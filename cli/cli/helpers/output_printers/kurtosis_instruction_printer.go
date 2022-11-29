package output_printers

import (
	"fmt"
	"github.com/bazelbuild/buildtools/build"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_args/run"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"strings"
	"sync"
	"time"
)

const (
	bazelBuildDefaultFilename = ""

	instructionPrefixString = "> "
	resultPrefixString      = ""

	progressBarLength = 20       // in characters
	progressBarChar   = "\u2588" // unicode for: █

	codeCommentPrefix = "# "
)

var (
	colorizeInstruction = color.New(color.FgCyan).SprintfFunc()
	colorizeResult      = color.New(color.FgWhite).SprintfFunc()
	colorizeError       = color.New(color.FgRed).SprintfFunc()

	colorizeProgressBarIsDone    = color.New(color.FgGreen).SprintfFunc()
	colorizeProgressBarRemaining = color.New(color.FgWhite).SprintfFunc()
)

var (
	writer = logrus.StandardLogger().Out

	spinnerChar  = spinner.CharSets[11]
	spinnerSpeed = 100 * time.Millisecond
	spinnerColor = spinner.WithColor("yellow")
)

type ExecutionPrinter struct {
	lock      *sync.Mutex
	spinner   *spinner.Spinner
	isStarted bool
}

func NewExecutionPrinter() *ExecutionPrinter {
	return &ExecutionPrinter{
		lock:      &sync.Mutex{},
		spinner:   nil,
		isStarted: false,
	}
}

func (printer *ExecutionPrinter) Start() error {
	if printer.isStarted {
		return stacktrace.NewError("printer already started")
	}
	printer.isStarted = true
	printer.spinner = spinner.New(spinnerChar, spinnerSpeed, spinnerColor, spinner.WithWriter(writer))
	printer.spinner.Start()
	return nil
}

func (printer *ExecutionPrinter) Stop() {
	if printer.isStarted {
		if printer.spinner != nil && printer.spinner.Active() {
			printer.spinner.Stop()
		}
	}
	printer.isStarted = false
}

// PrintKurtosisExecutionResponseLineToStdOut format and prints the instruction to StdOut. It returns a boolean indicating whether an error occurred during printing
func (printer *ExecutionPrinter) PrintKurtosisExecutionResponseLineToStdOut(responseLine *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, verbosity run.Verbosity) (bool, error) {
	// Printing is a 3 phase operation:
	// 1. stop spinner to clear the ephemeral progress info
	// 2. print whatever needs to be printed, could be nothing
	// 3. restart the spinner, potentially with an updated content
	// To avoid conflicts, we take a lock out of cautiousness (this method shouldn't be called concurrently anyway)
	printer.lock.Lock()
	defer printer.lock.Unlock()

	if !printer.isStarted {
		return false, stacktrace.NewError("Cannot print with a non started printer")
	}
	// Need to stop the current spinner otherwise it will conflict with whatever we're about to print
	printer.spinner.Stop()

	// process response payload
	isThisReponseLineAnError := false
	if responseLine.GetInstruction() != nil {
		formattedInstruction := formatInstruction(responseLine.GetInstruction(), verbosity)
		// we separate each tuple (instruction, result) with an additional newline
		formattedInstructionWithNewline := fmt.Sprintf("\n%s", formattedInstruction)
		if _, err := fmt.Fprintln(writer, formattedInstructionWithNewline); err != nil {
			return isThisReponseLineAnError, stacktrace.Propagate(err, "Error printing Kurtosis instruction: \n%v", formattedInstruction)
		}
	} else if responseLine.GetInstructionResult() != nil {
		formattedInstructionResult := formatInstructionResult(responseLine.GetInstructionResult())
		if _, err := fmt.Fprintln(logrus.StandardLogger().Out, formattedInstructionResult); err != nil {
			return isThisReponseLineAnError, stacktrace.Propagate(err, "Error printing Kurtosis instruction result: \n%v", formattedInstructionResult)
		}
	} else if responseLine.GetError() != nil {
		isThisReponseLineAnError = true
		var errorMsg string
		if responseLine.GetError().GetInterpretationError() != nil {
			errorMsg = fmt.Sprintf("There was an error interpreting Starlark code \n%v", responseLine.GetError().GetInterpretationError().GetErrorMessage())
		} else if responseLine.GetError().GetValidationError() != nil {
			errorMsg = fmt.Sprintf("There was an error validating Starlark code \n%v", responseLine.GetError().GetValidationError().GetErrorMessage())
		} else if responseLine.GetError().GetExecutionError() != nil {
			errorMsg = fmt.Sprintf("There was an error executing Starlark code \n%v", responseLine.GetError().GetExecutionError().GetErrorMessage())
		}
		formattedError := formatError(errorMsg)
		if _, err := fmt.Fprintln(writer, formattedError); err != nil {
			return isThisReponseLineAnError, stacktrace.Propagate(err, "An error happened executing Starlark code but the error couldn't be printed to the CLI output. Error message was: \n%v", errorMsg)
		}
	} else if responseLine.GetProgressInfo() != nil {
		progress := responseLine.GetProgressInfo()
		progressBarStr := formatProgressBar(progress.GetCurrentStepNumber(), progress.GetTotalSteps())
		spinnerInfoString := fmt.Sprintf("   %s %s", progressBarStr, progress.GetCurrentStepInfo())
		printer.spinner = spinner.New(spinnerChar, spinnerSpeed, spinnerColor, spinner.WithWriter(writer), spinner.WithSuffix(spinnerInfoString))
	}

	// re-start the spinner before exiting
	printer.spinner.Start()
	return isThisReponseLineAnError, nil
}

func formatError(errorMessage string) string {
	return colorizeError(errorMessage)
}

func formatInstruction(instruction *kurtosis_core_rpc_api_bindings.StarlarkInstruction, verbosity run.Verbosity) string {
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
	return colorizeInstruction(serializedInstruction)
}

func formatInstructionResult(instructionResult *kurtosis_core_rpc_api_bindings.StarlarkInstructionResult) string {
	serializedInstructionResult := fmt.Sprintf("%s%s", resultPrefixString, instructionResult.GetSerializedInstructionResult())
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

func formatProgressBar(currentStep uint32, totalSteps uint32) string {
	progressBar := strings.Builder{}
	threshold := currentStep * progressBarLength
	for i := uint32(0); i < progressBarLength; i++ {
		if i*totalSteps < threshold {
			progressBar.WriteString(colorizeProgressBarIsDone(progressBarChar))
		} else {
			progressBar.WriteString(colorizeProgressBarRemaining(progressBarChar))
		}
	}
	return progressBar.String()
}
