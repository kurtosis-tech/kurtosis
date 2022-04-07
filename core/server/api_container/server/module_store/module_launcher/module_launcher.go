/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package module_launcher

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/module"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/free-ip-addr-tracker-lib/lib"
	"github.com/kurtosis-tech/kurtosis-core/api/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-core/api/golang/module_launch_api"
	"github.com/kurtosis-tech/kurtosis-core/server/commons/current_time_str_provider"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"time"
)

const (
	waitForModuleAvailabilityTimeout = 10 * time.Second

	modulePortNum      = uint16(1111)
	modulePortProtocol = port_spec.PortProtocol_TCP

	// The filepath where the enclave data directory will be mounted on the module container
	enclaveDataDirMountFilepathOnContainer = "/kurtosis-enclave-data"

	// This should probably be configurable
	shouldPullContainerImageBeforeStarting = true

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
	enclaveId enclave.EnclaveID

	kurtosisBackend backend_interface.KurtosisBackend

	// Modules have a connection to the API container, so the launcher must know what socket to pass to modules
	apiContainerSocketInsideNetwork string

	enclaveDataDirPath string
	freeIpAddrTracker  *lib.FreeIpAddrTracker
}

func NewModuleLauncher(enclaveId enclave.EnclaveID, kurtosisBackend backend_interface.KurtosisBackend, apiContainerSocketInsideNetwork string, enclaveDataDirPath string, freeIpAddrTracker *lib.FreeIpAddrTracker) *ModuleLauncher {
	return &ModuleLauncher{
		enclaveId:                       enclaveId,
		kurtosisBackend:                 kurtosisBackend,
		apiContainerSocketInsideNetwork: apiContainerSocketInsideNetwork,
		freeIpAddrTracker:               freeIpAddrTracker,
		enclaveDataDirPath:              enclaveDataDirPath,
	}
}

func (launcher ModuleLauncher) Launch(
	ctx context.Context,
	moduleID module.ModuleID,
	containerImage string,
	serializedParams string,
) (
	resultModule *module.Module,
	resultClient kurtosis_core_rpc_api_bindings.ExecutableModuleServiceClient,
	resultErr error,
) {
	suffix := current_time_str_provider.GetCurrentTimeStr()
	moduleGUID := module.ModuleGUID(string(moduleID) + "-" + suffix)

	privateIpAddr, err := launcher.freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting a free IP address for new module")
	}

	args := module_launch_api.NewModuleContainerArgs(
		string(launcher.enclaveId),
		modulePortNum,
		launcher.apiContainerSocketInsideNetwork,
		serializedParams,
		enclaveDataDirMountFilepathOnContainer,
	)
	envVars, err := module_launch_api.GetEnvFromArgs(args)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting the module container environment variables from args '%+v'", args)
	}

	createdModule, err := launcher.kurtosisBackend.CreateModule(
		ctx,
		containerImage,
		launcher.enclaveId,
		moduleID,
		moduleGUID,
		privateIpAddr,
		modulePortNum,
		launcher.enclaveDataDirPath,
		envVars,
	)

	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred launching  for module '%v'", containerImage)
	}
	shouldKillContainer := true
	defer func() {
		if shouldKillContainer {
			_, failedModules, err := launcher.kurtosisBackend.DestroyModules(ctx, getModuleByModuleGUIDFilter(moduleGUID))
			if err != nil {
				logrus.Error("Launching the module failed, but an error occurred calling the backend to kill the module we started:")
				fmt.Fprintln(logrus.StandardLogger().Out, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually kill the module with GUID '%v'", moduleGUID)
			}
			if len(failedModules) > 0 {
				for failedModuleGUID, failedModuleErr := range failedModules {
					logrus.Error("Launching the module failed, but an error occurred killing module we started:")
					fmt.Fprintln(logrus.StandardLogger().Out, failedModuleErr)
					logrus.Errorf("ACTION REQUIRED: You'll need to manually kill the module with GUID '%v'", failedModuleGUID)
				}
			}
		}
	}()

	moduleSocket := fmt.Sprintf("%v:%v", privateIpAddr, modulePortNum)
	conn, err := grpc.Dial(
		moduleSocket,
		grpc.WithInsecure(), // TODO SECURITY: Use HTTPS to verify we're connecting to the correct module
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Couldn't dial module container '%v' at %v", moduleID, moduleSocket)
	}
	moduleClient := kurtosis_core_rpc_api_bindings.NewExecutableModuleServiceClient(conn)

	logrus.Debugf("Waiting for module container to become available...")
	if err := waitUntilModuleContainerIsAvailable(ctx, moduleClient); err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred while waiting for module '%v' to become available", moduleID)
	}

	shouldKillContainer = false
	return createdModule, moduleClient, nil
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

func getModuleByModuleGUIDFilter(guid module.ModuleGUID) *module.ModuleFilters {
	return &module.ModuleFilters{GUIDs: map[module.ModuleGUID]bool{guid: true}}
}
