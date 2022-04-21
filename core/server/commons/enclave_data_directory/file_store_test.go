/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_directory

import (
	"testing"
	"github.com/stretchr/testify/require"
	"strings"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
)

func TestFileStore_StoreFile(t *testing.T) {
	fileName := "TestFileStore_StoreFile.txt"
	data := []byte("Long live Kurtosis!")
	fileStore, err := StoreFile(fileName, data)
	require.NoError(t, err)
	defer os.Remove(fileStore.absolutePath)

	assert.Equal(t, 32, len(fileStore.uuid))
	expectedFileName := strings.Join([]string{fileStore.uuid, fileName}, "_")
	workingDir, _ := os.Getwd()
	expectedFilePath := filepath.Join(workingDir, expectedFileName)
	assert.Equal(t, expectedFilePath, fileStore.absolutePath)
	assert.Equal(t, "", fileStore.dirpathRelativeToDataDirRoot)
	assert.FileExists(t, expectedFilePath)
}