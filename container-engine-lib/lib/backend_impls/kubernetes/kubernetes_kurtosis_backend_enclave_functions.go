package kubernetes

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

func (backend *KubernetesKurtosisBackend) CreateEnclave(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	isPartitioningEnabled bool,
) (
	*enclave.Enclave,
	error,
) {
	if isPartitioningEnabled {
		return nil, stacktrace.NewError("Partitioning not supported for kubernetes-backed modules.")
	}
	namespaceName := fmt.Sprintf("kurtosis-%v", enclaveId)
	_, err := backend.kubernetesManager.GetNamespace(ctx, namespaceName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting namespaces using name '%+v', which is necessary to ensure that our enclave doesn't exist yet", namespaceName)
	}
	namespaceLabels := map[string]string{}
	namespace, err := backend.kubernetesManager.CreateNamespace(ctx, namespaceName, namespaceLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create namespace.")
	}
	logrus.Infof("Namespace: %+v", namespace)
	return nil, nil
}
