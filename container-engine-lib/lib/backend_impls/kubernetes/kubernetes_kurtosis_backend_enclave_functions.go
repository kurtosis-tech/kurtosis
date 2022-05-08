package kubernetes

import (
	"context"
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
	namespaceList, err := backend.kubernetesManager.GetNamespacesByLabels(ctx, searchNamespaceLabels)
	if err != nil {
		return nil, stacktrace.NewError("Failed to list namespaces from Kubernetes, so can not verify if enclave '%v' already exists.", enclaveId)
	}
	if len(namespaceList.Items) > 0 {
		return nil, stacktrace.NewError("Cannot create enclave with ID '%v' because an enclave with ID '%v' already exists", enclaveId, enclaveId)
	}

	// Make Enclave attributes provider
	enclaveObjAttrsProvider, err := backend.objAttrsProvider.ForEnclave(string(enclaveId))
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while trying to generate an object attributes provider for the enclave with ID '%v'", enclaveId)
	}

	enclaveNamespaceAttrs, err := enclaveObjAttrsProvider.ForEnclaveNamespace(isPartitioningEnabled)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while trying to get the enclave network attributes for the enclave with ID '%v'", enclaveId)
	}

	enclaveNamespaceName := enclaveNamespaceAttrs.GetName().GetString()
	enclaveNamespaceLabels := enclaveNamespaceAttrs.GetLabels()

	enclaveNamespaceLabelMap := map[string]string{}
	for kubernetesLabelKey, kubernetesLabelValue := range enclaveNamespaceLabels {
		enclaveNamespaceLabelKey := kubernetesLabelKey.GetString()
		enclaveNamespaceValue := kubernetesLabelValue.GetString()
		enclaveNamespaceLabelMap[enclaveNamespaceLabelKey] = enclaveNamespaceValue
	}

	_, err = backend.kubernetesManager.CreateNamespace(ctx, enclaveNamespaceName, enclaveNamespaceLabelMap)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create namespace with name '%v' for enclave '%v'", enclaveNamespaceName, enclaveId)
	}
	shouldDeleteNamespace := true
	defer func() {
		if shouldDeleteNamespace {
			if err := backend.kubernetesManager.RemoveNamespace(teardownContext, enclaveNamespaceName); err != nil {
				logrus.Errorf("Creating the enclave didn't complete successfully, so we tried to delete namespace '%v' that we created but an error was thrown:\n%v", enclaveNamespaceName, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove namespace with name '%v'!!!!!!!", enclaveNamespaceName)
			}
		}
	}()

	// Create Persistent Volume Claim for the enclave (associated with namespace)
	volumeSearchLabels :=  map[string]string{
		label_key_consts.AppIDLabelKey.GetString(): label_value_consts.AppIDLabelValue.GetString(),
		label_key_consts.EnclaveIDLabelKey.GetString(): string(enclaveId),
		label_key_consts.VolumeTypeLabelKey.GetString(): label_value_consts.EnclaveDataVolumeTypeLabelValue.GetString(),
	}
	foundVolumes, err := backend.kubernetesManager.GetPersistentVolumeClaimsByLabels(ctx, enclaveNamespaceName, volumeSearchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave data volumes matching labels '%+v'", volumeSearchLabels)
	}
	if len(foundVolumes.Items) > 0 {
		return nil, stacktrace.NewError("Cannot create enclave with ID '%v' because one or more enclave data volumes for that enclave already exists", enclaveId)
	}

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
