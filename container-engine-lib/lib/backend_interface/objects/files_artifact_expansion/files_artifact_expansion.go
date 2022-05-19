package files_artifact_expansion

import "github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"

type FilesArtifactExpansionGUID string

type FilesArtifactExpansion struct {
	guid FilesArtifactExpansionGUID
	serviceGUID service.ServiceGUID
}
