package docker

import (
	"context"
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
	files_artifact_expansion_volume.FilesArtifactExpansionVolumeName,
	error,
) {

	enclaveObjAttrsProvider, err := backend.objAttrsProvider.ForEnclave(enclaveId)
	if err != nil {
		return "", stacktrace.Propagate(err, "Couldn't get an object attribute provider for enclave '%v'", enclaveId)
	}

	volumeAttrs, err := enclaveObjAttrsProvider.ForFilesArtifactExpansionVolume(
		filesArtifactExpansion.GetGUID(),
		filesArtifactExpansion.GetServiceGUID())
	if err != nil {
		return "", stacktrace.Propagate(
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
		return "", stacktrace.Propagate(err, "An error occurred getting Docker volumes by name '%v'", volumeName)
	}
	if len(foundVolumes) > 0 {
		//We iterate to check if it is exactly the same name
		for _, foundVolumeName := range foundVolumes {
			if volumeName == foundVolumeName {
				return "", stacktrace.NewError("Volume can not be created because a volume with name '%v' already exists.", volumeName)
			}
		}
	}

	if err := backend.dockerManager.CreateVolume(ctx, volumeName, volumeLabels); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating the destination volume '%v' with labels '%+v'", volumeName, volumeLabels)
	}
	return files_artifact_expansion_volume.FilesArtifactExpansionVolumeName(volumeName), nil
}
