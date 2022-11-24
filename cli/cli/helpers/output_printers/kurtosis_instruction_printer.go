package output_printers

import (
	"fmt"
	"github.com/bazelbuild/buildtools/build"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"strings"
)

const (
	bazelBuildDefaultFilename = ""
)

// PrintKurtosisExecutionResponseLineToStdOut format and prints the instruction to StdOut. It returns a boolean indicating whether an error occurred during printing
// TODO(gb): scriptOutput buffer will be abandoned once we start printing the instruciton output right next to the instruction itself
func PrintKurtosisExecutionResponseLineToStdOut(responseLine *kurtosis_core_rpc_api_bindings.KurtosisExecutionResponseLine, scriptOutput *strings.Builder) (bool, error) {
	var errorRunningKurtosisCode bool
	if responseLine.GetInstruction() != nil {
		formattedInstruction := formatInstruction(responseLine.GetInstruction())
		if _, err := fmt.Fprintln(logrus.StandardLogger().Out, formattedInstruction); err != nil {
			return errorRunningKurtosisCode, stacktrace.Propagate(err, "Error printing Kurtosis instruction: \n%v", formattedInstruction)
		}
		if responseLine.GetInstruction().InstructionResult != nil {
			scriptOutput.WriteString(responseLine.GetInstruction().GetInstructionResult())
		}
	} else if responseLine.GetError() != nil {
		errorRunningKurtosisCode = true
		var errorMsg string
		if responseLine.GetError().GetInterpretationError() != nil {
			errorMsg = fmt.Sprintf("There was an error interpreting Starlark code \n%v", responseLine.GetError().GetInterpretationError().GetErrorMessage())
		} else if responseLine.GetError().GetValidationError() != nil {
			errorMsg = fmt.Sprintf("There was an error validating Starlark code \n%v", responseLine.GetError().GetValidationError().GetErrorMessage())
		} else if responseLine.GetError().GetExecutionError() != nil {
			errorMsg = fmt.Sprintf("There was an error executing Starlark code \n%v", responseLine.GetError().GetExecutionError().GetErrorMessage())
		}
		if _, err := fmt.Fprintln(logrus.StandardLogger().Out, errorMsg); err != nil {
			return errorRunningKurtosisCode, stacktrace.Propagate(err, "An error happened executing Starlark code but the error couldn't be printed to the CLI output. Error message was: \n%v", errorMsg)
		}
	}
	return errorRunningKurtosisCode, nil
}

func formatInstruction(instruction *kurtosis_core_rpc_api_bindings.KurtosisInstruction) string {
	canonicalizedInstructionWithComment := fmt.Sprintf(
		"# from %s[%d:%d]\n%s",
		instruction.GetPosition().GetFilename(),
		instruction.GetPosition().GetLine(),
		instruction.GetPosition().GetColumn(),
		instruction.GetExecutableInstruction(),
	)

	parsedInstruction, err := build.ParseDefault(bazelBuildDefaultFilename, []byte(canonicalizedInstructionWithComment))
	if err != nil {
		logrus.Warnf("Unable to format instruction. Will print it with no indentation. Problematic instruction was: \n%v", canonicalizedInstructionWithComment)
		return canonicalizedInstructionWithComment
	}

	multiLineInstruction := strings.Builder{}
	for _, statement := range parsedInstruction.Stmt {
		multiLineInstruction.WriteString(build.FormatString(statement))
	}
	return multiLineInstruction.String()
}
