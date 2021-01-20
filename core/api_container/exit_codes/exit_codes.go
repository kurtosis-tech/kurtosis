/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package exit_codes

const (
	Success = iota
	StartupErrorExitCode	    	// The API container hit an error while starting up
	ShutdownErrorExitCode			// The API container encountered errors during shutodwn
	// =============================== Test Execution exit codes ================================================
	// NOTE: If you add new test execution exit codes, make sure to modify the test_executor who consumes them!!
	NoTestSuiteRegisteredExitCode
	TestHitTimeoutExitCode
	ReceivedTermSignalExitCode
	OutOfOrderTestStatusExitCode
)
