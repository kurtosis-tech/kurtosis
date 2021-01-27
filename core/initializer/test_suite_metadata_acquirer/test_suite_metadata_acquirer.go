/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_suite_metadata_acquirer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/api_container/api_container_docker_consts/api_container_exit_codes"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/kurtosis-tech/kurtosis/commons/suite_execution_volume"
	"github.com/kurtosis-tech/kurtosis/initializer/banner_printer"
	"github.com/kurtosis-tech/kurtosis/initializer/test_suite_launcher"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
)

const (
	metadataAcquiringApiContainerTitle     = "Metadata-Acquiring Kurtosis API Container"
	metadataSendingTestsuiteContainerTitle = "Metadata-Sending Testsuite Container"
)

func GetTestSuiteMetadata(
		suiteExecutionVolName string,
		suiteExecutionVolume *suite_execution_volume.SuiteExecutionVolume,
		dockerClient *client.Client,
		launcher *test_suite_launcher.TestsuiteContainerLauncher,
		debuggerHostPortBinding nat.PortBinding) (*TestSuiteMetadata, error) {
	parentContext := context.Background()

	dockerManager, err := docker_manager.NewDockerManager(logrus.StandardLogger(), dockerClient)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the Docker manager")
	}

	logrus.Info("Launching metadata-acquiring containers...")
	testsuiteContainerId, apiContainerId, err := launcher.LaunchMetadataAcquiringContainers(
		parentContext,
		dockerManager,
		suiteExecutionVolName,
		debuggerHostPortBinding)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred running the metadata-acquiring containers")
	}
	// Safeguard to ensure we don't ever leak containers
	defer func() {
		if err := dockerManager.KillContainer(parentContext, testsuiteContainerId); err != nil {
			logrus.Errorf("An error occurred killing the suite metadata-gathering testsuite container:")
			fmt.Fprintln(logrus.StandardLogger().Out, err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually stop container with ID '%v'!", testsuiteContainerId)
		}
	}()
	defer func() {
		if err := dockerManager.KillContainer(parentContext, apiContainerId); err != nil {
			logrus.Errorf("An error occurred killing the suite metadata-gathering Kurtosis API container:")
			fmt.Fprintln(logrus.StandardLogger().Out, err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually stop container with ID '%v'!", testsuiteContainerId)
		}
	}()
	logrus.Infof(
		"Metadata-acquiring containers launched, with testsuite debugger port bound to host port %v:%v (if a debugger " +
			"is running in the testsuite, you may need to connect to this port to allow execution to proceed)",
		debuggerHostPortBinding.HostIP,
		debuggerHostPortBinding.HostPort)

	apiContainerExitCodeInt64, err := dockerManager.WaitForExit(
		parentContext,
		apiContainerId)
	if err != nil {
		printBothContainerLogs(parentContext, dockerManager, apiContainerId, testsuiteContainerId)
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the exit of the suite metadata-serializing API container")
	}
	apiContainerExitCode := int(apiContainerExitCodeInt64)

	acceptExitCodeVisitor, found := api_container_exit_codes.ExitCodeErrorVisitorAcceptFuncs[apiContainerExitCode]
	if !found {
		return nil, stacktrace.NewError("The Kurtosis API container exited with an unrecognized " +
			"exit code '%v' that doesn't have an accept listener; this is a code bug in Kurtosis",
			apiContainerExitCode)
	}
	visitor := metadataSerializationExitCodeErrorVisitor{}
	if err := acceptExitCodeVisitor(visitor); err != nil {
		printBothContainerLogs(parentContext, dockerManager, apiContainerId, testsuiteContainerId)
		return nil, stacktrace.Propagate(
			err,
			"The API container responsible for serializing suite metadata exited with an error")
	}

	// At this point we expect the testsuite container to already have exited
	testsuiteContainerExitCode, err := dockerManager.WaitForExit(
		parentContext,
		testsuiteContainerId)
	if err != nil {
		printBothContainerLogs(parentContext, dockerManager, apiContainerId, testsuiteContainerId)
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the exit of the testsuite container that " +
			"sends suite metadata to the API container")
	}
	if testsuiteContainerExitCode != 0 {
		printBothContainerLogs(parentContext, dockerManager, apiContainerId, testsuiteContainerId)
		return nil, stacktrace.NewError("The testsuite container that sends suite metadata to the API container exited " +
			"with a nonzero exit code")
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

func printBothContainerLogs(
		ctx context.Context,
		dockerManager *docker_manager.DockerManager,
		apiContainerId string,
		testsuiteContainerId string) {
	banner_printer.PrintContainerLogsWithBanners(dockerManager, ctx, apiContainerId, logrus.StandardLogger(), metadataAcquiringApiContainerTitle)
	banner_printer.PrintContainerLogsWithBanners(dockerManager, ctx, testsuiteContainerId, logrus.StandardLogger(), metadataSendingTestsuiteContainerTitle)
}
