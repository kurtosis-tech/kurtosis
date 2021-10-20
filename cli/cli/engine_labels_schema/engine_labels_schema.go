package engine_labels_schema

import "github.com/kurtosis-tech/kurtosis-core/commons/enclave_object_labels"

const (
	// TODO This needs to be merged with the labels in API container, and centralized into a labels library!
	ContainerTypeKurtosisEngine = "kurtosis-engine"
)

var EngineContainerLabels = map[string]string{
	// TODO These need refactoring!!! "ContainerTypeLabel" and "AppIDLabel" aren't just for enclave objects!!!
	//  See https://github.com/kurtosis-tech/kurtosis-cli/issues/24
	enclave_object_labels.AppIDLabel: enclave_object_labels.AppIDValue,
	enclave_object_labels.ContainerTypeLabel: ContainerTypeKurtosisEngine,
}
