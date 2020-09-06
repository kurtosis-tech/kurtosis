/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_suite_constants

import "github.com/palantir/stacktrace"

/*
A package to contain the contract of Docker environment variables that will be passed by the initializer to
	the testsuite to run it
 */

const (
	MetadataFilepathEnvVar = "METADATA_FILEPATH"
	TestEnvVar             = "TEST"
	KurtosisApiIpEnvVar    = "KURTOSIS_API_IP"
	ServicesDirpathEnvVar  = "SERVICES_DIRPATH"
	LogFilepathEnvVar      = "LOG_FILEPATH"
	LogLevelEnvVar         = "LOG_LEVEL"
)

/*
Generates the map of environment variables needed to run a test suite container

NOTE: exactly one of metadata_filepath or test_name must be non-empty!
 */
func GenerateTestSuiteEnvVars(
		metadataFilepathEnvVar string,
		testName string,
		kurtosisApiIp string,
		servicesDirpath string,
		logFilepath string,
		logLevel string,
		customEnvVars map[string]string) (map[string]string, error) {
	standardVars := map[string]string{
		MetadataFilepathEnvVar: metadataFilepathEnvVar,
		TestEnvVar:             testName,
		KurtosisApiIpEnvVar:    kurtosisApiIp,
		ServicesDirpathEnvVar:  servicesDirpath,
		LogFilepathEnvVar:      logFilepath,
		LogLevelEnvVar:         logLevel,
	}
	for key, val := range customEnvVars {
		if _, ok := standardVars[key]; ok {
			return nil, stacktrace.NewError(
				"Custom test suite environment variable binding %s=%s requested, but is not allowed because key is " +
					"already being used by Kurtosis.",
				key,
				val)
		}
		standardVars[key] = val
	}
	return standardVars, nil
}
