package facts_engine

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
	"io"
	"strconv"
	"time"
)

type FactsEngine struct {
	db              *bolt.DB
	recipeChannel   chan *kurtosis_core_rpc_api_bindings.FactRecipe
	doneChannelList []chan bool
	serviceNetwork  service_network.ServiceNetwork
}

var (
	factValuesBucketName  = []byte("fact_values")
	factRecipesBucketName = []byte("fact_recipes")
)

func NewFactsEngine(db *bolt.DB, serviceNetwork service_network.ServiceNetwork) *FactsEngine {
	return &FactsEngine{
		db,
		make(chan *kurtosis_core_rpc_api_bindings.FactRecipe),
		[]chan bool{},
		serviceNetwork,
	}
}

func (engine *FactsEngine) Start() {
	go engine.consumeRecipeChannel()
	err := engine.restoreStoredRecipes()
	if err != nil {
		logrus.Infof("No fact recipes were found on the database")
	}
}

func (engine *FactsEngine) Stop() {
	close(engine.recipeChannel)
	for _, doneChannel := range engine.doneChannelList {
		doneChannel <- true
	}
}

func (engine *FactsEngine) PushRecipe(recipe *kurtosis_core_rpc_api_bindings.FactRecipe) {
	engine.recipeChannel <- recipe
}

func (engine *FactsEngine) FetchLatestFactValue(factId string) (string, *kurtosis_core_rpc_api_bindings.FactValue, error) {
	returnFactValue := kurtosis_core_rpc_api_bindings.FactValue{}
	var returnTimestamp string
	err := engine.db.View(func(tx *bolt.Tx) error {
		factValuesBucket := tx.Bucket(factValuesBucketName)
		if factValuesBucket == nil {
			return stacktrace.NewError("An error occurred because the bucket '%v' wasn't found on the database", factValuesBucket)
		}
		factBucket := factValuesBucket.Bucket([]byte(factId))
		if factBucket == nil {
			return stacktrace.NewError("An error occurred because the bucket '%v' wasn't found on the database", factBucket)
		}
		timestamp, factValue := factBucket.Cursor().Last()
		if timestamp == nil {
			return stacktrace.NewError("An error occurred because no fact value was found for fact '%v'", factId)
		}
		returnTimestamp = string(timestamp)
		err := proto.Unmarshal(factValue, &returnFactValue)
		if err != nil {
			return stacktrace.NewError("An error occurred when unmarshalling fact value")
		}
		return err
	})
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred when fetching latest fact value")
	}
	return returnTimestamp, &returnFactValue, nil
}

func (engine *FactsEngine) restoreStoredRecipes() error {
	return engine.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(factRecipesBucketName)
		if bucket == nil {
			logrus.Infof("No fact recipes were found on the database")
			return nil
		}
		restoredRecipes := 0
		err := bucket.ForEach(func(key, value []byte) error {
			unmarshalledFactRecipe := &kurtosis_core_rpc_api_bindings.FactRecipe{}
			err := proto.Unmarshal(key, unmarshalledFactRecipe)
			if err != nil {
				return stacktrace.Propagate(err, "An error occurred when restoring recipe")
			}
			engine.recipeChannel <- unmarshalledFactRecipe
			restoredRecipes += 1
			return nil
		})
		logrus.Infof("%v fact recipes were restored from the database", restoredRecipes)
		if err != nil {
			return err
		}
		return nil
	})
}

func (engine *FactsEngine) consumeRecipeChannel() {
	for recipe := range engine.recipeChannel {
		err := engine.persistRecipe(recipe)
		if err != nil {
			logrus.Errorf(stacktrace.Propagate(err, "An error occurred when persisting recipe").Error())
			continue
		}
		doneChannel := make(chan bool)
		engine.doneChannelList = append(engine.doneChannelList, doneChannel)
		go engine.runRecipeLoop(doneChannel, recipe)
	}
}

func (engine *FactsEngine) persistRecipe(recipe *kurtosis_core_rpc_api_bindings.FactRecipe) error {
	return engine.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(factRecipesBucketName)
		if err != nil {
			return stacktrace.Propagate(err, "Failure creating or retrieving bucket '%v'", factRecipesBucketName)
		}
		marshaledFactRecipe, err := proto.Marshal(recipe)
		if err != nil {
			return stacktrace.Propagate(err, "Failure marshaling recipe '%v'", recipe)
		}
		err = bucket.Put(marshaledFactRecipe, []byte{})
		if err != nil {
			return stacktrace.Propagate(err, "Failure saving marshaled recipe '%v'", recipe)
		}
		return nil
	})
}

func (engine *FactsEngine) runRecipeLoop(done <-chan bool, recipe *kurtosis_core_rpc_api_bindings.FactRecipe) {
	for {
		select {
		case <-done:
			return
		default:
			factId := fmt.Sprintf("%v.%v", recipe.GetServiceId(), recipe.GetFactName())
			timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)
			factValue, err := engine.runRecipe(recipe)
			if err != nil {
				logrus.Errorf(stacktrace.Propagate(err, "An error occurred when running recipe").Error())
				// TODO(victor.colombo): Run exponential backoff
				time.Sleep(2 * time.Second)
				continue
			}
			marshaledFactValue, err := proto.Marshal(factValue)
			if err != nil {
				logrus.Errorf(stacktrace.Propagate(err, "An error occurred when marshaling fact value").Error())
				// TODO(victor.colombo): Run exponential backoff
				time.Sleep(2 * time.Second)
				continue
			}
			err = engine.updateFactValue(factId, timestamp, marshaledFactValue)
			if err != nil {
				logrus.Errorf(err.Error())
				time.Sleep(2 * time.Second)
				continue
			}
			time.Sleep(2 * time.Second)
		}
	}
}

func (engine *FactsEngine) runRecipe(recipe *kurtosis_core_rpc_api_bindings.FactRecipe) (*kurtosis_core_rpc_api_bindings.FactValue, error) {
	if recipe.GetConstantFact() != nil {
		return recipe.GetConstantFact().GetFactValue(), nil
	}
	if recipe.GetExecFact() != nil {
		_, result, err := engine.serviceNetwork.ExecCommand(context.Background(), service.ServiceID(recipe.GetServiceId()), []string{recipe.GetExecFact().GetExecString()})
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred when running exec recipe")
		}
		return &kurtosis_core_rpc_api_bindings.FactValue{
			FactValue: &kurtosis_core_rpc_api_bindings.FactValue_StringValue{
				StringValue: result,
			},
		}, nil
	}
	if recipe.GetHttpRequestFact() != nil {
		response, err := engine.serviceNetwork.HttpRequestService(
			context.Background(),
			service.ServiceID(recipe.GetServiceId()),
			recipe.GetHttpRequestFact().GetPortId(),
			recipe.GetHttpRequestFact().GetMethod().String(),
			recipe.GetHttpRequestFact().GetContentType(),
			recipe.GetHttpRequestFact().GetEndpoint(),
			recipe.GetHttpRequestFact().GetBody(),
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred when running HTTP request recipe")
		}
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred when reading HTTP response body")
		}
		return &kurtosis_core_rpc_api_bindings.FactValue{
			FactValue: &kurtosis_core_rpc_api_bindings.FactValue_StringValue{
				StringValue: string(body),
			},
		}, nil
	}
	panic("Recipe type not implemented!!!")
}

func (engine *FactsEngine) updateFactValue(factId string, timestamp string, value []byte) error {
	err := engine.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(factValuesBucketName)
		if err != nil {
			return stacktrace.Propagate(err, "Failure creating or retrieving bucket '%v'", factValuesBucketName)
		}
		factBucket, err := bucket.CreateBucketIfNotExists([]byte(factId))
		if err != nil {
			return stacktrace.Propagate(err, "Failure creating or retrieving bucket '%v'", factId)
		}
		err = factBucket.Put([]byte(timestamp), []byte(value))
		if err != nil {
			return stacktrace.Propagate(err, "Failure saving timestamp and value '%v' '%v'", timestamp, value)
		}
		return nil
	})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred when updating fact value '%v' '%v' '%v'", factId, timestamp, value)
	}
	return err
}
