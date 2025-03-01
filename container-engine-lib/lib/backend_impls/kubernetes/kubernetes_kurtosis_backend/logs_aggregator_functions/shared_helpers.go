package logs_aggregator_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_resource_collectors"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/stacktrace"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
)

func getLogsAggregatorKubernetesResourcesForCluster(ctx context.Context, kubernetesManager *kubernetes_manager.KubernetesManager) (*logsAggregatorKubernetesResources, error) {
	resourceTypeLabelKeyStr := kubernetes_label_key.KurtosisResourceTypeKubernetesLabelKey.GetString()
	logsAggregatorResourceTypeLabelValStr := label_value_consts.LogsAggregatorKurtosisResourceTypeKubernetesLabelValue.GetString()
	logsAggregatorDeploymentSearchLabels := map[string]string{
		kubernetes_label_key.AppIDKubernetesLabelKey.GetString(): label_value_consts.AppIDKubernetesLabelValue.GetString(),
		resourceTypeLabelKeyStr:                                  logsAggregatorResourceTypeLabelValStr,
	}

	logsAggregatorNamespaces, err := kubernetes_resource_collectors.CollectMatchingNamespaces(ctx, kubernetesManager, logsAggregatorDeploymentSearchLabels, resourceTypeLabelKeyStr, map[string]bool{logsAggregatorResourceTypeLabelValStr: true})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting namespace for logs aggregator.")
	}
	var namespace *apiv1.Namespace
	if logsAggregatorNamespaceForLabel, found := logsAggregatorNamespaces[logsAggregatorResourceTypeLabelValStr]; found {
		if len(logsAggregatorNamespaceForLabel) > 1 {
			return nil, stacktrace.NewError(
				"Expected at most one namespaces for the logs aggregator but found '%v'",
				len(logsAggregatorNamespaceForLabel),
			)
		}
		if len(logsAggregatorNamespaceForLabel) == 0 {
			// if no namespace for logs aggregator, assume it doesn't exist at all
			return &logsAggregatorKubernetesResources{
				deployment: nil,
				service:    nil,
				namespace:  nil,
			}, nil
		}
		namespace = logsAggregatorNamespaceForLabel[0]
	} else {
		return &logsAggregatorKubernetesResources{
			deployment: nil,
			service:    nil,
			namespace:  nil,
		}, nil
	}

	logsAggregatorConfigServices, err := kubernetes_resource_collectors.CollectMatchingServices(
		ctx,
		kubernetesManager,
		namespace.Name,
		logsAggregatorDeploymentSearchLabels,
		resourceTypeLabelKeyStr,
		map[string]bool{
			logsAggregatorResourceTypeLabelValStr: true,
		})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting service for logs aggregator in namespace '%v'", namespace.Name)
	}
	var service *apiv1.Service
	if serviceForForLabel, found := logsAggregatorConfigServices[logsAggregatorResourceTypeLabelValStr]; found {
		if len(serviceForForLabel) > 1 {
			return nil, stacktrace.NewError(
				"Expected at most one logs aggregator services in namespace '%v' for logs aggregator but found '%v'",
				namespace.Name,
				len(serviceForForLabel),
			)
		}
		if len(serviceForForLabel) == 0 {
			service = nil
		} else {
			service = serviceForForLabel[0]
		}
	}

	deployments, err := kubernetes_resource_collectors.CollectMatchingDeployments(
		ctx,
		kubernetesManager,
		namespace.Name,
		logsAggregatorDeploymentSearchLabels,
		resourceTypeLabelKeyStr,
		map[string]bool{
			logsAggregatorResourceTypeLabelValStr: true,
		})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting deployments for logs aggregator in namespace '%v'", namespace)
	}
	var deployment *appsv1.Deployment
	if logsAggregatorDeploymentsForLabel, found := deployments[logsAggregatorResourceTypeLabelValStr]; found {
		if len(logsAggregatorDeploymentsForLabel) > 1 {
			return nil, stacktrace.NewError(
				"Expected at most one logs aggregator deployment in namespace '%v' for logs aggregator but found '%v'",
				namespace.Name,
				len(logsAggregatorDeploymentsForLabel),
			)
		}
		if len(logsAggregatorDeploymentsForLabel) == 0 {
			deployment = nil
		} else {
			deployment = logsAggregatorDeploymentsForLabel[0]
		}
	}

	logsAggregatorKubernetesResources := &logsAggregatorKubernetesResources{
		deployment: deployment,
		service:    service,
		namespace:  namespace,
	}

	return logsAggregatorKubernetesResources, nil
}
