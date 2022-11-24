package output_printers

import (
	"fmt"
	"github.com/bazelbuild/buildtools/build"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/sirupsen/logrus"
	"strings"
)

const (
	bazelBuildDefaultFilename = ""
)

// PrintKurtosisExecutionResponseLineToStdOut format and prints the instruction to StdOut. It returns a boolean indicating whether an error occurred during printing
// TODO(gb): scriptOutput buffer will be abandoned once we start printing the instruciton output right next to the instruction itself
func PrintKurtosisExecutionResponseLineToStdOut(responseLine *kurtosis_core_rpc_api_bindings.KurtosisExecutionResponseLine, scriptOutput *strings.Builder) bool {
	var printingError error
	if responseLine.GetInstruction() != nil {
		_, printingError = fmt.Fprintln(logrus.StandardLogger().Out, formatInstruction(responseLine.GetInstruction()))
		if responseLine.GetInstruction().InstructionResult != nil {
			scriptOutput.WriteString(responseLine.GetInstruction().GetInstructionResult())
		}
	} else if responseLine.GetError() != nil {
		if responseLine.GetError().GetInterpretationError() != nil {
			errorMsg := fmt.Sprintf("There was an error interpreting Kurtosis code \n%v", responseLine.GetError().GetInterpretationError().GetErrorMessage())
			_, printingError = fmt.Fprintln(logrus.StandardLogger().Out, errorMsg)
		} else if responseLine.GetError().GetValidationError() != nil {
			errorMsg := fmt.Sprintf("There was an error validating Kurtosis code \n%v", responseLine.GetError().GetValidationError().GetErrorMessage())
			_, printingError = fmt.Fprintln(logrus.StandardLogger().Out, errorMsg)
		} else if responseLine.GetError().GetExecutionError() != nil {
			errorMsg := fmt.Sprintf("There was an error executing Kurtosis code \n%v", responseLine.GetError().GetExecutionError().GetErrorMessage())
			_, printingError = fmt.Fprintln(logrus.StandardLogger().Out, errorMsg)
		}
	}
	if printingError != nil {
		logrus.Errorf("An error occured printing Kurtosis output to the terminal: \n%v", responseLine.GetInstruction())
		return true
	}
	return false
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
