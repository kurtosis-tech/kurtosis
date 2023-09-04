package enclave_plan

import "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"

// Use NewEnclavePlanCapabilitiesBuilder to construct this object
type EnclavePlanCapabilities struct {
	// we do this way in order to have exported fields which can be marshalled
	// and an unexported type for encapsulation
	privateEnclavePlanCapabilities *privateEnclavePlanCapabilities
}

type privateEnclavePlanCapabilities struct {
	InstructionTypeStr string
	ServiceName        service.ServiceName
	ServiceNames       []service.ServiceName
	ArtifactName       string
	FilesArtifactMD5   []byte
}

func (capabilities *EnclavePlanCapabilities) GetInstructionTypeStr() string {
	return capabilities.privateEnclavePlanCapabilities.InstructionTypeStr
}

func (capabilities *EnclavePlanCapabilities) GetServiceName() service.ServiceName {
	return capabilities.privateEnclavePlanCapabilities.ServiceName
}

func (capabilities *EnclavePlanCapabilities) GetServiceNames() []service.ServiceName {
	return capabilities.privateEnclavePlanCapabilities.ServiceNames
}

func (capabilities *EnclavePlanCapabilities) GetArtifactName() string {
	return capabilities.privateEnclavePlanCapabilities.ArtifactName
}

func (capabilities *EnclavePlanCapabilities) GetFilesArtifactMD5() []byte {
	return capabilities.privateEnclavePlanCapabilities.FilesArtifactMD5
}
