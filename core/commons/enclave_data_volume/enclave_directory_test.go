/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package enclave_data_volume

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
)

func TestNewServiceDirectory(t *testing.T) {
	enclaveDirpath, err := ioutil.TempDir("", "")
	assert.Nil(t, err)

	enclaveDir := NewEnclaveDirectory(enclaveDirpath)

	testServiceId := "test-service"
	svcDir, err := enclaveDir.NewServiceDirectory(testServiceId)
	assert.Nil(t, err)

	// Test that all-services dir got created
	allSvcsDirpath := path.Join(enclaveDirpath, allServicesDirname)
	_, err = os.Stat(allSvcsDirpath)
	assert.Nil(t, err)

	files, err := ioutil.ReadDir(allSvcsDirpath)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(files))

	createdDir := files[0]
	// We add a UUID to each service so they don't conflict, so the name won't match exactly
	// This is why we use "contains"
	strings.Contains(createdDir.Name(), testServiceId)

	absoluteSvcDirpath := svcDir.absoluteDirpath
	assert.Equal(t, allSvcsDirpath, path.Dir(absoluteSvcDirpath))
	assert.True(t, strings.Contains(absoluteSvcDirpath, testServiceId))

	relativeSvcDirpath := svcDir.dirpathRelativeToVolRoot
	assert.Equal(t, path.Join(allServicesDirname), path.Dir(relativeSvcDirpath))
	assert.True(t, strings.Contains(relativeSvcDirpath, testServiceId))
}

func TestGetArtifactCache(t *testing.T) {
	enclaveDirpath, err := ioutil.TempDir("", "")
	assert.Nil(t, err)

	enclaveDir := NewEnclaveDirectory(enclaveDirpath)

	artifactCache, err := enclaveDir.GetArtifactCache()
	assert.Nil(t, err)

	expectedAbsDirpath := path.Join(enclaveDirpath, artifactCacheDirname)
	_, err = os.Stat(expectedAbsDirpath)
	assert.Nil(t, err)
	assert.Equal(t, expectedAbsDirpath, artifactCache.absoluteDirpath)

	expectedRelativeDirpath := artifactCacheDirname
	assert.Equal(t, expectedRelativeDirpath, artifactCache.dirpathRelativeToVolRoot)

}

func TestGetStaticFileCache(t *testing.T) {
	enclaveDirpath, err := ioutil.TempDir("", "")
	assert.Nil(t, err)

	enclaveDir := NewEnclaveDirectory(enclaveDirpath)

	staticFileCache, err := enclaveDir.GetStaticFileCache()
	assert.Nil(t, err)

	expectedAbsDirpath := path.Join(enclaveDirpath, staticFileCacheDirname)
	_, err = os.Stat(expectedAbsDirpath)
	assert.Nil(t, err)
	assert.Equal(t, expectedAbsDirpath, staticFileCache.absoluteDirpath)

	expectedRelativeDirpath := staticFileCacheDirname
	assert.Equal(t, expectedRelativeDirpath, staticFileCache.dirpathRelativeToVolRoot)

}
