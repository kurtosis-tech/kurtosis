/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_executor

import "github.com/palantir/stacktrace"

type testExecutionExitCodeErrorVisitor struct {}

func (t testExecutionExitCodeErrorVisitor) VisitSuccessfulExit() error {
	return nil
}

func (t testExecutionExitCodeErrorVisitor) VisitNoTestSuiteRegistered() error {
	return stacktrace.NewError("The Kurtosis API container timed out waiting for a testsuite container to register " +
		"itself; this is a bug in Kurtosis itself")
}

func (t testExecutionExitCodeErrorVisitor) VisitShutdownEventBeforeSuiteRegistration() error {
	return stacktrace.NewError("The Kurtosis API's main service told itself to shut down before a testsuite container was " +
		"registered; this is a bug in Kurtosis itself")
}

func (t testExecutionExitCodeErrorVisitor) VisitStartupError() error {
	return stacktrace.NewError("The Kurtosis API container encountered an error while " +
		"starting up; this is a bug in Kurtosis itself")
}

func (t testExecutionExitCodeErrorVisitor) VisitShutdownError() error {
	return stacktrace.NewError("The Kurtosis API container encountered an error while " +
		"shutting down; this is a bug in Kurtosis itself")
}

func (t testExecutionExitCodeErrorVisitor) VisitReceivedTermSignal() error {
	return stacktrace.NewError("The Kurtosis API container exited due to receiving " +
		"a shutdown signal; if this is not expected, it indicates a bug in Kurtosis itself")
}

func (t testExecutionExitCodeErrorVisitor) VisitSerializeNotCalled() error {
	return stacktrace.NewError("The Kurtosis API exited with a 'serialize not called' error, indicating that " +
		"the Kurtosis API was in the wrong mode; this is a bug with Kurtosis itself")
}

func (t testExecutionExitCodeErrorVisitor) VisitNoTestSetupRegistered() error {
	return stacktrace.NewError("The Kurtosis API container timed out waiting for registration of a test setup; " +
		"this is a bug in Kurtosis itself")
}

func (t testExecutionExitCodeErrorVisitor) VisitNoTestExecutionRegistered() error {
	return stacktrace.NewError("The Kurtosis API container timed out waiting for registration of a test execution; " +
		"this is a bug in Kurtosis itself")
}

func (t testExecutionExitCodeErrorVisitor) VisitTestHitSetupTimeout() error {
	return stacktrace.NewError("The test failed to set up within the test setup timeout")
}

func (t testExecutionExitCodeErrorVisitor) VisitTestHitExecutionTimeout() error {
	return stacktrace.NewError("The test failed to complete within the test execution timeout")
}

func (t testExecutionExitCodeErrorVisitor) VisitTestsuiteExitedDuringSetup() error {
	return stacktrace.NewError("The testsuite exited during the setup phase, which should never happen")
}

func (t testExecutionExitCodeErrorVisitor) VisitErrWaitingForSuiteContainerExit() error {
	return stacktrace.NewError("The Kurtosis API container encountered an error waiting for the testsuite " +
		"container to exit; this is a bug in Kurtosis itself")
}
