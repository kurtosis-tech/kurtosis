package output_printers

import (
	"errors"
	"fmt"
	"sync"

	"github.com/briandowns/spinner"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_args/run"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/interactive_terminal_decider"
	"github.com/kurtosis-tech/kurtosis/cli/cli/out"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

type DefaultExecutionPrinter struct {
	lock *sync.Mutex

	isStarted bool

	isSpinnerBeingUsed bool
	spinner            *spinner.Spinner
}

func NewDefaultExecutionPrinter() *DefaultExecutionPrinter {
	return &DefaultExecutionPrinter{
		lock:               &sync.Mutex{},
		isStarted:          false,
		isSpinnerBeingUsed: false,
		spinner:            nil,
	}
}

func (printer *DefaultExecutionPrinter) Start() error {
	if printer.isStarted {
		return stacktrace.NewError("printer already started")
	}
	printer.isStarted = true
	if !interactive_terminal_decider.IsInteractiveTerminal() {
		printer.isSpinnerBeingUsed = false
		logrus.Infof("Kurtosis CLI is running in a non interactive terminal. Everything will work but progress information and the progress bar will not be displayed.")
		return nil
	}

	// spinner setup
	printer.isSpinnerBeingUsed = true
	printer.spinner = spinner.New(spinnerChar, spinnerSpeed, spinnerColor, spinner.WithWriter(writer), spinner.WithSuffix(spinnerDefaultSuffix))
	printer.startSpinnerIfUsed()
	return nil
}

func (printer *DefaultExecutionPrinter) Stop() {
	printer.stopSpinnerIfUsed()
	printer.isStarted = false
}

// PrintKurtosisExecutionResponseLineToStdOut format and prints the instruction to StdOut.
func (printer *DefaultExecutionPrinter) PrintKurtosisExecutionResponseLineToStdOut(responseLine *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, verbosity run.Verbosity, dryRun bool) error {
	// Printing is a 3 phase operation:
	// 1. stop spinner to clear the ephemeral progress info
	// 2. print whatever needs to be printed, could be nothing
	// 3. restart the spinner, potentially with an updated content
	// To avoid conflicts, we take a lock out of cautiousness (this method shouldn't be called concurrently anyway)
	printer.lock.Lock()
	defer printer.lock.Unlock()

	if !printer.isStarted {
		return stacktrace.NewError("Cannot print with a non started printer")
	}

	// process response payload
	if responseLine.GetInstruction() != nil && verbosity != run.OutputOnly {
		formattedInstruction := formatInstruction(responseLine.GetInstruction(), verbosity)
		// we separate each tuple (instruction, result) with an additional newline
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

func (printer *DefaultExecutionPrinter) printPersistentLineToStdOut(lineToPrint string) error {
	// If spinner is being used, we have to stop spinner -> print -> start spinner in order to keep the spinner at the bottom of the output
	printer.stopSpinnerIfUsed()
	out.PrintOutLn(lineToPrint)
	printer.startSpinnerIfUsed()
	return nil
}

func (printer *DefaultExecutionPrinter) startSpinnerIfUsed() {
	if printer.isSpinnerBeingUsed {
		printer.spinner.Start()
	}
}

func (printer *DefaultExecutionPrinter) stopSpinnerIfUsed() {
	if printer.isSpinnerBeingUsed {
		printer.spinner.Stop()
	}
}
