/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package module_store

import (
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-core/api_container/server/module_store/module_launcher"
	"github.com/kurtosis-tech/kurtosis-core/api_container/server/module_store/module_store_types"
	"github.com/kurtosis-tech/kurtosis-module-api-lib/golang/kurtosis_module_rpc_api_bindings"
	"github.com/palantir/stacktrace"
	"net"
	"sync"
)

type moduleInfo struct {
	containerId           string
	ipAddr                net.IP
	client                kurtosis_module_rpc_api_bindings.ExecutableModuleServiceClient
	hostPortBinding *nat.PortBinding
}

type ModuleStore struct {
	mutex *sync.Mutex

	dockerManager *docker_manager.DockerManager

	// module_id -> IP addr, container ID, etc.
	modules map[module_store_types.ModuleID]moduleInfo

	moduleLauncher *module_launcher.ModuleLauncher
}

func NewModuleStore(dockerManager *docker_manager.DockerManager, moduleLauncher *module_launcher.ModuleLauncher) *ModuleStore {
	return &ModuleStore{
		mutex:          &sync.Mutex{},
		dockerManager:  dockerManager,
		modules:        map[module_store_types.ModuleID]moduleInfo{},
		moduleLauncher: moduleLauncher,
	}
}

func (store *ModuleStore) LoadModule(ctx context.Context, moduleId module_store_types.ModuleID, containerImage string, serializedParams string) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if _, found := store.modules[moduleId]; found {
		return stacktrace.NewError("Module ID '%v' already exists in the module map", moduleId)
	}

	// NOTE: We don't use module host port bindings for now; we could expose them in the future if it's useful
	containerId, containerIpAddr, client, hostPortBinding, err := store.moduleLauncher.Launch(ctx, moduleId, containerImage, serializedParams)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred launching module from container image '%v' and serialized params '%v'",
			containerImage,
			serializedParams,
		)
	}

	infoForModule :=  moduleInfo{
		containerId: containerId,
		ipAddr: containerIpAddr,
		client: client,
		hostPortBinding: hostPortBinding,
	}

	store.modules[moduleId] = infoForModule
	return nil
}

func (store *ModuleStore) UnloadModule(ctx context.Context, moduleId module_store_types.ModuleID) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	 infoForModule, found := store.modules[moduleId]
	 if !found {
		return stacktrace.NewError("Module ID '%v' does not exist in the module map", moduleId)
	}

	containerId := infoForModule.containerId
	if err := store.dockerManager.KillContainer(ctx, containerId); err != nil {
		return  stacktrace.Propagate(err, "An error occurred killing module container '%v' while unloading the module from the store", moduleId)
	}

	delete(store.modules, moduleId)

	return nil
}

func (store *ModuleStore) ExecuteModule(ctx context.Context, moduleId module_store_types.ModuleID, serializedParams string) (serializedResult string, resultErr error) {
	// NOTE: technically we don't need this mutex for this function since we're not modifying internal state, but we do need it to check isDestroyed
	// TODO PERF: Don't block the entire store on executing a module
	store.mutex.Lock()
	defer store.mutex.Unlock()

	info, found := store.modules[moduleId]
	if !found {
		return "", stacktrace.NewError("No module '%v' exists in the module store", moduleId)
	}
	client := info.client
	args := &kurtosis_module_rpc_api_bindings.ExecuteArgs{ParamsJson: serializedParams}
	resp, err := client.Execute(ctx, args)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred executing module '%v' with serialized params '%v'", moduleId, serializedParams)
	}
	return resp.ResponseJson, nil
}

func (store *ModuleStore) GetModuleIPAddrByID(moduleId module_store_types.ModuleID) (net.IP, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	info, found := store.modules[moduleId]
	if !found {
		return nil, stacktrace.NewError("No module with ID '%v' has been loaded", moduleId)
	}
	return info.ipAddr, nil
}

func (store *ModuleStore) GetModules() map[module_store_types.ModuleID]bool {

	moduleIDs := make(map[module_store_types.ModuleID]bool, len(store.modules))

	for key, _ := range store.modules {
		if _, ok := moduleIDs[key]; !ok{
			moduleIDs[key] = true
		}
	}
	return moduleIDs
}
