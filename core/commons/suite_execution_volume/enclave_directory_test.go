/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package suite_execution_volume

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
)

func TestNewServiceDirectory(t *testing.T) {
	suiteExVolDirpath, err := ioutil.TempDir("", "")
	assert.Nil(t, err)

	testId := "someTest"

	suiteExVol := NewSuiteExecutionVolume(suiteExVolDirpath)
	testExDir, err := suiteExVol.GetEnclaveDirectory([]string{testId})
	assert.Nil(t, err)

	serviceId := "someService"

	svcDir, err := testExDir.NewServiceDirectory(serviceId)
	assert.Nil(t, err)

	allSvcsDirpath := path.Join(suiteExVolDirpath, testId, allServicesDirname)
	_, err = os.Stat(allSvcsDirpath)
	assert.Nil(t, err)

	files, err := ioutil.ReadDir(allSvcsDirpath)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(files))

	createdDir := files[0]
	// We add a UUID to each service so they don't conflict, so the name won't match exactly
	// This is why we use "contains"
	strings.Contains(createdDir.Name(), serviceId)

	absoluteSvcDirpath := svcDir.absoluteDirpath
	assert.Equal(t, allSvcsDirpath, path.Dir(absoluteSvcDirpath))
	assert.True(t, strings.Contains(absoluteSvcDirpath, serviceId))

	relativeSvcDirpath := svcDir.dirpathRelativeToVolRoot
	assert.Equal(t, path.Join(testId, allServicesDirname), path.Dir(relativeSvcDirpath))
	assert.True(t, strings.Contains(relativeSvcDirpath, serviceId))
}

func TestGetArtifactCache(t *testing.T) {
	suiteExVolDirpath, err := ioutil.TempDir("", "")
	assert.Nil(t, err)

	testId := "someTest"

	suiteExVol := NewSuiteExecutionVolume(suiteExVolDirpath)
	testExDir, err := suiteExVol.GetEnclaveDirectory([]string{testId})
	assert.Nil(t, err)

	artifactCache, err := testExDir.GetArtifactCache()
	assert.Nil(t, err)

	expectedAbsDirpath := path.Join(suiteExVolDirpath, testId, artifactCacheDirname)
	_, err = os.Stat(expectedAbsDirpath)
	assert.Nil(t, err)
	assert.Equal(t, expectedAbsDirpath, artifactCache.absoluteDirpath)

	expectedRelativeDirpath := path.Join(testId, artifactCacheDirname)
	assert.Equal(t, expectedRelativeDirpath, artifactCache.dirpathRelativeToVolRoot)

}

func TestGetStaticFileCache(t *testing.T) {
	suiteExVolDirpath, err := ioutil.TempDir("", "")
	assert.Nil(t, err)

	testId := "someTest"

	suiteExVol := NewSuiteExecutionVolume(suiteExVolDirpath)
	testExDir, err := suiteExVol.GetEnclaveDirectory([]string{testId})
	assert.Nil(t, err)

	staticFileCache, err := testExDir.GetStaticFileCache()
	assert.Nil(t, err)

	expectedAbsDirpath := path.Join(suiteExVolDirpath, testId, staticFileCacheDirname)
	_, err = os.Stat(expectedAbsDirpath)
	assert.Nil(t, err)
	assert.Equal(t, expectedAbsDirpath, staticFileCache.absoluteDirpath)

	expectedRelativeDirpath := path.Join(testId, staticFileCacheDirname)
	assert.Equal(t, expectedRelativeDirpath, staticFileCache.dirpathRelativeToVolRoot)

}
