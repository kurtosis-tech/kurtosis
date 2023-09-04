package enclave_plan

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_plan_capabilities"
	"github.com/kurtosis-tech/stacktrace"
)

type EnclavePlanInstruction struct {
	// we do this way in order to have exported fields which can be marshalled
	// and an unexported type for encapsulation
	privateEnclavePlanInstruction *privateEnclavePlanInstruction
}

type privateEnclavePlanInstruction struct {
	KurtosisInstructionStr string
	Capabilities           *enclave_plan_capabilities.EnclavePlanCapabilities
	ReturnedValueStr       string
}

func NewEnclavePlanInstruction(
	kurtosisInstructionStr string,
	capabilities *enclave_plan_capabilities.EnclavePlanCapabilities,
	returnedValueStr string,
) *EnclavePlanInstruction {
	privatePlan := &privateEnclavePlanInstruction{
		KurtosisInstructionStr: kurtosisInstructionStr,
		Capabilities:           capabilities,
		ReturnedValueStr:       returnedValueStr,
	}
	return &EnclavePlanInstruction{
		privateEnclavePlanInstruction: privatePlan,
	}
}

func (instruction *EnclavePlanInstruction) GetKurtosisInstructionStr() string {
	return instruction.privateEnclavePlanInstruction.KurtosisInstructionStr
}

func (instruction *EnclavePlanInstruction) GetCapabilities() *enclave_plan_capabilities.EnclavePlanCapabilities {
	return instruction.privateEnclavePlanInstruction.Capabilities
}

func (instruction *EnclavePlanInstruction) GetReturnedValueStr() string {
	return instruction.privateEnclavePlanInstruction.ReturnedValueStr
}

func (instruction *EnclavePlanInstruction) MarshalJSON() ([]byte, error) {
	return json.Marshal(instruction.privateEnclavePlanInstruction)
}

func (instruction *EnclavePlanInstruction) UnmarshalJSON(data []byte) error {

	// Suppressing exhaustruct requirement because we want an object with zero values
	// nolint: exhaustruct
	unmarshalledPrivateStructPtr := &privateEnclavePlanInstruction{}

	if err := json.Unmarshal(data, unmarshalledPrivateStructPtr); err != nil {
		return stacktrace.Propagate(err, "An error occurred unmarshalling the private struct")
	}

	instruction.privateEnclavePlanInstruction = unmarshalledPrivateStructPtr
	return nil
}
