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
	volumeName := filesArtifactExpansionVolume.GetName()
	defer func() {
		// TODO TODO TODO Destroy the expansion volume
		panic("implement me")
		//backend.destroyFilesArtifactExpansionVolume(ctx, volumeName)
	}()

	_, err = backend.runFilesArtifactExpander(
		ctx,
		newFilesArtifactExpanderGUID(filesArtifactId, serviceGuid),
		enclaveId,
		volumeName,
		destVolMntDirpathOnExpander,
		filesArtifactFilepathRelativeToEnclaveDatadirRoot,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred running files artifact expander for user service with GUID '%v' and files artifact ID '%v' and files artifact expansion volume '%v' in enclave with ID '%v'", serviceGuid, filesArtifactId, volumeName, enclaveId)
	}
	return nil, nil
}

//Destroy files artifact expansion volume and expander using the given filters
func (backend *DockerKurtosisBackend)  DestroyFilesArtifactExpansion(
	ctx context.Context,
	filters  files_artifact_expansion.FilesArtifactExpansionFilters,
)(
	successfulFileArtifactExpansionGUIDs map[files_artifact_expansion.FilesArtifactExpansionGUID]bool,
	erroredFileArtifactExpansionGUIDs map[files_artifact_expansion.FilesArtifactExpansionGUID]error,
	resultErr error,
) {
	panic("IMPLEMENT ME")
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

func newFilesArtifactExpanderGUID(filesArtifactId service.FilesArtifactID, serviceGuid service.ServiceGUID) files_artifact_expander.FilesArtifactExpanderGUID {
	serviceRegistrationGuidStr := string(serviceGuid)
	filesArtifactIdStr := string(filesArtifactId)
	suffix := getCurrentTimeStr()
	guidStr := strings.Join([]string{serviceRegistrationGuidStr, filesArtifactIdStr, suffix}, guidElementSeparator)
	guid := files_artifact_expander.FilesArtifactExpanderGUID(guidStr)
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
