package logs_aggregator_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

func CreateLogsAggregator(
	ctx context.Context,
	logsAggregatorDeployment LogsAggregatorDeployment,
	objAttrProvider object_attributes_provider.KubernetesObjectAttributesProvider,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (*logs_aggregator.LogsAggregator, func(), error) {
	var logsAggregatorObj *logs_aggregator.LogsAggregator
	var kubernetesResources *logsAggregatorKubernetesResources
	shouldRemoveLogsAggregator := false // only gets set to true if a logs aggregator is created (and might need to be removed)
	var removeLogsAggregatorFunc func()
	var err error

	logsAggregatorObj, kubernetesResources, err = getLogsAggregatorObjAndResourcesForCluster(ctx, kubernetesManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting logs aggregator object and resources for cluster.")
	}

	if logsAggregatorObj != nil {
		logrus.Debug("Found existing logs aggregator deployment.")
	} else {
		logrus.Debug("Did not find existing logs aggregator, creating one...")
		service, deployment, namespace, configMap, removeLogsAggregatorFunc, err := logsAggregatorDeployment.CreateAndStart(
			ctx,
			defaultLogsListeningPortNum,
			objAttrProvider,
			kubernetesManager)
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred creating logs aggregator deployment.")
		}
		shouldRemoveLogsAggregator = false
		defer func() {
			if shouldRemoveLogsAggregator {
				removeLogsAggregatorFunc()
			}
		}()

		kubernetesResources = &logsAggregatorKubernetesResources{
			service:    service,
			deployment: deployment,
			namespace:  namespace,
			configMap:  configMap,
		}

		logsAggregatorObj, err = getLogsAggregatorObjectFromKubernetesResources(ctx, kubernetesManager, kubernetesResources)
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred getting the logs aggregator object from kubernetes resources.")
		}
	}

	// wait for availability

	return logsAggregatorObj, removeLogsAggregatorFunc, nil
}
