package logs_aggregator_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_resource_collectors"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
	"github.com/kurtosis-tech/stacktrace"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"net"
)

const (
// maxAvailabilityCheckRetries     = 30
// timeToWaitBetweenChecksDuration = 1 * time.Second
)

func getLogsAggregatorObjAndResourcesForCluster(ctx context.Context, kubernetesManager *kubernetes_manager.KubernetesManager) (*logs_aggregator.LogsAggregator, *logsAggregatorKubernetesResources, error) {
	kubernetesResources, err := getLogsAggregatorKubernetesResourcesForCluster(ctx, kubernetesManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes resources for logs aggregator.")
	}

	obj, err := getLogsAggregatorObjectFromKubernetesResources(ctx, kubernetesManager, kubernetesResources)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting logs aggregator object from kubernetes resources.")
	}
	return obj, kubernetesResources, nil
}

func getLogsAggregatorKubernetesResourcesForCluster(ctx context.Context, kubernetesManager *kubernetes_manager.KubernetesManager) (*logsAggregatorKubernetesResources, error) {
	resourceTypeLabelKeyStr := kubernetes_label_key.KurtosisResourceTypeKubernetesLabelKey.GetString()
	logsAggregatorResourceTypeLabelValStr := label_value_consts.LogsAggregatorKurtosisResourceTypeKubernetesLabelValue.GetString()
	logsAggregatorDeploymentSearchLabels := map[string]string{
		kubernetes_label_key.AppIDKubernetesLabelKey.GetString(): label_value_consts.AppIDKubernetesLabelValue.GetString(),
		resourceTypeLabelKeyStr:                                  logsAggregatorResourceTypeLabelValStr,
	}

	logsAggregatorNamespaces, err := kubernetes_resource_collectors.CollectMatchingNamespaces(
		ctx,
		kubernetesManager,
		logsAggregatorDeploymentSearchLabels,
		resourceTypeLabelKeyStr,
		map[string]bool{
			logsAggregatorResourceTypeLabelValStr: true,
		})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting namespace for logs aggregator.")
	}
	var namespace *apiv1.Namespace
	if logsAggregatorNamespaceForLabel, found := logsAggregatorNamespaces[logsAggregatorResourceTypeLabelValStr]; found {
		if len(logsAggregatorNamespaceForLabel) > 1 {
			return nil, stacktrace.NewError(
				"Expected at most one namespaces for the logs aggregator but found '%v'",
				len(logsAggregatorNamespaceForLabel),
			)
		}
		if len(logsAggregatorNamespaceForLabel) == 0 {
			// if no namespace for logs aggregator, assume it doesn't exist at all
			return &logsAggregatorKubernetesResources{
				deployment: nil,
				service:    nil,
				configMap:  nil,
				namespace:  nil,
			}, nil
		}
		namespace = logsAggregatorNamespaceForLabel[0]
	} else {
		return &logsAggregatorKubernetesResources{
			deployment: nil,
			service:    nil,
			configMap:  nil,
			namespace:  nil,
		}, nil
	}

	configMaps, err := kubernetes_resource_collectors.CollectMatchingConfigMaps(
		ctx,
		kubernetesManager,
		namespace.Name,
		logsAggregatorDeploymentSearchLabels,
		resourceTypeLabelKeyStr,
		map[string]bool{
			logsAggregatorResourceTypeLabelValStr: true,
		})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting config map for logs aggregator in namespace '%v'", namespace)
	}
	var configMap *apiv1.ConfigMap
	if logsAggregatorConfigMapsForLabel, found := configMaps[logsAggregatorResourceTypeLabelValStr]; found {
		if len(logsAggregatorConfigMapsForLabel) > 1 {
			return nil, stacktrace.NewError(
				"Expected at most one logs aggregator config map in namespace '%v' for logs aggregator but found '%v'",
				namespace.Name,
				len(logsAggregatorConfigMapsForLabel),
			)
		}
		if len(logsAggregatorConfigMapsForLabel) == 0 {
			configMap = nil
		} else {
			configMap = logsAggregatorConfigMapsForLabel[0]
		}
	}

	logsAggregatorConfigServices, err := kubernetes_resource_collectors.CollectMatchingServices(
		ctx,
		kubernetesManager,
		namespace.Name,
		logsAggregatorDeploymentSearchLabels,
		resourceTypeLabelKeyStr,
		map[string]bool{
			logsAggregatorResourceTypeLabelValStr: true,
		})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting service for logs aggregator in namespace '%v'", namespace.Name)
	}
	var service *apiv1.Service
	if serviceForForLabel, found := logsAggregatorConfigServices[logsAggregatorResourceTypeLabelValStr]; found {
		if len(serviceForForLabel) > 1 {
			return nil, stacktrace.NewError(
				"Expected at most one logs aggregator services in namespace '%v' for logs aggregator but found '%v'",
				namespace.Name,
				len(serviceForForLabel),
			)
		}
		if len(serviceForForLabel) == 0 {
			service = nil
		} else {
			service = serviceForForLabel[0]
		}
	}

	deployments, err := kubernetes_resource_collectors.CollectMatchingDeployments(
		ctx,
		kubernetesManager,
		namespace.Name,
		logsAggregatorDeploymentSearchLabels,
		resourceTypeLabelKeyStr,
		map[string]bool{
			logsAggregatorResourceTypeLabelValStr: true,
		})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting deployments for logs aggregator in namespace '%v'", namespace)
	}
	var deployment *appsv1.Deployment
	if logsAggregatorDeploymentsForLabel, found := deployments[logsAggregatorResourceTypeLabelValStr]; found {
		if len(logsAggregatorDeploymentsForLabel) > 1 {
			return nil, stacktrace.NewError(
				"Expected at most one logs aggregator deployment in namespace '%v' for logs aggregator but found '%v'",
				namespace.Name,
				len(logsAggregatorDeploymentsForLabel),
			)
		}
		if len(logsAggregatorDeploymentsForLabel) == 0 {
			deployment = nil
		} else {
			deployment = logsAggregatorDeploymentsForLabel[0]
		}
	}

	logsAggregatorKubernetesResources := &logsAggregatorKubernetesResources{
		deployment: deployment,
		service:    service,
		configMap:  configMap,
		namespace:  namespace,
	}

	return logsAggregatorKubernetesResources, nil
}

// getLogsAggregatorsObjectFromKubernetesResources returns a logs aggregator object if any only if all kubernetes resources required for the logs aggregator exists
// otherwise returns nil object or error
func getLogsAggregatorObjectFromKubernetesResources(ctx context.Context, kubernetesManager *kubernetes_manager.KubernetesManager, logsAggregatorKubernetesResources *logsAggregatorKubernetesResources) (*logs_aggregator.LogsAggregator, error) {
	if logsAggregatorKubernetesResources.namespace == nil || logsAggregatorKubernetesResources.deployment == nil || logsAggregatorKubernetesResources.configMap == nil && logsAggregatorKubernetesResources.service == nil {
		// if any resources not found for logs collector, don't return an object
		return nil, nil
	}

	var (
		logsAggregatorStatus container.ContainerStatus
		privateIpAddr        net.IP
		logsListeningPort    uint16
	)

	logsAggregatorStatus, err := getLogsAggregatorStatus(ctx, kubernetesManager, logsAggregatorKubernetesResources.deployment)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the status of the logs aggregator.")
	}

	privateIpAddr = net.ParseIP(logsAggregatorKubernetesResources.service.Spec.ClusterIP)
	logsListeningPort = defaultLogsListeningPortNum

	return logs_aggregator.NewLogsAggregator(
		logsAggregatorStatus,
		privateIpAddr,
		logsListeningPort,
	), nil
}

// TODO: container status is outdated for k8s pods (see TODO in shared_helpers.GetContainerStatusFromPod)
// in the meantime logs aggregator status is container.ContainerStatus_Running if all pods managed by the logs aggregator Deployment are running
// if one is failing/or stopped, the logs aggregator is to considered to be stopped
func getLogsAggregatorStatus(ctx context.Context, kubernetesManager *kubernetes_manager.KubernetesManager, logsAggregatorDeployment *appsv1.Deployment) (container.ContainerStatus, error) {
	logsAggregatorPods, err := kubernetesManager.GetPodsManagedByDeployment(ctx, logsAggregatorDeployment)
	if err != nil {
		return container.ContainerStatus_Stopped, stacktrace.Propagate(err, "An error occurred getting pods managed by logs aggregator deployment '%v'.", logsAggregatorDeployment.Name)
	}
	if len(logsAggregatorPods) < 1 {
		// if there are no pods associated with logs aggregator deployment, assume something is wrong and err
		return container.ContainerStatus_Stopped, stacktrace.NewError("No pods managed by logs aggregator deployment were found. There should be exactly one. This is likely a bug in Kurtosis.")
	}
	if len(logsAggregatorPods) > 1 {
		// if there is more than one pod associated with logs aggregator deployment, assume something is wrong and err
		return container.ContainerStatus_Stopped, stacktrace.NewError("No pods managed by logs aggregator deployment were found. There should be exactly one. This is likely a bug in Kurtosis.")
	}

	pod := logsAggregatorPods[0]
	podStatus, err := shared_helpers.GetContainerStatusFromPod(pod)
	if err != nil {
		return container.ContainerStatus_Stopped, stacktrace.Propagate(err, "An error occurred retrieving container status for a pod managed by logs aggregator deployment '%v' with name: %v\n", logsAggregatorDeployment.Name, pod.Name)
	}
	return podStatus, nil
}

//
//func waitForLogsAggregatorAvailability(
//	ctx context.Context,
//	k8sResources *logsAggregatorKubernetesResources,
//	kubernetesManager *kubernetes_manager.KubernetesManager) error {
//	logsAggregatorDeployment := k8sResources.deployment
//	pods, err := kubernetesManager.GetPodsManagedByDeployment(ctx, k8sResources.deployment)
//	if err != nil {
//		return stacktrace.Propagate(err, "An error occurred getting pods managed by logs aggregator daemon set '%v'", logsAggregatorDeployment.Name)
//	}
//	if len(pods) < 1 {
//		return stacktrace.NewError("No pods managed by logs aggregator deployment were found. There should be exactly one. This is likely a bug in Kurtosis.")
//	}
//	if len(pods) > 1 {
//		return stacktrace.NewError("No pods managed by logs aggregator deployment were found. There should be exactly one. This is likely a bug in Kurtosis.")
//	}
//
//	httpPortSpec, err := port_spec.NewPortSpec(defaultLogsListeningPortNum, port_spec.TransportProtocol_TCP, "http", nil, "")
//	if err != nil {
//		return stacktrace.Propagate(
//			err,
//			"An error occurred creating the log aggregator public HTTP port spec object using number '%v' and protocol '%v'",
//			defaultLogsListeningPortNum,
//			port_spec.TransportProtocol_TCP,
//		)
//	}
//	pod := pods[0]
//	if len(pod.Spec.Containers) < 1 {
//		return stacktrace.NewError("Pod '%v' managed by logs aggregator deployment '%v' doesn't have any containers associated with it. There should be at least one container.", pod.Name, logsAggregatorDeployment.Name)
//	}
//
//	vectorContainerName := pod.Spec.Containers[0].Name
//	if err = shared_helpers.WaitForPortAvailabilityUsingNetstat(
//		kubernetesManager,
//		k8sResources.namespace.Name,
//		pod.Name,
//		vectorContainerName,
//		httpPortSpec,
//		maxAvailabilityCheckRetries,
//		timeToWaitBetweenChecksDuration); err != nil {
//		return stacktrace.Propagate(err, "An error occurred while checking for availability of pod '%v' managed by logs aggregator deployment '%v'", pod.Name, logsAggregatorDeployment.Name)
//	}
//
//	return nil
//}
