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
