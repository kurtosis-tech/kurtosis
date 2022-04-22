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
	"github.com/stretchr/testify/require"
)

func TestFileStore_StoreFileSavedFilesUniquely(t *testing.T) {
	fileStore := getTestFileStore(t)
	testContent := "Long Live Kurtosis!"
	reader := strings.NewReader(testContent)
	uuid, enclaveFile, err := fileStore.StoreFile(reader)
	assert.Nil(t, err)
	assert.Equal(t, 32, len(uuid)) //UUID is 128 bits but in string it is hex represented chars so 32 chars

	//Test that it saved where it said it would.
	_, dirErr := os.Stat(enclaveFile.absoluteFilepath)
	assert.Nil(t, dirErr)
	file, readErr := ioutil.ReadFile(enclaveFile.absoluteFilepath)
	assert.Nil(t, readErr)
	assert.Equal(t, []byte(testContent), file)

	//Test that it will save another file in the same place without clashing.
	testContent = "Long Live Kurtosis Another Again!"
	reader = strings.NewReader(testContent)
	anotherUUID, anotherEnclaveFile, anotherError := fileStore.StoreFile(reader)
	assert.Nil(t, anotherError)
	_, dirErr = os.Stat(anotherEnclaveFile.absoluteFilepath)
	assert.Nil(t, dirErr)
	file, readErr = ioutil.ReadFile(anotherEnclaveFile.absoluteFilepath)
	assert.Equal(t, []byte(testContent), file)

	require.NotEqual(t, uuid, anotherUUID)
	//Make sure somehow we didn't just write over it.
	require.NotEqual(t, enclaveFile.GetAbsoluteFilepath(), anotherEnclaveFile.GetAbsoluteFilepath())
}

func getTestFileStore(t *testing.T) *FileStore {
	absDirpath, err := ioutil.TempDir("", "")
	assert.Nil(t, err)
	fileStore, err := newFileStore(absDirpath, "")
	assert.Nil(t, err)
	return fileStore
}