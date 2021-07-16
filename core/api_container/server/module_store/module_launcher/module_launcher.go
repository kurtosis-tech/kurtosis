/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package module_launcher

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis-client/golang/core_api_consts"
	"github.com/kurtosis-tech/kurtosis/api_container/server/module_store/module_store_types"
	"github.com/kurtosis-tech/kurtosis/api_container/server/optional_host_port_binding_supplier"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network/container_name_provider"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/kurtosis-tech/kurtosis/kurtosis_module/docker_api/kurtosis_module_env_vars"
	"github.com/kurtosis-tech/kurtosis/kurtosis_module/kurtosis_module_rpc_api/kurtosis_module_rpc_api_consts"
	"github.com/palantir/stacktrace"
	"net"
	"strconv"
)

type ModuleLauncher struct {
	dockerManager *docker_manager.DockerManager

	// Modules have a connection to the API container, so the module launcher must know about the API container's IP addr
	apiContainerIpAddr string

	containerNameElemsProvider *container_name_provider.ContainerNameElementsProvider

	freeIpAddrTracker *commons.FreeIpAddrTracker

	optionalHostPortBindingSupplier *optional_host_port_binding_supplier.OptionalHostPortBindingSupplier

	dockerNetworkId string
}

func NewModuleLauncher(dockerManager *docker_manager.DockerManager, apiContainerIpAddr string, containerNameElemsProvider *container_name_provider.ContainerNameElementsProvider, freeIpAddrTracker *commons.FreeIpAddrTracker, optionalHostPortBindingSupplier *optional_host_port_binding_supplier.OptionalHostPortBindingSupplier, dockerNetworkId string) *ModuleLauncher {
	return &ModuleLauncher{dockerManager: dockerManager, apiContainerIpAddr: apiContainerIpAddr, containerNameElemsProvider: containerNameElemsProvider, freeIpAddrTracker: freeIpAddrTracker, optionalHostPortBindingSupplier: optionalHostPortBindingSupplier, dockerNetworkId: dockerNetworkId}
}

func (launcher ModuleLauncher) Launch(ctx context.Context, moduleId module_store_types.ModuleID, containerImage string, paramsJsonStr string) (id string, ipAddr net.IP, hostPortBindings map[nat.Port]*nat.PortBinding, resultErr error) {
	moduleIpAddr, err := launcher.freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred getting a free IP address for new module")
	}

	modulePortNumStr := strconv.Itoa(kurtosis_module_rpc_api_consts.ListenPort)
	modulePortObj, err := nat.NewPort(kurtosis_module_rpc_api_consts.ListenProtocol, modulePortNumStr)
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(
			err,
			"An error occurred creating port object for module port %v/%v",
			kurtosis_module_rpc_api_consts.ListenProtocol,
			kurtosis_module_rpc_api_consts.ListenPort,
		)
	}
	usedPorts := map[nat.Port]bool {
		modulePortObj: true,
	}

	usedPortsWithHostBindings, err := launcher.optionalHostPortBindingSupplier.BindPortsToHostIfNeeded(usedPorts)
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred binding used ports to host ports")
	}

	apiContainerSocket := fmt.Sprintf("%v:%v", launcher.apiContainerIpAddr, core_api_consts.ListenPort)
	envVars := map[string]string{
		kurtosis_module_env_vars.ApiContainerSocketEnvVar: apiContainerSocket,
		kurtosis_module_env_vars.CustomParamsJsonEnvVar: paramsJsonStr,
	}

	containerId, err := launcher.dockerManager.CreateAndStartContainer(
		ctx,
		containerImage,
		launcher.containerNameElemsProvider.GetForModule(moduleId),
		launcher.dockerNetworkId,
		moduleIpAddr,
		map[docker_manager.ContainerCapability]bool{}, // No extra capapbilities needed for modules
		docker_manager.DefaultNetworkMode,
		usedPortsWithHostBindings,
		nil, // No ENTRYPOINT overrides; modules are configured using env vars
		nil, // No CMD overrides; modules are configured using env vars
		envVars,
		nil, // No bind mounts needed
		nil, // No volume mounts needed
		false, // Modules don't need to access the host machine
	)
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred launching the module container")
	}
	return containerId, moduleIpAddr, usedPortsWithHostBindings, nil
}