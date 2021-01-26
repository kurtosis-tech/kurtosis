/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package suite_execution_volume

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"path"
	"strings"
	"testing"
)

func TestCreateFile(t *testing.T) {
	suiteExVolDirpath, err := ioutil.TempDir("", "")
	assert.Nil(t, err)

	testId := "someTest"

	suiteExVol := NewSuiteExecutionVolume(suiteExVolDirpath)
	testExDir, err := suiteExVol.CreateTestExecutionDirectory(testId)
	assert.Nil(t, err)

	serviceId := "someService"

	svcDir, err := testExDir.CreateServiceDirectory(serviceId)
	assert.Nil(t, err)

	svcAbsDirpath := svcDir.absoluteDirpath
	svcRelDirpath := svcDir.dirpathRelativeToVolRoot

	filename := "someFile"

	file, err := svcDir.CreateFile(filename)
	assert.Nil(t, err)

	// Check file was actually created
	files, err := ioutil.ReadDir(svcAbsDirpath)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(files))
	fileInfo := files[0]
	assert.True(t, strings.Contains(fileInfo.Name(), filename))

	// Check File data structure is correct
	assert.Equal(
		t,
		svcAbsDirpath,
		path.Dir(file.absoluteFilepath),
	)
	assert.Equal(
		t,
		svcRelDirpath,
		path.Dir(file.filepathRelativeToVolRoot),
	)
}