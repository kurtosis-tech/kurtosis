/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package enclave_data_volume

import (
	"github.com/palantir/stacktrace"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestFileCache_AddAndGetArtifact(t *testing.T) {
	fileCache := getTestFileCache(t)

	testKey := "test-key"
	testContents := "test-file-contents"
	addedFileObj, err := fileCache.AddFile(
		testKey,
		func(fp *os.File) error {
			if _, err := fp.Write([]byte(testContents)); err != nil {
				return stacktrace.Propagate(err, "An error occurred writing the test file contents")
			}
			return nil
		},
	)
	assert.Nil(t, err)

	// Check that the contents are what we expect
	fp, err := os.Open(addedFileObj.GetAbsoluteFilepath())
	assert.Nil(t, err)
	testFileBytes, err := ioutil.ReadAll(fp)
	assert.Nil(t, err)
	assert.Equal(t, testContents, string(testFileBytes))

	// Check relative filepath was set correctly
	assert.Equal(t, addedFileObj.GetAbsoluteFilepath(), path.Join(fileCache.absoluteDirpath, addedFileObj.filepathRelativeToVolRoot))

	// Verify the retrieved file matches the file we just created
	retrievedFileObj, err := fileCache.GetFile(testKey)
	assert.Nil(t, err)
	assert.Equal(t, addedFileObj.GetAbsoluteFilepath(), retrievedFileObj.GetAbsoluteFilepath())
	assert.Equal(t, addedFileObj.GetFilepathRelativeToVolRoot(), retrievedFileObj.GetFilepathRelativeToVolRoot())
}

func TestFileCache_GetErrorsOnNonexistentKey(t *testing.T) {
	fileCache := getTestFileCache(t)

	_, err := fileCache.GetFile("nonexistent-key")
	assert.NotNil(t, err)
}

func TestFileCache_AddErrorsOnDuplicateAdd(t *testing.T) {
	fileCache := getTestFileCache(t)

	testKey := "test-key"
	_, err := fileCache.AddFile(testKey, func(fp *os.File) error { return nil })
	assert.Nil(t, err)

	_, err = fileCache.AddFile(testKey, func(fp *os.File) error { return nil })
	assert.NotNil(t, err)
}

func TestFileCache_FileDeletedOnSupplierError(t *testing.T) {
	fileCache := getTestFileCache(t)

	testKey := "test-key"
	_, err := fileCache.AddFile(testKey, func(fp *os.File) error { return stacktrace.NewError("TEST ERROR") })
	assert.NotNil(t, err)

	// Make sure the file cache directory is still empty
	files, err := ioutil.ReadDir(fileCache.absoluteDirpath)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(files))
}

func getTestFileCache(t *testing.T) *FileCache {
	absDirpath, err := ioutil.TempDir("", "")
	assert.Nil(t, err)
	return newFileCache(absDirpath, "")
}
