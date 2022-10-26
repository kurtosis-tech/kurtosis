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
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	"strconv"
	"sync"
	"time"
)

type FactId string

type FactsEngine struct {
	db             *bolt.DB
	exitChanMap    map[FactId]chan bool
	serviceNetwork service_network.ServiceNetwork
	lock           *sync.Mutex
}

var (
	factValuesBucketNameStr  = "fact_values"
	factValuesBucketName     = []byte(factValuesBucketNameStr)
	factRecipesBucketNameStr = "fact_recipes"
	factRecipesBucketName    = []byte(factRecipesBucketNameStr)
)

const (
	defaultWaitTimeBetweenRuns = 2 * time.Second
	factIdFormatStr            = "%v.%v"
)

func NewFactsEngine(db *bolt.DB, serviceNetwork service_network.ServiceNetwork) *FactsEngine {
	return &FactsEngine{
		db,
		make(map[FactId]chan bool),
		serviceNetwork,
		&sync.Mutex{},
	}
}

func (engine *FactsEngine) Start() {
	err := engine.restoreStoredRecipes()
	if err != nil {
		logrus.Info("No fact recipes were found on the database")
	}
}

func (engine *FactsEngine) Stop() {
	for _, exitChan := range engine.exitChanMap {
		exitChan <- true
		close(exitChan)
	}
}

func (engine *FactsEngine) PushRecipe(recipe *kurtosis_core_rpc_api_bindings.FactRecipe) error {
	// Locking avoid the condition where two recipes of the same fact are pushed at the same time
	engine.lock.Lock()
	defer engine.lock.Unlock()
	factId := GetFactId(recipe.GetServiceId(), recipe.GetFactName())
	if err := engine.persistRecipe(recipe); err != nil {
		return stacktrace.Propagate(err, "An error occurred when persisting recipe for fact '%v'", factId)
	}
	if err := engine.setupRunRecipeLoop(recipe); err != nil {
		return stacktrace.Propagate(err, "An error occurred when setting up run recipe loop for fact '%v'", factId)
	}
	return nil
}

func (engine *FactsEngine) setupRunRecipeLoop(recipe *kurtosis_core_rpc_api_bindings.FactRecipe) error {
	factId := GetFactId(recipe.GetServiceId(), recipe.GetFactName())
	exitChan, isRunning := engine.exitChanMap[factId]
	if isRunning {
		logrus.Infof("Stopped running fact '%v' to run new recipe", factId)
		exitChan <- true
	} else {
		logrus.Infof("Setting up and running fact '%v'", factId)
	}
	engine.exitChanMap[factId] = make(chan bool)
	go engine.runRecipeLoop(factId, engine.exitChanMap[factId], recipe)
	return nil
}

func (engine *FactsEngine) FetchLatestFactValue(factId FactId) (*kurtosis_core_rpc_api_bindings.FactValue, error) {
	returnFactValue := &kurtosis_core_rpc_api_bindings.FactValue{}
	err := engine.db.View(func(tx *bolt.Tx) error {
		factValuesBucket := tx.Bucket(factValuesBucketName)
		if factValuesBucket == nil {
			return stacktrace.NewError("An error occurred because the bucket '%v' wasn't found on the database", factValuesBucketNameStr)
		}
		factBucket := factValuesBucket.Bucket([]byte(factId))
		if factBucket == nil {
			return stacktrace.NewError("An error occurred because the bucket '%v' wasn't found on the database", factId)
		}
		timestamp, factValue := factBucket.Cursor().Last()
		if timestamp == nil {
			return stacktrace.NewError("An error occurred because no fact value was found for fact '%v'", factId)
		}
		err := proto.Unmarshal(factValue, returnFactValue)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred when unmarshalling fact value from '%v'", factId)
		}
		return err
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred when fetching latest fact value '%v'", factId)
	}
	return returnFactValue, nil
}

func (engine *FactsEngine) restoreStoredRecipes() error {
	return engine.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(factRecipesBucketName)
		if bucket == nil {
			logrus.Info("No fact recipes were found on the database")
			return nil
		}
		restoredRecipes := 0
		err := bucket.ForEach(func(key, value []byte) error {
			unmarshalledFactRecipe := &kurtosis_core_rpc_api_bindings.FactRecipe{}
			err := proto.Unmarshal(key, unmarshalledFactRecipe)
			if err != nil {
				return stacktrace.Propagate(err, "An error occurred when unmarshalling recipe")
			}
			err = engine.setupRunRecipeLoop(unmarshalledFactRecipe)
			if err != nil {
				return stacktrace.Propagate(err, "An error occurred when pushing restored recipe to engine")
			}
			restoredRecipes += 1
			return nil
		})
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred when restoring recipes")
		}
		return nil
	})
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

func (engine *FactsEngine) runRecipeLoop(factId FactId, exit <-chan bool, recipe *kurtosis_core_rpc_api_bindings.FactRecipe) {
	var ticker *time.Ticker
	if recipe.GetRefreshInterval() != nil {
		ticker = time.NewTicker(recipe.GetRefreshInterval().AsDuration())
	} else {
		ticker = time.NewTicker(defaultWaitTimeBetweenRuns)
	}
	for {
		select {
		case <-exit:
			return
		case <-ticker.C:
			// TODO(victor.colombo): Take hint from protobuf on how long to wait for it
			now := time.Now()
			timestamp := strconv.FormatInt(now.UnixNano(), 10)
			factValue, err := engine.runRecipe(recipe)
			factValue.UpdatedAt = timestamppb.New(time.Now())
			if err != nil {
				logrus.Errorf(stacktrace.Propagate(err, "An error occurred when running recipe").Error())
				// TODO(victor.colombo): Run exponential backoff
				continue
			}
			marshaledFactValue, err := proto.Marshal(factValue)
			if err != nil {
				logrus.Errorf(stacktrace.Propagate(err, "An error occurred when marshaling fact value").Error())
				// TODO(victor.colombo): Define what to do in case, and when this happens
				continue
			}
			err = engine.updateFactValue(factId, timestamp, marshaledFactValue)
			if err != nil {
				logrus.Errorf(stacktrace.Propagate(err, "An error occurred when updating fact value").Error())
				// TODO(victor.colombo): Define what to do in case, and when this happens
				continue
			}
		}
	}
}

func GetFactId(serviceId string, factName string) FactId {
	return FactId(fmt.Sprintf(factIdFormatStr, serviceId, factName))
}

func (engine *FactsEngine) runRecipe(recipe *kurtosis_core_rpc_api_bindings.FactRecipe) (*kurtosis_core_rpc_api_bindings.FactValue, error) {
	if recipe.GetConstantFact() != nil {
		return recipe.GetConstantFact().GetFactValue(), nil
	}
	if recipe.GetExecFact() != nil {
		_, result, err := engine.serviceNetwork.ExecCommand(context.Background(), service.ServiceID(recipe.GetServiceId()), recipe.GetExecFact().GetCmdArgs())
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

func (engine *FactsEngine) updateFactValue(factId FactId, timestamp string, value []byte) error {
	err := engine.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(factValuesBucketName)
		if err != nil {
			return stacktrace.Propagate(err, "Failure creating or retrieving bucket '%v'", factValuesBucketName)
		}
		factBucket, err := bucket.CreateBucketIfNotExists([]byte(factId))
		if err != nil {
			return stacktrace.Propagate(err, "Failure creating or retrieving bucket '%v'", factId)
		}
		err = factBucket.Put([]byte(timestamp), value)
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
