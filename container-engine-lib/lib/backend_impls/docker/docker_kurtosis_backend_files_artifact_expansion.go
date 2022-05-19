package docker

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact_expander"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact_expansion"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
)

const (
	guidElementSeparator = "-"
	// TODO Change this to base 16 to be more compact??
	guidBase = 10
	// Dirpath on the artifact expander container where the destination volume containing expanded files will be mounted
	destVolMntDirpathOnExpander = "/dest"

	// The Docker API's default is to return just containers whose status is "running"
	// However, we'd rather do our own filtering on what "running" means (because, e.g., "restarting"
	// should also be considered as running)
	shouldFetchAllExpansionContainersWhenRetrievingContainers = true
)

//Create a files artifact exansion volume for user service and file artifact id and runs a file artifact expander
func (backend *DockerKurtosisBackend) CreateFilesArtifactExpansion(ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceGuid service.ServiceGUID,
	filesArtifactId service.FilesArtifactID,
	filesArtifactFilepathRelativeToEnclaveDatadirRoot string) (*files_artifact_expansion.FilesArtifactExpansion, error) {

	filesArtifactExpansion := files_artifact_expansion.NewFilesArtifactExpansion(newFilesArtifactExpansionGUID(filesArtifactId, serviceGuid), serviceGuid)
	filesArtifactExpansionVolume, err := backend.createFilesArtifactExpansionVolume(
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

	_, err = backend.runFilesArtifactExpander(
		ctx,
		filesArtifactExpansion,
		enclaveId,
		filesArtifactExpansionVolume.GetName(),
		destVolMntDirpathOnExpander,
		filesArtifactFilepathRelativeToEnclaveDatadirRoot,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred running files artifact expander for user service with GUID '%v' and files artifact ID '%v' in enclave with ID '%v'", serviceGuid, filesArtifactId, enclaveId)
	}
	return filesArtifactExpansion, nil
}

//Destroy files artifact expansion volume and expander using the given filters
func (backend *DockerKurtosisBackend)  DestroyFilesArtifactExpansion(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters  files_artifact_expansion.FilesArtifactExpansionFilters,
)(
	successfulFileArtifactExpansionGUIDs map[files_artifact_expansion.FilesArtifactExpansionGUID]bool,
	erroredFileArtifactExpansionGUIDs map[files_artifact_expansion.FilesArtifactExpansionGUID]error,
	resultErr error,
) {
	erroredFileArtifactExpansionGUIDs = map[files_artifact_expansion.FilesArtifactExpansionGUID]error{}
	successfulFileArtifactExpansionGUIDs = filters.GUIDs
	artifactVolumeLabels := map[string]string{
		label_key_consts.AppIDDockerLabelKey.GetString(): label_value_consts.AppIDDockerLabelValue.GetString(),
		label_key_consts.VolumeTypeDockerLabelKey.GetString(): label_value_consts.FilesArtifactExpansionVolumeTypeDockerLabelValue.GetString(),
	}
	volumes, err := backend.dockerManager.GetVolumesByLabels(ctx, artifactVolumeLabels)
	if err != nil {
		return nil,nil, stacktrace.Propagate(err, "Failed to find volumes for filters '%+v'", filters)
	}
	artifactContainerLabels := map[string]string{
		label_key_consts.AppIDDockerLabelKey.GetString(): label_value_consts.AppIDDockerLabelValue.GetString(),
		label_key_consts.ContainerTypeDockerLabelKey.GetString(): label_value_consts.FilesArtifactExpanderContainerTypeDockerLabelValue.GetString(),
	}
	containers, err := backend.dockerManager.GetContainersByLabels(ctx, artifactContainerLabels, shouldFetchAllExpansionContainersWhenRetrievingContainers)
	if err != nil {
		return nil,nil, stacktrace.Propagate(err, "Failed to find volumes for filters '%+v'", filters)
	}
	for _, volume := range volumes {
		filesArtifactExpansionGUIDStr := volume.Labels[label_key_consts.GUIDDockerLabelKey.GetString()]
		filesArtifactExpansionGUID := files_artifact_expansion.FilesArtifactExpansionGUID(filesArtifactExpansionGUIDStr)
		if filters.GUIDs[filesArtifactExpansionGUID] {
			err := backend.dockerManager.RemoveVolume(ctx, volume.Name)
			if err != nil {
				erroredFileArtifactExpansionGUIDs[filesArtifactExpansionGUID] = err
				successfulFileArtifactExpansionGUIDs[filesArtifactExpansionGUID] = false
			}
		}
	}
	for _, container := range containers {
		filesArtifactExpansionGUIDStr := container.GetLabels()[label_key_consts.GUIDDockerLabelKey.GetString()]
		filesArtifactExpansionGUID := files_artifact_expansion.FilesArtifactExpansionGUID(filesArtifactExpansionGUIDStr)
		if filters.GUIDs[filesArtifactExpansionGUID] {
			err := backend.dockerManager.RemoveContainer(ctx, container.GetId())
			if err != nil {
				if erroredFileArtifactExpansionGUIDs[filesArtifactExpansionGUID] != nil {
					volumeErr := erroredFileArtifactExpansionGUIDs[filesArtifactExpansionGUID]
					erroredFileArtifactExpansionGUIDs[filesArtifactExpansionGUID] = stacktrace.NewError(
						"Failed to delete files artifact expansion with the following errors: '%v', '%v'",
						err,
						volumeErr)
				} else {
					erroredFileArtifactExpansionGUIDs[filesArtifactExpansionGUID] = err
					successfulFileArtifactExpansionGUIDs[filesArtifactExpansionGUID] = false
				}
			}
		}
	}
	return successfulFileArtifactExpansionGUIDs, erroredFileArtifactExpansionGUIDs, nil
}

// ====================== PRIVATE HELPERS =============================

func (backend *DockerKurtosisBackend) getMatchingFilesArtifactExpansions(
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

func newFilesArtifactExpansionGUID(filesArtifactId service.FilesArtifactID, serviceGuid service.ServiceGUID) files_artifact_expansion.FilesArtifactExpansionGUID {
	serviceRegistrationGuidStr := string(serviceGuid)
	filesArtifactIdStr := string(filesArtifactId)
	suffix := getCurrentTimeStr()
	guidStr := strings.Join([]string{serviceRegistrationGuidStr, filesArtifactIdStr, suffix}, guidElementSeparator)
	guid := files_artifact_expansion.FilesArtifactExpansionGUID(guidStr)
	return guid
}

// Provides the current time in string form, for use as a suffix to a container ID (e.g. service ID, module ID) that will
//  make it unique so it won't collide with other containers with the same ID
func getCurrentTimeStr() string {
	now := time.Now()
	// TODO make this UnixNano to reduce risk of collisions???
	nowUnixSecs := now.Unix()
	return strconv.FormatInt(nowUnixSecs, guidBase)
}
