package fluentbit

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/stacktrace"
	apiv1 "k8s.io/api/core/v1"
)

type fluentbitLogsCollectorDaemonSet struct{}

func NewFluentbitLogsCollectorDaemonSet() *fluentbitLogsCollectorDaemonSet {
	return &fluentbitLogsCollectorDaemonSet{}
}

func (fluentbitDaemonSet *fluentbitLogsCollectorDaemonSet) CreateAndStart(
	ctx context.Context,
	engineNamespace string,
	logsDatabaseHost string,
	logsDatabasePort uint16,
	tcpPortNumber uint16,
	httpPortNumber uint16,
	serviceAccountName string,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) error {
	logsCollectorConfigurationCreator := createFluentbitConfigurationCreatorForKurtosis(logsDatabaseHost, logsDatabasePort, tcpPortNumber, httpPortNumber)

	if err := logsCollectorConfigurationCreator.CreateConfiguration(context.Background(), engineNamespace, kubernetesManager); err != nil {
		return stacktrace.Propagate(err, "An error occurred running the logs collector configuration creator")
	}

	fluentbitConfigVolume, err := createVolume(configMapName, nil, &apiv1.ConfigMapVolumeSource{
		LocalObjectReference: apiv1.LocalObjectReference{
			Name: configMapName,
		},
		Items:       nil,
		DefaultMode: nil,
		Optional:    nil,
	})
	if err != nil {
		return stacktrace.Propagate(err, "Error creating volume '%s' for FluentBit. This is a Kurtosis internal error", configMapName)
	}

	varLogVolumeName := "varlog"
	varLogVolume, err := createVolume(varLogVolumeName, &apiv1.HostPathVolumeSource{
		Path: "/var/log",
		Type: nil,
	}, nil)
	if err != nil {
		return stacktrace.Propagate(err, "Error creating volume '%s' for FluentBit. This is a Kurtosis internal error", varLogVolumeName)
	}

	varLibDockerContainersVolumeName := "varlibdockercontainers"
	varLibDockerContainersVolume, err := createVolume(varLibDockerContainersVolumeName, &apiv1.HostPathVolumeSource{
		Path: "/var/lib/docker/containers",
		Type: nil,
	}, nil)
	if err != nil {
		return stacktrace.Propagate(err, "Error creating volume '%s' for FluentBit. This is a Kurtosis internal error", varLibDockerContainersVolumeName)
	}

	mntVolumeName := "mnt"
	mntVolume, err := createVolume(mntVolumeName, &apiv1.HostPathVolumeSource{
		Path: "/mnt",
		Type: nil,
	}, nil)
	if err != nil {
		return stacktrace.Propagate(err, "Error creating volume '%s' for FluentBit. This is a Kurtosis internal error", mntVolumeName)
	}

	fluentBitContainer := apiv1.Container{
		Name:         daemonSetName,
		Image:        containerImage,
		Command:      nil,
		Args:         nil,
		WorkingDir:   "",
		Ports:        nil,
		EnvFrom:      nil,
		Env:          nil,
		Resources:    apiv1.ResourceRequirements{},
		ResizePolicy: nil,
		VolumeMounts: []apiv1.VolumeMount{
			{
				Name:             configMapName,
				ReadOnly:         false,
				MountPath:        configDirpathInContainer,
				SubPath:          "",
				MountPropagation: nil,
				SubPathExpr:      "",
			},
			{
				Name:             varLogVolume.Name,
				ReadOnly:         false,
				MountPath:        "/var/log",
				SubPath:          "",
				MountPropagation: nil,
				SubPathExpr:      "",
			},
			{
				Name:             varLibDockerContainersVolume.Name,
				ReadOnly:         true,
				MountPath:        "/var/lib/docker/containers",
				SubPath:          "",
				MountPropagation: nil,
				SubPathExpr:      "",
			},
			{
				Name:             mntVolume.Name,
				ReadOnly:         true,
				MountPath:        "/mnt",
				SubPath:          "",
				MountPropagation: nil,
				SubPathExpr:      "",
			},
		},
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

	_, err = kubernetesManager.CreateDaemonSet(
		ctx,
		engineNamespace,
		daemonSetName,
		map[string]string{
			kubernetesAppLabel:              "fluent-bit-logging",
			"version":                       "v1",
			"kubernetes.io/cluster-service": "true",
		},
		map[string]string{},
		map[string]string{
			kubernetesAppLabel: "fluent-bit-logging",
		},
		[]apiv1.Container{
			fluentBitContainer,
		},
		[]apiv1.Volume{
			*fluentbitConfigVolume,
			*varLogVolume,
			*varLibDockerContainersVolume,
			*mntVolume,
		},
		serviceAccountName,
		10,
	)
	if err != nil {
		return stacktrace.Propagate(err, "Unable to create the Kubernetes DaemonSet for Fluentbit logs collector")
	}
	return nil
}

func (fluentbitDaemonSet *fluentbitLogsCollectorDaemonSet) AlreadyExists(
	ctx context.Context,
	engineNamespace string,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) bool {
	// TODO: this is a bit hacky b/c we just look at the error to check if the daemonset already exists. To be improved
	daemonSetMaybe, _ := kubernetesManager.GetDaemonSet(ctx, engineNamespace, daemonSetName)
	if daemonSetMaybe != nil {
		return true
	}
	return false
}

func createVolume(name string, hostPathMaybe *apiv1.HostPathVolumeSource, configMapMaybe *apiv1.ConfigMapVolumeSource) (*apiv1.Volume, error) {
	// the below condition is a logical XOR
	if (hostPathMaybe == nil) == (configMapMaybe == nil) {
		return nil, stacktrace.NewError("Need exactly one of hostPath or configMap")
	}

	volume := apiv1.Volume{
		Name: name,
		VolumeSource: apiv1.VolumeSource{
			HostPath:              hostPathMaybe,
			EmptyDir:              nil,
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
			ConfigMap:             configMapMaybe,
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
	}
	return &volume, nil

}
