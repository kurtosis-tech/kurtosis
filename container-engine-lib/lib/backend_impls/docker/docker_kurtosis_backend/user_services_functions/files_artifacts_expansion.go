package user_service_functions

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/kurtosis-tech/free-ip-addr-tracker-lib/lib"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/concurrent_writer"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"strings"
)

const (
	shouldFollowContainerLogsWhenExpanderHasError = false

	expanderContainerSuccessExitCode = 0
)

// Functions required to do files artifacts expansion
func doFilesArtifactExpansionAndGetUserServiceVolumes(
	ctx context.Context,
	serviceGuid service.ServiceGUID,
	objAttrsProvider object_attributes_provider.DockerEnclaveObjectAttributesProvider,
	freeIpAddrProvider *lib.FreeIpAddrTracker,
	enclaveNetworkId string,
	expanderImage string,
	expanderEnvVars map[string]string,
	expanderMountpointsToServiceMountpoints map[string]string,
	dockerManager *docker_manager.DockerManager,
) (map[string]string, error) {
	requestedExpanderMountpoints := map[string]bool{}
	for expanderMountpoint := range expanderMountpointsToServiceMountpoints {
		requestedExpanderMountpoints[expanderMountpoint] = true
	}
	expanderMountpointsToVolumeNames, err := createFilesArtifactsExpansionVolumes(
		ctx,
		serviceGuid,
		objAttrsProvider,
		requestedExpanderMountpoints,
		dockerManager,
	)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Couldn't create files artifact expansion volumes for requested expander mounpoints: %+v",
			requestedExpanderMountpoints,
		)
	}
	shouldDeleteVolumes := true
	defer func() {
		if shouldDeleteVolumes {
			for _, volumeName := range expanderMountpointsToVolumeNames {
				// Use background context so we delete these even if input context was cancelled
				if err := dockerManager.RemoveVolume(context.Background(), volumeName); err != nil {
					logrus.Errorf("Running the expansion failed so we tried to delete files artifact expansion volume '%v' that we created, but doing so threw an error:\n%v", volumeName, err)
					logrus.Errorf("You'll need to delete volume '%v' manually!", volumeName)
				}
			}
		}
	}()

	if err := runFilesArtifactsExpander(
		ctx,
		serviceGuid,
		objAttrsProvider,
		freeIpAddrProvider,
		expanderImage,
		expanderEnvVars,
		enclaveNetworkId,
		expanderMountpointsToVolumeNames,
		dockerManager,
	); err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred running files artifacts expander for service '%v'",
			serviceGuid,
		)
	}

	userServiceVolumeMounts := map[string]string{}
	for expanderMountpoint, userServiceMountpoint := range expanderMountpointsToServiceMountpoints {
		volumeName, found := expanderMountpointsToVolumeNames[expanderMountpoint]
		if !found {
			return nil, stacktrace.NewError(
				"Found expander mountpoint '%v' for which no expansion volume was created; this should never happen "+
					"and is a bug in Kurtosis",
				expanderMountpoint,
			)
		}
		userServiceVolumeMounts[volumeName] = userServiceMountpoint
	}

	shouldDeleteVolumes = false
	return userServiceVolumeMounts, nil
}

// Runs a single expander container which expands one or more files artifacts into multiple volumes
// NOTE: It is the caller's responsibility to handle the volumes that get returned
func runFilesArtifactsExpander(
	ctx context.Context,
	serviceGuid service.ServiceGUID,
	objAttrProvider object_attributes_provider.DockerEnclaveObjectAttributesProvider,
	freeIpAddrProvider *lib.FreeIpAddrTracker,
	image string,
	envVars map[string]string,
	enclaveNetworkId string,
	mountpointsToVolumeNames map[string]string,
	dockerManager *docker_manager.DockerManager,
) error {
	containerAttrs, err := objAttrProvider.ForFilesArtifactsExpanderContainer(serviceGuid)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while trying to get the files artifact expander container attributes for service '%v'", serviceGuid)
	}
	containerName := containerAttrs.GetName().GetString()
	containerLabels := map[string]string{}
	for labelKey, labelValue := range containerAttrs.GetLabels() {
		containerLabels[labelKey.GetString()] = labelValue.GetString()
	}

	volumeMounts := map[string]string{}
	for mountpointOnExpander, volumeName := range mountpointsToVolumeNames {
		volumeMounts[volumeName] = mountpointOnExpander
	}

	ipAddr, err := freeIpAddrProvider.GetFreeIpAddr()
	if err != nil {
		return stacktrace.Propagate(err, "Couldn't get a free IP to give the expander container '%v'", containerName)
	}
	defer freeIpAddrProvider.ReleaseIpAddr(ipAddr)

	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		image,
		containerName,
		enclaveNetworkId,
	).WithStaticIP(
		ipAddr,
	).WithEnvironmentVariables(
		envVars,
	).WithVolumeMounts(
		volumeMounts,
	).WithLabels(
		containerLabels,
	).Build()
	containerId, _, err := dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred creating files artifacts expander container '%v' for service '%v'",
			containerName,
			serviceGuid,
		)
	}
	defer func() {
		// We destroy the expander container, rather than leaving it around, so that we clean up the resource we created
		// in this function (meaning the caller doesn't have to do it)
		// We can do this because if an error occurs, we'll capture the logs of the container in the error we return
		// to the user
		if destroyContainerErr := dockerManager.RemoveContainer(ctx, containerId); destroyContainerErr != nil {
			logrus.Errorf(
				"We tried to remove the expander container '%v' with ID '%v' that we started, but doing so threw an error:\n%v",
				containerName,
				containerId,
				destroyContainerErr,
			)
			logrus.Errorf("ACTION REQUIRED: You'll need to remove files artifacts expander container '%v' manually", containerName)
		}
	}()

	exitCode, err := dockerManager.WaitForExit(ctx, containerId)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred waiting for files artifacts expander container '%v' to exit",
			containerName,
		)
	}
	if exitCode != expanderContainerSuccessExitCode {
		containerLogsBlockStr, err := getFilesArtifactsExpanderContainerLogsBlockStr(
			ctx,
			containerId,
			dockerManager,
		)
		if err != nil {
			return stacktrace.NewError(
				"Files artifacts expander container '%v' for service '%v' finished with non-%v exit code '%v' so we tried "+
					"to get the logs, but doing so failed with an error:\n%v",
				containerName,
				serviceGuid,
				expanderContainerSuccessExitCode,
				exitCode,
				err,
			)
		}
		return stacktrace.NewError(
			"Files artifacts expander container '%v' for service '%v' finished with non-%v exit code '%v' and logs:\n%v",
			containerName,
			serviceGuid,
			expanderContainerSuccessExitCode,
			exitCode,
			containerLogsBlockStr,
		)
	}

	return nil
}

// This seems like a lot of effort to go through to get the logs of a failed container, but easily seeing the reason an expander
// container has failed has proven to be very useful
func getFilesArtifactsExpanderContainerLogsBlockStr(
	ctx context.Context,
	containerId string,
	dockerManager *docker_manager.DockerManager,
) (string, error) {
	containerLogsReadCloser, err := dockerManager.GetContainerLogs(ctx, containerId, shouldFollowContainerLogsWhenExpanderHasError)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the logs for expander container with ID '%v'", containerId)
	}
	defer containerLogsReadCloser.Close()

	buffer := &bytes.Buffer{}
	concurrentBuffer := concurrent_writer.NewConcurrentWriter(buffer)

	// TODO Push this down into GetContainerLogs!!!! This code actually has a bug where it won't work if the container is a TTY
	//  container; the proper checking logic can be seen in the 'enclave dump' functions but should all be abstracted by GetContainerLogs
	//  The only reason I'm not doing it right now is because we have the huge ETH deadline tomorrow and I don't have time for any
	//  nice-to-have refactors (ktoday, 2022-05-22)
	if _, err := stdcopy.StdCopy(concurrentBuffer, concurrentBuffer, containerLogsReadCloser); err != nil {
		return "", stacktrace.Propagate(
			err,
			"An error occurred copying logs to memory for files artifact expander container '%v'",
			containerId,
		)
	}

	wrappedContainerLogsStrBuilder := strings.Builder{}
	wrappedContainerLogsStrBuilder.WriteString(fmt.Sprintf(
		">>>>>>>>>>>>>>>>>>>>>>>>>>>>>> Logs for container '%v' <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<\n",
		containerId,
	))
	wrappedContainerLogsStrBuilder.WriteString(buffer.String())
	wrappedContainerLogsStrBuilder.WriteString("\n")
	wrappedContainerLogsStrBuilder.WriteString(fmt.Sprintf(
		">>>>>>>>>>>>>>>>>>>>>>>>>>>> End logs for container '%v' <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<",
		containerId,
	))

	return wrappedContainerLogsStrBuilder.String(), nil
}

// Takes in a list of mountpoints on the expander container that the expander container wants populated with volumes,
// creates one volume per mountpoint location, and returns the volume_name -> mountpoint map that the container
// can use when starting
func createFilesArtifactsExpansionVolumes(
	ctx context.Context,
	serviceGuid service.ServiceGUID,
	enclaveObjAttrsProvider object_attributes_provider.DockerEnclaveObjectAttributesProvider,
	allMountpointsExpanderWants map[string]bool,
	dockerManager *docker_manager.DockerManager,
) (map[string]string, error) {
	shouldDeleteVolumes := true
	result := map[string]string{}
	for mountpointExpanderWants := range allMountpointsExpanderWants {
		volumeAttrs, err := enclaveObjAttrsProvider.ForSingleFilesArtifactExpansionVolume(serviceGuid)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating files artifact expansion volume for service '%v'", serviceGuid)
		}
		volumeNameStr := volumeAttrs.GetName().GetString()
		volumeLabelsStrs := map[string]string{}
		for key, value := range volumeAttrs.GetLabels() {
			volumeLabelsStrs[key.GetString()] = value.GetString()
		}
		if err := dockerManager.CreateVolume(
			ctx,
			volumeAttrs.GetName().GetString(),
			volumeLabelsStrs,
		); err != nil {
			return nil, stacktrace.Propagate(
				err,
				"An error occurred creating files artifact expansion volume for service '%v' that's intended to be mounted "+
					"on the expander container at path '%v'",
				serviceGuid,
				mountpointExpanderWants,
			)
		}
		//goland:noinspection GoDeferInLoop
		defer func() {
			if shouldDeleteVolumes {
				// Background context so we still run this even if the input context was cancelled
				if err := dockerManager.RemoveVolume(context.Background(), volumeNameStr); err != nil {
					logrus.Warnf(
						"Creating files artifact expansion volumes didn't complete successfully so we tried to delete volume '%v' that we created, but doing so threw an error:\n%v",
						volumeNameStr,
						err,
					)
					logrus.Warnf("You'll need to clean up volume '%v' manually!", volumeNameStr)
				}
			}
		}()

		result[mountpointExpanderWants] = volumeNameStr
	}
	shouldDeleteVolumes = false
	return result, nil
}
