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
	"strings"
	"path/filepath"
	"github.com/stretchr/testify/require"
)

func TestFileStore_StoreFileSavesFile(t *testing.T) {
	fileStore := getTestFileStore(t)
	testContent := "Long Live Kurtosis!"
	reader := strings.NewReader(testContent)
	uuid, err := fileStore.StoreFile(reader)
	assert.Nil(t, err)
	assert.Equal(t, 32, len(uuid)) //UUID is 128 bits but in string it is hex represented chars so 32 chars

	//Test that it saved where it said it would.
	expectedFilename := strings.Join([]string{uuid, artifactExtension}, ".")
	expectedFilepath := filepath.Join(fileStore.fileCache.absoluteDirpath, expectedFilename)
	_, dirErr := os.Stat(expectedFilepath)
	assert.Nil(t, dirErr)
	file, readErr := ioutil.ReadFile(expectedFilepath)
	assert.Nil(t, readErr)
	assert.Equal(t, []byte(testContent), file)
}

func TestFileStore_GetFilepathByUUIDProperFilepath(t *testing.T) {
	fileStore := getTestFileStore(t)
	testContent := "Long Live Kurtosis!"
	reader := strings.NewReader(testContent)
	uuid, err := fileStore.StoreFile(reader)
	assert.Nil(t, err)

	filepath, err := fileStore.GetFilepathByUUID(uuid)
	assert.Nil(t, err)

	_, dirErr := os.Stat(filepath)
	assert.Nil(t, dirErr)
	file, readErr := ioutil.ReadFile(filepath)
	assert.Nil(t, readErr)
	assert.Equal(t, []byte(testContent), file)
}

func TestFileStore_StoreFilesUniquely(t *testing.T){
	fileStore := getTestFileStore(t)
	testContent := "Long Live Kurtosis!"
	otherTestContent := "Long Live Kurtosis, But Different!"

	//Write Both Files
	reader := strings.NewReader(testContent)
	uuid, err := fileStore.StoreFile(reader)
	assert.Nil(t, err)

	reader = strings.NewReader(otherTestContent)
	anotherUUID, err := fileStore.StoreFile(reader)
	assert.Nil(t, err)
	require.NotEqual(t, uuid, anotherUUID)

	//Get their paths.
	filepath, err := fileStore.GetFilepathByUUID(uuid)
	assert.Nil(t, err)
	anotherFilepath, err := fileStore.GetFilepathByUUID(anotherUUID)
	assert.Nil(t, err)
	require.NotEqual(t, filepath, anotherFilepath)

	//Read and evaluate their content is different.
	file, readErr := ioutil.ReadFile(filepath)
	assert.Nil(t, readErr)
	anotherFile, readErr := ioutil.ReadFile(anotherFilepath)
	assert.Nil(t, readErr)
	require.NotEqual(t, file, anotherFile)
}

func getTestFileStore(t *testing.T) *FilesArtifactStore {
	absDirpath, err := ioutil.TempDir("", "")
	assert.Nil(t, err)
	fileStore := newFilesArtifactStore(absDirpath, "")
	assert.Nil(t, err)
	return fileStore
}