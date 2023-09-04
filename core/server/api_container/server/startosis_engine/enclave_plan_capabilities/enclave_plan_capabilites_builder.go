package enclave_plan_capabilities

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
)

type EnclavePlanCapabilitiesBuilder struct {
	instructionTypeStr string
	serviceName        service.ServiceName
	serviceNames       []service.ServiceName
	artifactName       string
	filesArtifactMD5   []byte
}

func NewEnclavePlanCapabilitiesBuilder(
	instructionTypeStr string,
) *EnclavePlanCapabilitiesBuilder {
	builder := &EnclavePlanCapabilitiesBuilder{
		instructionTypeStr: instructionTypeStr,
	}
	return builder
}

func (builder *EnclavePlanCapabilitiesBuilder) WitServiceName(serviceName service.ServiceName) *EnclavePlanCapabilitiesBuilder {
	builder.serviceName = serviceName
	return builder
}

func (builder *EnclavePlanCapabilitiesBuilder) WitServiceNames(serviceNames []service.ServiceName) *EnclavePlanCapabilitiesBuilder {
	builder.serviceNames = serviceNames
	return builder
}

func (builder *EnclavePlanCapabilitiesBuilder) WithArtifactName(artifactName string) *EnclavePlanCapabilitiesBuilder {
	builder.artifactName = artifactName
	return builder
}

func (builder *EnclavePlanCapabilitiesBuilder) WithFilesArtifactMD5(filesArtifactMD5 []byte) *EnclavePlanCapabilitiesBuilder {
	builder.filesArtifactMD5 = filesArtifactMD5
	return builder
}

func (builder *EnclavePlanCapabilitiesBuilder) Build() *EnclavePlanCapabilities {
	privateCapabilities := &privateEnclavePlanCapabilities{
		InstructionTypeStr: builder.instructionTypeStr,
		ServiceName:        builder.serviceName,
		ServiceNames:       builder.serviceNames,
		ArtifactName:       builder.artifactName,
		FilesArtifactMD5:   builder.filesArtifactMD5,
	}
	return &EnclavePlanCapabilities{
		privateEnclavePlanCapabilities: privateCapabilities,
	}
}
