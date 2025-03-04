package logs_collector_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
)

// DestroyLogsCollector destroys the logs collector and its associated resources
func DestroyLogsCollector(ctx context.Context, kubernetesManager *kubernetes_manager.KubernetesManager) error {
	logsCollectorResources, err := getLogsCollectorKubernetesResourcesForCluster(ctx, kubernetesManager)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred retrieving kubernetes resources for logs collector.")
	}

	var logsCollectorNamespace *apiv1.Namespace
	if logsCollectorResources.namespace != nil {
		logsCollectorNamespace = logsCollectorResources.namespace
	} else {
		// assume if there is no namespace, there's no logs collector but log a debug in case
		logrus.Debug("No logs collector namespace found. Returning without attempting to destroy remaining logs collector resources.")
		return nil
	}

	var destroyErr error
	if logsCollectorResources.daemonSet != nil {
		if err := kubernetesManager.RemoveDaemonSet(ctx, logsCollectorNamespace.Name, logsCollectorResources.daemonSet); err != nil {
			destroyErr = stacktrace.Propagate(err, "An error occurred removing logs collector daemon set.")
		}
	}

	if logsCollectorResources.configMap != nil {
		if err := kubernetesManager.RemoveConfigMap(ctx, logsCollectorNamespace.Name, logsCollectorResources.configMap); err != nil {
			destroyErr = stacktrace.Propagate(err, "An error occurred removing logs collector config map.")
		}
	}

	if logsCollectorResources.serviceAccount != nil {
		if err := kubernetesManager.RemoveServiceAccount(ctx, logsCollectorResources.serviceAccount); err != nil {
			destroyErr = stacktrace.Propagate(err, "An error occurred removing logs collector service account.")
		}
	}

	if logsCollectorResources.clusterRole != nil {
		if err := kubernetesManager.RemoveClusterRole(ctx, logsCollectorResources.clusterRole); err != nil {
			destroyErr = stacktrace.Propagate(err, "An error occurred removing logs collector cluster role.")
		}
	}

	if logsCollectorResources.clusterRoleBinding != nil {
		if err := kubernetesManager.RemoveClusterRoleBindings(ctx, logsCollectorResources.clusterRoleBinding); err != nil {
			destroyErr = stacktrace.Propagate(err, "An error occurred removing logs collector cluster role binding.")
		}
	}

	if err := kubernetesManager.RemoveNamespace(ctx, logsCollectorNamespace); err != nil {
		destroyErr = stacktrace.Propagate(err, "An error occurred removing logs collector namespace.")
	}

	if err := waitForNamespaceRemoval(ctx, logsCollectorNamespace.Name, kubernetesManager); err != nil {
		destroyErr = stacktrace.Propagate(err, "An error occurred waiting for logs collector namespace to be removed.")
	}

	return destroyErr
}
