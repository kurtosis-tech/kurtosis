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

func TestFileStore_StoreFileSimpleCase(t *testing.T) {
	fileStore := getTestFileStore(t)
	testContent := "Long Live Kurtosis!"
	reader := strings.NewReader(testContent)
	targetArtifactName := "test-artifact-name"
	filesArtifactUuid, err := fileStore.StoreFile(reader, targetArtifactName)
	require.Equal(t, 32, len(filesArtifactUuid)) //UUID is 128 bits but in string it is hex represented chars so 32 chars
	require.Nil(t, err)
	require.Len(t, fileStore.artifactNameToArtifactUuid, 1)
	require.Contains(t, fileStore.artifactNameToArtifactUuid, targetArtifactName)

	//Test that it saved where it said it would.
	expectedFilename := strings.Join(
		[]string{string(filesArtifactUuid), artifactExtension},
		".",
	)
	expectedFilepath := filepath.Join(fileStore.fileCache.absoluteDirpath, expectedFilename)
	_, dirErr := os.Stat(expectedFilepath)
	require.Nil(t, dirErr)
	file, readErr := ioutil.ReadFile(expectedFilepath)
	require.Nil(t, readErr)
	require.Equal(t, []byte(testContent), file)
}

func TestFileStore_StoringToExistingUUIDFails(t *testing.T) {
	fileStore := getTestFileStore(t)
	testContent := "Long Live Kurtosis!"
	reader := strings.NewReader(testContent)
	testArtifactName := "test-artifact-name"
	filesArtifactUuid, err := fileStore.StoreFile(reader, testArtifactName)
	require.Nil(t, err)
	require.Equal(t, 32, len(filesArtifactUuid)) //UUID is 128 bits but in string it is hex represented chars so 32 chars

	anotherTestContent := "This one should fail"
	anotherReader := strings.NewReader(anotherTestContent)
	_, err = fileStore.StoreFile(anotherReader, testArtifactName)
	require.NotNil(t, err)
}

func TestFileStore_GetFilepathByUUIDProperFilepath(t *testing.T) {
	fileStore := getTestFileStore(t)
	testContent := "Long Live Kurtosis!"
	testArtifactName := "test-artifact"
	reader := strings.NewReader(testContent)
	uuid, err := fileStore.StoreFile(reader, testArtifactName)
	require.Nil(t, err)

	enclaveDataFile, err := fileStore.GetFile(string(uuid))
	require.Nil(t, err)

	_, dirErr := os.Stat(enclaveDataFile.absoluteFilepath)
	require.Nil(t, dirErr)
	file, readErr := ioutil.ReadFile(enclaveDataFile.absoluteFilepath)
	require.Nil(t, readErr)
	require.Equal(t, []byte(testContent), file)
}

func TestFileStore_StoreFilesUniquely(t *testing.T) {
	fileStore := getTestFileStore(t)
	testContent := "Long Live Kurtosis!"
	otherTestContent := "Long Live Kurtosis, But Different!"

	//Write Both Files
	reader := strings.NewReader(testContent)
	testArtifact1 := "test-artifact-1"
	uuid, err := fileStore.StoreFile(reader, testArtifact1)
	require.Nil(t, err)

	reader = strings.NewReader(otherTestContent)
	testArtifact2 := "test-artifact-2"
	anotherUUID, err := fileStore.StoreFile(reader, testArtifact2)
	require.Nil(t, err)
	require.NotEqual(t, uuid, anotherUUID)

	//Get their paths.
	enclaveDataFile, err := fileStore.GetFile(string(uuid))
	require.Nil(t, err)
	anotherFilepath, err := fileStore.GetFile(string(anotherUUID))
	require.Nil(t, err)
	require.NotEqual(t, enclaveDataFile, anotherFilepath)

	//Read and evaluate their content is different.
	file, readErr := ioutil.ReadFile(enclaveDataFile.absoluteFilepath)
	require.Nil(t, readErr)
	anotherFile, readErr := ioutil.ReadFile(anotherFilepath.absoluteFilepath)
	require.Nil(t, readErr)
	require.NotEqual(t, file, anotherFile)
}

func TestFileStore_RemoveFileRemovesFileFromDisk(t *testing.T) {
	fileStore := getTestFileStore(t)
	testContent := "Long Live Kurtosis!"
	reader := strings.NewReader(testContent)
	testArtifactName := "test-artifact"
	uuid, err := fileStore.StoreFile(reader, testArtifactName)
	require.Nil(t, err)
	require.Len(t, fileStore.shortenedUuidToFullUuid, 1)
	require.Len(t, fileStore.artifactNameToArtifactUuid, 1)

	enclaveDataFile, err := fileStore.GetFile(string(uuid))
	require.Nil(t, err)

	err = fileStore.RemoveFile(string(uuid))
	require.Nil(t, err)
	require.Len(t, fileStore.shortenedUuidToFullUuid, 0)
	require.Len(t, fileStore.artifactNameToArtifactUuid, 0)

	_, err = os.Stat(enclaveDataFile.absoluteFilepath)
	require.NotNil(t, err)
	require.True(t, os.IsNotExist(err))
}

func TestFileStore_RemoveFileFailsForNonExistentId(t *testing.T) {
	fileStore := getTestFileStore(t)
	nonExistentId, err := NewFilesArtifactUUID()
	require.Nil(t, err)

	err = fileStore.RemoveFile(string(nonExistentId))
	require.NotNil(t, err)
}

func TestFilesArtifactStore_GetFileNamesAndUuids(t *testing.T) {
	fileStore := getTestFileStore(t)
	testContent := "Long Live Kurtosis!"
	otherTestContent := "Long Live Kurtosis, But Different!"

	//Write Both Files
	reader := strings.NewReader(testContent)
	testArtifact1 := "test-artifact-1"
	uuid, err := fileStore.StoreFile(reader, testArtifact1)
	require.Nil(t, err)

	reader = strings.NewReader(otherTestContent)
	testArtifact2 := "test-artifact-2"
	anotherUUID, err := fileStore.StoreFile(reader, testArtifact2)
	require.Nil(t, err)
	require.NotEqual(t, uuid, anotherUUID)

	fileNameAndUuids := fileStore.GetFileNamesAndUuids()
	require.Len(t, fileNameAndUuids, 2)
	require.Contains(t, fileNameAndUuids, FileNameAndUuid{uuid: uuid, name: testArtifact1})
	require.Contains(t, fileNameAndUuids, FileNameAndUuid{uuid: anotherUUID, name: testArtifact2})
}

func getTestFileStore(t *testing.T) *FilesArtifactStore {
	absDirpath, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	fileStore := newFilesArtifactStore(absDirpath, "")
	require.Nil(t, err)
	return fileStore
}

func Test_generateUniqueNameForFileArtifact_MaxRetriesOver(t *testing.T) {
	timesCalled := 0
	// this method should be call 4 time (maxRetries + 1)
	mockedGenerateNameMethod := func() string {
		timesCalled = timesCalled + 1
		if timesCalled == 4 {
			return "last-noun"
		}
		return "adjective-noun"
	}

	artifactIdToUUIDMap := map[string]FilesArtifactUUID{
		"adjective-noun": FilesArtifactUUID("a"),
		"last-noun":      FilesArtifactUUID("b"),
		"last-noun-1":    FilesArtifactUUID("b"),
	}

	fileArtifactStoreUnderTest := NewFilesArtifactStoreForTesting("/", "/", artifactIdToUUIDMap, nil, 3, mockedGenerateNameMethod)
	actual := fileArtifactStoreUnderTest.GenerateUniqueNameForFileArtifact()
	require.Equal(t, "last-noun-2", actual)
}

func Test_generateUniqueNameForFileArtifact_Found(t *testing.T) {
	timesCalled := 0
	mockedGenerateNameMethod := func() string {
		timesCalled = timesCalled + 1
		if timesCalled == 4 {
			return "unique-name"
		}
		return "non-unique-name"
	}

	artifactIdToUUIDMap := map[string]FilesArtifactUUID{
		"non-unique-name": FilesArtifactUUID("a"),
	}

	fileArtifactStoreUnderTest := NewFilesArtifactStoreForTesting("/", "/", artifactIdToUUIDMap, nil, 3, mockedGenerateNameMethod)
	actual := fileArtifactStoreUnderTest.GenerateUniqueNameForFileArtifact()
	require.Equal(t, "unique-name", actual)
}
