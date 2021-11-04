/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package api_container_docker_consts

const (

	// The location where the directory on the Docker host machine for storing enclave data will be bind-mounted
	//  on the API container
	// This COULD possibly vary across launcher API verisons, but we can cross that bridge when we come to it
	EnclaveDataDirMountpoint = "/kurtosis-enclave-data"
)
