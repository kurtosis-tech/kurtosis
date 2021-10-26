/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package module_launcher

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-client/golang/kurtosis_core_rpc_api_consts"
	"github.com/kurtosis-tech/kurtosis-core/api_container/server/module_store/module_store_types"
	"github.com/kurtosis-tech/kurtosis-core/commons"
	"github.com/kurtosis-tech/kurtosis-core/commons/current_time_str_provider"
	"github.com/kurtosis-tech/kurtosis-core/commons/object_labels_providers"
	"github.com/kurtosis-tech/kurtosis-core/commons/object_name_providers"
	"github.com/kurtosis-tech/kurtosis-module-api-lib/golang/kurtosis_module_docker_api"
	"github.com/kurtosis-tech/kurtosis-module-api-lib/golang/kurtosis_module_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-module-api-lib/golang/kurtosis_module_rpc_api_consts"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"net"
	"strconv"
	"time"
)

const (
	waitForModuleAvailabilityTimeout = 10 * time.Second
)

type ModuleLauncher struct {
	dockerManager *docker_manager.DockerManager

	// Modules have a connection to the API container, so the launcher must know about the API container's IP addr
	apiContainerIpAddr string

	enclaveObjNameProvider *object_name_providers.EnclaveObjectNameProvider

	enclaveObjLabelsProvider *object_labels_providers.EnclaveObjectLabelsProvider

	freeIpAddrTracker *commons.FreeIpAddrTracker

	// TODO Publish module ports always, to simplify
	shouldPublishPorts bool

	dockerNetworkId string

	enclaveDataVolName string
}

func NewModuleLauncher(dockerManager *docker_manager.DockerManager, apiContainerIpAddr string, enclaveObjNameProvider *object_name_providers.EnclaveObjectNameProvider, enclaveObjLabelsProvider *object_labels_providers.EnclaveObjectLabelsProvider, freeIpAddrTracker *commons.FreeIpAddrTracker, shouldPublishPorts bool, dockerNetworkId string, enclaveDataVolName string) *ModuleLauncher {
	return &ModuleLauncher{dockerManager: dockerManager, apiContainerIpAddr: apiContainerIpAddr, enclaveObjNameProvider: enclaveObjNameProvider, enclaveObjLabelsProvider: enclaveObjLabelsProvider, freeIpAddrTracker: freeIpAddrTracker, shouldPublishPorts: shouldPublishPorts, dockerNetworkId: dockerNetworkId, enclaveDataVolName: enclaveDataVolName}
}

func (launcher ModuleLauncher) Launch(
		ctx context.Context,
		moduleID module_store_types.ModuleID,
		containerImage string,
		serializedParams string) (newContainerId string, newContainerIpAddr net.IP, client kurtosis_module_rpc_api_bindings.ExecutableModuleServiceClient, moduleHostPortBinding *nat.PortBinding, resultErr error) {

	portNumStr := strconv.Itoa(kurtosis_module_rpc_api_consts.ListenPort)
	portObj, err := nat.NewPort(kurtosis_module_rpc_api_consts.ListenProtocol, portNumStr)
	if err != nil {
		return "", nil, nil, nil, stacktrace.Propagate(
			err,
			"An error occurred creating port object for module port %v/%v",
			kurtosis_module_rpc_api_consts.ListenProtocol,
			kurtosis_module_rpc_api_consts.ListenPort,
		)
	}
	portPublishSpec := docker_manager.NewNoPublishingSpec()
	if launcher.shouldPublishPorts {
		portPublishSpec = docker_manager.NewAutomaticPublishingSpec()
	}
	usedPorts := map[nat.Port]docker_manager.PortPublishSpec {
		portObj: portPublishSpec,
	}

	ipAddr, err := launcher.freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return "", nil, nil, nil, stacktrace.Propagate(err, "An error occurred getting a free IP address for new module")
	}

	apiContainerSocket := fmt.Sprintf("%v:%v", launcher.apiContainerIpAddr, kurtosis_core_rpc_api_consts.ListenPort)
	envVars := map[string]string{
		kurtosis_module_docker_api.ApiContainerSocketEnvVar: apiContainerSocket,
		kurtosis_module_docker_api.SerializedCustomParamsEnvVar: serializedParams,
	}

	volumeMounts := map[string]string{
		launcher.enclaveDataVolName: kurtosis_module_docker_api.ExecutionVolumeMountpoint,
	}

	suffix := current_time_str_provider.GetCurrentTimeStr()
	moduleGUID :=  module_store_types.ModuleGUID(string(moduleID) + "_" + suffix)

	containerName := launcher.enclaveObjNameProvider.ForModuleContainer(moduleGUID)
	containerLabels := launcher.enclaveObjLabelsProvider.ForModuleContainer(moduleGUID)
	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		containerImage,
		containerName,
		launcher.dockerNetworkId,
	).WithAlias(
		containerName,
	).WithStaticIP(
		ipAddr,
	).WithUsedPorts(
		usedPorts,
	).WithEnvironmentVariables(
		envVars,
	).WithVolumeMounts(
		volumeMounts,
	).WithLabels(
		containerLabels,
	).Build()
	containerId, allHostPortBindings, err := launcher.dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return "", nil, nil, nil, stacktrace.Propagate(err, "An error occurred launching the module container")
	}
	shouldDestroyContainer := true
	defer func() {
		if shouldDestroyContainer {
			if err := launcher.dockerManager.KillContainer(context.Background(), containerId); err != nil {
				logrus.Error("Launching the module container failed, but an error occurred killing container we started:")
				fmt.Fprintln(logrus.StandardLogger().Out, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually kill container with ID '%v'", containerId)
			}
		}
	}()

	var resultHostPortBinding *nat.PortBinding = nil
	hostPortBindingFromMap, found := allHostPortBindings[portObj]
	if found {
		resultHostPortBinding = hostPortBindingFromMap
	}

	moduleSocket := fmt.Sprintf("%v:%v", ipAddr, kurtosis_module_rpc_api_consts.ListenPort)
	conn, err := grpc.Dial(
		moduleSocket,
		grpc.WithInsecure(), // TODO SECURITY: Use HTTPS to verify we're connecting to the correct module
	)
	if err != nil {
		return "", nil, nil, nil, stacktrace.Propagate(err, "Couldn't dial module container '%v' at %v", moduleID, moduleSocket)
	}
	moduleClient := kurtosis_module_rpc_api_bindings.NewExecutableModuleServiceClient(conn)

	logrus.Debugf("Waiting for module container to become available...")
	if err := waitUntilModuleContainerIsAvailable(ctx, moduleClient); err != nil {
		return "", nil, nil, nil, stacktrace.Propagate(err, "An error occurred while waiting for module container '%v' to become available", moduleID)
	}
	logrus.Debugf("Module container '%v' became available", moduleID)

	shouldDestroyContainer = false
	return containerId, ipAddr, moduleClient, resultHostPortBinding, nil
}

// ==========================================================================================
//                                   Private helper functions
// ==========================================================================================
func waitUntilModuleContainerIsAvailable(ctx context.Context, client kurtosis_module_rpc_api_bindings.ExecutableModuleServiceClient) error {
	contextWithTimeout, cancelFunc := context.WithTimeout(ctx, waitForModuleAvailabilityTimeout)
	defer cancelFunc()
	if _, err := client.IsAvailable(contextWithTimeout, &emptypb.Empty{}, grpc.WaitForReady(true)); err != nil {
		return stacktrace.Propagate(err, "An error occurred waiting for the module container to become available")
	}
	return nil
}
