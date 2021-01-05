/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package api_container_env_vars

// Constants for the environment variables that are used in the Dockerfile, made available to Go code
const (
	ApiContainerIpAddrEnvVar       = "API_CONTAINER_IP"
	ApiLogFilepathEnvVar           = "LOG_FILEPATH"
	GatewayIpEnvVar                = "GATEWAY_IP"
	IsPartitioningEnabledEnvVar    = "IS_PARTITIONING_ENABLED"
	LogLevelEnvVar                 = "LOG_LEVEL"
	NetworkIdEnvVar                = "NETWORK_ID"
	SubnetMaskEnvVar               = "SUBNET_MASK"
	TestSuiteContainerIdEnvVar     = "TEST_SUITE_CONTAINER_ID"
	TestSuiteContainerIpAddrEnvVar = "TEST_SUITE_CONTAINER_IP"
	TestVolumeName                 = "TEST_VOLUME"
)
