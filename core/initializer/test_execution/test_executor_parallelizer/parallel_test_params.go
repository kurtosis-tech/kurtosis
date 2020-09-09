/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_executor_parallelizer

/*
Package struct containing the parameters for running a test
*/
type ParallelTestParams struct {
	// Name of the test to run
	TestName            string

	// Subnet mask that should be used for the Docker network that the test controller & network will run in
	SubnetMask          string
}

func NewParallelTestParams(testName string, subnetMask string) *ParallelTestParams {
	return &ParallelTestParams{
		TestName: testName,
		SubnetMask: subnetMask,
	}
}
