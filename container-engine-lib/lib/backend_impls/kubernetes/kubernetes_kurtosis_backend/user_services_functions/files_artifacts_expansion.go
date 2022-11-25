package user_services_functions

import (
	apiv1 "k8s.io/api/core/v1"
	"strconv"
)

const (
	filesArtifactExpanderInitContainerName = "files-artifact-expander"
	filesArtifactExpansionVolumeName       = "files-artifact-expansion"
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
				HostPath: nil,
				EmptyDir: &apiv1.EmptyDirVolumeSource{
					Medium:    "",
					SizeLimit: nil,
				},
				GCEPersistentDisk:     nil,
				AWSElasticBlockStore:  nil,
				GitRepo:               nil,
				Secret:                nil,
				NFS:                   nil,
				ISCSI:                 nil,
				Glusterfs:             nil,
				PersistentVolumeClaim: nil,
				RBD:                   nil,
				FlexVolume:            nil,
				Cinder:                nil,
				CephFS:                nil,
				Flocker:               nil,
				DownwardAPI:           nil,
				FC:                    nil,
				AzureFile:             nil,
				ConfigMap:             nil,
				VsphereVolume:         nil,
				Quobyte:               nil,
				AzureDisk:             nil,
				PhotonPersistentDisk:  nil,
				Projected:             nil,
				PortworxVolume:        nil,
				ScaleIO:               nil,
				StorageOS:             nil,
				CSI:                   nil,
				Ephemeral:             nil,
			},
		},
	}

	volumeMountsOnExpanderContainer := []apiv1.VolumeMount{}
	volumeMountsOnUserServiceContainer := []apiv1.VolumeMount{}
	volumeSubdirIndex := 0
	for requestedExpanderDirpath, requestedUserServiceDirpath := range expanderDirpathsToUserServiceDirpaths {
		subdirName := strconv.Itoa(volumeSubdirIndex)

		expanderContainerMount := apiv1.VolumeMount{
			Name:             filesArtifactExpansionVolumeName,
			ReadOnly:         isFilesArtifactExpansionVolumeReadOnly,
			MountPath:        requestedExpanderDirpath,
			SubPath:          subdirName,
			MountPropagation: nil,
			SubPathExpr:      "",
		}
		volumeMountsOnExpanderContainer = append(volumeMountsOnExpanderContainer, expanderContainerMount)

		userServiceContainerMount := apiv1.VolumeMount{
			Name:             filesArtifactExpansionVolumeName,
			ReadOnly:         isFilesArtifactExpansionVolumeReadOnly,
			MountPath:        requestedUserServiceDirpath,
			SubPath:          subdirName,
			MountPropagation: nil,
			SubPathExpr:      "",
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
			Name:      key,
			Value:     value,
			ValueFrom: nil,
		}
		expanderEnvVars = append(expanderEnvVars, envVar)
	}

	filesArtifactExpansionInitContainer := apiv1.Container{
		Name:       filesArtifactExpanderInitContainerName,
		Image:      image,
		Command:    nil,
		Args:       nil,
		WorkingDir: "",
		Ports:      nil,
		EnvFrom:    nil,
		Env:        expanderEnvVars,
		Resources: apiv1.ResourceRequirements{
			Limits:   nil,
			Requests: nil,
		},
		VolumeMounts:             volumeMounts,
		VolumeDevices:            nil,
		LivenessProbe:            nil,
		ReadinessProbe:           nil,
		StartupProbe:             nil,
		Lifecycle:                nil,
		TerminationMessagePath:   "",
		TerminationMessagePolicy: "",
		ImagePullPolicy:          "",
		SecurityContext:          nil,
		Stdin:                    false,
		StdinOnce:                false,
		TTY:                      false,
	}

	return filesArtifactExpansionInitContainer
}
