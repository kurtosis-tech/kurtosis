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
	// =============================== Test Execution exit codes ================================================
	// NOTE: If you add new test execution exit codes, make sure to modify the test_executor who consumes them!!
	NoTestExecutionRegistered	// A testsuite registered itself, but then didn't register a test execution
	TestHitTimeout
	ReceivedTermSignal
	ErrWaitingForSuiteContainerExit // An error occurred waiting for the testsuite container to exit
	OutOfOrderTestStatus
)
var AcceptVisitorFuncs = map[int]func(visitor ErrorRenderingVisitor) error {
	SuccessfulExit:  func(visitor ErrorRenderingVisitor) error { return visitor.VisitSuccessfulExit() },
	NoTestSuiteRegistered: func(visitor ErrorRenderingVisitor) error { return visitor.VisitNoTestSuiteRegistered() },
	ShutdownEventBeforeSuiteRegistration: func(visitor ErrorRenderingVisitor) error { return visitor.VisitNoTestSuiteRegistered() },
	StartupError: func(visitor ErrorRenderingVisitor) error { return visitor.VisitStartupError() },
	ShutdownError: func(visitor ErrorRenderingVisitor) error { return visitor.VisitShutdownError() },
	NoTestExecutionRegistered: func(visitor ErrorRenderingVisitor) error { return visitor.VisitNoTestExecutionRegistered() },
	TestHitTimeout: func(visitor ErrorRenderingVisitor) error { return visitor.VisitTestHitTimeout() },
	ReceivedTermSignal: func(visitor ErrorRenderingVisitor) error { return visitor.VisitReceivedTermSignal() },
	ErrWaitingForSuiteContainerExit: func(visitor ErrorRenderingVisitor) error { return visitor.VisitErrWaitingForSuiteContainerExit() },
	OutOfOrderTestStatus: func(visitor ErrorRenderingVisitor) error { return visitor.VisitOutOfOrderTestStatus() },
}

// Translates exit codes into Go 'error' types
type ErrorRenderingVisitor interface {
	VisitSuccessfulExit() error
	VisitNoTestSuiteRegistered() error
	VisitShutdownEventBeforeSuiteRegistration() error
	VisitStartupError() error
	VisitShutdownError() error
	VisitNoTestExecutionRegistered() error
	VisitTestHitTimeout() error
	VisitReceivedTermSignal() error
	VisitErrWaitingForSuiteContainerExit() error
	VisitOutOfOrderTestStatus() error
}
