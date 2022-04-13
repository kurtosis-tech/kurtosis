package files_artifact_expansion_volume

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
)

type FilesArtifactExpansionVolumeFilters struct {
	// Disjunctive set of files artifact expansion
	// volume names to find files artifact expansion
	// volumes for
	// If nil or empty, will match all names
	Names map[FilesArtifactExpansionVolumeName]bool

	// Disjunctive set of enclave IDs for which to return
	// files artifact expansion volumes
	// If nil or empty, will match all enclave IDs
	EnclaveIDs map[enclave.EnclaveID]bool
}
