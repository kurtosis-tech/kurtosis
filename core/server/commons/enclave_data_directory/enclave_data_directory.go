/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_directory

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages/git_package_content_provider"
	"github.com/kurtosis-tech/stacktrace"
	"path"
	"sync"
)

const (
	// The name of the directory INSIDE THE ENCLAVE DATA DIR where files artifacts are being stored.
	// This will replace artifactCacheDirname
	artifactStoreDirname = "artifact-store"

	// The name of the directory INSIDE THE ENCLAVE DATA DIR where Starlark packages will be stored
	startosisPackageStoreDirname = "startosis-packages"

	// The name of the directory INSIDE THE ENCLAVE DATA DIR where temporary packages will be stored
	// We place the temp folder here so that the move to the final destination is atomic
	// Move from places outside of the enclave data dir are not atomic as they're over the network
	tmpPackageStoreDirname = "tmp-startosis-packages"
)

// A directory containing all the data associated with a certain enclave (i.e. a Docker subnetwork where services are spun up)
// An enclave is created either per-test (in the testing framework) or per interactive instance (with Kurtosis Interactive)
type EnclaveDataDirectory struct {
	absMountDirpath string
}

var (
	// NOTE: This will be initialized exactly once (singleton pattern)
	currentFilesArtifactStore *FilesArtifactStore
	once                      sync.Once
)

func NewEnclaveDataDirectory(absMountDirpath string) *EnclaveDataDirectory {
	return &EnclaveDataDirectory{absMountDirpath: absMountDirpath}
}

func (dir EnclaveDataDirectory) GetFilesArtifactStore() (*FilesArtifactStore, error) {
	relativeDirpath := artifactStoreDirname
	absoluteDirpath := path.Join(dir.absMountDirpath, relativeDirpath)
	if err := ensureDirpathExists(absoluteDirpath); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred ensuring the files artifact store dirpath '%v' exists.", absoluteDirpath)
	}

	// NOTE: We use a 'once' to initialize the filesArtifactStore because it contains a mutex,
	// and we don't ever want multiple filesArtifactStore instances in existence
	once.Do(func() {
		currentFilesArtifactStore = newFilesArtifactStore(absoluteDirpath, relativeDirpath)
	})

	return currentFilesArtifactStore, nil
}

func (dir EnclaveDataDirectory) GetGitPackageContentProvider() (*git_package_content_provider.GitPackageContentProvider, error) {
	packageStoreDirpath := path.Join(dir.absMountDirpath, startosisPackageStoreDirname)
	if err := ensureDirpathExists(packageStoreDirpath); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred ensuring the Starlark package store dirpath '%v' exists.", packageStoreDirpath)
	}

	tempPackageStoreDirpath := path.Join(dir.absMountDirpath, tmpPackageStoreDirname)
	if err := ensureDirpathExists(tempPackageStoreDirpath); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred ensuring the Starlark temporary package store dirpath '%v' exists.", tempPackageStoreDirpath)
	}

	return git_package_content_provider.NewGitPackageContentProvider(packageStoreDirpath, tempPackageStoreDirpath), nil
}
