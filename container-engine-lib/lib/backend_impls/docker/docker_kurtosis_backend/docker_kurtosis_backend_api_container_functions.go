package docker_kurtosis_backend

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_registry_spec"
	"net"
	"time"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_key"

	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_operation_parallelizer"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_port_spec_serializer"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/api_container"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/network_helpers"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	// The API container uses gRPC so MUST listen on TCP (no other protocols are supported)
	apiContainerTransportProtocol = port_spec.TransportProtocol_TCP

	maxWaitForApiContainerAvailabilityRetries         = 10
	timeBetweenWaitForApiContainerAvailabilityRetries = 1 * time.Second

	apicDebugServerPort = 50103 // in ClI this is 50101 and in engine is 50102
)

// TODO: MIGRATE THIS FOLDER TO USE STRUCTURE OF USER_SERVICE_FUNCTIONS MODULE

var emptyRegistrySpecAsPublicImage *image_registry_spec.ImageRegistrySpec = nil

func (backend *DockerKurtosisBackend) CreateAPIContainer(
	ctx context.Context,
	image string,
	enclaveUuid enclave.EnclaveUUID,
	grpcPortNum uint16,
	// The dirpath on the API container where the enclave data volume should be mounted
	enclaveDataVolumeDirpath string,
	ownIpAddressEnvVar string,
	customEnvVars map[string]string,
	shouldStartInDebugMode bool,
) (*api_container.APIContainer, error) {
	logrus.Debugf("Creating the APIC for enclave '%v'", enclaveUuid)

	// Verify no API container already exists in the enclave
	apiContainersInEnclaveFilters := &api_container.APIContainerFilters{
		EnclaveIDs: map[enclave.EnclaveUUID]bool{
			enclaveUuid: true,
		},
		Statuses: nil,
	}
	preexistingApiContainersInEnclave, err := backend.GetAPIContainers(ctx, apiContainersInEnclaveFilters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred checking if API containers already exist in enclave '%v'", enclaveUuid)
	}
	if len(preexistingApiContainersInEnclave) > 0 {
		return nil, stacktrace.NewError("Found existing API container(s) in enclave '%v'; cannot start a new one", enclaveUuid)
	}

	enclaveDataVolumeName, err := backend.getEnclaveDataVolumeByEnclaveUuid(ctx, enclaveUuid)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the enclave data volume for enclave '%v'", enclaveUuid)
	}

	// Get the Docker network ID where we'll start the new API container
	enclaveNetwork, err := backend.getEnclaveNetworkByEnclaveUuid(ctx, enclaveUuid)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave network by enclave UUID '%v'", enclaveUuid)
	}

	enclaveLogsCollector, err := backend.GetLogsCollectorForEnclave(ctx, enclaveUuid)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting the logs collector for enclave '%v'; This is a bug in Kurtosis", enclaveUuid)
	}

	reverseProxy, err := backend.GetReverseProxy(ctx)
	if reverseProxy == nil {
		return nil, stacktrace.Propagate(err, "The reverse proxy is not running, This is a bug in Kurtosis")
	}
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting the reverse proxy, This is a bug in Kurtosis")
	}
	reverseProxyEnclaveNetworkIpAddress, found := reverseProxy.GetEnclaveNetworksIpAddress()[enclaveNetwork.GetId()]
	if !found {
		return nil, stacktrace.NewError("An error occurred while getting the reverse proxy enclave network IP address for enclave '%v', This is a bug in Kurtosis", enclaveUuid)
	}

	networkCidr := enclaveNetwork.GetIpAndMask()
	alreadyTakenIps := map[string]bool{
		networkCidr.IP.String():                                    true,
		enclaveNetwork.GetGatewayIp():                              true,
		enclaveLogsCollector.GetEnclaveNetworkIpAddress().String(): true,
		reverseProxyEnclaveNetworkIpAddress.String():               true,
	}

	ipAddr, err := network_helpers.GetFreeIpAddrFromSubnet(alreadyTakenIps, networkCidr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting an IP address for the API container")
	}

	// Set the own-IP environment variable
	if _, found := customEnvVars[ownIpAddressEnvVar]; found {
		return nil, stacktrace.NewError("Requested own IP environment variable '%v' conflicts with custom environment variable", ownIpAddressEnvVar)
	}
	envVarsWithOwnIp := map[string]string{
		ownIpAddressEnvVar: ipAddr.String(),
	}
	for key, value := range customEnvVars {
		envVarsWithOwnIp[key] = value
	}

	defaultWait, err := port_spec.CreateWaitWithDefaultValues()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a new wait with default values")
	}

	privateGrpcPortSpec, err := port_spec.NewPortSpec(grpcPortNum, apiContainerTransportProtocol, consts.HttpApplicationProtocol, defaultWait)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the API container's private grpc port spec object using number '%v' and protocol '%v'",
			grpcPortNum,
			apiContainerTransportProtocol,
		)
	}

	enclaveObjAttrProvider, err := backend.objAttrsProvider.ForEnclave(enclaveUuid)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Couldn't get an object attribute provider for enclave '%v'", enclaveUuid)
	}

	apiContainerAttrs, err := enclaveObjAttrProvider.ForApiContainer(
		ipAddr,
		consts.KurtosisInternalContainerGrpcPortId,
		privateGrpcPortSpec,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the object attributes for the API container")
	}

	privateGrpcDockerPort, err := shared_helpers.TransformPortSpecToDockerPort(privateGrpcPortSpec)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred transforming the private grpc port spec to a Docker port")
	}

	usedPorts := map[nat.Port]docker_manager.PortPublishSpec{
		privateGrpcDockerPort: docker_manager.NewAutomaticPublishingSpec(),
	}

	if shouldStartInDebugMode {
		debugServerPortSpec, err := port_spec.NewPortSpec(
			uint16(apicDebugServerPort),
			apiContainerTransportProtocol,
			consts.HttpApplicationProtocol,
			defaultWait,
		)
		if err != nil {
			return nil, stacktrace.Propagate(
				err,
				"An error occurred creating the API container's debug server port spec object using number '%v' and protocol '%v'",
				apicDebugServerPort,
				apiContainerTransportProtocol,
			)
		}

		debugServerDockerPort, err := shared_helpers.TransformPortSpecToDockerPort(debugServerPortSpec)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred transforming the debug server port spec to a Docker port")
		}

		usedPorts[debugServerDockerPort] = docker_manager.NewManualPublishingSpec(uint16(apicDebugServerPort))
	}

	bindMounts := map[string]string{
		// Necessary so that the API container can interact with the Docker engine
		consts.DockerSocketFilepath: consts.DockerSocketFilepath,
	}

	volumeMounts := map[string]string{
		enclaveDataVolumeName: enclaveDataVolumeDirpath,
	}

	labelStrs := map[string]string{}
	for labelKey, labelValue := range apiContainerAttrs.GetLabels() {
		labelStrs[labelKey.GetString()] = labelValue.GetString()
	}
	// TODO: configure the APIContainer to send the logs to the Fluentbit logs collector server

	createAndStartArgsBuilder := docker_manager.NewCreateAndStartContainerArgsBuilder(
		image,
		apiContainerAttrs.GetName().GetString(),
		enclaveNetwork.GetId(),
	).WithEnvironmentVariables(
		envVarsWithOwnIp,
	).WithBindMounts(
		bindMounts,
	).WithVolumeMounts(
		volumeMounts,
	).WithUsedPorts(
		usedPorts,
	).WithStaticIP(
		ipAddr,
	).WithLabels(
		labelStrs,
	).WithRestartPolicy(docker_manager.RestartOnFailure)

	if shouldStartInDebugMode {
		// Adding systrace capabilities when starting the debug server in the engine's container
		capabilities := map[docker_manager.ContainerCapability]bool{
			docker_manager.SysPtrace: true,
		}
		createAndStartArgsBuilder.WithAddedCapabilities(capabilities)

		// Setting security for debugging the engine's container
		securityOpts := map[docker_manager.ContainerSecurityOpt]bool{
			docker_manager.AppArmorUnconfined: true,
		}
		createAndStartArgsBuilder.WithSecurityOpts(securityOpts)
	}

	createAndStartArgs := createAndStartArgsBuilder.Build()

	if _, err = backend.dockerManager.FetchImageIfMissing(ctx, image, emptyRegistrySpecAsPublicImage); err != nil {
		logrus.Warnf("Failed to pull the latest version of API container image '%v'; you may be running an out-of-date version. Error:\n%v", image, err)
	}

	containerId, hostMachinePortBindings, err := backend.dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		logrus.Debugf("Error occurred starting the API container. Err:\n%v", err)
		return nil, stacktrace.Propagate(err, "An error occurred starting the API container")
	}
	shouldKillContainer := true
	defer func() {
		if shouldKillContainer {
			// NOTE: We use the background context here so that the kill will still go off even if the reason for
			// the failure was the original context being cancelled
			if err := backend.dockerManager.KillContainer(context.Background(), containerId); err != nil {
				logrus.Errorf(
					"Launching API container '%v' with container ID '%v' didn't complete successfully so we "+
						"tried to kill the container we started, but doing so exited with an error:\n%v",
					apiContainerAttrs.GetName(),
					containerId,
					err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually stop API container with ID '%v'!!!!!!", containerId)
			}
		}
	}()

	if err := shared_helpers.WaitForPortAvailabilityUsingNetstat(
		ctx,
		backend.dockerManager,
		containerId,
		privateGrpcPortSpec,
		maxWaitForApiContainerAvailabilityRetries,
		timeBetweenWaitForApiContainerAvailabilityRetries,
	); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the API container's grpc port to become available")
	}

	logrus.Debugf("Checking for the APIC availability in enclave '%v'...", enclaveUuid)
	if err := shared_helpers.WaitForPortAvailabilityUsingNetstat(
		ctx,
		backend.dockerManager,
		containerId,
		privateGrpcPortSpec,
		maxWaitForApiContainerAvailabilityRetries,
		timeBetweenWaitForApiContainerAvailabilityRetries,
	); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the API container's grpc port to become available")
	}
	logrus.Debugf("...APIC is available in enclave '%v'", enclaveUuid)

	bridgeNetworkIpAddress, err := backend.dockerManager.GetContainerIP(ctx, consts.NameOfNetworkToStartEngineAndLogServiceContainersIn, containerId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting bridge network ip address for enclave with id: '%v'", enclaveUuid)
	}

	result, err := getApiContainerObjectFromContainerInfo(containerId, labelStrs, types.ContainerStatus_Running, hostMachinePortBindings, bridgeNetworkIpAddress)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating an API container object from container with ID '%v'", containerId)
	}

	logrus.Debugf("APIC for enclave '%v' successfully created", enclaveUuid)

	shouldKillContainer = false
	return result, nil
}

func (backend *DockerKurtosisBackend) GetAPIContainers(ctx context.Context, filters *api_container.APIContainerFilters) (map[enclave.EnclaveUUID]*api_container.APIContainer, error) {
	matchingApiContainers, err := backend.getMatchingApiContainers(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting API containers matching the following filters: %+v", filters)
	}

	matchingApiContainersByEnclaveID := map[enclave.EnclaveUUID]*api_container.APIContainer{}
	for _, apicObj := range matchingApiContainers {
		matchingApiContainersByEnclaveID[apicObj.GetEnclaveID()] = apicObj
	}
	return matchingApiContainersByEnclaveID, nil
}

func (backend *DockerKurtosisBackend) StopAPIContainers(
	ctx context.Context,
	filters *api_container.APIContainerFilters,
) (
	resultSuccessfulEnclaveIds map[enclave.EnclaveUUID]bool,
	resultErroredEnclaveIds map[enclave.EnclaveUUID]error,
	resultErr error,
) {
	matchingApiContainersByContainerId, err := backend.getMatchingApiContainers(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting API containers matching filters '%+v'", filters)
	}

	var killApiContainerOperation docker_operation_parallelizer.DockerOperation = func(
		ctx context.Context,
		dockerManager *docker_manager.DockerManager,
		dockerObjectId string,
	) error {
		if err := dockerManager.KillContainer(ctx, dockerObjectId); err != nil {
			return stacktrace.Propagate(err, "An error occurred killing API container with ID '%v'", dockerObjectId)
		}
		return nil
	}

	successfulEnclaveIdStrs, erroredEnclaveIdStrs, err := docker_operation_parallelizer.RunDockerOperationInParallelForKurtosisObjects(
		ctx,
		matchingApiContainersByContainerId,
		backend.dockerManager,
		extractEnclaveIdApiContainer,
		killApiContainerOperation,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred killing API containers matching filters '%+v'", filters)
	}

	successfulEnclaveIds := map[enclave.EnclaveUUID]bool{}
	for enclaveIdStr := range successfulEnclaveIdStrs {
		successfulEnclaveIds[enclave.EnclaveUUID(enclaveIdStr)] = true
	}
	erroredEnclaveIds := map[enclave.EnclaveUUID]error{}
	for enclaveIdStr, killErr := range erroredEnclaveIdStrs {
		erroredEnclaveIds[enclave.EnclaveUUID(enclaveIdStr)] = killErr
	}

	return successfulEnclaveIds, erroredEnclaveIds, nil
}

func (backend *DockerKurtosisBackend) DestroyAPIContainers(ctx context.Context, filters *api_container.APIContainerFilters) (successfulApiContainerIds map[enclave.EnclaveUUID]bool, erroredApiContainerIds map[enclave.EnclaveUUID]error, resultErr error) {
	matchingApiContainersByContainerId, err := backend.getMatchingApiContainers(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting API containers matching the following filters: %+v", filters)
	}

	var removeApiContainerOperation docker_operation_parallelizer.DockerOperation = func(
		ctx context.Context,
		dockerManager *docker_manager.DockerManager,
		dockerObjectId string,
	) error {
		if err := dockerManager.RemoveContainer(ctx, dockerObjectId); err != nil {
			return stacktrace.Propagate(err, "An error occurred removing API container with ID '%v'", dockerObjectId)
		}
		return nil
	}

	successfulEnclaveIdStrs, erroredEnclaveIdStrs, err := docker_operation_parallelizer.RunDockerOperationInParallelForKurtosisObjects(
		ctx,
		matchingApiContainersByContainerId,
		backend.dockerManager,
		extractEnclaveIdApiContainer,
		removeApiContainerOperation,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred removing API containers matching filters '%+v'", filters)
	}

	successfulEnclaveIds := map[enclave.EnclaveUUID]bool{}
	for enclaveIdStr := range successfulEnclaveIdStrs {
		successfulEnclaveIds[enclave.EnclaveUUID(enclaveIdStr)] = true
	}
	erroredEnclaveIds := map[enclave.EnclaveUUID]error{}
	for enclaveIdStr, killErr := range erroredEnclaveIdStrs {
		erroredEnclaveIds[enclave.EnclaveUUID(enclaveIdStr)] = killErr
	}

	return successfulEnclaveIds, erroredEnclaveIds, nil
}

// ====================================================================================================
//
//	Private Helper Functions
//
// ====================================================================================================
// Gets API containers matching the search filters, indexed by their container ID
func (backend *DockerKurtosisBackend) getMatchingApiContainers(ctx context.Context, filters *api_container.APIContainerFilters) (map[string]*api_container.APIContainer, error) {

	apiContainerSearchLabels := map[string]string{
		docker_label_key.AppIDDockerLabelKey.GetString():         label_value_consts.AppIDDockerLabelValue.GetString(),
		docker_label_key.ContainerTypeDockerLabelKey.GetString(): label_value_consts.APIContainerContainerTypeDockerLabelValue.GetString(),
		// NOTE: we do NOT use the enclave UUID label here, and instead do postfiltering, because Docker has no way to do disjunctive search!
	}
	allApiContainers, err := backend.dockerManager.GetContainersByLabels(ctx, apiContainerSearchLabels, consts.ShouldFetchAllContainersWhenRetrievingContainers)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred fetching API containers using labels: %+v", apiContainerSearchLabels)
	}

	allMatchingApiContainers := map[string]*api_container.APIContainer{}
	for _, apiContainer := range allApiContainers {
		containerId := apiContainer.GetId()
		bridgeNetworkIpAddress, err := backend.dockerManager.GetContainerIP(ctx, consts.NameOfNetworkToStartEngineAndLogServiceContainersIn, containerId)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred while getting bridge network ip address for container with id: '%v'", containerId)
		}

		apicObj, err := getApiContainerObjectFromContainerInfo(
			containerId,
			apiContainer.GetLabels(),
			apiContainer.GetStatus(),
			apiContainer.GetHostPortBindings(),
			bridgeNetworkIpAddress,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred converting container with ID '%v' into an API container object", apiContainer.GetId())
		}

		// If the ID filter is specified, drop API containers not matching it
		if filters.EnclaveIDs != nil && len(filters.EnclaveIDs) > 0 {
			if _, found := filters.EnclaveIDs[apicObj.GetEnclaveID()]; !found {
				continue
			}
		}

		// If status filter is specified, drop API containers not matching it
		if filters.Statuses != nil && len(filters.Statuses) > 0 {
			if _, found := filters.Statuses[apicObj.GetStatus()]; !found {
				continue
			}
		}

		allMatchingApiContainers[containerId] = apicObj
	}

	return allMatchingApiContainers, nil
}

func getApiContainerObjectFromContainerInfo(
	containerId string,
	labels map[string]string,
	containerStatus types.ContainerStatus,
	allHostMachinePortBindings map[nat.Port]*nat.PortBinding,
	bridgeNetworkIpAddress string,
) (*api_container.APIContainer, error) {
	enclaveId, found := labels[docker_label_key.EnclaveUUIDDockerLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected the API container's enclave UUID to be found under label '%v' but the label wasn't present", docker_label_key.EnclaveUUIDDockerLabelKey.GetString())
	}

	privateIpAddrStr, found := labels[docker_label_key.PrivateIPDockerLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError(
			"Couldn't find the API container's private IP using label '%v'",
			docker_label_key.PrivateIPDockerLabelKey.GetString(),
		)
	}
	privateIpAddr := net.ParseIP(privateIpAddrStr)
	if privateIpAddr == nil {
		return nil, stacktrace.NewError("Couldn't parse private IP address string '%v' to an IP", privateIpAddrStr)
	}

	bridgeNetworkIpAddressAddr := net.ParseIP(bridgeNetworkIpAddress)
	if privateIpAddr == nil {
		return nil, stacktrace.NewError("Couldn't parse bridge network IP address string '%v' to an IP", bridgeNetworkIpAddressAddr)
	}

	privateGrpcPortSpec, err := getPrivateApiContainerPorts(labels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the API container's private port specs from container '%v' with labels: %+v", containerId, labels)
	}

	isContainerRunning, found := consts.IsContainerRunningDeterminer[containerStatus]
	if !found {
		// This should never happen because we enforce completeness in a unit test
		return nil, stacktrace.NewError("No is-running designation found for API container status '%v'; this is a bug in Kurtosis!", containerStatus.String())
	}
	var apiContainerStatus container.ContainerStatus
	if isContainerRunning {
		apiContainerStatus = container.ContainerStatus_Running
	} else {
		apiContainerStatus = container.ContainerStatus_Stopped
	}

	var publicIpAddr net.IP
	var publicGrpcPortSpec *port_spec.PortSpec
	if apiContainerStatus == container.ContainerStatus_Running {
		publicGrpcPortIpAddr, candidatePublicGrpcPortSpec, err := shared_helpers.GetPublicPortBindingFromPrivatePortSpec(privateGrpcPortSpec, allHostMachinePortBindings)
		if err != nil {
			return nil, stacktrace.Propagate(err, "The engine is running, but an error occurred getting the public port spec for the engine's grpc private port spec")
		}
		publicGrpcPortSpec = candidatePublicGrpcPortSpec
		publicIpAddr = publicGrpcPortIpAddr
	}

	result := api_container.NewAPIContainer(
		enclave.EnclaveUUID(enclaveId),
		apiContainerStatus,
		privateIpAddr,
		privateGrpcPortSpec,
		publicIpAddr,
		publicGrpcPortSpec,
		bridgeNetworkIpAddressAddr,
	)

	return result, nil
}

func getPrivateApiContainerPorts(containerLabels map[string]string) (
	resultGrpcPortSpec *port_spec.PortSpec,
	resultErr error,
) {
	serializedPortSpecs, found := containerLabels[docker_label_key.PortSpecsDockerLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected to find port specs label '%v' but none was found", docker_label_key.PortSpecsDockerLabelKey.GetString())
	}

	portSpecs, err := docker_port_spec_serializer.DeserializePortSpecs(serializedPortSpecs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred deserializing port specs string '%v'", serializedPortSpecs)
	}

	grpcPortSpec, foundGrpcPort := portSpecs[consts.KurtosisInternalContainerGrpcPortId]
	if !foundGrpcPort {
		return nil, stacktrace.NewError("No grpc port with ID '%v' found in the port specs", consts.KurtosisInternalContainerGrpcPortId)
	}

	return grpcPortSpec, nil
}

func extractEnclaveIdApiContainer(apiContainer *api_container.APIContainer) string {
	return string(apiContainer.GetEnclaveID())
}
