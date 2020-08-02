package parallelism

import "os"

/*
Package struct containing the output of a single test that was run in parallel
*/
type ParallelTestOutput struct {
	// Name of the test that was run
	TestName     string

	// Indicates whether an error occurred during the execution of the test that prevented it from running
	ExecutionErr error

	// Indicates whether the test passed or failed (undefined if the test had a setup error)
	TestPassed   bool

	// FP of the file where the test's logs were written to
	LogFp        *os.File
}
