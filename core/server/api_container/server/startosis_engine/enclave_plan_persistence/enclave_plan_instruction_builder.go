package enclave_plan_persistence

import (
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
)

type EnclavePlanInstructionBuilder struct {
	uuid string

	instructionType string

	starlarkCode string

	returnedValue string

	serviceNames []string

	filesArtifacts map[string][]byte
}

func NewEnclavePlanInstructionBuilder() *EnclavePlanInstructionBuilder {
	return &EnclavePlanInstructionBuilder{
		uuid:            "",
		instructionType: "",
		starlarkCode:    "",
		returnedValue:   "",
		serviceNames:    []string{},
		filesArtifacts:  map[string][]byte{},
	}
}

func (builder *EnclavePlanInstructionBuilder) SetUuid(uuid string) *EnclavePlanInstructionBuilder {
	builder.uuid = uuid
	return builder
}

func (builder *EnclavePlanInstructionBuilder) SetType(instructionType string) *EnclavePlanInstructionBuilder {
	builder.instructionType = instructionType
	return builder
}

func (builder *EnclavePlanInstructionBuilder) SetStarlarkCode(starlarkCode string) *EnclavePlanInstructionBuilder {
	builder.starlarkCode = starlarkCode
	return builder
}

func (builder *EnclavePlanInstructionBuilder) SetReturnedValue(serializedReturnedValue string) *EnclavePlanInstructionBuilder {
	builder.returnedValue = serializedReturnedValue
	return builder
}

func (builder *EnclavePlanInstructionBuilder) AddServiceName(serviceName service.ServiceName) *EnclavePlanInstructionBuilder {
	builder.serviceNames = append(builder.serviceNames, string(serviceName))
	return builder
}

func (builder *EnclavePlanInstructionBuilder) AddFilesArtifact(filesArtifactName string, filesArtifactMd5 []byte) *EnclavePlanInstructionBuilder {
	builder.filesArtifacts[filesArtifactName] = filesArtifactMd5
	return builder
}

func (builder *EnclavePlanInstructionBuilder) Build() (*EnclavePlanInstruction, error) {
	if builder.uuid == "" || builder.instructionType == "" || builder.starlarkCode == "" || builder.returnedValue == "" {
		return nil, stacktrace.NewError("Some required attributes aren't set on this builder")
	}
	return &EnclavePlanInstruction{
		Uuid:           builder.uuid,
		Type:           builder.instructionType,
		StarlarkCode:   builder.starlarkCode,
		ReturnedValue:  builder.returnedValue,
		ServiceNames:   builder.serviceNames,
		FilesArtifacts: builder.filesArtifacts,
	}, nil
}
