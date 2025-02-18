package fluentbit

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"time"
)

type fluentbitLogsCollector struct{}

func NewFluentbitLogsCollector() *fluentbitLogsCollector {
	return &fluentbitLogsCollector{}
}

func (fluentbit *fluentbitLogsCollector) CreateAndStart(
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
	*apiv1.Namespace,
	func(),
	error,
) {
	logsCollectorGuidStr, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return nil, nil, nil, nil, stacktrace.Propagate(err, "An error occurred creating uuid for logs collector.")
	}

	logsCollectorGuid := logs_collector.LogsCollectorGuid(logsCollectorGuidStr)
	logsCollectorAttrProvider := objAttrsProvider.ForLogsCollector(logsCollectorGuid)

	namespace, err := createLogsCollectorNamespace(ctx, logsCollectorAttrProvider, kubernetesManager)
	if err != nil {
		return nil, nil, nil, nil, stacktrace.Propagate(err, "An error occurred creating namespace for logs collector.")
	}
	removeNamespaceFunc := func() {
		removeCtx := context.Background()
		if err := kubernetesManager.RemoveNamespace(removeCtx, namespace); err != nil {
			logrus.Errorf(
				"Launching the logs collector daemon set with name '%v' didn't complete successfully so we "+
					"tried to remove the namespace we started, but doing so exited with an error:\n%v",
				namespace.Name,
				err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually remove the logs collector namespace with Kubernetes name '%v'!!!!!!", namespace.Name)
		}
	}
	shouldRemoveLogsCollectorNamespace := true
	defer func() {
		if shouldRemoveLogsCollectorNamespace {
			removeNamespaceFunc()
		}
	}()

	configMap, err := createLogsCollectorConfigMap(ctx, namespace.Name, logsCollectorAttrProvider, kubernetesManager)
	if err != nil {
		return nil, nil, nil, nil, stacktrace.Propagate(err, "An error occurred while trying to create config map for fluent-bit log collector.")
	}
	removeConfigMapFunc := func() {
		removeCtx := context.Background()
		if err := kubernetesManager.RemoveConfigMap(removeCtx, namespace.Name, configMap); err != nil {
			logrus.Errorf(
				"Launching the logs collector daemon set with name '%v' didn't complete successfully so we "+
					"tried to remove the config map we started, but doing so exited with an error:\n%v",
				configMap.Name,
				err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually remove the logs collector config map with Kubernetes name '%v' in namespace '%v'!!!!!!", configMap.Name, configMap.Namespace)
		}
	}
	shouldRemoveLogsCollectorConfigMap := true
	defer func() {
		if shouldRemoveLogsCollectorConfigMap {
			removeConfigMapFunc()
		}
	}()

	// TODO: Get port information

	daemonSet, err := createLogsCollectorDaemonSet(ctx, namespace.Name, configMap.Name, logsCollectorAttrProvider, kubernetesManager)
	if err != nil {
		return nil, nil, nil, nil, stacktrace.Propagate(err, "An error occurred while trying to daemonset for fluent-bit log collector.")
	}
	removeDaemonSetFunc := func() {
		removeCtx := context.Background()
		if err := kubernetesManager.RemoveDaemonSet(removeCtx, namespace.Name, daemonSet); err != nil {
			logrus.Errorf(
				"Launching the logs collector daemon with name '%v' didn't complete successfully so we "+
					"tried to remove the daemon set we started, but doing so exited with an error:\n%v",
				daemonSet.Name,
				err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually remove the logs collector daemon set with Kubernetes name '%v' in namespace '%v'!!!!!!", daemonSet.Name, daemonSet.Namespace)
		}
	}
	shouldRemoveLogsCollectorDaemonSet := true
	defer func() {
		if shouldRemoveLogsCollectorDaemonSet {
			removeDaemonSetFunc()
		}
	}()

	// wait until the first pod associated with this daemon set is online before returning
	if err = waitForAtLeastOneActivePodManagedByDaemonSet(ctx, daemonSet, kubernetesManager); err != nil {
		return nil, nil, nil, nil, stacktrace.Propagate(err, "An error occurred waiting for at least one active pod managed by logs collector daemon set '%v'", daemonSet.Name)
	}

	removeLogsCollectorFunc := func() {
		removeDaemonSetFunc()
		removeConfigMapFunc()
		removeNamespaceFunc()
	}

	shouldRemoveLogsCollectorNamespace = false
	shouldRemoveLogsCollectorConfigMap = false
	shouldRemoveLogsCollectorDaemonSet = false
	return daemonSet, configMap, namespace, removeLogsCollectorFunc, nil
}

func (fluentbit *fluentbitLogsCollector) GetHttpHealthCheckEndpoint() string {
	return healthCheckEndpointPath
}

func createLogsCollectorDaemonSet(
	ctx context.Context,
	namespace string,
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
		namespace,
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

func createLogsCollectorConfigMap(
	ctx context.Context,
	namespace string,
	objAttrProvider object_attributes_provider.KubernetesLogsCollectorObjectAttributesProvider,
	kubernetesManager *kubernetes_manager.KubernetesManager) (*apiv1.ConfigMap, error) {
	configMapAttrProvider, err := objAttrProvider.ForLogsCollectorConfigMap()
	if err != nil {
		return nil, err
	}
	name := configMapAttrProvider.GetName().GetString()
	labels := shared_helpers.GetStringMapFromLabelMap(configMapAttrProvider.GetLabels())
	annotations := shared_helpers.GetStringMapFromAnnotationMap(configMapAttrProvider.GetAnnotations())

	configMap, err := kubernetesManager.CreateConfigMap(
		ctx,
		namespace,
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

func createLogsCollectorNamespace(
	ctx context.Context,
	objAttrProvider object_attributes_provider.KubernetesLogsCollectorObjectAttributesProvider,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (*apiv1.Namespace, error) {
	namespaceProvider, err := objAttrProvider.ForLogsCollectorNamespace()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting logs collector namespace attributes provider.")
	}
	namespaceName := namespaceProvider.GetName().GetString()
	namespaceLabels := shared_helpers.GetStringMapFromLabelMap(namespaceProvider.GetLabels())
	namespaceAnnotations := shared_helpers.GetStringMapFromAnnotationMap(namespaceProvider.GetAnnotations())

	namespaceObj, err := kubernetesManager.CreateNamespace(ctx, namespaceName, namespaceLabels, namespaceAnnotations)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating namepsace for logs collector with name '%s'", namespaceName)
	}

	return namespaceObj, nil

}

func waitForAtLeastOneActivePodManagedByDaemonSet(ctx context.Context, logsCollectorDaemonSet *appsv1.DaemonSet, kubernetesManager *kubernetes_manager.KubernetesManager) error {
	retryInterval := 500 * time.Millisecond
	maxRetries := 10
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(maxRetries)*retryInterval)
	defer cancel()

	ticker := time.NewTicker(retryInterval)
	defer ticker.Stop()
	for attempt := 0; attempt < maxRetries; attempt++ {
		select {
		case <-timeoutCtx.Done():
			return stacktrace.NewError(
				"Timeout waiting for a pod managed by logs collector daemon set '%s' to come online",
				logsCollectorDaemonSet.Name,
			)
		case <-ticker.C:
			pods, err := kubernetesManager.GetPodsManagedByDaemonSet(ctx, logsCollectorDaemonSet)
			if err != nil {
				return stacktrace.Propagate(err, "An error occurred getting pods managed by logs collector daemon set '%v'", logsCollectorDaemonSet.Name)
			}
			if len(pods) > 0 {
				// found a pod, success
				return nil
			}
		}
	}
	return stacktrace.NewError(
		"Exceeded max retries (%d) waiting for a pod managed by daemon set '%s' to come online",
		maxRetries, logsCollectorDaemonSet.Name,
	)
}
