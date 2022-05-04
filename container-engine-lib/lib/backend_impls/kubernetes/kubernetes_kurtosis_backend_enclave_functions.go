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
		return nil, stacktrace.NewError("Partitioning not supported for Kubernetes-backed Kurtosis.")
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
	// TODO TODO TODO REPLACE THIS WITH NAME LABELS AND ANNOTATIONS FROM ATTRIBUTES PROVIDER
	namespaceLabels := map[string]string{}
	_, err = backend.kubernetesManager.CreateNamespace(ctx, namespaceName, namespaceLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create namespace with name '%v'", namespaceName)
	}
	shouldDeleteNamespace := true
	defer func() {
		if shouldDeleteNamespace {
			if err := backend.kubernetesManager.RemoveNamespace(teardownContext, namespaceName); err != nil {
				logrus.Errorf("Creating the enclave didn't complete successfully, so we tried to delete namespace '%v' that we created but an error was thrown:\n%v", namespaceName, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove namespace with name '%v'!!!!!!!", namespaceName)
			}
		}
	}()
	newEnclave := enclave.NewEnclave(enclaveId, enclave.EnclaveStatus_Empty, "", "", net.IP{}, nil)

	shouldDeleteNamespace = false
	return newEnclave, nil
}

func (backend *KubernetesKurtosisBackend) GetEnclaves(ctx context.Context, filters *enclave.EnclaveFilters) (map[enclave.EnclaveID]*enclave.Enclave, error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) StopEnclaves(ctx context.Context, filters *enclave.EnclaveFilters) (successfulEnclaveIds map[enclave.EnclaveID]bool, erroredEnclaveIds map[enclave.EnclaveID]error, resultErr error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) DumpEnclave(ctx context.Context, enclaveId enclave.EnclaveID, outputDirpath string) error {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) DestroyEnclaves(ctx context.Context, filters *enclave.EnclaveFilters) (successfulEnclaveIds map[enclave.EnclaveID]bool, erroredEnclaveIds map[enclave.EnclaveID]error, resultErr error) {
	//TODO implement me
	panic("implement me")
}
