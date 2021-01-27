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
)

type TestExecutionDirectory struct {
	absoluteDirpath          string
	dirpathRelativeToVolRoot string
}

func newTestExecutionDirectory(absoluteDirpath string, dirpathRelativeToVolRoot string) *TestExecutionDirectory {
	return &TestExecutionDirectory{absoluteDirpath: absoluteDirpath, dirpathRelativeToVolRoot: dirpathRelativeToVolRoot}
}

// Creates a new, unique service directory for a service with the given service ID
func (executionDir TestExecutionDirectory) GetServiceDirectory(serviceId string) (*ServiceDirectory, error) {
	allServicesAbsoluteDirpath := path.Join(executionDir.absoluteDirpath, allServicesDirname)
	if err := ensureDirpathExists(allServicesAbsoluteDirpath); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred ensuring all services dirpath '%v' exists inside test execution dir", allServicesAbsoluteDirpath)
	}
	allServicesRelativeDirpath := path.Join(executionDir.dirpathRelativeToVolRoot, allServicesDirname)

	uniqueId := uuid.New()
	serviceDirname := fmt.Sprintf("%v_%v", serviceId, uniqueId.String())
	absoluteServiceDirpath := path.Join(allServicesAbsoluteDirpath, serviceDirname)
	if err := ensureDirpathExists(absoluteServiceDirpath); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred ensuring service dirpath '%v' exists inside test execution dir", absoluteServiceDirpath)
	}
	relativeServiceDirpath := path.Join(allServicesRelativeDirpath, serviceDirname)
	return newServiceDirectory(absoluteServiceDirpath, relativeServiceDirpath), nil
}
