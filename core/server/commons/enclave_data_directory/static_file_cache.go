/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_directory

import (
	"github.com/kurtosis-tech/stacktrace"
	"strings"
)

/*
An interface for interacting with the static file cache directory that exists inside the enclave data dir
*/
type StaticFileCache struct {
	underlying *FileCache
}

func newStaticFileCache(absoluteDirpath string, dirpathRelativeToDataDirRoot string) *StaticFileCache {
	return &StaticFileCache{
		underlying: newFileCache(absoluteDirpath, dirpathRelativeToDataDirRoot),
	}
}

func (cache *StaticFileCache) RegisterStaticFile(key string) (*EnclaveDataDirFile, error) {
	// The static file cache needs to initialize an empty file, which the underlying filecache already does so need
	//  to do anything
	result, err := cache.underlying.AddFile(key, strings.NewReader(""))
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding an empty static file to the underlying cache")
	}
	return result, nil
}

func (cache *StaticFileCache) GetStaticFile(key string) (*EnclaveDataDirFile, error) {
	result, err := cache.underlying.GetFile(key)

	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting static file entry '%v' from the underlying file cache", key)
	}
	return result, nil

}
