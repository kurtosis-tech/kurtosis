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
	"github.com/kurtosis-tech/kurtosis/commons/suite_execution_volume"
	"github.com/kurtosis-tech/kurtosis/initializer/banner_printer"
	"github.com/kurtosis-tech/kurtosis/initializer/test_suite_constants"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
)

const (
	bridgeNetworkName = "bridge"

	metadataAcquiringApiContainerTitle     = "Suite Metadata-Acquiring API Container"
	metadataSendingTestsuiteContainerTitle = "Suite Metadata-Sending Testsuite Container"
)

func GetTestSuiteMetadata(
		suiteExecutionVolName string,
		suiteExecutionVolume *suite_execution_volume.SuiteExecutionVolume,
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


	logrus.Info("Launching metadata-acquiring testsuite container...")
	testsuiteContainerId, apiContainerId, err := launcher.LaunchMetadataAcquiringContainer(
		parentContext,
		dockerManager,
		bridgeNetworkId,
		suiteExecutionVolName,
		debuggerHostPortBinding)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred running the metadta-acquiring testsuite & API container")
	}
	logrus.Infof(
		"Metadata-acquiring testsuite container launched, with debugger port bound to host port %v:%v (if a debugger " +
			"is running in the testsuite, you may need to connect to this port to allow execution to proceed)",
		debuggerHostPortBinding.HostIP,
		debuggerHostPortBinding.HostPort)

	apiContainerExitCode, err := dockerManager.WaitForExit(
		parentContext,
		apiContainerId)
	if err != nil {
		banner_printer.PrintContainerLogsWithBanners(*dockerManager, parentContext, apiContainerId, logrus.StandardLogger(), metadataAcquiringApiContainerTitle)
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the exit of the suite metadata-serializing API container")
	}
	if apiContainerExitCode != 0 {
		banner_printer.PrintContainerLogsWithBanners(*dockerManager, parentContext, apiContainerId, logrus.StandardLogger(), metadataAcquiringApiContainerTitle)
		return nil, stacktrace.NewError("The API container for serializing suite metadata exited with a nonzero exit code")
	}

	// At this point we expect the testsuite container to already have exited
	testsuiteContainerExitCode, err := dockerManager.WaitForExit(
		parentContext,
		testsuiteContainerId)
	if err != nil {
		banner_printer.PrintContainerLogsWithBanners(*dockerManager, parentContext, testsuiteContainerId, logrus.StandardLogger(), metadataSendingTestsuiteContainerTitle)
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the exit of the testsuite container that sends suite metadata to the API container")
	}
	if testsuiteContainerExitCode != 0 {
		banner_printer.PrintContainerLogsWithBanners(*dockerManager, parentContext, apiContainerId, logrus.StandardLogger(), metadataSendingTestsuiteContainerTitle)
		return nil, stacktrace.NewError("The testsuite container that sends suite metadata to the API container exited with a nonzero exit code")
	}

	suiteMetadataFile := suiteExecutionVolume.GetSuiteMetadataFile()
	suiteMetadataFilepath := suiteMetadataFile.GetAbsoluteFilepath()

	suiteMetadataReaderFp, err := os.Open(suiteMetadataFilepath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred opening the testsuite metadata file at '%v' for reading", suiteMetadataFilepath)
	}
	defer suiteMetadataReaderFp.Close()

	jsonBytes, err := ioutil.ReadAll(suiteMetadataReaderFp)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred reading the testsuite metadata JSON string from file")
	}
	logrus.Debugf("Test suite metadata JSON: " + string(jsonBytes))

	var suiteMetadata TestSuiteMetadata
	if err := json.Unmarshal(jsonBytes, &suiteMetadata); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred deserializing the testsuite metadata JSON")
	}

	if err := validateTestSuiteMetadata(suiteMetadata); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred validating the test suite metadata")
	}

	return &suiteMetadata, nil
}
