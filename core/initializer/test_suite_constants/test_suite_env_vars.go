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
	// TODO get rid of these docs - we don't want the user to have to fuss about this
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! IMPORTANT !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	// When you change these, make sure to update the docs at:
	// 	https://github.com/kurtosis-tech/kurtosis-docs/blob/develop/testsuite-details.md#dockerfile
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! IMPORTANT !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	debuggerPortEnvVar            = "DEBUGGER_PORT"
	kurtosisApiSocketEnvVar		  = "KURTOSIS_API_SOCKET"
	logLevelEnvVar                = "LOG_LEVEL"
	customParamsJson			  = "CUSTOM_PARAMS_JSON"
)

