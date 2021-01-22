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
	allServicesDirname = "services"
)

// Interface for interacting with the contents of the suite execution volume
type SuiteExecutionVolume struct {
	mountDirpath string
}

func NewSuiteExecutionVolume(mountDirpath string) *SuiteExecutionVolume {
	return &SuiteExecutionVolume{mountDirpath: mountDirpath}
}

// TODO change type to be ServiceID
// Creates a new, unique service directory for a service with the given service ID
func (volume SuiteExecutionVolume) CreateServiceDirectory(serviceId string) (*ServiceDirectory, error) {
	allServicesRelativeDirpath := allServicesDirname
	uniqueId := uuid.New()
	serviceDirname := fmt.Sprintf("%v_%v", serviceId, uniqueId.String())
	relativeServiceDirpath := path.Join(allServicesRelativeDirpath, serviceDirname)
	absoluteServiceDirpath := path.Join(volume.mountDirpath, relativeServiceDirpath)
	if err := os.Mkdir(absoluteServiceDirpath, os.ModeDir); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating new service directory '%v'", absoluteServiceDirpath)
	}
	serviceDirectoryObj := newServiceDirectory(absoluteServiceDirpath, relativeServiceDirpath)
	return serviceDirectoryObj, nil
}
