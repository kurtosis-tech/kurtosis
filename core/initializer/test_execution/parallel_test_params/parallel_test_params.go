/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package parallel_test_params

/*
Package struct containing the parameters for running a test
*/
type ParallelTestParams struct {
	// Name of the test to run
	TestName            string

	TestSetupTimeoutSeconds uint32

	TestRunTimeoutSeconds uint32

	IsPartitioningEnabled bool

	// The number of bits of variability in the subnet mask of the network this test will run in
	// The number of IPs that the network can fit will be 2 ^ network_width_bits
	NetworkWidthBits uint32
}

func NewParallelTestParams(testName string, testSetupTimeoutSeconds uint32, testRunTimeoutSeconds uint32, isPartitioningEnabled bool, networkWidthBits uint32) *ParallelTestParams {
	return &ParallelTestParams{TestName: testName, TestSetupTimeoutSeconds: testSetupTimeoutSeconds, TestRunTimeoutSeconds: testRunTimeoutSeconds, IsPartitioningEnabled: isPartitioningEnabled, NetworkWidthBits: networkWidthBits}
}
