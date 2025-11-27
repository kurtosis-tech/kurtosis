package to_http

import (
	"github.com/dzobbe/PoTE-kurtosis/engine/server/engine/utils"
	"github.com/kurtosis-tech/stacktrace"

	rpc_api "github.com/dzobbe/PoTE-kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	api_type "github.com/dzobbe/PoTE-kurtosis/api/golang/http_rest/api_types"
)

func ToHttpStarlarkRunResponseLine(rpc_value *rpc_api.StarlarkRunResponseLine) (*api_type.StarlarkRunResponseLine, error) {
	var http_type api_type.StarlarkRunResponseLine

	if runError := rpc_value.GetError(); runError != nil {
		value, err := ToHttpStarlarkError(runError)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to serialize %T", runError)
		}
		if value != nil {
			if err := http_type.FromStarlarkError(*value); err != nil {
				return nil, stacktrace.Propagate(err, "failed to serialize %T", runError)
			}
			return &http_type, nil
		}
		return nil, nil
	}

	if runInfo := rpc_value.GetInfo(); runInfo != nil {
		if err := http_type.FromStarlarkInfo(ToHttpStarlarkInfo(runInfo)); err != nil {
			return nil, stacktrace.Propagate(err, "failed to serialize %T", runInfo)
		}
		return &http_type, nil
	}

	if runInstruction := rpc_value.GetInstruction(); runInstruction != nil {
		if err := http_type.FromStarlarkInstruction(ToHttpStarlarkInstruction(runInstruction)); err != nil {
			return nil, stacktrace.Propagate(err, "failed to serialize %T", runInstruction)
		}
		return &http_type, nil
	}

	if runInstructionResult := rpc_value.GetInstructionResult(); runInstructionResult != nil {
		if err := http_type.FromStarlarkInstructionResult(ToHttpStarlarkInstructionResult(runInstructionResult)); err != nil {
			return nil, stacktrace.Propagate(err, "failed to serialize %T", runInstructionResult)
		}
		return &http_type, nil
	}

	if runProgressInfo := rpc_value.GetProgressInfo(); runProgressInfo != nil {
		if err := http_type.FromStarlarkRunProgress(ToHttpStarlarkProgressInfo(runProgressInfo)); err != nil {
			return nil, stacktrace.Propagate(err, "failed to serialize %T", runProgressInfo)
		}
		return &http_type, nil
	}

	if runWarning := rpc_value.GetWarning(); runWarning != nil {
		if err := http_type.FromStarlarkWarning(ToHttpStarlarkWarning(runWarning)); err != nil {
			return nil, stacktrace.Propagate(err, "failed to serialize %T", runWarning)
		}
		return &http_type, nil
	}

	if runFinishedEvent := rpc_value.GetRunFinishedEvent(); runFinishedEvent != nil {
		if err := http_type.FromStarlarkRunFinishedEvent(ToHttpStarlarkRunFinishedEvent(runFinishedEvent)); err != nil {
			return nil, stacktrace.Propagate(err, "failed to serialize %T", runFinishedEvent)
		}
		return &http_type, nil
	}

	err := stacktrace.NewError("Unmatched gRPC to Http mapping, source type: %T", rpc_value)
	return nil, err
}

func ToHttpStarlarkError(rpc_value *rpc_api.StarlarkError) (*api_type.StarlarkError, error) {
	var http_type api_type.StarlarkError
	if runError := rpc_value.GetExecutionError(); runError != nil {
		if err := http_type.Error.FromStarlarkExecutionError(ToHttpStarlarkExecutionError(runError)); err != nil {
			return nil, stacktrace.Propagate(err, "failed to serialize %T", runError)
		}
		return &http_type, nil
	}

	if runError := rpc_value.GetInterpretationError(); runError != nil {
		if err := http_type.Error.FromStarlarkInterpretationError(ToHttpStarlarkInterpretationError(runError)); err != nil {
			return nil, stacktrace.Propagate(err, "failed to serialize %T", runError)
		}
		return &http_type, nil
	}

	if runError := rpc_value.GetValidationError(); runError != nil {
		if err := http_type.Error.FromStarlarkValidationError(ToHttpStarlarkValidationError(runError)); err != nil {
			return nil, stacktrace.Propagate(err, "failed to serialize %T", runError)
		}
		return &http_type, nil
	}

	err := stacktrace.NewError("Unmatched gRPC to Http mapping, source type: %T", rpc_value)
	return nil, err
}

func ToHttpStarlarkExecutionError(rpc_value *rpc_api.StarlarkExecutionError) api_type.StarlarkExecutionError {
	var http_type api_type.StarlarkExecutionError
	http_type.ExecutionError.ErrorMessage = rpc_value.ErrorMessage
	return http_type
}

func ToHttpStarlarkInterpretationError(rpc_value *rpc_api.StarlarkInterpretationError) api_type.StarlarkInterpretationError {
	var http_type api_type.StarlarkInterpretationError
	http_type.InterpretationError.ErrorMessage = rpc_value.ErrorMessage
	return http_type
}

func ToHttpStarlarkValidationError(rpc_value *rpc_api.StarlarkValidationError) api_type.StarlarkValidationError {
	var http_type api_type.StarlarkValidationError
	http_type.ValidationError.ErrorMessage = rpc_value.ErrorMessage
	return http_type
}

func ToHttpStarlarkInfo(rpc_value *rpc_api.StarlarkInfo) api_type.StarlarkInfo {
	var info api_type.StarlarkInfo
	info.Info.Instruction.InfoMessage = rpc_value.InfoMessage
	return info
}

func ToHttpStarlarkInstruction(rpc_value *rpc_api.StarlarkInstruction) api_type.StarlarkInstruction {
	position := ToHttpStarlarkInstructionPosition(rpc_value.Position)
	return api_type.StarlarkInstruction{
		ExecutableInstruction: rpc_value.ExecutableInstruction,
		IsSkipped:             rpc_value.IsSkipped,
		Position:              &position,
		InstructionName:       rpc_value.InstructionName,
		Arguments: utils.MapList(
			rpc_value.Arguments,
			ToHttpStarlarkInstructionArgument,
		),
	}
}

func ToHttpStarlarkInstructionPosition(rpc_value *rpc_api.StarlarkInstructionPosition) api_type.StarlarkInstructionPosition {
	return api_type.StarlarkInstructionPosition{
		Column:   rpc_value.Column,
		Line:     rpc_value.Line,
		Filename: rpc_value.Filename,
	}
}

func ToHttpStarlarkInstructionResult(rpc_value *rpc_api.StarlarkInstructionResult) api_type.StarlarkInstructionResult {
	var instructionResult api_type.StarlarkInstructionResult
	instructionResult.InstructionResult.SerializedInstructionResult = rpc_value.SerializedInstructionResult
	return instructionResult
}

func ToHttpStarlarkProgressInfo(rpc_value *rpc_api.StarlarkRunProgress) api_type.StarlarkRunProgress {
	var progress api_type.StarlarkRunProgress
	progress.ProgressInfo.CurrentStepInfo = rpc_value.CurrentStepInfo
	progress.ProgressInfo.CurrentStepNumber = int32(rpc_value.CurrentStepNumber)
	progress.ProgressInfo.TotalSteps = int32(rpc_value.TotalSteps)
	return progress
}

func ToHttpStarlarkWarning(rpc_value *rpc_api.StarlarkWarning) api_type.StarlarkWarning {
	var warning api_type.StarlarkWarning
	warning.Warning.WarningMessage = rpc_value.WarningMessage
	return warning
}

func ToHttpStarlarkRunFinishedEvent(rpc_value *rpc_api.StarlarkRunFinishedEvent) api_type.StarlarkRunFinishedEvent {
	var event api_type.StarlarkRunFinishedEvent
	event.RunFinishedEvent.IsRunSuccessful = rpc_value.IsRunSuccessful
	event.RunFinishedEvent.SerializedOutput = rpc_value.SerializedOutput
	return event
}

func ToHttpStarlarkInstructionArgument(rpc_value *rpc_api.StarlarkInstructionArg) api_type.StarlarkInstructionArgument {
	return api_type.StarlarkInstructionArgument{
		ArgName:            rpc_value.ArgName,
		IsRepresentative:   rpc_value.IsRepresentative,
		SerializedArgValue: rpc_value.SerializedArgValue,
	}
}
