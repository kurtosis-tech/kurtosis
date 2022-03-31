/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package user_service_launcher

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/free-ip-addr-tracker-lib/lib"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

/*
Convenience struct whose only purpose is launching user services
*/
type UserServiceLauncher struct {
	kurtosisBackend backend_interface.KurtosisBackend

	freeIpAddrTracker *lib.FreeIpAddrTracker
}

func NewUserServiceLauncher(kurtosisBackend backend_interface.KurtosisBackend, freeIpAddrTracker *lib.FreeIpAddrTracker) *UserServiceLauncher {
	return &UserServiceLauncher{kurtosisBackend: kurtosisBackend, freeIpAddrTracker: freeIpAddrTracker}
}

/**
Launches a testnet service with the given parameters

Returns:
	* The container ID of the newly-launched service
	* The mapping of used_port -> host_port_binding (if no host port is bound, then the value will be nil)
*/
func (launcher UserServiceLauncher) Launch(
	ctx context.Context,
	serviceGUID service.ServiceGUID,
	serviceId service.ServiceID,
	imageName string,
	privatePorts []*port_spec.PortSpec,
	entrypointArgs []string,
	cmdArgs []string,
	envVars map[string]string,
	enclaveDataDirMountDirpath string,
// Mapping files artifact ID -> mountpoint on the container to launch
	filesArtifactIdsToMountpoints map[string]string,
) (
	resultUserService *service.Service,
	resultErr error,
) {
	launchedUserService, err := launcher.kurtosisBackend.CreateUserService(
		ctx,
		serviceId,
		serviceGUID,
		imageName,
		privatePorts,
		entrypointArgs,
		cmdArgs,
		envVars,
		enclaveDataDirMountDirpath,
		filesArtifactIdsToMountpoints,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred starting the container for user service in enclave '%v' with image '%v'", imageName)
	}
	shouldKillService := true
	defer func() {
		if shouldKillService {
			_, erroredUserServices, err := launcher.kurtosisBackend.DestroyUserServices(ctx, getServiceByServiceGUIDFilter(serviceGUID))
			if err != nil {
				logrus.Errorf("Launching the service failed, but an error occurred calling the backend to destroy the service:\n%v", err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually kill service with ID '%v'", serviceGUID)
			}
			for serviceGUID, err := range erroredUserServices {
				logrus.Errorf("Launching the service failed, but an error occurred destroying the service:\n%v", err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually kill service with ID '%v'", serviceGUID)
			}
		}
	}()

	shouldKillService = false
	return launchedUserService, nil
}

func getServiceByServiceGUIDFilter(serviceGUID service.ServiceGUID) *service.ServiceFilters {
	return &service.ServiceFilters{
		GUIDs: map[service.ServiceGUID]bool{
			serviceGUID: true,
		},
	}
}
