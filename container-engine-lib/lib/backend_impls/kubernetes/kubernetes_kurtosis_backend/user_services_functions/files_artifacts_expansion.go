package user_services_functions

import (
	apiv1 "k8s.io/api/core/v1"
	"strconv"
)

const (
	filesArtifactExpanderInitContainerName = "files-artifact-expander"
	filesArtifactExpansionVolumeName = "files-artifact-expansion"
	isFilesArtifactExpansionVolumeReadOnly = false
)

// Functions required to do files artifacts expansion
func prepareFilesArtifactsExpansionResources(
	expanderImage string,
	expanderEnvVars map[string]string,
	expanderDirpathsToUserServiceDirpaths map[string]string,
) (
	resultPodVolumes []apiv1.Volume,
	resultUserServiceContainerVolumeMounts []apiv1.VolumeMount,
	resultPodInitContainers []apiv1.Container,
	resultErr error,
) {
	podVolumes := []apiv1.Volume{
		{
			Name: filesArtifactExpansionVolumeName,
			VolumeSource: apiv1.VolumeSource{
				EmptyDir: &apiv1.EmptyDirVolumeSource{},
			},
		},
	}

	volumeMountsOnExpanderContainer := []apiv1.VolumeMount{}
	volumeMountsOnUserServiceContainer := []apiv1.VolumeMount{}
	volumeSubdirIndex := 0
	for requestedExpanderDirpath, requestedUserServiceDirpath := range expanderDirpathsToUserServiceDirpaths {
		subdirName := strconv.Itoa(volumeSubdirIndex)

		expanderContainerMount := apiv1.VolumeMount{
			Name:      filesArtifactExpansionVolumeName,
			ReadOnly:  isFilesArtifactExpansionVolumeReadOnly,
			MountPath: requestedExpanderDirpath,
			SubPath:   subdirName,
		}
		volumeMountsOnExpanderContainer = append(volumeMountsOnExpanderContainer, expanderContainerMount)

		userServiceContainerMount := apiv1.VolumeMount{
			Name:      filesArtifactExpansionVolumeName,
			ReadOnly:  isFilesArtifactExpansionVolumeReadOnly,
			MountPath: requestedUserServiceDirpath,
			SubPath:   subdirName,
		}
		volumeMountsOnUserServiceContainer = append(volumeMountsOnUserServiceContainer, userServiceContainerMount)

		volumeSubdirIndex = volumeSubdirIndex + 1
	}

	filesArtifactExpansionInitContainer := getFilesArtifactExpansionInitContainerSpecs(
		expanderImage,
		expanderEnvVars,
		volumeMountsOnExpanderContainer,
	)

	podInitContainers := []apiv1.Container{
		filesArtifactExpansionInitContainer,
	}

	return podVolumes, volumeMountsOnUserServiceContainer, podInitContainers, nil
}

func getFilesArtifactExpansionInitContainerSpecs(
	image string,
	envVars map[string]string,
	volumeMounts []apiv1.VolumeMount,
) apiv1.Container {
	expanderEnvVars := []apiv1.EnvVar{}
	for key, value := range envVars {
		envVar := apiv1.EnvVar{
			Name:  key,
			Value: value,
		}
		expanderEnvVars = append(expanderEnvVars, envVar)
	}

	filesArtifactExpansionInitContainer := apiv1.Container{
		Name:         filesArtifactExpanderInitContainerName,
		Image:        image,
		Env:          expanderEnvVars,
		VolumeMounts: volumeMounts,
	}

	return filesArtifactExpansionInitContainer
}