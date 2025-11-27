package output_printers

import (
	"fmt"
	"strings"
	"time"

	"github.com/bazelbuild/buildtools/build"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/dzobbe/PoTE-kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_args/run"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/out"
	"github.com/sirupsen/logrus"
)

const (
	bazelBuildDefaultFilename = ""
	spinnerDefaultSuffix      = ""

	instructionPrefixString = "> "
	resultPrefixString      = ""

	progressBarLength = 20       // in characters
	progressBarChar   = "\u2588" // unicode for: █

	codeCommentPrefix = "# "

	newlineChar = "\n"
)

var (
	colorizeInstruction      = color.New(color.FgCyan).SprintFunc()
	colorizeResult           = color.New(color.FgWhite).SprintFunc()
	colorizeError            = color.New(color.FgRed).SprintFunc()
	colorizeWarning          = color.New(color.FgYellow).SprintFunc()
	colorizeRunSuccessfulMsg = color.New(color.FgGreen).SprintFunc()
	colorizeInfo             = color.New(color.FgCyan).SprintFunc()

	colorizeProgressBarIsDone    = color.New(color.FgGreen).SprintFunc()
	colorizeProgressBarRemaining = color.New(color.FgWhite).SprintFunc()
)

var (
	writer = out.GetOut()

	spinnerChar  = spinner.CharSets[11]
	spinnerSpeed = 250 * time.Millisecond
	spinnerColor = spinner.WithColor("yellow")
)

func FormatError(errorMessage string) string {
	return colorizeError(errorMessage)
}

func formatWarning(warningMessage string) string {
	return colorizeWarning(warningMessage)
}

func formatInfo(infoMessage string) string {
	return colorizeInfo(infoMessage)
}

// nolint:exhaustive
func formatInstruction(instruction *kurtosis_core_rpc_api_bindings.StarlarkInstruction, verbosity run.Verbosity) string {
	var serializedInstruction string
	switch verbosity {
	case run.Description:
		serializedInstruction = instruction.Description
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
	return colorizeInstruction(serializedInstruction)
}

func formatInstructionResult(instructionResult *kurtosis_core_rpc_api_bindings.StarlarkInstructionResult, verbosity run.Verbosity) string {
	serializedInstructionResult := fmt.Sprintf("%s%s", resultPrefixString, instructionResult.GetSerializedInstructionResult())
	if verbosity == run.Detailed {
		executionDuration := instructionResult.GetExecutionDuration()
		serializedInstructionResult = fmt.Sprintf("%s (execution duration: %s)", serializedInstructionResult, executionDuration.AsDuration().String())
	}
	return colorizeResult(serializedInstructionResult)
}

func formatInstructionToReadableString(instruction *kurtosis_core_rpc_api_bindings.StarlarkInstruction, exhaustive bool) string {
	serializedInstructionComponents := []string{instruction.GetInstructionName()}
	for _, arg := range instruction.GetArguments() {
		if exhaustive || arg.GetIsRepresentative() {
			var serializedArg string
			if arg.ArgName != nil {
				serializedArg = fmt.Sprintf("%s=%s", arg.GetArgName(), arg.GetSerializedArgValue())
			} else {
				serializedArg = arg.GetSerializedArgValue()
			}
			if serializedArg != "" {
				serializedInstructionComponents = append(serializedInstructionComponents, serializedArg)
			}
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

func formatInstructionToExecutable(instruction *kurtosis_core_rpc_api_bindings.StarlarkInstruction) string {
	serializedInstruction := fmt.Sprintf(
		"from %s[%d:%d]\n%s",
		instruction.GetPosition().GetFilename(),
		instruction.GetPosition().GetLine(),
		instruction.GetPosition().GetColumn(),
		instruction.GetExecutableInstruction(),
	)
	serializedInstructionWithComment := fmt.Sprintf("%s%s", codeCommentPrefix, serializedInstruction)

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

func formatProgressMessage(messageLines []string) string {
	return strings.Join(messageLines, newlineChar)
}

func formatProgressBar(currentStep uint32, totalSteps uint32, progressBarChar string) string {
	threshold := 0
	if totalSteps != 0 {
		threshold = int(currentStep * progressBarLength / totalSteps)
	}
	isDone := colorizeProgressBarIsDone(strings.Repeat(progressBarChar, threshold))
	remaining := colorizeProgressBarRemaining(strings.Repeat(progressBarChar, progressBarLength-threshold))
	return fmt.Sprintf("%s%s", isDone, remaining)
}

func formatRunOutput(runFinishedEvent *kurtosis_core_rpc_api_bindings.StarlarkRunFinishedEvent, dryRun bool, verbosity run.Verbosity) string {
	durationMsg := "."
	if verbosity == run.Detailed {
		durationMsg = durationMsg + fmt.Sprintf(" Total instruction execution time: %s.", runFinishedEvent.GetTotalExecutionDuration().AsDuration().String())
	}
	if !runFinishedEvent.GetIsRunSuccessful() {
		if dryRun {
			return colorizeError(fmt.Sprintf("Error encountered running Starlark code in dry-run mode%s", durationMsg))
		}
		return colorizeError(fmt.Sprintf("Error encountered running Starlark code%s", durationMsg))
	}
	// run was successful
	runSuccessMsg := strings.Builder{}
	runSuccessMsg.WriteString("Starlark code successfully run")
	if dryRun {
		runSuccessMsg.WriteString(" in dry-run mode")
	}
	runSuccessMsg.WriteString(durationMsg)
	if runFinishedEvent.GetSerializedOutput() != "" {
		runSuccessMsg.WriteString(fmt.Sprintf(" Output was:\n%v", runFinishedEvent.GetSerializedOutput()))
	} else {
		runSuccessMsg.WriteString(" No output was returned.")
	}
	return colorizeRunSuccessfulMsg(runSuccessMsg.String())
}
