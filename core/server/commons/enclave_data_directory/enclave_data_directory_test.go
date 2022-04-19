/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_directory

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
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

	enclaveDir := NewEnclaveDataDirectory(enclaveDirpath)

	testServiceGUID := service.ServiceGUID("test-service")
	svcDir, err := enclaveDir.GetServiceDirectory(testServiceGUID)
	assert.Nil(t, err)

	// Test that all-services dir got created
	allSvcsDirpath := path.Join(enclaveDirpath, allServicesDirname)
	_, err = os.Stat(allSvcsDirpath)
	assert.Nil(t, err)

	files, err := ioutil.ReadDir(allSvcsDirpath)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(files))

	createdDir := files[0]

	assert.Equal(t, createdDir.Name(),string(testServiceGUID))

	absoluteSvcDirpath := svcDir.absoluteDirpath
	assert.Equal(t, allSvcsDirpath, path.Dir(absoluteSvcDirpath))
	assert.True(t, strings.Contains(absoluteSvcDirpath, string(testServiceGUID)))

	relativeSvcDirpath := svcDir.dirpathRelativeToDataDirRoot
	assert.Equal(t, path.Join(allServicesDirname), path.Dir(relativeSvcDirpath))
	assert.True(t, strings.Contains(relativeSvcDirpath, string(testServiceGUID)))
}

func TestGetArtifactCache(t *testing.T) {
	enclaveDirpath, err := ioutil.TempDir("", "")
	assert.Nil(t, err)

	enclaveDir := NewEnclaveDataDirectory(enclaveDirpath)

	artifactCache, err := enclaveDir.GetFilesArtifactCache()
	assert.Nil(t, err)

	expectedAbsDirpath := path.Join(enclaveDirpath, artifactCacheDirname)
	_, err = os.Stat(expectedAbsDirpath)
	assert.Nil(t, err)
	assert.Equal(t, expectedAbsDirpath, artifactCache.underlying.absoluteDirpath)

	expectedRelativeDirpath := artifactCacheDirname
	assert.Equal(t, expectedRelativeDirpath, artifactCache.underlying.dirpathRelativeToDataDirRoot)

}

func TestGetStaticFileCache(t *testing.T) {
	enclaveDirpath, err := ioutil.TempDir("", "")
	assert.Nil(t, err)

	enclaveDir := NewEnclaveDataDirectory(enclaveDirpath)

	staticFileCache, err := enclaveDir.GetStaticFileCache()
	assert.Nil(t, err)

	expectedAbsDirpath := path.Join(enclaveDirpath, staticFileCacheDirname)
	_, err = os.Stat(expectedAbsDirpath)
	assert.Nil(t, err)
	assert.Equal(t, expectedAbsDirpath, staticFileCache.underlying.absoluteDirpath)

	expectedRelativeDirpath := staticFileCacheDirname
	assert.Equal(t, expectedRelativeDirpath, staticFileCache.underlying.dirpathRelativeToDataDirRoot)

}
