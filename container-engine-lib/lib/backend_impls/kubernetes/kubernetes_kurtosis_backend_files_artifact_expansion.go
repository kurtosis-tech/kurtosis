package kubernetes

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact_expansion"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
	"path"
)

const (
	// Docker image that will be used to launch the container that will expand the files artifact
	//  into a Docker volume
	dockerImage = "alpine:3.12"

	// Dirpath on the artifact expander container where the enclave data volume (which contains artifacts)
	//  will be mounted
	enclaveDataVolumeDirpathOnExpanderContainer = "/enclave-data"

	// Dirpath on the artifact expander container where the destination volume containing expanded files will be mounted
	destVolMntDirpathOnExpander = "/dest"

	pvcVolumeClaimReadOnly = false

	expanderContainerSuccessExitCode = 0

	// Based on example on k8s docs https://kubernetes.io/docs/concepts/workloads/controllers/job/
	ttlSecondsAfterFinishedExpanderJob = 100
)

//Create a files artifact exansion volume for user service and file artifact id and runs a file artifact expander
func (backend *KubernetesKurtosisBackend) CreateFilesArtifactExpansion(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceGuid service.ServiceGUID,
	filesArtifactId service.FilesArtifactID,
	filesArtifactFilepathRelativeToEnclaveDatadirRoot string)(*files_artifact_expansion.FilesArtifactExpansion, error) {
	filesArtifactExpansionGUIDStr, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to generate UUID for files artifact expansion.")
	}


	filesArtifactExpansionGUID := files_artifact_expansion.FilesArtifactExpansionGUID(filesArtifactExpansionGUIDStr)
	enclaveObjAttrsProvider := backend.objAttrsProvider.ForEnclave(enclaveId)
	pvcAttrs, err := enclaveObjAttrsProvider.ForFilesArtifactExpansionPersistentVolumeClaim(
		filesArtifactExpansionGUID,
		serviceGuid)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred while trying to get the files artifact expansion " +
				"volume attributes for service with GUID '%v'",
			serviceGuid,
		)
	}
	pvcName := pvcAttrs.GetName().GetString()
	pvcLabels := map[string]string{}
	for labelKey, labelValue := range pvcAttrs.GetLabels() {
		pvcLabels[labelKey.GetString()] = labelValue.GetString()
	}
	enclaveNamespaceName, err := backend.getEnclaveNamespaceName(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave namespace name for enclave with ID '%v'", enclaveId)
	}
	pvc, err := backend.kubernetesManager.CreatePersistentVolumeClaim(
		ctx,
		enclaveNamespaceName,
		pvcName,
		pvcLabels,
		backend.apiContainerModeArgs.filesArtifactExpansionVolumeSizeInMegabytes,
		backend.apiContainerModeArgs.storageClassName)
	shouldDestroyPVC := true
	defer func() {
		if shouldDestroyPVC {
			err := backend.kubernetesManager.RemovePersistentVolumeClaim(
				ctx, enclaveNamespaceName, pvcName)
			if err != nil {
				logrus.Errorf("Failed to destroy persistent volume claim with name '%v' in namespace '%v'. Got error '%v'." +
					"You should manually destroy this persistent volume claim.", pvcName, enclaveNamespaceName, err.Error())
			}
		}
	}()


	artifactFilepath := path.Join(enclaveDataVolumeDirpathOnExpanderContainer, filesArtifactFilepathRelativeToEnclaveDatadirRoot)
	extractionCommand := getExtractionCommand(artifactFilepath, destVolMntDirpathOnExpander)

	jobAttrs, err := enclaveObjAttrsProvider.ForFilesArtifactExpansionJob(filesArtifactExpansionGUID, serviceGUID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while trying to get the files artifact expander container attributes for files artifact expansion GUID '%v'", filesArtifactExpansionGUID)
	}
	jobContainerName := jobAttrs.GetName().GetString()

	job, err := backend.kubernetesManager.CreateJobWithPVCMount(ctx,
		enclaveNamespaceName,
		ttlSecondsAfterFinishedExpanderJob,
		dockerImage,
		extractionCommand,
		jobContainerName,
		pvc.Spec.VolumeName,
		pvc.Name,
		enclaveDataVolumeDirpathOnExpanderContainer,
		pvcVolumeClaimReadOnly,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create files artifact expansion job for expansion '%v'", filesArtifactExpansionGUID)
	}
	shouldDestroyPVC = false
	panic("IMPLEMENT ME")
}

//Destroy files artifact expansion volume and expander using the given filters
func (backend *KubernetesKurtosisBackend) DestroyFilesArtifactExpansions(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters  *files_artifact_expansion.FilesArtifactExpansionFilters,
)(
	successfulFileArtifactExpansionGUIDs map[files_artifact_expansion.FilesArtifactExpansionGUID]bool,
	erroredFileArtifactExpansionGUIDs map[files_artifact_expansion.FilesArtifactExpansionGUID]error,
	resultErr error,
) {
	panic("IMPLEMENT ME")
}

// ==================== PRIVATE HELPERS ===============

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

func (backend *KubernetesKurtosisBackend) runFilesArtifactExpander(
	ctx context.Context,
	filesArtifactExpansionGUID files_artifact_expansion.FilesArtifactExpansionGUID,
	serviceGUID service.ServiceGUID,
	enclaveId enclave.EnclaveID,
	filesArtifactExpansionVolumeName string,
	destVolMntDirpathOnExpander string,
	ipAddr net.IP,
	filesArtifactFilepathRelativeToEnclaveDataVolumeRoot string,
) (expanderContainerId string, resultErr error) {

	namespaceName, err := backend.getEnclaveNamespaceName(ctx, enclaveId)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting namespaceName by enclave ID '%v'", enclaveId)
	}

	enclaveDataVolumeName, err := backend.getEnclaveDataVolumeByEnclaveId(ctx, enclaveId)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the enclave data volume for enclave '%v'", enclaveId)
	}

	enclaveObjAttrsProvider, err := backend.objAttrsProvider.ForEnclave(enclaveId)
	if err != nil {
		return "", stacktrace.Propagate(err, "Couldn't get an object attribute provider for enclave '%v'", enclaveId)
	}

	containerAttrs, err := enclaveObjAttrsProvider.ForFilesArtifactExpansionContainer(filesArtifactExpansionGUID, serviceGUID)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while trying to get the files artifact expander container attributes for files artifact expansion GUID '%v'", filesArtifactExpansionGUID)
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
		return "", stacktrace.Propagate(err, "An error occurred creating the Docker container to expand the file artifact '%v' into the volume '%v'", filesArtifactFilepathRelativeToEnclaveDataVolumeRoot, filesArtifactExpansionVolumeName)
	}
	shouldKillContainer := true
	defer func() {
		if shouldKillContainer {
			containerTeardownErr := backend.dockerManager.KillContainer(ctx, containerId)
			if containerTeardownErr != nil {
				logrus.Errorf("Failed to tear down container ID '%v', you will need to manually remove it!", containerId)
			}
		}
	}()

	exitCode, err := backend.dockerManager.WaitForExit(ctx, containerId)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred waiting for the files artifact expander Docker container '%v' to exit", containerId)
	}
	if exitCode != expanderContainerSuccessExitCode {
		return "", stacktrace.NewError(
			"The files artifact expander Docker container '%v' exited with non-%v exit code: %v",
			containerId,
			expanderContainerSuccessExitCode,
			exitCode)
	}
	shouldKillContainer = false
	return containerId,nil
}

