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

func TestEnsureDirectoryExists(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "")
	assert.Nil(t, err)

	dirToCreate := path.Join(tempDir, "to-create")
	if _, err := os.Stat(dirToCreate); err == nil {
		t.Fatal("Expected directory not to exist")
	}

	assert.Nil(t, ensureDirpathExists(dirToCreate))

	if _, err := os.Stat(dirToCreate); err != nil {
		t.Fatal(t, err)
	}
}
