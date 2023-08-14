/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_directory

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db/file_artifacts_db"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/test_helpers"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileStore_StoreFileSimpleCase(t *testing.T) {
	fileStore, closer := getTestFileStore(t)
	defer closer()
	testContent := "Long Live Kurtosis!"
	reader := strings.NewReader(testContent)
	targetArtifactName := "test-artifact-name"
	fakeMd5 := []byte("blah")
	filesArtifactUuid, err := fileStore.StoreFile(reader, fakeMd5, targetArtifactName)
	require.Equal(t, 32, len(filesArtifactUuid)) //UUID is 128 bits but in string it is hex represented chars so 32 chars
	require.Nil(t, err)
	require.Len(t, fileStore.fileArtifactDb.GetArtifactUuidMap(), 1)
	require.Contains(t, fileStore.fileArtifactDb.GetArtifactUuidMap(), targetArtifactName)
	require.Len(t, fileStore.fileArtifactDb.GetContentMd5Map(), 1)
	require.Contains(t, fileStore.fileArtifactDb.GetContentMd5Map(), string(filesArtifactUuid))

	//Test that it saved where it said it would.
	expectedFilename := strings.Join(
		[]string{string(filesArtifactUuid), artifactExtension},
		".",
	)
	expectedFilepath := filepath.Join(fileStore.fileCache.absoluteDirpath, expectedFilename)
	_, dirErr := os.Stat(expectedFilepath)
	require.Nil(t, dirErr)
	file, readErr := os.ReadFile(expectedFilepath)
	require.Nil(t, readErr)
	require.Equal(t, []byte(testContent), file)
}

func TestFileStore_StoringToExistingUUIDFails(t *testing.T) {
	fileStore, closer := getTestFileStore(t)
	defer closer()
	testContent := "Long Live Kurtosis!"
	reader := strings.NewReader(testContent)
	testArtifactName := "test-artifact-name"
	fakeMd5 := []byte("blah")
	filesArtifactUuid, err := fileStore.StoreFile(reader, fakeMd5, testArtifactName)
	require.Nil(t, err)
	require.Equal(t, 32, len(filesArtifactUuid)) //UUID is 128 bits but in string it is hex represented chars so 32 chars

	anotherTestContent := "This one should fail"
	anotherReader := strings.NewReader(anotherTestContent)
	_, err = fileStore.StoreFile(anotherReader, fakeMd5, testArtifactName)
	require.NotNil(t, err)
}

func TestFileStore_GetFilepathByUUIDProperFilepath(t *testing.T) {
	fileStore, closer := getTestFileStore(t)
	defer closer()
	testContent := "Long Live Kurtosis!"
	testArtifactName := "test-artifact"
	reader := strings.NewReader(testContent)
	fakeMd5 := []byte("blah")
	uuid, err := fileStore.StoreFile(reader, fakeMd5, testArtifactName)
	require.Nil(t, err)

	returnedUuid, enclaveDataFile, returnedMd5, found, err := fileStore.GetFile(string(uuid))
	require.Nil(t, err)
	require.True(t, found)
	require.Equal(t, uuid, returnedUuid)
	require.Equal(t, fakeMd5, returnedMd5)

	_, dirErr := os.Stat(enclaveDataFile.absoluteFilepath)
	require.Nil(t, dirErr)
	file, readErr := os.ReadFile(enclaveDataFile.absoluteFilepath)
	require.Nil(t, readErr)
	require.Equal(t, []byte(testContent), file)
}

func TestFileStore_StoreFilesUniquely(t *testing.T) {
	fileStore, closer := getTestFileStore(t)
	defer closer()
	testContent := "Long Live Kurtosis!"
	otherTestContent := "Long Live Kurtosis, But Different!"

	//Write Both Files
	reader := strings.NewReader(testContent)
	testArtifact1 := "test-artifact-1"
	fakeMd5 := []byte("blah")
	uuid, err := fileStore.StoreFile(reader, fakeMd5, testArtifact1)
	require.Nil(t, err)

	reader = strings.NewReader(otherTestContent)
	testArtifact2 := "test-artifact-2"
	fakeMd52 := []byte("blah2")
	anotherUUID, err := fileStore.StoreFile(reader, fakeMd52, testArtifact2)
	require.Nil(t, err)
	require.NotEqual(t, uuid, anotherUUID)

	//Get their paths.
	_, enclaveDataFile, _, found, err := fileStore.GetFile(string(uuid))
	require.Nil(t, err)
	require.True(t, found)
	_, anotherFilepath, _, anotherFound, err := fileStore.GetFile(string(anotherUUID))
	require.Nil(t, err)
	require.True(t, anotherFound)
	require.NotEqual(t, enclaveDataFile, anotherFilepath)

	//Read and evaluate their content is different.
	file, readErr := os.ReadFile(enclaveDataFile.absoluteFilepath)
	require.Nil(t, readErr)
	anotherFile, readErr := os.ReadFile(anotherFilepath.absoluteFilepath)
	require.Nil(t, readErr)
	require.NotEqual(t, file, anotherFile)
}

func TestFileStore_RemoveFileRemovesFileFromDisk(t *testing.T) {
	fileStore, closer := getTestFileStore(t)
	defer closer()
	testContent := "Long Live Kurtosis!"
	reader := strings.NewReader(testContent)
	testArtifactName := "test-artifact"
	fakeMd5 := []byte("blah")
	uuid, err := fileStore.StoreFile(reader, fakeMd5, testArtifactName)
	require.Nil(t, err)
	require.Len(t, fileStore.fileArtifactDb.GetFullUuidMap(), 1)
	require.Len(t, fileStore.fileArtifactDb.GetArtifactUuidMap(), 1)

	_, enclaveDataFile, _, found, err := fileStore.GetFile(string(uuid))
	require.Nil(t, err)
	require.True(t, found)

	err = fileStore.RemoveFile(string(uuid))
	require.Nil(t, err)
	require.Len(t, fileStore.fileArtifactDb.GetFullUuidMap(), 0)
	require.Len(t, fileStore.fileArtifactDb.GetArtifactUuidMap(), 0)
	require.Len(t, fileStore.fileArtifactDb.GetContentMd5Map(), 0)

	_, err = os.Stat(enclaveDataFile.absoluteFilepath)
	require.NotNil(t, err)
	require.True(t, os.IsNotExist(err))
}

func TestFileStore_RemoveFileFailsForNonExistentId(t *testing.T) {
	fileStore, closer := getTestFileStore(t)
	defer closer()
	nonExistentId, err := NewFilesArtifactUUID()
	require.Nil(t, err)

	err = fileStore.RemoveFile(string(nonExistentId))
	require.NotNil(t, err)
}

func TestFilesArtifactStore_GetFileNamesAndUuids(t *testing.T) {
	fileStore, closer := getTestFileStore(t)
	defer closer()
	testContent := "Long Live Kurtosis!"
	otherTestContent := "Long Live Kurtosis, But Different!"

	//Write Both Files
	reader := strings.NewReader(testContent)
	testArtifact1 := "test-artifact-1"
	fakeMd5 := []byte("blah")
	uuid, err := fileStore.StoreFile(reader, fakeMd5, testArtifact1)
	require.Nil(t, err)

	reader = strings.NewReader(otherTestContent)
	testArtifact2 := "test-artifact-2"
	fakeMd52 := []byte("blah")
	anotherUUID, err := fileStore.StoreFile(reader, fakeMd52, testArtifact2)
	require.Nil(t, err)
	require.NotEqual(t, uuid, anotherUUID)

	fileNameAndUuids := fileStore.GetFileNamesAndUuids()
	require.Len(t, fileNameAndUuids, 2)
	require.Contains(t, fileNameAndUuids, FileNameAndUuid{uuid: uuid, name: testArtifact1})
	require.Contains(t, fileNameAndUuids, FileNameAndUuid{uuid: anotherUUID, name: testArtifact2})
}

func getTestFileStore(t *testing.T) (*FilesArtifactStore, func()) {
	absDirpath, err := os.MkdirTemp("", "")
	require.Nil(t, err)
	db, closer, err := test_helpers.CreateEnclaveDbForTesting()
	require.Nil(t, err)
	fileArtifactDb, err := file_artifacts_db.GetFileArtifactsDbForTesting(db, map[string]string{})
	require.Nil(t, err)
	fileStore := newFilesArtifactStoreFromDb(absDirpath, "", fileArtifactDb)
	require.Nil(t, err)
	return fileStore, closer
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

	artifactIdToUUIDMap := map[string]string{
		"adjective-noun": "a",
		"last-noun":      "b",
		"last-noun-1":    "b",
	}

	db, closer, err := test_helpers.CreateEnclaveDbForTesting()
	require.Nil(t, err)
	defer closer()
	fileArtifactDb, err := file_artifacts_db.GetFileArtifactsDbForTesting(db, artifactIdToUUIDMap)
	require.Nil(t, err)

	fileArtifactStoreUnderTest := NewFilesArtifactStoreForTesting("/", "/", fileArtifactDb, 3, mockedGenerateNameMethod)
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

	artifactIdToUUIDMap := map[string]string{
		"non-unique-name": "a",
	}

	db, closer, err := test_helpers.CreateEnclaveDbForTesting()
	require.Nil(t, err)
	defer closer()
	fileArtifactDb, err := file_artifacts_db.GetFileArtifactsDbForTesting(db, artifactIdToUUIDMap)
	require.Nil(t, err)

	fileArtifactStoreUnderTest := NewFilesArtifactStoreForTesting("/", "/", fileArtifactDb, 3, mockedGenerateNameMethod)
	actual := fileArtifactStoreUnderTest.GenerateUniqueNameForFileArtifact()
	require.Equal(t, "unique-name", actual)
}
