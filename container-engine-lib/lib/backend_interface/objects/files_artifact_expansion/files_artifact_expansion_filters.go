package files_artifact_expansion

import "github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"

type FilesArtifactExpansionFilters struct {
	// Disjunctive set of files artifact expansion GUIDs to find files artifact expansion for
	// If nil or empty, will match all GUIDs
	GUIDs map[FilesArtifactExpansionGUID]bool

	// Disjunctive set of serviceGUIDs that returned files artifact expansions must have
	// If nil or empty, will match all serviceGUIDs
	ServiceGUIDs map[service.ServiceGUID]bool
}
