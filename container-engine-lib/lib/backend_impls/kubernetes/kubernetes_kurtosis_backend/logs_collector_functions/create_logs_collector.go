package logs_collector_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

func CreateLogsCollector(
	ctx context.Context,
	logsCollectorTcpPortNumber uint16,
	logsCollectorHttpPortNumber uint16,
	logsCollectorDaemonSet LogsCollectorDaemonSet,
	logsAggregator *logs_aggregator.LogsAggregator,
	kubernetesManager *kubernetes_manager.KubernetesManager,
	objAttrsProvider object_attributes_provider.KubernetesObjectAttributesProvider,
) (
	*logs_collector.LogsCollector,
	func() error,
	error,
) {
	var logsCollectorObj *logs_collector.LogsCollector
	logsCollectorObj, err := getLogsCollectorObjForCluster(ctx, kubernetesManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting logs aggregator container.")
	}
	if logsCollectorObj != nil {
		logrus.Debug("Found existing logs collector daemon set.")
		return logsCollectorObj, nil, nil
	}

	// TODO: get logs collector tcp and http port id

	logrus.Debug("Did not find existing log collector, creating one...")
	daemonSet, configMap, removeLogsCollectorDaemonSetFunc, err := logsCollectorDaemonSet.CreateAndStart(
		ctx,
		"", // TODO: fill these in when adding aggregator to k8s
		0,  // TODO: fill these in when adding aggregator to k8s
		logsCollectorTcpPortNumber,
		logsCollectorHttpPortNumber,
		"",
		"",
		objAttrsProvider,
		kubernetesManager,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(
			err,
			"An error occurred starting the logs collector daemon set with logs aggregator host '%v', logs aggregator port '%v', HTTP port number '%v', TCP port id '%v', and HTTP port id '%v'",
			"",
			"",
			"",
			logsCollectorHttpPortNumber,
			logsCollectorTcpPortNumber,
		)
	}
	shouldRemoveLogsCollectorDaemonSet := true
	defer func() {
		if shouldRemoveLogsCollectorDaemonSet {
			removeLogsCollectorDaemonSetFunc()
		}
	}()

	kubernetesResources := &logsCollectorKubernetesResources{
		daemonSet: daemonSet,
		configMap: configMap,
	}

	logsCollectorObj, err = getLogsCollectorsObjectFromKubernetesResources(ctx, kubernetesManager, kubernetesResources)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting the logs collector object from kubernetes resources.")
	}

	// TODO: Availability check
	// figure out how to check availability for all pods in a daemon set
	// WaitForPortAvailabilityUsingNetstat <- could do this for all the pods / containers in the logs collector daemon set
	// could use the same availability checker that docker uses - if we can just turn ip add

	// need port info to do availability check
	// so need ip addresses or hostnames of all the pods started by the daemon set

	return logsCollectorObj, nil, nil
}
