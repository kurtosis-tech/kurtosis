package docker

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
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

	containerAttrs, err := enclaveObjAttrsProvider.ForFilesArtifactExpanderContainer(filesArtifactExpansion.GetGUID(), filesArtifactExpansion.GetServiceGUID())
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