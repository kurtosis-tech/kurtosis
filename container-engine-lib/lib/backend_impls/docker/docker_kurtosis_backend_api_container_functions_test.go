package docker

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestNewTempDirSuccessfullCreation(t *testing.T) {
	serviceGuid := "datastore-0-1650572435"
	tempDirPattern := tempDirForUserServiceFilesPrefix + string(serviceGuid) + tempDirForUserServiceFilesSuffix
	createdTempDir, err := newTempDir(tempDirPattern)
	require.NoErrorf(t, err, "An error occurred creating new temporary directory using pattern '%v'", tempDirPattern)
	require.Contains(t, createdTempDir, tempDirPattern, "The created temporary directory '%v' does not contain pattern '%v'", createdTempDir, tempDirPattern)

	fileInfo, err := os.Stat(createdTempDir)
	require.NoErrorf(t, err, "An error occurred getting file stat from temporary directory '%v'", createdTempDir)
	require.Truef(t, fileInfo.IsDir(), "Expected to create a new temporary directory with path '%v' but it is not a directory", createdTempDir)
}

