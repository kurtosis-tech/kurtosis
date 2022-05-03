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
	namespaceName := fmt.Sprintf("kurtosis-'%v'", enclaveId)
	namespaceLabels := map[string]string{}
	namespace, err := backend.kubernetesManager.CreateNamespace(ctx, namespaceName, namespaceLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create namespace.")
	}
	logrus.Infof("Namespace: %+v", namespace)
	return nil, nil
}
