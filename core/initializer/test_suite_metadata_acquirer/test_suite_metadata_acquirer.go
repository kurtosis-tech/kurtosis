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
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
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

	// The name of the directory that will be created inside the suite execution volume for storing files related
	//  to acquiring test suite metadata
	metadataAcquirerDirname = "metadata-acquirer"

	metadataAcquiringContainerDescription = "Testsuite Metadata-Acquiring Container"
	testSuiteMetadataFilename             = "test-suite-metadata.json"
)

/*
Spins up a testsuite container in metadata-acquiring mode and returns the metadata that the suite returns

Args:
	testSuiteImage: The name of the Docker image containing the test suite
	suiteExecutionVolume: The name of the Docker volume dedicated for storing file IO for the suite execution
	initializerContainerSuiteExVolDirpath: The dirpath on the INITIALIZER container where the suite execution volume is mounted
	dockerClient: The Docker client with which Docker requests will be made
	launcher: Launcher for creating test suite containers
	debuggerHostPortBinding: The host port binding to use for binding to the testsuite debugger port
*/
func GetTestSuiteMetadata(
		suiteExecutionVolume string,
		initializerContainerSuiteExVolDirpath string,
		dockerClient *client.Client,
		launcher *test_suite_constants.TestsuiteContainerLauncher,
		debuggerHostPortBinding nat.PortBinding) (*TestSuiteMetadata, error) {
	parentContext := context.Background()

	dockerManager, err := docker_manager.NewDockerManager(logrus.StandardLogger(), dockerClient)
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

	metadataAcquirerDirpathOnInitializer := path.Join(initializerContainerSuiteExVolDirpath, metadataAcquirerDirname)
	if err := os.Mkdir(metadataAcquirerDirpathOnInitializer, os.ModeDir); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a directory in the suite execution volume to " +
			"store data for the acquisition of test suite metadata")
	}
	metadataAcquirerDirpathOnSuite := path.Join(test_suite_constants.TestsuiteContainerSuiteExVolMountpoint, metadataAcquirerDirname)

	metadataFilepathOnSuite := path.Join(metadataAcquirerDirpathOnSuite, testSuiteMetadataFilename)

	logrus.Info("Launching metadata-acquiring testsuite container...")
	containerId, err := launcher.LaunchMetadataAcquiringContainer(
		parentContext,
		dockerManager,
		bridgeNetworkId,
		suiteExecutionVolume,
		metadataFilepathOnSuite,
		debuggerHostPortBinding)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred launching the metadata-acquiring testsuite container")
	}
	logrus.Infof(
		"Metadata-acquiring testsuite container launched, with debugger port bound to host port %v:%v (if a debugger " +
			"is running in the testsuite, you may need to connect to this port to allow execution to proceed)",
		debuggerHostPortBinding.HostIP,
		debuggerHostPortBinding.HostPort)

	exitCode, err := dockerManager.WaitForExit(
		parentContext,
		containerId)
	if err != nil {
		banner_printer.PrintContainerLogsWithBanners(*dockerManager, parentContext, containerId, logrus.StandardLogger(), metadataAcquiringContainerDescription)
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the exit of the testsuite container to return test metadata")
	}
	if exitCode != 0 {
		banner_printer.PrintContainerLogsWithBanners(*dockerManager, parentContext, containerId, logrus.StandardLogger(), metadataAcquiringContainerDescription)
		return nil, stacktrace.NewError("The testsuite container for acquiring suite metadata exited with a nonzero exit code")
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

	logrus.Debugf("Test suite metadata JSON: " + string(jsonBytes))

	var suiteMetadata TestSuiteMetadata
	if err := json.Unmarshal(jsonBytes, &suiteMetadata); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred deserializing the testsuite metadata JSON")
	}

	if err := validateTestSuiteMetadata(&suiteMetadata); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred validating the test suite metadata")
	}

	return &suiteMetadata, nil
}
