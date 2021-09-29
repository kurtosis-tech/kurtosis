/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package test_suite_metadata_acquirer

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-testsuite-api-lib/golang/kurtosis_testsuite_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/initializer/banner_printer"
	"github.com/kurtosis-tech/kurtosis/initializer/test_suite_launcher"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"io"
	"strings"
	"time"
)

const (
	waitForTestsuiteAvailabilityTimeout = 10 * time.Second

	metadataProvidingTestsuiteContainerTitle = "Metadata-Providing Testsuite Container"

	metadataAcquisitionTimeout = 20 * time.Second

	containerStopTimeout = 10 * time.Second

	shouldFollowTestsuiteLogsOnErr = false
)

func GetTestSuiteMetadata(
		dockerClient *client.Client,
		launcher *test_suite_launcher.TestsuiteContainerLauncher) (*kurtosis_testsuite_rpc_api_bindings.TestSuiteMetadata, error) {
	parentContext := context.Background()

	dockerManager := docker_manager.NewDockerManager(logrus.StandardLogger(), dockerClient)

	logrus.Debugf("Launching metadata-providing testsuite...")
	containerId, hostMachineRpcPortBinding, err := launcher.LaunchMetadataAcquiringContainer(
		parentContext,
		logrus.StandardLogger(),
		dockerManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred launching the metadata-providing testsuite container")
	}
	// Safeguard to ensure we don't ever leak containers
	defer func() {
		if err := dockerManager.KillContainer(parentContext, containerId); err != nil {
			logrus.Errorf("An error occurred killing the suite metadata-providing testsuite container:")
			fmt.Fprintln(logrus.StandardLogger().Out, err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually stop container with ID '%v'!", containerId)
		}
	}()
	logrus.Debugf("Metadata-providing testsuite container launched")

	testsuiteSocket := fmt.Sprintf("%v:%v", hostMachineRpcPortBinding.HostIP, hostMachineRpcPortBinding.HostPort)
	conn, err := grpc.Dial(
		testsuiteSocket,
		grpc.WithInsecure(), // TODO SECURITY: Use HTTPS to verify we're connecting to the correct testsuite
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Couldn't dial testsuite container at %v to get testsuite metadata", testsuiteSocket)
	}
	testsuiteClient := kurtosis_testsuite_rpc_api_bindings.NewTestSuiteServiceClient(conn)

	logrus.Debugf("Waiting for testsuite container to become available...")
	if err := waitUntilTestsuiteContainerIsAvailable(parentContext, testsuiteClient); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while waiting for the testsuite container to become available")
	}
	logrus.Debugf("Testsuite container became available")

	logrus.Debugf("Getting testsuite metadata...")
	ctxWithTimeout, cancelFunc := context.WithTimeout(parentContext, metadataAcquisitionTimeout)
	defer cancelFunc() // Safeguard to ensure we don't leak resources
	suiteMetadata, err := testsuiteClient.GetTestSuiteMetadata(
		ctxWithTimeout,
		&emptypb.Empty{},
		grpc.WaitForReady(true),
	)
	if err != nil {
		printContainerLogsWithBanners(
			dockerManager,
			parentContext,
			containerId,
			logrus.StandardLogger(),
			metadataProvidingTestsuiteContainerTitle,
		)
		return nil, stacktrace.Propagate(err, "An error occurred getting the test suite metadata")
	}
	logrus.Debugf("Successfully retrieved testsuite metadata: %+v", suiteMetadata)

	if err := validateTestSuiteMetadata(suiteMetadata); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred validating the test suite metadata")
	}

	if err := dockerManager.StopContainer(parentContext, containerId, containerStopTimeout); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred stopping the metadata-providing testsuite container")
	}

	return suiteMetadata, nil
}

func waitUntilTestsuiteContainerIsAvailable(ctx context.Context, client kurtosis_testsuite_rpc_api_bindings.TestSuiteServiceClient) error {
	contextWithTimeout, cancelFunc := context.WithTimeout(ctx, waitForTestsuiteAvailabilityTimeout)
	defer cancelFunc()
	if _, err := client.IsAvailable(contextWithTimeout, &emptypb.Empty{}, grpc.WaitForReady(true)); err != nil {
		return stacktrace.Propagate(err, "An error occurred waiting for the testsuite container to become available")
	}
	return nil
}

/*
Little helper function to print a container's logs with with banners indicating the start and end of the logs

Args:
	dockerManager: Docker manager to use when retrieving container logs
	context: The context in which to run the log retrieval Docker function
	containerId: ID of the Docker container from which to retrieve logs
	containerDescription: Short, human-readable description of the container whose logs are being printed
	logFilepath: Filepath of the file containing the container's logs
*/
func printContainerLogsWithBanners(
		dockerManager *docker_manager.DockerManager,
		context context.Context,
		containerId string,
		log *logrus.Logger,
		containerDescription string) {
	var logReader io.Reader
	var useDockerLogDemultiplexing bool
	logReadCloser, err := dockerManager.GetContainerLogs(context, containerId, shouldFollowTestsuiteLogsOnErr)
	if err != nil {
		errStr := fmt.Sprintf("Could not print container's logs due to the following error: %v", err)
		logReader = strings.NewReader(errStr)
		useDockerLogDemultiplexing = false
	} else {
		defer logReadCloser.Close()
		logReader = logReadCloser
		useDockerLogDemultiplexing = true
	}

	banner_printer.PrintSection(log, containerDescription + " Logs", false)
	var copyErr error
	if useDockerLogDemultiplexing {
		// Docker logs multiplex STDOUT and STDERR into a single stream, and need to be demultiplexed
		// See https://github.com/moby/moby/issues/32794
		_, copyErr = stdcopy.StdCopy(log.Out, log.Out, logReader)
	} else {
		_, copyErr = io.Copy(log.Out, logReader)
	}
	if copyErr != nil {
		log.Errorf("Could not print the test suite container's logs due to the following error when copying log contents:")
		fmt.Fprintln(log.Out, err)
	}
	banner_printer.PrintSection(log, "End " + containerDescription + " Logs", false)
}

func validateTestSuiteMetadata(suiteMetadata *kurtosis_testsuite_rpc_api_bindings.TestSuiteMetadata) error {
	if suiteMetadata.TestMetadata == nil {
		return stacktrace.NewError("Test metadata map is nil")
	}
	// NOTE: We don't check if the static file set is nil because an empty map will be deserialized as nil
	if len(suiteMetadata.TestMetadata) == 0 {
		return stacktrace.NewError("Test suite doesn't declare any tests")
	}
	for testName, testMetadata := range suiteMetadata.TestMetadata {
		if len(strings.TrimSpace(testName)) == 0 {
			return stacktrace.NewError("Test name '%v' is empty", testName)
		}
		if err := validateTestMetadata(testMetadata); err != nil {
			return stacktrace.Propagate(err, "An error occurred validating metadata for test '%v'", testName)
		}
	}
	return nil
}

func validateTestMetadata(testMetadata *kurtosis_testsuite_rpc_api_bindings.TestMetadata) error {
	if testMetadata.TestSetupTimeoutInSeconds <= 0 {
		return stacktrace.NewError("Test setup timeout is %v, but must be greater than 0.", testMetadata.TestSetupTimeoutInSeconds)
	}
	if testMetadata.TestRunTimeoutInSeconds <= 0 {
		return stacktrace.NewError("Test run timeout is %v, but must be greater than 0.", testMetadata.TestRunTimeoutInSeconds)
	}
	return nil
}
