/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package api_container_exit_codes

// Thanks to Go's absolutely stupid decision not to have enums, this is our best way to
//  have an enum value that also enforces handling all values at compile-time
const (
	// =============================== Commmon exit codes ======================================
	SuccessfulExit = iota
	NoTestSuiteRegistered
	ShutdownEventBeforeSuiteRegistration	   // Service sends a shutdown exit code before the testsuite is registered
	StartupError                       // The API container hit an error while starting up
	ShutdownError                      // The API container encountered errors during shutodwn
	ReceivedTermSignal
	// =============================== Test Execution exit codes ================================================
	// NOTE: If you add new test execution exit codes, make sure to modify the test_executor who consumes them!!
	NoTestExecutionRegistered	// A testsuite registered itself, but then didn't register a test execution
	TestHitTimeout
	ErrWaitingForSuiteContainerExit // An error occurred waiting for the testsuite container to exit
)
var ExitCodeErrorVisitorAcceptFuncs = map[int]func(visitor ExitCodeErrorVisitor) error {
	SuccessfulExit:  func(visitor ExitCodeErrorVisitor) error { return visitor.VisitSuccessfulExit() },
	NoTestSuiteRegistered: func(visitor ExitCodeErrorVisitor) error { return visitor.VisitNoTestSuiteRegistered() },
	ShutdownEventBeforeSuiteRegistration: func(visitor ExitCodeErrorVisitor) error { return visitor.VisitNoTestSuiteRegistered() },
	StartupError: func(visitor ExitCodeErrorVisitor) error { return visitor.VisitStartupError() },
	ShutdownError: func(visitor ExitCodeErrorVisitor) error { return visitor.VisitShutdownError() },
	ReceivedTermSignal: func(visitor ExitCodeErrorVisitor) error { return visitor.VisitReceivedTermSignal() },
	NoTestExecutionRegistered: func(visitor ExitCodeErrorVisitor) error { return visitor.VisitNoTestExecutionRegistered() },
	TestHitTimeout: func(visitor ExitCodeErrorVisitor) error { return visitor.VisitTestHitTimeout() },
	ErrWaitingForSuiteContainerExit: func(visitor ExitCodeErrorVisitor) error { return visitor.VisitErrWaitingForSuiteContainerExit() },
}

// Translates exit codes into Go 'error' types
type ExitCodeErrorVisitor interface {
	VisitSuccessfulExit() error
	VisitNoTestSuiteRegistered() error
	VisitShutdownEventBeforeSuiteRegistration() error
	VisitStartupError() error
	VisitShutdownError() error
	VisitReceivedTermSignal() error
	VisitNoTestExecutionRegistered() error
	VisitTestHitTimeout() error
	VisitErrWaitingForSuiteContainerExit() error
}
