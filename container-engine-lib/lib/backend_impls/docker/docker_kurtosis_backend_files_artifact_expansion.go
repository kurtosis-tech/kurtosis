package docker

import (
	"context"
	"github.com/docker/docker/api/types"
	docker_manager_types "github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact_expansion"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	// Dirpath on the artifact expander container where the destination volume containing expanded files will be mounted
	destVolMntDirpathOnExpander = "/dest"

	// The Docker API's default is to return just containers whose status is "running"
	// However, we'd rather do our own filtering on what "running" means (because, e.g., "restarting"
	// should also be considered as running)
	shouldFetchAllExpansionContainersWhenRetrievingContainers = true
)

type filesArtifactExpansionObjectsAndDockerResources struct {
	filesArtifactExpansionGUID *files_artifact_expansion.FilesArtifactExpansionGUID

	// May be nil if no pod has been started yet
	service *service.Service

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
	filesArtifactId service.FilesArtifactID,
	filesArtifactFilepathRelativeToEnclaveDatadirRoot string) (*files_artifact_expansion.FilesArtifactExpansion, error) {

	filesArtifactExpansionGUIDStr, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to generate UUID for files artifact expansion.")
	}
	filesArtifactExpansion := files_artifact_expansion.NewFilesArtifactExpansion(files_artifact_expansion.FilesArtifactExpansionGUID(filesArtifactExpansionGUIDStr), serviceGuid)
	filesArtifactExpansionVolumeName, err := backend.createFilesArtifactExpansionVolume(
		ctx,
		enclaveId,
		filesArtifactExpansion,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating files artifact expansion volume for user service with GUID '%v' and files artifact ID '%v' in enclave with ID '%v'", serviceGuid, filesArtifactId, enclaveId)
	}
	filesArtifactExpansionFilters := files_artifact_expansion.FilesArtifactExpansionFilters{
		GUIDs: map[files_artifact_expansion.FilesArtifactExpansionGUID]bool{
			filesArtifactExpansion.GetGUID(): true,
		},
	}
	defer func() {
		_, erroredVolumeNames, err := backend.DestroyFilesArtifactExpansion(ctx, enclaveId, filesArtifactExpansionFilters)
		if err != nil {
			logrus.Errorf("Failed to destroy expansion volumes for files artifact expansion '%v' - got error: \n'%v'", filesArtifactExpansion.GetGUID(), err)
		}
		for name, err := range erroredVolumeNames {
			logrus.Errorf("Failed to destroy expansion volume '%v' for files artifact expansion '%v' - get error: \n'%v'", name, filesArtifactExpansion.GetGUID(), err)
		}
	}()

	err = backend.runFilesArtifactExpander(
		ctx,
		filesArtifactExpansion,
		enclaveId,
		filesArtifactExpansionVolumeName,
		destVolMntDirpathOnExpander,
		filesArtifactFilepathRelativeToEnclaveDatadirRoot,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred running files artifact expander for user service with GUID '%v' and files artifact ID '%v' in enclave with ID '%v'", serviceGuid, filesArtifactId, enclaveId)
	}
	return filesArtifactExpansion, nil
}

// Destroy files artifact expansion volume and expander using the given filters
func (backend *DockerKurtosisBackend)  DestroyFilesArtifactExpansion(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters  files_artifact_expansion.FilesArtifactExpansionFilters,
)(
	successfulFileArtifactExpansionGUIDs map[files_artifact_expansion.FilesArtifactExpansionGUID]bool,
	erroredFileArtifactExpansionGUIDs map[files_artifact_expansion.FilesArtifactExpansionGUID]error,
	resultErr error,
) {
	matchingFilesArtifactExpansionDockerResources, err := backend.getMatchingFileArtifactExpansionDockerResources(
		ctx, enclaveId, filters.GUIDs, filters.ServiceGUIDs)
	if err != nil {
		return nil,nil,stacktrace.Propagate(err, "Failed to get files expansion docker resources in enclave '%v' for filters '%+v'",
			enclaveId, filters)
	}
	successMap := map[files_artifact_expansion.FilesArtifactExpansionGUID]bool{}
	errorMap := map[files_artifact_expansion.FilesArtifactExpansionGUID]error{}
	for filesArtifactGUID, dockerResources := range matchingFilesArtifactExpansionDockerResources {
		volume := dockerResources.volume
		container := dockerResources.container

		// Remove container
		if container != nil {
			containerErr := backend.dockerManager.RemoveContainer(ctx, container.GetName())
			if containerErr != nil {
				errorMap[filesArtifactGUID] = stacktrace.Propagate(containerErr,"Failed to clean up files artifact expansion artifact container. ")
			}
			continue
		}

		// Remove volume
		if volume != nil {
			volumeErr := backend.dockerManager.RemoveVolume(ctx, volume.Name)
			if volumeErr != nil {
				errorMap[filesArtifactGUID] = stacktrace.Propagate(volumeErr,"Failed to clean up files artifact expansion artifact volume. ")
			}
			continue
		}

		successMap[filesArtifactGUID] = true
	}
	return successMap, errorMap, nil
}

// ====================== PRIVATE HELPERS =============================

func (backend *DockerKurtosisBackend) getMatchingFileArtifactExpansionDockerResources(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	maybeFilesArtifactGuidsToMatch map[files_artifact_expansion.FilesArtifactExpansionGUID]bool,
	maybeServiceGuidsToMatch map[service.ServiceGUID]bool,
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
		filesArtifactExpansionGUIDStr := volume.Labels[label_key_consts.GUIDDockerLabelKey.GetString()]
		serviceGUIDstr := volume.Labels[label_key_consts.UserServiceGUIDDockerLabelKey.GetString()]
		filesArtifactExpansionGUID := files_artifact_expansion.FilesArtifactExpansionGUID(filesArtifactExpansionGUIDStr)
		serviceGUID := service.ServiceGUID(serviceGUIDstr)

		// Default true because empty filter set should return all
		filesArtifactGUIDFound := true
		if maybeFilesArtifactGuidsToMatch != nil && len(maybeFilesArtifactGuidsToMatch) > 0 {
			_, filesArtifactGUIDFound = maybeFilesArtifactGuidsToMatch[filesArtifactExpansionGUID]
		}

		// Default true because empty filter set should return all
		serviceGUIDFound := true
		if maybeServiceGuidsToMatch != nil && len(maybeServiceGuidsToMatch) > 0 {
			_, serviceGUIDFound = maybeServiceGuidsToMatch[serviceGUID]
		}

		if filesArtifactGUIDFound || serviceGUIDFound {
			resourcesByFilesArtifactGuid[filesArtifactExpansionGUID] = &filesArtifactExpansionDockerResources{volume: volume}
		}
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
		filesArtifactExpansionGUIDStr := container.GetLabels()[label_key_consts.GUIDDockerLabelKey.GetString()]
		serviceGUIDstr := container.GetLabels()[label_key_consts.UserServiceGUIDDockerLabelKey.GetString()]
		filesArtifactExpansionGUID := files_artifact_expansion.FilesArtifactExpansionGUID(filesArtifactExpansionGUIDStr)
		serviceGUID := service.ServiceGUID(serviceGUIDstr)

		// Default true because empty filter set should return all
		filesArtifactGUIDFound := true
		if maybeFilesArtifactGuidsToMatch != nil && len(maybeFilesArtifactGuidsToMatch) > 0 {
			_, filesArtifactGUIDFound = maybeFilesArtifactGuidsToMatch[filesArtifactExpansionGUID]
		}

		// Default true because empty filter set should return all
		serviceGUIDFound := true
		if maybeServiceGuidsToMatch != nil && len(maybeServiceGuidsToMatch) > 0 {
			_, serviceGUIDFound = maybeServiceGuidsToMatch[serviceGUID]
		}

		if filesArtifactGUIDFound || serviceGUIDFound {
			if resourcesByFilesArtifactGuid[filesArtifactExpansionGUID] == nil {
				resourcesByFilesArtifactGuid[filesArtifactExpansionGUID] = &filesArtifactExpansionDockerResources{container: container}
			} else {
				resourcesByFilesArtifactGuid[filesArtifactExpansionGUID].container = container
			}
		}
	}
	return resourcesByFilesArtifactGuid, nil
}
