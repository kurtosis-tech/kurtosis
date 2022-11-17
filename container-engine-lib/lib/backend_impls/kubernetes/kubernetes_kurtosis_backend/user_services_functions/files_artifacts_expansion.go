package user_services_functions

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strconv"
	"time"
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
				HostPath: &apiv1.HostPathVolumeSource{
					Path: "",
					Type: nil,
				},
				EmptyDir: &apiv1.EmptyDirVolumeSource{
					Medium: "",
					SizeLimit: &resource.Quantity{
						Format: "",
					},
				},
				GCEPersistentDisk: &apiv1.GCEPersistentDiskVolumeSource{
					PDName:    "",
					FSType:    "",
					Partition: 0,
					ReadOnly:  false,
				},
				AWSElasticBlockStore: &apiv1.AWSElasticBlockStoreVolumeSource{
					VolumeID:  "",
					FSType:    "",
					Partition: 0,
					ReadOnly:  false,
				},
				GitRepo: &apiv1.GitRepoVolumeSource{
					Repository: "",
					Revision:   "",
					Directory:  "",
				},
				Secret: &apiv1.SecretVolumeSource{
					SecretName:  "",
					Items:       nil,
					DefaultMode: nil,
					Optional:    nil,
				},
				NFS: &apiv1.NFSVolumeSource{
					Server:   "",
					Path:     "",
					ReadOnly: false,
				},
				ISCSI: &apiv1.ISCSIVolumeSource{
					TargetPortal:      "",
					IQN:               "",
					Lun:               0,
					ISCSIInterface:    "",
					FSType:            "",
					ReadOnly:          false,
					Portals:           nil,
					DiscoveryCHAPAuth: false,
					SessionCHAPAuth:   false,
					SecretRef: &apiv1.LocalObjectReference{
						Name: "",
					},
					InitiatorName: nil,
				},
				Glusterfs: &apiv1.GlusterfsVolumeSource{
					EndpointsName: "",
					Path:          "",
					ReadOnly:      false,
				},
				PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
					ClaimName: "",
					ReadOnly:  false,
				},
				RBD: &apiv1.RBDVolumeSource{
					CephMonitors: nil,
					RBDImage:     "",
					FSType:       "",
					RBDPool:      "",
					RadosUser:    "",
					Keyring:      "",
					SecretRef: &apiv1.LocalObjectReference{
						Name: "",
					},
					ReadOnly: false,
				},
				FlexVolume: &apiv1.FlexVolumeSource{
					Driver: "",
					FSType: "",
					SecretRef: &apiv1.LocalObjectReference{
						Name: "",
					},
					ReadOnly: false,
					Options:  nil,
				},
				Cinder: &apiv1.CinderVolumeSource{
					VolumeID: "",
					FSType:   "",
					ReadOnly: false,
					SecretRef: &apiv1.LocalObjectReference{
						Name: "",
					},
				},
				CephFS: &apiv1.CephFSVolumeSource{
					Monitors:   nil,
					Path:       "",
					User:       "",
					SecretFile: "",
					SecretRef: &apiv1.LocalObjectReference{
						Name: "",
					},
					ReadOnly: false,
				},
				Flocker: &apiv1.FlockerVolumeSource{
					DatasetName: "",
					DatasetUUID: "",
				},
				DownwardAPI: &apiv1.DownwardAPIVolumeSource{
					Items:       nil,
					DefaultMode: nil,
				},
				FC: &apiv1.FCVolumeSource{
					TargetWWNs: nil,
					Lun:        nil,
					FSType:     "",
					ReadOnly:   false,
					WWIDs:      nil,
				},
				AzureFile: &apiv1.AzureFileVolumeSource{
					SecretName: "",
					ShareName:  "",
					ReadOnly:   false,
				},
				ConfigMap: &apiv1.ConfigMapVolumeSource{
					LocalObjectReference: apiv1.LocalObjectReference{
						Name: "",
					},
					Items:       nil,
					DefaultMode: nil,
					Optional:    nil,
				},
				VsphereVolume: &apiv1.VsphereVirtualDiskVolumeSource{
					VolumePath:        "",
					FSType:            "",
					StoragePolicyName: "",
					StoragePolicyID:   "",
				},
				Quobyte: &apiv1.QuobyteVolumeSource{
					Registry: "",
					Volume:   "",
					ReadOnly: false,
					User:     "",
					Group:    "",
					Tenant:   "",
				},
				AzureDisk: &apiv1.AzureDiskVolumeSource{
					DiskName:    "",
					DataDiskURI: "",
					CachingMode: nil,
					FSType:      nil,
					ReadOnly:    nil,
					Kind:        nil,
				},
				PhotonPersistentDisk: &apiv1.PhotonPersistentDiskVolumeSource{
					PdID:   "",
					FSType: "",
				},
				Projected: &apiv1.ProjectedVolumeSource{
					Sources:     nil,
					DefaultMode: nil,
				},
				PortworxVolume: &apiv1.PortworxVolumeSource{
					VolumeID: "",
					FSType:   "",
					ReadOnly: false,
				},
				ScaleIO: &apiv1.ScaleIOVolumeSource{
					Gateway: "",
					System:  "",
					SecretRef: &apiv1.LocalObjectReference{
						Name: "",
					},
					SSLEnabled:       false,
					ProtectionDomain: "",
					StoragePool:      "",
					StorageMode:      "",
					VolumeName:       "",
					FSType:           "",
					ReadOnly:         false,
				},
				StorageOS: &apiv1.StorageOSVolumeSource{
					VolumeName:      "",
					VolumeNamespace: "",
					FSType:          "",
					ReadOnly:        false,
					SecretRef: &apiv1.LocalObjectReference{
						Name: "",
					},
				},
				CSI: &apiv1.CSIVolumeSource{
					Driver:           "",
					ReadOnly:         nil,
					FSType:           nil,
					VolumeAttributes: nil,
					NodePublishSecretRef: &apiv1.LocalObjectReference{
						Name: "",
					},
				},
				Ephemeral: &apiv1.EphemeralVolumeSource{
					VolumeClaimTemplate: &apiv1.PersistentVolumeClaimTemplate{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "",
							GenerateName:    "",
							Namespace:       "",
							SelfLink:        "",
							UID:             "",
							ResourceVersion: "",
							Generation:      0,
							CreationTimestamp: metav1.Time{
								Time: time.Time{},
							},
							DeletionTimestamp: &metav1.Time{
								Time: time.Time{},
							},
							DeletionGracePeriodSeconds: nil,
							Labels:                     nil,
							Annotations:                nil,
							OwnerReferences:            nil,
							Finalizers:                 nil,
							ZZZ_DeprecatedClusterName:  "",
							ManagedFields:              nil,
						},
						Spec: apiv1.PersistentVolumeClaimSpec{
							AccessModes: nil,
							Selector: &metav1.LabelSelector{
								MatchLabels:      nil,
								MatchExpressions: nil,
							},
							Resources: apiv1.ResourceRequirements{
								Limits:   nil,
								Requests: nil,
							},
							VolumeName:       "",
							StorageClassName: nil,
							VolumeMode:       nil,
							DataSource: &apiv1.TypedLocalObjectReference{
								APIGroup: nil,
								Kind:     "",
								Name:     "",
							},
							DataSourceRef: &apiv1.TypedLocalObjectReference{
								APIGroup: nil,
								Kind:     "",
								Name:     "",
							},
						},
					},
				},
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
			Name:  key,
			Value: value,
			ValueFrom: &apiv1.EnvVarSource{
				FieldRef: &apiv1.ObjectFieldSelector{
					APIVersion: "",
					FieldPath:  "",
				},
				ResourceFieldRef: &apiv1.ResourceFieldSelector{
					ContainerName: "",
					Resource:      "",
					Divisor: resource.Quantity{
						Format: "",
					},
				},
				ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
					LocalObjectReference: apiv1.LocalObjectReference{
						Name: "",
					},
					Key:      "",
					Optional: nil,
				},
				SecretKeyRef: &apiv1.SecretKeySelector{
					LocalObjectReference: apiv1.LocalObjectReference{
						Name: "",
					},
					Key:      "",
					Optional: nil,
				},
			},
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
		VolumeMounts:  volumeMounts,
		VolumeDevices: nil,
		LivenessProbe: &apiv1.Probe{
			ProbeHandler: apiv1.ProbeHandler{
				Exec: &apiv1.ExecAction{
					Command: nil,
				},
				HTTPGet: &apiv1.HTTPGetAction{
					Path: "",
					Port: intstr.IntOrString{
						Type:   0,
						IntVal: 0,
						StrVal: "",
					},
					Host:        "",
					Scheme:      "",
					HTTPHeaders: nil,
				},
				TCPSocket: &apiv1.TCPSocketAction{
					Port: intstr.IntOrString{
						Type:   0,
						IntVal: 0,
						StrVal: "",
					},
					Host: "",
				},
				GRPC: &apiv1.GRPCAction{
					Port:    0,
					Service: nil,
				},
			},
			InitialDelaySeconds:           0,
			TimeoutSeconds:                0,
			PeriodSeconds:                 0,
			SuccessThreshold:              0,
			FailureThreshold:              0,
			TerminationGracePeriodSeconds: nil,
		},
		ReadinessProbe: &apiv1.Probe{
			ProbeHandler: apiv1.ProbeHandler{
				Exec: &apiv1.ExecAction{
					Command: nil,
				},
				HTTPGet: &apiv1.HTTPGetAction{
					Path: "",
					Port: intstr.IntOrString{
						Type:   0,
						IntVal: 0,
						StrVal: "",
					},
					Host:        "",
					Scheme:      "",
					HTTPHeaders: nil,
				},
				TCPSocket: &apiv1.TCPSocketAction{
					Port: intstr.IntOrString{
						Type:   0,
						IntVal: 0,
						StrVal: "",
					},
					Host: "",
				},
				GRPC: &apiv1.GRPCAction{
					Port:    0,
					Service: nil,
				},
			},
			InitialDelaySeconds:           0,
			TimeoutSeconds:                0,
			PeriodSeconds:                 0,
			SuccessThreshold:              0,
			FailureThreshold:              0,
			TerminationGracePeriodSeconds: nil,
		},
		StartupProbe: &apiv1.Probe{
			ProbeHandler: apiv1.ProbeHandler{
				Exec: &apiv1.ExecAction{
					Command: nil,
				},
				HTTPGet: &apiv1.HTTPGetAction{
					Path: "",
					Port: intstr.IntOrString{
						Type:   0,
						IntVal: 0,
						StrVal: "",
					},
					Host:        "",
					Scheme:      "",
					HTTPHeaders: nil,
				},
				TCPSocket: &apiv1.TCPSocketAction{
					Port: intstr.IntOrString{
						Type:   0,
						IntVal: 0,
						StrVal: "",
					},
					Host: "",
				},
				GRPC: &apiv1.GRPCAction{
					Port:    0,
					Service: nil,
				},
			},
			InitialDelaySeconds:           0,
			TimeoutSeconds:                0,
			PeriodSeconds:                 0,
			SuccessThreshold:              0,
			FailureThreshold:              0,
			TerminationGracePeriodSeconds: nil,
		},
		Lifecycle: &apiv1.Lifecycle{
			PostStart: &apiv1.LifecycleHandler{
				Exec: &apiv1.ExecAction{
					Command: nil,
				},
				HTTPGet: &apiv1.HTTPGetAction{
					Path: "",
					Port: intstr.IntOrString{
						Type:   0,
						IntVal: 0,
						StrVal: "",
					},
					Host:        "",
					Scheme:      "",
					HTTPHeaders: nil,
				},
				TCPSocket: &apiv1.TCPSocketAction{
					Port: intstr.IntOrString{
						Type:   0,
						IntVal: 0,
						StrVal: "",
					},
					Host: "",
				},
			},
			PreStop: &apiv1.LifecycleHandler{
				Exec: &apiv1.ExecAction{
					Command: nil,
				},
				HTTPGet: &apiv1.HTTPGetAction{
					Path: "",
					Port: intstr.IntOrString{
						Type:   0,
						IntVal: 0,
						StrVal: "",
					},
					Host:        "",
					Scheme:      "",
					HTTPHeaders: nil,
				},
				TCPSocket: &apiv1.TCPSocketAction{
					Port: intstr.IntOrString{
						Type:   0,
						IntVal: 0,
						StrVal: "",
					},
					Host: "",
				},
			},
		},
		TerminationMessagePath:   "",
		TerminationMessagePolicy: "",
		ImagePullPolicy:          "",
		SecurityContext: &apiv1.SecurityContext{
			Capabilities: &apiv1.Capabilities{
				Add:  nil,
				Drop: nil,
			},
			Privileged: nil,
			SELinuxOptions: &apiv1.SELinuxOptions{
				User:  "",
				Role:  "",
				Type:  "",
				Level: "",
			},
			WindowsOptions: &apiv1.WindowsSecurityContextOptions{
				GMSACredentialSpecName: nil,
				GMSACredentialSpec:     nil,
				RunAsUserName:          nil,
				HostProcess:            nil,
			},
			RunAsUser:                nil,
			RunAsGroup:               nil,
			RunAsNonRoot:             nil,
			ReadOnlyRootFilesystem:   nil,
			AllowPrivilegeEscalation: nil,
			ProcMount:                nil,
			SeccompProfile: &apiv1.SeccompProfile{
				Type:             "",
				LocalhostProfile: nil,
			},
		},
		Stdin:     false,
		StdinOnce: false,
		TTY:       false,
	}

	return filesArtifactExpansionInitContainer
}
