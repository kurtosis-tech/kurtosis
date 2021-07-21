/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package enclave_data_volume

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
	"os"
	"path"
)

const (
	generatedFileFilenamePrefix = "generated"
	staticFileFilenamePrefix = "static"
)

// API for interacting with a service's directory inside the enclave data volume
type ServiceDirectory struct {
	absoluteDirpath          string
	dirpathRelativeToVolRoot string
}

func newServiceDirectory(absoluteDirpath string, dirpathRelativeToVolRoot string) *ServiceDirectory {
	return &ServiceDirectory{absoluteDirpath: absoluteDirpath, dirpathRelativeToVolRoot: dirpathRelativeToVolRoot}
}

func (directory ServiceDirectory) NewGeneratedFile(generatedFileKey string) (*EnclaveDataVolFile, error) {
	file, err := directory.getNewFilepath(generatedFileFilenamePrefix, generatedFileKey)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting filepath for generated file key '%v'", generatedFileKey)
	}
	return file, nil
}

func (directory ServiceDirectory) NewStaticFile(staticFileKey string) (*EnclaveDataVolFile, error) {
	file, err := directory.getNewFilepath(staticFileFilenamePrefix, staticFileKey)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting filepath for static file key '%v'", staticFileKey)
	}
	return file, nil
}

func (directory ServiceDirectory) getNewFilepath(prefix string, identifierFragment string) (*EnclaveDataVolFile, error) {
	uniqueId := uuid.New()
	uniqueFilename := fmt.Sprintf("%v_%v_%v", prefix, identifierFragment, uniqueId)

	absoluteFilepath := path.Join(directory.absoluteDirpath, uniqueFilename)
	fp, err := os.Create(absoluteFilepath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating file '%v'", absoluteFilepath)
	}
	fp.Close()

	relativeFilepath := path.Join(directory.dirpathRelativeToVolRoot, uniqueFilename)
	return newEnclaveDataVolFile(absoluteFilepath, relativeFilepath), nil
}


