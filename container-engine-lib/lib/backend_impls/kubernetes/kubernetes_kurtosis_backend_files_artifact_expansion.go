package kubernetes

/*
const (
	// Docker image that will be used to launch the container that will expand the files artifact
	//  into a Docker volume
	dockerImage = "alpine:3.12"

	// Dirpath on the artifact expander container where the enclave data volume (which contains artifacts)
	//  will be mounted
	enclaveDataVolumeDirpathOnExpanderContainer = "/enclave-data"

	// Dirpath on the artifact expander container where the destination volume containing expanded files will be mounted
	destinationVolumeMountDirpathOnExpanderContainer = "/dest"

	isEnclaveDataPersistentVolumeClaimReadOnly = true
	isDestinationPersistentVolumeClaimReadOnly = false

	jobStatusPollerInterval = time.Millisecond * 100

	jobStatusPollerTimeout = time.Second * 10

	// Based on example on k8s docs https://kubernetes.io/docs/concepts/workloads/controllers/job/
	ttlSecondsAfterFinishedExpanderJob = 100

	filesArtifactExpansionContainerName = "files-artifact-expansion-container"

	// We don't want any retries, under the principle of "less magic"
	numTarExpansionRetries = 0

	shouldFollowContainerLogsWhenArtifactExpansionJobFails = false
	shouldAddTimestampsToContainerLogsWhenArtifactExpansionJobFails = true

	// Kubernetes will label all Job pods with this label
	jobNamePodLabel = "job-name"
)

//Create a files artifact exansion volume for user service and file artifact id and runs a file artifact expander
func (backend *KubernetesKurtosisBackend) CreateFilesArtifactExpansion(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceGuid service.ServiceGUID,
	filesArtifactFilepathRelativeToEnclaveDatadirRoot string,
)(*files_artifact_expansion.FilesArtifactExpansion, error) {
	enclaveNamespaceName, err := backend.getEnclaveNamespaceName(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave namespace name for enclave with ID '%v'", enclaveId)
	}

	filesArtifactExpansionGUIDStr, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to generate UUID for files artifact expansion.")
	}

	enclaveDataPvc, err := backend.getEnclaveDataPersistentVolumeClaim(ctx, enclaveNamespaceName, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the enclave data volume for enclave '%v'", enclaveId)
	}

	filesArtifactExpansionGUID := files_artifact_expansion.FilesArtifactExpansionGUID(filesArtifactExpansionGUIDStr)
	enclaveObjAttrsProvider := backend.objAttrsProvider.ForEnclave(enclaveId)
	destinationPvcAttrs, err := enclaveObjAttrsProvider.ForFilesArtifactExpansionPersistentVolumeClaim(
		filesArtifactExpansionGUID,
		serviceGuid)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred while trying to get the destination files artifact expansion " +
				"volume attributes for service with GUID '%v'",
			serviceGuid,
		)
	}
	destPvcName := destinationPvcAttrs.GetName().GetString()
	destPvcLabelStrs := map[string]string{}
	for labelKey, labelValue := range destinationPvcAttrs.GetLabels() {
		destPvcLabelStrs[labelKey.GetString()] = labelValue.GetString()
	}
	destinationPvc, err := backend.kubernetesManager.CreatePersistentVolumeClaim(
		ctx,
		enclaveNamespaceName,
		destPvcName,
		destPvcLabelStrs,
		backend.apiContainerModeArgs.filesArtifactExpansionVolumeSizeInMegabytes,
		backend.apiContainerModeArgs.storageClassName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating destination persistent volume claim in namespace '%v' for files artifact expansion '%v'", enclaveNamespaceName, filesArtifactExpansionGUID)
	}
	shouldDestroyPVC := true
	defer func() {
		if shouldDestroyPVC {
			err := backend.kubernetesManager.RemovePersistentVolumeClaim(
				ctx, destinationPvc)
			if err != nil {
				logrus.Errorf(
					"Files artifact expansion '%v' failed so we tried to destroy destination persistent volume '%v' that we " +
						"created, but doing so threw an error:\n%v",
					filesArtifactExpansionGUID,
					destPvcName,
					err,
				)
				logrus.Errorf(
					"ACTION REQUIRED: You will need to destroy persistent volume claim '%v' in namespace '%v' manually!",
					destPvcName,
					enclaveNamespaceName,
				)
			}
		}
	}()

	artifactFilepath := path.Join(enclaveDataVolumeDirpathOnExpanderContainer, filesArtifactFilepathRelativeToEnclaveDatadirRoot)

	jobAttrs, err := enclaveObjAttrsProvider.ForFilesArtifactExpansionJob(filesArtifactExpansionGUID, serviceGuid)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while trying to get the files artifact expander container attributes for files artifact expansion GUID '%v'", filesArtifactExpansionGUID)
	}

	if err := backend.runExtractionJobToCompletion(
		ctx,
		enclaveNamespaceName,
		enclaveDataPvc,
		destinationPvc,
		artifactFilepath,
		destinationVolumeMountDirpathOnExpanderContainer,
		jobAttrs,
	); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred running the extraction job to completion")
	}

	filesArtifactExpansion := files_artifact_expansion.NewFilesArtifactExpansion(filesArtifactExpansionGUID, serviceGuid)

	shouldDestroyPVC = false
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

/*
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

 */

/*
// TODO Push into KubernetesBackend??
func (backend *KubernetesKurtosisBackend) runExtractionJobToCompletion(
	ctx context.Context,
	namespaceName string,
	enclaveDataPvc *apiv1.PersistentVolumeClaim,
	destinationPvc *apiv1.PersistentVolumeClaim,
	artifactFilepathOnVolume string,
	destinationDirpath string,
	jobAttrs object_attributes_provider.KubernetesObjectAttributes,
) error {
	extractionCommand := getExtractionCommand(artifactFilepathOnVolume, destinationDirpath)

	jobName := jobAttrs.GetName()

	enclaveDataVolume := apiv1.Volume{
		Name: enclaveDataPvc.Spec.VolumeName,
		VolumeSource: apiv1.VolumeSource{
			PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
				ClaimName: enclaveDataPvc.GetName(),
				ReadOnly:  isEnclaveDataPersistentVolumeClaimReadOnly,
			},
		},
	}

	destinationVolume := apiv1.Volume{
		Name: destinationPvc.Spec.VolumeName,
		VolumeSource: apiv1.VolumeSource{
			PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
				ClaimName: destinationPvc.GetName(),
				ReadOnly:  isDestinationPersistentVolumeClaimReadOnly,
			},
		},
	}

	container := apiv1.Container{
		Name:                     filesArtifactExpansionContainerName,
		Image:                    dockerImage,
		Command:                  extractionCommand,
		VolumeMounts:             []apiv1.VolumeMount{
			{
				Name: enclaveDataPvc.Spec.VolumeName,
				MountPath: enclaveDataVolumeDirpathOnExpanderContainer,
			},
			{
				Name:      destinationPvc.Spec.VolumeName,
				MountPath: destinationVolumeMountDirpathOnExpanderContainer,
			},
		},
	}

	job, err := backend.kubernetesManager.CreateJobWithContainerAndVolume(
		ctx,
		namespaceName,
		jobName,
		jobAttrs.GetLabels(),
		jobAttrs.GetAnnotations(),
		[]apiv1.Container{container},
		[]apiv1.Volume{
			enclaveDataVolume,
			destinationVolume,
		},
		numTarExpansionRetries,
		ttlSecondsAfterFinishedExpanderJob,
	)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to create files artifact expansion job '%v'", jobName.GetString())
	}
	shouldDeleteJob := true
	defer func(){
		if shouldDeleteJob {
			// We delete instead of kill/stop because Kubernetes doesn't have the concept of keeping around stopped jobs
			// https://stackoverflow.com/a/52608258
			deleteJobError := backend.kubernetesManager.DeleteJob(ctx, namespaceName, job)
			if deleteJobError != nil {
				logrus.Errorf(
					"Running the extraction job to completion failed so we tried to delete job '%v' in namespace '%v' that we created, but doing so threw an error:\n%v",
					jobName.GetString(),
					namespaceName,
					deleteJobError,
				)
			}
		}
	}()

	hasJobEnded := false
	didJobSucceed := false
	jobEndedPoller := time.Tick(jobStatusPollerInterval)
	jobEndedTimeout := time.After(jobStatusPollerTimeout)
	for !hasJobEnded {
		select {
		case <-jobEndedTimeout:
			return stacktrace.NewError("Timed out after %v waiting for job '%v' to complete.", jobStatusPollerTimeout, job.Name)
		case <-jobEndedPoller:
			hasJobEnded, didJobSucceed, err = backend.kubernetesManager.GetJobCompletionAndSuccessFlags(ctx, namespaceName, job.Name)
			if err != nil {
				return stacktrace.Propagate(
					err,
					"Failed to get status for job '%v' in namespace '%v'",
					job.Name,
					namespaceName,
				)
			}
		}
	}
	if !didJobSucceed {
		jobPodsLabels := map[string]string{
			jobNamePodLabel: jobName.GetString(),
		}
		pods, err := backend.kubernetesManager.GetPodsByLabels(ctx, namespaceName, jobPodsLabels)
		if err != nil {
			return stacktrace.NewError("Job '%v' did not succeed, and we couldn't grab pod logs due to the following error: %v", jobName, err)
		}
		if len(pods.Items) == 0 {
			return stacktrace.NewError("Job '%v' did not succeed, and we couldn't grab pod logs because Kubernetes didn't return any pods for the job", jobName)
		}

		// This *seems* like a huge amount of work to go through, but the logs are actually invaluable for debugging
		containerLogs := backend.getAllJobContainerLogs(ctx, namespaceName, pods.Items)

		return stacktrace.NewError(
			"Job '%v' in namespace '%v' did not succeed; container logs are as follows:\n%v",
			jobName,
			namespaceName,
			strings.Join(containerLogs, "\n"),
		)
	}

	shouldDeleteJob = false
	return nil
}

func (backend *KubernetesKurtosisBackend) getAllJobContainerLogs(
	ctx context.Context,
	namespaceName string,
	pods []apiv1.Pod,
) []string {
	// TODO Parallelize to increase perf? But make sure we don't explode memory with huge pod logs
	// We go through all this work so that the user can see why the job failed
	containerLogStrs := []string{}
	for _, pod := range pods {
		for _, podContainer := range pod.Spec.Containers {
			strBuilder := strings.Builder{}
			strBuilder.WriteString(fmt.Sprintf(
				">>>>>>>>>>>>>>>>>>>> Pod %v - Container %v <<<<<<<<<<<<<<<<<<<<\n",
				pod.Name,
				podContainer.Name,
			))
			containerLogs, err := backend.getSingleJobContainerLogs(ctx, namespaceName, pod.Name, podContainer.Name)
			if err != nil {
				strBuilder.WriteString(fmt.Sprintf("Couldn't get logs for container due to an error:\n%v", err))
			} else {
				strBuilder.WriteString(containerLogs)
			}
			strBuilder.WriteString(fmt.Sprintf(
				"\n>>>>>>>>>>>>>>>>>> End Pod %v - Container %v <<<<<<<<<<<<<<<<<<<<",
				pod.Name,
				podContainer.Name,
			))
			containerLogStrs = append(containerLogStrs, strBuilder.String())
		}
	}

	return containerLogStrs
}

func (backend *KubernetesKurtosisBackend) getSingleJobContainerLogs(ctx context.Context, namespaceName string, podName string, containerName string) (string, error) {
	logs, err := backend.kubernetesManager.GetContainerLogs(
		ctx,
		namespaceName,
		podName,
		containerName,
		shouldFollowContainerLogsWhenArtifactExpansionJobFails,
		shouldAddTimestampsToContainerLogsWhenArtifactExpansionJobFails,
	)
	if err != nil {
		return "", stacktrace.Propagate(
			err,
			"An error occurred copying logs from container '%v' in pod '%v' in namespace '%v'",
			containerName,
			podName,
			namespaceName,
		)
	}
	defer logs.Close()

	output := &bytes.Buffer{}
	if _, err := io.Copy(output, logs); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred copying logs of container '%v' in pod '%v' to a buffer", containerName, podName)
	}

	return output.String(), nil
}
*/