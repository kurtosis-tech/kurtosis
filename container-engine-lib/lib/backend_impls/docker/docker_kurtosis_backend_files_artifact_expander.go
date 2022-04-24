package docker

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact_expander"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact_expansion_volume"
	"github.com/kurtosis-tech/stacktrace"
	"net"
)

const (
	// Docker image that will be used to launch the container that will expand the files artifact
	//  into a Docker volume
	dockerImage = "alpine:3.12"

	// TODO Remove this when we switch fully to the data volume
	// Dirpath on the artifact expander container where the enclave data dir (which contains the artifacts)
	//  will be bind-mounted
	enclaveDataBindmountDirpathOnExpanderContainer = "/enclave-data"

	// The location where the enclave data volume will be mounted
	//  on the files artifact expansion container
	enclaveDataVolumeDirpathOnExpanderContainer = "/kurtosis-data"

	expanderContainerSuccessExitCode = 0
)

func (backend *DockerKurtosisBackend) RunFilesArtifactExpander(
	ctx context.Context,
	guid files_artifact_expander.FilesArtifactExpanderGUID,
	enclaveId enclave.EnclaveID,
	filesArtifactExpansionVolumeName files_artifact_expansion_volume.FilesArtifactExpansionVolumeName,
	enclaveDataDirpathOnHostMachine string,
	destVolMntDirpathOnExpander string,
	filesArtifactFilepathRelativeToEnclaveDatadirRoot string,
	ipAddr net.IP, // TODO REMOVE THIS ONCE WE FIX THE STATIC IP PROBLEM!!
)(*files_artifact_expander.FilesArtifactExpander, error){

	enclaveNetwork, err := backend.getEnclaveNetworkByEnclaveId(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave network by enclave ID '%v'", enclaveId)
	}

	enclaveDataVolumeName, err := backend.getEnclaveDataVolumeByEnclaveId(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the enclave data volume for enclave '%v'", enclaveId)
	}

	enclaveObjAttrsProvider, err := backend.objAttrsProvider.ForEnclave(enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Couldn't get an object attribute provider for enclave '%v'", enclaveId)
	}

	containerAttrs, err := enclaveObjAttrsProvider.ForFilesArtifactExpanderContainer(guid)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while trying to get the files artifact expander container attributes for files artifact expander GUID '%v'", guid)
	}
	containerName := containerAttrs.GetName().GetString()
	containerLabels := map[string]string{}
	for labelKey, labelValue := range containerAttrs.GetLabels() {
		containerLabels[labelKey.GetString()] = labelValue.GetString()
	}

	filesArtifactExpansionVolumeNameStr := string(filesArtifactExpansionVolumeName)

	bindMounts := map[string]string{
		enclaveDataDirpathOnHostMachine: enclaveDataBindmountDirpathOnExpanderContainer,
	}

	volumeMounts := map[string]string{
		filesArtifactExpansionVolumeNameStr: destVolMntDirpathOnExpander,
		enclaveDataVolumeName:               enclaveDataVolumeDirpathOnExpanderContainer,
	}

	containerCmd := getExtractionCommand(filesArtifactFilepathRelativeToEnclaveDatadirRoot, destVolMntDirpathOnExpander)

	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		dockerImage,
		containerName,
		enclaveNetwork.GetId(),
	).WithStaticIP(
		ipAddr,
	).WithCmdArgs(
		containerCmd,
	).WithBindMounts(
		bindMounts,
	).WithVolumeMounts(
		volumeMounts,
	).WithLabels(
		containerLabels,
	).Build()
	containerId, _, err := backend.dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the Docker container to expand the file artifact '%v' into the volume '%v'", filesArtifactFilepathRelativeToEnclaveDatadirRoot, filesArtifactExpansionVolumeNameStr)
	}

	exitCode, err := backend.dockerManager.WaitForExit(ctx, containerId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the files artifact expander Docker container '%v' to exit", containerId)
	}
	if exitCode != expanderContainerSuccessExitCode {
		return nil, stacktrace.NewError(
			"The files artifact expander Docker container '%v' exited with non-%v exit code: %v",
			containerId,
			expanderContainerSuccessExitCode,
			exitCode)
	}

	containerStatus := types.ContainerStatus_Exited

	newFilesArtifactExpander, err := getFilesArtifactExpanderObjectFromContainerInfo(containerLabels, containerStatus)

	return newFilesArtifactExpander, nil
}

func (backend *DockerKurtosisBackend) DestroyFilesArtifactExpanders(
	ctx context.Context,
	filters *files_artifact_expander.FilesArtifactExpanderFilters,
) (
	map[files_artifact_expander.FilesArtifactExpanderGUID]bool,
	map[files_artifact_expander.FilesArtifactExpanderGUID]error,
	error,
) {
	successfulExpanderGUIDs := map[files_artifact_expander.FilesArtifactExpanderGUID]bool{}
	erroredExpanderGUIDs  := map[files_artifact_expander.FilesArtifactExpanderGUID]error{}

	matchedExpanders, err := backend.getMatchingFilesArtifactExpanders(ctx, filters)
	if err != nil {
		return nil, nil,  stacktrace.Propagate(err, "An error occurred getting files artifact expanders matching filters '%+v'", filters)
	}

	//TODO execute concurrently to improve perf
	for containerId, expander := range matchedExpanders {
		expanderGuid := expander.GetGUID()
		if err := backend.dockerManager.RemoveContainer(ctx, containerId); err != nil {
			wrappedErr := stacktrace.Propagate(
				err,
				"An error occurred removing container with ID '%v'",
				containerId,
			)
			erroredExpanderGUIDs[expanderGuid] = wrappedErr
			continue
		}
		successfulExpanderGUIDs[expanderGuid] = true
	}

	return successfulExpanderGUIDs, erroredExpanderGUIDs, nil
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================
// Image-specific generator of the command that should be run to extract the artifact at the given filepath
//  to the destination
func getExtractionCommand(artifactFilepath string, destVolMntDirpathOnExpander string) (dockerRunCmd []string) {
	return []string{
		"tar",
		"-xzvf",
		artifactFilepath,
		"-C",
		destVolMntDirpathOnExpander,
	}
}

func (backend *DockerKurtosisBackend) getMatchingFilesArtifactExpanders(
	ctx context.Context,
	filters *files_artifact_expander.FilesArtifactExpanderFilters,
)(map[string]*files_artifact_expander.FilesArtifactExpander, error) {
	searchLabels := map[string]string{
		label_key_consts.AppIDLabelKey.GetString(): label_value_consts.AppIDLabelValue.GetString(),
		label_key_consts.ContainerTypeLabelKey.GetString(): label_value_consts.FilesArtifactExpanderContainerTypeLabelValue.GetString(),
	}
	matchingContainers, err := backend.dockerManager.GetContainersByLabels(ctx, searchLabels, shouldFetchAllContainersWhenRetrievingContainers)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred fetching containers using labels: %+v", searchLabels)
	}

	matchingObjects := map[string]*files_artifact_expander.FilesArtifactExpander{}
	for _, container := range matchingContainers {
		containerId := container.GetId()
		object, err := getFilesArtifactExpanderObjectFromContainerInfo(
			container.GetLabels(),
			container.GetStatus(),
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred converting container '%v' into a files artifact expander object", container.GetId())
		}

		if filters.EnclaveIDs != nil && len(filters.EnclaveIDs) > 0 {
			if _, found := filters.EnclaveIDs[object.GetEnclaveID()]; !found {
				continue
			}
		}

		if filters.GUIDs != nil && len(filters.GUIDs) > 0 {
			if _, found := filters.GUIDs[object.GetGUID()]; !found {
				continue
			}
		}

		if filters.Statuses != nil && len(filters.Statuses) > 0 {
			if _, found := filters.Statuses[object.GetStatus()]; !found {
				continue
			}
		}

		matchingObjects[containerId] = object
	}

	return matchingObjects, nil
}

func getFilesArtifactExpanderObjectFromContainerInfo(
	labels map[string]string,
	containerStatus types.ContainerStatus,
) (*files_artifact_expander.FilesArtifactExpander, error) {

	enclaveIdStr, found := labels[label_key_consts.EnclaveIDLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected the files artifact expander's enclave ID to be found under label '%v' but the label wasn't present", label_key_consts.EnclaveIDLabelKey.GetString())
	}

	guidStr, found := labels[label_key_consts.GUIDLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected to find GUID label key '%v' but none was found", label_key_consts.GUIDLabelKey.GetString())
	}

	isContainerRunning, found := isContainerRunningDeterminer[containerStatus]
	if !found {
		// This should never happen because we enforce completeness in a unit test
		return nil, stacktrace.NewError("No is-running designation found for files artifact expander container status '%v'; this is a bug in Kurtosis!", containerStatus.String())
	}
	var status container_status.ContainerStatus
	if isContainerRunning {
		status = container_status.ContainerStatus_Running
	} else {
		status = container_status.ContainerStatus_Stopped
	}

	newObject := files_artifact_expander.NewFilesArtifactExpander(
		files_artifact_expander.FilesArtifactExpanderGUID(guidStr),
		enclave.EnclaveID(enclaveIdStr),
		status,
	)

	return newObject, nil
}
