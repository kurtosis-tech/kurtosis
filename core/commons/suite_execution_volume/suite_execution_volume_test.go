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
	"testing"
)

func TestGetSuiteExecutionDirectory(t *testing.T) {
	suiteExVolDirpath, err := ioutil.TempDir("", "")
	assert.Nil(t, err)

	testId := "someTest"

	suiteExVol := NewSuiteExecutionVolume(suiteExVolDirpath)
	testExDir, err := suiteExVol.GetEnclaveDirectory([]string{testId})
	assert.Nil(t, err)

	expectedAbsDirpath := path.Join(suiteExVolDirpath, testId)
	_, err = os.Stat(expectedAbsDirpath)
	assert.Nil(t, err)

	assert.Equal(t, expectedAbsDirpath, testExDir.absoluteDirpath)
	assert.Equal(t, testId, testExDir.dirpathRelativeToVolRoot)
}

func TestGetArtifactCache(t *testing.T) {
	suiteExVolDirpath, err := ioutil.TempDir("", "")
	assert.Nil(t, err)

	suiteExVol := NewSuiteExecutionVolume(suiteExVolDirpath)
	artifactCache, err := suiteExVol.GetArtifactCache()
	assert.Nil(t, err)

	expectedAbsDirpath := path.Join(suiteExVolDirpath, artifactCacheDirname)
	_, err = os.Stat(expectedAbsDirpath)
	assert.Nil(t, err)

	assert.Equal(t, expectedAbsDirpath, artifactCache.absoluteDirpath)
	assert.Equal(t, artifactCacheDirname, artifactCache.dirpathRelativeToVolRoot)
}
