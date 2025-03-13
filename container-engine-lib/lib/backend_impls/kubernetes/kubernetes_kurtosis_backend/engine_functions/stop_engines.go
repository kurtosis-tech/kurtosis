package engine_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/logs_aggregator_functions"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/logs_collector_functions"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	applyconfigurationsv1 "k8s.io/client-go/applyconfigurations/core/v1"
)

func StopEngines(
	ctx context.Context,
	filters *engine.EngineFilters,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (
	resultSuccessfulEngineGuids map[engine.EngineGUID]bool,
	resultErroredEngineGuids map[engine.EngineGUID]error,
	resultErr error,
) {
	_, matchingKubernetesResources, err := getMatchingEngineObjectsAndKubernetesResources(ctx, filters, kubernetesManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting engines and Kubernetes resources matching filters '%+v'", filters)
	}

	successfulEngineGuids := map[engine.EngineGUID]bool{}
	erroredEngineGuids := map[engine.EngineGUID]error{}
	for engineGuid, resources := range matchingKubernetesResources {
		if resources.namespace == nil {
			// No namespace means nothing needs stopping
			successfulEngineGuids[engineGuid] = true
			continue
		}
		namespaceName := resources.namespace.Name

		if resources.pod != nil {
			podName := resources.pod.Name
			if err := kubernetesManager.RemovePod(ctx, resources.pod); err != nil {
				erroredEngineGuids[engineGuid] = stacktrace.Propagate(
					err,
					"An error occurred removing pod '%v' in namespace '%v' for engine '%v'",
					podName,
					namespaceName,
					engineGuid,
				)
				continue
			}
		}

		if resources.engineNodeName != "" {
			engineNodeName := resources.engineNodeName
			engineNodeSelectors := resources.engineNodeSelectors
			if err := kubernetesManager.RemoveLabelsFromNode(ctx, engineNodeName, engineNodeSelectors); err != nil {
				erroredEngineGuids[engineGuid] = stacktrace.Propagate(
					err,
					"An error occurred removing labels '%v' from node '%v' for engine '%v'",
					engineNodeSelectors,
					engineNodeName,
					engineGuid,
				)
				continue
			}

		}

		kubernetesService := resources.service
		if kubernetesService != nil {
			serviceName := kubernetesService.Name
			updateConfigurator := func(updatesToApply *applyconfigurationsv1.ServiceApplyConfiguration) {
				specUpdates := applyconfigurationsv1.ServiceSpec().WithSelector(nil)
				updatesToApply.WithSpec(specUpdates)
			}
			if _, err := kubernetesManager.UpdateService(ctx, namespaceName, serviceName, updateConfigurator); err != nil {
				erroredEngineGuids[engineGuid] = stacktrace.Propagate(
					err,
					"An error occurred removing selectors from service '%v' in namespace '%v' for engine '%v'",
					kubernetesService.Name,
					namespaceName,
					engineGuid,
				)
				continue
			}
		}

		successfulEngineGuids[engineGuid] = true
	}

	// Stop centralized logging components
	if err := logs_aggregator_functions.DestroyLogsAggregator(ctx, kubernetesManager); err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred removing the logs aggregator.")
	}
	logrus.Debug("Successfully destroyed logs aggregator.")

	if err := logs_collector_functions.DestroyLogsCollector(ctx, kubernetesManager); err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred removing the logs collector.")
	}
	logrus.Debug("Successfully destroyed logs collector.")

	return successfulEngineGuids, erroredEngineGuids, nil
}
