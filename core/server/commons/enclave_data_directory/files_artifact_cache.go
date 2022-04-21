/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_directory

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact"
	"github.com/kurtosis-tech/stacktrace"
	// This is a special type of import that includes the correct hashing algorithm that we use
	// If we don't have the "_" in front, Goland will complain it's unused
	_ "golang.org/x/crypto/sha3"
	"net"
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

func (cache FilesArtifactCache) DownloadFilesArtifact(artifactId string, method string, url string) error {

	connection, err := net.Dial(method, url)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred making the request to URL '%v' to get the data for aritfact '%v'", url, artifactId)
	}
	defer connection.Close()

	_, err = cache.underlying.AddFile(artifactId, connection)
	return err
}

// Gets the artifact with the given URL, or throws an error if it doesn't exist
func (cache FilesArtifactCache) GetFilesArtifact(artifactId files_artifact.FilesArtifactID) (*EnclaveDataDirFile, error) {
	artifactIdStr := string(artifactId)
	result, err := cache.underlying.GetFile(artifactIdStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting files artifact with ID '%v' from the cache", artifactId)
	}
	return result, nil
}
