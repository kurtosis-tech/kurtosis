package docker

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/api_container"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
)

func (backendCore *DockerKurtosisBackend) CreateAPIContainer(
	ctx context.Context,
	image string,
	grpcPortId string,
	grpcPortSpec *port_spec.PortSpec,
	grpcProxyPortId string,
	grpcProxyPortSpec *port_spec.PortSpec,
	enclaveDataDirpathOnHostMachine string,
	envVars map[string]string,
) (*api_container.APIContainer, error) {
	/*
	objAttrs, err := backend.objAttrsProvider.ForEngineServer()

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

	envVars, err := args.GetEnvFromArgs(argsObj)
	if err != nil {
		return "", nil, nil, nil, stacktrace.Propagate(err, "An error occurred generating the API container's environment variables")
	}

	containerImageAndTag := fmt.Sprintf(
		"%v:%v",
		containerImage,
		imageVersionTag,
	)

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
	*/
	panic("implement me")
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