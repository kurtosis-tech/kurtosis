/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_suite_metadata_acquirer

import (
	"context"
	"encoding/json"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/kurtosis-tech/kurtosis/initializer/banner_printer"
	"github.com/kurtosis-tech/kurtosis/initializer/test_suite_constants"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
)

const (
	bridgeNetworkName = "bridge"

	// The name of the dirctory that will be created inside the suite execution volume for storing files related
	//  to acquiring test suite metadata
	metadataAcquirerDirname = "metadata-acquirer"

	testListingContainerDescription = "Test-Listing Container"
	testSuiteMetadataFilename = "test-suite-metadata.json"
	logFilename = "test-listing.log"
)

/*
Spins up a testsuite container in test-listing mode and returns the "set" of tests that it spits out

Args:
	testSuiteImage: The name of the Docker image containing the test suite
	suiteExecutionVolume: The name of the Docker volume dedicated for storing file IO for the suite execution
	suiteExecutionVolumeMountDirpath: The dirpath where the suite execution volume is mounted on the initializer container
	dockerClient: The Docker client with which Docker requests will be made
	testSuiteLogLevel: The log level the test suite will output with
	customEnvVars: A key-value mapping of custom environment variables that will be set when running the test suite image
*/
func GetTestSuiteMetadata(
		testSuiteImage string,
		suiteExecutionVolume string,
		suiteExecutionVolumeMountDirpath string,
		dockerClient *client.Client,
		testSuiteLogLevel string,
		customEnvVars map[string]string) (*TestSuiteMetadata, error) {
	parentContext := context.Background()

	dockerManager, err := commons.NewDockerManager(logrus.StandardLogger(), dockerClient)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the Docker manager")
	}

	bridgeNetworkIds, err := dockerManager.GetNetworkIdsByName(parentContext, bridgeNetworkName)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting the network IDs matching the '%v' network",
			bridgeNetworkName)
	}
	if len(bridgeNetworkIds) == 0 || len(bridgeNetworkIds) > 1 {
		return nil, stacktrace.NewError(
			"%v Docker network IDs were returned for the '%v' network - this is very strange!",
			len(bridgeNetworkIds),
			bridgeNetworkName)
	}
	bridgeNetworkId := bridgeNetworkIds[0]

	metadataAcquirerDirpathOnInitializer := path.Join(suiteExecutionVolumeMountDirpath, metadataAcquirerDirname)
	if err := os.Mkdir(metadataAcquirerDirpathOnInitializer, os.ModeDir); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a directory in the suite execution volume to " +
			"store data for the acquisition of test suite metadata")
	}
	metadataAcquirerDirpathOnSuite := path.Join(test_suite_constants.SuiteExecutionVolumeMountpoint, metadataAcquirerDirname)

	metadataFilepathOnSuite := path.Join(metadataAcquirerDirpathOnSuite, testSuiteMetadataFilename)
	logFilepathOnSuite := path.Join(metadataAcquirerDirpathOnSuite, logFilename)

	envVars, err := test_suite_constants.GenerateTestSuiteEnvVars(
		metadataFilepathOnSuite,
		"", // We leave the test name blank to signify that we want test listing, not test execution
		"", // Because we're doing test listing, not test execution, the Kurtosis API IP can be blank
		"", // We leave the services dirpath blank because getting suite metadata doesn't require knowing this
		logFilepathOnSuite,
		testSuiteLogLevel,
		customEnvVars)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred generating the Docker environment variables for the test suite")
	}

	testListingContainerId, err := dockerManager.CreateAndStartContainer(
		parentContext,
		testSuiteImage,
		bridgeNetworkId,
		nil,  // Nil because the bridge network will assign IPs on its own; we don't need to (and won't know what IPs are already used)
		map[nat.Port]bool{},
		nil, // Nil start command args because we expect the test suite image to be parameterized with variables
		envVars,
		map[string]string{},
		map[string]string{
			suiteExecutionVolume: test_suite_constants.SuiteExecutionVolumeMountpoint,
		})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the test suite container to list the tests")
	}

	logFilepathOnInitializer := path.Join(metadataAcquirerDirpathOnInitializer, logFilename)
	testListingExitCode, err := dockerManager.WaitForExit(
		parentContext,
		testListingContainerId)
	if err != nil {
		banner_printer.PrintContainerLogsWithBanners(logrus.StandardLogger(), testListingContainerDescription, logFilepathOnInitializer)
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the exit of the testsuite container to list the tests")
	}
	if testListingExitCode != 0 {
		banner_printer.PrintContainerLogsWithBanners(logrus.StandardLogger(), testListingContainerDescription, logFilepathOnInitializer)
		return nil, stacktrace.NewError("The testsuite container for listing tests exited with a nonzero exit code")
	}

	metadataFilepathOnInitializer := path.Join(metadataAcquirerDirpathOnInitializer, testSuiteMetadataFilename)
	testSuiteMetadataReaderFp, err := os.Open(metadataFilepathOnInitializer)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred opening the temp filename containing test suite metadata for reading")
	}
	defer testSuiteMetadataReaderFp.Close()

	jsonBytes, err := ioutil.ReadAll(testSuiteMetadataReaderFp)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred reading the test suite metadata JSON string from file")
	}

	var suiteMetadata TestSuiteMetadata
	json.Unmarshal(jsonBytes, &suiteMetadata)

	return &suiteMetadata, nil
}
