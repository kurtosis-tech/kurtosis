/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package suite_execution_volume

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
	"path"
)

const (
	allServicesDirname = "services"

	// The name of the directory INSIDE THE TEST EXECUTION VOLUME where artifacts are being
	//  a) stored using the initializer and b) retrieved using the files artifact expander
	artifactCacheDirname = "artifact-cache"

	// The name of the directory INSIDE THE TEST EXECUTION VOLUME where static files from the
	//  testsuite container are stored, and used when launching services
	staticFileCacheDirname = "static-file-cache"
)

// A directory containing all the data associated with a certain enclave (i.e. a Docker subnetwork where services are spun up)
// In testing, this will correspond to the name of the test
// In interactive mode, this will be the enclave in which the user's test network is running
type EnclaveDirectory struct {
	absoluteDirpath          string
	dirpathRelativeToVolRoot string
}

func newEnclaveDirectory(absoluteDirpath string, dirpathRelativeToVolRoot string) *EnclaveDirectory {
	return &EnclaveDirectory{absoluteDirpath: absoluteDirpath, dirpathRelativeToVolRoot: dirpathRelativeToVolRoot}
}


func (enclaveDir EnclaveDirectory) GetArtifactCache() (*ArtifactCache, error) {
	relativeDirpath := path.Join(enclaveDir.dirpathRelativeToVolRoot, artifactCacheDirname)
	absoluteDirpath := path.Join(enclaveDir.absoluteDirpath, artifactCacheDirname)
	if err := ensureDirpathExists(absoluteDirpath); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred ensuring artifact cache dirpath '%v' exists", absoluteDirpath)
	}

	return newArtifactCache(absoluteDirpath, relativeDirpath), nil
}

func (enclaveDir EnclaveDirectory) GetStaticFileCache() (*StaticFileCache, error) {
	relativeDirpath := path.Join(enclaveDir.dirpathRelativeToVolRoot, staticFileCacheDirname)
	absoluteDirpath := path.Join(enclaveDir.absoluteDirpath, staticFileCacheDirname)
	if err := ensureDirpathExists(absoluteDirpath); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred ensuring static file cache dirpath '%v' exists", absoluteDirpath)
	}
	return newStaticFileCache(absoluteDirpath, relativeDirpath), nil
}

// Creates a new, unique service directory for a service with the given service ID
func (enclaveDir EnclaveDirectory) NewServiceDirectory(serviceId string) (*ServiceDirectory, error) {
	allServicesAbsoluteDirpath := path.Join(enclaveDir.absoluteDirpath, allServicesDirname)
	if err := ensureDirpathExists(allServicesAbsoluteDirpath); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred ensuring all services dirpath '%v' exists inside test execution dir", allServicesAbsoluteDirpath)
	}
	allServicesRelativeDirpath := path.Join(enclaveDir.dirpathRelativeToVolRoot, allServicesDirname)

	uniqueId := uuid.New()
	serviceDirname := fmt.Sprintf("%v_%v", serviceId, uniqueId.String())
	absoluteServiceDirpath := path.Join(allServicesAbsoluteDirpath, serviceDirname)
	if err := ensureDirpathExists(absoluteServiceDirpath); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred ensuring service dirpath '%v' exists inside test execution dir", absoluteServiceDirpath)
	}
	relativeServiceDirpath := path.Join(allServicesRelativeDirpath, serviceDirname)
	return newServiceDirectory(absoluteServiceDirpath, relativeServiceDirpath), nil
}
