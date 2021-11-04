/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_directory

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func TestFilesArtifactCache_SuccessfulDownload(t *testing.T) {
	cache := getTestFilesArtifactCache(t)

	testArtifactId := "test-artifact"
	url := "https://www.google.com"
	err := cache.DownloadFilesArtifact(testArtifactId, url)
	assert.Nil(t, err)
}

func TestDownloadArtifactToFilepath_ErrorDownload(t *testing.T) {
	cache := getTestFilesArtifactCache(t)

	testArtifactId := "test-artifact"
	url := "THIS URL DOESN'T EXIST"
	err := cache.DownloadFilesArtifact(testArtifactId, url)
	assert.NotNil(t, err)
}
func getTestFilesArtifactCache(t *testing.T) *FilesArtifactCache {
	absDirpath, err := ioutil.TempDir("", "")
	assert.Nil(t, err)
	return newFilesArtifactCache(absDirpath, "")
}
