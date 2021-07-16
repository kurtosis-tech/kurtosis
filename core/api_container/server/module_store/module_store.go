/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package module_store

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api_container/server/module_store/module_launcher"
	"github.com/kurtosis-tech/kurtosis/api_container/server/module_store/module_store_types"
	"github.com/palantir/stacktrace"
	"net"
	"sync"
)

type moduleType string

const (
	lambdaModuleType moduleType = "LAMBDA"
)

type moduleInfo struct {
	containerId string
	ipAddr net.IP
	moduleType moduleType
}

type ModuleContext struct {
	id     module_store_types.ModuleID
	ipAddr net.IP
}

type ModuleStore struct {
	mutex *sync.Mutex

	// module_id -> IP addr, container ID, etc.
	moduleInfo map[module_store_types.ModuleID]moduleInfo

	moduleLauncher *module_launcher.ModuleLauncher
}

func NewModuleStore(moduleLauncher *module_launcher.ModuleLauncher) *ModuleStore {
	return &ModuleStore{
		mutex: &sync.Mutex{},
		moduleInfo: map[module_store_types.ModuleID]moduleInfo{},
		moduleLauncher: moduleLauncher,
	}
}

// Loads a module and returns its module ID, IP address, and any host port bindings
func (store *ModuleStore) LoadModule(ctx context.Context, moduleId module_store_types.ModuleID, containerImage string, paramsJsonStr string) (net.IP, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if _, found := store.moduleInfo[moduleId]; found {
		return nil, stacktrace.NewError("Module ID '%v' already exists in the module info map", moduleId)
	}

	// NOTE: We don't use module host port bindings for now; we could expose them in the future if it's useful
	containerId, containerIpAddr, _, err := store.moduleLauncher.Launch(ctx, moduleId, containerImage, paramsJsonStr)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred launching module from container image '%v' and params JSON string '%v'",
			containerImage,
			paramsJsonStr,
		)
	}
	moduleData := moduleInfo{
		containerId: containerId,
		ipAddr:      containerIpAddr,
		moduleType:  "",
	}
	store.moduleInfo[moduleId] = moduleData

	return containerIpAddr, nil
}

func (store *ModuleStore) ExecuteLambda(moduleId module_store_types.ModuleID, paramsJson string) (responseJson string, resultErr error) {
	// NOTE: No need to lock the mutex because we don't modify the internal state!
}
