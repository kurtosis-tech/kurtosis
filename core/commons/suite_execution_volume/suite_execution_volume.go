/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package suite_execution_volume

import (
	"github.com/palantir/stacktrace"
	"path"
)

const (
	suiteMetadataFilename = "suite-metadata.json"
)

// Interface for interacting with the contents of the suite execution volume
type SuiteExecutionVolume struct {
	mountDirpath string
}

func NewSuiteExecutionVolume(mountDirpath string) *SuiteExecutionVolume {
	return &SuiteExecutionVolume{mountDirpath: mountDirpath}
}

func (volume SuiteExecutionVolume) GetSuiteMetadataFile() *File {
	relativeFilepath := suiteMetadataFilename
	absoluteFilepath := path.Join(volume.mountDirpath, relativeFilepath)
	return newFile(absoluteFilepath, relativeFilepath)
}

func (volume SuiteExecutionVolume) CreateTestExecutionDirectory(testExecutionId string) (*TestExecutionDirectory, error) {
	relativeDirpath := testExecutionId
	absoluteDirpath := path.Join(volume.mountDirpath, relativeDirpath)
	if err := ensureDirpathExists(absoluteDirpath); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred ensuring test execution dirpath '%v' exists", relativeDirpath)
	}
	return newTestExecutionDirectory(absoluteDirpath, relativeDirpath), nil
}


