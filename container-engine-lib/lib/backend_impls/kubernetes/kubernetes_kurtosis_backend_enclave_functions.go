package kubernetes

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
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
	teardownContext := context.Background()
	namespaceName := fmt.Sprintf("kurtosis-%v", enclaveId)
	namespaceList, err := backend.kubernetesManager.ListNamespaces(ctx)
	if err != nil {
		return nil, stacktrace.NewError("Failed to list namespaces from Kubernetes, so can not verify if namespace already exists.")
	}
	// Iterate through namespace list to do name matching because GetNamespace doesn't return a clear error
	// that distinguishes between failed lookup mechanism and namespace not existing
	for _, namespace := range namespaceList.Items {
		if namespace.GetName() == namespaceName {
			return nil, stacktrace.NewError("Namespace with name %v already exists.", namespaceName)
		}
	}
	namespaceLabels := map[string]string{}
	namespace, err := backend.kubernetesManager.CreateNamespace(ctx, namespaceName, namespaceLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create namespace.")
	}
	logrus.Infof("Namespace: %+v", namespace)
	shouldDeleteNetwork := true
	defer func() {
		if shouldDeleteNetwork {
			if err := backend.kubernetesManager.RemoveNamespace(teardownContext, namespaceName); err != nil {
				logrus.Errorf("Creating the enclave didn't complete successfully, so we tried to delete namespace '%v' that we created but an error was thrown:\n%v", namespaceName, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove namespace with name '%v'!!!!!!!", namespaceName)
			}
		}
	}()
	newEnclave := enclave.NewEnclave(enclaveId, enclave.EnclaveStatus_Empty, "", "", net.IP{}, nil)

	shouldDeleteNetwork = false
	return newEnclave, nil
}
