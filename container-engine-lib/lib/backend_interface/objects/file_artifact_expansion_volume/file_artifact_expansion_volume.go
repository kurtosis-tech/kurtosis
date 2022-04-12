package file_artifact_expansion_volume

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
)

type FileArtifactExpansionVolumeName string

type FileArtifactExpansionVolume struct {
	name FileArtifactExpansionVolumeName
	serviceGuid service.ServiceGUID
	enclaveId   enclave.EnclaveID
}

func NewFileArtifactExpansionVolume(name FileArtifactExpansionVolumeName, serviceGuid service.ServiceGUID, enclaveId enclave.EnclaveID) *FileArtifactExpansionVolume {
	return &FileArtifactExpansionVolume{name: name, serviceGuid: serviceGuid, enclaveId: enclaveId}
}

func (expansionVolume *FileArtifactExpansionVolume) GetName() FileArtifactExpansionVolumeName {
	return expansionVolume.name
}

func (expansionVolume *FileArtifactExpansionVolume) GetServiceGUID() service.ServiceGUID {
	return expansionVolume.serviceGuid
}

func (expansionVolume *FileArtifactExpansionVolume) GetEnclaveID() enclave.EnclaveID {
	return expansionVolume.enclaveId
}

