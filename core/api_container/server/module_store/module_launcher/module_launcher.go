/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package module_launcher

import (
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"github.com/kurtosis-tech/kurtosis/api_container/server/module_store"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network/container_name_provider"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/kurtosis-tech/kurtosis/commons/free_host_port_binding_supplier"
	"github.com/palantir/stacktrace"
)

type ModuleLauncher struct {
	dockerManager *docker_manager.DockerManager

	containerNameElemsProvider *container_name_provider.ContainerNameElementsProvider

	freeIpAddrTracker *commons.FreeIpAddrTracker

	// A nil value for this field indicates that no service port <-> host port bindings should be done
	freeHostPortBindingSupplier *free_host_port_binding_supplier.FreeHostPortBindingSupplier

	dockerNetworkId string
}

// TODO Constructor

func (launcher ModuleLauncher) Launch(ctx context.Context, containerImage string, paramsJsonStr string) (containerId string, containerIpAddr string, hostPortBindings map[nat.Port]*nat.PortBinding, resultErr error) {
	moduleIdUuid, err := uuid.NewUUID()
	if err != nil {
		return "", "", nil, stacktrace.Propagate(err, "An error occurred generating the UUID for the module ID")
	}
	moduleId := module_store.ModuleID(moduleIdUuid.String())

	moduleIpAddr, err := launcher.freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return "", "", nil, stacktrace.Propagate(err, "An error occurred getting a free IP address for new module")
	}

	// vvvvvvvvvvvv TODO TODO REPLACE THIS WITH PORT VALUE FROM MODULE LIBRARY vvvvvvvvvv
	moduleProtocol := "tcp"
	modulePort := "9613"
	// ^^^^^^^^^^^^ TODO TODO REPLACE THIS WITH PORT VALUE FROM MODULE LIBRARY ^^^^^^^^^^
	modulePortObj, err := nat.NewPort(moduleProtocol, modulePort)
	if err != nil {
		return "", "", nil, stacktrace.Propagate(err, "An error occurred creating port object for module port %v/%v", moduleProtocol, modulePort)
	}
	usedPorts := map[nat.Port]bool {
		modulePortObj: true,
	}

	hostPortBindings := map[nat.Port]*nat.PortBinding{}
	for port, _ := range usedPorts {
		if launcher.freeHostPortBindingSupplier != nil {
			binding, err := launcher.freeHostPortBindingSupplier.GetFreePortBinding()
			if err != nil {
				return "", "", nil, stacktrace.Propagate(
					err,
					"Host port binding was requested, but an error occurred getting a free host port to bind to service port %v",
					port.Port(),
				)
			}
			hostPortBindings[port] = &binding
		}
	}

	launcher.dockerManager.CreateAndStartContainer(
		ctx,
		containerImage,
		store.containerNameElemProvider.GetForModule(),
		store.containerNameElemProvider.GetForModule(moduleId),
		moduleIpAddr,
		map[docker_manager.ContainerCapability]bool{}, // No extra capapbilities needed for modules
		docker_manager.DefaultNetworkMode,
	)

	volumeMounts := map[string]string{
		launcher.suiteExecutionVolName: suiteExecutionVolMntDirpath,
	}
	for artifactVolName, mountpoint := range artifactVolToMountpoint {
		volumeMounts[artifactVolName] = mountpoint
	}

	containerId, err := launcher.dockerManager.CreateAndStartContainer(
		ctx,
		imageName,
		launcher.containerNameElemsProvider.GetForUserService(serviceId),
		launcher.dockerNetworkId,
		ipAddr,
		map[docker_manager.ContainerCapability]bool{},
		docker_manager.DefaultNetworkMode,
		hostPortBindings,
		entrypointArgs,
		cmdArgs,
		dockerEnvVars,
		map[string]string{}, // no bind mounts for services created via the Kurtosis API
		volumeMounts,
		false,		// User services definitely shouldn't be able to access the Docker host machine
	)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred starting the Docker container for service with image '%v'", imageName)
	}
	return containerId, hostPortBindings, nil
}