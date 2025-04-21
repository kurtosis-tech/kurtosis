package logs_collector_functions

import (
	"context"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	emptyUrl                = ""
	httpProtocolStr         = "http"
	logsCollectorTcpPortId  = "tcp"
	logsCollectorHttpPortId = "http"
)

var noWait *port_spec.Wait = nil

func CreateLogsCollector(
	ctx context.Context,
	logsCollectorTcpPortNumber uint16,
	logsCollectorHttpPortNumber uint16,
	logsCollectorDaemonSet LogsCollectorDaemonSet,
	logsAggregator *logs_aggregator.LogsAggregator,
	logsCollectorFilters []logs_collector.Filter,
	kubernetesManager *kubernetes_manager.KubernetesManager,
	objAttrsProvider object_attributes_provider.KubernetesObjectAttributesProvider,
) (
	*logs_collector.LogsCollector,
	func(),
	error,
) {
	var logsCollectorObj *logs_collector.LogsCollector
	var kubernetesResources *logsCollectorKubernetesResources
	shouldRemoveLogsCollector := false // only gets set to true if a logs collector is created (and might need to be removed)
	var removeLogsCollectorFunc func()
	var err error

	logsCollectorObj, kubernetesResources, err = getLogsCollectorObjAndResourcesForCluster(ctx, kubernetesManager)
	if err != nil {
		return nil, removeLogsCollectorFunc, stacktrace.Propagate(err, "An error occurred getting logs collector object and resources for cluster.")
	}

	if logsCollectorObj != nil {
		removeLogsCollectorFunc = func() {} // can't create remove in this situation so jus make it a no op
		logrus.Debug("Found existing logs collector daemon set.")
	} else {
		logrus.Debug("Did not find existing log collector, creating one...")
		daemonSet, configMap, namespace, serviceAccount, clusterRole, clusterRoleBinding, removeLogsCollectorFunc, err := logsCollectorDaemonSet.CreateAndStart(
			ctx,
			logsAggregator.GetMaybePrivateIpAddr().String(),
			logsAggregator.GetListeningPortNum(),
			logsCollectorTcpPortNumber,
			logsCollectorHttpPortNumber,
			logsCollectorTcpPortId,
			logsCollectorHttpPortId,
			logsCollectorFilters,
			objAttrsProvider,
			kubernetesManager,
		)
		if err != nil {
			return nil, removeLogsCollectorFunc, stacktrace.Propagate(
				err,
				"An error occurred starting the logs collector daemon set with logs collector with '%v', HTTP port number '%v', TCP port id '%v', and HTTP port id '%v'",
				logsCollectorTcpPortNumber,
				logsCollectorHttpPortNumber,
				logsCollectorTcpPortId,
				logsCollectorHttpPortId,
			)
		}
		shouldRemoveLogsCollector = true
		defer func() {
			if shouldRemoveLogsCollector {
				removeLogsCollectorFunc()
			}
		}()

		kubernetesResources = &logsCollectorKubernetesResources{
			daemonSet:          daemonSet,
			configMap:          configMap,
			serviceAccount:     serviceAccount,
			clusterRoleBinding: clusterRoleBinding,
			clusterRole:        clusterRole,
			namespace:          namespace,
		}

		logsCollectorObj, err = getLogsCollectorsObjectFromKubernetesResources(ctx, kubernetesManager, kubernetesResources)
		if err != nil {
			return nil, removeLogsCollectorFunc, stacktrace.Propagate(err, "An error occurred getting the logs collector object from kubernetes resources.")
		}
	}

	logrus.Debugf("Checking for logs collector availability in namespace '%v'...", kubernetesResources.namespace.Name)

	if err = waitForLogsCollectorAvailability(ctx, logsCollectorHttpPortNumber, kubernetesResources, kubernetesManager); err != nil {
		return nil, removeLogsCollectorFunc, stacktrace.Propagate(err, "An error occurred while waiting for the logs collector daemon set to become available")
	}
	logrus.Debugf("...logs collector is available in namepsace '%v'", kubernetesResources.namespace.Name)

	logrus.Debugf("Logs collector successfully created with name '%v' in namepsace'%v'", kubernetesResources.daemonSet.Name, kubernetesResources.namespace.Name)
	shouldRemoveLogsCollector = false
	return logsCollectorObj, removeLogsCollectorFunc, nil
}
