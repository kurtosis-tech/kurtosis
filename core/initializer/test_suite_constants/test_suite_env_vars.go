/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_suite_constants

/*
A package to contain the contract of Docker environment variables that will be passed by the initializer to
	the testsuite to run it
 */

// TODO Refactor these so that they're passed in as a single JSON str
const (
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! IMPORTANT !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	// When you change these, make sure to update the docs at:
	// 	https://github.com/kurtosis-tech/kurtosis-docs/blob/develop/testsuite-details.md#dockerfile
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! IMPORTANT !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	metadataFilepathEnvVar        = "METADATA_FILEPATH"
	testEnvVar                    = "TEST"
	kurtosisApiIpEnvVar           = "KURTOSIS_API_IP"
	servicesRelativeDirpathEnvVar = "SERVICES_RELATIVE_DIRPATH"
	logLevelEnvVar                = "LOG_LEVEL"
	debuggerPortEnvVar            = "DEBUGGER_PORT"
	modeEnvVar  				  = "MODE"

	// TODO break into separate file??
	printSuiteMetadataMode = "PRINT_SUITE_METADATA"
	executeTestMode = "EXECUTE_TEST_MODE"
)

