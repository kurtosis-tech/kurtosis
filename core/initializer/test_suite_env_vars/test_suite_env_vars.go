/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_suite_env_vars

import "github.com/palantir/stacktrace"

/*
A package to contain the contract of Docker environment variables that will be passed by the initializer to
	the testsuite to run it
 */

const (
	MetadataFilepathEnvVar = "METADATA_FILEPATH"
	TestEnvVar             = "TEST"
	KurtosisApiIpEnvVar    = "KURTOSIS_API_IP"
	LogFilepathEnvVar      = "LOG_FILEPATH"
	LogLevelEnvVar = "LOG_LEVEL"
)

/*
Generates the map of environment variables needed to run a test suite container

NOTE: exactly one of metadata_filepath or test_name must be non-empty!
 */
func GenerateTestSuiteEnvVars(
		metadataFilepathEnvVar string,
		testName string,
		kurtosisApiIp string,
		logFilepath string,
		logLevel string,
		customEnvVars map[string]string) (map[string]string, error) {
	standardVars := map[string]string{
		MetadataFilepathEnvVar: metadataFilepathEnvVar,
		TestEnvVar:             testName,
		KurtosisApiIpEnvVar:    kurtosisApiIp,
		LogFilepathEnvVar:      logFilepath,
		LogLevelEnvVar:         logLevel,
	}
	for key, val := range customEnvVars {
		if _, ok := standardVars[key]; ok {
			return nil, stacktrace.NewError(
				"Tried to manually add custom environment variable %s to the test controller container, but it is " +
					"already being used by Kurtosis.",
				key)
		}
		standardVars[key] = val
	}
	return standardVars, nil
}
