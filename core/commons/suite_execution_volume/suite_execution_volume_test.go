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

func TestGetSuiteMetadataFile(t *testing.T) {
	suiteExVolDirpath, err := ioutil.TempDir("", "")
	assert.Nil(t, err)

	suiteExVol := NewSuiteExecutionVolume(suiteExVolDirpath)
	metadataFile := suiteExVol.GetSuiteMetadataFile()

	assert.Equal(t, suiteMetadataFilename, metadataFile.GetFilepathRelativeToVolRoot())

	writeFp, err := os.Create(metadataFile.absoluteFilepath)
	assert.Nil(t, err)
	testStr := "this is a test string"
	writeFp.WriteString(testStr)
	assert.Nil(t, writeFp.Close())

	expectedFilepath := path.Join(suiteExVolDirpath, suiteMetadataFilename)
	readFp, err := os.Open(expectedFilepath)
	assert.Nil(t, err)
	fileBytes, err := ioutil.ReadAll(readFp)
	assert.Nil(t, err)

	assert.Equal(t, testStr, string(fileBytes))
}

func TestGetSuiteExecutionDirectory(t *testing.T) {
	suiteExVolDirpath, err := ioutil.TempDir("", "")
	assert.Nil(t, err)

	testId := "someTest"

	suiteExVol := NewSuiteExecutionVolume(suiteExVolDirpath)
	testExDir, err := suiteExVol.GetEnclaveDirectory(testId)
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
