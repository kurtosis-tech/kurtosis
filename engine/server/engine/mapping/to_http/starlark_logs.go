package to_http

import (
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/utils"

	rpc_api "github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	api_type "github.com/kurtosis-tech/kurtosis/api/golang/http_rest/api_types"
)

func ToHttpApiStarlarkRunResponseLine(line rpc_api.StarlarkRunResponseLine) api_type.StarlarkRunResponseLine {
	if runError := line.GetError(); runError != nil {
		var http_type api_type.StarlarkRunResponseLine
		http_type.FromStarlarkError(ToHttpStarlarkError(*runError))
		return http_type
	}

	if runInfo := line.GetInfo(); runInfo != nil {
		var http_type api_type.StarlarkRunResponseLine
		http_type.FromStarlarkInfo(ToHttpStarlarkInfo(*runInfo))
		return http_type
	}

	if runInstruction := line.GetInstruction(); runInstruction != nil {
		var http_type api_type.StarlarkRunResponseLine
		http_type.FromStarlarkInstruction(ToHttpStarlarkInstruction(*runInstruction))
		return http_type
	}

	if runInstructionResult := line.GetInstructionResult(); runInstructionResult != nil {
		var http_type api_type.StarlarkRunResponseLine
		http_type.FromStarlarkInstructionResult(ToHttpStarlarkInstructionResult(*runInstructionResult))
		return http_type
	}

	if runProgressInfo := line.GetProgressInfo(); runProgressInfo != nil {
		var http_type api_type.StarlarkRunResponseLine
		http_type.FromStarlarkRunProgress(ToHttpStarlarkProgressInfo(*runProgressInfo))
		return http_type
	}

	if runWarning := line.GetWarning(); runWarning != nil {
		var http_type api_type.StarlarkRunResponseLine
		http_type.FromStarlarkWarning(ToHttpStarlarkWarning(*runWarning))
		return http_type
	}

	if runFinishedEvent := line.GetRunFinishedEvent(); runFinishedEvent != nil {
		var http_type api_type.StarlarkRunResponseLine
		http_type.FromStarlarkRunFinishedEvent(ToHttpStarlarkRunFinishedEvent(*runFinishedEvent))
		return http_type
	}

	return api_type.StarlarkRunResponseLine{}
}

func ToHttpStarlarkError(rpc_value rpc_api.StarlarkError) api_type.StarlarkError {
	return api_type.StarlarkError{
		// Error: rpc_value.Error,
	}

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

func ToHttpStarlarkRunResponseLine(rpc_value rpc_api.StarlarkRunResponseLine) api_type.StarlarkRunResponseLine {
	return api_type.StarlarkRunResponseLine{}

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
