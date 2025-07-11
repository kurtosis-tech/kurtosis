package output_printers

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bazelbuild/buildtools/build"
	"github.com/briandowns/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fatih/color"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_args/run"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/interactive_terminal_decider"
	"github.com/kurtosis-tech/kurtosis/cli/cli/out"
	"github.com/kurtosis-tech/stacktrace"
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

type ExecutionPrinter struct {
	lock *sync.Mutex

	isStarted bool

	// Bubbletea integration
	isInteractive     bool
	bubbletteaModel   *ExecutionModel
	bubbletteaProgram *tea.Program
	messageChan       chan tea.Msg

	// Legacy spinner for non-interactive terminals
	isSpinnerBeingUsed bool
	spinner            *spinner.Spinner
}

func NewExecutionPrinter() *ExecutionPrinter {
	return &ExecutionPrinter{
		lock:               &sync.Mutex{},
		isSpinnerBeingUsed: false,
		spinner:            nil,
		isStarted:          false,
		isInteractive:      interactive_terminal_decider.IsInteractiveTerminal(),
		messageChan:        make(chan tea.Msg, 100), // Buffered channel
	}
}

func (printer *ExecutionPrinter) Start() error {
	return printer.StartWithVerbosity(run.Brief, false)
}

func (printer *ExecutionPrinter) StartWithVerbosity(verbosity run.Verbosity, dryRun bool) error {
	if printer.isStarted {
		return stacktrace.NewError("printer already started")
	}
	printer.isStarted = true

	if printer.isInteractive {
		// Initialize bubbletea model and program
		printer.bubbletteaModel = NewExecutionModel(verbosity, dryRun, true)
		printer.bubbletteaProgram = tea.NewProgram(
			printer.bubbletteaModel,
			tea.WithMouseCellMotion(),
		)

		// Start the bubbletea program in a goroutine
		go func() {
			if _, err := printer.bubbletteaProgram.Run(); err != nil {
				logrus.Errorf("Error running bubbletea program: %v", err)
			}
		}()

		// Start message processing goroutine
		go printer.processMessages()
	} else {
		// Fallback to legacy spinner for non-interactive terminals
		printer.isSpinnerBeingUsed = false
		logrus.Infof("Kurtosis CLI is running in a non interactive terminal. Everything will work but progress information and the progress bar will not be displayed.")
	}
	return nil
}

func (printer *ExecutionPrinter) Stop() {
	if printer.isInteractive && printer.bubbletteaProgram != nil {
		// Send completion message and quit
		printer.messageChan <- ExecutionCompleteMsg{Success: true, Error: nil}
		printer.bubbletteaProgram.Quit()
		close(printer.messageChan)
	} else {
		printer.stopSpinnerIfUsed()
	}
	printer.isStarted = false
}

// PrintKurtosisExecutionResponseLineToStdOut format and prints the instruction to StdOut.
func (printer *ExecutionPrinter) PrintKurtosisExecutionResponseLineToStdOut(responseLine *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, verbosity run.Verbosity, dryRun bool) error {
	printer.lock.Lock()
	defer printer.lock.Unlock()

	if !printer.isStarted {
		return stacktrace.NewError("Cannot print with a non started printer")
	}

	if printer.isInteractive {
		// Convert response line to bubbletea message and send
		msg, err := printer.convertResponseLineToMessage(responseLine, verbosity, dryRun)
		if err != nil {
			return stacktrace.Propagate(err, "Error converting response line to bubbletea message")
		}
		if msg != nil {
			select {
			case printer.messageChan <- msg:
				// Message sent successfully
			default:
				// Channel full, log warning but don't block
				logrus.Warnf("Message channel full, dropping message")
			}
		}
	} else {
		// Fallback to legacy printing for non-interactive terminals
		return printer.printToStdOutLegacy(responseLine, verbosity, dryRun)
	}
	return nil
}

func (printer *ExecutionPrinter) printPersistentLineToStdOut(lineToPrint string) error {
	// If spinner is being used, we have to stop spinner -> print -> start spinner in order to keep the spinner at the bottom of the output
	printer.stopSpinnerIfUsed()
	out.PrintOutLn(lineToPrint)
	printer.startSpinnerIfUsed()
	return nil
}

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
	// if verbosity == run.Detailed {
	executionDuration := instructionResult.GetExecutionDuration()
	serializedInstructionResult = fmt.Sprintf("%s (execution duration: %s)", serializedInstructionResult, executionDuration.AsDuration().String())
	// }
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
		totalExecutionDuration := runFinishedEvent.GetTotalExecutionDuration().AsDuration()
		totalParallelExecutionDuration := runFinishedEvent.GetTotalParallelExecutionDuration().AsDuration()
		if totalParallelExecutionDuration > 0 {
			percentageFaster := (totalExecutionDuration.Seconds() - totalParallelExecutionDuration.Seconds()) / totalParallelExecutionDuration.Seconds() * 100
			durationMsg = durationMsg + fmt.Sprintf(" Total instruction execution time: %s. Total parallel execution time: %s. Parallel execution is %f%% faster", totalExecutionDuration.String(), totalParallelExecutionDuration.String(), percentageFaster)
		} else {
			durationMsg = durationMsg + fmt.Sprintf(" Total instruction execution time: %s. Total parallel execution time: %s.", totalExecutionDuration.String(), totalParallelExecutionDuration.String())
		}
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

func (printer *ExecutionPrinter) startSpinnerIfUsed() {
	if printer.isSpinnerBeingUsed {
		printer.spinner.Start()
	}
}

func (printer *ExecutionPrinter) stopSpinnerIfUsed() {
	if printer.isSpinnerBeingUsed {
		printer.spinner.Stop()
	}
}

// processMessages handles bubbletea messages from the channel
func (printer *ExecutionPrinter) processMessages() {
	for msg := range printer.messageChan {
		if printer.bubbletteaProgram != nil {
			printer.bubbletteaProgram.Send(msg)
		}
	}
}

// convertResponseLineToMessage converts a Starlark response line to a bubbletea message
func (printer *ExecutionPrinter) convertResponseLineToMessage(responseLine *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, verbosity run.Verbosity, dryRun bool) (tea.Msg, error) {
	if responseLine.GetInstruction() != nil && verbosity != run.OutputOnly {
		instruction := responseLine.GetInstruction()
		instructionId := instruction.GetInstructionId()
		return InstructionStartedMsg{
			ID:   instructionId,
			Name: formatInstruction(instruction, verbosity),
		}, nil
	} else if responseLine.GetInstructionResult() != nil {
		result := responseLine.GetInstructionResult()
		instructionId := result.GetInstructionId()
		return InstructionCompletedMsg{
			ID:     instructionId,
			Result: formatInstructionResult(result, verbosity),
		}, nil
	} else if responseLine.GetError() != nil {
		var errorMsg string
		var instructionId string

		if responseLine.GetError().GetInterpretationError() != nil {
			errorMsg = fmt.Sprintf("There was an error interpreting Starlark code \n%v", responseLine.GetError().GetInterpretationError().GetErrorMessage())
			instructionId = "interpretation"
		} else if responseLine.GetError().GetValidationError() != nil {
			errorMsg = fmt.Sprintf("There was an error validating Starlark code \n%v", responseLine.GetError().GetValidationError().GetErrorMessage())
			instructionId = "validation"
		} else if responseLine.GetError().GetExecutionError() != nil {
			errorMsgWithStackTrace := errors.New(responseLine.GetError().GetExecutionError().GetErrorMessage())
			cleanedErrorFromStarlark := out.GetErrorMessageToBeDisplayedOnCli(errorMsgWithStackTrace)
			errorMsg = fmt.Sprintf("There was an error executing Starlark code \n%v", cleanedErrorFromStarlark)
			instructionId = "execution"
		}

		return InstructionFailedMsg{
			ID:    instructionId,
			Error: FormatError(errorMsg),
		}, nil
	} else if responseLine.GetProgressInfo() != nil {
		progress := responseLine.GetProgressInfo()
		progressRatio := float64(progress.GetCurrentStepNumber()) / float64(progress.GetTotalSteps())
		instructionId := progress.GetInstructionId()
		return InstructionProgressMsg{
			ID:       instructionId,
			Progress: progressRatio,
			Message:  formatProgressMessage(progress.GetCurrentStepInfo()),
		}, nil
	} else if responseLine.GetRunFinishedEvent() != nil {
		runFinished := responseLine.GetRunFinishedEvent()
		return ExecutionCompleteMsg{
			ID:      "execution",
			Info:    formatInfo(runFinished.GetSerializedOutput()),
			Success: runFinished.GetIsRunSuccessful(),
			Error:   nil,
		}, nil
	} else if responseLine.GetWarning() != nil {
		warning := responseLine.GetWarning()
		instructionId := "execution"
		return InstructionWarningMsg{
			ID:      instructionId,
			Warning: formatWarning(warning.GetWarningMessage()),
		}, nil
	} else if responseLine.GetInfo() != nil {
		info := responseLine.GetInfo()
		instructionId := "execution"
		return InstructionInfoMsg{
			ID:   instructionId,
			Info: formatInfo(info.GetInfoMessage()),
		}, nil
	}

	// No message to send for unknown response types
	return nil, nil
}

// printToStdOutLegacy handles printing for non-interactive terminals using the original logic
func (printer *ExecutionPrinter) printToStdOutLegacy(responseLine *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, verbosity run.Verbosity, dryRun bool) error {
	// Original printing logic for non-interactive terminals
	if responseLine.GetInstruction() != nil && verbosity != run.OutputOnly {
		formattedInstruction := formatInstruction(responseLine.GetInstruction(), verbosity)
		formattedInstructionWithNewline := fmt.Sprintf("\n%s", formattedInstruction)
		if err := printer.printPersistentLineToStdOut(formattedInstructionWithNewline); err != nil {
			return stacktrace.Propagate(err, "Error printing Kurtosis instruction: \n%v", formattedInstruction)
		}
	} else if responseLine.GetInstructionResult() != nil {
		formattedInstructionResult := formatInstructionResult(responseLine.GetInstructionResult(), verbosity)
		if err := printer.printPersistentLineToStdOut(formattedInstructionResult); err != nil {
			return stacktrace.Propagate(err, "Error printing Kurtosis instruction result: \n%v", formattedInstructionResult)
		}
	} else if responseLine.GetError() != nil {
		var errorMsg string
		if responseLine.GetError().GetInterpretationError() != nil {
			errorMsg = fmt.Sprintf("There was an error interpreting Starlark code \n%v", responseLine.GetError().GetInterpretationError().GetErrorMessage())
		} else if responseLine.GetError().GetValidationError() != nil {
			errorMsg = fmt.Sprintf("There was an error validating Starlark code \n%v", responseLine.GetError().GetValidationError().GetErrorMessage())
		} else if responseLine.GetError().GetExecutionError() != nil {
			errorMsgWithStackTrace := errors.New(responseLine.GetError().GetExecutionError().GetErrorMessage())
			cleanedErrorFromStarlark := out.GetErrorMessageToBeDisplayedOnCli(errorMsgWithStackTrace)
			errorMsg = fmt.Sprintf("There was an error executing Starlark code \n%v", cleanedErrorFromStarlark)
		}
		formattedError := FormatError(errorMsg)
		if err := printer.printPersistentLineToStdOut(formattedError); err != nil {
			return stacktrace.Propagate(err, "An error happened executing Starlark code but the error couldn't be printed to the CLI output. Error message was: \n%v", errorMsg)
		}
	} else if responseLine.GetProgressInfo() != nil {
		if printer.isSpinnerBeingUsed {
			progress := responseLine.GetProgressInfo()
			progressMessageStr := formatProgressMessage(progress.GetCurrentStepInfo())
			progressBarStr := formatProgressBar(progress.GetCurrentStepNumber(), progress.GetTotalSteps(), progressBarChar)
			printer.spinner.Suffix = fmt.Sprintf("   %s %s", progressBarStr, progressMessageStr)
		}
	} else if responseLine.GetRunFinishedEvent() != nil {
		formattedRunOutputMessage := formatRunOutput(responseLine.GetRunFinishedEvent(), dryRun, verbosity)
		formattedRunOutputMessageWithNewline := fmt.Sprintf("\n%s", formattedRunOutputMessage)
		if err := printer.printPersistentLineToStdOut(formattedRunOutputMessageWithNewline); err != nil {
			return stacktrace.Propagate(err, "Unable to print the success output message containing the serialized output object. Message was: \n%v", formattedRunOutputMessage)
		}
	} else if responseLine.GetWarning() != nil {
		formattedRunWarningMessage := formatWarning(responseLine.GetWarning().GetWarningMessage())
		formattedRunWarningMessageWithNewline := fmt.Sprintf("\n%s", formattedRunWarningMessage)
		if err := printer.printPersistentLineToStdOut(formattedRunWarningMessageWithNewline); err != nil {
			return stacktrace.Propagate(err, "Error printing warning message: %v", formattedRunWarningMessage)
		}
	} else if responseLine.GetInfo() != nil {
		formattedRunInfoMessage := formatInfo(responseLine.GetInfo().GetInfoMessage())
		formattedRunInfoMessageWithNewline := fmt.Sprintf("\n%s", formattedRunInfoMessage)
		if err := printer.printPersistentLineToStdOut(formattedRunInfoMessageWithNewline); err != nil {
			return stacktrace.Propagate(err, "Error printing info message: %v", formattedRunInfoMessage)
		}
	}
	return nil
}
