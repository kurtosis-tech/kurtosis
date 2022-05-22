package docker

/*
const (
	// Dirpath on the artifact expander container where the destination volume containing expanded files will be mounted
	destVolMntDirpathOnExpander = "/dest"

	// The Docker API's default is to return just containers whose status is "running"
	// However, we'd rather do our own filtering on what "running" means (because, e.g., "restarting"
	// should also be considered as running)
	shouldFetchAllExpansionContainersWhenRetrievingContainers = true

	// Docker image that will be used to launch the container that will expand the files artifact
	//  into a Docker volume
	dockerImage = "alpine:3.12"

	// Dirpath on the artifact expander container where the enclave data volume (which contains artifacts)
	//  will be mounted
	enclaveDataVolumeDirpathOnExpanderContainer = "/enclave-data"

	expanderContainerSuccessExitCode = 0
)

type filesArtifactExpansionObjectsAndDockerResources struct {
	filesArtifactExpansion *files_artifact_expansion.FilesArtifactExpansion

	// Will never be nil
	dockerResources *filesArtifactExpansionDockerResources
}

// Any of these values being nil indicates that the resource doesn't exist
type filesArtifactExpansionDockerResources struct {
	// Will never be nil because an API container is defined by its service
	volume *types.Volume

	container *docker_manager_types.Container
}

//Create a files artifact exansion volume for user service and file artifact id and runs a file artifact expander
func (backend *DockerKurtosisBackend) CreateFilesArtifactExpansion(ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceGuid service.ServiceGUID,
	filesArtifactFilepathRelativeToEnclaveDatadirRoot string) (*files_artifact_expansion.FilesArtifactExpansion, error) {

	filesArtifactExpansionGUIDStr, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to generate UUID for files artifact expansion.")
	}
	filesArtifactExpansionGUID := files_artifact_expansion.FilesArtifactExpansionGUID(filesArtifactExpansionGUIDStr)
	filesArtifactExpansionVolumeName, err := backend.createFilesArtifactExpansionVolume(
		ctx,
		enclaveId,
		filesArtifactExpansionGUID,
		serviceGuid,
	)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating files artifact expansion volume for user service with GUID '%v' and files " +
				"artifact relative filepath '%v' in enclave with ID '%v'",
			serviceGuid,
			filesArtifactFilepathRelativeToEnclaveDatadirRoot,
			enclaveId,
		)
	}
	// We don't delete the volume because we only defer-stop the expander container (not remove) and Docker won't let us remove a volume that's in use

	freeIpAddrProvider, found := backend.enclaveFreeIpProviders[enclaveId]
	if !found {
		return nil, stacktrace.NewError(
			"Received a request to run a files artifact expander attached to service with GUID '%v' in enclave '%v', but no free IP address provider was " +
				"defined for this enclave; this likely means that the request is being called where it shouldn't " +
				"be (i.e. outside the API container)",
			serviceGuid,
			enclaveId,
		)
	}

	ipAddr, err := freeIpAddrProvider.GetFreeIpAddr()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting a free IP address")
	}
	// We don't release the IP because we only defer-stop the expander container (not remove) so it will still be consuming the IP

	_, err = backend.runFilesArtifactExpander(
		ctx,
		filesArtifactExpansionGUID,
		serviceGuid,
		enclaveId,
		filesArtifactExpansionVolumeName,
		destVolMntDirpathOnExpander,
		ipAddr,
		filesArtifactFilepathRelativeToEnclaveDatadirRoot,
	)
	if err != nil {
		return nil,
		stacktrace.Propagate(
			err,
			"An error occurred running files artifact expander for user service with GUID '%v' and files artifact relative path '%v' in enclave with ID '%v'",
			serviceGuid,
			filesArtifactFilepathRelativeToEnclaveDatadirRoot,
			enclaveId,
		)
	}
	filesArtifactExpansion := files_artifact_expansion.NewFilesArtifactExpansion(filesArtifactExpansionGUID, serviceGuid)
	return filesArtifactExpansion, nil
}

// Destroy files artifact expansion volume and expander using the given filters
func (backend *DockerKurtosisBackend) DestroyFilesArtifactExpansions(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *files_artifact_expansion.FilesArtifactExpansionFilters,
)(
	resultSuccessfulFileArtifactExpansionGUIDs map[files_artifact_expansion.FilesArtifactExpansionGUID]bool,
	resultErroredFileArtifactExpansionGUIDs map[files_artifact_expansion.FilesArtifactExpansionGUID]error,
	resultErr error,
) {
	matchingFilesArtifactExpansionObjectsAndDockerResources, err := backend.getMatchingFilesArtifactExpansionObjectsAndDockerResources(
		ctx, enclaveId, filters)
	if err != nil {
		return nil,nil,stacktrace.Propagate(err, "Failed to get files expansion docker resources in enclave '%v' for filters '%+v'",
			enclaveId, filters)
	}
	successMap := map[files_artifact_expansion.FilesArtifactExpansionGUID]bool{}
	errorMap := map[files_artifact_expansion.FilesArtifactExpansionGUID]error{}
	for filesArtifactGUID, dockerResourcesAndObjects := range matchingFilesArtifactExpansionObjectsAndDockerResources {
		resources := dockerResourcesAndObjects.dockerResources

		if resources == nil {
			return nil, nil,
				stacktrace.NewError("Tried to delete Docker resources but none were given for files artifact expansion guid '%v'", filesArtifactGUID)
		}
		container := resources.container
		// Remove container
		if container != nil {
			containerErr := backend.dockerManager.RemoveContainer(ctx, container.GetName())
			if containerErr != nil {
				errorMap[filesArtifactGUID] = stacktrace.Propagate(containerErr, "An error occurred removing container '%v' for files artifact expansion '%v'", container.GetName(), filesArtifactGUID)
				continue
			}
		}

		volume := resources.volume
		// Remove volume
		if volume != nil {
			volumeErr := backend.dockerManager.RemoveVolume(ctx, volume.Name)
			if volumeErr != nil {
				errorMap[filesArtifactGUID] = stacktrace.Propagate(volumeErr, "An error occurred removing volume '%v' for files artifact expansion '%v'", volume.Name, filesArtifactGUID)
				continue
			}
		}

		successMap[filesArtifactGUID] = true
	}
	return successMap, errorMap, nil
}

// ====================== PRIVATE HELPERS =============================
func (backend *DockerKurtosisBackend) getMatchingFilesArtifactExpansionObjectsAndDockerResources(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *files_artifact_expansion.FilesArtifactExpansionFilters,
) (
	map[files_artifact_expansion.FilesArtifactExpansionGUID]*filesArtifactExpansionObjectsAndDockerResources,
	error,
) {
	matchingFilesArtifactExpansionDockerResources, err := backend.getMatchingFileArtifactExpansionDockerResources(
		ctx, enclaveId, filters.GUIDs)
	if err != nil {
		return nil,stacktrace.Propagate(err, "Failed to get files expansion docker resources in enclave '%v' for filters '%+v'",
			enclaveId, filters)
	}
	matchingFilesArtifactExpansionObjects, err := getFilesExpansionObjectsFromDockerResources(matchingFilesArtifactExpansionDockerResources)
	if err != nil {
		return nil,stacktrace.Propagate(err, "Failed to get files expansion objects in enclave '%v' for filters '%+v'",
			enclaveId, filters)
	}

	// Finally, apply the filters
	resultFilesArtifactExpansionObjectsAndDockerResources := map[files_artifact_expansion.FilesArtifactExpansionGUID]*filesArtifactExpansionObjectsAndDockerResources{}
	for filesArtifactExpansionGUID, filesArtifactExpansionObj := range matchingFilesArtifactExpansionObjects {
		if filters.GUIDs != nil && len(filters.GUIDs) > 0 {
			if _, found := filters.GUIDs[filesArtifactExpansionObj.GetGUID()]; !found {
				continue
			}
		}

		if filters.ServiceGUIDs != nil && len(filters.ServiceGUIDs) > 0 {
			if _, found := filters.ServiceGUIDs[filesArtifactExpansionObj.GetServiceGUID()]; !found {
				continue
			}
		}

		if _, found := matchingFilesArtifactExpansionDockerResources[filesArtifactExpansionGUID]; !found {
			return nil, stacktrace.NewError("Expected to find Docker resources matching files artifact expansion guid '%v' but none was found", filesArtifactExpansionGUID)
		}
		resultFilesArtifactExpansionObjectsAndDockerResources[filesArtifactExpansionGUID] = &filesArtifactExpansionObjectsAndDockerResources{
			filesArtifactExpansion: filesArtifactExpansionObj,
			dockerResources: matchingFilesArtifactExpansionDockerResources[filesArtifactExpansionGUID],
		}
	}
	return resultFilesArtifactExpansionObjectsAndDockerResources, nil
}

func (backend *DockerKurtosisBackend) getMatchingFileArtifactExpansionDockerResources(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	maybeFilesArtifactGuidsToMatch map[files_artifact_expansion.FilesArtifactExpansionGUID]bool,
) (map[files_artifact_expansion.FilesArtifactExpansionGUID]*filesArtifactExpansionDockerResources, error){
	resourcesByFilesArtifactGuid := map[files_artifact_expansion.FilesArtifactExpansionGUID]*filesArtifactExpansionDockerResources{}
	artifactVolumeLabels := map[string]string{
		label_key_consts.AppIDDockerLabelKey.GetString(): label_value_consts.AppIDDockerLabelValue.GetString(),
		label_key_consts.EnclaveIDDockerLabelKey.GetString():     string(enclaveId),
		label_key_consts.VolumeTypeDockerLabelKey.GetString(): label_value_consts.FilesArtifactExpansionVolumeTypeDockerLabelValue.GetString(),
	}
	volumes, err := backend.dockerManager.GetVolumesByLabels(ctx, artifactVolumeLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to find volumes for labels '%+v'", artifactVolumeLabels)
	}

	// FILTER VOLUMES BY GUIDS TO MATCH
	for _, volume := range volumes {
		filesArtifactExpansionGUIDStr, found := volume.Labels[label_key_consts.GUIDDockerLabelKey.GetString()]
		if !found {
			return nil, stacktrace.Propagate(err, "Failed to find GUID label for volume '%v'", volume.Name)
		}
		filesArtifactExpansionGUID := files_artifact_expansion.FilesArtifactExpansionGUID(filesArtifactExpansionGUIDStr)

		if maybeFilesArtifactGuidsToMatch != nil && len(maybeFilesArtifactGuidsToMatch) > 0 {
			if _, filesArtifactGUIDFound := maybeFilesArtifactGuidsToMatch[filesArtifactExpansionGUID]; !filesArtifactGUIDFound {
				continue
			}
		}

		resourcesByFilesArtifactGuid[filesArtifactExpansionGUID] = &filesArtifactExpansionDockerResources{volume: volume}
	}

	artifactContainerLabels := map[string]string{
		label_key_consts.AppIDDockerLabelKey.GetString(): label_value_consts.AppIDDockerLabelValue.GetString(),
		label_key_consts.EnclaveIDDockerLabelKey.GetString():     string(enclaveId),
		label_key_consts.ContainerTypeDockerLabelKey.GetString(): label_value_consts.FilesArtifactExpanderContainerTypeDockerLabelValue.GetString(),
	}

	containers, err := backend.dockerManager.GetContainersByLabels(ctx, artifactContainerLabels, shouldFetchAllExpansionContainersWhenRetrievingContainers)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to find volumes for labels '%+v'", artifactContainerLabels)
	}

	// FILTER CONTAINERS BY GUIDS TO MATCH
	for _, container := range containers {
		filesArtifactExpansionGUIDStr, found := container.GetLabels()[label_key_consts.GUIDDockerLabelKey.GetString()]
		if !found {
			return nil, stacktrace.NewError("Failed to find GUID label on docker container '%v'", container.GetId())
		}
		filesArtifactExpansionGUID := files_artifact_expansion.FilesArtifactExpansionGUID(filesArtifactExpansionGUIDStr)

		if maybeFilesArtifactGuidsToMatch != nil && len(maybeFilesArtifactGuidsToMatch) > 0 {
			if _, filesArtifactGUIDFound := maybeFilesArtifactGuidsToMatch[filesArtifactExpansionGUID]; !filesArtifactGUIDFound {
				continue
			}
		}

		resultObj, found := resourcesByFilesArtifactGuid[filesArtifactExpansionGUID]
		if !found {
			resultObj = &filesArtifactExpansionDockerResources{}
		}
		resultObj.container = container
		resourcesByFilesArtifactGuid[filesArtifactExpansionGUID] = resultObj
	}
	return resourcesByFilesArtifactGuid, nil
}

func getFilesExpansionObjectsFromDockerResources(
	allResources map[files_artifact_expansion.FilesArtifactExpansionGUID]*filesArtifactExpansionDockerResources,
) (map[files_artifact_expansion.FilesArtifactExpansionGUID]*files_artifact_expansion.FilesArtifactExpansion, error){
	filesArtifactExpansionObjects := map[files_artifact_expansion.FilesArtifactExpansionGUID]*files_artifact_expansion.FilesArtifactExpansion{}
	for filesArtifactExpansionGUID, dockerResource := range allResources {
		canonicalResource := dockerResource.volume
		serviceGUIDStr, found := canonicalResource.Labels[label_key_consts.UserServiceGUIDDockerLabelKey.GetString()]
		if !found {
			return nil, stacktrace.NewError("Found a volume as part of files artifact expansion '%v' without a service GUID label - this should never happen.", filesArtifactExpansionGUID)
		}
		serviceGUID := service.ServiceGUID(serviceGUIDStr)
		filesArtifactExpansionObjects[filesArtifactExpansionGUID] = files_artifact_expansion.NewFilesArtifactExpansion(filesArtifactExpansionGUID, serviceGUID)
	}
	return filesArtifactExpansionObjects, nil
}


func (backend *DockerKurtosisBackend) createFilesArtifactExpansionVolume(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filesArtifactExpansionGUID files_artifact_expansion.FilesArtifactExpansionGUID,
	serviceGUID service.ServiceGUID,
)(
	string,
	error,
) {

	enclaveObjAttrsProvider, err := backend.objAttrsProvider.ForEnclave(enclaveId)
	if err != nil {
		return "", stacktrace.Propagate(err, "Couldn't get an object attribute provider for enclave '%v'", enclaveId)
	}

	volumeAttrs, err := enclaveObjAttrsProvider.ForSingleFilesArtifactExpansionVolume(
		filesArtifactExpansionGUID,
		serviceGUID)
	if err != nil {
		return "", stacktrace.Propagate(
			err,
			"An error occurred while trying to get the files artifact expansion " +
				"volume attributes for service with GUID '%v'",
			serviceGUID,
		)
	}
	volumeName := volumeAttrs.GetName().GetString()
	volumeLabels := map[string]string{}
	for labelKey, labelValue := range volumeAttrs.GetLabels() {
		volumeLabels[labelKey.GetString()] = labelValue.GetString()
	}

	if err := backend.dockerManager.CreateVolume(ctx, volumeName, volumeLabels); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating the destination volume '%v' with labels '%+v'", volumeName, volumeLabels)
	}
	return volumeName, nil
}

func (backend *DockerKurtosisBackend) runFilesArtifactExpander(
	ctx context.Context,
	filesArtifactExpansionGUID files_artifact_expansion.FilesArtifactExpansionGUID,
	serviceGUID service.ServiceGUID,
	enclaveId enclave.EnclaveID,
	filesArtifactExpansionVolumeName string,
	destVolMntDirpathOnExpander string,
	ipAddr net.IP,
	filesArtifactFilepathRelativeToEnclaveDataVolumeRoot string,
) (expanderContainerId string, resultErr error) {

	enclaveNetwork, err := backend.getEnclaveNetworkByEnclaveId(ctx, enclaveId)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting enclave network by enclave ID '%v'", enclaveId)
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