package files_artifact_expander

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
)

type FilesArtifactExpanderFilters struct {
	// Disjunctive set of enclave IDs for which to return files artifact expanders
	// If nil or empty, will match all enclave IDs
	EnclaveIDs map[enclave.EnclaveID]bool

	// Disjunctive set of files artifact expander GUIDs to find files artifact expander for
	// If nil or empty, will match all GUIDs
	GUIDs map[FilesArtifactExpanderGUID]bool

	// Disjunctive set of statuses that returned files artifact expanders must conform to
	// If nil or empty, will match all statuses
	Statuses map[container_status.ContainerStatus]bool
}
