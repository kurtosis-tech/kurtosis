package engine_labels_schema

import "github.com/kurtosis-tech/kurtosis-core/commons/schema"

const (
	// TODO This needs to be merged with the labels in API container, and centralized into a labels library!
	ContainerTypeKurtosisEngine = "kurtosis-engine"
)

// !!!!! WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING !!!!!
// If you're thinking about changing this, be VERY careful - it's
//  possible to get in a situation where the user has an engine container running with the old labels,
//  they upgrade their CLI to your new version with the new labels, and their new CLI's 'enclave stop'
//  command can no longer find the old server to stop it (thereby leaking a running container!)
// !!!!! WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING !!!!!
var EngineContainerLabels = map[string]string{
	// TODO These need refactoring!!! "ContainerTypeLabel" and "AppIDLabel" aren't just for enclave objects!!!
	//  See https://github.com/kurtosis-tech/kurtosis-cli/issues/24
	schema.AppIDLabel:         schema.AppIDValue,
	schema.ContainerTypeLabel: ContainerTypeKurtosisEngine,
}
