package docker

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_operation_parallelizer"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact_expander"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact_expansion"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact_expansion_volume"
	"github.com/kurtosis-tech/stacktrace"
	"path"
)

const (
	// Docker image that will be used to launch the container that will expand the files artifact
	//  into a Docker volume
	dockerImage = "alpine:3.12"

	// Dirpath on the artifact expander container where the enclave data volume (which contains artifacts)
	//  will be mounted
	enclaveDataVolumeDirpathOnExpanderContainer = "/enclave-data"

	expanderContainerSuccessExitCode = 0
)

func (backend *DockerKurtosisBackend) runFilesArtifactExpander(
	ctx context.Context,
	filesArtifactExpansion *files_artifact_expansion.FilesArtifactExpansion,
	enclaveId enclave.EnclaveID,
	filesArtifactExpansionVolumeName files_artifact_expansion_volume.FilesArtifactExpansionVolumeName,
	destVolMntDirpathOnExpander string,
	filesArtifactFilepathRelativeToEnclaveDataVolumeRoot string,
)(*files_artifact_expander.FilesArtifactExpander, error){

	enclaveNetwork, err := backend.getEnclaveNetworkByEnclaveId(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave network by enclave ID '%v'", enclaveId)
	}

	freeIpAddrProvider, found := backend.enclaveFreeIpProviders[enclaveId]
	if !found {
		return nil, stacktrace.NewError(
			"Received a request to run a files artifact expander attached to service with GUID '%v' in enclave '%v', but no free IP address provider was " +
				"defined for this enclave; this likely means that the request is being called where it shouldn't " +
				"be (i.e. outside the API container)",
			filesArtifactExpansion.GetServiceGUID(),
			enclaveId,
		)
	}

	enclaveDataVolumeName, err := backend.getEnclaveDataVolumeByEnclaveId(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the enclave data volume for enclave '%v'", enclaveId)
	}

	enclaveObjAttrsProvider, err := backend.objAttrsProvider.ForEnclave(enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Couldn't get an object attribute provider for enclave '%v'", enclaveId)
	}

	containerAttrs, err := enclaveObjAttrsProvider.ForFilesArtifactExpanderContainer(filesArtifactExpansion.GetGUID())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while trying to get the files artifact expander container attributes for files artifact expansion GUID '%v'", filesArtifactExpansion.GetGUID())
	}
	containerName := containerAttrs.GetName().GetString()
	containerLabels := map[string]string{}
	for labelKey, labelValue := range containerAttrs.GetLabels() {
		containerLabels[labelKey.GetString()] = labelValue.GetString()
	}

	filesArtifactExpansionVolumeNameStr := string(filesArtifactExpansionVolumeName)

	volumeMounts := map[string]string{
		filesArtifactExpansionVolumeNameStr: destVolMntDirpathOnExpander,
		enclaveDataVolumeName:               enclaveDataVolumeDirpathOnExpanderContainer,
	}

	artifactFilepath := path.Join(enclaveDataVolumeDirpathOnExpanderContainer, filesArtifactFilepathRelativeToEnclaveDataVolumeRoot)
	containerCmd := getExtractionCommand(artifactFilepath, destVolMntDirpathOnExpander)

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

	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		dockerImage,
		containerName,
		enclaveNetwork.GetId(),
	).WithStaticIP(
		ipAddr,
	).WithCmdArgs(
		containerCmd,
	).WithVolumeMounts(
		volumeMounts,
	).WithLabels(
		containerLabels,
	).Build()
	containerId, _, err := backend.dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the Docker container to expand the file artifact '%v' into the volume '%v'", filesArtifactFilepathRelativeToEnclaveDataVolumeRoot, filesArtifactExpansionVolumeNameStr)
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

	shouldReleaseIp = true
	return newFilesArtifactExpander, nil
}

func (backend *DockerKurtosisBackend) destroyFilesArtifactExpanders(
	ctx context.Context,
	filters *files_artifact_expander.FilesArtifactExpanderFilters,
) (
	map[files_artifact_expander.FilesArtifactExpanderGUID]bool,
	map[files_artifact_expander.FilesArtifactExpanderGUID]error,
	error,
) {
	matchedExpanders, err := backend.getMatchingFilesArtifactExpanders(ctx, filters)
	if err != nil {
		return nil, nil,  stacktrace.Propagate(err, "An error occurred getting files artifact expanders matching filters '%+v'", filters)
	}

	// TODO PLEAAASE GO GENERICS... but we can't use 1.18 yet because it'll break all Kurtosis clients :(
	matchingUncastedExpandersByContainerId := map[string]interface{}{}
	for containerId, expanderObj := range matchedExpanders {
		matchingUncastedExpandersByContainerId[containerId] = interface{}(expanderObj)
	}

	var removeExpanderOperation docker_operation_parallelizer.DockerOperation = func(
		ctx context.Context,
		dockerManager *docker_manager.DockerManager,
		dockerObjectId string,
	) error {
		if err := dockerManager.RemoveContainer(ctx, dockerObjectId); err != nil {
			return stacktrace.Propagate(err, "An error occurred removing files artifact expander container with ID '%v'", dockerObjectId)
		}
		return nil
	}

	successfulExpanderGUIDStrs, erroredExpanderGUIDStrs, err := docker_operation_parallelizer.RunDockerOperationInParallelForKurtosisObjects(
		ctx,
		matchingUncastedExpandersByContainerId,
		backend.dockerManager,
		extractExpanderGUIDFromExpanderObj,
		removeExpanderOperation,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred removing files artifact expander containers matching filters '%+v'", filters)
	}

	successfulExpanderGUIDs := map[files_artifact_expander.FilesArtifactExpanderGUID]bool{}
	for expanderGuidStr := range successfulExpanderGUIDStrs {
		successfulExpanderGUIDs[files_artifact_expander.FilesArtifactExpanderGUID(expanderGuidStr)] = true
	}
	erroredExpanderGUIDs := map[files_artifact_expander.FilesArtifactExpanderGUID]error{}
	for expanderGuidStr, removalErr := range erroredExpanderGUIDStrs {
		erroredExpanderGUIDs[files_artifact_expander.FilesArtifactExpanderGUID(expanderGuidStr)] = removalErr
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
		label_key_consts.AppIDDockerLabelKey.GetString():         label_value_consts.AppIDDockerLabelValue.GetString(),
		label_key_consts.ContainerTypeDockerLabelKey.GetString(): label_value_consts.FilesArtifactExpanderContainerTypeDockerLabelValue.GetString(),
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

	enclaveIdStr, found := labels[label_key_consts.EnclaveIDDockerLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected the files artifact expander's enclave ID to be found under label '%v' but the label wasn't present", label_key_consts.EnclaveIDDockerLabelKey.GetString())
	}

	guidStr, found := labels[label_key_consts.GUIDDockerLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected to find GUID label key '%v' but none was found", label_key_consts.GUIDDockerLabelKey.GetString())
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

func extractExpanderGUIDFromExpanderObj(uncastedObj interface{}) (string, error) {
	castedObj, ok := uncastedObj.(*files_artifact_expander.FilesArtifactExpander)
	if !ok {
		return "", stacktrace.NewError("An error occurred downcasting the files artifact expander object")
	}
	return string(castedObj.GetGUID()), nil
}