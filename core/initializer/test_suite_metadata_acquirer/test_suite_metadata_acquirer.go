/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_suite_metadata_acquirer

import (
	"context"
	"encoding/json"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/kurtosis-tech/kurtosis/initializer/banner_printer"
	"github.com/kurtosis-tech/kurtosis/initializer/test_suite_env_vars"
	"github.com/kurtosis-tech/kurtosis/initializer/test_suite_mount_locations"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
)

const (
	bridgeNetworkName = "bridge"
	testListingContainerDescription = "Test-Listing Container"
	testSuiteMetadataFilename = "test-suite-metadata.json"
	logFilename = "test-listing.log"
)

/*
Spins up a testsuite container in test-listing mode and returns the "set" of tests that it spits out
*/
func GetTestSuiteMetadata(
		testSuiteImage string,
		dockerManager *commons.DockerManager) (*TestSuiteMetadata, error) {
	parentContext := context.Background()

	// Create the tempfile that the testsuite image will write test names to
	testSuiteMetadataFp, err := ioutil.TempFile("", testSuiteMetadataFilename)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the temp filepath to write test suite " +
			"metadata to")
	}
	testSuiteMetadataFp.Close()

	containerLogFp, err := ioutil.TempFile("", logFilename)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the temp filepath to store the " +
			"test suite container logs")
	}
	containerLogFp.Close()
	defer os.Remove(containerLogFp.Name())

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

	testSuiteMetadataFilepath := path.Join(test_suite_mount_locations.BindMountsDirpath, testSuiteMetadataFilename)
	logFilepath := path.Join(test_suite_mount_locations.BindMountsDirpath, logFilename)
	testListingContainerId, err := dockerManager.CreateAndStartContainer(
		parentContext,
		testSuiteImage,
		// TODO parameterize these
		bridgeNetworkId,
		nil,  // Nil because the bridge network will assign IPs on its own; we don't need to (and won't know what IPs are already used)
		map[nat.Port]bool{},
		nil,
		map[string]string{
			test_suite_env_vars.MetadataFilepathEnvVar:     testSuiteMetadataFilepath,
			test_suite_env_vars.TestEnvVar:                 "", // We leave this blank to signify that we want test listing, not test execution
			test_suite_env_vars.KurtosisApiIpEnvVar:        "", // Because we're doing test listing, this can be blank
			test_suite_env_vars.TestSuiteLogFilepathEnvVar: logFilepath,
		},
		map[string]string{
			testSuiteMetadataFp.Name(): testSuiteMetadataFilepath,
			containerLogFp.Name():      logFilepath,
		},
		map[string]string{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the test suite container to list the tests")
	}

	testListingExitCode, err := dockerManager.WaitForExit(
		parentContext,
		testListingContainerId)
	if err != nil {
		banner_printer.PrintContainerLogsWithBanners(logrus.StandardLogger(), testListingContainerDescription, containerLogFp.Name())
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the exit of the testsuite container to list the tests")
	}
	if testListingExitCode != 0 {
		banner_printer.PrintContainerLogsWithBanners(logrus.StandardLogger(), testListingContainerDescription, containerLogFp.Name())
		return nil, stacktrace.NewError("The testsuite container for listing tests exited with a nonzero exit code")
	}

	testSuiteMetadataReaderFp, err := os.Open(testSuiteMetadataFp.Name())
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
