package dependency_graph

import (
	"fmt"
	"os"
	"testing"

	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers/magic_string_helper"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/types"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
	"go.starlark.net/starlark"
)

const (
	testStringRuntimeValue = starlark.String("test-runtime-value")
	testRuntimeValueField  = "field.subfield"
	starlarkThreadName     = "starlark-value-serde-for-test-in-dependency-graph-thread"
)

func TestServiceDependencies(t *testing.T) {
	instruction1 := types.ScheduledInstructionUuid("instruction1")
	instruction2 := types.ScheduledInstructionUuid("instruction2")
	serviceName := "test-service"

	graph := NewInstructionDependencyGraph([]types.ScheduledInstructionUuid{instruction1, instruction2})

	graph.ProducesService(instruction1, serviceName)
	graph.ConsumesService(instruction2, serviceName)

	dependencies := graph.GenerateDependencyGraph()

	require.Len(t, dependencies[instruction2], 1)
	require.Equal(t, instruction1, dependencies[instruction2][0])
	require.Len(t, dependencies[instruction1], 0)
}

func TestFilesArtifactDependencies(t *testing.T) {
	instruction1 := types.ScheduledInstructionUuid("instruction1")
	instruction2 := types.ScheduledInstructionUuid("instruction2")
	artifactName := "test-artifact"

	graph := NewInstructionDependencyGraph([]types.ScheduledInstructionUuid{instruction1, instruction2})

	graph.ProducesFilesArtifact(instruction1, artifactName)
	graph.ConsumesFilesArtifact(instruction2, artifactName)

	dependencies := graph.GenerateDependencyGraph()

	require.Len(t, dependencies[instruction2], 1)
	require.Equal(t, instruction1, dependencies[instruction2][0])
	require.Len(t, dependencies[instruction1], 0)
}

func TestRuntimeValueDependencies(t *testing.T) {
	runtimeValue := getTestRuntimeValue(t)
	instruction1 := types.ScheduledInstructionUuid("instruction1")
	instruction2 := types.ScheduledInstructionUuid("instruction2")

	graph := NewInstructionDependencyGraph([]types.ScheduledInstructionUuid{instruction1, instruction2})

	graph.ProducesRuntimeValue(instruction1, runtimeValue)
	graph.ConsumesAnyRuntimeValuesInString(instruction2, fmt.Sprintf("some text with %s", runtimeValue))

	dependencies := graph.GenerateDependencyGraph()

	require.Len(t, dependencies[instruction2], 1)
	require.Equal(t, instruction1, dependencies[instruction2][0])
	require.Len(t, dependencies[instruction1], 0)
}

func TestRuntimeValueDependenciesInList(t *testing.T) {
	runtimeValue := getTestRuntimeValue(t)
	instruction1 := types.ScheduledInstructionUuid("instruction1")
	instruction2 := types.ScheduledInstructionUuid("instruction2")

	graph := NewInstructionDependencyGraph([]types.ScheduledInstructionUuid{instruction1, instruction2})

	graph.ProducesRuntimeValue(instruction1, runtimeValue)
	graph.ConsumesAnyRuntimeValuesInString(instruction2, fmt.Sprintf("some text with %s", runtimeValue))

	dependencies := graph.GenerateDependencyGraph()

	require.Len(t, dependencies[instruction2], 1)
	require.Equal(t, instruction1, dependencies[instruction2][0])
	require.Len(t, dependencies[instruction1], 0)
}

func TestPrintInstructionDependencies(t *testing.T) {
	instruction1 := types.ScheduledInstructionUuid("instruction1")
	instruction2 := types.ScheduledInstructionUuid("instruction2")

	graph := NewInstructionDependencyGraph([]types.ScheduledInstructionUuid{instruction1, instruction2})

	graph.AddPrintInstruction(instruction2)

	dependencies := graph.GenerateDependencyGraph()

	require.Len(t, dependencies[instruction2], 1)
	require.Equal(t, instruction1, dependencies[instruction2][0])
	require.Len(t, dependencies[instruction1], 0)
}

func TestMultipleDependencies(t *testing.T) {
	instruction1 := types.ScheduledInstructionUuid("instruction1")
	instruction2 := types.ScheduledInstructionUuid("instruction2")
	instruction3 := types.ScheduledInstructionUuid("instruction3")
	serviceName := "test-service"
	artifactName := "test-artifact"

	graph := NewInstructionDependencyGraph([]types.ScheduledInstructionUuid{instruction1, instruction2, instruction3})

	graph.ProducesService(instruction1, serviceName)
	graph.ProducesFilesArtifact(instruction2, artifactName)

	graph.ConsumesService(instruction3, serviceName)
	graph.ConsumesFilesArtifact(instruction3, artifactName)

	dependencies := graph.GenerateDependencyGraph()

	require.Len(t, dependencies[instruction3], 2)
	require.Contains(t, dependencies[instruction3], instruction1)
	require.Contains(t, dependencies[instruction3], instruction2)
	require.Len(t, dependencies[instruction1], 0)
	require.Len(t, dependencies[instruction2], 0)
}

func TestNoDependencies(t *testing.T) {
	instruction1 := types.ScheduledInstructionUuid("instruction1")
	instruction2 := types.ScheduledInstructionUuid("instruction2")

	graph := NewInstructionDependencyGraph([]types.ScheduledInstructionUuid{instruction1, instruction2})

	dependencies := graph.GenerateDependencyGraph()

	require.Len(t, dependencies[instruction1], 0)
	require.Len(t, dependencies[instruction2], 0)
}

func TestComplexDependencyChain(t *testing.T) {
	instruction1 := types.ScheduledInstructionUuid("instruction1")
	instruction2 := types.ScheduledInstructionUuid("instruction2")
	instruction3 := types.ScheduledInstructionUuid("instruction3")
	instruction4 := types.ScheduledInstructionUuid("instruction4")

	serviceName := "service-1"
	artifactName := "artifact-1"
	runtimeValue := getTestRuntimeValue(t)

	graph := NewInstructionDependencyGraph([]types.ScheduledInstructionUuid{instruction1, instruction2, instruction3, instruction4})

	graph.ProducesService(instruction1, serviceName)
	graph.ConsumesService(instruction2, serviceName)

	graph.ProducesFilesArtifact(instruction2, artifactName)
	graph.ConsumesFilesArtifact(instruction3, artifactName)

	graph.ProducesRuntimeValue(instruction3, runtimeValue)
	graph.ConsumesAnyRuntimeValuesInString(instruction4, fmt.Sprintf("some text with %s", runtimeValue))

	dependencies := graph.GenerateDependencyGraph()

	require.Len(t, dependencies[instruction2], 1)
	require.Equal(t, instruction1, dependencies[instruction2][0])

	require.Len(t, dependencies[instruction3], 1)
	require.Equal(t, instruction2, dependencies[instruction3][0])

	require.Len(t, dependencies[instruction4], 1)
	require.Equal(t, instruction3, dependencies[instruction4][0])

	require.Len(t, dependencies[instruction1], 0)
}

func TestRemoveServiceFromDependencyGraph(t *testing.T) {
	instruction1 := types.ScheduledInstructionUuid("remove-service(name = 'service-1')")

	serviceName := "service-1"

	graph := NewInstructionDependencyGraph([]types.ScheduledInstructionUuid{instruction1})

	graph.ConsumesService(instruction1, serviceName)

	dependencies := graph.GenerateDependencyGraph()

	require.Len(t, dependencies[instruction1], 0)
}

func getEnclaveDBForTest(t *testing.T) *enclave_db.EnclaveDB {
	file, err := os.CreateTemp("/tmp", "*.db")
	defer func() {
		err = os.Remove(file.Name())
		require.NoError(t, err)
	}()

	require.NoError(t, err)
	db, err := bolt.Open(file.Name(), 0666, nil)
	require.NoError(t, err)
	enclaveDb := &enclave_db.EnclaveDB{
		DB: db,
	}

	return enclaveDb
}

func newDummyStarlarkValueSerDeForTest() *kurtosis_types.StarlarkValueSerde {
	starlarkThread := &starlark.Thread{
		Name:       starlarkThreadName,
		Print:      nil,
		Load:       nil,
		OnMaxSteps: nil,
		Steps:      0,
	}
	starlarkEnv := starlark.StringDict{}

	serde := kurtosis_types.NewStarlarkValueSerde(starlarkThread, starlarkEnv)

	return serde
}

func getTestRuntimeValue(t *testing.T) string {
	enclaveDb := getEnclaveDBForTest(t)

	dummySerde := newDummyStarlarkValueSerDeForTest()
	runtimeValueStore, err := runtime_value_store.CreateRuntimeValueStore(dummySerde, enclaveDb)
	require.NoError(t, err)

	stringValueUuid, err := runtimeValueStore.CreateValue()
	require.NoError(t, err)
	err = runtimeValueStore.SetValue(stringValueUuid, map[string]starlark.Comparable{testRuntimeValueField: testStringRuntimeValue})
	require.NoError(t, err)

	return fmt.Sprintf(magic_string_helper.RuntimeValueReplacementPlaceholderFormat, stringValueUuid, testRuntimeValueField)
}
