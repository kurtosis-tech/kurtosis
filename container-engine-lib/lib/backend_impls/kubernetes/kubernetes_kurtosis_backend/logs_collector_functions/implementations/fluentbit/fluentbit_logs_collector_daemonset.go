package fluentbit

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
)

type fluentbitLogsCollector struct{}

func NewFluentbitLogsCollector() *fluentbitLogsCollector {
	return &fluentbitLogsCollector{}
}

func (fluentbitPod *fluentbitLogsCollector) CreateAndStart(
	ctx context.Context,
	logsAggregatorHost string,
	logsAggregatorPort uint16,
	tcpPortNumber uint16,
	httpPortNumber uint16,
	logsCollectorTcpPortId string,
	logsCollectorHttpPortId string,
	objAttrsProvider object_attributes_provider.KubernetesObjectAttributesProvider,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (
	*appsv1.DaemonSet,
	*apiv1.ConfigMap,
	func(),
	error,
) {
	logsCollectorGuidStr, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return nil, nil, func() { return }, stacktrace.Propagate(err, "An error occurred creating uuid for logs collector.")
	}
	logsCollectorGuid := logs_collector.LogsCollectorGuid(logsCollectorGuidStr)
	logsCollectorAttrProvider := objAttrsProvider.ForLogsCollector(logsCollectorGuid)

	configMap, err := CreateLogsCollectorConfigMap(ctx, logsCollectorAttrProvider, kubernetesManager)
	if err != nil {
		return nil, nil, func() { return }, stacktrace.Propagate(err, "An error occurred while trying to create config map for fluent-bit log collector.")
	}
	// TODO: create remove function

	// TODO: Get port information

	daemonSet, err := CreateLogsCollectorDaemonSet(ctx, configMap.Name, logsCollectorAttrProvider, kubernetesManager)
	if err != nil {
		return nil, nil, func() { return }, stacktrace.Propagate(err, "An error occurred while trying to daemonset for fluent-bit log collector.")
	}
	// TODO: create remove function
	//removeContainerFunc := func() {
	//	removeCtx := context.Background()
	//
	//	if err := kubernetesManager.RemovePod(removeCtx, containerId); err != nil {
	//		logrus.Errorf(
	//			"Launching the logs collector container with ID '%v' didn't complete successfully so we "+
	//				"tried to remove the container we started, but doing so exited with an error:\n%v",
	//			containerId,
	//			err)
	//		logrus.Errorf("ACTION REQUIRED: You'll need to manually remove the logs collector server with Docker container ID '%v'!!!!!!", containerId)
	//	}
	//}
	//shouldRemoveLogsCollectorContainer := true
	//defer func() {
	//	if shouldRemoveLogsCollectorContainer {
	//		removeContainerFunc()
	//	}
	//}()

	//shouldRemoveLogsCollectorContainer = false
	//shouldRemoveLogsCollectorConfigMap = false
	return daemonSet, configMap, func() { return }, nil
}

func CreateLogsCollectorDaemonSet(
	ctx context.Context,
	fluentBitCfgConfigMapName string,
	objAttrProvider object_attributes_provider.KubernetesLogsCollectorObjectAttributesProvider,
	kubernetesManager *kubernetes_manager.KubernetesManager) (*appsv1.DaemonSet, error) {

	daemonSetAttrProvider, err := objAttrProvider.ForLogsCollectorDaemonSet()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting logs collector daemon set attributes provider.")
	}
	name := daemonSetAttrProvider.GetName().GetString()
	labels := shared_helpers.GetStringMapFromLabelMap(daemonSetAttrProvider.GetLabels())
	annotations := shared_helpers.GetStringMapFromAnnotationMap(daemonSetAttrProvider.GetAnnotations())

	namespaceProvider, err := objAttrProvider.ForLogsCollectorNamespace()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting logs collector namespace attributes provider.")
	}
	namespaceName := namespaceProvider.GetName().GetString()

	containers := []apiv1.Container{
		{
			Name:  "fluent-bit", // how should it be named?
			Image: fluentBitImage,
			Args: []string{
				"/fluent-bit/bin/fluent-bit",
				"--config=/fluent-bit/etc/conf/fluent-bit.conf",
			},
			Ports: []apiv1.ContainerPort{
				{ContainerPort: 80},
			},
			VolumeMounts: []apiv1.VolumeMount{
				{
					Name:             "varlog",
					ReadOnly:         false,
					MountPath:        "/var/log",
					SubPath:          "",
					MountPropagation: nil,
					SubPathExpr:      "",
				},
				{
					Name:             "varlibdockercontainers",
					ReadOnly:         false,
					MountPath:        "/var/lib/docker/containers",
					SubPath:          "",
					MountPropagation: nil,
					SubPathExpr:      "",
				},
				{
					Name:             "varlogcontainers",
					ReadOnly:         false,
					MountPath:        "/var/log/containers",
					SubPath:          "",
					MountPropagation: nil,
					SubPathExpr:      "",
				},
				{
					Name:             "fluent-bit-config",
					ReadOnly:         false,
					MountPath:        "/fluent-bit/etc/conf",
					SubPath:          "",
					MountPropagation: nil,
					SubPathExpr:      "",
				},
				{
					Name:             "fluent-bit-host-logs",
					ReadOnly:         false,
					MountPath:        "/fluent-bit-logs",
					SubPath:          "",
					MountPropagation: nil,
					SubPathExpr:      "",
				},
			},
		},
	}

	volumes := []apiv1.Volume{
		{
			Name: "varlog",
			VolumeSource: apiv1.VolumeSource{
				HostPath: &apiv1.HostPathVolumeSource{
					Path: "/var/log",
					Type: nil,
				},
			},
		},
		{
			// is this where docker container logs are stored across all kubernetes clusters?
			Name: "varlibdockercontainers",
			VolumeSource: apiv1.VolumeSource{
				HostPath: &apiv1.HostPathVolumeSource{
					Path: "/var/lib/docker/containers",
					Type: nil,
				},
			},
		},
		{
			Name: "varlogcontainers",
			VolumeSource: apiv1.VolumeSource{
				HostPath: &apiv1.HostPathVolumeSource{
					Path: "/var/log/containers",
					Type: nil,
				},
			},
		},
		{
			Name: "fluent-bit-config",
			VolumeSource: apiv1.VolumeSource{
				ConfigMap: &apiv1.ConfigMapVolumeSource{
					LocalObjectReference: apiv1.LocalObjectReference{Name: fluentBitCfgConfigMapName},
					Items:                nil,
					DefaultMode:          nil,
					Optional:             nil,
				},
			},
		},
		{
			Name: "fluent-bit-host-logs",
			VolumeSource: apiv1.VolumeSource{
				HostPath: &apiv1.HostPathVolumeSource{
					Path: "/var/log/fluentbit",
				},
			},
		},
	}

	logsCollectorDaemonSet, err := kubernetesManager.CreateDaemonSet(
		ctx,
		namespaceName,
		name,
		labels,
		annotations,
		[]apiv1.Container{},
		containers,
		volumes,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating daemon set for fluent bit logs collector.")
	}

	return logsCollectorDaemonSet, nil
}

func CreateLogsCollectorConfigMap(ctx context.Context,
	objAttrProvider object_attributes_provider.KubernetesLogsCollectorObjectAttributesProvider,
	kubernetesManager *kubernetes_manager.KubernetesManager) (*apiv1.ConfigMap, error) {
	configMapAttrProvider, err := objAttrProvider.ForLogsCollectorConfigMap()
	if err != nil {
		return nil, err
	}
	name := configMapAttrProvider.GetName().GetString()
	labels := shared_helpers.GetStringMapFromLabelMap(configMapAttrProvider.GetLabels())
	annotations := shared_helpers.GetStringMapFromAnnotationMap(configMapAttrProvider.GetAnnotations())

	namespaceProvider, err := objAttrProvider.ForLogsCollectorNamespace()
	if err != nil {
		return nil, err
	}
	namespaceName := namespaceProvider.GetName().GetString()

	configMap, err := kubernetesManager.CreateConfigMap(
		ctx,
		namespaceName,
		name,
		labels,
		annotations,
		map[string]string{
			"fluent-bit.conf": fluentBitConfigStr,
		},
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating config map for fluent bit log collector config.")
	}

	return configMap, nil
}
