/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_directory

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/kurtosis-tech/stacktrace"
	"os"
	"path"
)

const (
	generatedFileFilenamePrefix = "generated"
	staticFileFilenamePrefix = "static"
)

// API for interacting with a service's directory inside the enclave data dir
type ServiceDirectory struct {
	absoluteDirpath              string
	dirpathRelativeToDataDirRoot string
}

func newServiceDirectory(absoluteDirpath string, dirpathRelativeToDataDirRoot string) *ServiceDirectory {
	return &ServiceDirectory{absoluteDirpath: absoluteDirpath, dirpathRelativeToDataDirRoot: dirpathRelativeToDataDirRoot}
}

func (directory ServiceDirectory) GetDirpathRelativeToDataDirRoot() string {
	return directory.dirpathRelativeToDataDirRoot
}

func (directory ServiceDirectory) NewGeneratedFile(generatedFileKey string) (*EnclaveDataDirFile, error) {
	file, err := directory.getNewFilepath(generatedFileFilenamePrefix, generatedFileKey)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting filepath for generated file key '%v'", generatedFileKey)
	}
	return file, nil
}

func (directory ServiceDirectory) NewStaticFile(staticFileKey string) (*EnclaveDataDirFile, error) {
	file, err := directory.getNewFilepath(staticFileFilenamePrefix, staticFileKey)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting filepath for static file key '%v'", staticFileKey)
	}
	return file, nil
}

func (directory ServiceDirectory) getNewFilepath(prefix string, identifierFragment string) (*EnclaveDataDirFile, error) {
	uniqueId := uuid.New()
	uniqueFilename := fmt.Sprintf("%v_%v_%v", prefix, identifierFragment, uniqueId)

	absoluteFilepath := path.Join(directory.absoluteDirpath, uniqueFilename)
	fp, err := os.Create(absoluteFilepath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating file '%v'", absoluteFilepath)
	}
	fp.Close()

	relativeFilepath := path.Join(directory.dirpathRelativeToDataDirRoot, uniqueFilename)
	return newEnclaveDataDirFile(absoluteFilepath, relativeFilepath), nil
}


