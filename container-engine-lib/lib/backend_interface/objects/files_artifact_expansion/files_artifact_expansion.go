package files_artifact_expansion

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
)

type FilesArtifactExpansionGUID string
type FilesArtifactExpanderGUID string
type FilesArtifactExpansionVolumeName string

type FilesArtifactExpansion struct {
	guid FilesArtifactExpansionGUID
	expansionVolumeName FilesArtifactExpansionVolumeName
	expanderGUID FilesArtifactExpanderGUID
	expanderStatus container_status.ContainerStatus
}
