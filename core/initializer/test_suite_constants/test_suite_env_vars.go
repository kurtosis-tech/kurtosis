/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_suite_constants

/*
A package to contain the contract of Docker environment variables that will be passed by the initializer to
	the testsuite to run it
 */

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
	debuggerPort				  = "DEBUGGER_PORT"
)
