/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package suite_execution_volume

import (
	"github.com/palantir/stacktrace"
	"os"
	"path"
	"sync"
)

/*
An interface for interacting with the static file cache directory that exists inside the suite execution volume,
	and is a) populated by the testsuite copying its files into it and b) accessed by the API container as it starts services
*/
type StaticFileCache struct {
	absoluteDirpath string

	// Guard to make sure concurrency can't mess us up
	mutex *sync.Mutex

	dirpathRelativeToVolRoot string
}

func newStaticFileCache(absoluteDirpath string, dirpathRelativeToVolRoot string) *StaticFileCache {
	return &StaticFileCache{absoluteDirpath: absoluteDirpath, dirpathRelativeToVolRoot: dirpathRelativeToVolRoot}
}

// Registers a new static file identified by the given key, which will be filled in by the client
func (cache *StaticFileCache) RegisterEntry(key string) (*File, error) {
	absFilepath := path.Join(cache.absoluteDirpath, key)
	if _, err := os.Stat(absFilepath); err == nil {
		return nil, stacktrace.NewError("Cannot register key '%v'; a static file with that key is already registered", key)
	}

	// Create an empty file to mark off this key
	fp, err := os.Create(absFilepath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating empty file in the static file cache at '%v'", absFilepath)
	}
	fp.Close()

	relativeFilepath := path.Join(cache.dirpathRelativeToVolRoot, key)
	file := newFile(absFilepath, relativeFilepath)
	return file, nil
}

func (cache *StaticFileCache) GetEntry(key string) (*File, error) {
	absFilepath := path.Join(cache.absoluteDirpath, key)
	if _, err := os.Stat(absFilepath); os.IsNotExist(err) {
		return nil, stacktrace.NewError("No static file entry in the cache with key '%v'", key)
	}
	relativeFilepath := path.Join(cache.dirpathRelativeToVolRoot, key)
	return newFile(absFilepath, relativeFilepath), nil
}
