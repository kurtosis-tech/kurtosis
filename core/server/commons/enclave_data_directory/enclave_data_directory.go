/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_directory

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_modules/git_module_content_provider"
	"github.com/kurtosis-tech/stacktrace"
	"path"
)

const (
	// The name of the directory INSIDE THE ENCLAVE DATA DIR where files artifacts are being stored.
	// This will replace artifactCacheDirname
	artifactStoreDirname = "artifact-store"

	// The name of the directory INSIDE THE ENCLAVE DATA DIR where startosis modules will be stored
	startosisModuleStoreDirname = "startosis-modules"

	// The name of the directory INSIDE THE ENCLAVE DATA DIR where temporary modules will be stored
	// We place the temp folder here so that the move to the final destination is atomic
	// Move from places outside of the enclave data dir are not atomic as they're over the network
	tmpModuleStoreDirname = "tmp-startosis-modules"
)

// A directory containing all the data associated with a certain enclave (i.e. a Docker subnetwork where services are spun up)
// An enclave is created either per-test (in the testing framework) or per interactive instance (with Kurtosis Interactive)
type EnclaveDataDirectory struct {
	absMountDirpath string
}

func NewEnclaveDataDirectory(absMountDirpath string) *EnclaveDataDirectory {
	return &EnclaveDataDirectory{absMountDirpath: absMountDirpath}
}

func (dir EnclaveDataDirectory) GetFilesArtifactStore() (*FilesArtifactStore, error) {
	relativeDirpath := artifactStoreDirname
	absoluteDirpath := path.Join(dir.absMountDirpath, relativeDirpath)
	if err := ensureDirpathExists(absoluteDirpath); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred ensuring the files artifact store dirpath '%v' exists.", absoluteDirpath)
	}

	return newFilesArtifactStore(absoluteDirpath, relativeDirpath), nil
}

func (dir EnclaveDataDirectory) GetGitModuleManager() (*git_module_content_provider.GitModuleManager, error) {
	moduleStoreDirpath := path.Join(dir.absMountDirpath, startosisModuleStoreDirname)
	if err := ensureDirpathExists(moduleStoreDirpath); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred ensuring the Startosis module store dirpath '%v' exists.", moduleStoreDirpath)
	}

	tempModuleStoreDirpath := path.Join(dir.absMountDirpath, tmpModuleStoreDirname)
	if err := ensureDirpathExists(tempModuleStoreDirpath); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred ensuring the Startosis temporary module store dirpath '%v' exists.", tempModuleStoreDirpath)
	}

	return git_module_content_provider.NewGitModuleManager(moduleStoreDirpath, tempModuleStoreDirpath), nil
}
