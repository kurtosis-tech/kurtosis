package store_service_files

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/enclave_data_directory"
	"github.com/stretchr/testify/require"
	"testing"
)

var emptyServiceNetwork = service_network.NewEmptyMockServiceNetwork()

func TestStoreFilesFromService_StringRepresentationWorks(t *testing.T) {
	testFilesArtifactUuid, err := enclave_data_directory.NewFilesArtifactUUID()
	require.Nil(t, err)
	storeFileFromServiceInstruction := NewStoreFilesFromServiceInstruction(
		emptyServiceNetwork,
		kurtosis_instruction.NewInstructionPosition(1, 1, "dummyFile"),
		"example-service-id",
		"/tmp/foo",
		testFilesArtifactUuid,
	)
	expectedMultiLineStr := `# from: dummyFile[1:1]
store_service_files(
	artifact_id="` + string(testFilesArtifactUuid) + `",
	service_id="example-service-id",
	src="/tmp/foo"
)`
	require.Equal(t, expectedMultiLineStr, storeFileFromServiceInstruction.GetCanonicalInstruction())
	expectedSingleLineStr := `store_service_files(artifact_id="` + string(testFilesArtifactUuid) + `", service_id="example-service-id", src_path="/tmp/foo")`
	require.Equal(t, expectedSingleLineStr, storeFileFromServiceInstruction.String())
}
