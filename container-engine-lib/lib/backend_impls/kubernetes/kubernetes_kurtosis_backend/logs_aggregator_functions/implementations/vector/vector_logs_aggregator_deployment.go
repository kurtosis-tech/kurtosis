package vector

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_value"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/concurrent_writer"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"time"
)

const (
	retryInterval        = 1 * time.Second
	maxRetries           = 30
	successExitCode      = 0
	preCleanNumReplicas  = 0
	postCleanNumReplicas = 1
)

type vectorLogsAggregatorDeployment struct{}

func NewVectorLogsAggregatorDeployment() *vectorLogsAggregatorDeployment {
	return &vectorLogsAggregatorDeployment{}
}

func (logsAggregator *vectorLogsAggregatorDeployment) CreateAndStart(
	ctx context.Context,
	logsListeningPortNum uint16,
	engineNamespace string,
	objAttrsProvider object_attributes_provider.KubernetesObjectAttributesProvider,
	kubernetesManager *kubernetes_manager.KubernetesManager) (
	*apiv1.Service,
	*appsv1.Deployment,
	*apiv1.Namespace,
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

	configMap, err := createLogsAggregatorConfigMap(ctx, namespace.Name, logsListeningPortNum, logsAggregatorAttrProvider, kubernetesManager)
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

	if err = waitForPodManagedByDeployment(ctx, deployment, kubernetesManager); err != nil {
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
					TopologyKey:       "kubernetes.io/hostname",
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

func createLogsAggregatorConfigMap(
	ctx context.Context,
	namespace string,
	logListeningPortNum uint16,
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
			vectorConfigFileName: fmt.Sprintf(
				vectorConfigFmtStr,
				vectorDataDirMountPath,
				apiPortStr,
				logListeningPortNum,
				kurtosisLogsMountPath),
		},
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating config map for vector log aggregator config.")
	}

	return configMap, nil
}

func waitForPodManagedByDeployment(ctx context.Context, logsAggregatorDeployment *appsv1.Deployment, kubernetesManager *kubernetes_manager.KubernetesManager) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(maxRetries)*retryInterval)
	defer cancel()

	ticker := time.NewTicker(retryInterval)
	defer ticker.Stop()
	for attempt := 0; attempt < maxRetries; attempt++ {
		select {
		case <-timeoutCtx.Done():
			return stacktrace.NewError(
				"Timeout waiting for a pod managed by logs aggregator deployment '%s' to come online",
				logsAggregatorDeployment.Name,
			)
		case <-ticker.C:
			pods, err := kubernetesManager.GetPodsManagedByDeployment(ctx, logsAggregatorDeployment)
			if err != nil {
				return stacktrace.Propagate(err, "An error occurred getting pods managed by logs aggregator deployment'%v'", logsAggregatorDeployment.Name)
			}
			if len(pods) > 0 && len(pods[0].Status.ContainerStatuses) > 0 && pods[0].Status.ContainerStatuses[0].Ready {
				// found a pod with a running vector container
				return nil
			}
		}
	}
	return stacktrace.NewError(
		"Exceeded max retries (%d) waiting for a pod managed by deployment '%s' to come online",
		maxRetries, logsAggregatorDeployment.Name,
	)
}

func (vector *vectorLogsAggregatorDeployment) GetLogsBaseDirPath() string {
	return kurtosisLogsMountPath
}

func (vector *vectorLogsAggregatorDeployment) GetHTTPHealthCheckEndpointAndPort() (string, uint16) {
	return "/health", apiPort
}

// Clean cleans up the data directory created by vector to store buffer information, to do this:
// 1) scales down the vector logs aggregator deployment
// 2) creates a privileged pod with access to underlying nodes filesystem
// 3) removes vector data directory on node's filesystem
func (vector *vectorLogsAggregatorDeployment) Clean(ctx context.Context, logsAggregatorDeployment *appsv1.Deployment, kubernetesManager *kubernetes_manager.KubernetesManager) error {
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

	logrus.Infof("Cleaning the vector logs aggregator deployment...")
	pod := pods[0]
	nodeName := pod.Spec.NodeName

	// scale deployment down to zero in order to unmount volume
	if err := kubernetesManager.ScaleDeployment(ctx, logsAggregatorDeployment.Namespace, logsAggregatorDeployment.Name, preCleanNumReplicas); err != nil {
		return stacktrace.Propagate(err, "An error occurred scaling deployment '%v' in namespace '%v' to '%v'.", logsAggregatorDeployment.Name, logsAggregatorDeployment.Namespace, preCleanNumReplicas)
	}
	if err := kubernetesManager.WaitForPodTermination(ctx, pod.Namespace, pod.Name); err != nil {
		return stacktrace.Propagate(err, "An error occurred waiting for pod '%v' in namespace '%v' to terminate.", pod.Namespace, pod.Name)
	}

	// rm the vector directory from the node using a privileged pod
	removeContainerName := "remove-vector-data-container"
	// pod needs to be privileged to access host filesystem
	isPrivileged := true
	hasHostPidAccess := true
	hasHostNetworkAcess := true
	removePodMountPath := fmt.Sprintf("host%v", vectorDataDirMountPath)
	removePod, err := kubernetesManager.CreatePod(ctx,
		logsAggregatorDeployment.Namespace,
		"remove-vector-data-pod",
		nil,
		nil,
		nil,
		[]apiv1.Container{
			{
				Name:  removeContainerName,
				Image: "busybox",
				Command: []string{
					"sh",
					"-c",
					"sleep 10000000s",
				},
				Args:       nil,
				WorkingDir: "",
				Ports:      nil,
				EnvFrom:    nil,
				Env:        nil,
				Resources: apiv1.ResourceRequirements{
					Limits:   nil,
					Requests: nil,
					Claims:   nil,
				},
				ResizePolicy: nil,
				VolumeMounts: []apiv1.VolumeMount{
					{
						Name:             vectorDataDirVolumeName,
						ReadOnly:         false,
						MountPath:        removePodMountPath,
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
				SecurityContext: &apiv1.SecurityContext{
					Privileged: &isPrivileged,
				},
				Stdin:     false,
				StdinOnce: false,
				TTY:       false,
			},
		},
		[]apiv1.Volume{
			{
				Name:         vectorDataDirVolumeName,
				VolumeSource: kubernetesManager.GetVolumeSourceForHostPath(vectorDataDirMountPath),
			},
		},
		"",
		"",
		nil,
		map[string]string{
			"kubernetes.io/hostname": nodeName,
		},
		hasHostPidAccess,
		hasHostNetworkAcess,
	)
	defer func() {
		// Don't block on removing this remove directory pod because this can take a while sometimes in k8s
		go func() {
			removeCtx := context.Background()
			if removePod != nil {
				err := kubernetesManager.RemovePod(removeCtx, removePod)
				if err != nil {
					logrus.Warnf("Attempted to remove pod '%v' in namespace '%v' but an error occurred:\n%v", removePod.Name, logsAggregatorDeployment.Namespace, err.Error())
					logrus.Warn("You may have to remove this pod manually.")
				}
			}
		}()
	}()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating pod '%v' in namespace '%v'.", "availabilityChecker", removePod)
	}

	cleanCmd := []string{"rm", "-rf", removePodMountPath}
	output := &bytes.Buffer{}
	concurrentWriter := concurrent_writer.NewConcurrentWriter(output)
	resultExitCode, err := kubernetesManager.RunExecCommand(
		removePod.Namespace,
		removePod.Name,
		removeContainerName,
		cleanCmd,
		concurrentWriter,
		concurrentWriter,
	)
	logrus.Debugf("Output of clean '%v': %v, exit code: %v", cleanCmd, output.String(), resultExitCode)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred running exec command '%v' on pod '%v' in namespace '%v'.", cleanCmd, pod.Name, pod.Namespace)
	}
	if resultExitCode != successExitCode {
		return stacktrace.NewError("Running exec command '%v' on pod '%v' in namespace '%v' returned a non-%v exit code: '%v'.", cleanCmd, pod.Name, pod.Namespace, successExitCode, resultExitCode)
	}
	if output.String() != "" {
		return stacktrace.NewError("Expected empty output from running exec command '%v' but instead retrieved output string '%v'", cleanCmd, output.String())
	}

	// scale up the deployment again
	if err := kubernetesManager.ScaleDeployment(ctx, logsAggregatorDeployment.Namespace, logsAggregatorDeployment.Name, postCleanNumReplicas); err != nil {
		return stacktrace.Propagate(err, "An error occurred scaling deployment '%v' in namespace '%v' to '%v'.", logsAggregatorDeployment.Name, logsAggregatorDeployment.Namespace, postCleanNumReplicas)
	}

	logrus.Debugf("Successfully cleaned logs aggregator.")

	return nil
}
