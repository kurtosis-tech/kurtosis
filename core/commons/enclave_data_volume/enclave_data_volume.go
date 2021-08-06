/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_volume

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
	"path"
)

const (
	allServicesDirname = "services"

	// The name of the directory INSIDE THE ENCLAVE DATA VOLUME where files artifacts are being stored
	artifactCacheDirname = "artifact-cache"

	// The name of the directory INSIDE THE ENCLAVE DATA VOLUME where static files from the
	//  testsuite container are stored, and used when launching services
	staticFileCacheDirname = "static-file-cache"
)

// A directory containing all the data associated with a certain enclave (i.e. a Docker subnetwork where services are spun up)
// An enclave is created either per-test (in the testing framework) or per interactive instance (with Kurtosis Interactive)
type EnclaveDataVolume struct {
	absMountDirpath string
}

func NewEnclaveDataVolume(absMountDirpath string) *EnclaveDataVolume {
	return &EnclaveDataVolume{absMountDirpath: absMountDirpath}
}


func (volume EnclaveDataVolume) GetFilesArtifactCache() (*FilesArtifactCache, error) {
	relativeDirpath := artifactCacheDirname
	absoluteDirpath := path.Join(volume.absMountDirpath, artifactCacheDirname)
	if err := ensureDirpathExists(absoluteDirpath); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred ensuring artifact cache dirpath '%v' exists", absoluteDirpath)
	}

	return newFilesArtifactCache(absoluteDirpath, relativeDirpath), nil
}

func (volume EnclaveDataVolume) GetStaticFileCache() (*StaticFileCache, error) {
	relativeDirpath := staticFileCacheDirname
	absoluteDirpath := path.Join(volume.absMountDirpath, staticFileCacheDirname)
	if err := ensureDirpathExists(absoluteDirpath); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred ensuring static file cache dirpath '%v' exists", absoluteDirpath)
	}
	return newStaticFileCache(absoluteDirpath, relativeDirpath), nil
}

// Creates a new, unique service directory for a service with the given service ID
func (volume EnclaveDataVolume) NewServiceDirectory(serviceId string) (*ServiceDirectory, error) {
	allServicesRelativeDirpath := allServicesDirname
	allServicesAbsoluteDirpath := path.Join(volume.absMountDirpath, allServicesDirname)
	if err := ensureDirpathExists(allServicesAbsoluteDirpath); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred ensuring all services dirpath '%v' exists inside the enclave data dir", allServicesAbsoluteDirpath)
	}

	uniqueId := uuid.New()
	serviceDirname := fmt.Sprintf("%v_%v", serviceId, uniqueId.String())
	absoluteServiceDirpath := path.Join(allServicesAbsoluteDirpath, serviceDirname)
	if err := ensureDirpathExists(absoluteServiceDirpath); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred ensuring service dirpath '%v' exists inside the enclave data dir", absoluteServiceDirpath)
	}
	relativeServiceDirpath := path.Join(allServicesRelativeDirpath, serviceDirname)
	return newServiceDirectory(absoluteServiceDirpath, relativeServiceDirpath), nil
}
