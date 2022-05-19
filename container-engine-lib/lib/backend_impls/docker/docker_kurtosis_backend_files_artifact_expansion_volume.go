package docker

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact_expansion"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact_expansion_volume"
	"github.com/kurtosis-tech/stacktrace"
)

func (backend *DockerKurtosisBackend) createFilesArtifactExpansionVolume(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filesArtifactExpansion *files_artifact_expansion.FilesArtifactExpansion,
)(
	*files_artifact_expansion_volume.FilesArtifactExpansionVolume,
	error,
) {

	enclaveObjAttrsProvider, err := backend.objAttrsProvider.ForEnclave(enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Couldn't get an object attribute provider for enclave '%v'", enclaveId)
	}

	volumeAttrs, err := enclaveObjAttrsProvider.ForFilesArtifactExpansionVolume(filesArtifactExpansion.GetGUID())
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred while trying to get the files artifact expansion " +
				 "volume attributes for service with GUID '%v'",
			filesArtifactExpansion.GetServiceGUID(),
		)
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


// ====================================================================================================
//                                     Private Helper Methods
// ====================================================================================================

func getFileArtifactExpansionVolumeFromDockerVolumeInfo(
	name string,
	labels map[string]string,
) (*files_artifact_expansion_volume.FilesArtifactExpansionVolume, error) {
	enclaveId, found := labels[label_key_consts.EnclaveIDDockerLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected the volume's enclave ID to be found under label '%v' but the label wasn't present", label_key_consts.EnclaveIDDockerLabelKey.GetString())
	}

	newObject := files_artifact_expansion_volume.NewFilesArtifactExpansionVolume(
		files_artifact_expansion_volume.FilesArtifactExpansionVolumeName(name),
		enclave.EnclaveID(enclaveId),
	)

	return newObject, nil
}