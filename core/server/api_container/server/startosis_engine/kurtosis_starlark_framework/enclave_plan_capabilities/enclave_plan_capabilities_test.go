package enclave_plan_capabilities

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	filesArtifactContentStr = "my-artifact-test-content"
)

func TestEnclavePlanCapabilitiesMarshallers(t *testing.T) {
	originalEnclavePlanCapabilities := getEnclavePlanCapabilitiesForTest()

	marshaledEnclavePlanCapabilities, err := json.Marshal(originalEnclavePlanCapabilities)
	require.NoError(t, err)
	require.NotNil(t, marshaledEnclavePlanCapabilities)

	// Suppressing exhaustruct requirement because we want an object with zero values
	// nolint: exhaustruct
	newEnclavePlanCapabilities := &EnclavePlanCapabilities{}

	err = json.Unmarshal(marshaledEnclavePlanCapabilities, newEnclavePlanCapabilities)
	require.NoError(t, err)

	require.Equal(t, originalEnclavePlanCapabilities, newEnclavePlanCapabilities)
	require.Equal(t, filesArtifactContentStr, string(newEnclavePlanCapabilities.GetFilesArtifactMD5()))
}

func TestEnclavePlanCapabilitiesMarshallers_WithAllZeroValues(t *testing.T) {
	originalEnclavePlanCapabilities := &EnclavePlanCapabilities{
		// Suppressing exhaustruct requirement because we want an object with zero values
		// nolint: exhaustruct
		privateEnclavePlanCapabilities: &privateEnclavePlanCapabilities{},
	}

	marshaledEnclavePlanCapabilities, err := json.Marshal(originalEnclavePlanCapabilities)
	require.NoError(t, err)
	require.NotNil(t, marshaledEnclavePlanCapabilities)

	// Suppressing exhaustruct requirement because we want an object with zero values
	// nolint: exhaustruct
	newEnclavePlanCapabilities := &EnclavePlanCapabilities{}

	err = json.Unmarshal(marshaledEnclavePlanCapabilities, newEnclavePlanCapabilities)
	require.NoError(t, err)

	require.Equal(t, originalEnclavePlanCapabilities, newEnclavePlanCapabilities)
}

func getEnclavePlanCapabilitiesForTest() *EnclavePlanCapabilities {
	privateCapabilities := &privateEnclavePlanCapabilities{
		InstructionTypeStr: "test-instruction",
		ServiceNames: []service.ServiceName{
			"my-service-test-1",
			"my-service-test-2",
			"my-service-test-3",
		},
		ServiceName:      "my-service-test",
		ArtifactName:     "my-artifact-test",
		FilesArtifactMD5: []byte(filesArtifactContentStr),
	}
	return &EnclavePlanCapabilities{
		privateEnclavePlanCapabilities: privateCapabilities,
	}
}
