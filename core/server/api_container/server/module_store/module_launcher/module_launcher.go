/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package module_launcher

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/free-ip-addr-tracker-lib/lib"
	"github.com/kurtosis-tech/kurtosis-core/api/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-core/api/golang/module_launch_api"
	"github.com/kurtosis-tech/kurtosis-core/launcher/enclave_container_launcher"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/module_store/module_store_types"
	"github.com/kurtosis-tech/kurtosis-core/server/commons/current_time_str_provider"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"net"
	"time"
)

const (
	waitForModuleAvailabilityTimeout = 10 * time.Second

	modulePortNum      = uint16(1111)
	modulePortProtocol = enclave_container_launcher.EnclaveContainerPortProtocol_TCP

	// The filepath where the enclave data directory will be mounted on the module container
	enclaveDataDirMountFilepathOnContainer = "/kurtosis-enclave-data"

	// For now, we let users update their module images if they want to, though this should probably be configurable
	shouldPullContainerImageBeforeStarting = false

	// Modules don't need to access the Docker engine directly, instead doing that through the API container
	shouldBindMountDockerSocket = false

	// Indicates that no alias should be set for the module
	moduleAlias = ""

)
// These values indicate "don't override the ENTRYPOINT/CMD args" (since modules are configured via envvars)
var entrypointArgs []string = nil
var cmdArgs []string = nil
var volumeMounts map[string]string = nil

type ModuleLauncher struct {
	// The ID of the enclave that the API container is running inside
	enclaveId string

	dockerManager *docker_manager.DockerManager

	// Modules have a connection to the API container, so the launcher must know what socket to pass to modules
	apiContainerSocketInsideNetwork string

	enclaveContainerLauncher *enclave_container_launcher.EnclaveContainerLauncher

	freeIpAddrTracker *lib.FreeIpAddrTracker

	dockerNetworkId string
}

func NewModuleLauncher(enclaveId string, dockerManager *docker_manager.DockerManager, apiContainerSocketInsideNetwork string, enclaveContainerLauncher *enclave_container_launcher.EnclaveContainerLauncher, freeIpAddrTracker *lib.FreeIpAddrTracker, dockerNetworkId string) *ModuleLauncher {
	return &ModuleLauncher{enclaveId: enclaveId, dockerManager: dockerManager, apiContainerSocketInsideNetwork: apiContainerSocketInsideNetwork, enclaveContainerLauncher: enclaveContainerLauncher, freeIpAddrTracker: freeIpAddrTracker, dockerNetworkId: dockerNetworkId}
}

func (launcher ModuleLauncher) Launch(
	ctx context.Context,
	moduleID module_store_types.ModuleID,
	containerImage string,
	serializedParams string,
) (
	resultContainerId string,
	resultPrivateIp net.IP,
	resultPrivatePort *enclave_container_launcher.EnclaveContainerPort,
	resultPublicIp net.IP,
	resultPublicPort *enclave_container_launcher.EnclaveContainerPort,
	client kurtosis_core_rpc_api_bindings.ExecutableModuleServiceClient,
	resultErr error,
) {
	privatePort, err := enclave_container_launcher.NewEnclaveContainerPort(modulePortNum, modulePortProtocol)
	if err != nil {
		return "", nil, nil, nil, nil, nil, stacktrace.Propagate(
			err,
			"Couldn't create module container port object using num '%v' and protocol '%v'",
			modulePortNum,
			modulePortProtocol,
		)
	}
	privatePorts := map[string]*enclave_container_launcher.EnclaveContainerPort{
		schema.KurtosisInternalContainerGRPCPortID: privatePort,
	}

	suffix := current_time_str_provider.GetCurrentTimeStr()
	moduleGUID :=  module_store_types.ModuleGUID(string(moduleID) + "_" + suffix)
	objAttrsSupplier := func(enclaveObjAttrsProvider schema.EnclaveObjectAttributesProvider) (schema.ObjectAttributes, error) {
		moduleContainerAttrs, err := enclaveObjAttrsProvider.ForModuleContainer(
			string(moduleGUID),
			modulePortNum,
		)
		if err != nil {
			return nil, stacktrace.Propagate(
				err,
				"An error occurred getting the module container object attributes using port num '%v'",
				modulePortNum,
			)
		}
		return moduleContainerAttrs, nil
	}

	privateIpAddr, err := launcher.freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return "", nil, nil, nil, nil, nil, stacktrace.Propagate(err, "An error occurred getting a free IP address for new module")
	}

	args := module_launch_api.NewModuleContainerArgs(
		launcher.enclaveId,
		modulePortNum,
		launcher.apiContainerSocketInsideNetwork,
		serializedParams,
		enclaveDataDirMountFilepathOnContainer,
	)
	envVars, err := module_launch_api.GetEnvFromArgs(args)
	if err != nil {
		return "", nil, nil, nil, nil, nil, stacktrace.Propagate(err, "An error occurred getting the module container environment variables from args '%+v'", args)
	}

	containerId, publicIpAddr, publicPorts, err := launcher.enclaveContainerLauncher.Launch(
		ctx,
		containerImage,
		shouldPullContainerImageBeforeStarting,
		privateIpAddr,
		launcher.dockerNetworkId,
		enclaveDataDirMountFilepathOnContainer,
		privatePorts,
		objAttrsSupplier,
		envVars,
		shouldBindMountDockerSocket,
		moduleAlias,
		entrypointArgs,
		cmdArgs,
		volumeMounts,
	)
	shouldKillContainer := true
	defer func() {
		if shouldKillContainer {
			if err := launcher.dockerManager.KillContainer(context.Background(), containerId); err != nil {
				logrus.Error("Launching the module container failed, but an error occurred killing container we started:")
				fmt.Fprintln(logrus.StandardLogger().Out, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually kill container with ID '%v'", containerId)
			}
		}
	}()

	publicPort, found := publicPorts[schema.KurtosisInternalContainerGRPCPortID]
	if !found {
		return "", nil, nil, nil, nil, nil, stacktrace.NewError(
			"Expected to find the module's public port information using port ID '%v', but none was found",
			schema.KurtosisInternalContainerGRPCPortID,
		)
	}

	moduleSocket := fmt.Sprintf("%v:%v", privateIpAddr, modulePortNum)
	conn, err := grpc.Dial(
		moduleSocket,
		grpc.WithInsecure(), // TODO SECURITY: Use HTTPS to verify we're connecting to the correct module
	)
	if err != nil {
		return "", nil, nil, nil, nil, nil, stacktrace.Propagate(err, "Couldn't dial module container '%v' at %v", moduleID, moduleSocket)
	}
	moduleClient := kurtosis_core_rpc_api_bindings.NewExecutableModuleServiceClient(conn)

	logrus.Debugf("Waiting for module container to become available...")
	if err := waitUntilModuleContainerIsAvailable(ctx, moduleClient); err != nil {
		return "", nil, nil, nil, nil, nil, stacktrace.Propagate(err, "An error occurred while waiting for module container '%v' to become available", moduleID)
	}
	logrus.Debugf("Module container '%v' became available", moduleID)

	shouldKillContainer = false
	return containerId, privateIpAddr, privatePort, publicIpAddr, publicPort, moduleClient, nil
}

// ==========================================================================================
//                                   Private helper functions
// ==========================================================================================
func waitUntilModuleContainerIsAvailable(ctx context.Context, client kurtosis_core_rpc_api_bindings.ExecutableModuleServiceClient) error {
	contextWithTimeout, cancelFunc := context.WithTimeout(ctx, waitForModuleAvailabilityTimeout)
	defer cancelFunc()
	if _, err := client.IsAvailable(contextWithTimeout, &emptypb.Empty{}, grpc.WaitForReady(true)); err != nil {
		return stacktrace.Propagate(err, "An error occurred waiting for the module container to become available")
	}
	return nil
}
