package docker

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"strings"
)

const (
	shouldFetchStoppedContainersWhenGettingEnclaveStatus = true
)

func (backend *DockerKurtosisBackend) CreateEnclave(
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

	networks, err := backend.getEnclaveNetworksByEnclaveIds(ctx, enclaveIDs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred checking for networks with ID '%v', which is necessary to ensure that our enclave doesn't exist yet", enclaveId)
	}
	if len(networks) > 0 {
		return nil, stacktrace.NewError("Cannot create enclave with ID '%v' because an enclave with ID '%v' already exists", enclaveId, enclaveId)
	}

	enclaveObjAttrsProvider, err := backend.objAttrsProvider.ForEnclave(enclaveId)
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
				logrus.Errorf("Creating the enclave didn't complete successfully, so we tried to delete network '%v' that we created but an error was thrown:", networkId)
				fmt.Fprintln(logrus.StandardLogger().Out, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove network with ID '%v'!!!!!!!", networkId)
			}
		}
	}()
	logrus.Debugf("Docker network '%v' created successfully with ID '%v' and subnet CIDR '%v'", enclaveId, networkId, networkIpAndMask.String())

	newEnclave := enclave.NewEnclave(enclaveId, enclave.EnclaveStatus_Empty, networkId, networkIpAndMask.String(), gatewayIp, freeIpAddrTracker)

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

	networks, err := backend.getEnclaveNetworksByEnclaveIds(ctx, filters.IDs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave networks by enclave IDs '%+v'", filters.IDs)
	}

	result := map[enclave.EnclaveID]*enclave.Enclave{}

	for _, network := range networks {
		enclaveId, err := getEnclaveIdFromNetwork(network)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting enclave ID from network '%+v'", network)
		}

		enclaveStatus, _, err := backend.getEnclaveStatusAndContainers(ctx, enclaveId)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting enclave status and containers for enclave with ID '%v'", enclaveId)
		}

		if filters.Statuses == nil || isEnclaveStatusInEnclaveFilters(enclaveStatus, filters) {
			// TODO We're returning nil here for gatewayIp and freeIpAddrProvider as a temporary hack, until we can fully push all Docker stuff under the KurtosisBackend
			enclave := enclave.NewEnclave(enclaveId, enclaveStatus, network.GetId(), network.GetIpAndMask().String(), nil, nil)
			result[enclaveId] = enclave
		}
	}

	return result, nil
}

// TODO MAYYYYYYYBE DumpEnclaves?

// Stops enclaves matching the given filters
func (backend *DockerKurtosisBackend) StopEnclaves(
	ctx context.Context,
	filters *enclave.EnclaveFilters,
) (
	successfulEnclaveIds map[enclave.EnclaveID]bool,
	erroredEnclaveIds map[enclave.EnclaveID]error,
	resultErr error,
) {

	networks, err := backend.getEnclaveNetworksByEnclaveIds(ctx, filters.IDs)
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

		enclaveStatus, containers, err := backend.getEnclaveStatusAndContainers(ctx, enclaveId)
		if err != nil {
			erroredEnclaveIds[enclaveId] = stacktrace.Propagate(err, "An error occurred getting enclave status and containers for enclave with ID '%v'", enclaveId)
			continue
		}

		if filters.Statuses == nil || isEnclaveStatusInEnclaveFilters(enclaveStatus, filters) {

			logrus.Debugf("Containers in enclave '%v' that will be stopped: %+v", enclaveId, containers)

			if _, erroredContainers := backend.killContainers(ctx, containers); len(erroredContainers) > 0 {
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
			if _, erroredContainers := backend.waitForExitContainers(ctx, containers); len(erroredContainers) > 0 {
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

// Destroys enclaves matching the given filters
func (backend *DockerKurtosisBackend) DestroyEnclaves(
	ctx context.Context,
	filters *enclave.EnclaveFilters,
) (
	successfulEnclaveIds map[enclave.EnclaveID]bool,
	erroredEnclaveIds map[enclave.EnclaveID]error,
	resultErr error,
) {
	// Stop containers
	resultSuccessfulEnclaveIds, resultErroredEnclaveIds, err := backend.StopEnclaves(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred stopping enclaves using filters '%v'", filters)
	}

	for enclaveId, err := range resultErroredEnclaveIds {
		erroredEnclaveIds[enclaveId] = err
	}

	// Remove containers
	for enclaveId := range resultSuccessfulEnclaveIds {
		_, containers, err := backend.getEnclaveStatusAndContainers(ctx, enclaveId)
		if err != nil {
			erroredEnclaveIds[enclaveId] = stacktrace.Propagate(err, "An error occurred getting enclave status and containers for enclave with ID '%v'", enclaveId)
			continue
		}

		if _, erroredContainers := backend.removeContainers(ctx, containers); len(erroredContainers) > 0 {
			removeContainerErrorStrs := []string{}
			for _, err = range erroredContainers {
				removeContainerErrorStrs = append(removeContainerErrorStrs, err.Error())
			}
			errorStr := strings.Join(removeContainerErrorStrs, "\n\n")

			erroredEnclaveIds[enclaveId] = stacktrace.NewError(
				"An error occurred removing one or more containers in enclave '%v':\n%v",
				enclaveId,
				errorStr,
			)
		}
		successfulEnclaveIds[enclaveId] = true
	}

	// Remove the networks
	networks, err := backend.getEnclaveNetworksByEnclaveIds(ctx, successfulEnclaveIds)
	if err != nil {
		return successfulEnclaveIds, erroredEnclaveIds, stacktrace.Propagate(err, "An error occurred getting enclave networks by enclave IDs '%+v'", successfulEnclaveIds)
	}

	for _, network := range networks {
		if err := backend.dockerManager.RemoveNetwork(ctx, network.GetId()); err != nil {
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
func (backend *DockerKurtosisBackend) getEnclaveNetworkByEnclaveId(ctx context.Context, enclaveId enclave.EnclaveID) (*types.Network, error) {
	enclaveIDs := map[enclave.EnclaveID]bool{
		enclaveId: true,
	}

	enclaveNetworksFound, err := backend.getEnclaveNetworksByEnclaveIds(ctx, enclaveIDs)
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

func (backend *DockerKurtosisBackend) getEnclaveNetworksByEnclaveIds(ctx context.Context, enclaveIds map[enclave.EnclaveID]bool) ([]*types.Network, error) {
	enclaveNetworks := []*types.Network{}
	if len(enclaveIds) == 0 {
		return enclaveNetworks, nil
	}

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
		if _, found := enclaveIds[enclaveId]; found {
			enclaveNetworks = append(enclaveNetworks, network)
		}
	}

	return enclaveNetworks, nil
}

func (backend *DockerKurtosisBackend) getEnclaveStatusAndContainers(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
) (

	enclave.EnclaveStatus,
	[]*types.Container,
	error,
) {
	allEnclaveContainers, err := backend.getEnclaveContainers(ctx, enclaveId)
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

func (backend *DockerKurtosisBackend) getEnclaveContainers(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
) ([]*types.Container, error) {
	searchLabels := map[string]string{
		label_key_consts.EnclaveIDLabelKey.GetString(): string(enclaveId),
	}
	containers, err := backend.dockerManager.GetContainersByLabels(ctx, searchLabels, shouldFetchStoppedContainersWhenGettingEnclaveStatus)
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
