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

func TestGetEnclaveDirectory(t *testing.T) {
	suiteExVolDirpath, err := ioutil.TempDir("", "")
	assert.Nil(t, err)

	testId := "someTest"

	suiteExVol := NewSuiteExecutionVolume(suiteExVolDirpath)
	enclaveDir, err := suiteExVol.GetEnclaveDirectory([]string{testId})
	assert.Nil(t, err)

	expectedAbsDirpath := path.Join(suiteExVolDirpath, testId)
	_, err = os.Stat(expectedAbsDirpath)
	assert.Nil(t, err)

	assert.Equal(t, expectedAbsDirpath, enclaveDir.absoluteDirpath)
	assert.Equal(t, testId, enclaveDir.dirpathRelativeToVolRoot)
}
