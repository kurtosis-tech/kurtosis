/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_directory

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db/file_artifacts_db"
	"github.com/kurtosis-tech/stacktrace"
	"path"
	"sync"
)

const (
	// The name of the directory INSIDE THE ENCLAVE DATA DIR where files artifacts are being stored.
	// This will replace artifactCacheDirname
	artifactStoreDirname = "artifact-store"

	// The name of the directory INSIDE THE ENCLAVE DATA DIR where Starlark packages will be stored
	repositoriesStoreDirname = "repositories"

	// The name of the directory INSIDE THE ENCLAVE DATA DIR where temporary repositories will be stored
	// We place the temp folder here so that the move to the final destination is atomic
	// Move from places outside the enclave data dir are not atomic as they're over the network
	tmpRepositoriesStoreDirname = "tmp-repositories"

	// Name of directory INSIDE THE ENCLAVE DATA DIR at [absMountDirPath]  that contains info for authenticating GitHub operations
	githubAuthStoreDirname = "github-auth"

	// Name of directory INSIDE THE ENCLAVE DATA DIR containing the enclave database (currently the bolt dB is implemented)
	enclaveDatabase = "enclave-database"
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

	var dbError error
	// NOTE: We use a 'once' to initialize the filesArtifactStore because it contains a mutex,
	// and we don't ever want multiple filesArtifactStore instances in existence
	once.Do(func() {
		db, err := file_artifacts_db.GetOrCreateNewFileArtifactsDb()
		if err != nil {
			dbError = stacktrace.Propagate(err, "Failed to get file artifacts db")
			return
		}
		currentFilesArtifactStore = newFilesArtifactStoreFromDb(absoluteDirpath, relativeDirpath, db)
	})

	return currentFilesArtifactStore, dbError
}

func (dir EnclaveDataDirectory) GetEnclaveDataDirectoryPaths() (string, string, string, string, error) {
	repositoriesStoreDirpath := path.Join(dir.absMountDirpath, repositoriesStoreDirname)
	if err := ensureDirpathExists(repositoriesStoreDirpath); err != nil {
		return "", "", "", "", stacktrace.Propagate(err, "An error occurred ensuring the repositories store dirpath '%v' exists.", repositoriesStoreDirpath)
	}

	tempRepositoriesStoreDirpath := path.Join(dir.absMountDirpath, tmpRepositoriesStoreDirname)
	if err := ensureDirpathExists(tempRepositoriesStoreDirpath); err != nil {
		return "", "", "", "", stacktrace.Propagate(err, "An error occurred ensuring the temporary repositories store dirpath '%v' exists.", tempRepositoriesStoreDirpath)
	}

	githubAuthStoreDirpath := path.Join(dir.absMountDirpath, githubAuthStoreDirname)
	if err := ensureDirpathExists(githubAuthStoreDirpath); err != nil {
		return "", "", "", "", stacktrace.Propagate(err, "An error occurred ensuring the GitHub auth store dirpath '%v' exists.", githubAuthStoreDirpath)
	}

	enclaveDatabaseDirpath := path.Join(dir.absMountDirpath, enclaveDatabase)
	if err := ensureDirpathExists(enclaveDatabaseDirpath); err != nil {
		return "", "", "", "", stacktrace.Propagate(err, "An error occurred ensuring the enclave database store dirpath '%v' exists.", enclaveDatabaseDirpath)
	}

	return repositoriesStoreDirpath, tempRepositoriesStoreDirpath, githubAuthStoreDirpath, enclaveDatabaseDirpath, nil
}
