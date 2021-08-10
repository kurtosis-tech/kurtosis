/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package api_container_env_vars

// Constants for the environment variables that are used in the Dockerfile, made available to Go code
const (
	LogLevelEnvVar                 = "LOG_LEVEL"

	// JSON-serialized string containing all the parameters that the API container needs to start
	// We JSON-serilaize these values rather than passing them in individually as separate env vars so that
	//  1) We don't need to modify the Dockerfile every time we add an env var
	//  2) We don't need to modify the API container's main.go flag parsing whenever we add an env var
	ParamsJsonEnvVar			   = "PARAMS_JSON"
)
