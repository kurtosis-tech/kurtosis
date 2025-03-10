package logs_aggregator_functions

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/docker/docker/api/types/volume"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_port_spec_serializer"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	shouldShowStoppedLogsAggregatorContainers = true
	curlContainerSuccessExitCode              = 0
	successHealthCheckStatusCode              = 200
	healthcheckImage                          = "badouralix/curl-jq"
	healthcheckContainerNamePrefix            = "logs-aggregator-healthcheck"
	healthcheckCmdMaxRetries                  = 30
	healthcheckCmdDelayInRetries              = 200 * time.Millisecond
	sleepSeconds                              = 1800
)

func getLogsAggregatorPrivatePorts(containerLabels map[string]string) (*port_spec.PortSpec, error) {
	serializedPortSpecs, found := containerLabels[docker_label_key.PortSpecsDockerLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected to find port specs label '%v' but none was found", docker_label_key.PortSpecsDockerLabelKey.GetString())
	}

	portSpecs, err := docker_port_spec_serializer.DeserializePortSpecs(serializedPortSpecs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Couldn't deserialize port spec string '%v'", serializedPortSpecs)
	}

	httpPortSpec, foundHttpPort := portSpecs[logsAggregatorHttpPortId]
	if !foundHttpPort {
		return nil, stacktrace.NewError("No logs aggregator HTTP port with ID '%v' found in the logs aggregator port specs", logsAggregatorHttpPortId)
	}

	return httpPortSpec, nil
}

// Returns nil [LogsAggregator] object if no container is found
func getLogsAggregatorObjectAndContainerId(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
) (*logs_aggregator.LogsAggregator, string, error) {
	logsAggregatorContainer, found, err := getLogsAggregatorContainer(ctx, dockerManager)
	if err != nil {
		return nil, "", stacktrace.Propagate(err, "An error occurred getting all logs aggregator containers")
	}
	if !found {
		return nil, "", nil
	}

	logsAggregatorContainerID := logsAggregatorContainer.GetId()

	logsAggregatorObject, err := getLogsAggregatorObjectFromContainerInfo(
		ctx,
		logsAggregatorContainerID,
		logsAggregatorContainer.GetLabels(),
		logsAggregatorContainer.GetStatus(),
		dockerManager,
	)
	if err != nil {
		return nil, "", stacktrace.Propagate(err, "An error occurred getting the logs Aggregator object using container ID '%v', labels '%+v' and the status '%v'", logsAggregatorContainer.GetId(), logsAggregatorContainer.GetLabels(), logsAggregatorContainer.GetStatus())
	}

	return logsAggregatorObject, logsAggregatorContainerID, nil
}

// Returns nil [Container] object and false if no logs aggregator container is found
func getLogsAggregatorContainer(ctx context.Context, dockerManager *docker_manager.DockerManager) (*types.Container, bool, error) {
	logsAggregatorContainerSearchLabels := map[string]string{
		docker_label_key.AppIDDockerLabelKey.GetString():         label_value_consts.AppIDDockerLabelValue.GetString(),
		docker_label_key.ContainerTypeDockerLabelKey.GetString(): label_value_consts.LogsAggregatorTypeDockerLabelValue.GetString(),
	}

	matchingLogsAggregatorContainers, err := dockerManager.GetContainersByLabels(ctx, logsAggregatorContainerSearchLabels, shouldShowStoppedLogsAggregatorContainers)
	if err != nil {
		return nil, false, stacktrace.Propagate(err, "An error occurred fetching the logs aggregator container using labels: %+v", logsAggregatorContainerSearchLabels)
	}

	if len(matchingLogsAggregatorContainers) == 0 {
		return nil, false, nil
	}
	if len(matchingLogsAggregatorContainers) > 1 {
		return nil, false, stacktrace.NewError("Found more than one logs aggregator Docker container'; this is a bug in Kurtosis")
	}
	return matchingLogsAggregatorContainers[0], true, nil
}

func getLogsAggregatorObjectFromContainerInfo(
	ctx context.Context,
	containerId string,
	labels map[string]string,
	containerStatus types.ContainerStatus,
	dockerManager *docker_manager.DockerManager,
) (*logs_aggregator.LogsAggregator, error) {
	var privateIpAddr net.IP

	isContainerRunning, found := consts.IsContainerRunningDeterminer[containerStatus]
	if !found {
		// This should never happen because we enforce completeness in a unit test
		return nil, stacktrace.NewError("No is-running designation found for logs aggregator container status '%v'; this is a bug in Kurtosis!", containerStatus.String())
	}

	privateHttpPortSpec, err := getLogsAggregatorPrivatePorts(labels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the logs collector container's private port specs from container '%v' with labels: %+v", containerId, labels)
	}

	var logsAggregatorStatus container.ContainerStatus
	if isContainerRunning {
		logsAggregatorStatus = container.ContainerStatus_Running

		privateIpAddrStr, err := dockerManager.GetContainerIP(ctx, consts.NameOfNetworkToStartEngineAndLogServiceContainersIn, containerId)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting the private IP address of container '%v' in network '%v'", containerId, consts.NameOfNetworkToStartEngineAndLogServiceContainersIn)
		}
		privateIpAddr = net.ParseIP(privateIpAddrStr)
		if privateIpAddr == nil {
			return nil, stacktrace.NewError("Couldn't parse private IP address string '%v' to an IP", privateIpAddrStr)
		}
	} else {
		logsAggregatorStatus = container.ContainerStatus_Stopped
	}

	logsAggregatorObj := logs_aggregator.NewLogsAggregator(
		logsAggregatorStatus,
		privateIpAddr,
		defaultLogsListeningPortNum,
		privateHttpPortSpec,
	)

	return logsAggregatorObj, nil
}

// if nothing is found we return empty volume name
func getLogsAggregatorVolumeName(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
) (string, error) {

	var volumes []*volume.Volume

	searchLabels := map[string]string{
		docker_label_key.AppIDDockerLabelKey.GetString():      label_value_consts.AppIDDockerLabelValue.GetString(),
		docker_label_key.VolumeTypeDockerLabelKey.GetString(): label_value_consts.LogsAggregatorVolumeTypeDockerLabelValue.GetString(),
	}

	volumes, err := dockerManager.GetVolumesByLabels(ctx, searchLabels)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the volumes for logs aggregator by labels '%+v'", searchLabels)
	}

	if len(volumes) == 0 {
		return "", nil
	}

	if len(volumes) > 1 {
		return "", stacktrace.NewError("Attempted to get logs collector volume name for logs aggregator but got more than one matches")
	}

	return volumes[0].Name, nil
}

func waitForLogsAggregatorAvailability(
	ctx context.Context,
	healthcheckIpAddr net.IP,
	healthCheckEndpoint string,
	healthCheckPortNum uint16,
	targetNetworkId string,
	dockerManager *docker_manager.DockerManager,
) error {
	entrypointArgs := []string{
		"/bin/sh",
		"-c",
		fmt.Sprintf("sleep %v", sleepSeconds),
	}

	uuid, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred generating a UUID for the configurator container name")
	}

	containerName := fmt.Sprintf("%s-%s", healthcheckContainerNamePrefix, uuid)

	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		healthcheckImage,
		containerName,
		targetNetworkId,
	).WithEntrypointArgs(
		entrypointArgs,
	).Build()

	containerId, _, err := dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred starting the logs aggregator healthcheck container with these args '%+v'", createAndStartArgs)
	}
	//The killing step has to be executed always in the success and also in the failed case
	defer func() {
		if err = dockerManager.RemoveContainer(context.Background(), containerId); err != nil {
			logrus.Errorf(
				"Launching the logs aggregator healthcheck container with container ID '%v' didn't complete successfully so we "+
					"tried to remove the container we started, but doing so exited with an error:\n%v",
				containerId,
				err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually remove the container with ID '%v'!!!!!!", containerId)
		}
	}()

	healthcheckUrl := fmt.Sprintf("http://%v:%v%v", healthcheckIpAddr, healthCheckPortNum, healthCheckEndpoint)
	execCmd := []string{
		"/bin/sh",
		"-c",
		fmt.Sprintf("curl -s -o /dev/null -w \"%%{http_code}\" %v", healthcheckUrl),
	}

	for i := uint(0); i < healthcheckCmdMaxRetries; i++ {
		outputBuffer := &bytes.Buffer{}
		exitCode, err := dockerManager.RunUserServiceExecCommands(ctx, containerId, "", execCmd, outputBuffer)
		if err == nil {
			healthCheckStatusCode, err := strconv.Atoi(outputBuffer.String())
			if err != nil {
				return stacktrace.Propagate(err, "Expected to be able to convert '%v', output from '%v' to an int but was unable to.", outputBuffer.String(), execCmd)
			}

			logrus.Debugf("Logs aggregator healthcheck command '%v' returned health status code: %v", execCmd, healthCheckStatusCode)
			if healthCheckStatusCode == successHealthCheckStatusCode {
				return nil
			}

			logrus.Debugf(
				"Logs aggregator healthcheck command command '%v' returned without a Docker error, but exited with non-%v exit code '%v' and logs:\n%v",
				execCmd,
				curlContainerSuccessExitCode,
				exitCode,
				outputBuffer.String(),
			)
		} else {
			logrus.Debugf(
				"Logs aggregator healthcheck command '%v' experienced a Docker error:\n%v",
				execCmd,
				err,
			)
		}

		// Tiny optimization to not sleep if we're not going to run the loop again
		if i < healthcheckCmdMaxRetries-1 {
			time.Sleep(healthcheckCmdDelayInRetries)
		}
	}

	return stacktrace.NewError(
		"Logs aggregator healthcheck didn't return success (as measured by the command '%v') even after retrying %v times with %v between retries",
		execCmd,
		healthcheckCmdMaxRetries,
		healthcheckCmdDelayInRetries,
	)
}
