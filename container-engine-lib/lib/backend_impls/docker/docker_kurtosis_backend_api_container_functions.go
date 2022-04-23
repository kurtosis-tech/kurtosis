package docker

import (
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_task_parallelizer"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/port_spec_serializer"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/api_container"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

const (
	// The location where the enclave data directory (on the Docker host machine) will be bind-mounted
	//  on the API container
	enclaveDataDirpathOnAPIContainer = "/kurtosis-enclave-data"

	// The API container uses gRPC so MUST listen on TCP (no other protocols are supported), which also
	// means that its grpc-proxy must listen on TCP
	apiContainerPortProtocol = port_spec.PortProtocol_TCP

	maxWaitForApiContainerAvailabilityRetries         = 10
	timeBetweenWaitForApiContainerAvailabilityRetries = 1 * time.Second

	// We use a short timeout so the API container has time to clean up but the user isn't stuck waiting on a long timeout
	apiContainerStopTimeout = 2 * time.Second

	// TODO Delete this after 2022-05-28
	pre_2022_03_28_IpAddrLabel = "com.kurtosistech.api-container-ip"
)

func (backend *DockerKurtosisBackend) CreateAPIContainer(
	ctx context.Context,
	image string,
	enclaveId enclave.EnclaveID,
	ipAddr net.IP, // TODO REMOVE THIS ONCE WE FIX THE STATIC IP PROBLEM!!
	grpcPortNum uint16,
	grpcProxyPortNum uint16,
	enclaveDataDirpathOnHostMachine string,
	envVars map[string]string,
) (*api_container.APIContainer, error) {
	// Verify no API container already exists in the enclave
	apiContainersInEnclaveFilters := &api_container.APIContainerFilters{
		EnclaveIDs: map[enclave.EnclaveID]bool{
			enclaveId: true,
		},
	}
	preexistingApiContainersInEnclave, err := backend.GetAPIContainers(ctx, apiContainersInEnclaveFilters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred checking if API containers already exist in enclave '%v'", enclaveId)
	}
	if len(preexistingApiContainersInEnclave) > 0 {
		return nil, stacktrace.NewError("Found existing API container(s) in enclave '%v'; cannot start a new one", enclaveId)
	}

	// Get the Docker network ID where we'll start the new API container
	enclaveNetwork, err := backend.getEnclaveNetworkByEnclaveId(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave network by enclave ID '%v'", enclaveId)
	}

	privateGrpcPortSpec, err := port_spec.NewPortSpec(grpcPortNum, apiContainerPortProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the API container's private grpc port spec object using number '%v' and protocol '%v'",
			grpcPortNum,
			enginePortProtocol.String(),
		)
	}
	privateGrpcProxyPortSpec, err := port_spec.NewPortSpec(grpcProxyPortNum, apiContainerPortProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the API container's private grpc proxy port spec object using number '%v' and protocol '%v'",
			grpcProxyPortNum,
			enginePortProtocol.String(),
		)
	}

	enclaveObjAttrProvider, err := backend.objAttrsProvider.ForEnclave(enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Couldn't get an object attribute provider for enclave '%v'", enclaveId)
	}

	apiContainerAttrs, err := enclaveObjAttrProvider.ForApiContainer(
		ipAddr,
		kurtosisInternalContainerGrpcPortId,
		privateGrpcPortSpec,
		kurtosisInternalContainerGrpcProxyPortId,
		privateGrpcProxyPortSpec,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the object attributes for the API container")
	}

	privateGrpcDockerPort, err := transformPortSpecToDockerPort(privateGrpcPortSpec)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred transforming the private grpc port spec to a Docker port")
	}
	privateGrpcProxyDockerPort, err := transformPortSpecToDockerPort(privateGrpcProxyPortSpec)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred transforming the private grpc proxy port spec to a Docker port")
	}
	usedPorts := map[nat.Port]docker_manager.PortPublishSpec{
		privateGrpcDockerPort:      docker_manager.NewAutomaticPublishingSpec(),
		privateGrpcProxyDockerPort: docker_manager.NewAutomaticPublishingSpec(),
	}

	bindMounts := map[string]string{
		// Necessary so that the API container can interact with the Docker engine
		dockerSocketFilepath:            dockerSocketFilepath,
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
	).WithStaticIP(
		ipAddr,
	).WithLabels(
		labelStrs,
	).Build()

	// Best-effort pull attempt
	if err = backend.dockerManager.PullImage(ctx, image); err != nil {
		logrus.Warnf("Failed to pull the latest version of API container image '%v'; you may be running an out-of-date version", image)
	}

	containerId, hostMachinePortBindings, err := backend.dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
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

	if err := waitForPortAvailabilityUsingNetstat(
		ctx,
		backend.dockerManager,
		containerId,
		privateGrpcPortSpec,
		maxWaitForApiContainerAvailabilityRetries,
		timeBetweenWaitForApiContainerAvailabilityRetries,
	); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the API container's grpc port to become available")
	}

	if err := waitForPortAvailabilityUsingNetstat(
		ctx,
		backend.dockerManager,
		containerId,
		privateGrpcPortSpec,
		maxWaitForApiContainerAvailabilityRetries,
		timeBetweenWaitForApiContainerAvailabilityRetries,
	); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the API container's grpc port to become available")
	}

	result, err := getApiContainerObjectFromContainerInfo(containerId, labelStrs, types.ContainerStatus_Running, hostMachinePortBindings)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating an API container object from container with ID '%v'", containerId)
	}

	shouldKillContainer = false
	return result, nil
}

func (backend *DockerKurtosisBackend) GetAPIContainers(ctx context.Context, filters *api_container.APIContainerFilters) (map[enclave.EnclaveID]*api_container.APIContainer, error) {
	matchingApiContainers, err := backend.getMatchingApiContainers(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting API containers matching the following filters: %+v", filters)
	}

	matchingApiContainersByEnclaveID := map[enclave.EnclaveID]*api_container.APIContainer{}
	for _, apicObj := range matchingApiContainers {
		matchingApiContainersByEnclaveID[apicObj.GetEnclaveID()] = apicObj
	}

	return matchingApiContainersByEnclaveID, nil
}

func (backend *DockerKurtosisBackend) StopAPIContainers(
	ctx context.Context,
	filters *api_container.APIContainerFilters,
) (
	resultSuccessfulEnclaveIds map[enclave.EnclaveID]bool,
	resultErroredEnclaveIds map[enclave.EnclaveID]error,
	resultErr error,
) {
	matchingApiContainersByContainerId, err := backend.getMatchingApiContainers(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting API containers matching filters '%+v'", filters)
	}

	// TODO PLEAAASE GO GENERICS... but we can't use 1.18 yet because it'll break all Kurtosis clients :(
	matchingUncastedApiContainersByContainerId := map[string]interface{}{}
	for containerId, apiContainerObj := range matchingApiContainersByContainerId {
		matchingUncastedApiContainersByContainerId[containerId] = interface{}(apiContainerObj)
	}

	var killApiContainerOperation docker_task_parallelizer.DockerOperation = func(
		ctx context.Context,
		dockerManager *docker_manager.DockerManager,
		dockerObjectId string,
	) error {
		if err := dockerManager.KillContainer(ctx, dockerObjectId); err != nil {
			return stacktrace.Propagate(err, "An error occurred killing API container with ID '%v'", dockerObjectId)
		}
		return nil
	}

	successfulEnclaveIdStrs, erroredEnclaveIdStrs, err := docker_task_parallelizer.RunDockerOperationInParallelForKurtosisObjects(
		ctx,
		matchingUncastedApiContainersByContainerId,
		backend.dockerManager,
		extractEnclaveIdFromUncastedApiContainerObj,
		killApiContainerOperation,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred killing API containers matching filters '%+v'", filters)
	}

	successfulEnclaveIds := map[enclave.EnclaveID]bool{}
	for enclaveIdStr := range successfulEnclaveIdStrs {
		successfulEnclaveIds[enclave.EnclaveID(enclaveIdStr)] = true
	}
	erroredEnclaveIds := map[enclave.EnclaveID]error{}
	for enclaveIdStr, killErr := range erroredEnclaveIdStrs {
		erroredEnclaveIds[enclave.EnclaveID(enclaveIdStr)] = killErr
	}

	return successfulEnclaveIds, erroredEnclaveIds, nil
}

func (backend *DockerKurtosisBackend) DestroyAPIContainers(ctx context.Context, filters *api_container.APIContainerFilters) (successfulApiContainerIds map[enclave.EnclaveID]bool, erroredApiContainerIds map[enclave.EnclaveID]error, resultErr error) {
	matchingApiContainersByContainerId, err := backend.getMatchingApiContainers(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting API containers matching the following filters: %+v", filters)
	}

	// TODO PLEAAASE GO GENERICS... but we can't use 1.18 yet because it'll break all Kurtosis clients :(
	matchingUncastedApiContainersByContainerId := map[string]interface{}{}
	for containerId, apiContainerObj := range matchingApiContainersByContainerId {
		matchingUncastedApiContainersByContainerId[containerId] = interface{}(apiContainerObj)
	}

	var removeApiContainerOperation docker_task_parallelizer.DockerOperation = func(
		ctx context.Context,
		dockerManager *docker_manager.DockerManager,
		dockerObjectId string,
	) error {
		if err := dockerManager.RemoveContainer(ctx, dockerObjectId); err != nil {
			return stacktrace.Propagate(err, "An error occurred removing API container with ID '%v'", dockerObjectId)
		}
		return nil
	}

	successfulEnclaveIdStrs, erroredEnclaveIdStrs, err := docker_task_parallelizer.RunDockerOperationInParallelForKurtosisObjects(
		ctx,
		matchingUncastedApiContainersByContainerId,
		backend.dockerManager,
		extractEnclaveIdFromUncastedApiContainerObj,
		removeApiContainerOperation,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred removing API containers matching filters '%+v'", filters)
	}

	successfulEnclaveIds := map[enclave.EnclaveID]bool{}
	for enclaveIdStr := range successfulEnclaveIdStrs {
		successfulEnclaveIds[enclave.EnclaveID(enclaveIdStr)] = true
	}
	erroredEnclaveIds := map[enclave.EnclaveID]error{}
	for enclaveIdStr, killErr := range erroredEnclaveIdStrs {
		erroredEnclaveIds[enclave.EnclaveID(enclaveIdStr)] = killErr
	}

	return successfulEnclaveIds, erroredEnclaveIds, nil
}

// ====================================================================================================
//                                      Private Helper Functions
// ====================================================================================================
// Gets API containers matching the search filters, indexed by their container ID
func (backend *DockerKurtosisBackend) getMatchingApiContainers(ctx context.Context, filters *api_container.APIContainerFilters) (map[string]*api_container.APIContainer, error) {

	apiContainerSearchLabels := map[string]string{
		label_key_consts.AppIDLabelKey.GetString():         label_value_consts.AppIDLabelValue.GetString(),
		label_key_consts.ContainerTypeLabelKey.GetString(): label_value_consts.APIContainerContainerTypeLabelValue.GetString(),
		// NOTE: we do NOT use the enclave ID label here, and instead do postfiltering, because Docker has no way to do disjunctive search!
	}
	allApiContainers, err := backend.dockerManager.GetContainersByLabels(ctx, apiContainerSearchLabels, shouldFetchAllContainersWhenRetrievingContainers)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred fetching API containers using labels: %+v", apiContainerSearchLabels)
	}

	allMatchingApiContainers := map[string]*api_container.APIContainer{}
	for _, apiContainer := range allApiContainers {
		containerId := apiContainer.GetId()
		apicObj, err := getApiContainerObjectFromContainerInfo(
			containerId,
			apiContainer.GetLabels(),
			apiContainer.GetStatus(),
			apiContainer.GetHostPortBindings(),
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
) (*api_container.APIContainer, error) {
	enclaveId, found := labels[label_key_consts.EnclaveIDLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected the API container's enclave ID to be found under label '%v' but the label wasn't present", label_key_consts.EnclaveIDLabelKey.GetString())
	}

	privateIpAddrStr, found := labels[label_key_consts.PrivateIPLabelKey.GetString()]
	if !found {
		// TODO DELETE THIS AFTER 2022-05-28 WHEN NO API CONTAINERS WON'T HAVE A PRIVATE IP
		candidateIpAddrStr, found := labels[pre_2022_03_28_IpAddrLabel]
		if !found {
			return nil, stacktrace.NewError(
				"Couldn't find the API container's private IP using label '%v' nor '%v'",
				label_key_consts.PrivateIPLabelKey.GetString(),
				pre_2022_03_28_IpAddrLabel,
			)
		}
		privateIpAddrStr = candidateIpAddrStr
	}
	privateIpAddr := net.ParseIP(privateIpAddrStr)
	if privateIpAddr == nil {
		return nil, stacktrace.NewError("Couldn't parse private IP address string '%v' to an IP", privateIpAddrStr)
	}

	privateGrpcPortSpec, privateGrpcProxyPortSpec, err := getPrivateApiContainerPorts(labels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the API container's private port specs from container '%v' with labels: %+v", containerId, labels)
	}

	isContainerRunning, found := isContainerRunningDeterminer[containerStatus]
	if !found {
		// This should never happen because we enforce completeness in a unit test
		return nil, stacktrace.NewError("No is-running designation found for API container status '%v'; this is a bug in Kurtosis!", containerStatus.String())
	}
	var apiContainerStatus container_status.ContainerStatus
	if isContainerRunning {
		apiContainerStatus = container_status.ContainerStatus_Running
	} else {
		apiContainerStatus = container_status.ContainerStatus_Stopped
	}

	var publicIpAddr net.IP
	var publicGrpcPortSpec *port_spec.PortSpec
	var publicGrpcProxyPortSpec *port_spec.PortSpec
	if apiContainerStatus == container_status.ContainerStatus_Running {
		publicGrpcPortIpAddr, candidatePublicGrpcPortSpec, err := getPublicPortBindingFromPrivatePortSpec(privateGrpcPortSpec, allHostMachinePortBindings)
		if err != nil {
			return nil, stacktrace.Propagate(err, "The engine is running, but an error occurred getting the public port spec for the engine's grpc private port spec")
		}
		publicGrpcPortSpec = candidatePublicGrpcPortSpec

		publicGrpcProxyPortIpAddr, candidatePublicGrpcProxyPortSpec, err := getPublicPortBindingFromPrivatePortSpec(privateGrpcProxyPortSpec, allHostMachinePortBindings)
		if err != nil {
			return nil, stacktrace.Propagate(err, "The engine is running, but an error occurred getting the public port spec for the engine's grpc private port spec")
		}
		publicGrpcProxyPortSpec = candidatePublicGrpcProxyPortSpec

		if publicGrpcPortIpAddr.String() != publicGrpcProxyPortIpAddr.String() {
			return nil, stacktrace.NewError(
				"Expected the engine's grpc port public IP address '%v' and grpc-proxy port public IP address '%v' to be the same, but they were different",
				publicGrpcPortIpAddr.String(),
				publicGrpcProxyPortIpAddr.String(),
			)
		}
		publicIpAddr = publicGrpcPortIpAddr
	}

	result := api_container.NewAPIContainer(
		enclave.EnclaveID(enclaveId),
		apiContainerStatus,
		privateIpAddr,
		privateGrpcPortSpec,
		privateGrpcProxyPortSpec,
		publicIpAddr,
		publicGrpcPortSpec,
		publicGrpcProxyPortSpec,
	)

	return result, nil
}

func getPrivateApiContainerPorts(containerLabels map[string]string) (
	resultGrpcPortSpec *port_spec.PortSpec,
	resultGrpcProxyPortSpec *port_spec.PortSpec,
	resultErr error,
) {
	serializedPortSpecs, found := containerLabels[label_key_consts.PortSpecsLabelKey.GetString()]
	if !found {
		return nil, nil, stacktrace.NewError("Expected to find port specs label '%v' but none was found", label_key_consts.PortSpecsLabelKey.GetString())
	}

	portSpecs, err := port_spec_serializer.DeserializePortSpecs(serializedPortSpecs)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred deserializing port specs string '%v'", serializedPortSpecs)
	}

	grpcPortSpec, foundGrpcPort := portSpecs[kurtosisInternalContainerGrpcPortId]
	if !foundGrpcPort {
		return nil, nil, stacktrace.NewError("No grpc port with ID '%v' found in the port specs", kurtosisInternalContainerGrpcPortId)
	}

	grpcProxyPortSpec, foundGrpcProxyPort := portSpecs[kurtosisInternalContainerGrpcProxyPortId]
	if !foundGrpcProxyPort {
		return nil, nil, stacktrace.NewError("No grpc-proxy port with ID '%v' found in the port specs", kurtosisInternalContainerGrpcProxyPortId)
	}

	return grpcPortSpec, grpcProxyPortSpec, nil
}

func extractEnclaveIdFromUncastedApiContainerObj(uncastedApiContainerObj interface{}) (string, error) {
	castedObj, ok := uncastedApiContainerObj.(*api_container.APIContainer)
	if !ok {
		return "", stacktrace.NewError("An error occurred downcasting the API container object")
	}
	return string(castedObj.GetEnclaveID()), nil
}
