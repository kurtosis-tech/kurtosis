package enclave_plan_persistence

type EnclavePlanInstruction struct {
	Uuid string `json:"uuid"`

	Type string `json:"type"`

	StarlarkCode string `json:"starlarkCode"`

	ReturnedValue string `json:"returnedValue"`

	ServiceNames []string `json:"serviceNames"`

	FilesArtifactNames []string `json:"filesArtifactNames"`

	FilesArtifactMd5s [][]byte `json:"filesArtifactMd5s"` // FSA byte arrays are automatically serialized as base64 encoded strings
}

func NewEnclavePlanInstruction(uuid string, instructionType string, starlarkCode string, returnedValue string, serviceNames []string, filesArtifactNames []string, filesArtifactMd5s [][]byte) *EnclavePlanInstruction {
	return &EnclavePlanInstruction{
		Uuid:               uuid,
		Type:               instructionType,
		StarlarkCode:       starlarkCode,
		ReturnedValue:      returnedValue,
		ServiceNames:       serviceNames,
		FilesArtifactNames: filesArtifactNames,
		FilesArtifactMd5s:  filesArtifactMd5s,
	}
}

func (enclavePlanInstruction *EnclavePlanInstruction) Clone() *EnclavePlanInstruction {
	return &EnclavePlanInstruction{
		Uuid:               enclavePlanInstruction.Uuid,
		Type:               enclavePlanInstruction.Type,
		StarlarkCode:       enclavePlanInstruction.StarlarkCode,
		ReturnedValue:      enclavePlanInstruction.ReturnedValue,
		ServiceNames:       enclavePlanInstruction.ServiceNames,
		FilesArtifactNames: enclavePlanInstruction.FilesArtifactNames,
		FilesArtifactMd5s:  enclavePlanInstruction.FilesArtifactMd5s,
	}
}
