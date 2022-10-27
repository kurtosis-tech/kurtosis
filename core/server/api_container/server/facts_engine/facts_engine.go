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
type cursorMovement func(*bolt.Cursor) ([]byte, []byte)
type cursorInitializer func(*bolt.Bucket) (*bolt.Cursor, []byte, []byte)

type FactsEngine struct {
	db             *bolt.DB
	exitChanMap    map[FactId]chan bool
	serviceNetwork service_network.ServiceNetwork
	lock           *sync.Mutex
}

var (
	factValuesBucketName  = []byte(factValuesBucketNameStr)
	factRecipesBucketName = []byte(factRecipesBucketNameStr)
)

const (
	factValuesBucketNameStr    = "fact_values"
	factRecipesBucketNameStr   = "fact_recipes"
	defaultWaitTimeBetweenRuns = 2 * time.Second
	factIdFormatStr            = "%v.%v"
	keyStringFormat            = "%020s"
	maxResultCount             = 100
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
	logrus.Info("Facts engine has started")
}

func (engine *FactsEngine) Stop() {
	for _, exitChan := range engine.exitChanMap {
		exitChan <- true
		close(exitChan)
	}
	logrus.Info("Facts engine has stopped")
}

func (engine *FactsEngine) PushRecipe(recipe *kurtosis_core_rpc_api_bindings.FactRecipe) error {
	// Locking avoid the condition where two recipes of the same fact are pushed at the same time
	engine.lock.Lock()
	defer engine.lock.Unlock()
	factId := GetFactId(recipe.GetServiceId(), recipe.GetFactName())
	if err := engine.persistRecipe(recipe); err != nil {
		return stacktrace.Propagate(err, "An error occurred when persisting recipe for fact '%v'", factId)
	}
	engine.setupRunRecipeLoop(factId, recipe)
	return nil
}

func (engine *FactsEngine) setupRunRecipeLoop(factId FactId, recipe *kurtosis_core_rpc_api_bindings.FactRecipe) {
	exitChan, isRunning := engine.exitChanMap[factId]
	if isRunning {
		logrus.Infof("Stopped running fact '%v' to run new recipe", factId)
		exitChan <- true
	} else {
		logrus.Infof("Setting up and running fact '%v'", factId)
	}
	engine.exitChanMap[factId] = make(chan bool)
	go engine.runRecipeLoop(factId, engine.exitChanMap[factId], recipe)
}

func (engine *FactsEngine) FetchLatestFactValues(factId FactId) ([]*kurtosis_core_rpc_api_bindings.FactValue, error) {
	return engine.getFactValues(factId, lastCursorInitializer, maxResultCount, cursorForwardStep)
}

func (engine *FactsEngine) FetchFactValuesAfter(factId FactId, afterTimestamp time.Time) ([]*kurtosis_core_rpc_api_bindings.FactValue, error) {
	timestampKey := []byte(getKeyFromTimestamp(afterTimestamp))
	return engine.getFactValues(factId, createSeekCursorInitializer(timestampKey), maxResultCount, cursorBackwardsStep)
}

func (engine *FactsEngine) getFactValues(factId FactId, initializer cursorInitializer, resultCount int, movement cursorMovement) ([]*kurtosis_core_rpc_api_bindings.FactValue, error) {
	returnFactValues := []*kurtosis_core_rpc_api_bindings.FactValue{}
	err := engine.db.View(func(tx *bolt.Tx) error {
		factValuesBucket := tx.Bucket(factValuesBucketName)
		if factValuesBucket == nil {
			return stacktrace.NewError("An error occurred because the bucket '%v' wasn't found on the database", factValuesBucketNameStr)
		}
		factBucket := factValuesBucket.Bucket([]byte(factId))
		if factBucket == nil {
			return stacktrace.NewError("An error occurred because the fact bucket '%v' wasn't found on the database", factId)
		}
		factValues := []*kurtosis_core_rpc_api_bindings.FactValue{}
		for cursor, timestampKey, factValue := initializer(factBucket); timestampKey != nil && resultCount > 0; timestampKey, factValue = movement(cursor) {
			unmarshalledFactValue := &kurtosis_core_rpc_api_bindings.FactValue{}
			if err := proto.Unmarshal(factValue, unmarshalledFactValue); err != nil {
				return stacktrace.Propagate(err, "An error occurred when unmarshalling fact value on key '%v'", string(timestampKey))
			}
			factValues = append(factValues, unmarshalledFactValue)
			resultCount--
		}
		returnFactValues = factValues
		return nil
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred when fetching latest fact value '%v'", factId)
	}
	return returnFactValues, nil
}

func (engine *FactsEngine) restoreStoredRecipes() error {
	return engine.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(factRecipesBucketName)
		if bucket == nil {
			logrus.Info("No fact recipes were found on the database")
			return nil
		}
		restoredRecipes := 0
		err := bucket.ForEach(func(storedRecipe, _ []byte) error {
			unmarshalledFactRecipe := &kurtosis_core_rpc_api_bindings.FactRecipe{}
			err := proto.Unmarshal(storedRecipe, unmarshalledFactRecipe)
			if err != nil {
				return stacktrace.Propagate(err, "An error occurred when unmarshalling recipe")
			}
			factId := GetFactId(unmarshalledFactRecipe.GetServiceId(), unmarshalledFactRecipe.GetFactName())
			engine.setupRunRecipeLoop(factId, unmarshalledFactRecipe)
			restoredRecipes += 1
			return nil
		})
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred when restoring recipes")
		}
		logrus.Infof("%d fact recipes were restored from the database", restoredRecipes)
		return nil
	})
}

func (engine *FactsEngine) persistRecipe(recipe *kurtosis_core_rpc_api_bindings.FactRecipe) error {
	return engine.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(factRecipesBucketName)
		if err != nil {
			return stacktrace.Propagate(err, "Failure creating or retrieving bucket '%v'", factRecipesBucketNameStr)
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
			now := time.Now()
			timestampKey := getKeyFromTimestamp(now)
			factValue, err := engine.runRecipe(recipe)
			factValue.UpdatedAt = timestamppb.New(now)
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
			err = engine.updateFactValue(factId, timestampKey, marshaledFactValue)
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

func (engine *FactsEngine) updateFactValue(factId FactId, timestampKey string, factValue []byte) error {
	err := engine.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(factValuesBucketName)
		if err != nil {
			return stacktrace.Propagate(err, "Failure creating or retrieving bucket '%v'", factValuesBucketName)
		}
		factBucket, err := bucket.CreateBucketIfNotExists([]byte(factId))
		if err != nil {
			return stacktrace.Propagate(err, "Failure creating or retrieving bucket '%v'", factId)
		}
		if err := factBucket.Put([]byte(timestampKey), factValue); err != nil {
			return stacktrace.Propagate(err, "Failure saving timestamp and value '%v' '%v'", timestampKey, factValue)
		}
		return nil
	})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred when updating fact value '%v' '%v' '%v'", factId, timestampKey, factValue)
	}
	return err
}

func getKeyFromTimestamp(timestamp time.Time) string {
	timestampStr := strconv.FormatInt(timestamp.UnixNano(), 10)
	return fmt.Sprintf(keyStringFormat, timestampStr)
}

func lastCursorInitializer(bucket *bolt.Bucket) (*bolt.Cursor, []byte, []byte) {
	cursor := bucket.Cursor()
	key, value := cursor.Last()
	return cursor, key, value
}

func createSeekCursorInitializer(seekKey []byte) cursorInitializer {
	return func(bucket *bolt.Bucket) (*bolt.Cursor, []byte, []byte) {
		cursor := bucket.Cursor()
		key, value := cursor.Seek(seekKey)
		return cursor, key, value
	}
}

func cursorForwardStep(cursor *bolt.Cursor) (key []byte, value []byte) {
	return cursor.Next()
}

func cursorBackwardsStep(cursor *bolt.Cursor) (key []byte, value []byte) {
	return cursor.Prev()
}
