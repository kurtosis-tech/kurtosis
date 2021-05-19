/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package module_store

import (
	"context"
	"github.com/docker/docker/api/types/container"
	"github.com/google/uuid"
	"github.com/kurtosis-tech/kurtosis/api_container/server/module_store/module_launcher"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network/container_name_provider"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/palantir/stacktrace"
	"net"
	"sync"
)

type ModuleID string

type moduleInfo struct {
	ipAddr net.IP

}

type ModuleStore struct {
	mutex *sync.Mutex

	// module_id -> module_ip_addr
	moduleIpAddrs map[ModuleID]string

	// module_id -> container_id
	moduleContainerIds map[ModuleID]string

	moduleLauncher *module_launcher.ModuleLauncher
}

// TODO Constructor

func (store *ModuleStore) LoadModule(ctx context.Context, containerImage string, paramsJsonStr string) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	moduleIdUuid, err := uuid.NewUUID()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred generating a UUID for module with image '%v' and params JSON '%v'", containerImage, paramsJsonStr)
	}
	moduleId := ModuleID(moduleIdUuid.String())

	containerId, containerIpAddr, usedHostPortBindings, err := store.moduleLauncher.Launch(ctx, moduleId, containerImage, paramsJsonStr)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred launching module from container image '%v' and params JSON string '%v'",
			containerImage,
			paramsJsonStr,
		)
	}



}

