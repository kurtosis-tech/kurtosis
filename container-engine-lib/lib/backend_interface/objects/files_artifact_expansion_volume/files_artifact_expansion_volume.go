package files_artifact_expansion_volume

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
)

type FilesArtifactExpansionGUID string
type FilesArtifactExpansionVolumeName string

type FilesArtifactExpansionVolume struct {
	guid		FilesArtifactExpansionGUID
	name        FilesArtifactExpansionVolumeName
	enclaveId   enclave.EnclaveID
}

func NewFilesArtifactExpansionVolume(name FilesArtifactExpansionVolumeName, enclaveId enclave.EnclaveID) *FilesArtifactExpansionVolume {
	return &FilesArtifactExpansionVolume{name: name, enclaveId: enclaveId}
}

func (expansionVolume *FilesArtifactExpansionVolume) GetName() FilesArtifactExpansionVolumeName {
	return expansionVolume.name
}

func (expansionVolume *FilesArtifactExpansionVolume) GetEnclaveID() enclave.EnclaveID {
	return expansionVolume.enclaveId
}

