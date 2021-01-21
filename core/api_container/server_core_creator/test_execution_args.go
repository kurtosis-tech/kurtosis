/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package server_core_creator

// Fields are public for JSON de/serialization
type TestExecutionArgs struct {
	ExecutionInstanceId      string	`json:"executionInstanceId"`
	NetworkId                string `json:"networkId"`
	SubnetMask               string	`json:"subnetMask"`
	GatewayIpAddr            string	`json:"gatewayIpAddr"`
	TestName                 string	`json:"testName"`
	SuiteExecutionVolumeName string	`json:"suiteExecutionVolumeName"`
	TestSuiteContainerId     string	`json:"testSuiteContainerId"`

	// It seems weird that we require this given that the test suite container doesn't run a server, but it's only so
	//  that our free IP address tracker knows not to dole out the test suite container's IP address
	TestSuiteContainerIpAddr string	`json:"testSuiteContainerIpAddr"`
	ApiContainerIpAddr       string	`json:"apiContainerIpAddr"`

	// TODO can we remove this by having suite register itself with its metadata?? Or by passing over the suite
	//  metadata?
	IsPartitioningEnabled bool	`json:"isPartitioningEnabled"`
}
