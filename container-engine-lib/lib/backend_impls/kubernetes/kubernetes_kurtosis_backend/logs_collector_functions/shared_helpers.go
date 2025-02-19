package logs_collector_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_resource_collectors"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"net"
	"time"
)

const (
	maxAvailabilityCheckRetries     = 10
	timeToWaitBetweenChecksDuration = 500 * time.Millisecond
)

func getLogsCollectorObjAndResourcesForCluster(ctx context.Context, kubernetesManager *kubernetes_manager.KubernetesManager) (*logs_collector.LogsCollector, *logsCollectorKubernetesResources, error) {
	kubernetesResources, err := getLogsCollectorKubernetesResourcesForCluster(ctx, kubernetesManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes resources for logs collectors.")
	}

	obj, err := getLogsCollectorsObjectFromKubernetesResources(ctx, kubernetesManager, kubernetesResources)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting logs collector object from kubernetes resources.")
	}
	return obj, kubernetesResources, nil
}

func getLogsCollectorKubernetesResourcesForCluster(ctx context.Context, kubernetesManager *kubernetes_manager.KubernetesManager) (*logsCollectorKubernetesResources, error) {
	resourceTypeLabelKeyStr := kubernetes_label_key.KurtosisResourceTypeKubernetesLabelKey.GetString()
	logsCollectorResourceTypeLabelValStr := label_value_consts.LogsCollectorKurtosisResourceTypeKubernetesLabelValue.GetString()
	logsCollectorDaemonSetSearchLabels := map[string]string{
		kubernetes_label_key.AppIDKubernetesLabelKey.GetString(): label_value_consts.AppIDKubernetesLabelValue.GetString(),
		resourceTypeLabelKeyStr:                                  logsCollectorResourceTypeLabelValStr,
		// could retrieve the logs collector by the logs collector guid if we added guid labels, but for now just retrieve by resource type
	}

	logsCollectorNamespaces, err := kubernetes_resource_collectors.CollectMatchingNamespaces(ctx, kubernetesManager, logsCollectorDaemonSetSearchLabels, resourceTypeLabelKeyStr, map[string]bool{logsCollectorResourceTypeLabelValStr: true})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting namespace for logs collector.")
	}
	var namespace *apiv1.Namespace
	if logsCollectorNamespaceForLabel, found := logsCollectorNamespaces[logsCollectorResourceTypeLabelValStr]; found {
		if len(logsCollectorNamespaceForLabel) > 1 {
			return nil, stacktrace.NewError(
				"Expected at most one namespaces for the logs collector but found '%v'",
				len(logsCollectorNamespaceForLabel),
			)
		}
		if len(logsCollectorNamespaceForLabel) == 0 {
			// if no namespace for logs collector, assume it doesn't exist at all
			return &logsCollectorKubernetesResources{
				daemonSet: nil,
				configMap: nil,
				namespace: nil,
			}, nil
		} else {
			namespace = logsCollectorNamespaceForLabel[0]
		}
	} else {
		return &logsCollectorKubernetesResources{
			daemonSet: nil,
			configMap: nil,
			namespace: nil,
		}, nil
	}

	logsCollectorCfgConfigMaps, err := kubernetes_resource_collectors.CollectMatchingConfigMaps(
		ctx,
		kubernetesManager,
		namespace.Name,
		logsCollectorDaemonSetSearchLabels,
		resourceTypeLabelKeyStr,
		map[string]bool{
			logsCollectorResourceTypeLabelValStr: true,
		})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting config map for logs collector in namespace '%v'", namespace.Name)
	}
	var configMap *apiv1.ConfigMap
	if configMapForForLabel, found := logsCollectorCfgConfigMaps[logsCollectorResourceTypeLabelValStr]; found {
		if len(configMapForForLabel) > 1 {
			return nil, stacktrace.NewError(
				"Expected at most one logs collector config maps in namespace '%v' for logs collector but found '%v'",
				namespace.Name,
				len(configMapForForLabel),
			)
		}
		configMap = configMapForForLabel[0]
	}

	daemonSets, err := kubernetes_resource_collectors.CollectMatchingDaemonSets(
		ctx,
		kubernetesManager,
		namespace.Name,
		logsCollectorDaemonSetSearchLabels,
		resourceTypeLabelKeyStr,
		map[string]bool{
			logsCollectorResourceTypeLabelValStr: true,
		})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting config map for logs collector in namespace '%v'", namespace)
	}
	var daemonSet *v1.DaemonSet
	if logsCollectorDaemonSetsForLabel, found := daemonSets[logsCollectorResourceTypeLabelValStr]; found {
		if len(logsCollectorDaemonSetsForLabel) > 1 {
			return nil, stacktrace.NewError(
				"Expected at most one logs collector daemon set in namespace '%v' for logs collector but found '%v'",
				namespace.Name,
				len(logsCollectorDaemonSetsForLabel),
			)
		}
		if len(logsCollectorDaemonSetsForLabel) == 0 {
			daemonSet = nil
		} else {
			daemonSet = logsCollectorDaemonSetsForLabel[0]
		}
	}

	logsCollectorKubernetesResources := &logsCollectorKubernetesResources{
		daemonSet: daemonSet,
		configMap: configMap,
		namespace: namespace,
	}

	return logsCollectorKubernetesResources, nil
}

func getLogsCollectorsObjectFromKubernetesResources(ctx context.Context, kubernetesManager *kubernetes_manager.KubernetesManager, logsCollectorKubernetesResources *logsCollectorKubernetesResources) (*logs_collector.LogsCollector, error) {
	if logsCollectorKubernetesResources.namespace == nil || logsCollectorKubernetesResources.daemonSet == nil || logsCollectorKubernetesResources.configMap == nil {
		// if any resources not found for logs collector, don't return an object
		return nil, nil
	}

	var (
		logsCollectorStatus container.ContainerStatus
		privateIpAddr       net.IP
		bridgeNetworkIpAddr net.IP
		err                 error
	)

	logsCollectorStatus, err = getLogsCollectorStatus(ctx, kubernetesManager, logsCollectorKubernetesResources.daemonSet)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the status of the logs collector.")
	}
	logrus.Infof("Log collector has status %v", logsCollectorStatus)

	// daemon sets don't have an entrypoint so leave these blank for now
	// if at some point, we need to access fluent bit log collectors via IP in some way, can create an entrypoint that allows
	// accessing log collectors via IP
	privateIpAddr = net.IP{}
	bridgeNetworkIpAddr = net.IP{}

	dummyPortSpecOne, err := port_spec.NewPortSpec(0, port_spec.TransportProtocol(0), "HTTP", nil, "")
	if err != nil {
		return nil, err
	}

	dummyPortSpecTwo, err := port_spec.NewPortSpec(0, port_spec.TransportProtocol(0), "HTTP", nil, "")
	if err != nil {
		return nil, err
	}

	return logs_collector.NewLogsCollector(
		logsCollectorStatus,
		privateIpAddr,
		bridgeNetworkIpAddr,
		dummyPortSpecOne,
		dummyPortSpecTwo,
	), nil
}

// TODO: container status is outdated for k8s pods (see TODO in shared_helpers.GetContainerStatusFromPod)
// in the meantime logs collector status is container.ContainerStatus_Running if all pods managed by the logs collector DaemonSet are running
// if one is failing/or stopped, the logs collector is to considered to be stopped
func getLogsCollectorStatus(ctx context.Context, kubernetesManager *kubernetes_manager.KubernetesManager, logsCollectorDaemonSet *v1.DaemonSet) (container.ContainerStatus, error) {
	logsCollectorPods, err := kubernetesManager.GetPodsManagedByDaemonSet(ctx, logsCollectorDaemonSet)
	if err != nil {
		return container.ContainerStatus_Stopped, stacktrace.Propagate(err, "An error occurred getting pods managed by logs collector daemon set '%v'.", logsCollectorDaemonSet.Name)
	}
	if len(logsCollectorPods) < 1 {
		// if there are no pods associated with logs collector daemon set, assume something is wrong and err
		return container.ContainerStatus_Stopped, stacktrace.NewError("No pods managed by logs collector daemon set were found. There should be at least one. This is likely a bug in Kurtosis.")
	}

	logsCollectorStatus := container.ContainerStatus_Running
	for _, pod := range logsCollectorPods {
		podStatus, err := shared_helpers.GetContainerStatusFromPod(pod)
		if err != nil {
			return container.ContainerStatus_Stopped, stacktrace.Propagate(err, "An error occurred retrieving container status for a pod managed by logs collectors collector daemon set '%v' with name: %v\n", logsCollectorDaemonSet.Name, pod.Name)
		}

		switch podStatus {
		case container.ContainerStatus_Running:
			continue
		case container.ContainerStatus_Stopped:
			logsCollectorStatus = container.ContainerStatus_Stopped
			return logsCollectorStatus, nil
		}
	}

	return logsCollectorStatus, nil
}

func waitForLogsCollectorAvailability(
	ctx context.Context,
	logsCollectorHttpPortNumber uint16,
	k8sResources *logsCollectorKubernetesResources,
	kubernetesManager *kubernetes_manager.KubernetesManager) error {
	logsCollectorDaemonSet := k8sResources.daemonSet
	pods, err := kubernetesManager.GetPodsManagedByDaemonSet(ctx, k8sResources.daemonSet)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting pods managed by logs collector daemon set '%v'", logsCollectorDaemonSet.Name)
	}

	// this port spec represents the http port that each log collector container (on each pod managed by the daemon set) wll have a port exposed on
	httpPortSpec, err := port_spec.NewPortSpec(logsCollectorHttpPortNumber, port_spec.TransportProtocol_TCP, httpProtocolStr, noWait, emptyUrl)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred creating the log collectors public HTTP port spec object using number '%v' and protocol '%v'",
			logsCollectorHttpPortNumber,
			port_spec.TransportProtocol_TCP,
		)
	}
	for _, pod := range pods {
		if len(pod.Spec.Containers) < 1 {
			return stacktrace.NewError("Pod '%v' managed by logs collector daemon set '%v' doesn't have any containers associated with it. There should be at least one container.", pod.Name, logsCollectorDaemonSet.Name)
		}

		// NOTE: ideally we'd actually curl the health endpoint of the fluent bit container (like we do for Docker)
		// this part of code runs on users machine logs collector isn't exposed so we exec onto the pod containers - which may not have curl on them, hence netstat check
		if err = shared_helpers.WaitForPortAvailabilityUsingNetstat(kubernetesManager, k8sResources.namespace.Name, pod.Name, pod.Spec.Containers[0].Name, httpPortSpec, maxAvailabilityCheckRetries, timeToWaitBetweenChecksDuration); err != nil {
			return stacktrace.Propagate(err, "An error occurred while checking for availability of pod '%v' managed by logs collector daemon set '%v'", pod.Name, logsCollectorDaemonSet.Name)
		}
	}

	return nil

}
