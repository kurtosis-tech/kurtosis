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
	engineNamespace string,
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
		removeLogsAggregatorFunc = func() {} // can't create remove in this situation so jus make it a no op
		logrus.Debug("Found existing logs aggregator deployment.")
	} else {
		logrus.Debug("Did not find existing logs aggregator, creating one...")
		service, deployment, namespace, configMap, removeLogsAggregatorFunc, err := logsAggregatorDeployment.CreateAndStart(
			ctx,
			defaultLogsListeningPortNum,
			engineNamespace,
			objAttrProvider,
			kubernetesManager)
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred creating logs aggregator deployment.")
		}
		shouldRemoveLogsAggregator = true
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

	logrus.Debugf("Checking for logs aggregator availability in namespace '%v'...", kubernetesResources.namespace.Name)

	healthCheckEndpoint, healthCheckPortNum := logsAggregatorDeployment.GetHTTPHealthCheckEndpointAndPort()
	if err = waitForLogsAggregatorAvailability(ctx, healthCheckEndpoint, healthCheckPortNum, kubernetesResources, kubernetesManager); err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred while waiting for the logs aggregator deployment to become available")
	}
	logrus.Debugf("...logs aggregator is available in namepsace '%v'", kubernetesResources.namespace.Name)

	shouldRemoveLogsAggregator = false
	return logsAggregatorObj, removeLogsAggregatorFunc, nil
}
