package docker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/gammazero/workerpool"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
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
	numContainersToDumpAtOnce = 20

	// Permisssions for the files & directories we create as a result of the dump
	createdDirPerms  = 0755
	createdFilePerms = 0644

	shouldFollowContainerLogsWhenDumping    = false

	containerSpecJsonSerializationIndent = "  "
	containerSpecJsonSerializationPrefix = ""
)

func (backendCore *DockerKurtosisBackend) CreateEnclave(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	isPartitioningEnabled bool,
) (
	*enclave.Enclave,
	error,
) {
	teardownCtx := context.Background() // Separate context for tearing stuff down in case the input context is cancelled

	enclaveIDs := map[enclave.EnclaveID]bool{
		enclaveId: true,
	}

	networks, err := backendCore.getEnclaveNetworksByEnclaveIds(ctx, enclaveIDs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred checking for networks with ID '%v', which is necessary to ensure that our enclave doesn't exist yet", enclaveId)
	}
	if len(networks) > 0 {
		return nil, stacktrace.NewError("Cannot create enclave with ID '%v' because an enclave with ID '%v' already exists", enclaveId, enclaveId)
	}

	enclaveObjAttrsProvider, err := backendCore.objAttrsProvider.ForEnclave(enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while trying to generate an object attributes provider for the enclave with ID '%v'", enclaveId)
	}

	enclaveNetworkAttrs, err := enclaveObjAttrsProvider.ForEnclaveNetwork(isPartitioningEnabled)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while trying to get the enclave network attributes for the enclave with ID '%v'", enclaveId)
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
	networkId, networkIpAndMask, gatewayIp, freeIpAddrTracker, err := backendCore.dockerNetworkAllocator.CreateNewNetwork(
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
			if err := backendCore.dockerManager.RemoveNetwork(teardownCtx, networkId); err != nil {
				logrus.Errorf("Creating the enclave didn't complete successfully, so we tried to delete network '%v' that we created but an error was thrown:", networkId)
				if _, err := fmt.Fprintln(logrus.StandardLogger().Out, err); err != nil {
					logrus.Errorf("And error occurred trying to print the remove network error")
				}
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove network with ID '%v'!!!!!!!", networkId)
			}
		}
	}()
	logrus.Debugf("Docker network '%v' created successfully with ID '%v' and subnet CIDR '%v'", enclaveId, networkId, networkIpAndMask.String())

	newEnclave := enclave.NewEnclave(enclaveId, enclave.EnclaveStatus_Empty, networkId, networkIpAndMask.String(), gatewayIp, freeIpAddrTracker)

	return newEnclave, nil
}

// Gets enclaves matching the given filters
func (backendCore *DockerKurtosisBackend) GetEnclaves(
	ctx context.Context,
	filters *enclave.EnclaveFilters,
) (
	map[enclave.EnclaveID]*enclave.Enclave,
	error,
) {

	networks, err := backendCore.getEnclaveNetworksByEnclaveIds(ctx, filters.IDs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave networks by enclave IDs '%+v'", filters.IDs)
	}

	result := map[enclave.EnclaveID]*enclave.Enclave{}

	for _, network := range networks {
		enclaveId, err := getEnclaveIdFromNetwork(network)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting enclave ID from network '%+v'", network)
		}

		enclaveStatus, _, err := backendCore.getEnclaveStatusAndContainers(ctx, enclaveId)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting enclave status and containers for enclave with ID '%v'", enclaveId)
		}

		if filters.Statuses == nil || isEnclaveStatusInEnclaveFilters(enclaveStatus, filters) {
			// TODO We're returning nil here for gatewayIp and freeIpAddrProvider as a temporary hack, until we can fully push all Docker stuff under the KurtosisBackend
			newEnclave := enclave.NewEnclave(enclaveId, enclaveStatus, network.GetId(), network.GetIpAndMask().String(), nil, nil)
			result[enclaveId] = newEnclave
		}
	}

	return result, nil
}

// Stops enclaves matching the given filters
func (backendCore *DockerKurtosisBackend) StopEnclaves(
	ctx context.Context,
	filters *enclave.EnclaveFilters,
) (
	successfulEnclaveIds map[enclave.EnclaveID]bool,
	erroredEnclaveIds map[enclave.EnclaveID]error,
	resultErr error,
) {

	networks, err := backendCore.getEnclaveNetworksByEnclaveIds(ctx, filters.IDs)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting enclave networks by enclave IDs '%+v'", filters.IDs)
	}
	if len(networks) == 0 {
		return nil, nil, stacktrace.Propagate(err, "No Enclave was found with IDs '%+v'", filters.IDs)
	}

	for _, network := range networks {
		enclaveId, err := getEnclaveIdFromNetwork(network)
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred getting enclave ID from network '%+v'", network)
		}

		enclaveStatus, containers, err := backendCore.getEnclaveStatusAndContainers(ctx, enclaveId)
		if err != nil {
			erroredEnclaveIds[enclaveId] = stacktrace.Propagate(err, "An error occurred getting enclave status and containers for enclave with ID '%v'", enclaveId)
			continue
		}

		if filters.Statuses == nil || isEnclaveStatusInEnclaveFilters(enclaveStatus, filters) {

			logrus.Debugf("Containers in enclave '%v' that will be stopped: %+v", enclaveId, containers)

			if _, erroredContainers := backendCore.killContainers(ctx, containers); len(erroredContainers) > 0 {
				containerKillErrorStrs := []string{}
				for _, err = range erroredContainers{
					containerKillErrorStrs = append(containerKillErrorStrs, err.Error())
				}
				errorStr := strings.Join(containerKillErrorStrs, "\n\n")
				erroredEnclaveIds[enclaveId] = stacktrace.NewError(
					"One or more errors occurred killing the containers in enclave '%v':\n%v",
					enclaveId,
					errorStr,
				)
				continue
			}

			// If all the kills went off successfully, wait for all the containers we just killed to definitively exit
			//  before we return
			if _, erroredContainers := backendCore.waitForContainerExits(ctx, containers); len(erroredContainers) > 0 {
				containerWaitErrorStrs := []string{}
				for _, err = range erroredContainers{
					containerWaitErrorStrs = append(containerWaitErrorStrs, err.Error())
				}
				errorStr := strings.Join(containerWaitErrorStrs, "\n\n")
				erroredEnclaveIds[enclaveId] = stacktrace.NewError(
					"One or more errors occurred waiting for containers in enclave '%v' to exit after killing, meaning we can't guarantee the enclave is completely stopped:\n%v",
					enclaveId,
					errorStr,
				)
				continue
			}

			successfulEnclaveIds[enclaveId] = true
		}
	}
	return successfulEnclaveIds, erroredEnclaveIds, nil
}

func (backend *DockerKurtosisBackend) DumpEnclave(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	outputDirpath string,
) error {
	enclaveContainerSearchLabels := map[string]string{
		label_key_consts.AppIDLabelKey.GetString(): label_value_consts.AppIDLabelValue.GetString(),
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
func (backendCore *DockerKurtosisBackend) DestroyEnclaves(
	ctx context.Context,
	filters *enclave.EnclaveFilters,
) (
	successfulEnclaveIds map[enclave.EnclaveID]bool,
	erroredEnclaveIds map[enclave.EnclaveID]error,
	resultErr error,
) {
	// Stop containers
	resultSuccessfulEnclaveIds, resultErroredEnclaveIds, err := backendCore.StopEnclaves(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred stopping enclaves using filters '%v'", filters)
	}

	for enclaveId, err := range resultErroredEnclaveIds {
		erroredEnclaveIds[enclaveId] = err
	}

	// Remove containers
	for enclaveId := range resultSuccessfulEnclaveIds {
		containers, err := backendCore.getEnclaveContainers(ctx, enclaveId)
		if err != nil {
			erroredEnclaveIds[enclaveId] = stacktrace.Propagate(err, "An error occurred getting enclave containers for enclave with ID '%v'", enclaveId)
			continue
		}

		for _, container := range containers {
			removeContainerErrorStrs := []string{}
			if err := backendCore.dockerManager.RemoveContainer(ctx, container.GetId()); err != nil {
				wrappedErrStr := fmt.Sprintf(
					"An error occurred removing container with ID '%v':\n%v",
					container.GetId(),
					err.Error(),
				)
				removeContainerErrorStrs = append(removeContainerErrorStrs, wrappedErrStr)
				continue
			}
			if len(removeContainerErrorStrs) > 0 {
				errorStr := strings.Join(removeContainerErrorStrs, "\n\n")
				erroredEnclaveIds[enclaveId] = stacktrace.NewError(
					"An error occurred removing one or more containers in enclave '%v':\n%v",
					enclaveId,
					errorStr,
				)
			}
		}

		successfulEnclaveIds[enclaveId] = true
	}

	// Remove the networks
	networks, err := backendCore.getEnclaveNetworksByEnclaveIds(ctx, successfulEnclaveIds)
	if err != nil {
		return successfulEnclaveIds, erroredEnclaveIds, stacktrace.Propagate(err, "An error occurred getting enclave networks by enclave IDs '%+v'", successfulEnclaveIds)
	}

	for _, network := range networks {
		if err := backendCore.dockerManager.RemoveNetwork(ctx, network.GetId()); err != nil {
			networkName := network.GetName()
			enclaveId := enclave.EnclaveID(networkName)
			delete(successfulEnclaveIds, enclaveId)
			erroredEnclaveIds[enclaveId] = stacktrace.Propagate(err, "An error occurred removing the network for enclave '%v'", enclaveId)
		}
	}

	return successfulEnclaveIds, erroredEnclaveIds, nil
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================
func (backendCore *DockerKurtosisBackend) getEnclaveNetworkByEnclaveId(ctx context.Context, enclaveId enclave.EnclaveID) (*types.Network, error) {
	enclaveIDs := map[enclave.EnclaveID]bool{
		enclaveId: true,
	}

	enclaveNetworksFound, err := backendCore.getEnclaveNetworksByEnclaveIds(ctx, enclaveIDs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Docker networks by enclave ID '%v'", enclaveId)
	}
	numMatchingNetworks := len(enclaveNetworksFound)
	if numMatchingNetworks == 0 {
		return nil, stacktrace.Propagate(err, "No network was found for enclave with ID '%v'; it should never happens, it's a bug in Kurtosis", enclaveId)
	}
	if numMatchingNetworks > 1 {
		return nil, stacktrace.NewError(
			"Expected exactly one network matching enclave ID '%v', but got %v",
			enclaveId,
			numMatchingNetworks,
		)
	}
	return  enclaveNetworksFound[0], nil
}

func (backendCore *DockerKurtosisBackend) getEnclaveNetworksByEnclaveIds(ctx context.Context, enclaveIds map[enclave.EnclaveID]bool) ([]*types.Network, error) {
	enclaveNetworks := []*types.Network{}
	if len(enclaveIds) == 0 {
		return enclaveNetworks, nil
	}

	kurtosisNetworkLabels := map[string]string{
		label_key_consts.AppIDLabelKey.GetString(): label_value_consts.AppIDLabelValue.GetString(),
	}

	appNetworks, err := backendCore.dockerManager.GetNetworksByLabels(ctx, kurtosisNetworkLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kurtosis networks")
	}

	for _, network := range appNetworks {
		enclaveId, err := getEnclaveIdFromNetwork(network)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting enclave ID from network '%+v'; it should never happens it's a bug in Kurtosis", network)
		}
		if _, found := enclaveIds[enclaveId]; found {
			enclaveNetworks = append(enclaveNetworks, network)
		}
	}

	return enclaveNetworks, nil
}

func (backendCore *DockerKurtosisBackend) getEnclaveStatusAndContainers(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
) (

	enclave.EnclaveStatus,
	[]*types.Container,
	error,
) {
	allEnclaveContainers, err := backendCore.getEnclaveContainers(ctx, enclaveId)
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

func (backendCore *DockerKurtosisBackend) getEnclaveContainers(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
) ([]*types.Container, error) {
	searchLabels := map[string]string{
		label_key_consts.EnclaveIDLabelKey.GetString(): string(enclaveId),
	}
	containers, err := backendCore.dockerManager.GetContainersByLabels(ctx, searchLabels, shouldFetchStoppedContainersWhenGettingEnclaveStatus)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the containers for enclave '%v' by labels '%+v'", enclaveId, searchLabels)
	}
	return containers, nil
}

func isEnclaveStatusInEnclaveFilters(enclaveStatus enclave.EnclaveStatus, filters *enclave.EnclaveFilters) bool {
	if _, found := filters.Statuses[enclaveStatus]; found {
		return true
	}

	return false
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
