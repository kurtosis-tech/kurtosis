package facts_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
	"strconv"
	"sync"
	"time"
)

type FactId string

type FactsEngine struct {
	db          *bolt.DB
	exitChanMap map[FactId]chan bool
	lock        *sync.Mutex
}

var (
	factValuesBucketNameStr  = "fact_values"
	factValuesBucketName     = []byte(factValuesBucketNameStr)
	factRecipesBucketNameStr = "fact_recipes"
	factRecipesBucketName    = []byte(factRecipesBucketNameStr)
)

const defaultWaitTimeBetweenRuns = 2 * time.Second

func NewFactsEngine(db *bolt.DB) *FactsEngine {
	return &FactsEngine{
		db,
		make(map[FactId]chan bool),
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
	factId := GetFactIdFromRecipe(recipe)
	if err := engine.persistRecipe(recipe); err != nil {
		return stacktrace.Propagate(err, "An error occurred when persisting recipe for fact '%v'", factId)
	}
	if err := engine.setupRunRecipeLoop(recipe); err != nil {
		return stacktrace.Propagate(err, "An error occurred when setting up run recipe loop for fact '%v'", factId)
	}
	return nil
}

func (engine *FactsEngine) setupRunRecipeLoop(recipe *kurtosis_core_rpc_api_bindings.FactRecipe) error {
	factId := GetFactIdFromRecipe(recipe)
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

func (engine *FactsEngine) FetchLatestFactValue(factId FactId) (string, *kurtosis_core_rpc_api_bindings.FactValue, error) {
	returnFactValue := &kurtosis_core_rpc_api_bindings.FactValue{}
	var returnTimestamp string
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
		returnTimestamp = string(timestamp)
		err := proto.Unmarshal(factValue, returnFactValue)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred when unmarshalling fact value from '%v'", factId)
		}
		return err
	})
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred when fetching latest fact value '%v'", factId)
	}
	return returnTimestamp, returnFactValue, nil
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
			timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)
			factValue, err := engine.runRecipe(recipe)
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

func GetFactIdFromRecipe(recipe *kurtosis_core_rpc_api_bindings.FactRecipe) FactId {
	return FactId(fmt.Sprintf("%v.%v", recipe.GetServiceId(), recipe.GetFactName()))
}

func (engine *FactsEngine) runRecipe(recipe *kurtosis_core_rpc_api_bindings.FactRecipe) (*kurtosis_core_rpc_api_bindings.FactValue, error) {
	if recipe.GetConstantFact() != nil {
		return recipe.GetConstantFact().GetFactValue(), nil
	}
	return nil, stacktrace.NewError("An error occurred when running recipe")
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
