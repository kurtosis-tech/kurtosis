/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package api_container_env_vars

// Constants for the environment variables that are used in the Dockerfile, made available to Go code
const (
	LogLevelEnvVar                 = "LOG_LEVEL"
	ModeEnvVar					   = "MODE"
	ParamsJsonEnvVar			   = "PARAMS_JSON"

	// TODO break into separate file???
	SuiteMetadataPrintingMode = "PRINT_SUITE_METADATA"
	TestExecutionMode = "EXECUTE_TEST"
)

var AllModes = map[string]bool{
	SuiteMetadataPrintingMode: true,
	TestExecutionMode: true,
}
