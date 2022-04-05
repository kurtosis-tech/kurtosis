/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package module_store

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/module"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis-core/api/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/module_store/module_launcher"
	"github.com/kurtosis-tech/stacktrace"
	"net"
	"strings"
	"sync"
)

type moduleInfo struct {
	moduleGUID    module.ModuleGUID
	privateIpAddr net.IP
	privatePort   *port_spec.PortSpec
	publicIpAddr  net.IP
	publicPort    *port_spec.PortSpec
	// NOTE: When we want restart-able enclaves, we'll need to not store this client and instead recreate one
	//  from the port num/protocol stored as a label on the module container
	client kurtosis_core_rpc_api_bindings.ExecutableModuleServiceClient
}

type ModuleStore struct {
	mutex *sync.Mutex

	kurtosisBackend backend_interface.KurtosisBackend

	// module_id -> IP addr, container ID, etc.
	modules map[module.ModuleID]moduleInfo

	moduleLauncher *module_launcher.ModuleLauncher
}

func NewModuleStore(kurtosisBackend backend_interface.KurtosisBackend, moduleLauncher *module_launcher.ModuleLauncher) *ModuleStore {
	return &ModuleStore{
		mutex:           &sync.Mutex{},
		kurtosisBackend: kurtosisBackend,
		modules:         map[module.ModuleID]moduleInfo{},
		moduleLauncher:  moduleLauncher,
	}
}

func (store *ModuleStore) LoadModule(
	ctx context.Context,
	moduleId module.ModuleID,
	containerImage string,
	serializedParams string,
) (
	resultPrivateIp net.IP,
	resultPrivatePort *port_spec.PortSpec,
	resultPublicIp net.IP,
	resultPublicPort *port_spec.PortSpec,
	resultErr error,
) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if _, found := store.modules[moduleId]; found {
		return nil, nil, nil, nil, stacktrace.NewError("Module ID '%v' already exists in the module map", moduleId)
	}

	launchedModule, launchedModuleClient, err := store.moduleLauncher.Launch(
		ctx,
		moduleId,
		containerImage,
		serializedParams,
	)
	if err != nil {
		return nil, nil, nil, nil, stacktrace.Propagate(
			err,
			"An error occurred launching module from container image '%v' and serialized params '%v'",
			containerImage,
			serializedParams,
		)
	}

	privateIpAddr := launchedModule.GetPrivateIp()
	publicIpAddr := launchedModule.GetPublicIp()
	privatePort := launchedModule.GetPrivatePort()
	publicPort := launchedModule.GetPublicPort()
	infoForModule := moduleInfo{
		moduleGUID:    launchedModule.GetGUID(),
		privateIpAddr: privateIpAddr,
		privatePort:   privatePort,
		publicIpAddr:  publicIpAddr,
		publicPort:    publicPort,
		client:        launchedModuleClient,
	}

	store.modules[moduleId] = infoForModule
	return privateIpAddr, privatePort, publicIpAddr, publicPort, nil
}

func (store *ModuleStore) UnloadModule(ctx context.Context, moduleId module.ModuleID) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	infoForModule, found := store.modules[moduleId]
	if !found {
		return stacktrace.NewError("Module ID '%v' does not exist in the module map", moduleId)
	}

	moduleGuid := infoForModule.moduleGUID
	_, failedToDestroyModules, err := store.kurtosisBackend.DestroyModules(ctx, getModuleByModuleGUIDFilter(moduleGuid))
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred killing module container '%v' while unloading the module from the store", moduleId)
	}
	if len(failedToDestroyModules) > 0 {
		moduleStopErrs := []string{}
		for moduleGUID, err := range failedToDestroyModules {
			wrappedErr := stacktrace.Propagate(
				err,
				"An error occurred stopping module `%v'",
				moduleGUID,
			)
			moduleStopErrs = append(moduleStopErrs, wrappedErr.Error())
		}
		return stacktrace.NewError(
			"One or more errors occurred stopping the module(s):\n%v",
			strings.Join(
				moduleStopErrs,
				"\n\n",
			),
		)
	}
	delete(store.modules, moduleId)

	return nil
}

func (store *ModuleStore) ExecuteModule(ctx context.Context, moduleId module.ModuleID, serializedParams string) (serializedResult string, resultErr error) {
	// NOTE: technically we don't need this mutex for this function since we're not modifying internal state, but we do need it to check isDestroyed
	// TODO PERF: Don't block the entire store on executing a module
	store.mutex.Lock()
	defer store.mutex.Unlock()

	info, found := store.modules[moduleId]
	if !found {
		return "", stacktrace.NewError("No module '%v' exists in the module store", moduleId)
	}
	client := info.client
	args := &kurtosis_core_rpc_api_bindings.ExecuteArgs{ParamsJson: serializedParams}
	resp, err := client.Execute(ctx, args)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred executing module '%v' with serialized params '%v'", moduleId, serializedParams)
	}
	return resp.ResponseJson, nil
}

func (store *ModuleStore) GetModuleInfo(moduleId module.ModuleID) (
	resultPrivateIp net.IP,
	resultPrivatePort *port_spec.PortSpec,
	resultPublicIp net.IP,
	resultPublicPort *port_spec.PortSpec,
	resultErr error,
) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	info, found := store.modules[moduleId]
	if !found {
		return nil, nil, nil, nil, stacktrace.NewError("No module with ID '%v' has been loaded", moduleId)
	}
	return info.privateIpAddr, info.privatePort, info.publicIpAddr, info.publicPort, nil
}

func (store *ModuleStore) GetModules() map[module.ModuleID]bool {

	moduleIDs := make(map[module.ModuleID]bool, len(store.modules))

	for key, _ := range store.modules {
		if _, ok := moduleIDs[key]; !ok {
			moduleIDs[key] = true
		}
	}
	return moduleIDs
}

func getModuleByModuleGUIDFilter(guid module.ModuleGUID) *module.ModuleFilters {
	return &module.ModuleFilters{GUIDs: map[module.ModuleGUID]bool{guid: true}}
}
