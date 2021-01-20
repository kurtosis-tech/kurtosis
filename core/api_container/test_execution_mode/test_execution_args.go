/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_execution_mode

// Fields are public for JSON de/serialization
type TestExecutionArgs struct {
	ExecutionInstanceId      string
	NetworkId                string
	SubnetMask               string
	GatewayIpAddr            string
	TestName                 string
	SuiteExecutionVolumeName string
	TestSuiteContainerId     string

	// It seems weird that we require this given that the test suite container doesn't run a server, but it's only so
	//  that our free IP address tracker knows not to dole out the test suite container's IP address
	TestSuiteContainerIpAddr string
	ApiContainerIpAddr       string

	IsPartitioningEnabled bool
}

func NewTestExecutionArgs(executionInstanceId string, networkId string, subnetMask string, gatewayIpAddr string, testName string, suiteExecutionVolumeName string, testSuiteContainerId string, testSuiteContainerIpAddr string, apiContainerIpAddr string, isPartitioningEnabled bool) *TestExecutionArgs {
	return &TestExecutionArgs{ExecutionInstanceId: executionInstanceId, NetworkId: networkId, SubnetMask: subnetMask, GatewayIpAddr: gatewayIpAddr, TestName: testName, SuiteExecutionVolumeName: suiteExecutionVolumeName, TestSuiteContainerId: testSuiteContainerId, TestSuiteContainerIpAddr: testSuiteContainerIpAddr, ApiContainerIpAddr: apiContainerIpAddr, IsPartitioningEnabled: isPartitioningEnabled}
}

