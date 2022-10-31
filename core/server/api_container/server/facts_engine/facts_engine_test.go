package facts_engine

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
	"os"
	"testing"
	"time"
)

const (
	refreshInterval          = time.Millisecond
	waitUntilFactsAreUpdated = 2000 * refreshInterval
)

func TestFactEngineLoop(t *testing.T) {
	file, err := os.CreateTemp("/tmp", "*.db")
	defer os.Remove(file.Name())
	require.Nil(t, err)
	db, err := bolt.Open(file.Name(), 0666, nil)
	require.Nil(t, err)
	defer func(db *bolt.DB) {
		err := db.Close()
		if err != nil {
			require.Nil(t, err)
		}
	}(db)
	factsEngine := NewFactsEngine(db, service_network.NewEmptyMockServiceNetwork())
	factsEngine.Start()
	factValue := &kurtosis_core_rpc_api_bindings.FactValue{
		FactValue: &kurtosis_core_rpc_api_bindings.FactValue_StringValue{
			StringValue: "value",
		},
	}
	factRecipe := binding_constructors.NewConstantFactRecipe("service_id", "fact_name", &kurtosis_core_rpc_api_bindings.ConstantFactRecipe{FactValue: factValue}, refreshInterval)
	err = factsEngine.PushRecipe(factRecipe)
	require.Nil(t, err)
	time.Sleep(waitUntilFactsAreUpdated) // Wait for the background workers to perform operations
	fetchedFactValues, err := factsEngine.FetchLatestFactValues("service_id.fact_name")
	require.Nil(t, err)
	require.NotEmpty(t, fetchedFactValues)
	require.Equal(t, fetchedFactValues[len(fetchedFactValues)-1].GetStringValue(), factValue.GetStringValue())
}

func TestFactRecipePersistence(t *testing.T) {
	file, err := os.CreateTemp("/tmp", "*.db")
	defer os.Remove(file.Name())
	require.Nil(t, err)
	db, err := bolt.Open(file.Name(), 0666, nil)
	require.Nil(t, err)
	defer func(db *bolt.DB) {
		err := db.Close()
		if err != nil {
			require.Nil(t, err)
		}
	}(db)
	factsEngine := NewFactsEngine(db, service_network.NewEmptyMockServiceNetwork())
	factsEngine.Start()
	factValue := &kurtosis_core_rpc_api_bindings.FactValue{
		FactValue: &kurtosis_core_rpc_api_bindings.FactValue_StringValue{
			StringValue: "value",
		},
	}
	factRecipe := binding_constructors.NewConstantFactRecipe("service_id", "fact_name", &kurtosis_core_rpc_api_bindings.ConstantFactRecipe{FactValue: factValue}, refreshInterval)
	err = factsEngine.PushRecipe(factRecipe)
	require.Nil(t, err)
	time.Sleep(waitUntilFactsAreUpdated) // Wait for the background workers to perform operations
	factsEngine.Stop()
	err = db.Close()
	require.Nil(t, err)
	otherDb, err := bolt.Open(file.Name(), 0666, nil)
	require.Nil(t, err)
	defer func() {
		err := otherDb.Close()
		require.Nil(t, err)
	}()
	secondEngineTimestamp := time.Now().UnixNano()
	otherFactsEngine := NewFactsEngine(otherDb, service_network.NewEmptyMockServiceNetwork())
	otherFactsEngine.Start()
	time.Sleep(waitUntilFactsAreUpdated) // Wait for the background workers to perform operations
	fetchedFactValues, err := otherFactsEngine.FetchLatestFactValues("service_id.fact_name")
	require.Nil(t, err)
	require.NotEmpty(t, fetchedFactValues)
	require.Greater(t, fetchedFactValues[len(fetchedFactValues)-1].GetUpdatedAt().AsTime().UnixNano(), secondEngineTimestamp)
}

func TestFactRecipeFetchValueAfter(t *testing.T) {
	startTestTimestamp := time.Now()
	file, err := os.CreateTemp("/tmp", "*.db")
	defer os.Remove(file.Name())
	require.Nil(t, err)
	db, err := bolt.Open(file.Name(), 0666, nil)
	require.Nil(t, err)
	defer func(db *bolt.DB) {
		err := db.Close()
		if err != nil {
			require.Nil(t, err)
		}
	}(db)
	factsEngine := NewFactsEngine(db, service_network.NewEmptyMockServiceNetwork())
	factsEngine.Start()
	factValue := &kurtosis_core_rpc_api_bindings.FactValue{
		FactValue: &kurtosis_core_rpc_api_bindings.FactValue_StringValue{
			StringValue: "value",
		},
	}
	factRecipe := binding_constructors.NewConstantFactRecipe("service_id", "fact_name", &kurtosis_core_rpc_api_bindings.ConstantFactRecipe{FactValue: factValue}, refreshInterval)
	err = factsEngine.PushRecipe(factRecipe)
	require.Nil(t, err)
	time.Sleep(waitUntilFactsAreUpdated) // Wait for the background workers to perform operations
	fetchedFactValues, err := factsEngine.FetchFactValuesAfter("service_id.fact_name", startTestTimestamp)
	require.Nil(t, err)
	require.NotEmpty(t, fetchedFactValues)
	require.Greater(t, fetchedFactValues[len(fetchedFactValues)-1].GetUpdatedAt().AsTime().UnixNano(), fetchedFactValues[0].GetUpdatedAt().AsTime().UnixNano())
	require.Equal(t, fetchedFactValues[0].GetStringValue(), factValue.GetStringValue())

}
