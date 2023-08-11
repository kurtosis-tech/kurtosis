/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_directory

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"testing"
)

func TestGetFilesArtifactStore(t *testing.T) {
	enclaveDirpath, err := os.MkdirTemp("", "")
	assert.Nil(t, err)

	enclaveDir := NewEnclaveDataDirectory(enclaveDirpath)

	artifactStore, err := enclaveDir.GetFilesArtifactStore()
	assert.Nil(t, err)
	defer func() {
		require.Nil(t, enclave_db.EraseDatabase())
	}()

	expectedAbsDirpath := path.Join(enclaveDirpath, artifactStoreDirname)
	_, err = os.Stat(expectedAbsDirpath)
	assert.Nil(t, err)
	assert.Equal(t, expectedAbsDirpath, artifactStore.fileCache.absoluteDirpath)

	expectedRelativeDirpath := artifactStoreDirname
	assert.Equal(t, expectedRelativeDirpath, artifactStore.fileCache.dirpathRelativeToDataDirRoot)
}
