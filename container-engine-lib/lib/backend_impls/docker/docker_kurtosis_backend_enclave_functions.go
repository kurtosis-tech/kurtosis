package docker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	docker_types "github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/gammazero/workerpool"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_task_parallelizer"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

const (
	shouldFetchStoppedContainersWhenGettingEnclaveStatus = true

	containerInspectResultFilename = "spec.json"
	containerLogsFilename          = "output.log"

	shouldFetchStoppedContainersWhenDumpingEnclave = true
	numContainersToDumpAtOnce                      = 20

	// Permissions for the files & directories we create as a result of the dump
	createdDirPerms  = 0755
	createdFilePerms = 0644

	shouldFollowContainerLogsWhenDumping = false

	containerSpecJsonSerializationIndent = "  "
	containerSpecJsonSerializationPrefix = ""
)

type matchingNetworkInformation struct {
	enclaveId enclave.EnclaveID
	enclaveStatus enclave.EnclaveStatus
	dockerNetwork *types.Network
	containers []*types.Container
}

func (backend *DockerKurtosisBackend) CreateEnclave(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	isPartitioningEnabled bool,
) (
	*enclave.Enclave,
	error,
) {
	teardownCtx := context.Background() // Separate context for tearing stuff down in case the input context is cancelled

	searchNetworkLabels := map[string]string{
		label_key_consts.AppIDLabelKey.GetString(): label_value_consts.AppIDLabelValue.GetString(),
		label_key_consts.EnclaveIDLabelKey.GetString(): string(enclaveId),
	}

	networks, err := backend.dockerManager.GetNetworksByLabels(ctx, searchNetworkLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting networks using labels '%+v', which is necessary to ensure that our enclave doesn't exist yet", searchNetworkLabels)
	}
	if len(networks) > 0 {
		return nil, stacktrace.NewError("Cannot create enclave with ID '%v' because an enclave with ID '%v' already exists", enclaveId, enclaveId)
	}

	volumeSearchLabels :=  map[string]string{
		label_key_consts.AppIDLabelKey.GetString(): label_value_consts.AppIDLabelValue.GetString(),
		label_key_consts.EnclaveIDLabelKey.GetString(): string(enclaveId),
		label_key_consts.VolumeTypeLabelKey.GetString(): label_value_consts.EnclaveDataVolumeTypeLabelValue.GetString(),
	}
	foundVolumes, err := backend.dockerManager.GetVolumesByLabels(ctx, volumeSearchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave data volumes matching labels '%+v'", volumeSearchLabels)
	}
	if len(foundVolumes) > 0 {
		return nil, stacktrace.NewError("Cannot create enclave with ID '%v' because one or more enclave data volume for that enclave already exists", enclaveId)
	}

	enclaveObjAttrsProvider, err := backend.objAttrsProvider.ForEnclave(enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while trying to generate an object attributes provider for the enclave with ID '%v'", enclaveId)
	}

	enclaveNetworkAttrs, err := enclaveObjAttrsProvider.ForEnclaveNetwork(isPartitioningEnabled)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while trying to get the enclave network attributes for the enclave with ID '%v'", enclaveId)
	}

	enclaveDataVolumeAttrs, err := enclaveObjAttrsProvider.ForEnclaveDataVolume()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while trying to get the enclave data volume attributes for the enclave with ID '%v'", enclaveId)
	}

	enclaveNetworkName := enclaveNetworkAttrs.GetName()
	enclaveNetworkDockerObjectLabels := enclaveNetworkAttrs.GetLabels()

	enclaveNetworkLabels := map[string]string{}
	for dockerLabelKey, dockerLabelValue := range enclaveNetworkDockerObjectLabels {
		enclaveNetworkLabelKey := dockerLabelKey.GetString()
		enclaveNetworkLabelValue := dockerLabelValue.GetString()
		enclaveNetworkLabels[enclaveNetworkLabelKey] = enclaveNetworkLabelValue
	}

	logrus.Debugf("Creating Docker network for enclave '%v'...", enclaveId)
	networkId, networkIpAndMask, gatewayIp, freeIpAddrTracker, err := backend.dockerNetworkAllocator.CreateNewNetwork(
		ctx,
		enclaveNetworkName.GetString(),
		enclaveNetworkLabels,
	)
	if err != nil {
		// TODO If the user Ctrl-C's while the CreateNetwork call is ongoing then the CreateNetwork will error saying
		//  that the Context was cancelled as expected, but *the Docker engine will still create the network*!!! We'll
		//  need to parse the log message for the string "context canceled" and, if found, do another search for
		//  networks with our network name and delete them
		return nil, stacktrace.Propagate(err, "An error occurred allocating a new network for enclave '%v'", enclaveId)
	}
	shouldDeleteNetwork := true
	defer func() {
		if shouldDeleteNetwork {
			if err := backend.dockerManager.RemoveNetwork(teardownCtx, networkId); err != nil {
				logrus.Errorf("Creating the enclave didn't complete successfully, so we tried to delete network '%v' that we created but an error was thrown:\n%v", networkId, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove network with ID '%v'!!!!!!!", networkId)
			}
		}
	}()
	logrus.Debugf("Docker network '%v' created successfully with ID '%v' and subnet CIDR '%v'", enclaveId, networkId, networkIpAndMask.String())

	enclaveDataVolumeNameStr := enclaveDataVolumeAttrs.GetName().GetString()
	enclaveDataVolumeLabelStrs := map[string]string{}
	for labelKey, labelValue := range enclaveDataVolumeAttrs.GetLabels() {
		enclaveDataVolumeLabelStrs[labelKey.GetString()] = labelValue.GetString()
	}
	if err := backend.dockerManager.CreateVolume(ctx, enclaveDataVolumeNameStr, enclaveDataVolumeLabelStrs); err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating enclave data volume with name '%v' and labels '%+v'",
			enclaveDataVolumeNameStr,
			enclaveDataVolumeLabelStrs,
		)
	}
	shouldDeleteVolume := true
	defer func() {
		if shouldDeleteVolume {
			if err := backend.dockerManager.RemoveVolume(teardownCtx, enclaveDataVolumeNameStr); err != nil {
				logrus.Errorf(
					"Creating the enclave didn't complete successfully, so we tried to delete enclave data volume '%v' " +
						"that we created but an error was thrown:\n%v",
					enclaveDataVolumeNameStr,
					err,
				)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove volume with name '%v'!!!!!!!", enclaveDataVolumeNameStr)
			}
		}
	}()

	newEnclave := enclave.NewEnclave(enclaveId, enclave.EnclaveStatus_Empty, networkId, networkIpAndMask.String(), gatewayIp, freeIpAddrTracker)

	shouldDeleteNetwork = false
	shouldDeleteVolume = false
	return newEnclave, nil
}

// Gets enclaves matching the given filters
func (backend *DockerKurtosisBackend) GetEnclaves(
	ctx context.Context,
	filters *enclave.EnclaveFilters,
) (
	map[enclave.EnclaveID]*enclave.Enclave,
	error,
) {

	allMatchingNetworkInfo, err := backend.getMatchingEnclaveNetworkInfo(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave networks matching filters '%+v'", filters)
	}

	result := map[enclave.EnclaveID]*enclave.Enclave{}
	for enclaveId, matchingNetworkInfo := range allMatchingNetworkInfo {
		result[enclaveId] = enclave.NewEnclave(
			enclaveId,
			matchingNetworkInfo.enclaveStatus,
			matchingNetworkInfo.dockerNetwork.GetId(),
			matchingNetworkInfo.dockerNetwork.GetIpAndMask().String(),
			// TODO We're returning nil here for gatewayIp and freeIpAddrProvider as a temporary hack, until we can fully push all Docker stuff under the KurtosisBackend
			nil,
			nil,
		)
	}

	return result, nil
}

// Stops enclaves matching the given filters
func (backend *DockerKurtosisBackend) StopEnclaves(
	ctx context.Context,
	filters *enclave.EnclaveFilters,
) (
	resultSuccessfulEnclaveIds map[enclave.EnclaveID]bool,
	resultErroredEnclaveIds map[enclave.EnclaveID]error,
	resultErr error,
) {

	matchingNetworkInfo, err := backend.getMatchingEnclaveNetworkInfo(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting enclave network info using filters '%+v'", filters)
	}

	// For all the enclaves to stop, gather all the containers that should be stopped
	enclaveIdsForContainerIdsToStop := map[string]enclave.EnclaveID{}
	containerIdsToStopToUncastedContainerId := map[string]interface{}{}
	for enclaveId, networkInfo := range matchingNetworkInfo {
		for _, container := range networkInfo.containers {
			containerId := container.GetId()
			enclaveIdsForContainerIdsToStop[containerId] = enclaveId
			containerIdsToStopToUncastedContainerId[containerId] = interface{}(containerId)
		}
	}

	var stopEnclaveContainerOperation docker_task_parallelizer.DockerOperation = func(ctx context.Context, dockerManager *docker_manager.DockerManager, dockerObjectId string) error {
		if err := dockerManager.KillContainer(ctx, dockerObjectId); err != nil {
			return stacktrace.Propagate(err, "An error occurred killing enclave container with ID '%v'", dockerObjectId)
		}
		return nil
	}

	_, erroredContainerIds, err := docker_task_parallelizer.RunDockerOperationInParallelForKurtosisObject(
		ctx,
		containerIdsToStopToUncastedContainerId,
		backend.dockerManager,
		func(uncastedContainerId interface{}) (string, error) {
			containerIdStr, ok := uncastedContainerId.(string)
			if !ok {
				return "", stacktrace.NewError("Failed to cast uncasted container ID to a casted string container ID")
			}
			return containerIdStr, nil
		},
		stopEnclaveContainerOperation,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred stopping the containers of enclaves matching filters '%+v'", filters)
	}

	// Do we need to explicitly wait until the containers exit?

	containerKillErrorStrsByEnclave := map[enclave.EnclaveID][]string{}
	for erroredContainerId, killContainerErr := range erroredContainerIds {
		containerEnclaveId, found := enclaveIdsForContainerIdsToStop[erroredContainerId]
		if !found {
			return nil, nil, stacktrace.NewError("An error occurred stopping container '%v' in an enclave we didn't request", erroredContainerId)
		}

		existingEnclaveErrors, found := containerKillErrorStrsByEnclave[containerEnclaveId]
		if !found {
			existingEnclaveErrors = []string{}
		}
		containerKillErrorStrsByEnclave[containerEnclaveId] = append(existingEnclaveErrors, killContainerErr.Error())
	}


	successfulEnclaveIds := map[enclave.EnclaveID]bool{}
	erroredEnclaveIds := map[enclave.EnclaveID]error{}
	for enclaveId, containerKillErrorStrs := range containerKillErrorStrsByEnclave {
		if len(containerKillErrorStrs) == 0 {
			successfulEnclaveIds[enclaveId] = true
			continue
		}
		errorStr := strings.Join(containerKillErrorStrs, "\n\n")
		erroredEnclaveIds[enclaveId] = stacktrace.NewError(
			"One or more errors occurred killing the containers in enclave '%v':\n%v",
			enclaveId,
			errorStr,
		)
		continue
	}
	return successfulEnclaveIds, erroredEnclaveIds, nil
}

func (backend *DockerKurtosisBackend) DumpEnclave(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	outputDirpath string,
) error {
	enclaveContainerSearchLabels := map[string]string{
		label_key_consts.AppIDLabelKey.GetString():     label_value_consts.AppIDLabelValue.GetString(),
		label_key_consts.EnclaveIDLabelKey.GetString(): string(enclaveId),
	}

	enclaveContainers, err := backend.dockerManager.GetContainersByLabels(ctx, enclaveContainerSearchLabels, shouldFetchStoppedContainersWhenDumpingEnclave)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred getting the containers in enclave '%v' for dumping the enclave",
			enclaveId,
		)
	}

	// Create output directory
	if _, err := os.Stat(outputDirpath); !os.IsNotExist(err) {
		return stacktrace.NewError("Cannot create output directory at '%v'; directory already exists", outputDirpath)
	}
	if err := os.Mkdir(outputDirpath, createdDirPerms); err != nil {
		return stacktrace.Propagate(err, "An error occurred creating output directory at '%v'", outputDirpath)
	}

	workerPool := workerpool.New(numContainersToDumpAtOnce)
	resultErrsChan := make(chan error, len(enclaveContainers))
	for _, container := range enclaveContainers {
		containerName := container.GetName()
		containerId := container.GetId()
		logrus.Debugf("Submitting job to dump info about container with name '%v' and ID '%v'", containerName, containerId)

		/*
			!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
			It's VERY important that the actual `func()` job function get created inside a helper function!!
			This is because variables declared inside for-loops are created BY REFERENCE rather than by-value, which
				means that if we inline the `func() {....}` creation here then all the job functions would get a REFERENCE to
				any variables they'd use.
			This means that by the time the job functions were run in the worker pool (long after the for-loop finished)
				then all the job functions would be using a reference from the last iteration of the for-loop.

			For more info, see the "Variables declared in for loops are passed by reference" section of:
				https://www.calhoun.io/gotchas-and-common-mistakes-with-closures-in-go/ for more details
			!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
		*/
		jobToSubmit := createDumpContainerJob(
			ctx,
			backend.dockerManager,
			outputDirpath,
			resultErrsChan,
			containerName,
			containerId,
		)

		workerPool.Submit(jobToSubmit)
	}
	workerPool.StopWait()
	close(resultErrsChan)

	allResultErrStrs := []string{}
	for resultErr := range resultErrsChan {
		allResultErrStrs = append(allResultErrStrs, resultErr.Error())
	}

	if len(allResultErrStrs) > 0 {
		allIndexedResultErrStrs := []string{}
		for idx, resultErrStr := range allResultErrStrs {
			indexedResultErrStr := fmt.Sprintf(">>>>>>>>>>>>>>>>> ERROR %v <<<<<<<<<<<<<<<<<\n%v", idx, resultErrStr)
			allIndexedResultErrStrs = append(allIndexedResultErrStrs, indexedResultErrStr)
		}

		// NOTE: We don't use stacktrace here because the actual stacktraces we care about are the ones from the threads!
		return errors.New(fmt.Sprintf(
			"The following errors occurred when trying to dump information about enclave '%v':\n%v",
			enclaveId,
			strings.Join(allIndexedResultErrStrs, "\n\n"),
		))
	}
	return nil
}

// Destroys enclaves matching the given filters
func (backend *DockerKurtosisBackend) DestroyEnclaves(
	ctx context.Context,
	filters *enclave.EnclaveFilters,
) (
	resultSuccessfulEnclaveIds map[enclave.EnclaveID]bool,
	resultErroredEnclaveIds map[enclave.EnclaveID]error,
	resultErr error,
) {
	matchingNetworkInfo, err := backend.getMatchingEnclaveNetworkInfo(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting enclave network info using filters '%+v'", filters)
	}

	erroredEnclaveIds := map[enclave.EnclaveID]error{}

	successfulContainerRemovalEnclaveIds, erroredContainerRemovalEnclaveIds, err := destroyContainersInEnclaves(ctx, backend.dockerManager, matchingNetworkInfo)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred destroying containers in enclaves matching filters '%+v'", filters)
	}
	for enclaveId, containerRemovalErr := range erroredContainerRemovalEnclaveIds {
		erroredEnclaveIds[enclaveId] = containerRemovalErr
	}

	successfulVolumeRemovalEnclaveIds, erroredVolumeRemovalEnclaveIds, err := destroyVolumesInEnclaves(ctx, backend.dockerManager, successfulContainerRemovalEnclaveIds)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred destroying containers in enclaves for which containers were successfully destroyed: %+v", successfulContainerRemovalEnclaveIds)
	}
	for enclaveId, volumeRemovalErr := range erroredVolumeRemovalEnclaveIds {
		erroredEnclaveIds[enclaveId] = volumeRemovalErr
	}

	// Remove the networks
	networkIdsToRemove


	return successfulErroredEnclaveIds, erroredEnclaveIds, nil
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================
func (backend *DockerKurtosisBackend) getMatchingEnclaveNetworkInfo(
	ctx context.Context,
	filters *enclave.EnclaveFilters,
) (
	// Keyed by network ID
	map[enclave.EnclaveID]*matchingNetworkInformation,
	error,
) {
	kurtosisNetworkLabels := map[string]string{
		label_key_consts.AppIDLabelKey.GetString(): label_value_consts.AppIDLabelValue.GetString(),
		// NOTE: we don't search by enclave ID here because Docker has no way to do disjunctive search
	}

	allKurtosisNetworks, err := backend.dockerManager.GetNetworksByLabels(ctx, kurtosisNetworkLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kurtosis networks")
	}

	// First, filter by enclave IDs
	matchingKurtosisEnclaveIdsByNetworkId := map[enclave.EnclaveID]*types.Network{}
	for _, kurtosisNetwork := range allKurtosisNetworks {
		enclaveId, err := getEnclaveIdFromNetwork(kurtosisNetwork)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting enclave ID from network '%+v'; this is a bug in Kurtosis", kurtosisNetwork)
		}

		if filters.IDs != nil && len(filters.IDs) > 0 {
			if _, found := filters.IDs[enclaveId]; !found {
				continue
			}
		}

		matchingKurtosisEnclaveIdsByNetworkId[enclaveId] = kurtosisNetwork
	}

	// Next, filter by enclave status
	result := map[enclave.EnclaveID]*matchingNetworkInformation{}
	for enclaveId, kurtosisNetwork := range matchingKurtosisEnclaveIdsByNetworkId {
		status, containers, err := backend.getEnclaveStatusAndContainers(ctx, enclaveId)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting enclave status and containers from network for enclave '%v'", enclaveId)
		}

		if filters.Statuses != nil && len(filters.Statuses) > 0 {
			if _, found := filters.Statuses[status]; !found {
				continue
			}
		}

		result[enclaveId] = &matchingNetworkInformation{
			enclaveId:     enclaveId,
			enclaveStatus: status,
			dockerNetwork: kurtosisNetwork,
			containers:    containers,
		}
	}

	return result, nil
}



/*
func (backend *DockerKurtosisBackend) getEnclaveNetworksByEnclaveIds(ctx context.Context, enclaveIds map[enclave.EnclaveID]bool) ([]*types.Network, error) {
	enclaveNetworks := []*types.Network{}

	kurtosisNetworkLabels := map[string]string{
		label_key_consts.AppIDLabelKey.GetString(): label_value_consts.AppIDLabelValue.GetString(),
	}

	appNetworks, err := backend.dockerManager.GetNetworksByLabels(ctx, kurtosisNetworkLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kurtosis networks")
	}

	for _, network := range appNetworks {
		enclaveId, err := getEnclaveIdFromNetwork(network)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting enclave ID from network '%+v'; it should never happens it's a bug in Kurtosis", network)
		}
		numOfEnclaveIds := len(enclaveIds)
		if _, found := enclaveIds[enclaveId]; found || numOfEnclaveIds == 0 {
			enclaveNetworks = append(enclaveNetworks, network)
		}
	}

	return enclaveNetworks, nil
}

 */

func (backend *DockerKurtosisBackend) getEnclaveStatusAndContainers(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
) (

	enclave.EnclaveStatus,
	[]*types.Container,
	error,
) {
	allEnclaveContainers, err := backend.getAllEnclaveContainers(ctx, enclaveId)
	if err != nil {
		return 0, nil, stacktrace.Propagate(err, "An error occurred getting the containers for enclave '%v'", enclaveId)
	}
	if len(allEnclaveContainers) == 0 {
		return enclave.EnclaveStatus_Empty, nil, nil
	}

	resultEnclaveStatus := enclave.EnclaveStatus_Stopped
	for _, enclaveContainer := range allEnclaveContainers {
		containerStatus := enclaveContainer.GetStatus()
		isContainerRunning, found := isContainerRunningDeterminer[containerStatus]
		if !found {
			// This should never happen because we enforce completeness in a unit test
			return 0, nil, stacktrace.NewError("No is-running designation found for enclave container status '%v'; this is a bug in Kurtosis!", containerStatus.String())
		}
		if isContainerRunning {
			resultEnclaveStatus = enclave.EnclaveStatus_Running
		}
	}

	return resultEnclaveStatus, allEnclaveContainers, nil
}

func (backend *DockerKurtosisBackend) getAllEnclaveContainers(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
) ([]*types.Container, error) {

	containers := []*types.Container{}

	searchLabels := map[string]string{
		label_key_consts.AppIDLabelKey.GetString(): label_value_consts.AppIDLabelValue.GetString(),
		label_key_consts.EnclaveIDLabelKey.GetString(): string(enclaveId),
	}
	containers, err := backend.dockerManager.GetContainersByLabels(ctx, searchLabels, shouldFetchStoppedContainersWhenGettingEnclaveStatus)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the containers for enclave '%v' by labels '%+v'", enclaveId, searchLabels)
	}
	return containers, nil
}

func getAllEnclaveVolumes(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
	enclaveId enclave.EnclaveID,
) ([]*docker_types.Volume, error) {

	volumes := []*docker_types.Volume{}

	searchLabels := map[string]string{
		label_key_consts.AppIDLabelKey.GetString(): label_value_consts.AppIDLabelValue.GetString(),
		label_key_consts.EnclaveIDLabelKey.GetString(): string(enclaveId),
	}

	volumes, err := dockerManager.GetVolumesByLabels(ctx, searchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the volumes for enclave '%v' by labels '%+v'", enclaveId, searchLabels)
	}

	return volumes, nil
}

func createDumpContainerJob(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
	enclaveOutputDirpath string,
	resultErrsChan chan error,
	containerName string,
	containerId string,
) func() {
	return func() {
		if err := dumpContainerInfo(ctx, dockerManager, enclaveOutputDirpath, containerName, containerId); err != nil {
			resultErrsChan <- stacktrace.Propagate(
				err,
				"An error occurred dumping container info for container with name '%v' and ID '%v'",
				containerName,
				containerId,
			)
		}
	}
}

func dumpContainerInfo(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
	enclaveOutputDirpath string,
	containerName string,
	containerId string,
) error {
	// Make output directory
	containerOutputDirpath := path.Join(enclaveOutputDirpath, containerName)
	if err := os.Mkdir(containerOutputDirpath, createdDirPerms); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred creating directory '%v' to hold the output of container with name '%v' and ID '%v'",
			containerOutputDirpath,
			containerName,
			containerId,
		)
	}

	// Write container inspect results to file
	inspectResult, err := dockerManager.InspectContainer(
		ctx,
		containerId,
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred inspecting container with ID '%v'", containerId)
	}
	jsonSerializedInspectResultBytes, err := json.MarshalIndent(inspectResult, containerSpecJsonSerializationPrefix, containerSpecJsonSerializationIndent)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred serializing the results of inspecting container with ID '%v' to JSON", containerId)
	}
	specOutputFilepath := path.Join(containerOutputDirpath, containerInspectResultFilename)
	if err := ioutil.WriteFile(specOutputFilepath, jsonSerializedInspectResultBytes, createdFilePerms); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred writing the inspect output of container with name '%v' and ID '%v' to file '%v'",
			containerName,
			containerId,
			specOutputFilepath,
		)
	}

	// Write container logs to file
	containerLogsReadCloser, err := dockerManager.GetContainerLogs(ctx, containerId, shouldFollowContainerLogsWhenDumping)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the logs for container with ID '%v'", containerId)
	}
	defer containerLogsReadCloser.Close()
	logsOutputFilepath := path.Join(containerOutputDirpath, containerLogsFilename)
	logsOutputFp, err := os.Create(logsOutputFilepath)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred creating file '%v' to hold the logs of container with name '%v' and ID '%v'",
			logsOutputFilepath,
			containerName,
			containerId,
		)
	}

	// TODO Figure out a way to abstract this!!! This check-if-the-container-is-TTY-and-use-io.Copy-if-so-and-stdcopy-if-not
	//  is copied straight from the Docker CLI, but it REALLY sucks that a Kurtosis dev magically needs to know that that's what
	//  they have to do if they want to read container logs
	// If we don't have this, reading the logs from REPL container breaks
	if inspectResult.Config.Tty {
		if _, err := io.Copy(logsOutputFp, containerLogsReadCloser); err != nil {
			return stacktrace.Propagate(
				err,
				"An error occurred copying the TTY container logs stream to file '%v' for container with name '%v' and ID '%v'",
				logsOutputFilepath,
				containerName,
				containerId,
			)
		}
	} else {
		if _, err := stdcopy.StdCopy(logsOutputFp, logsOutputFp, containerLogsReadCloser); err != nil {
			return stacktrace.Propagate(
				err,
				"An error occurred copying the non-TTY container logs stream to file '%v' for container with name '%v' and ID '%v'",
				logsOutputFilepath,
				containerName,
				containerId,
			)
		}
	}

	return nil
}

func (backend *DockerKurtosisBackend) waitForContainerExits(
	ctx context.Context,
	containers []*types.Container,
) (
	resultSuccessfulContainers map[string]bool,
	resultErroredContainers map[string]error,
) {
	successfulContainers := map[string]bool{}
	erroredContainers := map[string]error{}
	// TODO Parallelize for perf
	for _, container := range containers {
		containerId := container.GetId()
		if _, err := backend.dockerManager.WaitForExit(ctx, containerId); err != nil {
			containerError := stacktrace.Propagate(
				err,
				"An error occurred waiting for container '%v' with ID '%v' to exit",
				container.GetName(),
				containerId,
			)
			erroredContainers[container.GetId()] = containerError
			continue
		}
		successfulContainers[containerId] = true
	}

	return successfulContainers, erroredContainers
}

func destroyContainersInEnclaves(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
	enclaves map[enclave.EnclaveID]*matchingNetworkInformation,
) (
	map[enclave.EnclaveID]bool,
	map[enclave.EnclaveID]error,
	error,
){
	// For all the enclaves to destroy, gather all the containers that should be destroyed
	enclaveIdsForContainerIdsToRemove := map[string]enclave.EnclaveID{}
	containerIdsToRemoveToUncastedContainerId := map[string]interface{}{}
	for enclaveId, networkInfo := range enclaves {
		for _, container := range networkInfo.containers {
			containerId := container.GetId()
			enclaveIdsForContainerIdsToRemove[containerId] = enclaveId
			containerIdsToRemoveToUncastedContainerId[containerId] = interface{}(containerId)
		}
	}

	var removeEnclaveContainerOperation docker_task_parallelizer.DockerOperation = func(ctx context.Context, dockerManager *docker_manager.DockerManager, dockerObjectId string) error {
		if err := dockerManager.RemoveContainer(ctx, dockerObjectId); err != nil {
			return stacktrace.Propagate(err, "An error occurred removing enclave container with ID '%v'", dockerObjectId)
		}
		return nil
	}

	_, erroredContainerIds, err := docker_task_parallelizer.RunDockerOperationInParallelForKurtosisObject(
		ctx,
		containerIdsToRemoveToUncastedContainerId,
		dockerManager,
		func(uncastedContainerId interface{}) (string, error) {
			containerIdStr, ok := uncastedContainerId.(string)
			if !ok {
				return "", stacktrace.NewError("Failed to cast uncasted container ID to a casted string container ID")
			}
			return containerIdStr, nil
		},
		removeEnclaveContainerOperation,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred removing the containers of enclaves matching filters '%+v'", filters)
	}

	containerRemovalErrorStrsByEnclave := map[enclave.EnclaveID][]string{}
	for erroredContainerId, removeContainerErr := range erroredContainerIds {
		containerEnclaveId, found := enclaveIdsForContainerIdsToRemove[erroredContainerId]
		if !found {
			return nil, nil, stacktrace.NewError("An error occurred stopping container '%v' in an enclave we didn't request", erroredContainerId)
		}

		existingEnclaveErrors, found := containerRemovalErrorStrsByEnclave[containerEnclaveId]
		if !found {
			existingEnclaveErrors = []string{}
		}
		containerRemovalErrorStrsByEnclave[containerEnclaveId] = append(existingEnclaveErrors, removeContainerErr.Error())
	}

	erroredEnclaveIds := map[enclave.EnclaveID]error{}
	successfulEnclaveIds := map[enclave.EnclaveID]bool{}
	for enclaveId, containerRemovalErrorStrs := range containerRemovalErrorStrsByEnclave {
		if len(containerRemovalErrorStrs) == 0 {
			successfulEnclaveIds[enclaveId] = true
			continue
		}
		errorStr := strings.Join(containerRemovalErrorStrs, "\n\n")
		erroredEnclaveIds[enclaveId] = stacktrace.NewError(
			"One or more errors occurred removing the containers in enclave '%v':\n%v",
			enclaveId,
			errorStr,
		)
		continue
	}

	return successfulEnclaveIds, erroredEnclaveIds, nil
}

func destroyVolumesInEnclaves(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
	enclaves map[enclave.EnclaveID]bool,
) (
	map[enclave.EnclaveID]bool,
	map[enclave.EnclaveID]error,
	error,
) {
	// After we've tried to destroy all the containers from the enclaves, take the successful ones and destroy their volumes
	enclaveIdsForVolumeIdsToRemove := map[string]enclave.EnclaveID{}
	volumeIdsToRemoveToUncastedVolumeId := map[string]interface{}{}
	for enclaveId := range enclaves {
		enclaveVolumeIds, err := getAllEnclaveVolumes(ctx, dockerManager, enclaveId)
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred getting the volumes for enclave '%v'", enclaveId)
		}

		for _, volume := range enclaveVolumeIds {
			volumeId := volume.Name
			enclaveIdsForVolumeIdsToRemove[volumeId] = enclaveId
			volumeIdsToRemoveToUncastedVolumeId[volumeId] = interface{}(volumeId)
		}
	}

	var removeEnclaveVolumeOperation docker_task_parallelizer.DockerOperation = func(ctx context.Context, dockerManager *docker_manager.DockerManager, dockerObjectId string) error {
		if err := dockerManager.RemoveVolume(ctx, dockerObjectId); err != nil {
			return stacktrace.Propagate(err, "An error occurred removing enclave volume with ID '%v'", dockerObjectId)
		}
		return nil
	}

	_, erroredVolumeIds, err := docker_task_parallelizer.RunDockerOperationInParallelForKurtosisObject(
		ctx,
		volumeIdsToRemoveToUncastedVolumeId,
		dockerManager,
		func(uncastedVolumeId interface{}) (string, error) {
			volumeIdStr, ok := uncastedVolumeId.(string)
			if !ok {
				return "", stacktrace.NewError("Failed to cast uncasted volume ID to a casted string volume ID")
			}
			return volumeIdStr, nil
		},
		removeEnclaveVolumeOperation,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred trying to remove the volumes for enclaves whose containers were successfully destroyed")
	}

	volumeRemovalErrorStrsByEnclave := map[enclave.EnclaveID][]string{}
	for erroredVolumeId, removeVolumeErr := range erroredVolumeIds {
		volumeEnclaveId, found := enclaveIdsForVolumeIdsToRemove[erroredVolumeId]
		if !found {
			return nil, nil, stacktrace.NewError("An error occurred removing volume '%v' in an enclave we didn't request", erroredVolumeId)
		}

		existingEnclaveErrors, found := volumeRemovalErrorStrsByEnclave[volumeEnclaveId]
		if !found {
			existingEnclaveErrors = []string{}
		}
		volumeRemovalErrorStrsByEnclave[volumeEnclaveId] = append(existingEnclaveErrors, removeVolumeErr.Error())
	}

	erroredEnclaveIds := map[enclave.EnclaveID]error{}
	successfulEnclaveIds := map[enclave.EnclaveID]bool{}
	for enclaveId, volumeRemovalErrorStrs := range volumeRemovalErrorStrsByEnclave {
		if len(volumeRemovalErrorStrs) == 0 {
			successfulEnclaveIds[enclaveId] = true
			continue
		}
		errorStr := strings.Join(volumeRemovalErrorStrs, "\n\n")
		erroredEnclaveIds[enclaveId] = stacktrace.NewError(
			"One or more errors occurred removing the volumes in enclave '%v':\n%v",
			enclaveId,
			errorStr,
		)
		continue
	}

	return successfulEnclaveIds, erroredEnclaveIds, nil
}

func destroyEnclaveNetworks(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
	enclaveNetworkIds map[enclave.EnclaveID]string,
) (map[enclave.EnclaveID]bool, map[enclave.EnclaveID]error, error) {
	networkIdsToUncastedNetworkId := map[string]interface{}{}
	for _, networkId := range enclaveNetworkIds {
		networkIdsToUncastedNetworkId[networkId] = interface{}(networkId)
	}

	var removeNetworkOperation docker_task_parallelizer.DockerOperation = func(ctx context.Context, dockerManager *docker_manager.DockerManager, dockerObjectId string) error {
		if err := dockerManager.RemoveNetwork(ctx, dockerObjectId); err != nil {
			return stacktrace.Propagate(err, "An error occurred removing enclave network with ID '%v'", dockerObjectId)
		}
		return nil
	}

	successfulNetworkIds, erroredNetworkIds, err := docker_task_parallelizer.RunDockerOperationInParallelForKurtosisObject(
		ctx,
		networkIdsToUncastedNetworkId,
		dockerManager,
		func(uncastedVolumeId interface{}) (string, error) {
			volumeIdStr, ok := uncastedVolumeId.(string)
			if !ok {
				return "", stacktrace.NewError("Failed to cast uncasted volume ID to a casted string volume ID")
			}
			return volumeIdStr, nil
		},
		removeNetworkOperation,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred trying to remove the volumes for enclaves whose containers were successfully destroyed")
	}




	successfulDeletedNetworksEnclaveIds := map[enclave.EnclaveID]bool{}
	for _, network := range networks {
		networkName := network.GetName()
		enclaveId := enclave.EnclaveID(networkName)
		if err := backend.dockerManager.RemoveNetwork(ctx, network.GetId()); err != nil {
			erroredEnclaveIds[enclaveId] = stacktrace.Propagate(err, "An error occurred removing the network for enclave '%v'", enclaveId)
			continue
		}
		successfulDeletedNetworksEnclaveIds[enclaveId] = true
	}

	successfulContainerRemovalEnclaveIds := successfulDeletedNetworksEnclaveIds
}