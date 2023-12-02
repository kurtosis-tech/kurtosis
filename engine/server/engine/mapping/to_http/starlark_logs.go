package to_http

import (
	"reflect"

	"github.com/kurtosis-tech/kurtosis/engine/server/engine/utils"
	"github.com/sirupsen/logrus"

	rpc_api "github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	api_type "github.com/kurtosis-tech/kurtosis/api/golang/http_rest/api_types"
)

func ToHttpStarlarkRunResponseLine(rpc_value rpc_api.StarlarkRunResponseLine) *api_type.StarlarkRunResponseLine {
	var http_type api_type.StarlarkRunResponseLine
	if runError := rpc_value.GetError(); runError != nil {
		if value := ToHttpStarlarkError(*runError); value != nil {
			http_type.FromStarlarkError(*value)
			return &http_type
		}
		return nil
	}

	if runInfo := rpc_value.GetInfo(); runInfo != nil {
		http_type.FromStarlarkInfo(ToHttpStarlarkInfo(*runInfo))
		return &http_type
	}

	if runInstruction := rpc_value.GetInstruction(); runInstruction != nil {
		http_type.FromStarlarkInstruction(ToHttpStarlarkInstruction(*runInstruction))
		return &http_type
	}

	if runInstructionResult := rpc_value.GetInstructionResult(); runInstructionResult != nil {
		http_type.FromStarlarkInstructionResult(ToHttpStarlarkInstructionResult(*runInstructionResult))
		return &http_type
	}

	if runProgressInfo := rpc_value.GetProgressInfo(); runProgressInfo != nil {
		http_type.FromStarlarkRunProgress(ToHttpStarlarkProgressInfo(*runProgressInfo))
		return &http_type
	}

	if runWarning := rpc_value.GetWarning(); runWarning != nil {
		http_type.FromStarlarkWarning(ToHttpStarlarkWarning(*runWarning))
		return &http_type
	}

	if runFinishedEvent := rpc_value.GetRunFinishedEvent(); runFinishedEvent != nil {
		http_type.FromStarlarkRunFinishedEvent(ToHttpStarlarkRunFinishedEvent(*runFinishedEvent))
		return &http_type
	}

	logrus.WithFields(logrus.Fields{
		"type":  reflect.TypeOf(rpc_value).Name(),
		"value": rpc_value.String(),
	}).Warnf("Unmatched gRPC to Http mapping, returning empty value")
	return nil
}

func ToHttpStarlarkError(rpc_value rpc_api.StarlarkError) *api_type.StarlarkError {
	var http_type api_type.StarlarkError
	if runError := rpc_value.GetExecutionError(); runError != nil {
		http_type.Error.FromStarlarkExecutionError(ToHttpStarlarkExecutionError(*runError))
		return &http_type
	}

	if runError := rpc_value.GetInterpretationError(); runError != nil {
		http_type.Error.FromStarlarkInterpretationError(ToHttpStarlarkInterpretationError(*runError))
		return &http_type
	}

	if runError := rpc_value.GetValidationError(); runError != nil {
		http_type.Error.FromStarlarkValidationError(ToHttpStarlarkValidationError(*runError))
		return &http_type
	}

	logrus.WithFields(logrus.Fields{
		"type":  reflect.TypeOf(rpc_value).Name(),
		"value": rpc_value.String(),
	}).Warnf("Unmatched gRPC to Http mapping, returning empty value")
	return nil
}

func ToHttpStarlarkExecutionError(rpc_value rpc_api.StarlarkExecutionError) api_type.StarlarkExecutionError {
	var http_type api_type.StarlarkExecutionError
	http_type.ExecutionError.ErrorMessage = rpc_value.ErrorMessage
	return http_type
}

func ToHttpStarlarkInterpretationError(rpc_value rpc_api.StarlarkInterpretationError) api_type.StarlarkInterpretationError {
	var http_type api_type.StarlarkInterpretationError
	http_type.InterpretationError.ErrorMessage = rpc_value.ErrorMessage
	return http_type
}

func ToHttpStarlarkValidationError(rpc_value rpc_api.StarlarkValidationError) api_type.StarlarkValidationError {
	var http_type api_type.StarlarkValidationError
	http_type.ValidationError.ErrorMessage = rpc_value.ErrorMessage
	return http_type
}

func ToHttpStarlarkInfo(rpc_value rpc_api.StarlarkInfo) api_type.StarlarkInfo {
	var info api_type.StarlarkInfo
	info.Info.Instruction.InfoMessage = ""
	return info
}

func ToHttpStarlarkInstruction(rpc_value rpc_api.StarlarkInstruction) api_type.StarlarkInstruction {
	return api_type.StarlarkInstruction{
		Arguments: utils.MapList(
			utils.FilterListNils(rpc_value.Arguments),
			ToHttpStarlarkInstructionArgument,
		),
	}
}

func ToHttpStarlarkInstructionResult(rpc_value rpc_api.StarlarkInstructionResult) api_type.StarlarkInstructionResult {
	var instructionResult api_type.StarlarkInstructionResult
	instructionResult.InstructionResult.SerializedInstructionResult = rpc_value.SerializedInstructionResult
	return instructionResult
}

func ToHttpStarlarkProgressInfo(rpc_value rpc_api.StarlarkRunProgress) api_type.StarlarkRunProgress {
	var progress api_type.StarlarkRunProgress
	progress.ProgressInfo.CurrentStepInfo = rpc_value.CurrentStepInfo
	progress.ProgressInfo.CurrentStepNumber = int32(rpc_value.CurrentStepNumber)
	progress.ProgressInfo.TotalSteps = int32(rpc_value.TotalSteps)
	return progress
}

func ToHttpStarlarkWarning(rpc_value rpc_api.StarlarkWarning) api_type.StarlarkWarning {
	var warning api_type.StarlarkWarning
	warning.Warning.WarningMessage = rpc_value.WarningMessage
	return warning
}

func ToHttpStarlarkRunFinishedEvent(rpc_value rpc_api.StarlarkRunFinishedEvent) api_type.StarlarkRunFinishedEvent {
	var event api_type.StarlarkRunFinishedEvent
	event.RunFinishedEvent.IsRunSuccessful = rpc_value.IsRunSuccessful
	event.RunFinishedEvent.SerializedOutput = rpc_value.SerializedOutput
	return event
}

func ToHttpStarlarkInstructionArgument(rpc_value rpc_api.StarlarkInstructionArg) api_type.StarlarkInstructionArgument {
	return api_type.StarlarkInstructionArgument{
		ArgName:            rpc_value.ArgName,
		IsRepresentative:   rpc_value.IsRepresentative,
		SerializedArgValue: rpc_value.SerializedArgValue,
	}
}
