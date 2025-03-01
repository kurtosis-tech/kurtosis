package logs_aggregator_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
)

func CreateLogsAggregator(
	ctx context.Context,
	logsAggregatorDeployment LogsAggregatorDeployment,
	objAttrProvider object_attributes_provider.KubernetesObjectAttributesProvider,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (*logs_aggregator.LogsAggregator, func(), error) {
	// check if it exists

	// if it doesn't create

	// if it does use that object and return

	// wait for availability

	return nil, nil, nil
}
