package test_helpers

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"strings"
)

const (
	newlineChar = "\n"
)

func ReadStreamContentUntilClosed(responseLines chan *kurtosis_core_rpc_api_bindings.KurtosisExecutionResponseLine) (string, []*kurtosis_core_rpc_api_bindings.KurtosisInstruction, *kurtosis_core_rpc_api_bindings.KurtosisInterpretationError, []*kurtosis_core_rpc_api_bindings.KurtosisValidationError, *kurtosis_core_rpc_api_bindings.KurtosisExecutionError) {
	scriptOutput := strings.Builder{}
	instructions := make([]*kurtosis_core_rpc_api_bindings.KurtosisInstruction, 0)
	var interpretationError *kurtosis_core_rpc_api_bindings.KurtosisInterpretationError
	validationErrors := make([]*kurtosis_core_rpc_api_bindings.KurtosisValidationError, 0)
	var executionError *kurtosis_core_rpc_api_bindings.KurtosisExecutionError

	for responseLine := range responseLines {
		if responseLine.GetInstruction() != nil {
			instructions = append(instructions, responseLine.GetInstruction())
		} else if responseLine.GetInstructionResult() != nil {
			scriptOutput.WriteString(responseLine.GetInstructionResult().GetSerializedInstructionResult())
			scriptOutput.WriteString(newlineChar)
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
	return scriptOutput.String(), instructions, interpretationError, validationErrors, executionError
}
