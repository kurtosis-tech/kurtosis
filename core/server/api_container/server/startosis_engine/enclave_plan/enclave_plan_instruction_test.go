package enclave_plan

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_plan_capabilities"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEnclavePlanInstructionMarshallers(t *testing.T) {
	originalEnclavePlanInstruction := getEnclavePlanInstructionForTest()

	marshaledEnclavePlanInstruciton, err := json.Marshal(originalEnclavePlanInstruction)
	require.NoError(t, err)
	require.NotNil(t, marshaledEnclavePlanInstruciton)

	// Suppressing exhaustruct requirement because we want an object with zero values
	// nolint: exhaustruct
	newEnclavePlanInstruction := &EnclavePlanInstruction{}

	err = json.Unmarshal(marshaledEnclavePlanInstruciton, newEnclavePlanInstruction)
	require.NoError(t, err)

	require.Equal(t, originalEnclavePlanInstruction, newEnclavePlanInstruction)
}

func getEnclavePlanInstructionForTest() *EnclavePlanInstruction {

	enclavePlanCapabilities := getEnclavePlanCapabilitiesForTest()

	privatePlan := &privateEnclavePlanInstruction{
		KurtosisInstructionStr: "add_service(name=\"test-service-name\", config=ServiceConfig(image=\"kurtosistech/example-datastore-server\"))",
		Capabilities:           enclavePlanCapabilities,
		ReturnedValueStr:       "Service(hostname = \"{{kurtosis:7892804afe9d407cb8975856047ed94b:hostname.runtime_value}}\", ip_address = \"{{kurtosis:7892804afe9d407cb8975856047ed94b:ip_address.runtime_value}}\", name = \"test-service-name\", ports = {})",
	}

	return &EnclavePlanInstruction{
		privateEnclavePlanInstruction: privatePlan,
	}
}

func getEnclavePlanCapabilitiesForTest() *enclave_plan_capabilities.EnclavePlanCapabilities {

	builder := enclave_plan_capabilities.NewEnclavePlanCapabilitiesBuilder("test-instruction")
	builder.WitServiceName("my-service-test")
	builder.WitServiceNames([]service.ServiceName{
		"my-service-test-1",
		"my-service-test-2",
		"my-service-test-3",
	})
	builder.WithArtifactName("my-artifact-test")
	builder.WithFilesArtifactMD5([]byte("files artifact content"))

	capabilities := builder.Build()

	return capabilities
}
