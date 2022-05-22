package kubernetes

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact_expansion"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	"path"
	"time"
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

	isPersistentVolumeClaimReadOnly = false

	jobStatusPollerInterval = time.Millisecond * 100

	jobStatusPollerTimeout = time.Second * 10

	// Based on example on k8s docs https://kubernetes.io/docs/concepts/workloads/controllers/job/
	ttlSecondsAfterFinishedExpanderJob = 100

	filesArtifactExpansionContainerName = "files-artifact-expansion-container"
)

//Create a files artifact exansion volume for user service and file artifact id and runs a file artifact expander
func (backend *KubernetesKurtosisBackend) CreateFilesArtifactExpansion(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceGuid service.ServiceGUID,
	filesArtifactFilepathRelativeToEnclaveDatadirRoot string,
)(*files_artifact_expansion.FilesArtifactExpansion, error) {
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
	pvcLabelStrs := map[string]string{}
	for labelKey, labelValue := range pvcAttrs.GetLabels() {
		pvcLabelStrs[labelKey.GetString()] = labelValue.GetString()
	}
	enclaveNamespaceName, err := backend.getEnclaveNamespaceName(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave namespace name for enclave with ID '%v'", enclaveId)
	}
	pvc, err := backend.kubernetesManager.CreatePersistentVolumeClaim(
		ctx,
		enclaveNamespaceName,
		pvcName,
		pvcLabelStrs,
		backend.apiContainerModeArgs.filesArtifactExpansionVolumeSizeInMegabytes,
		backend.apiContainerModeArgs.storageClassName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a persistent volume claim in namespace '%v' for files artifact expansion '%v'", enclaveNamespaceName, filesArtifactExpansionGUID)
	}
	shouldDestroyPVC := true
	defer func() {
		if shouldDestroyPVC {
			err := backend.kubernetesManager.RemovePersistentVolumeClaim(
				ctx, pvc)
			if err != nil {
				logrus.Errorf("Failed to destroy persistent volume claim with name '%v' in namespace '%v'. Got error '%v'." +
					"You should manually destroy this persistent volume claim.", pvcName, enclaveNamespaceName, err.Error())
			}
		}
	}()

	artifactFilepath := path.Join(enclaveDataVolumeDirpathOnExpanderContainer, filesArtifactFilepathRelativeToEnclaveDatadirRoot)
	extractionCommand := getExtractionCommand(artifactFilepath, destVolMntDirpathOnExpander)

	volume := apiv1.Volume{
		Name:         pvc.Spec.VolumeName,
		VolumeSource: apiv1.VolumeSource{
			PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
				ClaimName: pvc.GetName(),
				ReadOnly:  isPersistentVolumeClaimReadOnly,
			},
		},
	}

	container := apiv1.Container{
		Name:                     filesArtifactExpansionContainerName,
		Image:                    dockerImage,
		Command:                  extractionCommand,
		VolumeMounts:             []apiv1.VolumeMount{
			{
				Name:             pvc.Spec.VolumeName,
				MountPath:        enclaveDataVolumeDirpathOnExpanderContainer,
			},
		},
	}

	jobAttrs, err := enclaveObjAttrsProvider.ForFilesArtifactExpansionJob(filesArtifactExpansionGUID, serviceGuid)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while trying to get the files artifact expander container attributes for files artifact expansion GUID '%v'", filesArtifactExpansionGUID)
	}
	job, err := backend.kubernetesManager.CreateJobWithContainerAndVolume(ctx,
		enclaveNamespaceName,
		jobAttrs.GetName(),
		jobAttrs.GetLabels(),
		jobAttrs.GetAnnotations(),
		[]apiv1.Container{container},
		[]apiv1.Volume{volume},
		ttlSecondsAfterFinishedExpanderJob,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create files artifact expansion job for expansion '%v'", filesArtifactExpansionGUID)
	}
	jobHasNotSucceeded := true
	defer func(){
		if jobHasNotSucceeded {
			// We delete instead of kill/stop because Kubernetes doesn't have the concept of keeping around stopped jobs
			// https://stackoverflow.com/a/52608258
			deleteJobError := backend.kubernetesManager.DeleteJob(ctx, enclaveNamespaceName, job)
			if deleteJobError != nil {
				logrus.Errorf("Failed to delete job '%v' in namespace '%v', which did not succeed. Error in deletion: '%v'",
					enclaveNamespaceName, job.Name, deleteJobError.Error())
			}
		}
	}()
	shouldKeepPollingJob := true
	jobFinishedPoller := time.Tick(jobStatusPollerInterval)
	jobSucceededTimeout := time.After(jobStatusPollerTimeout)
	for shouldKeepPollingJob {
		select {
		case <- jobSucceededTimeout:
			return nil, stacktrace.NewError("Timed out after '%v' seconds waiting for job '%v' for files artifact expansion '%v' to complete.", jobStatusPollerTimeout.Seconds(), job.Name, filesArtifactExpansionGUID)
		case <-jobFinishedPoller:
			hasJobCompleted, hasJobSucceededInPoll, err := backend.kubernetesManager.GetJobCompletionAndSuccessFlags(ctx, enclaveNamespaceName, job.Name)
			if err != nil {
				return nil, stacktrace.Propagate(err, "Failed to get status for job '%v' for files artifact expansion '%v'",
					job.Name, filesArtifactExpansionGUID)
			}
			if hasJobCompleted {
				shouldKeepPollingJob = false
				jobHasNotSucceeded = !hasJobSucceededInPoll
			}
		}
	}
	if jobHasNotSucceeded {
		return nil, stacktrace.NewError("Job '%v' for files artifact expansion '%v' did not succeed.", job.Name, filesArtifactExpansionGUID)
	}
	jobHasNotSucceeded = false
	shouldDestroyPVC = false
	filesArtifactExpansion := files_artifact_expansion.NewFilesArtifactExpansion(filesArtifactExpansionGUID, serviceGuid)
	return filesArtifactExpansion, nil
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

