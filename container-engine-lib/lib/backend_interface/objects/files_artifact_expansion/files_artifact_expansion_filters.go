package files_artifact_expansion

import "github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"

type FilesArtifactExpansionFilters struct {
	// Disjunctive set of files artifact expansion GUIDs to find files artifact expansion for
	// If nil or empty, will match all GUIDs
	GUIDs map[FilesArtifactExpansionGUID]bool

	// Disjunctive set of expander container statuses that returned files artifact expnasions must conform to
	// If nil or empty, will match all statuses
	ServiceGUIDs map[service.ServiceGUID]bool
}
