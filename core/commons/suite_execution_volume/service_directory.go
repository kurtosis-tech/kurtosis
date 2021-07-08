/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package suite_execution_volume

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

// API for interacting with a service's directory inside the suite execution volume
type ServiceDirectory struct {
	absoluteDirpath          string
	dirpathRelativeToVolRoot string
}

func newServiceDirectory(absoluteDirpath string, dirpathRelativeToVolRoot string) *ServiceDirectory {
	return &ServiceDirectory{absoluteDirpath: absoluteDirpath, dirpathRelativeToVolRoot: dirpathRelativeToVolRoot}
}

func (directory ServiceDirectory) GetGeneratedFile(generatedFileKey string) (*File, error) {
	file, err := directory.getFile(generatedFileFilenamePrefix, generatedFileKey)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting generated file for generated file key '%v'", generatedFileKey)
	}
	return file, nil
}

func (directory ServiceDirectory) GetStaticFile(staticFileKey string) (*File, error) {
	file, err := directory.getFile(staticFileFilenamePrefix, staticFileKey)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting static file for static file key '%v'", staticFileKey)
	}
	return file, nil
}

func (directory ServiceDirectory) getFile(prefix string, identifierFragment string) (*File, error) {
	uniqueId := uuid.New()
	uniqueFilename := fmt.Sprintf("%v_%v_%v", prefix, identifierFragment, uniqueId)

	absoluteFilepath := path.Join(directory.absoluteDirpath, uniqueFilename)
	fp, err := os.Create(absoluteFilepath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating file '%v'", absoluteFilepath)
	}
	fp.Close()

	relativeFilepath := path.Join(directory.dirpathRelativeToVolRoot, uniqueFilename)
	return newFile(absoluteFilepath, relativeFilepath), nil
}


