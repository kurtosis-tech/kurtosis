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
	factValuesBucketName  = []byte(factValuesBucketNameStr)
	factRecipesBucketName = []byte(factRecipesBucketNameStr)
)

const (
	factValuesBucketNameStr    = "fact_values"
	factRecipesBucketNameStr   = "fact_recipes"
	defaultWaitTimeBetweenRuns = 2 * time.Second
	factIdFormatStr            = "%v.%v"
)

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
			return stacktrace.NewError("An error occurred because the fact bucket '%v' wasn't found on the database", factId)
		}
		timestamp, factValue := factBucket.Cursor().Last()
		// If the bucket is empty then a nil key and value are returned.
		if timestamp == nil {
			return stacktrace.NewError("An error occurred because no fact value was found for fact '%v'", factId)
		}
		returnTimestamp = string(timestamp)
		if err := proto.Unmarshal(factValue, returnFactValue); err != nil {
			return stacktrace.Propagate(err, "An error occurred when unmarshalling fact value from '%v'", factId)
		}
		return nil
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
		err := bucket.ForEach(func(storedRecipe, _ []byte) error {
			unmarshalledFactRecipe := &kurtosis_core_rpc_api_bindings.FactRecipe{}
			err := proto.Unmarshal(storedRecipe, unmarshalledFactRecipe)
			if err != nil {
				return stacktrace.Propagate(err, "An error occurred when unmarshalling recipe")
			}
			factId := GetFactIdFromRecipe(unmarshalledFactRecipe)
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
			timestamp := strconv.FormatInt(now.UnixNano(), 10)
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
	return FactId(fmt.Sprintf(factIdFormatStr, recipe.GetServiceId(), recipe.GetFactName()))
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
		if err := factBucket.Put([]byte(timestamp), value); err != nil {
			return stacktrace.Propagate(err, "Failure saving timestamp and value '%v' '%v'", timestamp, value)
		}
		return nil
	})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred when updating fact value '%v' '%v' '%v'", factId, timestamp, value)
	}
	return err
}
