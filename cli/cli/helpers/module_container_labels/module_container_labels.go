package module_container_labels

import (
	"github.com/kurtosis-tech/object-attributes-schema-lib/forever_constants"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
)

func GetModuleContainerLabelsWithEnclaveIDAndModuleId(enclaveId string, moduleId string) map[string]string {
	labels := map[string]string{}
	labels[forever_constants.ContainerTypeLabel] = schema.ContainerTypeModuleContainer
	labels[schema.EnclaveIDContainerLabel] = enclaveId
	labels[schema.IDLabel] = moduleId
	return labels
}
