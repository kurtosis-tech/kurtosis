package enclave_plan_persistence

import (
	"bytes"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/service"
)

type EnclavePlanInstruction struct {
	Uuid string `json:"uuid"`

	Type string `json:"type"`

	StarlarkCode string `json:"starlarkCode"`

	ReturnedValue string `json:"returnedValue"`

	ServiceNames []string `json:"serviceNames"`

	// mapping between files artifact name and files artifact MD5
	FilesArtifacts map[string][]byte `json:"filesArtifacts"` // FSA byte arrays are automatically serialized as base64 encoded strings
}

// HasOnlyServiceName is a convenience function that returns true if the enclave plan instruction has only
// one service and its name is equal to the parameter of this function
func (enclavePlanInstruction *EnclavePlanInstruction) HasOnlyServiceName(serviceName service.ServiceName) bool {
	if len(enclavePlanInstruction.ServiceNames) != 1 {
		return false
	}
	return enclavePlanInstruction.ServiceNames[0] == string(serviceName)
}

// HasOnlyFilesArtifactName is a convenience function that returns true if the enclave plan instruction has only
// one files artifact and its name is equal to the parameter of this function
func (enclavePlanInstruction *EnclavePlanInstruction) HasOnlyFilesArtifactName(filesArtifactName string) bool {
	if len(enclavePlanInstruction.FilesArtifacts) != 1 {
		return false
	}
	_, found := enclavePlanInstruction.FilesArtifacts[filesArtifactName]
	return found
}

// HasOnlyFilesArtifactMd5 is a convenience function that returns true if the enclave plan instruction has only
// one files artifact and its MD5 is equal to the parameter of this function
func (enclavePlanInstruction *EnclavePlanInstruction) HasOnlyFilesArtifactMd5(filesArtifactMd5 []byte) bool {
	if len(enclavePlanInstruction.FilesArtifacts) != 1 {
		return false
	}
	for _, filesArtifactMd5FromInstruction := range enclavePlanInstruction.FilesArtifacts {
		if bytes.Equal(filesArtifactMd5, filesArtifactMd5FromInstruction) {
			return true
		}
	}
	return false
}

func (enclavePlanInstruction *EnclavePlanInstruction) Clone() *EnclavePlanInstruction {
	clonedServiceNames := make([]string, len(enclavePlanInstruction.ServiceNames))
	copy(enclavePlanInstruction.ServiceNames, clonedServiceNames)

	clonedFilesArtifacts := make(map[string][]byte, len(enclavePlanInstruction.FilesArtifacts))
	for filesArtifactName, filesArtifactMd5 := range enclavePlanInstruction.FilesArtifacts {
		clonedFilesArtifactMd5 := make([]byte, len(filesArtifactMd5))
		copy(filesArtifactMd5, clonedFilesArtifactMd5)
		clonedFilesArtifacts[filesArtifactName] = clonedFilesArtifactMd5
	}
	return &EnclavePlanInstruction{
		Uuid:           enclavePlanInstruction.Uuid,
		Type:           enclavePlanInstruction.Type,
		StarlarkCode:   enclavePlanInstruction.StarlarkCode,
		ReturnedValue:  enclavePlanInstruction.ReturnedValue,
		ServiceNames:   clonedServiceNames,
		FilesArtifacts: clonedFilesArtifacts,
	}
}
