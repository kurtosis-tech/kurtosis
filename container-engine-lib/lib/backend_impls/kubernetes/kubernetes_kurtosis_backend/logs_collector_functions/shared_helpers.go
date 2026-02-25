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
	rbacv1 "k8s.io/api/rbac/v1"
	"net"
	"time"
)

const (
	maxAvailabilityChecksRetries                = 30
	timeToWaitBetweenAvailabilityChecksDuration = 1 * time.Second
	maxTriesToWaitForNamespaceRemoval           = 30
	timeToWaitBetweenNamespaceRemovalChecks     = 1 * time.Second
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
				daemonSet:          nil,
				configMap:          nil,
				namespace:          nil,
				serviceAccount:     nil,
				clusterRoleBinding: nil,
				clusterRole:        nil,
			}, nil
		}
		namespace = logsCollectorNamespaceForLabel[0]
	} else {
		return &logsCollectorKubernetesResources{
			daemonSet:          nil,
			configMap:          nil,
			namespace:          nil,
			serviceAccount:     nil,
			clusterRoleBinding: nil,
			clusterRole:        nil,
		}, nil
	}

	logsCollectorConfigClusterRoles, err := kubernetes_resource_collectors.CollectMatchingClusterRoles(
		ctx,
		kubernetesManager,
		logsCollectorDaemonSetSearchLabels,
		resourceTypeLabelKeyStr,
		map[string]bool{
			logsCollectorResourceTypeLabelValStr: true,
		})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting cluster role for logs collector in namespace '%v'", namespace.Name)
	}
	var clusterRole *rbacv1.ClusterRole
	if clusterRoleForLabel, found := logsCollectorConfigClusterRoles[logsCollectorResourceTypeLabelValStr]; found {
		if len(clusterRoleForLabel) > 1 {
			return nil, stacktrace.NewError(
				"Expected at most one logs collector cluster role in namespace '%v' for logs collector but found '%v'",
				namespace.Name,
				len(clusterRoleForLabel),
			)
		}
		if len(clusterRoleForLabel) == 0 {
			clusterRole = nil
		} else {
			clusterRole = clusterRoleForLabel[0]
		}
	}

	logsCollectorConfigClusterRoleBindings, err := kubernetes_resource_collectors.CollectMatchingClusterRoleBindings(
		ctx,
		kubernetesManager,
		logsCollectorDaemonSetSearchLabels,
		resourceTypeLabelKeyStr,
		map[string]bool{
			logsCollectorResourceTypeLabelValStr: true,
		})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting cluster role binding for logs collector in namespace '%v'", namespace.Name)
	}
	var cluseterRoleBinding *rbacv1.ClusterRoleBinding
	if clusterRoleBindingsForLabel, found := logsCollectorConfigClusterRoleBindings[logsCollectorResourceTypeLabelValStr]; found {
		if len(clusterRoleBindingsForLabel) > 1 {
			return nil, stacktrace.NewError(
				"Expected at most one logs collector cluster role bindings in namespace '%v' for logs collector but found '%v'",
				namespace.Name,
				len(clusterRoleBindingsForLabel),
			)
		}
		if len(clusterRoleBindingsForLabel) == 0 {
			cluseterRoleBinding = nil
		} else {
			cluseterRoleBinding = clusterRoleBindingsForLabel[0]
		}
	}

	logsCollectorConfigServiceAccounts, err := kubernetes_resource_collectors.CollectMatchingServiceAccounts(
		ctx,
		kubernetesManager,
		namespace.Namespace,
		logsCollectorDaemonSetSearchLabels,
		resourceTypeLabelKeyStr,
		map[string]bool{
			logsCollectorResourceTypeLabelValStr: true,
		})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting service accounts for logs collector in namespace '%v'", namespace.Name)
	}
	var serviceAccount *apiv1.ServiceAccount
	if serviceAccountForLabel, found := logsCollectorConfigServiceAccounts[logsCollectorResourceTypeLabelValStr]; found {
		if len(serviceAccountForLabel) > 1 {
			return nil, stacktrace.NewError(
				"Expected at most one logs collector service account in namespace '%v' for logs collector but found '%v'",
				namespace.Name,
				len(serviceAccountForLabel),
			)
		}
		if len(serviceAccountForLabel) == 0 {
			serviceAccount = nil
		} else {
			serviceAccount = serviceAccountForLabel[0]
		}
	}

	logsCollectorConfigConfigMaps, err := kubernetes_resource_collectors.CollectMatchingConfigMaps(
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
	if configMapForForLabel, found := logsCollectorConfigConfigMaps[logsCollectorResourceTypeLabelValStr]; found {
		if len(configMapForForLabel) > 1 {
			return nil, stacktrace.NewError(
				"Expected at most one logs collector config maps in namespace '%v' for logs collector but found '%v'",
				namespace.Name,
				len(configMapForForLabel),
			)
		}
		if len(configMapForForLabel) == 0 {
			configMap = nil
		} else {
			configMap = configMapForForLabel[0]
		}
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
		daemonSet:          daemonSet,
		configMap:          configMap,
		namespace:          namespace,
		clusterRole:        clusterRole,
		clusterRoleBinding: cluseterRoleBinding,
		serviceAccount:     serviceAccount,
	}

	return logsCollectorKubernetesResources, nil
}

// getLogsCollectorsObjectFromKubernetesResources returns a logs collector object if any only if all kubernetes resources required for the logs collector exists
// otherwise returns nil object or error
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

	// no way to access from logs collector daemon set from outside the cluster so leave these blank for now
	// if at some point, we need to access fluent bit log collectors via IP in some way, can create an entrypoint that allows
	// accessing log collectors via IP
	privateIpAddr = net.IP{}
	bridgeNetworkIpAddr = net.IP{}
	var privateHttpPortSpec *port_spec.PortSpec
	var privateTcpPortSpec *port_spec.PortSpec

	return logs_collector.NewLogsCollector(
		logsCollectorStatus,
		privateIpAddr,
		bridgeNetworkIpAddr,
		privateTcpPortSpec,
		privateHttpPortSpec,
	), nil
}

// TODO: container status is outdated for k8s pods (see TODO in shared_helpers.GetContainerStatusFromPod)
// logs collector status is container.ContainerStatus_Running if at least one pod managed by the logs collector DaemonSet is running
// if some pods are stopped while others are running, a warning is logged to surface partial degradation
func getLogsCollectorStatus(ctx context.Context, kubernetesManager *kubernetes_manager.KubernetesManager, logsCollectorDaemonSet *v1.DaemonSet) (container.ContainerStatus, error) {
	logsCollectorPods, err := kubernetesManager.GetPodsManagedByDaemonSet(ctx, logsCollectorDaemonSet)
	if err != nil {
		return container.ContainerStatus_Stopped, stacktrace.Propagate(err, "An error occurred getting pods managed by logs collector daemon set '%v'.", logsCollectorDaemonSet.Name)
	}
	if len(logsCollectorPods) < 1 {
		// if there are no pods associated with logs collector daemon set, assume something is wrong and err
		return container.ContainerStatus_Stopped, stacktrace.NewError("No pods managed by logs collector daemon set were found. There should be at least one. This is likely a bug in Kurtosis.")
	}

	runningPods := 0
	stoppedPods := 0
	for _, pod := range logsCollectorPods {
		podStatus, err := shared_helpers.GetContainerStatusFromPod(pod)
		if err != nil {
			return container.ContainerStatus_Stopped, stacktrace.Propagate(err, "An error occurred retrieving container status for a pod managed by logs collectors collector daemon set '%v' with name: %v\n", logsCollectorDaemonSet.Name, pod.Name)
		}

		if podStatus == container.ContainerStatus_Running {
			runningPods++
		} else {
			stoppedPods++
		}
	}

	if runningPods == 0 {
		return container.ContainerStatus_Stopped, nil
	}

	if stoppedPods > 0 {
		logrus.Warnf("Logs collector daemon set '%v' has %d stopped pods out of %d total pods. The collector is partially degraded.", logsCollectorDaemonSet.Name, stoppedPods, len(logsCollectorPods))
	}

	return container.ContainerStatus_Running, nil
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

	// this port spec represents the http port that each log collector container (on each pod managed by the daemon set) will have a port exposed on
	httpPortSpec, err := port_spec.NewPortSpec(logsCollectorHttpPortNumber, port_spec.TransportProtocol_TCP, httpProtocolStr, noWait, emptyUrl)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred creating the log collectors public HTTP port spec object using number '%v' and protocol '%v'",
			logsCollectorHttpPortNumber,
			port_spec.TransportProtocol_TCP,
		)
	}
	// Check that at least one DaemonSet pod is available rather than requiring all pods.
	// On large clusters, some nodes may be slow to start pods (e.g. disk pressure) which
	// would cause the engine startup to fail unnecessarily.
	var lastErr error
	for _, pod := range pods {
		if len(pod.Spec.Containers) < 1 {
			lastErr = stacktrace.NewError("Pod '%v' managed by logs collector daemon set '%v' doesn't have any containers associated with it. There should be at least one container.", pod.Name, logsCollectorDaemonSet.Name)
			continue
		}

		fluentBitContainerName := pod.Spec.Containers[0].Name
		if err = shared_helpers.WaitForPortAvailabilityUsingNetstat(
			kubernetesManager,
			k8sResources.namespace.Name,
			pod.Name,
			fluentBitContainerName,
			httpPortSpec,
			maxAvailabilityChecksRetries,
			timeToWaitBetweenAvailabilityChecksDuration); err != nil {
			lastErr = stacktrace.Propagate(err, "An error occurred while checking for availability of pod '%v' managed by logs collector daemon set '%v'", pod.Name, logsCollectorDaemonSet.Name)
			continue
		}
		return nil
	}

	return stacktrace.Propagate(lastErr, "No logs collector pods managed by daemon set '%v' became available", logsCollectorDaemonSet.Name)
}

func waitForNamespaceRemoval(
	ctx context.Context,
	namespace string,
	kubernetesManager *kubernetes_manager.KubernetesManager) error {

	for i := uint(0); i < maxTriesToWaitForNamespaceRemoval; i++ {
		if _, err := kubernetesManager.GetNamespace(ctx, namespace); err != nil {
			// if err was returned, namespace doesn't exist, or it's been marked for deleted
			return nil
		}

		// Tiny optimization to not sleep if we're not going to run the loop again
		if i < maxTriesToWaitForNamespaceRemoval-1 {
			time.Sleep(timeToWaitBetweenNamespaceRemovalChecks)
		}
	}

	return stacktrace.NewError("Attempted to wait for namespace '%v' removal or to be marked for deletion '%v' times but namespace was not removed.", namespace, maxTriesToWaitForNamespaceRemoval)
}
