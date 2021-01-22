/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package exit_codes

type ApiContainerExitCode int

// TODO Refactor this to be a class with a visitor, so all versions MUST be handled!!!!
const (
	SuccessExitCode       ApiContainerExitCode = iota
	NoTestSuiteRegisteredExitCode
	ShutdownEventBeforeSuiteRegistration	   // Service sends a shutdown exit code before the testsuite is registered
	StartupErrorExitCode                       // The API container hit an error while starting up
	ShutdownErrorExitCode                      // The API container encountered errors during shutodwn
	// =============================== Test Execution exit codes ================================================
	// NOTE: If you add new test execution exit codes, make sure to modify the test_executor who consumes them!!
	NoTestExecutionRegisteredExitCode	// A testsuite registered itself, but then didn't register a test execution
	TestHitTimeoutExitCode
	ReceivedTermSignalExitCode
	ErrWaitingForSuiteContainerExitExitCode // An error occurred waiting for the testsuite container to exit
	OutOfOrderTestStatusExitCode
)
