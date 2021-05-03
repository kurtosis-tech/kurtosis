/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_executor_parallelizer

import (
	"github.com/kurtosis-tech/kurtosis/test_suite/rpc_api/bindings"
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
	TestMetadata bindings.TestMetadata
}

func NewParallelTestParams(testName string, subnetMask string, testMetadata bindings.TestMetadata) *ParallelTestParams {
	return &ParallelTestParams{TestName: testName, SubnetMask: subnetMask, TestMetadata: testMetadata}
}

