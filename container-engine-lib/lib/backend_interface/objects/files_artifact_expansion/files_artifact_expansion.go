package files_artifact_expansion

// TODO TODO DELETE THIS ENTIRE PACKAGE!

import "github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"

type FilesArtifactExpansionGUID string

type FilesArtifactExpansion struct {
	guid FilesArtifactExpansionGUID
	serviceGUID service.ServiceGUID
}

func NewFilesArtifactExpansion(guid FilesArtifactExpansionGUID, serviceGUID service.ServiceGUID) *FilesArtifactExpansion {
	return &FilesArtifactExpansion{guid: guid, serviceGUID: serviceGUID}
}

func (filesArtifactExpansion *FilesArtifactExpansion) GetGUID() FilesArtifactExpansionGUID {
	return filesArtifactExpansion.guid
}

func (filesArtifactExpansion *FilesArtifactExpansion) GetServiceGUID() service.ServiceGUID {
	return filesArtifactExpansion.serviceGUID
}