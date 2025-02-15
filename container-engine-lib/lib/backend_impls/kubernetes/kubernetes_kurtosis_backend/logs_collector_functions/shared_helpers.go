package logs_collector_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_resource_collectors"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	v1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"net"
)

func getLogsCollectorObjForCluster(ctx context.Context, kubernetesManager *kubernetes_manager.KubernetesManager) (*logs_collector.LogsCollector, error) {
	kubernetesResources, err := getLogsCollectorKubernetesResourcesForCluster(ctx, "kube-system", kubernetesManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes resources for logs collectors.")
	}

	obj, err := getLogsCollectorsObjectFromKubernetesResources(ctx, kubernetesManager, kubernetesResources)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting logs collector object from kubernetes resources.")
	}
	return obj, nil
}

func getLogsCollectorKubernetesResourcesForCluster(ctx context.Context, namespace string, kubernetesManager *kubernetes_manager.KubernetesManager) (*logsCollectorKubernetesResources, error) {
	logsCollectorDaemonSetSearchLabels := map[string]string{
		kubernetes_label_key.AppIDKubernetesLabelKey.GetString():                label_value_consts.AppIDKubernetesLabelValue.GetString(),
		kubernetes_label_key.KurtosisResourceTypeKubernetesLabelKey.GetString(): label_value_consts.LogsCollectorKurtosisResourceTypeKubernetesLabelValue.GetString(),
	}

	logsCollectorCfgConfigMaps, err := kubernetes_resource_collectors.CollectMatchingConfigMaps(
		ctx,
		kubernetesManager,
		namespace,
		logsCollectorDaemonSetSearchLabels, // could retrieve the logs collector by the logs collector guid, but for now just retrieve by resource type
		kubernetes_label_key.KurtosisResourceTypeKubernetesLabelKey.GetString(),
		map[string]bool{
			label_value_consts.LogsCollectorKurtosisResourceTypeKubernetesLabelValue.GetString(): true,
		})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting config map for logs collector in namespace '%v'", namespace)
	}
	var configMap *apiv1.ConfigMap
	if configMapForId, found := logsCollectorCfgConfigMaps[label_value_consts.LogsCollectorKurtosisResourceTypeKubernetesLabelValue.GetString()]; found {
		if len(configMapForId) > 1 {
			return nil, stacktrace.NewError(
				"Expected at most one logs collector config maps in namespace '%v' for logs collector but found '%v'",
				namespace,
				len(logsCollectorCfgConfigMaps),
			)
		}
		configMap = configMapForId[0]
	}

	daemonSets, err := kubernetes_resource_collectors.CollectMatchingDaemonSets(
		ctx,
		kubernetesManager,
		namespace,
		logsCollectorDaemonSetSearchLabels,
		kubernetes_label_key.KurtosisResourceTypeKubernetesLabelKey.GetString(),
		map[string]bool{
			label_value_consts.LogsCollectorKurtosisResourceTypeKubernetesLabelValue.GetString(): true,
		})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting config map for logs collector in namespace '%v'", namespace)
	}
	var daemonSet *v1.DaemonSet
	if logsCollectorDaemonSet, found := daemonSets[label_value_consts.LogsCollectorKurtosisResourceTypeKubernetesLabelValue.GetString()]; found {
		if len(logsCollectorDaemonSet) > 1 {
			return nil, stacktrace.NewError(
				"Expected at most one logs collector daemon set in namespace '%v' for logs collector but found '%v'",
				namespace,
				len(logsCollectorCfgConfigMaps),
			)
		}
		if len(logsCollectorDaemonSet) == 0 {
			daemonSet = nil
		} else {
			daemonSet = logsCollectorDaemonSet[0]
		}
	}

	logsCollectorKubernetesResources := &logsCollectorKubernetesResources{
		daemonSet: daemonSet,
		configMap: configMap,
	}

	return logsCollectorKubernetesResources, nil
}

func getLogsCollectorsObjectFromKubernetesResources(ctx context.Context, kubernetesManager *kubernetes_manager.KubernetesManager, logsCollectorKubernetesResources *logsCollectorKubernetesResources) (*logs_collector.LogsCollector, error) {
	if logsCollectorKubernetesResources.daemonSet == nil || logsCollectorKubernetesResources.configMap == nil {
		// if any resources not found for logs collector, don't return an object
		return nil, nil
	}

	var (
		logsCollectorStatus container.ContainerStatus
		privateIpAddr       net.IP
		bridgeNetworkIpAddr net.IP
		err                 error
	)

	// TODO: get the status of all the fluentbit pods/containers - if they are all running, set to running, else set to stopped
	logsCollectorStatus = container.ContainerStatus_Running

	// daemon sets don't have a entrypoint so leave these blank for now
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
