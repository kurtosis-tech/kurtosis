package enclave_plan_instruction

import (
	"encoding/json"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	enclave_plan_capabilities2 "github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/enclave_plan_capabilities"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	instructionNameFormat         = "test-instruction-%v"
	serviceNameStrFormat          = "my-service-test-%v"
	serviceNameAInList            = service.ServiceName("service-test-A")
	serviceNameBInList            = service.ServiceName("service-test-B")
	artifactNameStrFormat         = "my-artifact-test-%v"
	filesArtifactContentStrFormat = "files artifact N%v content"

	addServiceInstructionStrFormat = "add_service(name=\"%s\", config=ServiceConfig(image=\"kurtosistech/example-datastore-server\"))"
	returnedValueStrFormat         = "Service(hostname = \"{{kurtosis:%v892804afe9d407cb8975856047ed94b:hostname.runtime_value}}\", ip_address = \"{{kurtosis:%v892804afe9d407cb8975856047ed94b:ip_address.runtime_value}}\", name = \"%s\", ports = {})"
)

func TestEnclavePlanInstructionMarshallers(t *testing.T) {
	originalEnclavePlanInstruction := getEnclavePlanInstructionForTest(1)[0]

	marshaledEnclavePlanInstruciton, err := json.Marshal(originalEnclavePlanInstruction)
	require.NoError(t, err)
	require.NotNil(t, marshaledEnclavePlanInstruciton)

	// Suppressing exhaustruct requirement because we want an object with zero values
	// nolint: exhaustruct
	newEnclavePlanInstruction := &EnclavePlanInstructionImpl{}

	err = json.Unmarshal(marshaledEnclavePlanInstruciton, newEnclavePlanInstruction)
	require.NoError(t, err)

	require.Equal(t, originalEnclavePlanInstruction, newEnclavePlanInstruction)
}

func getEnclavePlanInstructionForTest(amount int) []*EnclavePlanInstructionImpl {

	allEnclavePlanInstructions := []*EnclavePlanInstructionImpl{}

	allEnclaveCapabilities := getEnclavePlanCapabilitiesForTest(amount)

	for _, enclavePlanCapabilities := range allEnclaveCapabilities {
		addServiceInstructionStr := fmt.Sprintf(addServiceInstructionStrFormat, enclavePlanCapabilities.GetServiceName())

		privatePlan := &privateEnclavePlanInstruction{
			KurtosisInstructionStr: addServiceInstructionStr,
			Capabilities:           enclavePlanCapabilities,
		}

		newEnclavePlanInstruction := &EnclavePlanInstructionImpl{
			privateEnclavePlanInstruction: privatePlan,
		}

		allEnclavePlanInstructions = append(allEnclavePlanInstructions, newEnclavePlanInstruction)
	}

	return allEnclavePlanInstructions
}

func getEnclavePlanCapabilitiesForTest(amount int) []*enclave_plan_capabilities2.EnclavePlanCapabilities {

	allEnclavePlanCapabilities := []*enclave_plan_capabilities2.EnclavePlanCapabilities{}

	for i := 1; i <= amount; i++ {
		instructionName := fmt.Sprintf(instructionNameFormat, i)
		serviceName := service.ServiceName(fmt.Sprintf(serviceNameStrFormat, i))
		artifactName := fmt.Sprintf(artifactNameStrFormat, i)
		filesArtifactContent := []byte(fmt.Sprintf(filesArtifactContentStrFormat, i))

		builder := enclave_plan_capabilities2.NewEnclavePlanCapabilitiesBuilder(instructionName)
		builder.WithServiceName(serviceName)
		builder.WithServiceNames([]service.ServiceName{
			serviceNameAInList,
			serviceNameBInList,
		})
		builder.WithArtifactName(artifactName)
		builder.WithFilesArtifactMD5(filesArtifactContent)

		newCapabilities := builder.Build()

		allEnclavePlanCapabilities = append(allEnclavePlanCapabilities, newCapabilities)
	}

	return allEnclavePlanCapabilities
}
