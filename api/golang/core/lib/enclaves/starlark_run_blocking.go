package enclaves

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"strings"
)

const (
	starlarkRunOutputLinesSplit = "\n"
)

type StarlarkRunMultilineOutput string

type StarlarkRunResult struct {
	RunOutput StarlarkRunMultilineOutput

	Instructions []*kurtosis_core_rpc_api_bindings.StarlarkInstruction

	InterpretationError *kurtosis_core_rpc_api_bindings.StarlarkInterpretationError

	ValidationErrors []*kurtosis_core_rpc_api_bindings.StarlarkValidationError

	ExecutionError *kurtosis_core_rpc_api_bindings.StarlarkExecutionError
}

func NewStarlarkRunResult(runOutput StarlarkRunMultilineOutput, instructions []*kurtosis_core_rpc_api_bindings.StarlarkInstruction, interpretationError *kurtosis_core_rpc_api_bindings.StarlarkInterpretationError, validationErrors []*kurtosis_core_rpc_api_bindings.StarlarkValidationError, executionError *kurtosis_core_rpc_api_bindings.StarlarkExecutionError) *StarlarkRunResult {
	return &StarlarkRunResult{
		RunOutput:           runOutput,
		Instructions:        instructions,
		InterpretationError: interpretationError,
		ValidationErrors:    validationErrors,
		ExecutionError:      executionError,
	}
}

func ReadStarlarkRunResponseLineBlocking(starlarkRunResponseLines <-chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine) *StarlarkRunResult {
	scriptOutput := strings.Builder{}
	instructions := make([]*kurtosis_core_rpc_api_bindings.StarlarkInstruction, 0)
	var interpretationError *kurtosis_core_rpc_api_bindings.StarlarkInterpretationError
	validationErrors := make([]*kurtosis_core_rpc_api_bindings.StarlarkValidationError, 0)
	var executionError *kurtosis_core_rpc_api_bindings.StarlarkExecutionError

	for responseLine := range starlarkRunResponseLines {
		if responseLine.GetInstruction() != nil {
			instructions = append(instructions, responseLine.GetInstruction())
		} else if responseLine.GetInstructionResult() != nil {
			scriptOutput.WriteString(responseLine.GetInstructionResult().GetSerializedInstructionResult())
			scriptOutput.WriteString(starlarkRunOutputLinesSplit)
		} else if responseLine.GetError() != nil {
			if responseLine.GetError().GetInterpretationError() != nil {
				interpretationError = responseLine.GetError().GetInterpretationError()
			} else if responseLine.GetError().GetValidationError() != nil {
				validationErrors = append(validationErrors, responseLine.GetError().GetValidationError())
			} else if responseLine.GetError().GetExecutionError() != nil {
				executionError = responseLine.GetError().GetExecutionError()
			}
		} else if responseLine.GetRunFinishedEvent() != nil {
			runFinishedEvent := responseLine.GetRunFinishedEvent()
			if runFinishedEvent.GetIsRunSuccessful() && runFinishedEvent.GetSerializedOutput() != "" {
				scriptOutput.WriteString(runFinishedEvent.GetSerializedOutput())
				scriptOutput.WriteString(starlarkRunOutputLinesSplit)
			}
		}
	}
	return NewStarlarkRunResult(
		StarlarkRunMultilineOutput(scriptOutput.String()),
		instructions,
		interpretationError,
		validationErrors,
		executionError)
}
