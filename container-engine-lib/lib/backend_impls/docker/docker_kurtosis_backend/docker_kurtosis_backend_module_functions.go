package docker_kurtosis_backend

import (
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_log_streaming_readcloser"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_operation_parallelizer"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_port_spec_serializer"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/module"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"time"
)

const (
	// The module container uses gRPC so MUST listen on TCP (no other protocols are supported)
	moduleContainerPortProtocol = port_spec.PortProtocol_TCP

	// The location where the enclave data volume will be mounted
	//  on the module container
	enclaveDataVolumeDirpathOnModuleContainer = "/kurtosis-data"

	maxWaitForModuleContainerAvailabilityRetries         = 10
	timeBetweenWaitForModuleContainerAvailabilityRetries = 1 * time.Second
)

// TODO: MIGRATE THIS FOLDER TO USE STRUCTURE OF USER_SERVICE_FUNCTIONS MODULE

func (backend *DockerKurtosisBackend) CreateModule(
	ctx context.Context,
	image string,
	enclaveId enclave.EnclaveID,
	id module.ModuleID,
	grpcPortNum uint16,
	envVars map[string]string,
) (
	newModule *module.Module,
	resultErr error,
) {

	uuidStr, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred generating a UUID string for module with ID '%v'", id)
	}
	guid := module.ModuleGUID(string(id) + "-" + uuidStr)

	freeIpAddrProvider, found := backend.enclaveFreeIpProviders[enclaveId]
	if !found {
		return nil, stacktrace.NewError(
			"Received a request to create module with ID '%v' in enclave '%v', but no free IP address provider was " +
				"defined for this enclave; this likely means that the request is being called where it shouldn't " +
				"be (i.e. outside the API container)",
			id,
			enclaveId,
		)
	}

	// Get the Docker network ID where we'll start the new module
	enclaveNetwork, err := backend.getEnclaveNetworkByEnclaveId(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave network by enclave ID '%v'", enclaveId)
	}

	enclaveDataVolumeName, err := backend.getEnclaveDataVolumeByEnclaveId(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the enclave data volume for enclave '%v'", enclaveId)
	}

	privateGrpcPortSpec, err := port_spec.NewPortSpec(grpcPortNum, moduleContainerPortProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the module container's private grpc port spec object using number '%v' and protocol '%v'",
			grpcPortNum,
			consts.EnginePortProtocol.String(),
		)
	}

	enclaveObjAttrProvider, err := backend.objAttrsProvider.ForEnclave(enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Couldn't get an object attribute provider for enclave '%v'", enclaveId)
	}

	ipAddr, err := freeIpAddrProvider.GetFreeIpAddr()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting a free IP address")
	}
	shouldReleaseIp := true
	defer func() {
		if shouldReleaseIp {
			freeIpAddrProvider.ReleaseIpAddr(ipAddr)
		}
	}()

	moduleContainerAttrs, err := enclaveObjAttrProvider.ForModuleContainer(
		ipAddr,
		id,
		guid,
		consts.KurtosisInternalContainerGrpcPortId,
		privateGrpcPortSpec,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the object attributes for the module container")
	}

	privateGrpcDockerPort, err := shared_helpers.TransformPortSpecToDockerPort(privateGrpcPortSpec)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred transforming the private grpc port spec to a Docker port")
	}
	usedPorts := map[nat.Port]docker_manager.PortPublishSpec{
		privateGrpcDockerPort: docker_manager.NewAutomaticPublishingSpec(),
	}

	volumeMounts := map[string]string{
		enclaveDataVolumeName: enclaveDataVolumeDirpathOnModuleContainer,
	}

	labelStrs := map[string]string{}
	for labelKey, labelValue := range moduleContainerAttrs.GetLabels() {
		labelStrs[labelKey.GetString()] = labelValue.GetString()
	}

	// Best-effort pull attempt
	if err = backend.dockerManager.PullImage(ctx, image); err != nil {
		logrus.Warnf("Failed to pull the latest version of module container image '%v'; you may be running an out-of-date version", image)
	}

	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		image,
		moduleContainerAttrs.GetName().GetString(),
		enclaveNetwork.GetId(),
	).WithEnvironmentVariables(
		envVars,
	).WithVolumeMounts(
		volumeMounts,
	).WithStaticIP(
		ipAddr,
	).WithUsedPorts(
		usedPorts,
	).WithLabels(
		labelStrs,
	).Build()

	containerId, hostMachinePortBindings, err := backend.dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred starting the module container")
	}
	shouldKillContainer := true
	defer func() {
		if shouldKillContainer {
			// NOTE: We use the background context here so that the kill will still go off even if the reason for
			// the failure was the original context being cancelled
			if err := backend.dockerManager.KillContainer(context.Background(), containerId); err != nil {
				logrus.Errorf(
					"Launching module container '%v' with container ID '%v' didn't complete successfully so we "+
						"tried to kill the container we started, but doing so exited with an error:\n%v",
					moduleContainerAttrs.GetName(),
					containerId,
					err,
				)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually stop module container with ID '%v'!!!!!!", containerId)
			}
		}
	}()

	if err := shared_helpers.WaitForPortAvailabilityUsingNetstat(
		ctx,
		backend.dockerManager,
		containerId,
		privateGrpcPortSpec,
		maxWaitForModuleContainerAvailabilityRetries,
		timeBetweenWaitForModuleContainerAvailabilityRetries,
	); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the module container's grpc port to become available")
	}

	result, err := getModuleObjectFromContainerInfo(containerId, labelStrs, types.ContainerStatus_Running, hostMachinePortBindings)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a module container object from container with ID '%v'", containerId)
	}

	shouldKillContainer = false
	shouldReleaseIp = false
	return result, nil
}

func (backend *DockerKurtosisBackend) GetModules(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *module.ModuleFilters,
) (
	map[module.ModuleGUID]*module.Module,
	error,
) {
	matchingModuleContainers, err := backend.getMatchingModules(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting module containers matching the following filters: %+v", filters)
	}

	matchingModuleContainersByModuleID := map[module.ModuleGUID]*module.Module{}
	for _, moduleObj := range matchingModuleContainers {
		matchingModuleContainersByModuleID[moduleObj.GetGUID()] = moduleObj
	}

	return matchingModuleContainersByModuleID, nil
}

func (backend *DockerKurtosisBackend) GetModuleLogs(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *module.ModuleFilters,
	shouldFollowLogs bool,
) (
	map[module.ModuleGUID]io.ReadCloser,
	map[module.ModuleGUID]error,
	error,
) {
	matchingModulesByContainerId, err := backend.getMatchingModules(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting modules matching filters '%+v'", filters)
	}

	successfulModuleLogs := map[module.ModuleGUID]io.ReadCloser{}
	erroredModules := map[module.ModuleGUID]error{}

	//TODO use concurrency to improve perf
	shouldCloseLogStreams := true
	for containerId, module := range matchingModulesByContainerId {
		rawLogStream, err := backend.dockerManager.GetContainerLogs(ctx, containerId, shouldFollowLogs)
		if err != nil {
			serviceError := stacktrace.Propagate(err, "An error occurred getting logs for module with GUID '%v' and container ID '%v'", module.GetGUID(), containerId)
			erroredModules[module.GetGUID()] = serviceError
			continue
		}
		defer func() {
			if shouldCloseLogStreams {
				rawLogStream.Close()
			}
		}()

		demultiplexedStream := docker_log_streaming_readcloser.NewDockerLogStreamingReadCloser(rawLogStream)
		defer func() {
			if shouldCloseLogStreams {
				rawLogStream.Close()
			}
		}()

		successfulModuleLogs[module.GetGUID()] = demultiplexedStream
	}

	shouldCloseLogStreams = false
	return successfulModuleLogs, erroredModules, nil
}

func (backend *DockerKurtosisBackend) StopModules(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *module.ModuleFilters,
) (
	resultSuccessfulModuleGuids map[module.ModuleGUID]bool,
	resultErroredModuleGuids map[module.ModuleGUID]error,
	resultErr error,
) {
	matchingModulesByContainerId, err := backend.getMatchingModules(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting modules matching filters '%+v'", filters)
	}

	// TODO PLEAAASE GO GENERICS... but we can't use 1.18 yet because it'll break all Kurtosis clients :(
	matchingUncastedObjectsByContainerId := map[string]interface{}{}
	for containerId, object := range matchingModulesByContainerId {
		matchingUncastedObjectsByContainerId[containerId] = interface{}(object)
	}

	var killOperation docker_operation_parallelizer.DockerOperation = func(
		ctx context.Context,
		dockerManager *docker_manager.DockerManager,
		dockerObjectId string,
	) error {
		if err := dockerManager.KillContainer(ctx, dockerObjectId); err != nil {
			return stacktrace.Propagate(err, "An error occurred killing module container with ID '%v'", dockerObjectId)
		}
		return nil
	}

	successfulGuidStrs, erroredGuidStrs, err := docker_operation_parallelizer.RunDockerOperationInParallelForKurtosisObjects(
		ctx,
		matchingUncastedObjectsByContainerId,
		backend.dockerManager,
		extractModuleGuidFromObj,
		killOperation,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred killing modules containers matching filters '%+v'", filters)
	}

	successfulGuids := map[module.ModuleGUID]bool{}
	for enclaveIdStr := range successfulGuidStrs {
		successfulGuids[module.ModuleGUID(enclaveIdStr)] = true
	}
	erroredGuids := map[module.ModuleGUID]error{}
	for enclaveIdStr, killErr := range erroredGuidStrs {
		erroredGuids[module.ModuleGUID(enclaveIdStr)] = killErr
	}

	return successfulGuids, erroredGuids, nil
}

func (backend *DockerKurtosisBackend) DestroyModules(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *module.ModuleFilters,
) (
	successfulModuleIds map[module.ModuleGUID]bool,
	erroredModuleIds map[module.ModuleGUID]error,
	resultErr error,
) {
	matchingModulesByContainerId, err := backend.getMatchingModules(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting module containers matching the following filters: %+v", filters)
	}

	// TODO PLEAAASE GO GENERICS... but we can't use 1.18 yet because it'll break all Kurtosis clients :(
	matchingUncastedObjectsByContainerId := map[string]interface{}{}
	for containerId, object := range matchingModulesByContainerId {
		matchingUncastedObjectsByContainerId[containerId] = interface{}(object)
	}

	var dockerOperation docker_operation_parallelizer.DockerOperation = func(
		ctx context.Context,
		dockerManager *docker_manager.DockerManager,
		dockerObjectId string,
	) error {
		if err := dockerManager.RemoveContainer(ctx, dockerObjectId); err != nil {
			return stacktrace.Propagate(err, "An error occurred removing module container with ID '%v'", dockerObjectId)
		}
		return nil
	}

	successfulGuidStrs, erroredGuidStrs, err := docker_operation_parallelizer.RunDockerOperationInParallelForKurtosisObjects(
		ctx,
		matchingUncastedObjectsByContainerId,
		backend.dockerManager,
		extractModuleGuidFromObj,
		dockerOperation,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred removing modules containers matching filters '%+v'", filters)
	}

	successfulGuids := map[module.ModuleGUID]bool{}
	for guidStr := range successfulGuidStrs {
		successfulGuids[module.ModuleGUID(guidStr)] = true
	}
	erroredGuids := map[module.ModuleGUID]error{}
	for guidStr, removalErr := range erroredGuidStrs {
		erroredGuids[module.ModuleGUID(guidStr)] = removalErr
	}

	return successfulGuids, erroredGuids, nil
}

// ====================================================================================================
//                                     Private Helper Methods
// ====================================================================================================
// Gets modules matching the search filters, indexed by their container ID
func (backend *DockerKurtosisBackend) getMatchingModules(ctx context.Context, filters *module.ModuleFilters) (map[string]*module.Module, error) {

	moduleContainerSearchLabels := map[string]string{
		label_key_consts.AppIDDockerLabelKey.GetString():         label_value_consts.AppIDDockerLabelValue.GetString(),
		label_key_consts.ContainerTypeDockerLabelKey.GetString(): label_value_consts.ModuleContainerTypeDockerLabelValue.GetString(),
	}
	matchingModuleContainers, err := backend.dockerManager.GetContainersByLabels(ctx, moduleContainerSearchLabels, consts.ShouldFetchAllContainersWhenRetrievingContainers)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred fetching module containers using labels: %+v", moduleContainerSearchLabels)
	}

	matchingModuleObjects := map[string]*module.Module{}
	for _, moduleContainer := range matchingModuleContainers {
		containerId := moduleContainer.GetId()
		moduleObj, err := getModuleObjectFromContainerInfo(
			containerId,
			moduleContainer.GetLabels(),
			moduleContainer.GetStatus(),
			moduleContainer.GetHostPortBindings(),
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred converting container with ID '%v' into a module object", moduleContainer.GetId())
		}

		// If the ID filter is specified, drop modules not matching it
		if filters.GUIDs != nil && len(filters.GUIDs) > 0 {
			if _, found := filters.GUIDs[moduleObj.GetGUID()]; !found {
				continue
			}
		}

		// If status filter is specified, drop modules	 not matching it
		if filters.Statuses != nil && len(filters.Statuses) > 0 {
			if _, found := filters.Statuses[moduleObj.GetStatus()]; !found {
				continue
			}
		}

		matchingModuleObjects[containerId] = moduleObj
	}

	return matchingModuleObjects, nil
}

func getModuleObjectFromContainerInfo(
	containerId string,
	labels map[string]string,
	containerStatus types.ContainerStatus,
	allHostMachinePortBindings map[nat.Port]*nat.PortBinding,
) (*module.Module, error) {
	enclaveId, found := labels[label_key_consts.EnclaveIDDockerLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected the module's enclave ID to be found under label '%v' but the label wasn't present", label_key_consts.EnclaveIDDockerLabelKey.GetString())
	}

	id, found := labels[label_key_consts.IDDockerLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected to find module ID label key '%v' but none was found", label_key_consts.IDDockerLabelKey.GetString())
	}

	guid, found := labels[label_key_consts.GUIDDockerLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected to find module GUID label key '%v' but none was found", label_key_consts.GUIDDockerLabelKey.GetString())
	}

	var privateIpAddr net.IP
	privateIpAddrStr, found := labels[label_key_consts.PrivateIPDockerLabelKey.GetString()]
	// UNCOMMENT THIS AFTER 2022-06-30 WHEN NOBODY HAS MODULES WITHOUT THE PRIVATE IP ADDRESS LABEL
	/*
		if !found {
			return nil, stacktrace.NewError("Expected to find module private IP label key '%v' but none was found", label_key_consts.PrivateIPDockerLabelKey.GetString())
		}
	*/
	if found {
		candidatePrivateIpAddr := net.ParseIP(privateIpAddrStr)
		if candidatePrivateIpAddr == nil {
			return nil, stacktrace.NewError("Couldn't parse private IP address string '%v' to an IP", privateIpAddrStr)
		}
		privateIpAddr = candidatePrivateIpAddr
	}

	privateGrpcPortSpec, err := getPrivateModulePorts(labels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the module container's private port specs from container '%v' with labels: %+v", containerId, labels)
	}

	isContainerRunning, found := shared_helpers.IsContainerRunningDeterminer[containerStatus]
	if !found {
		// This should never happen because we enforce completeness in a unit test
		return nil, stacktrace.NewError("No is-running designation found for module container status '%v'; this is a bug in Kurtosis!", containerStatus.String())
	}
	var moduleStatus container_status.ContainerStatus
	if isContainerRunning {
		moduleStatus = container_status.ContainerStatus_Running
	} else {
		moduleStatus = container_status.ContainerStatus_Stopped
	}

	var publicIpAddr net.IP
	var publicGrpcPortSpec *port_spec.PortSpec
	if moduleStatus == container_status.ContainerStatus_Running {
		publicGrpcPortIpAddr, candidatePublicGrpcPortSpec, err := shared_helpers.GetPublicPortBindingFromPrivatePortSpec(privateGrpcPortSpec, allHostMachinePortBindings)
		if err != nil {
			return nil, stacktrace.Propagate(err, "The module is running, but an error occurred getting the public port spec for the module's grpc private port spec")
		}
		publicGrpcPortSpec = candidatePublicGrpcPortSpec
		publicIpAddr = publicGrpcPortIpAddr
	}

	result := module.NewModule(
		enclave.EnclaveID(enclaveId),
		module.ModuleID(id),
		module.ModuleGUID(guid),
		moduleStatus,
		privateIpAddr,
		privateGrpcPortSpec,
		publicIpAddr,
		publicGrpcPortSpec,
	)

	return result, nil
}

func getPrivateModulePorts(containerLabels map[string]string) (
	resultGrpcPortSpec *port_spec.PortSpec,
	resultErr error,
) {
	serializedPortSpecs, found := containerLabels[label_key_consts.PortSpecsDockerLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected to find port specs label '%v' but none was found", label_key_consts.PortSpecsDockerLabelKey.GetString())
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

func extractModuleGuidFromObj(uncastedModuleObj interface{}) (string, error) {
	castedObj, ok := uncastedModuleObj.(*module.Module)
	if !ok {
		return "", stacktrace.NewError("An error occurred downcasting the module object")
	}
	return string(castedObj.GetGUID()), nil
}
