package files_artifact_expander

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact"
)

type FilesArtifactExpanderGUID string

type FilesArtifactExpander struct {
	guid FilesArtifactExpanderGUID
	enclaveId enclave.EnclaveID
	filesArtifactId files_artifact.FilesArtifactID
	status container_status.ContainerStatus
}

func NewFilesArtifactExpander(guid FilesArtifactExpanderGUID, enclaveId enclave.EnclaveID, filesArtifactId files_artifact.FilesArtifactID, status container_status.ContainerStatus) *FilesArtifactExpander {
	return &FilesArtifactExpander{guid: guid, enclaveId: enclaveId, filesArtifactId: filesArtifactId, status: status}
}

func (expander *FilesArtifactExpander) GetGUID() FilesArtifactExpanderGUID {
	return expander.guid
}

func (expander *FilesArtifactExpander) GetEnclaveID() enclave.EnclaveID {
	return expander.enclaveId
}

func (expander *FilesArtifactExpander) GetFilesArtifactID() files_artifact.FilesArtifactID {
	return expander.filesArtifactId
}

func (expander *FilesArtifactExpander) GetStatus() container_status.ContainerStatus {
	return expander.status
}
