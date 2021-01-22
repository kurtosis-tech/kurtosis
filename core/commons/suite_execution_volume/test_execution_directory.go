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
	absoluteDirpath string
	relativeDirpath string
}

func newTestExecutionDirectory(absoluteDirpath string, relativeDirpath string) *TestExecutionDirectory {
	return &TestExecutionDirectory{absoluteDirpath: absoluteDirpath, relativeDirpath: relativeDirpath}
}

// TODO change types to be ServiceID type
// Creates a new, unique service directory for a service with the given service ID
func (executionDir TestExecutionDirectory) CreateServiceDirectory(serviceId string) (*ServiceDirectory, error) {
	allServicesRelativeDirpath := allServicesDirname
	allServicesAbsoluteDirpath := path.Join(executionDir.absoluteDirpath, allServicesRelativeDirpath)
	if err := ensureDirpathExists(allServicesAbsoluteDirpath); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred ensuring all services dirpath '%v' exists inside test execution dir", allServicesAbsoluteDirpath)
	}

	uniqueId := uuid.New()
	serviceDirname := fmt.Sprintf("%v_%v", serviceId, uniqueId.String())
	relativeServiceDirpath := path.Join(allServicesRelativeDirpath, serviceDirname)
	absoluteServiceDirpath := path.Join(executionDir.absoluteDirpath, relativeServiceDirpath)
	if err := ensureDirpathExists(absoluteServiceDirpath); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred ensuring service dirpath '%v' exists inside test execution dir", absoluteServiceDirpath)
	}
	return newServiceDirectory(absoluteServiceDirpath, relativeServiceDirpath), nil
}
