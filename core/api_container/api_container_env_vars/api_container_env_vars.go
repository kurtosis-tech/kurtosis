/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package api_container_env_vars

// Constants for the environment variables that are used in the Dockerfile, made available to Go code
const (
	ApiContainerIpAddrEnvVar       = "API_CONTAINER_IP"
	ApiLogFilepathEnvVar           = "LOG_FILEPATH"
	ExecutionInstanceIdEnvVar      = "EXECUTION_INSTANCE_ID"
	GatewayIpEnvVar                = "GATEWAY_IP"
	IsPartitioningEnabledEnvVar    = "IS_PARTITIONING_ENABLED"
	LogLevelEnvVar                 = "LOG_LEVEL"

	// Indicates the mode that the API container should be running in
	ModeEnvVar					   = "MODE"

	NetworkIdEnvVar                = "NETWORK_ID"

	ParamsJsonEnvVar			   = "PARAMS_JSON"




	SubnetMaskEnvVar               = "SUBNET_MASK"
	TestNameEnvVar                 = "TEST_NAME"
	TestSuiteContainerIdEnvVar     = "TEST_SUITE_CONTAINER_ID"
	TestSuiteContainerIpAddrEnvVar = "TEST_SUITE_CONTAINER_IP"
	TestVolumeNameEnvVar           = "TEST_VOLUME"

	// TODO break into separate file???
	SuiteMetadataPrintingMode = "PRINT_SUITE_METADATA"
	TestExecutionMode = "EXECUTE_TEST"
)

var AllModes = map[string]bool{
	SuiteMetadataPrintingMode: true,
	TestExecutionMode: true,
}
