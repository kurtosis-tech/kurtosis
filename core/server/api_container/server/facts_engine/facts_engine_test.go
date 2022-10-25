package facts_engine

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/types/known/durationpb"
	"os"
	"strconv"
	"testing"
	"time"
)

const (
	refreshInterval          = time.Millisecond
	waitUntilFactsAreUpdated = 100 * refreshInterval
)

func TestFactEngineLoop(t *testing.T) {
	file, err := os.CreateTemp("/tmp", "*.db")
	defer os.Remove(file.Name())
	require.Nil(t, err)
	db, err := bolt.Open(file.Name(), 0666, nil)
	defer func() {
		err := db.Close()
		require.Nil(t, err)
	}()
	require.Nil(t, err)
	factsEngine := NewFactsEngine(db)
	factsEngine.Start()
	factValue := &kurtosis_core_rpc_api_bindings.FactValue{
		FactValue: &kurtosis_core_rpc_api_bindings.FactValue_StringValue{
			StringValue: "value",
		},
	}
	err = factsEngine.PushRecipe(&kurtosis_core_rpc_api_bindings.FactRecipe{
		ServiceId: "service_id",
		FactName:  "fact_name",
		FactRecipe: &kurtosis_core_rpc_api_bindings.FactRecipe_ConstantFact{
			ConstantFact: &kurtosis_core_rpc_api_bindings.ConstantFactRecipe{
				FactValue: factValue,
			},
		},
		RefreshInterval: durationpb.New(refreshInterval),
	})
	require.Nil(t, err)
	time.Sleep(waitUntilFactsAreUpdated) // Wait for the background workers to perform operations
	_, fetchedFactValue, err := factsEngine.FetchLatestFactValue("service_id.fact_name")
	require.Nil(t, err)
	require.Equal(t, fetchedFactValue.GetStringValue(), factValue.GetStringValue())
}

func TestFactRecipePersistence(t *testing.T) {
	file, err := os.CreateTemp("/tmp", "*.db")
	defer os.Remove(file.Name())
	require.Nil(t, err)
	db, err := bolt.Open(file.Name(), 0666, nil)
	require.Nil(t, err)
	factsEngine := NewFactsEngine(db)
	factsEngine.Start()
	factValue := &kurtosis_core_rpc_api_bindings.FactValue{
		FactValue: &kurtosis_core_rpc_api_bindings.FactValue_StringValue{
			StringValue: "value",
		},
	}
	err = factsEngine.PushRecipe(&kurtosis_core_rpc_api_bindings.FactRecipe{
		ServiceId: "service_id",
		FactName:  "fact_name",
		FactRecipe: &kurtosis_core_rpc_api_bindings.FactRecipe_ConstantFact{
			ConstantFact: &kurtosis_core_rpc_api_bindings.ConstantFactRecipe{
				FactValue: factValue,
			},
		},
		RefreshInterval: durationpb.New(refreshInterval),
	})
	require.Nil(t, err)
	time.Sleep(waitUntilFactsAreUpdated) // Wait for the background workers to perform operations
	factsEngine.Stop()
	err = db.Close()
	require.Nil(t, err)
	otherDb, err := bolt.Open(file.Name(), 0666, nil)
	defer func() {
		err := otherDb.Close()
		require.Nil(t, err)
	}()
	require.Nil(t, err)
	secondEngineTimestamp := time.Now().UnixNano()
	otherFactsEngine := NewFactsEngine(otherDb)
	otherFactsEngine.Start()
	time.Sleep(waitUntilFactsAreUpdated) // Wait for the background workers to perform operations
	savedTimestampStr, _, err := otherFactsEngine.FetchLatestFactValue("service_id.fact_name")
	require.Nil(t, err)
	savedTimestamp, err := strconv.ParseInt(savedTimestampStr, 10, 64)
	require.Nil(t, err)
	require.Greater(t, savedTimestamp, secondEngineTimestamp)
}
