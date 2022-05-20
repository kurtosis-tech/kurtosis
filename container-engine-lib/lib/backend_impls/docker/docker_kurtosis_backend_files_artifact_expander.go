package docker

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact_expansion"
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
	filesArtifactExpansionVolumeName string,
	destVolMntDirpathOnExpander string,
	filesArtifactFilepathRelativeToEnclaveDataVolumeRoot string,
) error {

	enclaveNetwork, err := backend.getEnclaveNetworkByEnclaveId(ctx, enclaveId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting enclave network by enclave ID '%v'", enclaveId)
	}

	freeIpAddrProvider, found := backend.enclaveFreeIpProviders[enclaveId]
	if !found {
		return stacktrace.NewError(
			"Received a request to run a files artifact expander attached to service with GUID '%v' in enclave '%v', but no free IP address provider was " +
				"defined for this enclave; this likely means that the request is being called where it shouldn't " +
				"be (i.e. outside the API container)",
			filesArtifactExpansion.GetServiceGUID(),
			enclaveId,
		)
	}

	enclaveDataVolumeName, err := backend.getEnclaveDataVolumeByEnclaveId(ctx, enclaveId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave data volume for enclave '%v'", enclaveId)
	}

	enclaveObjAttrsProvider, err := backend.objAttrsProvider.ForEnclave(enclaveId)
	if err != nil {
		return stacktrace.Propagate(err, "Couldn't get an object attribute provider for enclave '%v'", enclaveId)
	}

	containerAttrs, err := enclaveObjAttrsProvider.ForFilesArtifactExpanderContainer(filesArtifactExpansion.GetGUID(), filesArtifactExpansion.GetServiceGUID())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while trying to get the files artifact expander container attributes for files artifact expansion GUID '%v'", filesArtifactExpansion.GetGUID())
	}
	containerName := containerAttrs.GetName().GetString()
	containerLabels := map[string]string{}
	for labelKey, labelValue := range containerAttrs.GetLabels() {
		containerLabels[labelKey.GetString()] = labelValue.GetString()
	}

	volumeMounts := map[string]string{
		filesArtifactExpansionVolumeName: destVolMntDirpathOnExpander,
		enclaveDataVolumeName:               enclaveDataVolumeDirpathOnExpanderContainer,
	}

	artifactFilepath := path.Join(enclaveDataVolumeDirpathOnExpanderContainer, filesArtifactFilepathRelativeToEnclaveDataVolumeRoot)
	containerCmd := getExtractionCommand(artifactFilepath, destVolMntDirpathOnExpander)

	ipAddr, err := freeIpAddrProvider.GetFreeIpAddr()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting a free IP address")
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
		return stacktrace.Propagate(err, "An error occurred creating the Docker container to expand the file artifact '%v' into the volume '%v'", filesArtifactFilepathRelativeToEnclaveDataVolumeRoot, filesArtifactExpansionVolumeName)
	}

	exitCode, err := backend.dockerManager.WaitForExit(ctx, containerId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred waiting for the files artifact expander Docker container '%v' to exit", containerId)
	}
	if exitCode != expanderContainerSuccessExitCode {
		return stacktrace.NewError(
			"The files artifact expander Docker container '%v' exited with non-%v exit code: %v",
			containerId,
			expanderContainerSuccessExitCode,
			exitCode)
	}
	shouldReleaseIp = true
	return nil
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