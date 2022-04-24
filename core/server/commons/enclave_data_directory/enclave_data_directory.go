/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_directory

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"path"
)

const (
	allServicesDirname = "services"

	// The name of the directory INSIDE THE ENCLAVE DATA DIR where files artifacts are being stored
	artifactCacheDirname = "artifact-cache"

	// The name of the directory INSIDE THE ENCLAVE DATA DIR where files artifacts are being stored.
	// This will replace artifactCacheDirname
	artifactStoreDirname = "artifact-store"

	// The name of the directory INSIDE THE ENCLAVE DATA DIR where static files from the
	//  testsuite container are stored, and used when launching services
	staticFileCacheDirname = "static-file-cache"
)

// A directory containing all the data associated with a certain enclave (i.e. a Docker subnetwork where services are spun up)
// An enclave is created either per-test (in the testing framework) or per interactive instance (with Kurtosis Interactive)
type EnclaveDataDirectory struct {
	absMountDirpath string
}

func NewEnclaveDataDirectory(absMountDirpath string) *EnclaveDataDirectory {
	return &EnclaveDataDirectory{absMountDirpath: absMountDirpath}
}


func (dir EnclaveDataDirectory) GetFilesArtifactCache() (*FilesArtifactCache, error) {
	relativeDirpath := artifactCacheDirname
	absoluteDirpath := path.Join(dir.absMountDirpath, artifactCacheDirname)
	if err := ensureDirpathExists(absoluteDirpath); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred ensuring artifact cache dirpath '%v' exists.", absoluteDirpath)
	}

	return newFilesArtifactCache(absoluteDirpath, relativeDirpath), nil
}

func (dir EnclaveDataDirectory) GetStaticFileCache() (*StaticFileCache, error) {
	relativeDirpath := staticFileCacheDirname
	absoluteDirpath := path.Join(dir.absMountDirpath, staticFileCacheDirname)
	if err := ensureDirpathExists(absoluteDirpath); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred ensuring static file cache dirpath '%v' exists.", absoluteDirpath)
	}
	return newStaticFileCache(absoluteDirpath, relativeDirpath), nil
}

func (dir EnclaveDataDirectory) GetFilesArtifactStore() (*FilesArtifactStore,error) {
	relativeDirpath := artifactStoreDirname
	absoluteDirpath := path.Join(dir.absMountDirpath, relativeDirpath)
	if err := ensureDirpathExists(absoluteDirpath); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred ensuring the files artifact store dirpath '%v' exists.", absoluteDirpath)
	}

	return newFilesArtifactStore(absoluteDirpath, relativeDirpath), nil
}

// Get the unique service directory for a service with the given service GUID
func (dir EnclaveDataDirectory) GetServiceDirectory(serviceGuid service.ServiceGUID) (*ServiceDirectory, error) {
	allServicesRelativeDirpath := allServicesDirname
	allServicesAbsoluteDirpath := path.Join(dir.absMountDirpath, allServicesDirname)
	if err := ensureDirpathExists(allServicesAbsoluteDirpath); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred ensuring all services dirpath '%v' exists inside the enclave data dir.", allServicesAbsoluteDirpath)
	}

	absoluteServiceDirpath := path.Join(allServicesAbsoluteDirpath, string(serviceGuid))
	if err := ensureDirpathExists(absoluteServiceDirpath); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred ensuring service dirpath '%v' exists inside the enclave data dir.", absoluteServiceDirpath)
	}
	relativeServiceDirpath := path.Join(allServicesRelativeDirpath, string(serviceGuid))
	return newServiceDirectory(absoluteServiceDirpath, relativeServiceDirpath), nil
}