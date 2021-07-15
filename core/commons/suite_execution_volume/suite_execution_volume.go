/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package suite_execution_volume

import (
	"github.com/palantir/stacktrace"
	"path"
	"strings"
)

const (
	suiteMetadataFilename = "suite-metadata.json"

	// The name of the directory INSIDE THE TEST EXECUTION VOLUME where artifacts are being
	//  a) stored using the initializer and b) retrieved using the files artifact expander
	artifactCacheDirname = "artifact-cache"

	// The name of the directory INSIDE THE TEST EXECUTION VOLUME where static files from the
	//  testsuite container are stored, and used when launching services
	staticFileCacheDirname = "static-file-cache"

	enclaveNameJoinChar = "_"
)

// Interface for interacting with the contents of the suite execution volume
type SuiteExecutionVolume struct {
	mountDirpath string
}

func NewSuiteExecutionVolume(mountDirpath string) *SuiteExecutionVolume {
	return &SuiteExecutionVolume{mountDirpath: mountDirpath}
}

// TODO Refactor this entire thing so that there's one volume per enclave, which requires pushing the artifact cache
//  into a separate volume (or better yet, on the local filesystem)
func (volume SuiteExecutionVolume) GetEnclaveDirectory(enclaveNameElems []string) (*EnclaveDirectory, error) {
	enclaveName := strings.Join(enclaveNameElems, enclaveNameJoinChar)
	relativeDirpath := enclaveName
	absoluteDirpath := path.Join(volume.mountDirpath, relativeDirpath)
	if err := ensureDirpathExists(absoluteDirpath); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred ensuring test execution dirpath '%v' exists", absoluteDirpath)
	}
	return newEnclaveDirectory(absoluteDirpath, relativeDirpath), nil
}

func (volume SuiteExecutionVolume) GetArtifactCache() (*ArtifactCache, error) {
	relativeDirpath := artifactCacheDirname
	absoluteDirpath := path.Join(volume.mountDirpath, relativeDirpath)
	if err := ensureDirpathExists(absoluteDirpath); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred ensuring artifact cache dirpath '%v' exists", absoluteDirpath)
	}

	return newArtifactCache(absoluteDirpath, relativeDirpath), nil
}

func (volume SuiteExecutionVolume) GetStaticFileCache() (*StaticFileCache, error) {
	relativeDirpath := staticFileCacheDirname
	absoluteDirpath := path.Join(volume.mountDirpath, relativeDirpath)
	if err := ensureDirpathExists(absoluteDirpath); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred ensuring static file cache dirpath '%v' exists", absoluteDirpath)
	}
	return newStaticFileCache(absoluteDirpath, relativeDirpath), nil
}

