/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package module_store

import (
	"context"
	"github.com/docker/docker/api/types/container"
	"github.com/google/uuid"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network/container_name_provider"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/palantir/stacktrace"
	"sync"
)

type ModuleID string

type ModuleStore struct {
	mutex *sync.Mutex

	// module_id -> module_ip_addr
	moduleIpAddrs map[ModuleID]string

	// module_id -> container_id
	moduleContainerIds map[ModuleID]string

	dockerManager *docker_manager.DockerManager

	containerNameElemProvider *container_name_provider.ContainerNameElementsProvider

	freeIpAddrTracker *commons.FreeIpAddrTracker
}

// TODO Constructor

func (store *ModuleStore) LoadModule(ctx context.Context, containerImage string, paramsJsonStr string) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()



}

