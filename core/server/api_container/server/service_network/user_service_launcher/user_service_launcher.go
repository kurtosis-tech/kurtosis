/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package user_service_launcher

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/free-ip-addr-tracker-lib/lib"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/service_network/user_service_launcher/files_artifact_expander"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
)

/*
Convenience struct whose only purpose is launching user services
*/
type UserServiceLauncher struct {
	kurtosisBackend backend_interface.KurtosisBackend
	filesArtifactExpander *files_artifact_expander.FilesArtifactExpander
	freeIpAddrTracker *lib.FreeIpAddrTracker
}

func NewUserServiceLauncher(kurtosisBackend backend_interface.KurtosisBackend, filesArtifactExpander *files_artifact_expander.FilesArtifactExpander, freeIpAddrTracker *lib.FreeIpAddrTracker) *UserServiceLauncher {
	return &UserServiceLauncher{kurtosisBackend: kurtosisBackend, filesArtifactExpander: filesArtifactExpander, freeIpAddrTracker: freeIpAddrTracker}
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
	enclaveId enclave.EnclaveID,
	ipAddr net.IP,
	imageName string,
	privatePorts map[string]*port_spec.PortSpec,
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
	usedArtifactIdSet := map[string]bool{}
	for artifactId := range filesArtifactIdsToMountpoints {
		usedArtifactIdSet[artifactId] = true
	}

	// First expand the files artifacts into volumes, so that any errors get caught early
	// NOTE: if users don't need to investigate the volume contents, we could keep track of the volumes we create
	//  and delete them at the end of the test to keep things cleaner
	artifactIdsToVolumes, err := launcher.filesArtifactExpander.ExpandArtifactsIntoVolumes(ctx, serviceGUID, usedArtifactIdSet)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred expanding the requested files artifacts into volumes")
	}

	artifactVolumeMounts := map[string]string{}
	for artifactId, mountpoint := range filesArtifactIdsToMountpoints {
		artifactVolume, found := artifactIdsToVolumes[artifactId]
		if !found {
			return nil, stacktrace.NewError(
				"Even though we declared that we need files artifact '%v' to be expanded, no volume containing the "+
					"expanded contents was found; this is a bug in Kurtosis",
				artifactId,
			)
		}
		artifactVolumeMounts[artifactVolume] = mountpoint
	}

	launchedUserService, err := launcher.kurtosisBackend.CreateUserService(
		ctx,
		serviceId,
		serviceGUID,
		imageName,
		enclaveId,
		ipAddr,
		privatePorts,
		entrypointArgs,
		cmdArgs,
		envVars,
		enclaveDataDirMountDirpath,
		artifactVolumeMounts,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred starting the container for user service in enclave '%v' with image '%v'", enclaveId, imageName)
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
