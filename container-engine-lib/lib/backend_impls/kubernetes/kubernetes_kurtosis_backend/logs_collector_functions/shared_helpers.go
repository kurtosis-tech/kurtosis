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
	"github.com/sirupsen/logrus"
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
	logrus.Info("size of daemon sets: %v", len(logCollectorDaemonSets))

	// There should every only be one log collector daemonset deployed in a cluster so error if more are found
	if len(logCollectorDaemonSets) > 1 {
		return nil, stacktrace.NewError("Multiple log collector daemonsets were detected inside the cluster, but there should only be one. This is likely a bug inside Kurtosis.")
	}

	if len(logCollectorDaemonSets) == 0 {
		return nil, nil
	}

	return &logCollectorDaemonSets[0], nil
}

func getLogsCollectorDaemonSetObject(ctx context.Context, logsCollectorDaemonSetResource *v1.DaemonSet) (*logs_collector.LogsCollector, error) {
	// translate log collector daemonset info to logs collector info
	dummyPortSpecOne, err := port_spec.NewPortSpec(1234, port_spec.TransportProtocol(0), "HTTP", nil, "")
	if err != nil {
		return nil, err
	}

	dummyPortSpecTwo, err := port_spec.NewPortSpec(1234, port_spec.TransportProtocol(0), "HTTP", nil, "")
	if err != nil {
		return nil, err
	}

	return logs_collector.NewLogsCollector(
		container.ContainerStatus_Running,
		net.IP{},
		net.IP{},
		dummyPortSpecOne,
		dummyPortSpecTwo,
	), nil
}
