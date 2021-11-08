/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package test_suite_launcher

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/host_machine_directories"
	"github.com/kurtosis-tech/kurtosis-client/golang/kurtosis_core_rpc_api_consts"
	"github.com/kurtosis-tech/kurtosis-core/commons/object_labels_providers"
	"github.com/kurtosis-tech/kurtosis-core/commons/object_name_providers"
	"github.com/kurtosis-tech/kurtosis-testsuite-api-lib/golang/kurtosis_testsuite_docker_api"
	"github.com/kurtosis-tech/kurtosis-testsuite-api-lib/golang/kurtosis_testsuite_rpc_api_consts"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
	"os"
	"path"
	"strconv"
)

const (
	bridgeNetworkName = "bridge"

	// When a debugger is used inside a testsuite, these are the port & protocol that the debugger inside the container
	// will be told to listen on
	portForDebuggersRunningOnTestsuite     = 2778
	protocolForDebuggersRunningOnTestsuite = "tcp"

	// TODO This constant represents "forbidden knowledge" about the directory that the engine-server uses to store enclave
	//  data in!!! It can be deleted as soon as the engine server is the only interface for interacting with an enclave
	allEnclavesDirnameInsideEngineDataDir = "enclaves"
)

var testsuiteRpcPort = nat.Port(fmt.Sprintf("%v/%v", kurtosis_testsuite_rpc_api_consts.ListenPort, kurtosis_testsuite_rpc_api_consts.ListenProtocol))
var testsuiteDebuggerPort = nat.Port(fmt.Sprintf("%v/%v", portForDebuggersRunningOnTestsuite, protocolForDebuggersRunningOnTestsuite))

// We always publish the testsuite container's ports so that the CLI can call setup/run methods from outside the enclave
var testsuitePortPublishSpec = docker_manager.NewAutomaticPublishingSpec()

type TestsuiteContainerLauncher struct {
	testsuiteExObjNameProvider *object_name_providers.TestsuiteExecutionObjectNameProvider

	testsuiteExObjLabelsProvider *object_labels_providers.TestsuiteExecutionObjectLabelsProvider

	testsuiteImage string

	// The log level string that will be passed as-is to the testsuite (should be meaningful to the testsuite)
	suiteLogLevel string

	// The JSON-serialized custom params object that will be passed as-is to the testsuite
	customParamsJson string
}

func NewTestsuiteContainerLauncher(testsuiteExObjNameProvider *object_name_providers.TestsuiteExecutionObjectNameProvider, testsuiteExObjLabelsProvider *object_labels_providers.TestsuiteExecutionObjectLabelsProvider, testsuiteImage string, suiteLogLevel string, customParamsJson string) *TestsuiteContainerLauncher {
	return &TestsuiteContainerLauncher{testsuiteExObjNameProvider: testsuiteExObjNameProvider, testsuiteExObjLabelsProvider: testsuiteExObjLabelsProvider, testsuiteImage: testsuiteImage, suiteLogLevel: suiteLogLevel, customParamsJson: customParamsJson}
}

/*
Launches a new testsuite container for providing testsuite metadata to the initializer
*/
func (launcher TestsuiteContainerLauncher) LaunchMetadataAcquiringContainer(
		ctx context.Context,
		log *logrus.Logger,
		dockerManager *docker_manager.DockerManager) (containerId string, hostMachineRpcPortBinding *nat.PortBinding, err error) {
	functionCompletedSuccessfully := false

	networks, err := dockerManager.GetNetworksByName(ctx, bridgeNetworkName)
	if err != nil {
		return "", nil, stacktrace.Propagate(
			err,
			"An error occurred getting the network  matching the '%v' network name",
			bridgeNetworkName)
	}
	if len(networks) == 0 || len(networks) > 1 {
		return "", nil, stacktrace.NewError(
			"%v Docker network were returned for the '%v' network - this is very strange!",
			len(networks),
			bridgeNetworkName)
	}

	bridgeNetworkId := networks[0].GetId()

	testsuiteEnvVars, err := launcher.generateMetadataProvidingEnvVars()
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred generating the testsuite container env vars")
	}

	suiteContainerDesc := "metadata-providing testsuite container"
	log.Debugf("Launching %v...", suiteContainerDesc)
	containerName := launcher.testsuiteExObjNameProvider.ForMetadataAcquiringTestsuiteContainer()
	labels := launcher.testsuiteExObjLabelsProvider.ForMetadataAcquiringTestsuiteContainer()
	testsuiteContainerId, hostPortBindings, err := launcher.createAndStartTestsuiteContainerWithDebuggingPort(
		ctx,
		dockerManager,
		containerName,
		bridgeNetworkId,
		nil,   // Nil because the bridge network will assign IPs on its own (and won't know what IPs are already used)
		testsuiteEnvVars,
		map[string]string{}, // No bind mounts necessary for metadata acquisition
		labels,
	)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred launching the testsuite container to provide metadata to the Kurtosis API container")
	}
	defer killContainerIfNotFunctionSuccess(
		ctx,
		log,
		dockerManager,
		testsuiteContainerId,
		func() bool { return functionCompletedSuccessfully},
	)
	logSuccessfulSuiteContainerLaunch(log, suiteContainerDesc, hostPortBindings)

	rpcPortOnHostMachine, found := hostPortBindings[testsuiteRpcPort]
	if !found {
		return "", nil, stacktrace.NewError("No host machine port binding found for the testsuite RPC port; this should never happen since testsuite RPC ports should always be bound to the host machine")
	}

	functionCompletedSuccessfully = true
	return testsuiteContainerId, rpcPortOnHostMachine, nil
}

/*
Launches a new testsuite container to run a test
*/
func (launcher TestsuiteContainerLauncher) LaunchTestRunningContainer(
		ctx context.Context,
		log *logrus.Logger,
		dockerManager *docker_manager.DockerManager,
		enclaveId string,
		networkId string,
		containerName string,
		kurtosisApiContainerIp net.IP,
		testsuiteContainerIp net.IP,
		labels map[string]string) (string, *nat.PortBinding, error){
	log.Debugf(
		"Test suite container IP: %v; kurtosis API container IP: %v",
		testsuiteContainerIp.String(),
		kurtosisApiContainerIp.String())

	functionCompletedSuccessfully := false

	testSuiteEnvVars, err := launcher.generateTestRunningEnvVars(kurtosisApiContainerIp.String())
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred generating the test-running testsuite container env vars")
	}

	suiteContainerDesc := "test-running testsuite container"
	log.Debugf("Launching %v....", suiteContainerDesc)
	enclaveDataDirpath, err := getEnclaveDataDirpath(enclaveId)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred getting enclave data dirpath for enclave '%v'", enclaveId)
	}
	bindMounts := map[string]string{
		enclaveDataDirpath: kurtosis_testsuite_docker_api.EnclaveDataDirMountpoint,
	}
	suiteContainerId, hostPortBindings, err := launcher.createAndStartTestsuiteContainerWithDebuggingPort(
		ctx,
		dockerManager,
		containerName,
		networkId,
		testsuiteContainerIp,
		testSuiteEnvVars,
		bindMounts,
		labels,
	)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred creating the test-running testsuite container")
	}
	defer killContainerIfNotFunctionSuccess(
		ctx,
		log,
		dockerManager,
		suiteContainerId,
		func() bool { return functionCompletedSuccessfully },
	)
	logSuccessfulSuiteContainerLaunch(log, suiteContainerDesc, hostPortBindings)

	rpcPortOnHostMachine, found := hostPortBindings[testsuiteRpcPort]
	if !found {
		return "", nil, stacktrace.NewError("No host machine port binding found for the testsuite RPC port; this should never happen since testsuite RPC ports should always be bound to the host machine")
	}

	functionCompletedSuccessfully = true
	return suiteContainerId, rpcPortOnHostMachine, nil
}

// ===============================================================================================
//                                 Private helper functions
// ===============================================================================================
// NOTE: The port binding will be nil if no host port was bound
func (launcher TestsuiteContainerLauncher) createAndStartTestsuiteContainerWithDebuggingPort(
		ctx context.Context,
		dockerManager *docker_manager.DockerManager,
		name string,
		networkId string,
		containerIpAddr net.IP,
		envVars map[string]string,
		bindMounts map[string]string,
		labels map[string]string,
	) (string, map[nat.Port]*nat.PortBinding, error) {

	usedPorts := map[nat.Port]docker_manager.PortPublishSpec{
		testsuiteRpcPort: testsuitePortPublishSpec,
		// TODO only set the debugger port if we're in debug mode, which would allow the Dockerfile
		//  to conditionally start the testsuite in dlv
		testsuiteDebuggerPort: testsuitePortPublishSpec,
	}

	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		launcher.testsuiteImage,
		name,
		networkId,
	).WithStaticIP(
		containerIpAddr,
	).WithUsedPorts(
		usedPorts,
	).WithEnvironmentVariables(
		envVars,
	).WithBindMounts(
		bindMounts,
	).WithLabels(
		labels,
	).Build()
	containerId, hostPortBindings, err := dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred creating and starting the testsuite container")
	}

	return containerId, hostPortBindings, nil
}

func logSuccessfulSuiteContainerLaunch(
		log *logrus.Logger,
		suiteContainerDesc string,
		hostPortBindings map[nat.Port]*nat.PortBinding) {
	suiteLaunchSupplementalLogInfo := ""
	debuggerHostPortBinding, found := hostPortBindings[testsuiteDebuggerPort]
	if found {
		suiteLaunchSupplementalLogInfo = fmt.Sprintf(
			" with debugger port bound to host port %v:%v (if a debugger is running, you may need to connect to this port to proceed)",
			debuggerHostPortBinding.HostIP,
			debuggerHostPortBinding.HostPort,
		)
	}
	log.Debugf("Successfully created %v%v", suiteContainerDesc, suiteLaunchSupplementalLogInfo)

}

func (launcher TestsuiteContainerLauncher) generateMetadataProvidingEnvVars() (map[string]string, error) {
	result, err := launcher.generateTestSuiteEnvVars("")
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred generating the metadata-providing env vars")
	}
	return result, nil
}

func (launcher TestsuiteContainerLauncher) generateTestRunningEnvVars(kurtosisApiIp string) (map[string]string, error) {
	kurtosisApiSocket := fmt.Sprintf("%v:%v", kurtosisApiIp, kurtosis_core_rpc_api_consts.ListenPort)
	result, err := launcher.generateTestSuiteEnvVars(kurtosisApiSocket)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred generating the test-running env vars")
	}
	return result, nil
}

/*
Generates the map of environment variables needed to run a test suite container
*/
func (launcher TestsuiteContainerLauncher) generateTestSuiteEnvVars(kurtosisApiSocket string) (map[string]string, error) {
	// TODO switch to the envVars requiring a visitor to hit, so we get them all
	standardVars := map[string]string{
		kurtosis_testsuite_docker_api.CustomParamsJsonEnvVar:  launcher.customParamsJson,
		kurtosis_testsuite_docker_api.DebuggerPortEnvVar:      strconv.Itoa(portForDebuggersRunningOnTestsuite),
		kurtosis_testsuite_docker_api.KurtosisApiSocketEnvVar: kurtosisApiSocket,
		kurtosis_testsuite_docker_api.LogLevelEnvVar:          launcher.suiteLogLevel,
	}
	return standardVars, nil
}

// This function is intended to be run as a deferred function, to kill a container that was started if the
//  outer function exits with an error
func killContainerIfNotFunctionSuccess(
		ctx context.Context,
		log *logrus.Logger,
		dockerManager *docker_manager.DockerManager,
		containerId string,
		// This needs to be a function so that it gets evaluated at time of killContainer... function
		didFunctionCompleteSuccessfully func() bool) {
	if !didFunctionCompleteSuccessfully() {
		if err := dockerManager.KillContainer(ctx, containerId); err != nil {
			log.Errorf("A container was started but the function that started it exited with an error. The container needed " +
				"to be stopped to avoid leaking a running container, but the following error occurred when attempting to stop the " +
				"container:")
			fmt.Fprintln(log.Out, err)
			log.Errorf("ACTION REQUIRED: You will need to stop the testsuite container with ID '%v' manually!", containerId)
		}
	} else {
		log.Debugf("Skipping killing container '%v' because function completed successfully", containerId)
	}
}

// TODO Delete this function!!! It uses "forbidden knowledge" about how the engine server creates enclave data directories
//  to get the enclave data directory. The whole testing framework needs to go away, so that all engine/enclave data
//  directory management is done inside the engine-server
func getEnclaveDataDirpath(enclaveId string) (string, error) {
	engineDataDirpath, err := host_machine_directories.GetEngineDataDirpath()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the engine data dirpath")
	}
	// TODO This is the forbidden knoweldge part! This code here shouldn't know anything about the internal
	//  structure of the engine data dirpath - that should be for the engine alone
	enclaveDataDirpath := path.Join(engineDataDirpath, allEnclavesDirnameInsideEngineDataDir, enclaveId)
	if _, err := os.Stat(enclaveDataDirpath); os.IsNotExist(err) {
		return "", stacktrace.NewError("Expected enclave data dirpath '%v' to exist, but doesn't", enclaveDataDirpath)
	}
	return enclaveDataDirpath, nil
}
