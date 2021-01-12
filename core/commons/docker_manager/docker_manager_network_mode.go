/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
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
