/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_execution_mode

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json2"
	"github.com/kurtosis-tech/kurtosis/api_container/api_container_docker_consts"
	"github.com/kurtosis-tech/kurtosis/api_container/exit_codes"
	"github.com/kurtosis-tech/kurtosis/api_container/test_execution_mode/api"
	"github.com/kurtosis-tech/kurtosis/api_container/test_execution_mode/execution/test_execution_status"
	"github.com/kurtosis-tech/kurtosis/api_container/test_execution_mode/service_network"
	"github.com/kurtosis-tech/kurtosis/api_container/test_execution_mode/service_network/networking_sidecar"
	"github.com/kurtosis-tech/kurtosis/api_container/test_execution_mode/service_network/user_service_launcher"
	"github.com/kurtosis-tech/kurtosis/api_container/test_execution_mode/service_network/user_service_launcher/files_artifact_expander"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	// If no test suite registers a test execution in this time, the API container will shut itself down of its own accord
	idleShutdownTimeout = 15 * time.Second

	// When shutting down the service network, the maximum amount of time we'll give a container to stop gracefully
	//  before hard-killing it
	containerStopTimeout = 10 * time.Second
)

type TestExecutionArgs struct {
	executionInstanceId      string
	networkId                string
	subnetMask               string
	gatewayIpAddr            string
	testName                 string
	suiteExecutionVolumeName string
	testSuiteContainerId     string

	// It seems weird that we require this given that the test suite container doesn't run a server, but it's only so
	//  that our free IP address tracker knows not to dole out the test suite container's IP address
	testSuiteContainerIpAddr string
	apiContainerIpAddr string

	isPartitioningEnabled bool


}

type TestExecutionCodepath struct {
	args TestExecutionArgs
}

func NewTestExecutionCodepath(args TestExecutionArgs) *TestExecutionCodepath {
	return &TestExecutionCodepath{args: args}
}

func (t TestExecutionCodepath) Execute() (int, error) {
	args := t.args

	dockerManager, err := createDockerManager()
	if err != nil {
		return exit_codes.StartupErrorExitCode, stacktrace.Propagate(err, "An error occurred creating the Docker manager")
	}

	freeIpAddrTracker, err := createFreeIpAddrTracker(
		args.subnetMask,
		args.gatewayIpAddr,
		args.apiContainerIpAddr,
		args.testSuiteContainerIpAddr)
	if err != nil {
		return exit_codes.StartupErrorExitCode, stacktrace.Propagate(err, "An error occurred creating the free IP address tracker")
	}

	serviceNetwork := createServiceNetwork(
		args.executionInstanceId,
		args.testName,
		args.suiteExecutionVolumeName,
		dockerManager,
		freeIpAddrTracker,
		args.networkId,
		args.isPartitioningEnabled)

	// A value on this channel indicates a change in the test execution status
	testExecutionStatusChan := make(chan test_execution_status.TestExecutionStatus, 1)

	server, err := createServer(
		dockerManager,
		testExecutionStatusChan,
		args.testSuiteContainerId,
		serviceNetwork)
	if err != nil {
		return exit_codes.StartupErrorExitCode, stacktrace.Propagate(err, "An error occurred creating the RPC server for receiving test execution commands")
	}

	go func(){
		server.ListenAndServe()
	}()

	// Docker will send SIGTERM to end the process, and we need to catch it to stop gracefully
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	logrus.Info("Waiting for stop signal or test completion...")
	exitCode := waitUntilExitCondition(testExecutionStatusChan, signalChan)
	var exitErr error = nil

	// NOTE: Might need to kick off a timeout thread to separately close the context if it's taking too long or if
	//  the server hangs forever trying to shutdown
	logrus.Info("Shutting down JSON RPC server...")
	if err := server.Shutdown(context.Background()); err != nil {
		exitCode = exit_codes.ShutdownErrorExitCode
		exitErr = stacktrace.Propagate(err, "An error occurred shutting down the JSON RPC server")
	} else {
		logrus.Info("JSON RPC server shut down successfully")
	}

	// NOTE: Might need to kick off a timeout thread to separately close the context if it's taking too long or if
	//  the service network hangs forever trying to shutdown
	logrus.Info("Destroying service network...")
	if err := serviceNetwork.Destroy(context.Background(), containerStopTimeout); err != nil {
		exitCode = exit_codes.ShutdownErrorExitCode
		exitErr = stacktrace.Propagate(err, "An error occurred destroying the service network")
	} else {
		logrus.Info("Service network destroyed successfully")
	}

	return exitCode, exitErr
}

func createDockerManager() (*docker_manager.DockerManager, error) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not initialize a Docker client from the environment")
	}

	dockerManager, err := docker_manager.NewDockerManager(logrus.StandardLogger(), dockerClient)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the Docker manager")
	}

	return dockerManager, nil
}

func createFreeIpAddrTracker(
	networkSubnetMask string,
	gatewayIp string,
	apiContainerIp string,
	testSuiteContainerIp string) (*commons.FreeIpAddrTracker, error){
	freeIpAddrTracker, err := commons.NewFreeIpAddrTracker(
		logrus.StandardLogger(),
		networkSubnetMask,
		map[string]bool{
			gatewayIp:      true,
			apiContainerIp: true,
			testSuiteContainerIp: true,
		})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the free IP address tracker")
	}
	return freeIpAddrTracker, nil
}

func createServiceNetwork(
	executionInstanceId string,
	testName string,
	suiteExecutionVolName string,
	dockerManager *docker_manager.DockerManager,
	freeIpAddrTracker *commons.FreeIpAddrTracker,
	dockerNetworkId string,
	isPartitioningEnabled bool) *service_network.ServiceNetwork {

	filesArtifactExpander := files_artifact_expander.NewFilesArtifactExpander(
		suiteExecutionVolName,
		dockerManager,
		dockerNetworkId,
		freeIpAddrTracker)

	userServiceLauncher := user_service_launcher.NewUserServiceLauncher(
		executionInstanceId,
		testName,
		dockerManager,
		freeIpAddrTracker,
		filesArtifactExpander,
		dockerNetworkId,
		suiteExecutionVolName)

	networkingSidecarManager := networking_sidecar.NewStandardNetworkingSidecarManager(
		dockerManager,
		freeIpAddrTracker,
		dockerNetworkId)

	serviceNetwork := service_network.NewServiceNetwork(
		isPartitioningEnabled,
		freeIpAddrTracker,
		dockerManager,
		userServiceLauncher,
		networkingSidecarManager)
	return serviceNetwork
}

func createServer(
	dockerManager *docker_manager.DockerManager,
	testExecutionStatusChan chan test_execution_status.TestExecutionStatus,
	testSuiteContainerId string,
	serviceNetwork *service_network.ServiceNetwork) (*http.Server, error) {
	kurtosisService := api.NewKurtosisService(
		testSuiteContainerId,
		testExecutionStatusChan,
		dockerManager,
		serviceNetwork)

	logrus.Info("Launching server...")
	httpHandler := rpc.NewServer()
	jsonCodec := json2.NewCodec()
	httpHandler.RegisterCodec(jsonCodec, "application/json")
	if err := httpHandler.RegisterService(kurtosisService, ""); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred registering the Kurtosis service to the HTTP handler")
	}
	server := &http.Server{
		Addr:    fmt.Sprintf(":%v", api_container_docker_consts.ContainerPort),
		Handler: httpHandler,
	}

	return server, nil
}

func waitUntilExitCondition(testExecutionStatusChan chan test_execution_status.TestExecutionStatus, signalChan chan os.Signal) int {
	// If no test suite registers within our idle timeout, shut the container down
	select {
	case receivedStatus := <-testExecutionStatusChan:
		if receivedStatus != test_execution_status.Running {
			logrus.Errorf(
				"Received out-of-order test execution status update; got %v when we were expecting %v",
				receivedStatus,
				test_execution_status.Running)
			return exit_codes.OutOfOrderTestStatusExitCode
		}
	case <-time.After(idleShutdownTimeout):
		logrus.Errorf("No test suite registered itself after waiting %v; this likely means the test suite had a fatal error", idleShutdownTimeout)
		return exit_codes.NoTestSuiteRegisteredExitCode
	case signal := <-signalChan:
		logrus.Infof("Received signal %v while awaiting test registration; server will shut down", signal)
		return exit_codes.ReceivedTermSignalExitCode
	}
	logrus.Info("Received notification that a test was registered; proceeding to await test completion...")

	select {
	case receivedStatus := <-testExecutionStatusChan:
		if !(receivedStatus == test_execution_status.HitTimeout || receivedStatus == test_execution_status.CompletedBeforeTimeout) {
			logrus.Errorf(
				"Received out-of-order test execution status update; got %v when we were expecting %v or %v",
				receivedStatus,
				test_execution_status.HitTimeout,
				test_execution_status.CompletedBeforeTimeout)
			return exit_codes.OutOfOrderTestStatusExitCode
		}
		if receivedStatus == test_execution_status.HitTimeout {
			logrus.Error("Test execution hit timeout before test completion")
			return exit_codes.TestHitTimeoutExitCode
		}
	case signal := <-signalChan:
		logrus.Infof("Received signal %v while awaiting test completion; server will shut down", signal)
		return exit_codes.ReceivedTermSignalExitCode
	}

	logrus.Info("Test completed before reaching the timeout")
	return exit_codes.TestCompletedInTimeoutExitCode
}
