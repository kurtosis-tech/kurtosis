package service_container_labels_by_enclave_id

import (
	"github.com/kurtosis-tech/object-attributes-schema-lib/forever_constants"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
)

func GetUserServiceContainerLabelsWithEnclaveID(enclaveId string) map[string]string {
	labels := map[string]string{}
	labels[forever_constants.ContainerTypeLabel] = schema.ContainerTypeUserServiceContainer
	labels[schema.EnclaveIDContainerLabel] = enclaveId
	return labels
}
