package logs_collector_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	v1 "k8s.io/api/apps/v1"
	"net"
)

func getLogsCollectorDaemonSetForCluster(ctx context.Context, namespace string, kubernetesManager *kubernetes_manager.KubernetesManager) (*v1.DaemonSet, error) {
	logsCollectorDaemonSetSearchLabels := map[string]string{
		kubernetes_label_key.AppIDKubernetesLabelKey.GetString():                label_value_consts.AppIDKubernetesLabelValue.GetString(),
		kubernetes_label_key.KurtosisResourceTypeKubernetesLabelKey.GetString(): label_value_consts.LogsCollectorKurtosisResourceTypeKubernetesLabelValue.GetString(),
	}

	matchingLogsCollectorDaemonsets, err := kubernetesManager.GetDaemonSetsByLabels(ctx, namespace, logsCollectorDaemonSetSearchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred fetching logs collector containers using labels: %+v", logsCollectorDaemonSetSearchLabels)
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

	return &logCollectorDaemonSets[0], nil
}

func getLogsCollectorDaemonSetObject(ctx context.Context, logsCollectorDaemonSetResource *v1.DaemonSet) (*logs_collector.LogsCollector, error) {

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
