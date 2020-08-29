/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_suite_env_vars

/*
A package to contain the contract of Docker environment variables that will be passed by the initializer to
	the testsuite to run it
 */

const (
	TestNamesFilepathEnvVar    = "TEST_NAMES_FILEPATH"
	TestEnvVar                 = "TEST"
	KurtosisApiIpEnvVar        = "KURTOSIS_API_IP"
	TestSuiteLogFilepathEnvVar = "LOG_FILEPATH"
)