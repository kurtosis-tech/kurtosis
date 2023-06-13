package logs_collector_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/logs_collector_functions/implementations/fluentbit"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_database"
	"github.com/kurtosis-tech/stacktrace"
)

func CreateLogsCollectorForCluster(
	ctx context.Context,
	namespace string,
	serviceAccountName string,
	logsDatabase *logs_database.LogsDatabase,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) error {
	logsCollector := fluentbit.NewFluentbitLogsCollectorDaemonSet()

	if logsCollector.AlreadyExists(ctx, namespace, kubernetesManager) {
		return stacktrace.NewError("Found existing logs collector DaemonSet; cannot start a new one")
	}

	if logsDatabase.GetMaybePrivateIpAddr() == nil {
		return stacktrace.NewError("Expected the logs database has private IP address but this is nil")
	}

	logsDatabaseHost := logsDatabase.GetMaybePrivateIpAddr().String()
	logsDatabasePort := logsDatabase.GetPrivateHttpPort().GetNumber()

	err := logsCollector.CreateAndStart(
		ctx,
		namespace,
		logsDatabaseHost,
		logsDatabasePort,
		0, // unused for now
		0, // unused for now
		serviceAccountName,
		kubernetesManager,
	)
	if err != nil {
		return stacktrace.Propagate(err, "Unable to create the Kubernetes DaemonSet for Fluentbit logs collector")
	}
	// TODO: maybe check for availability as we do in Docker
	return nil
}
