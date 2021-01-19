/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package exit_codes

const (
	// NOTE: If you add new codes, make sure to modify the test_executor who consumes them!!
	TestCompletedInTimeoutExitCode = iota
	StartupErrorExitCode           	// The API container hit an error while starting up
	ShutdownErrorExitCode			// The API container encountered erros during shutodwn
	NoTestSuiteRegisteredExitCode
	TestHitTimeoutExitCode
	ReceivedTermSignalExitCode
	OutOfOrderTestStatusExitCode
)
