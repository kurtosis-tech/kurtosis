package vector

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_value"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type vectorLogsAggregatorDeployment struct{}

func NewVectorLogsAggregatorDeployment() *vectorLogsAggregatorDeployment {
	return &vectorLogsAggregatorDeployment{}
}

func (logsAggregator *vectorLogsAggregatorDeployment) CreateAndStart(
	ctx context.Context,
	logsListeningPort uint16,
	objAttrsProvider object_attributes_provider.KubernetesObjectAttributesProvider,
	kubernetesManager *kubernetes_manager.KubernetesManager) (
	*apiv1.Service,
	*apiv1.Namespace,
	*appsv1.Deployment,
	*apiv1.ConfigMap,
	func(),
	error) {
	logsAggregatorGuidStr, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return nil, nil, nil, nil, nil, stacktrace.Propagate(err, "An error occurred creating uuid for logs collector.")
	}
	logsAggregatorGuid := logs_aggregator.LogsAggregatorGuid(logsAggregatorGuidStr)
	logsAggregatorAttrProvider := objAttrsProvider.ForLogsAggregator(logsAggregatorGuid)

	namespace, err := createLogsAggregatorNamespace(ctx, logsAggregatorAttrProvider, kubernetesManager)
	if err != nil {
		return nil, nil, nil, nil, nil, stacktrace.Propagate(err, "An error occurred creating namespace for logs aggregator.")
	}
	removeNamespaceFunc := func() {
		removeCtx := context.Background()
		if err := kubernetesManager.RemoveNamespace(removeCtx, namespace); err != nil {
			logrus.Errorf(
				"Launching the logs aggregator deployment with name '%v' didn't complete successfully so we "+
					"tried to remove the namespace we started, but doing so exited with an error:\n%v",
				namespace.Name,
				err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually remove the logs aggregator namespace with Kubernetes name '%v'!!!!!!", namespace.Name)
		}
	}
	shouldRemoveLogsAggregatorNamespace := true
	defer func() {
		if shouldRemoveLogsAggregatorNamespace {
			removeNamespaceFunc()
		}
	}()

	configMap, err := createLogsAggregatorConfigMap(ctx, namespace.Name, logsAggregatorAttrProvider, kubernetesManager)
	if err != nil {
		return nil, nil, nil, nil, nil, stacktrace.Propagate(err, "An error occurred while trying to create config map for vector logs aggregator.")
	}
	removeConfigMapFunc := func() {
		removeCtx := context.Background()
		if err := kubernetesManager.RemoveConfigMap(removeCtx, namespace.Name, configMap); err != nil {
			logrus.Errorf(
				"Launching the logs aggregator deployment with name '%v' didn't complete successfully so we "+
					"tried to remove the config map we started, but doing so exited with an error:\n%v",
				configMap.Name,
				err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually remove the logs aggregator config map with Kubernetes name '%v' in namespace '%v'!!!!!!", configMap.Name, configMap.Namespace)
		}
	}
	shouldRemoveLogsAggregatorConfigMap := false
	defer func() {
		if shouldRemoveLogsAggregatorConfigMap {
			removeConfigMapFunc()
		}
	}()

	deployment, deploymentLabels, err := createLogsAggregatorDeployment(ctx, namespace.Name, logsListeningPort, configMap.Name, logsAggregatorAttrProvider, kubernetesManager)
	if err != nil {
		return nil, nil, nil, nil, nil, stacktrace.Propagate(err, "An error occurred while trying to create daemon set for fluent bit logs collector.")
	}
	removeDeploymentFunc := func() {
		removeCtx := context.Background()
		if err := kubernetesManager.RemoveDeployment(removeCtx, namespace.Name, deployment); err != nil {
			logrus.Errorf(
				"Launching the logs aggregator deployment with name '%v' didn't complete successfully so we "+
					"tried to remove the daemon set we started, but doing so exited with an error:\n%v",
				deployment.Name,
				err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually remove the logs aggregator deployment with Kubernetes name '%v' in namespace '%v'!!!!!!", deployment.Name, deployment.Namespace)
		}
	}
	shouldRemoveLogsAggregatorDeployment := true
	defer func() {
		if shouldRemoveLogsAggregatorDeployment {
			removeDeploymentFunc()
		}
	}()

	service, err := createLogsAggregatorService(ctx, namespace.Name, logsListeningPort, deploymentLabels, logsAggregatorAttrProvider, kubernetesManager)
	if err != nil {
		return nil, nil, nil, nil, nil, stacktrace.Propagate(err, "An error occurred creating service for logs aggregator.")
	}
	removeServiceFunc := func() {
		removeCtx := context.Background()
		if err := kubernetesManager.RemoveService(removeCtx, service); err != nil {
			logrus.Errorf(
				"Launching the logs aggregator deployment with name '%v' didn't complete successfully so we "+
					"tried to remove the service we started, but doing so exited with an error:\n%v",
				service.Name,
				err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually remove the logs aggregator service with Kubernetes name '%v'!!!!!!", service.Name)
		}
	}
	shouldRemoveLogsAggregatorService := true
	defer func() {
		if shouldRemoveLogsAggregatorService {
			removeServiceFunc()
		}
	}()

	removeLogsAggregatorFunc := func() {
		removeConfigMapFunc()
		removeDeploymentFunc()
		removeServiceFunc()
		removeNamespaceFunc()
	}

	shouldRemoveLogsAggregatorDeployment = false
	shouldRemoveLogsAggregatorService = false
	shouldRemoveLogsAggregatorNamespace = false
	return service, namespace, deployment, nil, removeLogsAggregatorFunc, nil
}

func createLogsAggregatorDeployment(
	ctx context.Context,
	namespace string,
	logsListeningPort uint16,
	configMapName string,
	objAttrProvider object_attributes_provider.KubernetesLogsAggregatorObjectAttributesProvider,
	kubernetesManager *kubernetes_manager.KubernetesManager) (*appsv1.Deployment, map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue, error) {
	deploymentAttrProvider, err := objAttrProvider.ForLogsAggregatorDeployment()
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting logs aggregator deployment attributes provider.")
	}

	labels := shared_helpers.GetStringMapFromLabelMap(deploymentAttrProvider.GetLabels())
	annotations := shared_helpers.GetStringMapFromAnnotationMap(deploymentAttrProvider.GetAnnotations())
	name := deploymentAttrProvider.GetName().GetString()

	containers := []apiv1.Container{
		{
			Name:       vectorContainerName,
			Image:      vectorImage,
			Command:    nil,
			Args:       []string{"--config", fmt.Sprintf("%s/vector.toml", vectorConfigMountPath)},
			WorkingDir: "",
			Ports: []apiv1.ContainerPort{
				{
					Name:          "fluent",
					HostPort:      0,
					ContainerPort: int32(logsListeningPort),
					Protocol:      "",
					HostIP:        "",
				},
			},
			EnvFrom:      nil,
			Env:          nil,
			Resources:    apiv1.ResourceRequirements{},
			ResizePolicy: nil,
			VolumeMounts: []apiv1.VolumeMount{
				{
					Name:             vectorConfigVolumeName,
					ReadOnly:         false,
					MountPath:        vectorConfigMountPath,
					SubPath:          "",
					MountPropagation: nil,
					SubPathExpr:      "",
				},
				{
					Name:             kurtosisLogsVolumeName,
					ReadOnly:         false,
					MountPath:        kurtosisLogsMountPath,
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
		},
	}

	volumes := []apiv1.Volume{
		{
			Name:         vectorConfigVolumeName,
			VolumeSource: kubernetesManager.GetVolumeSourceForConfigMap(configMapName),
		},
		{
			Name:         kurtosisLogsMountPath,
			VolumeSource: kubernetesManager.GetVolumeSourceForHostPath(kurtosisLogsMountPath),
		},
	}

	logsAggregatorDeployment, err := kubernetesManager.CreateDeployment(
		ctx,
		namespace,
		name,
		labels,
		annotations,
		[]apiv1.Container{}, // no need init containers
		containers,
		volumes,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred creating deployment for vector logs aggregator.")
	}

	return logsAggregatorDeployment, deploymentAttrProvider.GetLabels(), nil
}

func createLogsAggregatorNamespace(
	ctx context.Context,
	objAttrProvider object_attributes_provider.KubernetesLogsAggregatorObjectAttributesProvider,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (*apiv1.Namespace, error) {
	namespaceAttrProvider, err := objAttrProvider.ForLogsAggregatorNamespace()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting logs aggregator namespace attributes provider.")
	}
	namespaceName := namespaceAttrProvider.GetName().GetString()
	namespaceLabels := shared_helpers.GetStringMapFromLabelMap(namespaceAttrProvider.GetLabels())
	namespaceAnnotations := shared_helpers.GetStringMapFromAnnotationMap(namespaceAttrProvider.GetAnnotations())

	namespaceObj, err := kubernetesManager.CreateNamespace(ctx, namespaceName, namespaceLabels, namespaceAnnotations)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating namespace for logs aggregator with name '%s'", namespaceName)
	}

	return namespaceObj, nil
}

func createLogsAggregatorService(
	ctx context.Context,
	namespace string,
	logListeningPort uint16,
	logsAggregatorDeploymentLabels map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue,
	objAttrProvider object_attributes_provider.KubernetesLogsAggregatorObjectAttributesProvider,
	kubernetesManager *kubernetes_manager.KubernetesManager) (*apiv1.Service, error) {
	serviceAttrProvider, err := objAttrProvider.ForLogsAggregatorNamespace()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting logs aggregator namespace attributes provider.")
	}
	serviceName := serviceAttrProvider.GetName().GetString()
	serviceLabels := shared_helpers.GetStringMapFromLabelMap(serviceAttrProvider.GetLabels())
	serviceAnnotations := shared_helpers.GetStringMapFromAnnotationMap(serviceAttrProvider.GetAnnotations())

	// for now logs aggregator will only be reachable form within the cluster
	// this might need to changed later for exporting logs cluster
	serviceType := apiv1.ServiceTypeClusterIP

	ports := []apiv1.ServicePort{
		{
			Name:        "logs-collector-forwarding",
			Protocol:    "",
			AppProtocol: nil,
			Port:        int32(logListeningPort),
			TargetPort:  intstr.IntOrString{IntVal: int32(logListeningPort)},
			NodePort:    0,
		},
	}

	matchPodLabels := shared_helpers.GetStringMapFromLabelMap(logsAggregatorDeploymentLabels)
	serviceObj, err := kubernetesManager.CreateService(ctx, namespace, serviceName, serviceLabels, serviceAnnotations, matchPodLabels, serviceType, ports)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating service object for logs aggregator.")
	}

	return serviceObj, nil
}

func createLogsAggregatorConfigMap(
	ctx context.Context,
	namespace string,
	objAttrProvider object_attributes_provider.KubernetesLogsAggregatorObjectAttributesProvider,
	kubernetesManager *kubernetes_manager.KubernetesManager) (*apiv1.ConfigMap, error) {
	configMapAttrProvider, err := objAttrProvider.ForLogsAggregatorConfigMap()
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
			vectorConfigFileName: vectorConfigStr,
		},
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating config map for vector log aggregator config.")
	}

	return configMap, nil
}
