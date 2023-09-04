package enclave_plan

type EnclavePlanInstruction struct {
	kurtosisInstructionStr string
	capabilities           *EnclavePlanCapabilities
	returnedValueStr       string
}

func NewEnclavePlanInstruction(
	kurtosisInstructionStr string,
	capabilities *EnclavePlanCapabilities,
) *EnclavePlanInstruction {
	instruction := &EnclavePlanInstruction{
		kurtosisInstructionStr: kurtosisInstructionStr,
		capabilities:           capabilities,
	}
	return instruction
}

func (data *EnclavePlanInstruction) GetKurtosisInstructionStr() string {
	return data.kurtosisInstructionStr
}

func (data *EnclavePlanInstruction) GetCapabilities() *EnclavePlanCapabilities {
	return data.capabilities
}
