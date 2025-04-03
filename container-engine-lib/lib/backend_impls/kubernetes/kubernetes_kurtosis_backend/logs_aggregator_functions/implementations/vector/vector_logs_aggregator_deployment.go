package vector

import (
	"context"
	"time"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_value"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	retryInterval        = 1 * time.Second
	maxRetries           = 30
	preCleanNumReplicas  = 0
	postCleanNumReplicas = 1
)

type vectorLogsAggregatorResourcesManager struct{}

func NewVectorLogsAggregatorResourcesManager() *vectorLogsAggregatorResourcesManager {
	return &vectorLogsAggregatorResourcesManager{}
}

func (logsAggregator *vectorLogsAggregatorResourcesManager) CreateAndStart(
	ctx context.Context,
	logsListeningPortNum uint16,
	sinks logs_aggregator.Sinks,
	httpPortNumber uint16,
	engineNamespace string,
	objAttrsProvider object_attributes_provider.KubernetesObjectAttributesProvider,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (
	*apiv1.Service,
	*appsv1.Deployment,
	*apiv1.Namespace,
	*apiv1.ConfigMap,
	func(),
	error,
) {
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

	vectorConfigurationCreatorObj := createVectorConfigurationCreatorForKurtosis(logsListeningPortNum, httpPortNumber, sinks)

	configMap, removeConfigMapFunc, err := vectorConfigurationCreatorObj.CreateConfiguration(ctx, namespace.Name, logsAggregatorAttrProvider, kubernetesManager)
	if err != nil {
		return nil, nil, nil, nil, nil, stacktrace.Propagate(err, "An error occurred while trying to create config map for vector logs aggregator.")
	}
	shouldRemoveLogsAggregatorConfigMap := false
	defer func() {
		if shouldRemoveLogsAggregatorConfigMap {
			removeConfigMapFunc()
		}
	}()

	deployment, deploymentLabels, err := createLogsAggregatorDeployment(ctx, engineNamespace, namespace.Name, logsListeningPortNum, configMap.Name, logsAggregatorAttrProvider, kubernetesManager)
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

	service, err := createLogsAggregatorService(ctx, namespace.Name, logsListeningPortNum, deploymentLabels, logsAggregatorAttrProvider, kubernetesManager)
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

	if err = kubernetesManager.WaitForPodManagedByDeployment(ctx, deployment, maxRetries, retryInterval); err != nil {
		return nil, nil, nil, nil, nil, stacktrace.Propagate(err, "An error occurred waiting for active pod managed by logs aggregator deployment '%v'", deployment.Name)
	}

	removeLogsAggregatorFunc := func() {
		removeConfigMapFunc()
		removeDeploymentFunc()
		removeServiceFunc()
		removeNamespaceFunc()
	}

	shouldRemoveLogsAggregatorConfigMap = false
	shouldRemoveLogsAggregatorDeployment = false
	shouldRemoveLogsAggregatorService = false
	shouldRemoveLogsAggregatorNamespace = false
	return service, deployment, namespace, configMap, removeLogsAggregatorFunc, nil
}

func createLogsAggregatorDeployment(
	ctx context.Context,
	engineNamespace string,
	namespace string,
	logsListeningPort uint16,
	configMapName string,
	objAttrProvider object_attributes_provider.KubernetesLogsAggregatorObjectAttributesProvider,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (
	*appsv1.Deployment,
	map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue,
	error,
) {
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
			Args:       []string{"--config", vectorConfigFilePath},
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
			EnvFrom: nil,
			Env:     nil,
			Resources: apiv1.ResourceRequirements{
				Limits:   nil,
				Requests: nil,
				Claims:   nil,
			},
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
				{
					Name:             vectorDataDirVolumeName,
					ReadOnly:         false,
					MountPath:        vectorDataDirMountPath,
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
			Name:         kurtosisLogsVolumeName,
			VolumeSource: kubernetesManager.GetVolumeSourceForHostPath(kurtosisLogsMountPath),
		},
		{
			Name:         vectorDataDirVolumeName,
			VolumeSource: kubernetesManager.GetVolumeSourceForHostPath(vectorDataDirMountPath),
		},
	}

	affinity := &apiv1.Affinity{
		NodeAffinity: nil,
		PodAffinity: &apiv1.PodAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []apiv1.PodAffinityTerm{
				{
					LabelSelector: &v1.LabelSelector{
						// Always schedule the logs aggregator pods to run on the same node as the engine pods
						// they need to share a node's filesystem because aggregator writes to log files that engine reads from
						MatchLabels: map[string]string{
							// use resource label to match engine pods (which should only be one at any time)
							kubernetes_label_key.KurtosisResourceTypeKubernetesLabelKey.GetString(): label_value_consts.EngineKurtosisResourceTypeKubernetesLabelValue.GetString(),
						},
						MatchExpressions: nil,
					},
					Namespaces:        []string{engineNamespace},
					TopologyKey:       apiv1.LabelHostname,
					NamespaceSelector: nil,
				},
			},
			PreferredDuringSchedulingIgnoredDuringExecution: nil,
		},
		PodAntiAffinity: nil,
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
		affinity,
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
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (*apiv1.Service, error) {
	serviceAttrProvider, err := objAttrProvider.ForLogsAggregatorNamespace()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting logs aggregator namespace attributes provider.")
	}
	serviceName := serviceAttrProvider.GetName().GetString()
	serviceLabels := shared_helpers.GetStringMapFromLabelMap(serviceAttrProvider.GetLabels())
	serviceAnnotations := shared_helpers.GetStringMapFromAnnotationMap(serviceAttrProvider.GetAnnotations())

	// for now logs aggregator will only be reachable from within the cluster
	// this might need to changed later for exporting logs cluster
	serviceType := apiv1.ServiceTypeClusterIP

	ports := []apiv1.ServicePort{
		{
			Name:        "logs-collector-forwarding",
			Protocol:    "",
			AppProtocol: nil,
			Port:        int32(logListeningPort),
			TargetPort: intstr.IntOrString{
				IntVal: int32(logListeningPort),
				StrVal: "",
				Type:   0,
			},
			NodePort: 0,
		},
		{
			Name:        "api",
			Protocol:    "",
			AppProtocol: nil,
			Port:        int32(apiPort),
			TargetPort: intstr.IntOrString{
				IntVal: int32(apiPort),
				StrVal: "",
				Type:   0,
			},
			NodePort: 0,
		},
	}

	matchPodLabels := shared_helpers.GetStringMapFromLabelMap(logsAggregatorDeploymentLabels)
	serviceObj, err := kubernetesManager.CreateService(ctx, namespace, serviceName, serviceLabels, serviceAnnotations, matchPodLabels, serviceType, ports)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating service object for logs aggregator.")
	}

	return serviceObj, nil
}

func (vector *vectorLogsAggregatorResourcesManager) GetLogsBaseDirPath() string {
	return kurtosisLogsMountPath
}

func (vector *vectorLogsAggregatorResourcesManager) GetHTTPHealthCheckEndpointAndPort() (string, uint16) {
	return "/health", apiPort
}

// Clean cleans up the data directory created by vector to store buffer information, to do this:
// 1) scales down the vector logs aggregator deployment
// 2) creates a privileged pod with access to underlying nodes filesystem
// 3) removes vector data directory on node's filesystem
func (vector *vectorLogsAggregatorResourcesManager) Clean(ctx context.Context, logsAggregatorDeployment *appsv1.Deployment, kubernetesManager *kubernetes_manager.KubernetesManager) error {
	pods, err := kubernetesManager.GetPodsManagedByDeployment(ctx, logsAggregatorDeployment)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting pods managed by deployment '%v' in namespace '%v'.", logsAggregatorDeployment.Name, logsAggregatorDeployment.Namespace)
	}
	if len(pods) == 0 {
		return stacktrace.Propagate(err, "No pods found for logs aggregator deployment '%v' in namespace '%v'.", logsAggregatorDeployment.Name, logsAggregatorDeployment.Namespace)
	}
	if len(pods) > 1 {
		return stacktrace.Propagate(err, "More than one pod found for logs aggregator deployment '%v' in namespace '%v'.", logsAggregatorDeployment.Name, logsAggregatorDeployment.Namespace)
	}

	logrus.Debugf("Cleaning the vector logs aggregator deployment...")
	pod := pods[0]
	nodeName := pod.Spec.NodeName

	// scale deployment down to zero in order to unmount volume
	// need to do this in order to unmount data directory volume - otherwise removing directory from underlying nodes file system won't work since processes are accessing it
	if err := kubernetesManager.ScaleDeployment(ctx, logsAggregatorDeployment.Namespace, logsAggregatorDeployment.Name, preCleanNumReplicas); err != nil {
		return stacktrace.Propagate(err, "An error occurred scaling deployment '%v' in namespace '%v' to '%v'.", logsAggregatorDeployment.Name, logsAggregatorDeployment.Namespace, preCleanNumReplicas)
	}
	if err := kubernetesManager.WaitForPodTermination(ctx, pod.Namespace, pod.Name); err != nil {
		return stacktrace.Propagate(err, "An error occurred waiting for pod '%v' in namespace '%v' to terminate.", pod.Namespace, pod.Name)
	}

	err = kubernetesManager.RemoveDirPathFromNode(ctx, logsAggregatorDeployment.Namespace, nodeName, vectorDataDirMountPath)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred removing dir path '%v' from node '%v' via a pod in namespace '%v'.", vectorDataDirMountPath, nodeName, logsAggregatorDeployment.Namespace)
	}

	// scale up the deployment again
	if err := kubernetesManager.ScaleDeployment(ctx, logsAggregatorDeployment.Namespace, logsAggregatorDeployment.Name, postCleanNumReplicas); err != nil {
		return stacktrace.Propagate(err, "An error occurred scaling deployment '%v' in namespace '%v' to '%v'.", logsAggregatorDeployment.Name, logsAggregatorDeployment.Namespace, postCleanNumReplicas)
	}

	// before continuing, ensure logs aggregator is up again
	if err := kubernetesManager.WaitForPodManagedByDeployment(ctx, logsAggregatorDeployment, maxRetries, retryInterval); err != nil {
		return stacktrace.Propagate(err, "An error occurred waiting for a pod managed by deployment '%v' to become available.", logsAggregatorDeployment.Name)
	}

	logrus.Debugf("Successfully cleaned logs aggregator.")

	return nil
}
