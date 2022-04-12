package docker

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/file_artifact"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/file_artifacts_expansion_volume"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
)

func (backend *DockerKurtosisBackend) CreateFileArtifactsExpansionVolume(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceGuid service.ServiceGUID,
	fileArtifactId file_artifact.FilterArtifactID,
)(
	*file_artifacts_expansion_volume.FileArtifactsExpansionVolume,
	error,
) {

	enclaveObjAttrsProvider, err := backend.objAttrsProvider.ForEnclave(enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Couldn't get an object attribute provider for enclave '%v'", enclaveId)
	}

	containerAttrs, err := enclaveObjAttrsProvider.ForFilesArtifactExpansionVolume(serviceGuid, fileArtifactId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while trying to get the file artifacts expansion volume attributes for user service with GUID '%v'", serviceGuid)
	}
	volumeName := containerAttrs.GetName().GetString()
	volumeLabels := map[string]string{}
	for labelKey, labelValue := range containerAttrs.GetLabels() {
		volumeLabels[labelKey.GetString()] = labelValue.GetString()
	}

	foundedVolumes, err := backend.dockerManager.GetVolumesByName(ctx, volumeName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Docker volumes by name '%v'", volumeName)
	}
	if len(foundedVolumes) > 0 {
		//We iterate to check if it is exactly the same name
		for _, foundedVolumeName := range foundedVolumes {
			if volumeName == foundedVolumeName {
				return nil, stacktrace.NewError("Volume can not be created because a volume with name '%v' already exists.", volumeName)
			}
		}
	}

	if err := backend.dockerManager.CreateVolume(ctx, volumeName, volumeLabels); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the destination volume '%v' with labels '%+v'", volumeName, volumeLabels)
	}

	guid := file_artifacts_expansion_volume.FileArtifactExpansionVolumeGUID(volumeName)

	newFileArtifactExpansionVolume := file_artifacts_expansion_volume.NewFileArtifactsExpansionVolume(guid)

	return newFileArtifactExpansionVolume, nil
}
