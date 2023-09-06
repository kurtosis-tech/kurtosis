package instructions_plan

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_plan_capabilities"
	"github.com/kurtosis-tech/stacktrace"
)

type EnclavePlanInstructionImpl struct {
	// we do this way in order to have exported fields which can be marshalled
	// and an unexported type for encapsulation
	privateEnclavePlanInstruction *privateEnclavePlanInstruction
}

type privateEnclavePlanInstruction struct {
	KurtosisInstructionStr string
	Capabilities           *enclave_plan_capabilities.EnclavePlanCapabilities
	ReturnedValueStr       string
}

func NewEnclavePlanInstructionImpl(
	kurtosisInstructionStr string,
	capabilities *enclave_plan_capabilities.EnclavePlanCapabilities,
	returnedValueStr string,
) *EnclavePlanInstructionImpl {
	privatePlan := &privateEnclavePlanInstruction{
		KurtosisInstructionStr: kurtosisInstructionStr,
		Capabilities:           capabilities,
		ReturnedValueStr:       returnedValueStr,
	}
	return &EnclavePlanInstructionImpl{
		privateEnclavePlanInstruction: privatePlan,
	}
}

func (instruction *EnclavePlanInstructionImpl) GetKurtosisInstructionStr() string {
	return instruction.privateEnclavePlanInstruction.KurtosisInstructionStr
}

func (instruction *EnclavePlanInstructionImpl) GetCapabilities() *enclave_plan_capabilities.EnclavePlanCapabilities {
	return instruction.privateEnclavePlanInstruction.Capabilities
}

func (instruction *EnclavePlanInstructionImpl) GetReturnedValueStr() string {
	return instruction.privateEnclavePlanInstruction.ReturnedValueStr
}

func (instruction *EnclavePlanInstructionImpl) MarshalJSON() ([]byte, error) {
	return json.Marshal(instruction.privateEnclavePlanInstruction)
}

func (instruction *EnclavePlanInstructionImpl) UnmarshalJSON(data []byte) error {

	// Suppressing exhaustruct requirement because we want an object with zero values
	// nolint: exhaustruct
	unmarshalledPrivateStructPtr := &privateEnclavePlanInstruction{}

	if err := json.Unmarshal(data, unmarshalledPrivateStructPtr); err != nil {
		return stacktrace.Propagate(err, "An error occurred unmarshalling the private struct")
	}

	instruction.privateEnclavePlanInstruction = unmarshalledPrivateStructPtr
	return nil
}
