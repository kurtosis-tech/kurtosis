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

func getLogsCollectorKubernetesResourcesForCluster(ctx context.Context, namespace string, kubernetesManager *kubernetes_manager.KubernetesManager) (*logsCollectorKubernetesResources, error) {
	logsCollectorDaemonSetSearchLabels := map[string]string{
		kubernetes_label_key.AppIDKubernetesLabelKey.GetString():                label_value_consts.AppIDKubernetesLabelValue.GetString(),
		kubernetes_label_key.KurtosisResourceTypeKubernetesLabelKey.GetString(): label_value_consts.LogsCollectorKurtosisResourceTypeKubernetesLabelValue.GetString(),
	}

	logsCollectorGuidStr := ""
	logsCollectorCfgConfigMaps, err := kubernetes_resource_collectors.CollectMatchingConfigMaps(ctx, kubernetesManager, namespace, logsCollectorDaemonSetSearchLabels, "kurtosis-logs-collector", map[string]bool{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting config map for logs collector in namespace '%v'", namespace)
	}
	var configMap *apiv1.ConfigMap
	if configMapForId, found := logsCollectorCfgConfigMaps[logsCollectorGuidStr]; found {
		if len(configMapForId) > 1 {
			return nil, stacktrace.NewError(
				"Expected at most one logs collector config maps in namespace '%v' for logs collector with GUID '%v' but found '%v'",
				namespace,
				logsCollectorGuidStr,
				len(logsCollectorCfgConfigMaps),
			)
		}
		configMap = configMapForId[0]
	}

	logsCollectorCfgConfigMaps, err := kubernetes_resource_collectors.CollectMatchingConfigMaps(ctx, kubernetesManager, namespace, logsCollectorDaemonSetSearchLabels, "kurtosis-logs-collector", map[string]bool{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting config map for logs collector in namespace '%v'", namespace)
	}
	var configMap *apiv1.ConfigMap
	if configMapForId, found := logsCollectorCfgConfigMaps[logsCollectorGuidStr]; found {
		if len(configMapForId) > 1 {
			return nil, stacktrace.NewError(
				"Expected at most one logs collector config maps in namespace '%v' for logs collector with GUID '%v' but found '%v'",
				namespace,
				logsCollectorGuidStr,
				len(logsCollectorCfgConfigMaps),
			)
		}
		configMap = configMapForId[0]
	}

	var logCollectorDaemonSets []v1.DaemonSet
	logCollectorDaemonSets = append(logCollectorDaemonSets, matchingLogsCollectorDaemonsets.Items...)

	// There should every only be one log collector daemonset deployed in a cluster so error if more are found
	if len(logCollectorDaemonSets) > 1 {
		return nil, stacktrace.NewError("Multiple log collector daemon sets were detected inside the cluster, but there should only be one. This is likely a bug inside Kurtosis.")
	}

	if len(logCollectorDaemonSets) == 0 {
		return nil, nil
	}

	logsCollectorKubernetesResources :=

	return &logCollectorDaemonSets[0], nil
}

func getLogsCollectorsObjectFromKubernetesResources(ctx context.Context, logsCollectorKubernetesResources *logsCollectorKubernetesResources) (*logs_collector.LogsCollector, error) {
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
