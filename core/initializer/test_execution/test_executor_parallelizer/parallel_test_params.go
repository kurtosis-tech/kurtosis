/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_executor_parallelizer

import (
	"github.com/kurtosis-tech/kurtosis/initializer/test_suite_metadata_acquirer"
)

/*
Package struct containing the parameters for running a test
*/
type ParallelTestParams struct {
	// Name of the test to run
	TestName            string

	// Subnet mask that should be used for the Docker network that the test controller & network will run in
	SubnetMask          string

	// Special options declared by the test itself
	TestMetadata test_suite_metadata_acquirer.TestMetadata
}

func NewParallelTestParams(testName string, subnetMask string, testMetadata test_suite_metadata_acquirer.TestMetadata) *ParallelTestParams {
	return &ParallelTestParams{TestName: testName, SubnetMask: subnetMask, TestMetadata: testMetadata}
}
