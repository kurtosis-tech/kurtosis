package kubernetes

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
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

	searchNamespaceLabels := map[string]string{
		label_key_consts.AppIDLabelKey.GetString(): label_value_consts.AppIDLabelValue.GetString(),
		label_key_consts.EnclaveIDLabelKey.GetString(): string(enclaveId),
	}
	namespaceName := fmt.Sprintf("kurtosis-%v", enclaveId)
	namespaceList, err := backend.kubernetesManager.GetNamespacesByLabels(ctx, searchNamespaceLabels)
	if err != nil {
		return nil, stacktrace.NewError("Failed to list namespaces from Kubernetes, so can not verify if enclave '%v' already exists.", enclaveId)
	}
	if len(namespaceList.Items) > 0 {
		return nil, stacktrace.NewError("Cannot create enclave with ID '%v' because an enclave with ID '%v' already exists", enclaveId, enclaveId)
	}

	// Make Enclave attributes provider
	enclaveObjAttrsProvider, err := backend.objAttrsProvider.ForEnclave()

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
