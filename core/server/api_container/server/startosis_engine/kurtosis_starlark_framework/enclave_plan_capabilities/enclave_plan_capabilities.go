package enclave_plan_capabilities

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
)

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

func (capabilities *EnclavePlanCapabilities) MarshalJSON() ([]byte, error) {
	return json.Marshal(capabilities.privateEnclavePlanCapabilities)
}

func (capabilities *EnclavePlanCapabilities) UnmarshalJSON(data []byte) error {

	// Suppressing exhaustruct requirement because we want an object with zero values
	// nolint: exhaustruct
	unmarshalledPrivateStructPtr := &privateEnclavePlanCapabilities{}

	if err := json.Unmarshal(data, unmarshalledPrivateStructPtr); err != nil {
		return stacktrace.Propagate(err, "An error occurred unmarshalling the private struct")
	}

	capabilities.privateEnclavePlanCapabilities = unmarshalledPrivateStructPtr
	return nil
}
