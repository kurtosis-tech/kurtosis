/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_directory

import (
	"github.com/kurtosis-tech/stacktrace"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
)

type failedReader struct {}

func (reader failedReader) Read(b []byte) (int, error) {
	return 0, stacktrace.Propagate(stacktrace.NewError("This is a test failure."),
		 "You are not supposed to see this test failure. Please contact developers if you are.")
}

func TestFileCache_AddAndGetAndRemoveArtifact(t *testing.T) {
	fileCache := getTestFileCache(t)
	testKey := "test-key"
	testContents := "test-file-contents"

	reader := strings.NewReader(testContents);
	addedFileObj, err := fileCache.AddFile(testKey, reader)
	assert.Nil(t, err)

	// Check that the contents are what we expect
	fp, err := os.Open(addedFileObj.GetAbsoluteFilepath())
	assert.Nil(t, err)
	testFileBytes, err := ioutil.ReadAll(fp)
	assert.Nil(t, err)
	assert.Equal(t, testContents, string(testFileBytes))

	// Check relative filepath was set correctly
	assert.Equal(t, addedFileObj.GetAbsoluteFilepath(), path.Join(fileCache.absoluteDirpath, addedFileObj.filepathRelativeToDataDirRoot))

	// Verify the retrieved file matches the file we just created
	retrievedFileObj, err := fileCache.GetFile(testKey)
	assert.Nil(t, err)
	assert.Equal(t, addedFileObj.GetAbsoluteFilepath(), retrievedFileObj.GetAbsoluteFilepath())
	assert.Equal(t, addedFileObj.GetFilepathRelativeToDataDirRoot(), retrievedFileObj.GetFilepathRelativeToDataDirRoot())

	// Verify that remove file works on the key and removes the file
	err = fileCache.RemoveFile(testKey)
	assert.Nil(t, err)
	_, err = os.Stat(addedFileObj.GetAbsoluteFilepath())
	assert.NotNil(t, err)
	assert.True(t, os.IsNotExist(err))
}

func TestFileCache_GetErrorsOnNonexistentKey(t *testing.T) {
	fileCache := getTestFileCache(t)

	_, err := fileCache.GetFile("nonexistent-key")
	assert.NotNil(t, err)
}

func TestFileCache_AddErrorsOnDuplicateAdd(t *testing.T) {
	fileCache := getTestFileCache(t)

	testKey := "test-key"
	_, err := fileCache.AddFile(testKey, strings.NewReader(""))
	assert.Nil(t, err)

	_, err = fileCache.AddFile(testKey, strings.NewReader(""))
	assert.NotNil(t, err)
}

func TestFileCache_FileDeletedOnReaderError(t *testing.T) {
	fileCache := getTestFileCache(t)
	var readerThatFails failedReader

	testKey := "test-keys"
	_, err := fileCache.AddFile(testKey, readerThatFails)
	assert.NotNil(t, err)

	// Make sure the file cache directory is still empty
	files, err := ioutil.ReadDir(fileCache.absoluteDirpath)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(files))
}

func TestFileCache_RemoveFileFailsForNonExistingKeys(t *testing.T) {
	fileCache := getTestFileCache(t)
	testKey := "test-key"

	err := fileCache.RemoveFile(testKey)
	assert.NotNil(t, err)
}

func getTestFileCache(t *testing.T) *FileCache {
	absDirpath, err := ioutil.TempDir("", "")
	assert.Nil(t, err)
	return newFileCache(absDirpath, "")
}
