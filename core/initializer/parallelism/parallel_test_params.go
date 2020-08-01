package parallelism

import (
	"github.com/docker/distribution/uuid"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
	"os"
)

/*
Package struct containing the parameters for running a test
*/
type ParallelTestParams struct {
	// Name of the test to run
	TestName            string

	// Logic of the test to run
	Test 				testsuite.Test

	// FP of the file where the test's log should be written to
	LogFp               *os.File

	// Subnet mask that should be used for the Docker network that the test controller & network will run in
	SubnetMask          string

	// UUID representing an a single execution of one or more tests from the test suite, to which this test execution belongs
	ExecutionInstanceId uuid.UUID
}

func NewParallelTestParams(testName string, test testsuite.Test, logFp *os.File, subnetMask string, executionInstanceId uuid.UUID) *ParallelTestParams {
	return &ParallelTestParams{TestName: testName, Test: test, LogFp: logFp, SubnetMask: subnetMask, ExecutionInstanceId: executionInstanceId}
}
