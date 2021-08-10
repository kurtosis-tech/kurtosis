/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_volume

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestStaticFileCache_TestAddFile(t *testing.T) {
	tmpDirpath, err := ioutil.TempDir("", "")
	assert.Nil(t, err)
	cache := newStaticFileCache(tmpDirpath, "")

	file, err := cache.RegisterStaticFile("test-key")
	assert.Nil(t, err)

	// Check that the result exists
	_, err = os.Stat(file.absoluteFilepath)
	assert.Nil(t, err)
}
