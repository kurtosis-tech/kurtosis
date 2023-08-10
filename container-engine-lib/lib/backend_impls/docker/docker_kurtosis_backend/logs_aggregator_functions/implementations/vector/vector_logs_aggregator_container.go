package vector

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
)

type vectorLogsAggregatorContainer struct{}

func NewVectorLogsAggregatorContainer() *vectorLogsAggregatorContainer {
	return &vectorLogsAggregatorContainer{}
}

func (vectorContainer *vectorLogsAggregatorContainer) CreateAndStart(
	ctx context.Context,
	logsListeningPortNumber uint16,
	targetNetworkId string,
	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
	dockerManager *docker_manager.DockerManager,
) (string, map[string]string, func(), error) {
	vectorContainerConfigProviderObj := createVectorContainerConfigProvider(logsListeningPortNumber)

	logsAggregatorVolumeAttrs, err := objAttrsProvider.ForLogsAggregatorVolume()
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred getting the logs collector volume attributes")
	}
	volumeName := logsAggregatorVolumeAttrs.GetName().GetString()
	volumeLabelStrs := map[string]string{}
	for labelKey, labelValue := range logsAggregatorVolumeAttrs.GetLabels() {
		volumeLabelStrs[labelKey.GetString()] = labelValue.GetString()
	}
	if err := dockerManager.CreateVolume(ctx, volumeName, volumeLabelStrs); err != nil {
		return "", nil, nil, stacktrace.Propagate(
			err,
			"An error occurred creating logs aggregator volume with name '%v' and labels '%+v'",
			volumeName,
			volumeLabelStrs,
		)
	}

	logsAggregatorAttrs, err := objAttrsProvider.ForLogsAggregator()
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred getting the logs aggregator container attributes.")
	}
	containerName := logsAggregatorAttrs.GetName().GetString()
	containerLabelStrs := map[string]string{}
	for labelKey, labelValue := range logsAggregatorAttrs.GetLabels() {
		containerLabelStrs[labelKey.GetString()] = labelValue.GetString()
	}
	createAndStartArgs, err := vectorContainerConfigProviderObj.GetContainerArgs(containerName, containerLabelStrs, volumeName, targetNetworkId)
	if err != nil {
		return "", nil, nil, err
	}

	containerId, _, err := dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred starting the logs aggregator container with these args '%+v'", createAndStartArgs)
	}
	removeContainerFunc := func() {
		removeCtx := context.Background()

		if err := dockerManager.RemoveContainer(removeCtx, containerId); err != nil {
			logrus.Errorf(
				"Launching the logs aggregator server with container ID '%v' didn't complete successfully so we "+
					"tried to remove the container we started, but doing so exited with an error:\n%v",
				containerId,
				err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually remove the logs aggregator server with Docker container ID '%v'!!!!!!", containerId)
		}
	}
	shouldRemoveLogsAggregatorContainer := true
	defer func() {
		if shouldRemoveLogsAggregatorContainer {
			removeContainerFunc()
		}
	}()

	aggregatorIPAddr, err := getLogsAggregatorIPAddr(ctx, containerId, dockerManager)
	if err != nil {
		return "", nil, nil, err
	}
	availabilityChecker := NewVectorAvailabilityChecker(aggregatorIPAddr)
	if err = availabilityChecker.WaitForAvailability(); err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred while waiting for the logs aggregator container to become available")
	}

	shouldRemoveLogsAggregatorContainer = false
	return containerId, containerLabelStrs, removeContainerFunc, nil
}

func getLogsAggregatorIPAddr(ctx context.Context, containerId string, manager *docker_manager.DockerManager) (net.IP, error) {
	privateIpAddrStr, err := manager.GetContainerIP(ctx, consts.NameOfNetworkToStartEngineAndLogServiceContainersIn, containerId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the private IP address of logs aggregator container '%v' in network '%v'", containerId, consts.NameOfNetworkToStartEngineAndLogServiceContainersIn)
	}
	privateIpAddr := net.ParseIP(privateIpAddrStr)
	if privateIpAddr == nil {
		return nil, stacktrace.NewError("Couldn't parse private IP address string '%v' to an IP", privateIpAddrStr)
	}
	return privateIpAddr, nil
}
