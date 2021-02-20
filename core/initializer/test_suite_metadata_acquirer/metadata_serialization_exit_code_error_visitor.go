/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_suite_metadata_acquirer

import "github.com/palantir/stacktrace"

type metadataSerializationExitCodeErrorVisitor struct {}

func (m metadataSerializationExitCodeErrorVisitor) VisitSuccessfulExit() error {
	return nil
}

func (t metadataSerializationExitCodeErrorVisitor) VisitNoTestSuiteRegistered() error {
	return stacktrace.NewError("The Kurtosis API container timed out waiting for a testsuite container to register " +
		"itself; this is a bug in Kurtosis itself")
}

func (t metadataSerializationExitCodeErrorVisitor) VisitShutdownEventBeforeSuiteRegistration() error {
	return stacktrace.NewError("The Kurtosis API's main service told itself to shut down before a testsuite container was " +
		"registered; this is a bug in Kurtosis itself")
}

func (t metadataSerializationExitCodeErrorVisitor) VisitStartupError() error {
	return stacktrace.NewError("The Kurtosis API container encountered an error while " +
		"starting up; this is a bug in Kurtosis itself")
}

func (t metadataSerializationExitCodeErrorVisitor) VisitShutdownError() error {
	return stacktrace.NewError("The Kurtosis API container encountered an error while " +
		"shutting down; this is a bug in Kurtosis itself")
}

func (m metadataSerializationExitCodeErrorVisitor) VisitReceivedTermSignal() error {
	return stacktrace.NewError("The Kurtosis API container exited due to receiving " +
		"a shutdown signal; if this is not expected, it indicates a bug in Kurtosis itself")
}

func (m metadataSerializationExitCodeErrorVisitor) VisitSerializeNotCalled() error {
	return stacktrace.NewError("The testsuite container registered itself with the API container, but " +
		"didn't call serialize; this is a bug in Kurtosis itself")
}

func (m metadataSerializationExitCodeErrorVisitor) VisitNoTestSetupRegistered() error {
	return getWrongModeError("no test setup registered")
}

func (m metadataSerializationExitCodeErrorVisitor) VisitNoTestExecutionRegistered() error {
	return getWrongModeError("no test execution registered")
}

func (m metadataSerializationExitCodeErrorVisitor) VisitTestHitSetupTimeout() error {
	return getWrongModeError("test hit setup timeout")
}

func (m metadataSerializationExitCodeErrorVisitor) VisitTestHitExecutionTimeout() error {
	return getWrongModeError("test hit execution timeout")
}

func (m metadataSerializationExitCodeErrorVisitor) VisitTestsuiteExitedDuringSetup() error {
	return getWrongModeError("testsuite exited during setup")
}

func (m metadataSerializationExitCodeErrorVisitor) VisitErrWaitingForSuiteContainerExit() error {
	return getWrongModeError("error waiting for testsuite container to exit")
}

func getWrongModeError(errorCodeDesc string) error {
	return stacktrace.NewError(
		"The Kurtosis API container exited with a '%v' error, which means it was in test execution " +
			"mode rather than metadata serialization mode; this is a bug with Kurtosis itself",
		errorCodeDesc)
}

