/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_directory

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestGetFilesArtifactStore(t *testing.T) {
	enclaveDirpath, err := ioutil.TempDir("", "")
	assert.Nil(t, err)

	enclaveDir := NewEnclaveDataDirectory(enclaveDirpath)

	artifactStore, err := enclaveDir.GetFilesArtifactStore()
	assert.Nil(t, err)

	expectedAbsDirpath := path.Join(enclaveDirpath, artifactStoreDirname)
	_, err = os.Stat(expectedAbsDirpath)
	assert.Nil(t, err)
	assert.Equal(t, expectedAbsDirpath, artifactStore.fileCache.absoluteDirpath)

	expectedRelativeDirpath := artifactStoreDirname
	assert.Equal(t, expectedRelativeDirpath, artifactStore.fileCache.dirpathRelativeToDataDirRoot)

}