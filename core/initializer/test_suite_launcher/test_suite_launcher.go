/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_suite_launcher

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"github.com/kurtosis-tech/kurtosis/api_container/api_container_docker_consts/api_container_env_vars"
	"github.com/kurtosis-tech/kurtosis/api_container/api_container_docker_consts/api_container_mountpoints"
	"github.com/kurtosis-tech/kurtosis/api_container/api_container_env_var_values/api_container_modes"
	"github.com/kurtosis-tech/kurtosis/api_container/api_container_env_var_values/api_container_params_json"
	"github.com/kurtosis-tech/kurtosis/api_container/server/api_container_server_consts"
	"github.com/kurtosis-tech/kurtosis/commons/docker_constants"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/kurtosis-tech/kurtosis/commons/free_host_port_binding_supplier"
	"github.com/kurtosis-tech/kurtosis/initializer/test_suite_docker_consts/test_suite_container_mountpoints"
	"github.com/kurtosis-tech/kurtosis/initializer/test_suite_docker_consts/test_suite_env_vars"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
	"strconv"
)

const (
	bridgeNetworkName = "bridge"

	dockerSocket = "/var/run/docker.sock"

	// When a debugger is used inside a testsuite, these are the port & protocol that the debugger inside the container
	// will be told to listen on
	portForDebuggersRunningOnTestsuite     = 2778
	protocolForDebuggersRunningOnTestsuite = "tcp"

	// Can make these configurable if needed
	hostPortTrackerInterfaceIp = "127.0.0.1"
	hostPortTrackerStartRange = 8000
	hostPortTrackerEndRange = 10000
)

type TestsuiteContainerLauncher struct {
	executionInstanceId uuid.UUID

	suiteExecutionVolName string

	kurtosisApiImage string

	kurtosisApiLogLevel logrus.Level

	testsuiteImage string

	// The log level string that will be passed as-is to the testsuite (should be meaningful to the testsuite)
	suiteLogLevel string

	// The JSON-serialized custom params object that will be passed as-is to the testsuite
	customParamsJson string

	// This is the port object inside the testsuite container that a debugger might be listening on, if any is running at all
	// We'll always bind this port on the testsuite container to a port on the user's machine, so they can attach
	//  a debugger if desired
	debuggerPortObj nat.Port

	// Supplier for getting free host ports, which will only be non-nil if doDebuggerHostPortBinding is set to true in
	//  the constructor
	hostPortBindingSupplier *free_host_port_binding_supplier.FreeHostPortBindingSupplier
}

func NewTestsuiteContainerLauncher(
		executionInstanceId uuid.UUID,
		suiteExecutionVolName string,
		kurtosisApiImage string,
		kurtosisApiLogLevel logrus.Level,
		testsuiteImage string,
		testsuiteLogLevel string,
		customParamsJson string,
		doDebuggerHostPortBinding bool) (*TestsuiteContainerLauncher, error) {
	var hostPortBindingSupplier *free_host_port_binding_supplier.FreeHostPortBindingSupplier = nil
	if doDebuggerHostPortBinding {
		supplier, err := free_host_port_binding_supplier.NewFreeHostPortBindingSupplier(
			docker_constants.HostMachineDomainInsideContainer,
			hostPortTrackerInterfaceIp,
			protocolForDebuggersRunningOnTestsuite,
			hostPortTrackerStartRange,
			hostPortTrackerEndRange,
			map[uint32]bool{}, // We don't know of any taken ports at this point
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred instantiating the free host port binding supplier")
		}
		hostPortBindingSupplier = supplier
	}
	return &TestsuiteContainerLauncher{
		executionInstanceId:       executionInstanceId,
		suiteExecutionVolName:     suiteExecutionVolName,
		kurtosisApiImage:          kurtosisApiImage,
		kurtosisApiLogLevel:       kurtosisApiLogLevel,
		testsuiteImage:            testsuiteImage,
		suiteLogLevel:             testsuiteLogLevel,
		customParamsJson:          customParamsJson,
		hostPortBindingSupplier:   hostPortBindingSupplier,
	}, nil
}

/*
Launches a new testsuite container to acquire testsuite metadata
 */
func (launcher TestsuiteContainerLauncher) LaunchMetadataAcquiringContainers(
		ctx context.Context,
		log *logrus.Logger,
		dockerManager *docker_manager.DockerManager) (testsuiteContainerId string, kurtosisApiContainerId string, err error) {
	functionCompletedSuccessfully := false

	bridgeNetworkIds, err := dockerManager.GetNetworkIdsByName(ctx, bridgeNetworkName)
	if err != nil {
		return "", "", stacktrace.Propagate(
			err,
			"An error occurred getting the network IDs matching the '%v' network",
			bridgeNetworkName)
	}
	if len(bridgeNetworkIds) == 0 || len(bridgeNetworkIds) > 1 {
		return "", "", stacktrace.NewError(
			"%v Docker network IDs were returned for the '%v' network - this is very strange!",
			len(bridgeNetworkIds),
			bridgeNetworkName)
	}
	bridgeNetworkId := bridgeNetworkIds[0]

	apiContainerEnvVars, err := launcher.genSuiteMetadataSerializationApiContainerEnvVars()
	if err != nil {
		return "", "", stacktrace.Propagate(err, "An error occurred generating the API container env vars")
	}

	log.Debug("Launching Kurtosis API container...")
	kurtosisApiPort := nat.Port(fmt.Sprintf("%v/%v", api_container_server_consts.ListenPort, api_container_server_consts.ListenProtocol))
	kurtosisApiContainerId, err = dockerManager.CreateAndStartContainer(
		ctx,
		launcher.kurtosisApiImage,
		bridgeNetworkId,
		nil,	// We're connecting to the bridge network, which will assign an IP automatically
		map[docker_manager.ContainerCapability]bool{}, // No extra capabilities needed for the API container
		docker_manager.DefaultNetworkMode,
		map[nat.Port]*nat.PortBinding{
			kurtosisApiPort: nil,
		},
		nil, // No ENTRYPOINT overriding needed because the API container is launched via env vars
		nil, // No CMD overriding needed for the same reason
		apiContainerEnvVars,
		map[string]string{},   // We don't need to bind mount the Docker socket because this API container won't interact with Docker
		map[string]string{
			launcher.suiteExecutionVolName: api_container_mountpoints.SuiteExecutionVolumeMountDirpath,
		},
		false,	// During metadata-acquisition, the API container doesn't need access to the Docker host machine
	)
	if err != nil {
		return "", "", stacktrace.Propagate(err, "An error occurred launching the Kurtosis API container")
	}
	defer killContainerIfNotFunctionSuccess(
		ctx,
		log,
		dockerManager,
		kurtosisApiContainerId,
		func() bool { return functionCompletedSuccessfully },
	)
	log.Debug("Successfully launched the Kurtosis API container")

	apiContainerIp, err := dockerManager.GetContainerIP(ctx, bridgeNetworkName, kurtosisApiContainerId)
	if err != nil {
		return "", "", stacktrace.Propagate(err, "An error occurred getting the API container's IP on network '%v'", bridgeNetworkName)
	}

	testsuiteEnvVars, err := launcher.generateTestSuiteEnvVars(apiContainerIp)
	if err != nil {
		return "", "", stacktrace.Propagate(err, "An error occurred generating the testsuite container env vars")
	}

	suiteContainerDesc := "metadata-reporting testsuite container"
	log.Infof("Launching %v...", suiteContainerDesc)
	testsuiteContainerId, debuggerPortHostBinding, err := launcher.createAndStartTestsuiteContainerWithDebuggingPortIfNecessary(
		ctx,
		dockerManager,
		bridgeNetworkId,
		nil,   // Nil because the bridge network will assign IPs on its own (and won't know what IPs are already used)
		testsuiteEnvVars,
	)
	if err != nil {
		return "", "", stacktrace.Propagate(err, "An error occurred launching the testsuite container to send metadata to the Kurtosis API container")
	}
	defer killContainerIfNotFunctionSuccess(
		ctx,
		log,
		dockerManager,
		testsuiteContainerId,
		func() bool { return functionCompletedSuccessfully},
	)
	logSuccessfulSuiteContainerLaunch(log, suiteContainerDesc, debuggerPortHostBinding)

	functionCompletedSuccessfully = true
	return testsuiteContainerId, kurtosisApiContainerId, nil
}

/*
Launches a new testsuite container to run a test
*/
func (launcher TestsuiteContainerLauncher) LaunchTestRunningContainers(
		ctx context.Context,
		log *logrus.Logger,
		dockerManager *docker_manager.DockerManager,
		networkId string,
		subnetMask string,
		gatewayIpAddr net.IP,
		testName string,
		kurtosisApiContainerIp net.IP,
		testsuiteContainerIp net.IP,
		testSetupTimeoutInSeconds uint32,
		testRunTimeoutInSeconds uint32,
		isPartitioningEnabled bool) (testsuiteContainerId string, kurtosisApiContainerId string, resultErr error){
	log.Debugf(
		"Test suite container IP: %v; kurtosis API container IP: %v",
		testsuiteContainerIp.String(),
		kurtosisApiContainerIp.String())

	functionCompletedSuccessfully := false

	testSuiteEnvVars, err := launcher.generateTestSuiteEnvVars(kurtosisApiContainerIp.String())
	if err != nil {
		return "", "", stacktrace.Propagate(err, "An error occurred generating the test-running testsuite container env vars")
	}

	suiteContainerDesc := "test-running testsuite container"
	log.Infof("Launching %v....", suiteContainerDesc)
	suiteContainerId, debuggerPortHostBinding, err := launcher.createAndStartTestsuiteContainerWithDebuggingPortIfNecessary(
		ctx,
		dockerManager,
		networkId,
		testsuiteContainerIp,
		testSuiteEnvVars)
	if err != nil {
		return "", "", stacktrace.Propagate(err, "An error occurred creating the test-running testsuite container")
	}
	defer killContainerIfNotFunctionSuccess(
		ctx,
		log,
		dockerManager,
		suiteContainerId,
		func() bool { return functionCompletedSuccessfully },
	)
	logSuccessfulSuiteContainerLaunch(log, suiteContainerDesc, debuggerPortHostBinding)


	apiContainerEnvVars, err := launcher.genTestExecutionApiContainerEnvVars(
		networkId,
		subnetMask,
		gatewayIpAddr,
		testName,
		suiteContainerId,
		testsuiteContainerIp,
		kurtosisApiContainerIp,
		testSetupTimeoutInSeconds,
		testRunTimeoutInSeconds,
		isPartitioningEnabled)
	if err != nil {
		return "", "", stacktrace.Propagate(err, "An error occurred generating the API container's environment variables")
	}

	log.Info("Launching Kurtosis API container...")
	kurtosisApiPort := nat.Port(fmt.Sprintf("%v/%v", api_container_server_consts.ListenPort, api_container_server_consts.ListenProtocol))
	kurtosisApiContainerId, err = dockerManager.CreateAndStartContainer(
		ctx,
		launcher.kurtosisApiImage,
		networkId,
		kurtosisApiContainerIp,
		map[docker_manager.ContainerCapability]bool{}, // No extra capabilities needed for the API container
		docker_manager.DefaultNetworkMode,
		map[nat.Port]*nat.PortBinding{
			kurtosisApiPort: nil,
		},
		nil, // Nil ENTRYPOINT args because the API container is launched by setting env vars
		nil, // Nil CMD args because the API container is launched by setting env vars
		apiContainerEnvVars,
		map[string]string{
			dockerSocket: dockerSocket,
		},
		map[string]string{
			launcher.suiteExecutionVolName: api_container_mountpoints.SuiteExecutionVolumeMountDirpath,
		},
		launcher.hostPortBindingSupplier != nil, // If we're expecting ot dole out host ports, the API container WILL need access to the host machine running Docker
	)
	defer killContainerIfNotFunctionSuccess(
		ctx,
		log,
		dockerManager,
		kurtosisApiContainerId,
		func() bool { return functionCompletedSuccessfully })
	if err != nil {
		return "", "", stacktrace.Propagate(err, "An error occurred launching the Kurtosis API container")
	}
	log.Infof("Successfully launched the Kurtosis API container")

	functionCompletedSuccessfully = true
	return suiteContainerId, kurtosisApiContainerId, nil
}

// ===============================================================================================
//                                 Private helper functions
// ===============================================================================================
func (launcher TestsuiteContainerLauncher) createAndStartTestsuiteContainerWithDebuggingPortIfNecessary(
		ctx context.Context,
		dockerManager *docker_manager.DockerManager,
		networkId string,
		containerIpAddr net.IP,
		envVars map[string]string,
	) (containerId string, debuggingPortBindingOnHost *nat.PortBinding, resultErr error) {

	hostPortBindings := map[nat.Port]*nat.PortBinding{}
	var debuggerPortBinding *nat.PortBinding = nil
	if launcher.hostPortBindingSupplier != nil {
		freePortBinding, err := launcher.hostPortBindingSupplier.GetFreePortBinding()
		if err != nil {
			return "", nil, stacktrace.Propagate(err, "An error occurred getting a free host port binding for the testsuite container")
		}
		debuggerPortBinding = &freePortBinding

		debuggerPortInsideTestsuite, err := nat.NewPort(protocolForDebuggersRunningOnTestsuite, strconv.Itoa(portForDebuggersRunningOnTestsuite))
		if err != nil {
			return "", nil, stacktrace.Propagate(
				err,
				"An error occurred creating the port object on '%v/%v' to represent the debugger's port inside the testsuite",
				protocolForDebuggersRunningOnTestsuite,
				portForDebuggersRunningOnTestsuite,
			)
		}

		hostPortBindings[debuggerPortInsideTestsuite] = debuggerPortBinding
	}

	containerId, err := dockerManager.CreateAndStartContainer(
		ctx,
		launcher.testsuiteImage,
		networkId,
		containerIpAddr,
		map[docker_manager.ContainerCapability]bool{}, 	// No extra capabilities needed for testsuite containers
		docker_manager.DefaultNetworkMode,  			// No special networking modes for testsuite containers
		hostPortBindings,
		nil, // Nil ENTRYPOINT args because we expect the test suite image to be parameterized with variables
		nil, // Nil CMD args because we expect the test suite image to be parameterized with variables
		envVars,
		map[string]string{}, 		// No bind mounts for a testsuite container
		map[string]string{
			launcher.suiteExecutionVolName: test_suite_container_mountpoints.TestsuiteContainerSuiteExVolMountpoint,
		},
		false, // The testsuite container should never be able to access the machine hosting Docker
	)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred creating and starting the testsuite container")
	}

	return containerId, debuggerPortBinding, nil
}

func logSuccessfulSuiteContainerLaunch(
		log *logrus.Logger,
		suiteContainerDesc string,
		debuggerPortHostBinding *nat.PortBinding) {
	suiteLaunchSupplementalLogInfo := ""
	if debuggerPortHostBinding != nil {
		suiteLaunchSupplementalLogInfo = fmt.Sprintf(
			" with debugger port bound to host port %v:%v (if a debugger is running, you may need to connect to this port to proceed)",
			debuggerPortHostBinding.HostIP,
			debuggerPortHostBinding.HostPort,
		)
	}
	log.Infof("Successfully created %v%v", suiteContainerDesc, suiteLaunchSupplementalLogInfo)

}


/*
Generates the map of environment variables needed to run a test suite container

NOTE: exactly one of metadata_filepath or test_name must be non-empty!
*/
func (launcher TestsuiteContainerLauncher) generateTestSuiteEnvVars(kurtosisApiIp string) (map[string]string, error) {
	debuggerPortIntStr := strconv.Itoa(launcher.debuggerPortObj.Int())
	kurtosisApiSocket := fmt.Sprintf("%v:%v", kurtosisApiIp, api_container_server_consts.ListenPort)
	// TODO switch to the envVars requiring a visitor to hit, so we get them all
	standardVars := map[string]string{
		test_suite_env_vars.KurtosisApiSocketEnvVar: kurtosisApiSocket,
		test_suite_env_vars.LogLevelEnvVar:          launcher.suiteLogLevel,
		test_suite_env_vars.DebuggerPortEnvVar:      debuggerPortIntStr,
		test_suite_env_vars.CustomParamsJson: launcher.customParamsJson,
	}
	return standardVars, nil
}

func genApiContainerEnvVars(
		logLevel logrus.Level,
		mode api_container_modes.ApiContainerMode,
		paramsJsonStr string) map[string]string {
	// TODO switch to the envVars requiring a visitor to hit, so we get them all
	return map[string]string{
		api_container_env_vars.LogLevelEnvVar: logLevel.String(),
		api_container_env_vars.ModeEnvVar: string(mode),
		api_container_env_vars.ParamsJsonEnvVar: paramsJsonStr,
	}
}

func (launcher TestsuiteContainerLauncher) genTestExecutionApiContainerEnvVars(
		networkId string,
		subnetMask string,
		gatewayIpAddr net.IP,
		testName string,
		testSuiteContainerId string,
		testSuiteContainerIpAddr net.IP,
		apiContainerIpAddr net.IP,
		testSetupTimeoutInSeconds uint32,
		testRunTimeoutInSeconds uint32,
		isPartitioningEnabled bool) (map[string]string, error) {
	var hostPortBindingSupplierParams *api_container_params_json.HostPortBindingSupplierParams = nil
	hostPortBindingSupplier := launcher.hostPortBindingSupplier
	if hostPortBindingSupplier != nil {
		hostPortBindingSupplierParams = &api_container_params_json.HostPortBindingSupplierParams{
			InterfaceIp:    hostPortBindingSupplier.GetInterfaceIp(),
			Protocol:       hostPortBindingSupplier.GetProtocol(),
			PortRangeStart: hostPortBindingSupplier.GetPortRangeStart(),
			PortRangeEnd:   hostPortBindingSupplier.GetPortRangeEnd(),
		}
	}

	args, err := api_container_params_json.NewTestExecutionArgs(
		launcher.executionInstanceId.String(),
		networkId,
		subnetMask,
		gatewayIpAddr.String(),
		testName,
		launcher.suiteExecutionVolName,
		testSuiteContainerId,
		testSuiteContainerIpAddr.String(),
		apiContainerIpAddr.String(),
		testSetupTimeoutInSeconds,
		testRunTimeoutInSeconds,
		isPartitioningEnabled,
		hostPortBindingSupplierParams)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the test execution args")
	}

	argsBytes, err := json.Marshal(args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred serializing API container test execution args to JSON")
	}

	argsStr := string(argsBytes)
	return genApiContainerEnvVars(
		launcher.kurtosisApiLogLevel,
		api_container_modes.TestExecutionMode,
		argsStr), nil
}

func (launcher TestsuiteContainerLauncher) genSuiteMetadataSerializationApiContainerEnvVars() (map[string]string, error) {
	args, err := api_container_params_json.NewSuiteMetadataSerializationArgs()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the suite metadata serialization args")
	}

	argsBytes, err := json.Marshal(args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred serializing API container suite metadata-serializing args to JSON")
	}

	argsStr := string(argsBytes)
	return genApiContainerEnvVars(
		launcher.kurtosisApiLogLevel,
		api_container_modes.SuiteMetadataSerializingMode,
		argsStr), nil

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
