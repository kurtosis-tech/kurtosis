package store_service_files

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

var emptyServiceNetwork = service_network.NewEmptyMockServiceNetwork()

func TestStoreFilesFromService_StringRepresentationWorks(t *testing.T) {
	position := kurtosis_instruction.NewInstructionPosition(1, 1, "dummyFile")
	testArtifactName := "test-artifact"
	storeFileFromServiceInstruction := newEmptyStoreServiceFilesInstruction(emptyServiceNetwork, position)
	storeFileFromServiceInstruction.starlarkKwargs = starlark.StringDict{}
	storeFileFromServiceInstruction.starlarkKwargs[serviceNameArgName] = starlark.String("example-service-id")
	storeFileFromServiceInstruction.starlarkKwargs[srcArgName] = starlark.String("/tmp/foo")
	storeFileFromServiceInstruction.starlarkKwargs[artifactNameArgName] = starlark.String(testArtifactName)

	expectedStr := `store_service_files(name="` + testArtifactName + `", service_name="example-service-id", src="/tmp/foo")`
	require.Equal(t, expectedStr, storeFileFromServiceInstruction.String())

	canonicalInstruction := binding_constructors.NewStarlarkInstruction(
		position.ToAPIType(),
		StoreServiceFilesBuiltinName,
		expectedStr,
		[]*kurtosis_core_rpc_api_bindings.StarlarkInstructionArg{
			binding_constructors.NewStarlarkInstructionKwarg(`"example-service-id"`, serviceNameArgName, true),
			binding_constructors.NewStarlarkInstructionKwarg(`"/tmp/foo"`, srcArgName, true),
			binding_constructors.NewStarlarkInstructionKwarg(`"`+testArtifactName+`"`, artifactNameArgName, true),
		})
	require.Equal(t, canonicalInstruction, storeFileFromServiceInstruction.GetCanonicalInstruction())
}
