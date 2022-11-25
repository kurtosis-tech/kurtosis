package test_helpers

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"strings"
)

func GenerateScriptOutput(instructions []*kurtosis_core_rpc_api_bindings.KurtosisInstruction) string {
	scriptOutput := strings.Builder{}
	for _, instruction := range instructions {
		if instruction.InstructionResult != nil {
			scriptOutput.WriteString(instruction.GetInstructionResult())
		}
	}
	return scriptOutput.String()
}

func ReadStreamContentUntilClosed(responseLines chan *kurtosis_core_rpc_api_bindings.KurtosisExecutionResponseLine) (*kurtosis_core_rpc_api_bindings.KurtosisInterpretationError, []*kurtosis_core_rpc_api_bindings.KurtosisValidationError, *kurtosis_core_rpc_api_bindings.KurtosisExecutionError, []*kurtosis_core_rpc_api_bindings.KurtosisInstruction) {
	var interpretationError *kurtosis_core_rpc_api_bindings.KurtosisInterpretationError
	validationErrors := make([]*kurtosis_core_rpc_api_bindings.KurtosisValidationError, 0)
	var executionError *kurtosis_core_rpc_api_bindings.KurtosisExecutionError
	instructions := make([]*kurtosis_core_rpc_api_bindings.KurtosisInstruction, 0)

	for responseLine := range responseLines {
		if responseLine.GetInstruction() != nil {
			instructions = append(instructions, responseLine.GetInstruction())
		} else if responseLine.GetError() != nil {
			if responseLine.GetError().GetInterpretationError() != nil {
				interpretationError = responseLine.GetError().GetInterpretationError()
			} else if responseLine.GetError().GetValidationError() != nil {
				validationErrors = append(validationErrors, responseLine.GetError().GetValidationError())
			} else if responseLine.GetError().GetExecutionError() != nil {
				executionError = responseLine.GetError().GetExecutionError()
			}
		}
	}

	return interpretationError, validationErrors, executionError, instructions
}
