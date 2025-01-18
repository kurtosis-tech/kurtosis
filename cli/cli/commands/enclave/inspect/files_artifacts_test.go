package inspect

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/core/kurtosis_core_rpc_api_bindings"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFilesArtifacts_sortFileNamesAndUuids(t *testing.T) {
	var fileNamesAndUuids []*kurtosis_core_rpc_api_bindings.FilesArtifactNameAndUuid

	fileNameAndUuid1 := &kurtosis_core_rpc_api_bindings.FilesArtifactNameAndUuid{
		FileName: "test-file-1",
		FileUuid: "000-000-0",
	}

	fileNameAndUuid2 := &kurtosis_core_rpc_api_bindings.FilesArtifactNameAndUuid{
		FileName: "test-file-0",
		FileUuid: "123-edf-0",
	}

	fileNamesAndUuids = append(fileNamesAndUuids, fileNameAndUuid1)
	fileNamesAndUuids = append(fileNamesAndUuids, fileNameAndUuid2)

	sortedFileNamesAndUuids := sortFileNamesAndUuids(fileNamesAndUuids)
	require.Len(t, sortedFileNamesAndUuids, 2)
	require.Equal(t, sortedFileNamesAndUuids[0], fileNameAndUuid2)
	require.Equal(t, sortedFileNamesAndUuids[1], fileNameAndUuid1)
}
