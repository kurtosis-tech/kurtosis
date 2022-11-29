package test_helpers

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"strings"
)

const (
	newlineChar = "\n"
)

func ReadStreamContentUntilClosed(responseLines chan *kurtosis_core_rpc_api_bindings.StarlarkExecutionResponseLine) (string, []*kurtosis_core_rpc_api_bindings.StarlarkInstruction, *kurtosis_core_rpc_api_bindings.StarlarkInterpretationError, []*kurtosis_core_rpc_api_bindings.StarlarkValidationError, *kurtosis_core_rpc_api_bindings.StarlarkExecutionError) {
	scriptOutput := strings.Builder{}
	instructions := make([]*kurtosis_core_rpc_api_bindings.StarlarkInstruction, 0)
	var interpretationError *kurtosis_core_rpc_api_bindings.StarlarkInterpretationError
	validationErrors := make([]*kurtosis_core_rpc_api_bindings.StarlarkValidationError, 0)
	var executionError *kurtosis_core_rpc_api_bindings.StarlarkExecutionError

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
