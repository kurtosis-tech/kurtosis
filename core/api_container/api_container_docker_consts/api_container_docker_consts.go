/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package api_container_docker_consts

const (

	// TODO Push inside the server
	// The port that the API container will listen on (hardcoded, because it runs in a Docker container so no real
	//  reason to make configurable)

	// The location where the suite execution Docker volume will be mounted on the API container
	SuiteExecutionVolumeMountDirpath = "/suite-execution"
)
