/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_directory

import (
	"testing"
	"os"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"github.com/stretchr/testify/require"
)

func TestFileHandler_ChangeDirectory(t *testing.T) {
	startDirectory, _ := os.Getwd()
	testHigherDirectory := "../../gopher"
	handler := createFileHandler(startDirectory)
	handler.ChangeDirectory(testHigherDirectory)
	assert.Equal(t, filepath.Join(startDirectory, testHigherDirectory), handler.currentWorkingDirectory)

	//Absolute directory change test.
	handler.ChangeDirectory(startDirectory)
	assert.Equal(t, startDirectory, handler.currentWorkingDirectory)

	testThisDirectory := "./gopher"
	handler.ChangeDirectory(testThisDirectory)
	assert.Equal(t, filepath.Join(startDirectory, testThisDirectory), handler.currentWorkingDirectory)
}

func TestFileHandler_SaveBytesToPath(t *testing.T) {
	startDirectory, _ := os.Getwd()
	fileName := "TestFileHandler_SaveBytesToPath.txt"
	content := []byte("Long live Kurtosis.")
	handler := createFileHandler(startDirectory)
	err := handler.SaveBytesToPath(fileName, "", content)
	require.NoError(t, err)

	fileLocation := filepath.Join(startDirectory, fileName)
	defer os.Remove(fileLocation)
	assert.FileExists(t, fileLocation)

	fileContents,_ := os.ReadFile(fileLocation)
	assert.Equal(t, content, fileContents)
}

func createFileHandler(workingDir string) *FileHandler{
	return newFileHandler(workingDir)
}
