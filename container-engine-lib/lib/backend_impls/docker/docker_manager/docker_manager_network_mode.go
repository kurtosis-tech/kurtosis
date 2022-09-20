/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package docker_manager

import "github.com/docker/docker/api/types/container"

const (
	defaultNetworkModeStr = "default"
)

type DockerManagerNetworkMode container.NetworkMode

var DefaultNetworkMode = DockerManagerNetworkMode(defaultNetworkModeStr)

func NewContainerNetworkMode(containerId string) DockerManagerNetworkMode {
	str := "container:" + containerId
	return DockerManagerNetworkMode(str)
}
