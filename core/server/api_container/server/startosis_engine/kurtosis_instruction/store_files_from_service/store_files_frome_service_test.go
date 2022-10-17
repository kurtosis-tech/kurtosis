package store_files_from_service

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/stretchr/testify/require"
	"testing"
)

var emptyServiceNetwork = service_network.NewEmptyMockServiceNetwork()

func TestStoreFilesFromService_StringRepresentationWorks(t *testing.T) {
	storeFileFromServiceInstruction := NewStoreFilesFromServicePosition(
		emptyServiceNetwork,
		*kurtosis_instruction.NewInstructionPosition(1, 1),
		"example-service-id",
		"/tmp/foo",
	)
	expectedStr := `store_file_from_service(service_id="example-service-id", src_path="/tmp/foo")`
	require.Equal(t, expectedStr, storeFileFromServiceInstruction.GetCanonicalInstruction())
	require.Equal(t, expectedStr, storeFileFromServiceInstruction.String())
}
