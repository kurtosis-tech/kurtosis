/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package api_container_rpc_api_consts

const (
	// Protocol that the API container will listen on
	ListenProtocol = "tcp"

	// The port that the API container will listen on (hardcoded, because it runs in a Docker container so no real
	//  reason to make configurable)
	ListenPort = 7443
)
