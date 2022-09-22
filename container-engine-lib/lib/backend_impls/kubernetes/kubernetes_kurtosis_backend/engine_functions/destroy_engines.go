package engine_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/stacktrace"
)

func DestroyEngines(
	ctx context.Context,
	filters *engine.EngineFilters,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (
	resultSuccessfulEngineGuids map[engine.EngineGUID]bool,
	resultErroredEngineGuids map[engine.EngineGUID]error,
	resultErr error,
) {
	_, matchingResources, err := getMatchingEngineObjectsAndKubernetesResources(ctx, filters, kubernetesManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting engine Kubernetes resources matching filters: %+v", filters)
	}

	successfulEngineGuids := map[engine.EngineGUID]bool{}
	erroredEngineGuids := map[engine.EngineGUID]error{}
	for engineGuid, resources := range matchingResources {
		// Remove ClusterRoleBinding
		if resources.clusterRoleBinding != nil {
			roleBindingName := resources.clusterRoleBinding.Name
			if err := kubernetesManager.RemoveClusterRoleBindings(ctx, resources.clusterRoleBinding); err != nil {
				erroredEngineGuids[engineGuid] = stacktrace.Propagate(
					err,
					"An error occurred removing cluster role binding '%v' for engine '%v'",
					roleBindingName,
					engineGuid,
				)
				continue
			}
		}

		// Remove ClusterRole
		if resources.clusterRole != nil {
			roleName := resources.clusterRole.Name
			if err := kubernetesManager.RemoveClusterRole(ctx, resources.clusterRole); err != nil {
				erroredEngineGuids[engineGuid] = stacktrace.Propagate(
					err,
					"An error occurred removing cluster role '%v' for engine '%v'",
					roleName,
					engineGuid,
				)
				continue
			}
		}

		// Remove the namespace
		if resources.namespace != nil {
			namespaceName := resources.namespace.Name
			if err := kubernetesManager.RemoveNamespace(ctx, resources.namespace); err != nil {
				erroredEngineGuids[engineGuid] = stacktrace.Propagate(
					err,
					"An error occurred removing namespace '%v' for engine '%v'",
					namespaceName,
					engineGuid,
				)
				continue
			}
		}

		successfulEngineGuids[engineGuid] = true
	}
	return successfulEngineGuids, erroredEngineGuids, nil
}
