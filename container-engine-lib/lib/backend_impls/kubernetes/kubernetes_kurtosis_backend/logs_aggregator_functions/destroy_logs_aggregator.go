package logs_aggregator_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
)

// DestroyLogsAggregator destroys the logs aggregator and its associated resources
func DestroyLogsAggregator(ctx context.Context, kubernetesManager *kubernetes_manager.KubernetesManager) error {
	logsAggregatorResources, err := getLogsAggregatorKubernetesResourcesForCluster(ctx, kubernetesManager)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred retrieving kubernetes resources for logs aggregator.")
	}

	var logsAggregatorNamespace *apiv1.Namespace
	if logsAggregatorResources.namespace != nil {
		logsAggregatorNamespace = logsAggregatorResources.namespace
	} else {
		// assume if there is no namespace, there's no logs aggregator but log a debug in case
		logrus.Debug("No logs aggregator namespace found. Returning without attempting to destroy remaining logs aggregator resources.")
		return nil
	}

	var destroyErr error
	if logsAggregatorResources.deployment != nil {
		if err := kubernetesManager.RemoveDeployment(ctx, logsAggregatorNamespace.Name, logsAggregatorResources.deployment); err != nil {
			destroyErr = stacktrace.Propagate(err, "An error occurred removing logs aggregator deployment.")
		}
	}

	if logsAggregatorResources.configMap != nil {
		if err := kubernetesManager.RemoveConfigMap(ctx, logsAggregatorNamespace.Name, logsAggregatorResources.configMap); err != nil {
			destroyErr = stacktrace.Propagate(err, "An error occurred removing logs aggregator config map.")
		}
	}

	if logsAggregatorResources.service != nil {
		if err := kubernetesManager.RemoveService(ctx, logsAggregatorResources.service); err != nil {
			destroyErr = stacktrace.Propagate(err, "An error occurred removing logs aggregator service.")
		}
	}

	if err := kubernetesManager.RemoveNamespace(ctx, logsAggregatorNamespace); err != nil {
		destroyErr = stacktrace.Propagate(err, "An error occurred removing logs aggregator namespace.")
	}

	return destroyErr
}
