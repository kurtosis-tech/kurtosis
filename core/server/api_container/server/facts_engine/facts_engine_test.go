package facts_engine

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
	"os"
	"strconv"
	"testing"
	"time"
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
	recipeChannel := make(chan *kurtosis_core_rpc_api_bindings.FactRecipe)
	defer close(recipeChannel)
	factsEngine := NewFactsEngine(db, recipeChannel)
	factsEngine.Start()
	factValue := &kurtosis_core_rpc_api_bindings.FactValue{
		FactValue: &kurtosis_core_rpc_api_bindings.FactValue_StringValue{
			StringValue: "value",
		},
	}
	recipeChannel <- &kurtosis_core_rpc_api_bindings.FactRecipe{
		ServiceId: "service_id",
		FactName:  "fact_name",
		FactRecipe: &kurtosis_core_rpc_api_bindings.FactRecipe_ConstantFact{
			ConstantFact: &kurtosis_core_rpc_api_bindings.ConstantFactRecipe{
				FactValue: factValue,
			},
		},
	}
	time.Sleep(1 * time.Second) // Wait for the background workers to perform operations
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
	recipeChannel := make(chan *kurtosis_core_rpc_api_bindings.FactRecipe)
	defer close(recipeChannel)
	factsEngine := NewFactsEngine(db, recipeChannel)
	factsEngine.Start()
	factValue := &kurtosis_core_rpc_api_bindings.FactValue{
		FactValue: &kurtosis_core_rpc_api_bindings.FactValue_StringValue{
			StringValue: "value",
		},
	}
	recipeChannel <- &kurtosis_core_rpc_api_bindings.FactRecipe{
		ServiceId: "service_id",
		FactName:  "fact_name",
		FactRecipe: &kurtosis_core_rpc_api_bindings.FactRecipe_ConstantFact{
			ConstantFact: &kurtosis_core_rpc_api_bindings.ConstantFactRecipe{
				FactValue: factValue,
			},
		},
	}
	time.Sleep(1 * time.Second) // Wait for the background workers to perform operations
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
	otherRecipeChannel := make(chan *kurtosis_core_rpc_api_bindings.FactRecipe)
	otherFactsEngine := NewFactsEngine(otherDb, otherRecipeChannel)
	otherFactsEngine.Start()
	time.Sleep(1 * time.Second) // Wait for the background workers to perform operations
	savedTimestampStr, _, err := otherFactsEngine.FetchLatestFactValue("service_id.fact_name")
	require.Nil(t, err)
	savedTimestamp, err := strconv.ParseInt(savedTimestampStr, 10, 64)
	require.Nil(t, err)
	require.Greater(t, savedTimestamp, secondEngineTimestamp)
}
