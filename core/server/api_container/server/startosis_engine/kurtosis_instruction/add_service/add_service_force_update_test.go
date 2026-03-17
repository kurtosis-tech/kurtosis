package add_service

import (
	"testing"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_plan_persistence"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
	"github.com/stretchr/testify/require"
)

func TestTryResolveWith_ForceUpdateTrue_ReturnsInstructionIsUpdate(t *testing.T) {
	serviceName := service.ServiceName("test-service")

	capabilities := &AddServiceCapabilities{
		serviceNetwork:               nil,
		runtimeValueStore:            nil,
		serviceName:                  serviceName,
		serviceConfig:                nil,
		readyCondition:               nil,
		packageId:                    "",
		packageContentProvider:       nil,
		packageReplaceOptions:        nil,
		imageVal:                     nil,
		interpretationTimeValueStore: nil,
		resultUuid:                   "",
		returnValue:                  nil,
		description:                  "",
		imageDownloadMode:            image_download_mode.ImageDownloadMode_Missing,
		forceUpdate:                  true,
	}

	enclaveComponents := enclave_structure.NewEnclaveComponents()

	otherInstruction := &enclave_plan_persistence.EnclavePlanInstruction{
		Uuid:           "test-uuid",
		Type:           AddServiceBuiltinName,
		StarlarkCode:   "",
		ReturnedValue:  "",
		ServiceNames:   []string{string(serviceName)},
		FilesArtifacts: nil,
	}

	// Even when instructions are "equal", force_update should return InstructionIsUpdate
	result := capabilities.TryResolveWith(true, otherInstruction, enclaveComponents)

	require.Equal(t, enclave_structure.InstructionIsUpdate, result)
}

func TestTryResolveWith_ForceUpdateTrue_InstructionsNotEqual_ReturnsInstructionIsUpdate(t *testing.T) {
	serviceName := service.ServiceName("test-service")

	capabilities := &AddServiceCapabilities{
		serviceNetwork:               nil,
		runtimeValueStore:            nil,
		serviceName:                  serviceName,
		serviceConfig:                nil,
		readyCondition:               nil,
		packageId:                    "",
		packageContentProvider:       nil,
		packageReplaceOptions:        nil,
		imageVal:                     nil,
		interpretationTimeValueStore: nil,
		resultUuid:                   "",
		returnValue:                  nil,
		description:                  "",
		imageDownloadMode:            image_download_mode.ImageDownloadMode_Missing,
		forceUpdate:                  true,
	}

	enclaveComponents := enclave_structure.NewEnclaveComponents()

	otherInstruction := &enclave_plan_persistence.EnclavePlanInstruction{
		Uuid:           "test-uuid",
		Type:           AddServiceBuiltinName,
		StarlarkCode:   "",
		ReturnedValue:  "",
		ServiceNames:   []string{string(serviceName)},
		FilesArtifacts: nil,
	}

	// When force_update is true and instructions are not equal, should return InstructionIsUpdate
	result := capabilities.TryResolveWith(false, otherInstruction, enclaveComponents)

	require.Equal(t, enclave_structure.InstructionIsUpdate, result)
}
