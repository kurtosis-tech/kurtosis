package fluentbit

import (
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	kubeSystemNamespaceName = "kube-system"

	fluentBitImage = "fluent/fluent-bit:latest-debug"

	fluentBitConfigMapName = "fluent-bit-config"
	fluentBitConfigStr     = `
[SERVICE]
    Parsers_File /fluent-bit/etc/parsers.conf

[INPUT]
    Name              tail
    Tag               kurtosis.*
    Path              /var/log/containers/*_kt-*_user-service-container-*.log
    Parser            docker

[OUTPUT]
    Name              stdout
    Match             *
    Format            json_lines

[OUTPUT]
    Name              file
    Match             *
    Path              /fluent-bit-logs/
    File              fluentbit-output.log
    Format            plain

[FILTER]
    Name              kubernetes
    Match             kurtosis.*
    Merge_Log         On
    Merge_Log_Key     On
    Labels            On
    Annotations       On
    Kube_Tag_Prefix   kurtosis.var.log.containers.
`
)

type fluentbitLogsCollector struct{}

func NewFluentbitLogsCollector() *fluentbitLogsCollector {
	return &fluentbitLogsCollector{}
}

func (fluentbitPod *fluentbitLogsCollector) CreateAndStart(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	logsAggregatorHost string,
	logsAggregatorPort uint16,
	tcpPortNumber uint16,
	httpPortNumber uint16,
	logsCollectorTcpPortId string,
	logsCollectorHttpPortId string,
	targetNetworkId string,
	objAttrsProvider object_attributes_provider.KubernetesObjectAttributesProvider,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (
	resultContainerId string,
	resultContainerLabels map[string]string,
	resultHostMachinePortBindings map[nat.Port]*nat.PortBinding,
	resultRemoveLogsCollectorContainerFunc func(),
	resultErr error,
) {
	// TODO: the creation of this will likely move
	logsCollectorGuidStr, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return "", map[string]string{}, make(map[nat.Port]*nat.PortBinding), func() { return }, stacktrace.Propagate(err, "An error occurred creating uuid for logs collector.")
	}
	logsCollectorGuid := logs_collector.LogsCollectorGuid(logsCollectorGuidStr)
	logsCollectorAttrProvider := objAttrsProvider.ForLogsCollector(logsCollectorGuid)

	// create config creator - this can be done with config map + init container
	// create fluentbit container config provider - this can with config map + an init container
	err = CreateLogsCollectorConfigMap(ctx, logsCollectorAttrProvider, kubernetesManager)
	if err != nil {
		return "", map[string]string{}, make(map[nat.Port]*nat.PortBinding), func() { return }, stacktrace.Propagate(err, "An error occurred while trying to create config map for fluent-bit log collector.")
	}

	// TODO figure out what ports are needed for this container
	// get information about ports

	// create and start daemonest
	//logsCollectorDaemonSet, err := kubernetesManager.CreatePod(ctx, createAndStartArgs)
	//if err != nil {
	//	return "", nil, nil, nil, stacktrace.Propagate(err, "An error occurred starting the logs collector container with these args '%+v'", createAndStartArgs)
	//}
	_, err = CreateLogsCollectorDaemonSet(ctx, logsCollectorAttrProvider, kubernetesManager)
	if err != nil {
		return "", map[string]string{}, make(map[nat.Port]*nat.PortBinding), func() { return }, stacktrace.Propagate(err, "An error occurred while trying to daemonset for fluent-bit log collector.")
	}

	// TODO: enable remove mechanism
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
	return "", map[string]string{}, make(map[nat.Port]*nat.PortBinding), func() { return }, nil
}

func CreateLogsCollectorDaemonSet(
	ctx context.Context,
	objAttrProvider object_attributes_provider.KubernetesLogsCollectorObjectAttributesProvider,
	manager *kubernetes_manager.KubernetesManager) (*appsv1.DaemonSet, error) {
	daemonSetClient := manager.KubernetesClientSet.AppsV1().DaemonSets(kubeSystemNamespaceName)

	daemonSetAttrProvider, err := objAttrProvider.ForLogsCollectorDaemonSet()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating fluentbit daemonset.")
	}
	namespaceProvider, err := objAttrProvider.ForLogsCollectorNamespace()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating fluentbit daemonset.")
	}
	namespaceName := namespaceProvider.GetName().GetString()
	name := daemonSetAttrProvider.GetName().GetString()
	labels := shared_helpers.GetStringMapFromLabelMap(daemonSetAttrProvider.GetLabels())
	annotations := shared_helpers.GetStringMapFromAnnotationMap(daemonSetAttrProvider.GetAnnotations())

	daemonSet := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespaceName,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "fluent-bit"},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "fluent-bit"},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "fluent-bit",
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
					},
					Volumes: []apiv1.Volume{
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
									LocalObjectReference: apiv1.LocalObjectReference{Name: fluentBitConfigMapName},
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
					},
					InitContainers: []apiv1.Container{},
				},
			},
		},
	}

	// deploy the daemon set
	logsCollectorDaemonSet, err := daemonSetClient.Create(ctx, daemonSet, metav1.CreateOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating fluentbit daemonset.")
	}

	// TODO: wait for daemonset availability

	return logsCollectorDaemonSet, nil
}

func CreateLogsCollectorConfigMap(ctx context.Context,
	objAttrProvider object_attributes_provider.KubernetesLogsCollectorObjectAttributesProvider,
	manager *kubernetes_manager.KubernetesManager) error {
	configMapClient := manager.KubernetesClientSet.CoreV1().ConfigMaps(kubeSystemNamespaceName)

	configMapAttrProvider, err := objAttrProvider.ForLogsCollectorConfigMap()
	if err != nil {
		return err
	}
	namespaceProvider, err := objAttrProvider.ForLogsCollectorNamespace()
	if err != nil {
		return err
	}
	namespaceName := namespaceProvider.GetName().GetString()
	name := configMapAttrProvider.GetName().GetString()
	labels := shared_helpers.GetStringMapFromLabelMap(configMapAttrProvider.GetLabels())
	annotations := shared_helpers.GetStringMapFromAnnotationMap(configMapAttrProvider.GetAnnotations())

	configMap := &apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespaceName,
			Labels:      labels,
			Annotations: annotations,
		},
		Data: map[string]string{
			"fluent-bit.conf": fluentBitConfigStr,
		},
	}

	_, err = configMapClient.Create(ctx, configMap, metav1.CreateOptions{})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while creating config map for fluentbit log collector config.")
	}

	return nil
}
