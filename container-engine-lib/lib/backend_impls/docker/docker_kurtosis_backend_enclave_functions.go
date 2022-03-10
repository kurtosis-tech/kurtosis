package docker

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"strings"
)

const (
	shouldFetchStoppedContainersWhenGettingEnclaveStatus = true
)

func (backend *DockerKurtosisBackend) CreateEnclave(
	ctx context.Context,
	enclaveId string,
) (
	*enclave.Enclave,
	error,
) {
	teardownCtx := context.Background() // Separate context for tearing stuff down in case the input context is cancelled

	_, found, err := backend.getEnclaveNetwork(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred checking for networks with name '%v', which is necessary to ensure that our enclave doesn't exist yet", enclaveId)
	}
	if found {
		return nil, stacktrace.NewError("Cannot create enclave '%v' because an enclave with that name already exists", enclaveId)
	}

	enclaveObjAttrsProvider, err := backend.objAttrsProvider.ForEnclave(enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while trying to generate an object attributes provider for the enclave with name '%v'", enclaveId)
	}
	enclaveNetworkAttrs, err := enclaveObjAttrsProvider.ForEnclaveNetwork()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while trying to get the enclave network attributes for the enclave with name '%v'", enclaveId)
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
	map[string]*enclave.Enclave,
	error,
) {

	networks, err := backend.getEnclaveNetworksByEnclaveIds(ctx, filters.IDs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave networks by enclave IDs '%+v'", filters.IDs)
	}

	result := map[string]*enclave.Enclave{}

	for _, network := range networks {
		enclaveId := network.GetName()

		enclaveStatus, _, err := backend.getEnclaveStatusAndContainers(ctx, enclaveId)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting enclave status and containers for enclave with ID '%v'", enclaveId)
		}

		if filters.Statuses == nil || isEnclaveStatusInEnclaveFilters(enclaveStatus, filters) {
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
	successfulEnclaveIds map[string]bool,
	erroredEnclaveIds map[string]error,
	resultErr error,
) {

	networks, err := backend.getEnclaveNetworksByEnclaveIds(ctx, filters.IDs)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting enclave networks by enclave IDs '%+v'", filters.IDs)
	}
	if len(networks) == 0 {
		return nil, nil, stacktrace.Propagate(err, "No Enclave was founded with some of these IDs '%+v'", filters.IDs)
	}

	for _, network := range networks {
		enclaveId := network.GetName()

		enclaveStatus, containers, err := backend.getEnclaveStatusAndContainers(ctx, enclaveId)
		if err != nil {
			erroredEnclaveIds[enclaveId] = stacktrace.Propagate(err, "An error occurred getting enclave status and containers for enclave with ID '%v'", enclaveId)
			continue
		}

		if filters.Statuses == nil || isEnclaveStatusInEnclaveFilters(enclaveStatus, filters) {

			logrus.Debugf("Containers in enclave '%v' that will be stopped: %+v", enclaveId, containers)

			// TODO Parallelize for perf
			containerKillErrorStrs := []string{}
			for _, container := range containers {
				containerId := container.GetId()
				containerName := container.GetName()
				if err := backend.dockerManager.KillContainer(ctx, containerId); err != nil {
					wrappedContainerKillErr := stacktrace.Propagate(
						err,
						"An error occurred killing container '%v' with ID '%v'",
						containerName,
						containerId,
					)
					containerKillErrorStrs = append(
						containerKillErrorStrs,
						wrappedContainerKillErr.Error(),
					)
				}
			}

			if len(containerKillErrorStrs) > 0 {
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
			containerWaitErrorStrs := []string{}
			for _, container := range containers {
				containerName := container.GetName()
				containerId := container.GetId()
				if _, err := backend.dockerManager.WaitForExit(ctx, containerId); err != nil {
					wrappedContainerWaitErr := stacktrace.Propagate(
						err,
						"An error occurred waiting for container '%v' with ID '%v' to exit after killing",
						containerName,
						containerId,
					)
					containerWaitErrorStrs = append(
						containerWaitErrorStrs,
						wrappedContainerWaitErr.Error(),
					)
				}
			}

			if len(containerWaitErrorStrs) > 0 {
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
	successfulEnclaveIds map[string]bool,
	erroredEnclaveIds map[string]error,
	resultErr error,
) {
	// Stop containers
	successfulEnclaveIds, erroredEnclaveIds, err := backend.StopEnclaves(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred stopping enclaves using filters '%v'", filters)
	}

	// Remove containers
	for enclaveId := range successfulEnclaveIds {
		_, containers, err := backend.getEnclaveStatusAndContainers(ctx, enclaveId)
		if err != nil {
			delete(successfulEnclaveIds, enclaveId)
			erroredEnclaveIds[enclaveId] = stacktrace.Propagate(err, "An error occurred getting enclave status and containers for enclave with ID '%v'", enclaveId)
			continue
		}

		removeContainerErrorStrs := []string{}
		for _, container := range containers {
			containerName := container.GetName()
			containerId := container.GetId()
			if err := backend.dockerManager.RemoveContainer(ctx, containerId); err != nil {
				wrappedErr := stacktrace.Propagate(
					err,
					"An error occurred removing container '%v' with ID '%v'",
					containerName,
					containerId,
				)
				removeContainerErrorStrs = append(
					removeContainerErrorStrs,
					wrappedErr.Error(),
				)
			}
		}

		if len(removeContainerErrorStrs) > 0 {
			delete(successfulEnclaveIds, enclaveId)
			erroredEnclaveIds[enclaveId] = stacktrace.NewError(
				"An error occurred removing one or more containers in enclave '%v':\n%v",
				enclaveId,
				strings.Join(
					removeContainerErrorStrs,
					"\n\n",
				),
			)
		}
	}

	// Remove the networks
	networks, err := backend.getEnclaveNetworksByEnclaveIds(ctx, successfulEnclaveIds)
	if err != nil {
		return successfulEnclaveIds, erroredEnclaveIds, stacktrace.Propagate(err, "An error occurred getting enclave networks by enclave IDs '%+v'", successfulEnclaveIds)
	}
	if len(networks) == 0 {
		return successfulEnclaveIds, erroredEnclaveIds, stacktrace.Propagate(err, "No Enclave was founded with some of these IDs '%+v'", successfulEnclaveIds)
	}

	for _, network := range networks {
		if err := backend.dockerManager.RemoveNetwork(ctx, network.GetId()); err != nil {
			enclaveId := network.GetName()
			delete(successfulEnclaveIds, enclaveId)
			erroredEnclaveIds[enclaveId] = stacktrace.Propagate(err, "An error occurred removing the network for enclave '%v'", enclaveId)
		}
	}

	return successfulEnclaveIds, erroredEnclaveIds, nil
}

// Repartition the enclave network
func (backendCore *DockerKurtosisBackend) RepartitionEnclave(
	ctx context.Context,
	enclaveId string,
	servicesConnections map[service.ServiceID]map[service.ServiceID]enclave.NetworkConnection,
) (
	resultErr error,
) {
	panic("Implement me")
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================
func (backend *DockerKurtosisBackend) getEnclaveNetworksByEnclaveIds(ctx context.Context, enclaveIds map[string]bool) ([]*types.Network, error) {
	kurtosisNetworkLabels := map[string]string{
		label_key_consts.AppIDLabelKey.GetString(): label_value_consts.AppIDLabelValue.GetString(),
	}

	for enclaveId := range enclaveIds {
		kurtosisNetworkLabels[label_key_consts.EnclaveIDLabelKey.GetString()] = enclaveId
	}

	networks, err := backend.dockerManager.GetNetworksByLabels(ctx, kurtosisNetworkLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kurtosis networks")
	}

	return networks, nil
}

// There is a 1:1 mapping between Docker network and enclave - no network, no enclave, and vice versa
// We therefore use this function to check for the existence of an enclave, as well as get network info about existing enclaves
func (backend *DockerKurtosisBackend) getEnclaveNetwork(ctx context.Context, enclaveId string) (*types.Network, bool, error) {
	allNetworksWithEnclaveIdInName, err := backend.dockerManager.GetNetworksByName(ctx, enclaveId)
	if err != nil {
		return nil, false, stacktrace.Propagate(err, "An error occurred getting networks matching name '%v'", enclaveId)
	}

	// NOTE: GetNetworksByName will match networks that have the enclaveId *even as a substring*, so we have to filter again
	// to get the network (if any) that has a name *exactly* == enclave ID
	matchingNetworks := []*types.Network{}
	for _, networkWithEnclaveId := range allNetworksWithEnclaveIdInName {
		if networkWithEnclaveId.GetName() == enclaveId {
			matchingNetworks = append(matchingNetworks, networkWithEnclaveId)
		}
	}

	numMatchingNetworks := len(matchingNetworks)
	logrus.Debugf("Found %v networks matching name '%v': %+v", numMatchingNetworks, enclaveId, matchingNetworks)
	if numMatchingNetworks > 1 {
		return nil, false, stacktrace.NewError(
			"Found %v networks matching name '%v' when we expected just one - this is likely a bug in Kurtosis!",
			numMatchingNetworks,
			enclaveId,
		)
	}
	if numMatchingNetworks == 0 {
		return nil, false, nil
	}
	network := matchingNetworks[0]
	return network, true, nil
}

func (backend *DockerKurtosisBackend) getEnclaveStatusAndContainers(
	ctx context.Context,
	enclaveId string,
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
		isEnclaveContainerRunning := containerStatus == types.ContainerStatus_Running || containerStatus == types.ContainerStatus_Restarting || containerStatus == types.ContainerStatus_Removing
		if isEnclaveContainerRunning {
			resultEnclaveStatus = enclave.EnclaveStatus_Running
		}
	}

	return resultEnclaveStatus, allEnclaveContainers, nil
}

func (backend *DockerKurtosisBackend) getEnclaveContainers(
	ctx context.Context,
	enclaveId string,
) ([]*types.Container, error) {
	searchLabels := map[string]string{
		label_key_consts.EnclaveIDLabelKey.GetString(): enclaveId,
	}
	containers, err := backend.dockerManager.GetContainersByLabels(ctx, searchLabels, shouldFetchStoppedContainersWhenGettingEnclaveStatus)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the containers for enclave '%v' by labels '%+v'", enclaveId, searchLabels)
	}
	return containers, nil
}

func isEnclaveStatusInEnclaveFilters(enclaveStatus enclave.EnclaveStatus, filters *enclave.EnclaveFilters) bool {
	for expectedEnclaveStatus := range filters.Statuses {
		if enclaveStatus == expectedEnclaveStatus {
			return true
		}
	}
	return false
}
