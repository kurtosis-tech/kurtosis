package file_artifact_expansion_volume

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
)

type FileArtifactExpansionVolumeFilters struct {
	// Disjunctive set of file artifact expansion
	// volume names to find file artifact expansion
	// volumes for
	// If nil or empty, will match all names
	Names map[FileArtifactExpansionVolumeName]bool

	// Disjunctive set of user service GUIDs to find
	// file artifact expansion volumes for
	// If nil or empty, will match all GUIDs
	UserServiceGUIDs map[service.ServiceGUID]bool

	// Disjunctive set of enclave IDs for which to return
	// file artifact expansion volumes
	// If nil or empty, will match all enclave IDs
	EnclaveIDs map[enclave.EnclaveID]bool
}
