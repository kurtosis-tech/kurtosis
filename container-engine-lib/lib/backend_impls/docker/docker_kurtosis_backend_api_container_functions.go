package docker

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/api_container"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
	"strconv"
)

const (
	// The location where the enclave data directory (on the Docker host machine) will be bind-mounted
	//  on the API container
	enclaveDataDirpathOnAPIContainer = "/kurtosis-enclave-data"
)

func (backendCore *DockerKurtosisBackend) CreateAPIContainer(
	ctx context.Context,
	image string,
	enclaveId string,
	ipAddr net.IP, // TODO REMOVE THIS ONCE WE FIX THE STATIC IP PROBLEM!!
	grpcPortId string,
	grpcPortSpec *port_spec.PortSpec,
	grpcProxyPortId string,
	grpcProxyPortSpec *port_spec.PortSpec,
	enclaveDataDirpathOnHostMachine string,
	envVars map[string]string,
) (*api_container.APIContainer, error) {
	// Verify no API container already exists in the enclave
	apiContainersInEnclaveFilters := &api_container.APIContainerFilters{
		EnclaveIDs: map[string]bool{
			enclaveId: true,
		},
	}
	preexistingApiContainersInEnclave, err := backendCore.GetAPIContainers(ctx, apiContainersInEnclaveFilters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred checking if API containers already exist in encalve '%v'", enclaveId)
	}
	if len(preexistingApiContainersInEnclave) > 0 {
		return nil, stacktrace.NewError("Found existing API container(s) in enclave '%v'; cannot start a new one", enclaveId)
	}

	// Get the Docker network ID where we'll start the new API container
	matchingNetworks, err := backendCore.dockerManager.GetNetworksByLabels(ctx, map[string]string{
		label_key_consts.IDLabelKey.GetString(): enclaveId,
	})
	numMatchingNetworks := len(matchingNetworks)
	if numMatchingNetworks == 0 {
		return nil, stacktrace.NewError("No network found for enclave with ID '%v'", enclaveId)
	}
	if numMatchingNetworks > 1 {
		return nil, stacktrace.NewError("Found '%v' enclave networks with ID '%v', which shouldn't happen", numMatchingNetworks, enclaveId)
	}
	enclaveNetwork := matchingNetworks[0]

	enclaveObjAttrProvider, err := backendCore.objAttrsProvider.ForEnclave(enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Couldn't get an object attribute provider for enclave '%v'", enclaveId)
	}

	apiContainerAttrs, err := enclaveObjAttrProvider.ForApiContainer(
		ipAddr,
		grpcPortId,
		grpcPortSpec,
		grpcProxyPortId,
		grpcProxyPortSpec,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the object attributes for the API container")
	}

	privateGrpcDockerPort, err := transformPortSpecToDockerPort(grpcPortSpec)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred transforming the private grpc port spec to a Docker port")
	}
	privateGrpcProxyDockerPort, err := transformPortSpecToDockerPort(grpcProxyPortSpec)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred transforming the private grpc proxy port spec to a Docker port")
	}
	usedPorts := map[nat.Port]docker_manager.PortPublishSpec{
		privateGrpcDockerPort:      docker_manager.NewManualPublishingSpec(grpcPortSpec.GetNumber()),
		privateGrpcProxyDockerPort: docker_manager.NewManualPublishingSpec(grpcProxyPortSpec.GetNumber()),
	}

	bindMounts := map[string]string{
		// Necessary so that the API container can interact with the Docker engine
		dockerSocketFilepath:           dockerSocketFilepath,
		enclaveDataDirpathOnHostMachine: enclaveDataDirpathOnAPIContainer,
	}

	labelStrs := map[string]string{}
	for labelKey, labelValue := range apiContainerAttrs.GetLabels() {
		labelStrs[labelKey.GetString()] = labelValue.GetString()
	}

	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		image,
		apiContainerAttrs.GetName().GetString(),
		enclaveNetwork.GetId(),
	).WithEnvironmentVariables(
		envVars,
	).WithBindMounts(
		bindMounts,
	).WithUsedPorts(
		usedPorts,
	).WithLabels(
		labelStrs,
	).Build()

	// Best-effort pull attempt
	if err = backendCore.dockerManager.PullImage(ctx, image); err != nil {
		logrus.Warnf("Failed to pull the latest version of API container image '%v'; you may be running an out-of-date version", containerImageAndTag)
	}

	// Best-effort pull attempt
	if err = backendCore.dockerManager.PullImage(ctx, containerImageAndTag); err != nil {
		logrus.Warnf("Failed to pull the latest version of engine server image '%v'; you may be running an out-of-date version", containerImageAndTag)
	}

	containerId, hostMachinePortBindings, err := backendCore.dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred starting the Kurtosis engine container")
	}
	shouldKillEngineContainer := true
	defer func() {
		if shouldKillEngineContainer {
			// NOTE: We use the background context here so that the kill will still go off even if the reason for
			// the failure was the original context being cancelled
			if err := backendCore.dockerManager.KillContainer(context.Background(), containerId); err != nil {
				logrus.Errorf(
					"Launching the engine server with ID '%v' and container ID '%v' didn't complete successfully so we " +
						"tried to kill the container we started, but doing so exited with an error:\n%v",
					engineIdStr,
					containerId,
					err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually stop engine server with ID '%v'!!!!!!", engineIdStr)
			}
		}
	}()



	/*



	objAttrsSupplier := func(enclaveObjAttrsProvider schema.EnclaveObjectAttributesProvider) (schema.ObjectAttributes, error) {
		apiContainerAttrs, err := enclaveObjAttrsProvider.ForApiContainer(
			apiContainerIpAddr,
			grpcPortNum,
			grpcProxyPortNum,
		)
		if err != nil {
			return nil, stacktrace.Propagate(
				err,
				"An error occurred getting the API container object attributes using port num '%v' and proxy port '%v'",
				grpcPortNum,
				grpcProxyPortNum,
			)
		}
		return apiContainerAttrs, nil
	}

	takenIpAddrStrSet := map[string]bool{
		gatewayIpAddr.String():      true,
		apiContainerIpAddr.String(): true,
	}
	argsObj, err := args.NewAPIContainerArgs(
		imageVersionTag,
		logLevel.String(),
		grpcPortNum,
		grpcProxyPortNum,
		enclaveId,
		networkId,
		subnetMask,
		apiContainerIpAddr.String(),
		takenIpAddrStrSet,
		isPartitioningEnabled,
		enclaveDataDirpathOnAPIContainer,
		enclaveDataDirpathOnHostMachine,
		metricsUserID,
		didUserAcceptSendingMetrics,
	)
	if err != nil {
		return "", nil, nil, nil, stacktrace.Propagate(err, "An error occurred creating the API container args")
	}

	 */

	/*
	grpcPort, err := enclave_container_launcher.NewEnclaveContainerPort(grpcPortNum, enclave_container_launcher.EnclaveContainerPortProtocol_TCP)
	if err != nil {
		return "", nil, nil, nil, stacktrace.Propagate(err, "An error occurred constructing the enclave container port object representing the API container's gRPC port '%v'", grpcPortNum)
	}

	grpcProxyPort, err := enclave_container_launcher.NewEnclaveContainerPort(grpcProxyPortNum, enclave_container_launcher.EnclaveContainerPortProtocol_TCP)
	if err != nil {
		return "", nil, nil, nil, stacktrace.Propagate(err, "An error occurred constructing the enclave container port object representing the API container's gRPC port with portNum '%v' and grpcPortNum '%v'", grpcPortNum, grpcProxyPortNum)
	}

	privatePorts := map[string]*enclave_container_launcher.EnclaveContainerPort{
		schema.KurtosisInternalContainerGRPCPortID:      grpcPort,
		schema.KurtosisInternalContainerGRPCProxyPortID: grpcProxyPort,
	}

	log.Debugf("Launching Kurtosis API container...")
	containerId, publicIpAddr, publicPorts, err := launcher.enclaveContainerLauncher.Launch(
		ctx,
		containerImageAndTag,
		shouldPullImageBeforeLaunching,
		apiContainerIpAddr,
		networkId,
		enclaveDataDirpathOnAPIContainer,
		privatePorts,
		objAttrsSupplier,
		envVars,
		shouldBindMountDockerSocket,
		containerAlias,
		entrypointArgs,
		cmdArgs,
		volumeMounts,
	)
	if err != nil {
		return "", nil, nil, nil, stacktrace.Propagate(err, "An error occurred launching the API container")
	}
	 */

	portIdsForDockerPortObjs, publishSpecs, err := getPortMapsBeforeContainerStart(privatePorts)
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred getting the ports maps required for starting an enclave container")
	}

	bindMounts := map[string]string{
		launcher.enclaveDataDirpathOnHostMachine: enclaveDataDirMountDirpath,
	}
	if shouldBindMountDockerSocket {
		bindMounts[dockerSocketFilepath] = dockerSocketFilepath
	}

	objectAttributes, err := objectAttributesSupplier(launcher.enclaveObjAttrsProvider)
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred getting the container attributes using the supplier")
	}

	containerName := objectAttributes.GetName()
	containerLabels := objectAttributes.GetLabels()
	createAndStartArgsBuilder := docker_manager.NewCreateAndStartContainerArgsBuilder(
		image,
		containerName,
		dockerNetworkId,
	).WithStaticIP(
		ipAddr,
	).WithUsedPorts(
		publishSpecs,
	).WithEnvironmentVariables(
		environmentVariables,
	).WithBindMounts(
		bindMounts,
	).WithLabels(
		containerLabels,
	)
	if maybeAlias != "" {
		createAndStartArgsBuilder.WithAlias(maybeAlias)
	}
	if maybeEntrypointArgs != nil {
		createAndStartArgsBuilder.WithEntrypointArgs(maybeEntrypointArgs)
	}
	if maybeCmdArgs != nil {
		createAndStartArgsBuilder.WithCmdArgs(maybeCmdArgs)
	}
	if maybeVolumeMounts != nil {
		createAndStartArgsBuilder.WithVolumeMounts(maybeVolumeMounts)
	}
	createAndStartArgs := createAndStartArgsBuilder.Build()
	containerId, hostPortBindingsByPortObj, err := launcher.dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred starting the Docker container for service with image '%v'", image)
	}
	shouldKillContainer := true
	defer func() {
		if shouldKillContainer {
			if err := launcher.dockerManager.KillContainer(context.Background(), containerId); err != nil {
				logrus.Error("Launching the service container failed, but an error occurred killing container we started:")
				fmt.Fprintln(logrus.StandardLogger().Out, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually kill container with ID '%v'", containerId)
			}
		}
	}()

	var maybePublicIpAddr net.IP = nil
	publicPorts := map[string]*EnclaveContainerPort{}
	if len(privatePorts) > 0 {
		maybePublicIpAddr, publicPorts, err = condensePublicNetworkInfoFromHostMachineBindings(
			hostPortBindingsByPortObj,
			privatePorts,
			portIdsForDockerPortObjs,
		)
		if err != nil {
			return "", nil, nil, stacktrace.Propagate(err, "An error occurred extracting public IP addr & ports from the host machine ports returned by the container engine")
		}
	}

	shouldKillContainer := true
	defer func() {
		if shouldKillContainer {
			if killErr := launcher.dockerManager.KillContainer(context.Background(), containerId); killErr != nil {
				logrus.Errorf("The function to create the API container didn't finish successful so we tried to kill the container we created, but the killing threw an error:")
				logrus.Error(killErr)
			}
		}
	}()

	if err := waitForAvailability(ctx, launcher.dockerManager, containerId, grpcPortNum); err != nil {
		return "", nil, nil, nil, stacktrace.Propagate(err, "An error occurred waiting for the API container's grpc port to become available")
	}

	if err := waitForAvailability(ctx, launcher.dockerManager, containerId, grpcProxyPortNum); err != nil {
		return "", nil, nil, nil, stacktrace.Propagate(err, "An error occurred waiting for the API container's grpc-proxy port to become available")
	}

	publicGrpcPort, found := publicPorts[schema.KurtosisInternalContainerGRPCPortID]
	if !found {
		return "", nil, nil, nil, stacktrace.NewError("No public port was found for '%v' - this is very strange!", schema.KurtosisInternalContainerGRPCPortID)
	}

	publicGrpcProxyPort, found := publicPorts[schema.KurtosisInternalContainerGRPCProxyPortID]
	if !found {
		return "", nil, nil, nil, stacktrace.NewError("No public port was found for '%v' - this is very strange!", schema.KurtosisInternalContainerGRPCProxyPortID)
	}

	shouldKillContainer = false
	return containerId, publicIpAddr, publicGrpcPort, publicGrpcProxyPort, nil
}

func (backendCore *DockerKurtosisBackend) GetAPIContainers(ctx context.Context, filters *api_container.APIContainerFilters) (map[string]*api_container.APIContainer, error) {
	//TODO implement me
	panic("implement me")
}

func (backendCore *DockerKurtosisBackend) StopAPIContainers(ctx context.Context, filters *enclave.EnclaveFilters) (successApiContainerIds map[string]bool, erroredApiContainerIds map[string]error, resultErr error) {
	//TODO implement me
	panic("implement me")
}

func (backendCore *DockerKurtosisBackend) DestroyAPIContainers(ctx context.Context, filters *enclave.EnclaveFilters) (successApiContainerIds map[string]bool, erroredApiContainerIds map[string]error, resultErr error) {
	//TODO implement me
	panic("implement me")
}

// ====================================================================================================
//                                      Private Helper Functions
// ====================================================================================================
// Takes in the ports used by a container and provides the necessary maps required for:
//  1) getting the container's labels
//  2) starting the service
//  3) getting the service's host machine port bindings after the service is started
func getPortMapsBeforeContainerStart(
	privatePorts map[string]*port_spec.PortSpec,
) (
	resultPortIdsForDockerPortObjs map[nat.Port]string,
	resultPublishSpecs map[nat.Port]docker_manager.PortPublishSpec, // Used by container engine
	resultErr error,
) {
	portIdsForDockerPortObjs := map[nat.Port]string{}
	publishSpecs := map[nat.Port]docker_manager.PortPublishSpec{}
	for portId, enclaveContainerPort := range privatePorts {
		portNum := enclaveContainerPort.GetNumber()
		portProto := enclaveContainerPort.GetProtocol()

		dockerPortObj, err := nat.NewPort(
			string(portProto),
			fmt.Sprintf("%v", portNum),
		)
		if err != nil {
			return nil, nil, stacktrace.Propagate(
				err,
				"An error occurred creating a Docker port object using port num '%v' and protocol string '%v'",
				portNum,
				portProto,
			)
		}

		if preexistingPortId, found := portIdsForDockerPortObjs[dockerPortObj]; found {
			return nil, nil, stacktrace.NewError(
				"Port '%v' declares Docker port spec '%v', but this port spec is already in use by port '%v'",
				portId,
				dockerPortObj,
				preexistingPortId,
			)
		}
		portIdsForDockerPortObjs[dockerPortObj] = portId

		publishSpecs[dockerPortObj] = docker_manager.NewAutomaticPublishingSpec()
	}
	return portIdsForDockerPortObjs, publishSpecs, nil
}

// condensePublicNetworkInfoFromHostMachineBindings
// Condenses declared private port bindings and the host machine port bindings returned by the container engine lib into:
//  1) a single host machine IP address
//  2) a map of private port binding IDs -> public ports
// An error is thrown if there are multiple host machine IP addresses
func condensePublicNetworkInfoFromHostMachineBindings(
	hostMachinePortBindings map[nat.Port]*nat.PortBinding,
	privatePorts map[string]*EnclaveContainerPort,
	portIdsForDockerPortObjs map[nat.Port]string,
) (
	resultPublicIpAddr net.IP,
	resultPublicPorts map[string]*EnclaveContainerPort,
	resultErr error,
) {
	if len(hostMachinePortBindings) == 0 {
		return nil, nil, stacktrace.NewError("Cannot condense public network info if no host machine port bindings are provided")
	}

	publicIpAddrStr := uninitializedPublicIpAddrStrValue
	publicPorts := map[string]*EnclaveContainerPort{}
	for dockerPortObj, hostPortBinding := range hostMachinePortBindings {
		portId, found := portIdsForDockerPortObjs[dockerPortObj]
		if !found {
			// If the container engine reports a host port binding that wasn't declared in the input used-ports object, ignore it
			// This could happen if a port is declared in the Dockerfile
			continue
		}

		privatePort, found := privatePorts[portId]
		if !found {
			return nil,  nil, stacktrace.NewError(
				"The container engine returned a host machine port binding for Docker port spec '%v', but this port spec didn't correspond to any port ID; this is very likely a bug in Kurtosis",
				dockerPortObj,
			)
		}

		hostIpAddr := hostPortBinding.HostIP
		if publicIpAddrStr == uninitializedPublicIpAddrStrValue {
			publicIpAddrStr = hostIpAddr
		} else if publicIpAddrStr != hostIpAddr {
			return nil, nil, stacktrace.NewError(
				"A public IP address '%v' was already declared for the service, but Docker port object '%v' declares a different public IP address '%v'",
				publicIpAddrStr,
				dockerPortObj,
				hostIpAddr,
			)
		}

		hostPortStr := hostPortBinding.HostPort
		hostPortUint64, err := strconv.ParseUint(hostPortStr, enclaveContainerPortNumUintBase, encalveContainerPortNumUintBits)
		if err != nil {
			return nil, nil, stacktrace.Propagate(
				err,
				"An error occurred parsing host machine port string '%v' into a uint with %v bits and base %v",
				hostPortStr,
				encalveContainerPortNumUintBits,
				enclaveContainerPortNumUintBase,
			)
		}
		hostPortUint16 := uint16(hostPortUint64) // Safe to do because our ParseUint declares the expected number of bits
		portProto := privatePort.GetProtocol()
		publicPort, err := NewEnclaveContainerPort(hostPortUint16, portProto)
		if err != nil {
			return nil, nil, stacktrace.Propagate(
				err,
				"An error occurred creating public port object with num '%v' and protocol '%v'; this should never happen and likely means a bug in Kurtosis",
				hostPortUint16,
				portProto,
			)
		}
		publicPorts[portId] = publicPort
	}
	if publicIpAddrStr == uninitializedPublicIpAddrStrValue {
		return nil, nil, stacktrace.NewError("No public IP address string was retrieved from host port bindings: %+v", hostMachinePortBindings)
	}
	publicIpAddr := net.ParseIP(publicIpAddrStr)
	if publicIpAddr == nil {
		return nil, nil, stacktrace.NewError("Couldn't parse service's public IP address string '%v' to an IP object", publicIpAddrStr)
	}
	return publicIpAddr, publicPorts, nil
}
