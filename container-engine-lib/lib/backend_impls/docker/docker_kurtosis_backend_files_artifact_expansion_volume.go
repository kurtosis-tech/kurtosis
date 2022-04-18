package docker

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact_expansion_volume"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
)

func (backend *DockerKurtosisBackend) CreateFilesArtifactExpansionVolume(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceGuid service.ServiceGUID,
	filesArtifactId files_artifact.FilesArtifactID,
)(
	*files_artifact_expansion_volume.FilesArtifactExpansionVolume,
	error,
) {

	enclaveObjAttrsProvider, err := backend.objAttrsProvider.ForEnclave(enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Couldn't get an object attribute provider for enclave '%v'", enclaveId)
	}

	volumeAttrs, err := enclaveObjAttrsProvider.ForFilesArtifactExpansionVolume(serviceGuid, filesArtifactId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while trying to get the files artifact expansion volume attributes for user service with GUID '%v' and files artifact ID '%v'", serviceGuid, filesArtifactId)
	}
	volumeName := volumeAttrs.GetName().GetString()
	volumeLabels := map[string]string{}
	for labelKey, labelValue := range volumeAttrs.GetLabels() {
		volumeLabels[labelKey.GetString()] = labelValue.GetString()
	}

	foundVolumes, err := backend.dockerManager.GetVolumesByName(ctx, volumeName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Docker volumes by name '%v'", volumeName)
	}
	if len(foundVolumes) > 0 {
		//We iterate to check if it is exactly the same name
		for _, foundVolumeName := range foundVolumes {
			if volumeName == foundVolumeName {
				return nil, stacktrace.NewError("Volume can not be created because a volume with name '%v' already exists.", volumeName)
			}
		}
	}

	if err := backend.dockerManager.CreateVolume(ctx, volumeName, volumeLabels); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the destination volume '%v' with labels '%+v'", volumeName, volumeLabels)
	}

	newFileArtifactExpansionVolume, err := getFileArtifactExpansionVolumeFromDockerVolumeInfo(volumeName, volumeLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred converting volume with name '%v' into a file artifact expansion volume object", volumeName)
	}

	return newFileArtifactExpansionVolume, nil
}

func (backend *DockerKurtosisBackend) DestroyFilesArtifactExpansionVolumes(
	ctx context.Context,
	filters *files_artifact_expansion_volume.FilesArtifactExpansionVolumeFilters,
) (
	map[files_artifact_expansion_volume.FilesArtifactExpansionVolumeName]bool,
	map[files_artifact_expansion_volume.FilesArtifactExpansionVolumeName]error,
	error,
) {
	successfulExpansionVolumeNames := map[files_artifact_expansion_volume.FilesArtifactExpansionVolumeName]bool{}
	erroredExpansionVolumeNames  := map[files_artifact_expansion_volume.FilesArtifactExpansionVolumeName]error{}

	expansionVolumes, err := backend.getMatchingFileArtifactExpansionVolumes(ctx, filters)
	if err != nil {
		return nil, nil,  stacktrace.Propagate(err, "An error occurred getting files artifact expansion volumes matching filters '%+v'", filters)
	}

	//TODO execute concurrently to improve perf
	for expansionVolumeName := range expansionVolumes {
		volumeName := string(expansionVolumeName)
		if err := backend.dockerManager.RemoveVolume(ctx, volumeName); err != nil {
			wrappedErr := stacktrace.Propagate(
				err,
				"An error occurred removing volume with name '%v'",
				volumeName,
			)
			erroredExpansionVolumeNames[expansionVolumeName] = wrappedErr
			continue
		}
		successfulExpansionVolumeNames[expansionVolumeName] = true
	}

	return successfulExpansionVolumeNames, erroredExpansionVolumeNames, nil
}

// ====================================================================================================
//                                     Private Helper Methods
// ====================================================================================================
func (backend *DockerKurtosisBackend) getMatchingFileArtifactExpansionVolumes(
	ctx context.Context,
	filters *files_artifact_expansion_volume.FilesArtifactExpansionVolumeFilters,
) (map[files_artifact_expansion_volume.FilesArtifactExpansionVolumeName]*files_artifact_expansion_volume.FilesArtifactExpansionVolume, error) {
	searchLabels := map[string]string{
		label_key_consts.AppIDLabelKey.GetString(): label_value_consts.AppIDLabelValue.GetString(),
	}
	matchingVolumes, err := backend.dockerManager.GetVolumesByLabels(ctx, searchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred fetching volumes using labels: %+v", searchLabels)
	}

	matchingObjects := map[files_artifact_expansion_volume.FilesArtifactExpansionVolumeName]*files_artifact_expansion_volume.FilesArtifactExpansionVolume{}
	for _, volume := range matchingVolumes {
		object, err := getFileArtifactExpansionVolumeFromDockerVolumeInfo(
			volume.Name,
			volume.Labels,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred converting volume with name '%v' into a files artifact expansion volume object", volume.Name)
		}

		if filters.Names != nil && len(filters.Names) > 0 {
			if _, found := filters.Names[object.GetName()]; !found {
				continue
			}
		}

		if filters.EnclaveIDs != nil && len(filters.EnclaveIDs) > 0 {
			if _, found := filters.EnclaveIDs[object.GetEnclaveID()]; !found {
				continue
			}
		}

		matchingObjects[object.GetName()] = object
	}

	return matchingObjects, nil
}

func getFileArtifactExpansionVolumeFromDockerVolumeInfo(
	name string,
	labels map[string]string,
) (*files_artifact_expansion_volume.FilesArtifactExpansionVolume, error) {
	enclaveId, found := labels[label_key_consts.EnclaveIDLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected the volume's enclave ID to be found under label '%v' but the label wasn't present", label_key_consts.EnclaveIDLabelKey.GetString())
	}

	newObject := files_artifact_expansion_volume.NewFilesArtifactExpansionVolume(
		files_artifact_expansion_volume.FilesArtifactExpansionVolumeName(name),
		enclave.EnclaveID(enclaveId),
	)

	return newObject, nil
}
