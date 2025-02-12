package logs_collector_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
	"github.com/kurtosis-tech/stacktrace"
)

func CreateLogsCollectorForEnclave(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	logsCollectorTcpPortNumber uint16,
	logsCollectorHttpPortNumber uint16,
	logsCollectorContainer LogsCollectorContainer,
	logsAggregator *logs_aggregator.LogsAggregator,
	kubernetesManager *kubernetes_manager.KubernetesManager,
	objAttrsProvider object_attributes_provider.KubernetesObjectAttributesProvider,
) (
	*logs_collector.LogsCollector,
	error,
) {
	// get logs collector for enclave

	// get enclave network

	// get logs aggregator host an port number

	// create and start the logs collector container
	containerId, _, _, _, err := logsCollectorContainer.CreateAndStart(
		ctx,
		enclaveUuid,
		"",
		0,
		logsCollectorTcpPortNumber,
		logsCollectorHttpPortNumber,
		"",
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

	// connect the container to the enclave network

	// get logs collector object

	// check for availability
	return nil, nil
}
