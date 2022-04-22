/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_directory

import (
	"testing"
	"io/ioutil"
	"github.com/stretchr/testify/assert"
	"os"
	"github.com/stretchr/testify/require"
	"strings"
)

func TestFileStore_FileStoreCreatedProperly(t *testing.T) {
	fileStore := getTestFileStore(t)
	assert.Equal(t, 32, len(fileStore.uuid)) //UUID is 128bits (16bytes, but expanded to hex string so 32chars)
	_, err := os.Stat(fileStore.fileCache.absoluteDirpath)
	assert.Nil(t, err) //Folder path exists
	assert.Equal(t, "", fileStore.fileCache.dirpathRelativeToDataDirRoot)
	anotherFileStore := getTestFileStore(t)
	require.NotEqual(t, fileStore.uuid, anotherFileStore) //To check uuid is actually unique.
}

func TestFileStore_StoreFile(t *testing.T) {
	fileStore := getTestFileStore(t)
	testContent := "Long Live Kurtosis!"
	reader := strings.NewReader(testContent)
	enclaveFile, err := fileStore.StoreFile(reader)
	assert.Nil(t, err)
	assert.Equal(t, enclaveFile.absoluteFilepath, fileStore.GetFilePath())

	_, dirErr := os.Stat(enclaveFile.absoluteFilepath)
	assert.Nil(t, dirErr)
	file, readErr := ioutil.ReadFile(fileStore.GetFilePath())
	assert.Nil(t, readErr)
	assert.Equal(t, []byte(testContent), file)
}

func getTestFileStore(t *testing.T) *FileStore {
	absDirpath, err := ioutil.TempDir("", "")
	assert.Nil(t, err)
	fileStore, err := newFileStore(absDirpath, "")
	assert.Nil(t, err)
	return fileStore
}