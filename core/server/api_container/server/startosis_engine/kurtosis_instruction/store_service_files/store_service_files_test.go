package store_service_files

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/enclave_data_directory"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

var emptyServiceNetwork = service_network.NewEmptyMockServiceNetwork()

func TestStoreFilesFromService_StringRepresentationWorks(t *testing.T) {
	position := kurtosis_instruction.NewInstructionPosition(1, 1, "dummyFile")
	testFilesArtifactId, err := enclave_data_directory.NewFilesArtifactUUID()
	require.Nil(t, err)
	storeFileFromServiceInstruction := newEmptyStoreServiceFilesInstruction(emptyServiceNetwork, position)
	storeFileFromServiceInstruction.starlarkKwargs = starlark.StringDict{}
	storeFileFromServiceInstruction.starlarkKwargs[serviceIdArgName] = starlark.String("example-service-id")
	storeFileFromServiceInstruction.starlarkKwargs[srcArgName] = starlark.String("/tmp/foo")
	storeFileFromServiceInstruction.starlarkKwargs[nonOptionalArtifactIdArgName] = starlark.String(testFilesArtifactId)

	expectedStr := `store_service_files(artifact_id="` + string(testFilesArtifactId) + `", service_id="example-service-id", src="/tmp/foo")`
	require.Equal(t, expectedStr, storeFileFromServiceInstruction.String())

	canonicalInstruction := binding_constructors.NewKurtosisInstruction(
		position.ToAPIType(),
		StoreServiceFilesBuiltinName,
		expectedStr,
		[]*kurtosis_core_rpc_api_bindings.KurtosisInstructionArg{
			binding_constructors.NewKurtosisInstructionKwarg(`"example-service-id"`, serviceIdArgName, true),
			binding_constructors.NewKurtosisInstructionKwarg(`"/tmp/foo"`, srcArgName, true),
			binding_constructors.NewKurtosisInstructionKwarg(`"`+string(testFilesArtifactId)+`"`, nonOptionalArtifactIdArgName, true),
		})
	require.Equal(t, canonicalInstruction, storeFileFromServiceInstruction.GetCanonicalInstruction())
}
