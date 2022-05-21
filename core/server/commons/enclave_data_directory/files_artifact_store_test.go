/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_directory

import (
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileStore_StoreFileSavesFile(t *testing.T) {
	fileStore := getTestFileStore(t)
	testContent := "Long Live Kurtosis!"
	reader := strings.NewReader(testContent)
	filesArtifactId, err := fileStore.StoreFile(reader)
	require.Nil(t, err)
	require.Equal(t, 36, len(filesArtifactId)) //UUID is 128 bits but in string it is hex represented chars so 32 chars

	//Test that it saved where it said it would.
	expectedFilename := strings.Join(
		[]string{string(filesArtifactId), artifactExtension},
		".",
	)
	expectedFilepath := filepath.Join(fileStore.fileCache.absoluteDirpath, expectedFilename)
	_, dirErr := os.Stat(expectedFilepath)
	require.Nil(t, dirErr)
	file, readErr := ioutil.ReadFile(expectedFilepath)
	require.Nil(t, readErr)
	require.Equal(t, []byte(testContent), file)
}

func TestFileStore_GetFilepathByUUIDProperFilepath(t *testing.T) {
	fileStore := getTestFileStore(t)
	testContent := "Long Live Kurtosis!"
	reader := strings.NewReader(testContent)
	uuid, err := fileStore.StoreFile(reader)
	require.Nil(t, err)

	enclaveDataFile, err := fileStore.GetFile(uuid)
	require.Nil(t, err)

	_, dirErr := os.Stat(enclaveDataFile.absoluteFilepath)
	require.Nil(t, dirErr)
	file, readErr := ioutil.ReadFile(enclaveDataFile.absoluteFilepath)
	require.Nil(t, readErr)
	require.Equal(t, []byte(testContent), file)
}

func TestFileStore_StoreFilesUniquely(t *testing.T){
	fileStore := getTestFileStore(t)
	testContent := "Long Live Kurtosis!"
	otherTestContent := "Long Live Kurtosis, But Different!"

	//Write Both Files
	reader := strings.NewReader(testContent)
	uuid, err := fileStore.StoreFile(reader)
	require.Nil(t, err)

	reader = strings.NewReader(otherTestContent)
	anotherUUID, err := fileStore.StoreFile(reader)
	require.Nil(t, err)
	require.NotEqual(t, uuid, anotherUUID)

	//Get their paths.
	enclaveDataFile, err := fileStore.GetFile(uuid)
	require.Nil(t, err)
	anotherFilepath, err := fileStore.GetFile(anotherUUID)
	require.Nil(t, err)
	require.NotEqual(t, enclaveDataFile, anotherFilepath)

	//Read and evaluate their content is different.
	file, readErr := ioutil.ReadFile(enclaveDataFile.absoluteFilepath)
	require.Nil(t, readErr)
	anotherFile, readErr := ioutil.ReadFile(anotherFilepath.absoluteFilepath)
	require.Nil(t, readErr)
	require.NotEqual(t, file, anotherFile)
}

func getTestFileStore(t *testing.T) *FilesArtifactStore {
	absDirpath, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	fileStore := newFilesArtifactStore(absDirpath, "")
	require.Nil(t, err)
	return fileStore
}