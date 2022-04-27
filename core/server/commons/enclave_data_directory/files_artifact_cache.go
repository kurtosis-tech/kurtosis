/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_directory

import (
	"bufio"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"io"
	// This is a special type of import that includes the correct hashing algorithm that we use
	// If we don't have the "_" in front, Goland will complain it's unused
	_ "golang.org/x/crypto/sha3"
	"net/http"
)

/*
A file cache inside the for storing files artifacts, downloaded from the URL
 */
type FilesArtifactCache struct {
	underlying *FileCache
}

func newFilesArtifactCache(absoluteDirpath string, dirpathRelativeToDataDirRoot string) *FilesArtifactCache {
	return &FilesArtifactCache{
		underlying: newFileCache(absoluteDirpath, dirpathRelativeToDataDirRoot),
	}
}

// StoreFile: Saves file to disk.
func (cache FilesArtifactCache) StoreFile(reader io.Reader, filename string) (string, error) {
	uuid, err := getUniversallyUniqueID()
	if err != nil{
		return "", stacktrace.Propagate(err, "Could not generate Universally Unique ID.")
	}
	_, err = cache.underlying.AddFile(filename, reader)
	if err != nil{
		return "", stacktrace.Propagate(err, "Could not add file with UUID %s at %s.", uuid,
			cache.underlying.absoluteDirpath)
	}
	return uuid, nil
}

func (cache FilesArtifactCache) DownloadFilesArtifact(artifactId string, url string) error {

	resp, err := http.Get(url)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred making the request to URL '%v' to get the data for artifact '%v'", url, artifactId)
	}
	defer resp.Body.Close()
	body := bufio.NewReader(resp.Body)

	if _, err := cache.underlying.AddFile(artifactId, body); err != nil {
		return stacktrace.Propagate(err, "An error occurred downloading the files artifact '%v' from URL '%v'", artifactId, url)
	}
	return nil
}

// Gets the artifact with the given URL, or throws an error if it doesn't exist
func (cache FilesArtifactCache) GetFilesArtifact(artifactId service.FilesArtifactID) (*EnclaveDataDirFile, error) {
	artifactIdStr := string(artifactId)
	result, err := cache.underlying.GetFile(artifactIdStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting files artifact with ID '%v' from the cache", artifactId)
	}
	return result, nil
}
