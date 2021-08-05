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
}

func NewParallelTestParams(testName string, testSetupTimeoutSeconds uint32, testRunTimeoutSeconds uint32, isPartitioningEnabled bool) *ParallelTestParams {
	return &ParallelTestParams{TestName: testName, TestSetupTimeoutSeconds: testSetupTimeoutSeconds, TestRunTimeoutSeconds: testRunTimeoutSeconds, IsPartitioningEnabled: isPartitioningEnabled}
}
