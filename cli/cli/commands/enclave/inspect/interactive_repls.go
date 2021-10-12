package inspect

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-core/commons/enclave_object_labels"
)

func printInteractiveRepls(ctx context.Context) error {

}

func getLabelsForListInteractiveRepls(enclaveId string) map[string]string {
	labels := map[string]string{}
	labels[enclave_object_labels.ContainerTypeLabel] = enclave_object_labels.ContainerTypeAPIContainer
	labels[enclave_object_labels.EnclaveIDContainerLabel] = enclaveId
	return labels
}
