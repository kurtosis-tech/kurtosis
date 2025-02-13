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
	// TODO: if the logs collector lifecycle is managed by the engine, might want to return a removal function so the engine can get rid of if defer undoing
	error,
) {
	//get logs collector for enclave
	logsCollectorDaemonSetResource, err := getLogsCollectorDaemonSetForCluster(ctx, "kube-system", kubernetesManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting logs aggregator container.")
	}
	if logsCollectorDaemonSetResource != nil {
		logrus.Info("Found existing logs collector daemon set.")
		logsCollectorDaemonSetObj, err := getLogsCollectorDaemonSetObject(ctx, logsCollectorDaemonSetResource)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting logs collector daemon set object.")
		}
		// removeFunc := func() {
		//		removeLogsCollectorContainerFunc()
		//}()
		return logsCollectorDaemonSetObj, nil
	}

	// get logs aggregator host an port number

	logrus.Info("Did not find existing log collector, creating one...")
	// create and start the logs collector container
	// TODO: adjust return values to whats needed for daemonset
	containerId, _, _, _, err := logsCollectorDaemonSet.CreateAndStart(
		ctx,
		"",
		0,
		logsCollectorTcpPortNumber,
		logsCollectorHttpPortNumber,
		"",
		"",
		objAttrsProvider,
		kubernetesManager,
	)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred running the logs collector container with container ID '%v' with logs aggregator host '%v', logs aggregator port '%v', HTTP port number '%v', TCP port id '%v', and HTTP port id '%v' in Docker network with ID '%v'",
			containerId,
			"",
			"",
			logsCollectorHttpPortNumber,
			"",
			"",
			"",
		)
	}
	//shouldRemoveLogsCollectorContainerFunc := true
	//defer func() {
	//	if shouldRemoveLogsCollectorContainerFunc {
	//		removeLogsCollectorContainerFunc()
	//	}
	//}()

	// get logs collector object

	// check for availability
	// TODO: figure out how to check availability for all pods in a daemonset
	// WaitForPortAvailabilityUsingNetstat <- could do this for all the pods / containers in the logs collector daemon set
	// could use the same availability checker that docker uses - if we can just turn ip add

	//shouldRemoveLogsCollectorContainerFunc := true
	return nil, nil
}
