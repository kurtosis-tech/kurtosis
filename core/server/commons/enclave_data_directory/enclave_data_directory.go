/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_directory

import (
	"github.com/kurtosis-tech/stacktrace"
	"path"
)

const (
	// The name of the directory INSIDE THE ENCLAVE DATA DIR where files artifacts are being stored.
	// This will replace artifactCacheDirname
	artifactStoreDirname = "artifact-store"
)

// A directory containing all the data associated with a certain enclave (i.e. a Docker subnetwork where services are spun up)
// An enclave is created either per-test (in the testing framework) or per interactive instance (with Kurtosis Interactive)
type EnclaveDataDirectory struct {
	absMountDirpath string
}

func NewEnclaveDataDirectory(absMountDirpath string) *EnclaveDataDirectory {
	return &EnclaveDataDirectory{absMountDirpath: absMountDirpath}
}

func (dir EnclaveDataDirectory) GetFilesArtifactStore() (*FilesArtifactStore,error) {
	relativeDirpath := artifactStoreDirname
	absoluteDirpath := path.Join(dir.absMountDirpath, relativeDirpath)
	if err := ensureDirpathExists(absoluteDirpath); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred ensuring the files artifact store dirpath '%v' exists.", absoluteDirpath)
	}

	return newFilesArtifactStore(absoluteDirpath, relativeDirpath), nil
}