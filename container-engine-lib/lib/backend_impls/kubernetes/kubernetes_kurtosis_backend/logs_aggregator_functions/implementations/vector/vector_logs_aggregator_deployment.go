package vector

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
)

type VectorLogsAggregatorDeployment struct{}

func NewVectorLogsAggregatorDeployment() *VectorLogsAggregatorDeployment {
	return &VectorLogsAggregatorDeployment{}
}

func (logsAggregator *VectorLogsAggregatorDeployment) CreateAndStart(
	ctx context.Context,
	logsListeningPort uint16,
	objAttrProvider object_attributes_provider.KubernetesObjectAttributesProvider,
	kubernetesManager *kubernetes_manager.KubernetesManager) (
	*apiv1.Service,
	*appsv1.Deployment,
	*apiv1.Namespace,
	func(),
	error) {
	// create namespace

	// create service

	// create deployment

	// create removal functions

	return nil, nil, nil, nil, nil
}
