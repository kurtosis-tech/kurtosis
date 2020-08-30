/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package api_container_docker_consts

const (
	// The port that the API container will listen on (hardcoded, because it runs in a Docker container so no real
	//  reason to make configurable)
	ContainerPort = 7443

	// The location on the API container image where the API container will write its logs to (and which will be bound-
	//  mounted so the initializer can read it)
	LogMountFilepath = "/bind-mounts/api.log"

	// The location where the test volume will be mounted on the API container
	TestVolumeDirpath = "/volume-mounts/test-volume"
)
