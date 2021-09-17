/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_volume

import (
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network/service_network_types"
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

	enclaveDir := NewEnclaveDataVolume(enclaveDirpath)

	testServiceGUID := service_network_types.ServiceGUID("test-service")
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

	relativeSvcDirpath := svcDir.dirpathRelativeToVolRoot
	assert.Equal(t, path.Join(allServicesDirname), path.Dir(relativeSvcDirpath))
	assert.True(t, strings.Contains(relativeSvcDirpath, string(testServiceGUID)))
}

func TestGetArtifactCache(t *testing.T) {
	enclaveDirpath, err := ioutil.TempDir("", "")
	assert.Nil(t, err)

	enclaveDir := NewEnclaveDataVolume(enclaveDirpath)

	artifactCache, err := enclaveDir.GetFilesArtifactCache()
	assert.Nil(t, err)

	expectedAbsDirpath := path.Join(enclaveDirpath, artifactCacheDirname)
	_, err = os.Stat(expectedAbsDirpath)
	assert.Nil(t, err)
	assert.Equal(t, expectedAbsDirpath, artifactCache.underlying.absoluteDirpath)

	expectedRelativeDirpath := artifactCacheDirname
	assert.Equal(t, expectedRelativeDirpath, artifactCache.underlying.dirpathRelativeToVolRoot)

}

func TestGetStaticFileCache(t *testing.T) {
	enclaveDirpath, err := ioutil.TempDir("", "")
	assert.Nil(t, err)

	enclaveDir := NewEnclaveDataVolume(enclaveDirpath)

	staticFileCache, err := enclaveDir.GetStaticFileCache()
	assert.Nil(t, err)

	expectedAbsDirpath := path.Join(enclaveDirpath, staticFileCacheDirname)
	_, err = os.Stat(expectedAbsDirpath)
	assert.Nil(t, err)
	assert.Equal(t, expectedAbsDirpath, staticFileCache.underlying.absoluteDirpath)

	expectedRelativeDirpath := staticFileCacheDirname
	assert.Equal(t, expectedRelativeDirpath, staticFileCache.underlying.dirpathRelativeToVolRoot)

}
