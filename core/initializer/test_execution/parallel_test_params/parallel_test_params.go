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

	// Subnet mask that should be used for the Docker network that the test controller & network will run in
	SubnetMask          string

	TestSetupTimeoutSeconds uint32

	TestRunTimeoutSeconds uint32

	IsPartitioningEnabled bool
}

func NewParallelTestParams(testName string, subnetMask string, testSetupTimeoutSeconds uint32, testRunTimeoutSeconds uint32, isPartitioningEnabled bool) *ParallelTestParams {
	return &ParallelTestParams{TestName: testName, SubnetMask: subnetMask, TestSetupTimeoutSeconds: testSetupTimeoutSeconds, TestRunTimeoutSeconds: testRunTimeoutSeconds, IsPartitioningEnabled: isPartitioningEnabled}
}
