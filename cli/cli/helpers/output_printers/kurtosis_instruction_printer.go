package output_printers

import (
	"fmt"
	"github.com/bazelbuild/buildtools/build"
	"github.com/fatih/color"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_args/run"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"strings"
)

const (
	bazelBuildDefaultFilename = ""

	instructionPrefixString = "> "
	resultPrefixString      = ""
)

var (
	colorizeInstruction = color.New(color.FgCyan).SprintfFunc()
	colorizeResult      = color.New(color.FgWhite).SprintfFunc()
	colorizeError       = color.New(color.FgRed).SprintfFunc()
)

// PrintKurtosisExecutionResponseLineToStdOut format and prints the instruction to StdOut. It returns a boolean indicating whether an error occurred during printing
func PrintKurtosisExecutionResponseLineToStdOut(responseLine *kurtosis_core_rpc_api_bindings.KurtosisExecutionResponseLine, verbosity run.Verbosity) (bool, error) {
	var errorRunningKurtosisCode bool
	if responseLine.GetInstruction() != nil {
		formattedInstruction := formatInstruction(responseLine.GetInstruction(), verbosity)
		// we separate each tuple (instruction, result) with an additional newline
		formattedInstructionWithNewline := fmt.Sprintf("%s\n", formattedInstruction)
		if _, err := fmt.Fprintln(logrus.StandardLogger().Out, formattedInstructionWithNewline); err != nil {
			return errorRunningKurtosisCode, stacktrace.Propagate(err, "Error printing Kurtosis instruction: \n%v", formattedInstruction)
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
		formattedError := formatError(errorMsg)
		if _, err := fmt.Fprintln(logrus.StandardLogger().Out, formattedError); err != nil {
			return errorRunningKurtosisCode, stacktrace.Propagate(err, "An error happened executing Starlark code but the error couldn't be printed to the CLI output. Error message was: \n%v", errorMsg)
		}
	}
	return errorRunningKurtosisCode, nil
}

func formatError(errorMessage string) string {
	return colorizeError(errorMessage)
}

func formatInstruction(instruction *kurtosis_core_rpc_api_bindings.KurtosisInstruction, verbosity run.Verbosity) string {
	var serializedInstruction string
	switch verbosity {
	case run.Brief:
		serializedInstruction = formatInstructionToReadableString(instruction, false)
	case run.Detailed:
		serializedInstruction = formatInstructionToReadableString(instruction, true)
	case run.Executable:
		serializedInstruction = formatInstructionToExecutable(instruction)
	default:
		logrus.Warnf("Unsupported verbosity flag: '%s'. Supported values are: (%s). The instruction will be printed with the default verbosity '%s'",
			verbosity.String(), strings.Join(run.VerbosityStrings(), ", "), run.Brief.String())
		serializedInstruction = formatInstructionToReadableString(instruction, false)
	}

	if instruction.InstructionResult != nil {
		serializedResult := formatInstructionResult(instruction.GetInstructionResult())
		return fmt.Sprintf("%s\n%s", colorizeInstruction(serializedInstruction), colorizeResult(serializedResult))
	}
	return colorizeInstruction(serializedInstruction)
}

func formatInstructionToReadableString(instruction *kurtosis_core_rpc_api_bindings.KurtosisInstruction, exhaustive bool) string {
	serializedInstructionComponents := []string{instruction.GetInstructionName()}
	for _, arg := range instruction.GetArguments() {
		if exhaustive || arg.GetIsRepresentative() {
			var serializedArg string
			if arg.ArgName != nil {
				serializedArg = fmt.Sprintf("%s=%s", arg.GetArgName(), arg.GetSerializedArgValue())
			} else {
				serializedArg = arg.GetSerializedArgValue()
			}
			serializedInstructionComponents = append(serializedInstructionComponents, serializedArg)
		}
	}

	var serializedInstruction string
	if exhaustive {
		separator := fmt.Sprintf("\n%s\t", instructionPrefixString)
		serializedInstruction = strings.Join(serializedInstructionComponents, separator)
	} else {
		separator := " "
		serializedInstruction = strings.Join(serializedInstructionComponents, separator)
	}
	return fmt.Sprintf("%s%s", instructionPrefixString, serializedInstruction)
}

func formatInstructionToExecutable(instruction *kurtosis_core_rpc_api_bindings.KurtosisInstruction) string {
	serializedInstructionWithComment := fmt.Sprintf(
		"# from %s[%d:%d]\n%s",
		instruction.GetPosition().GetFilename(),
		instruction.GetPosition().GetLine(),
		instruction.GetPosition().GetColumn(),
		instruction.GetExecutableInstruction(),
	)

	parsedInstruction, err := build.ParseDefault(bazelBuildDefaultFilename, []byte(serializedInstructionWithComment))
	if err != nil {
		logrus.Warnf("Unable to format instruction. Will print it with no indentation. Problematic instruction was: \n%v", serializedInstructionWithComment)
		return serializedInstructionWithComment
	}

	multiLineInstruction := strings.Builder{}
	for _, statement := range parsedInstruction.Stmt {
		multiLineInstruction.WriteString(build.FormatString(statement))
	}
	return multiLineInstruction.String()
}

func formatInstructionResult(instructionResult string) string {
	return fmt.Sprintf("%s%s", resultPrefixString, instructionResult)
}
