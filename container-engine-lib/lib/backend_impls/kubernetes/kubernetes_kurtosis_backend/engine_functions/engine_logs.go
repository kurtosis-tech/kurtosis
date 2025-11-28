package engine_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_resource_collectors"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/stacktrace"
	apiv1 "k8s.io/api/core/v1"
)

// GetEngineLogs dump pods, into the received output dirpath, for the current engine
// it assumes that there is a single engine per cluster
func GetEngineLogs(
	ctx context.Context,
	outputDirpath string,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) error {
	engineMatchLabels := getEngineMatchLabels()

	engineGuidStrs := map[string]bool{}

	// Namespaces
	namespaces, err := kubernetes_resource_collectors.CollectMatchingNamespaces(
		ctx,
		kubernetesManager,
		engineMatchLabels,
		kubernetes_label_key.IDKubernetesLabelKey.GetString(),
		engineGuidStrs,
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting engine namespaces")
	}

	if len(namespaces) == 0 {
		return stacktrace.NewError("Didn't find the engine's namespace")
	}
	if len(namespaces) > 1 {
		return stacktrace.NewError("Found more than one engine's namespace; this is a bug in Kurtosis")
	}

	for engineGuidStr, namespacesForId := range namespaces {
		engineGuid := engine.EngineGUID(engineGuidStr)
		if len(namespacesForId) > 1 {
			return stacktrace.NewError(
				"Expected at most one namespace to match engine GUID '%v', but got '%v'",
				engineGuidStr,
				len(namespacesForId),
			)
		}

		engineNamespace := namespacesForId[0]
		if engineNamespace == nil {
			return stacktrace.NewError("Expected at most one namespace to match engine GUID '%v', but it wasn't found; this is a bug in Kurtosis", engineGuid)
		}
		namespaceName := engineNamespace.GetName()

		// get the engine's pod
		pods, err := kubernetes_resource_collectors.CollectMatchingPods(
			ctx,
			kubernetesManager,
			namespaceName,
			engineMatchLabels,
			kubernetes_label_key.IDKubernetesLabelKey.GetString(),
			map[string]bool{
				engineGuidStr: true,
			},
		)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting pods matching engine GUID '%v' in namespace '%v'", engineGuid, namespaceName)
		}

		podsForId, found := pods[engineGuidStr]
		if !found {
			return stacktrace.NewError("Didn't find any Kubernetes pod for engine with GUID '%v' in namespace '%v'", engineGuid, namespaceName)
		}

		if len(podsForId) == 0 {
			return stacktrace.NewError(
				"Expected to find one engine pod in namespace '%v' for engine with GUID '%v' "+
					"but none was found",
				namespaceName,
				engineGuid,
			)
		}
		if len(podsForId) > 1 {
			return stacktrace.NewError(
				"Expected at most one engine pod in namespace '%v' for engine with GUID '%v' "+
					"but found '%v'",
				namespaceName,
				engineGuid,
				len(pods),
			)
		}

		podsToDump := []apiv1.Pod{
			*podsForId[0],
		}

		if err = shared_helpers.DumpNamespacePods(ctx, kubernetesManager, engineNamespace, podsToDump, outputDirpath); err != nil {
			return stacktrace.Propagate(err, "An error occurred dumping pods '%+v' in namespace '%v'", podsToDump, namespaceName)
		}
	}

	return nil
}
