package shared_utils

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

const (
	defaultPerm = 0700
)

func TestListFilesInPathDeterministic(t *testing.T) {
	dirPath := createRandomFolderStruct(t)

	files, err := listFilesInPathDeterministic(dirPath, false)
	require.NoError(t, err)
	require.Len(t, files, 3)
}

func TestListFilesInPathDeterministicRecursive(t *testing.T) {
	dirPath := createRandomFolderStruct(t)

	files, err := listFilesInPathDeterministic(dirPath, true)
	require.NoError(t, err)
	require.Len(t, files, 4)
}

func TestCompressPath(t *testing.T) {
	dirPath := createRandomFolderStruct(t)

	compressedData, md5, err := CompressPath(dirPath, false)
	require.NoError(t, err)

	os.WriteFile(dirPath+"/result.tgz", compressedData, defaultPerm)
	require.NotNil(t, compressedData)
	require.NotNil(t, md5)
}

func createRandomFolderStruct(t *testing.T) string {
	dirPath, err := os.MkdirTemp("", "test-dir-*")
	require.NoError(t, err)

	_, err = os.Create(dirPath + "/file_1.txt")
	require.NoError(t, err)
	_, err = os.Create(dirPath + "/file_2.txt")
	require.NoError(t, err)

	err = os.Mkdir(dirPath+"/level1", defaultPerm)
	require.NoError(t, err)
	_, err = os.Create(dirPath + "/level1" + "/file_3.txt")
	require.NoError(t, err)

	err = os.Mkdir(dirPath+"/level1"+"/level2", defaultPerm)
	require.NoError(t, err)
	_, err = os.Create(dirPath + "/level1" + "/level2" + "/file_3.txt")
	require.NoError(t, err)

	err = os.Mkdir(dirPath+"/level1"+"/emptydir", defaultPerm)
	require.NoError(t, err)
	return dirPath
}
