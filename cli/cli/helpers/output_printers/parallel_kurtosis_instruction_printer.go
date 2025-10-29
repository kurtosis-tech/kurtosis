package output_printers

import (
	"errors"
	"fmt"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_args/run"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/interactive_terminal_decider"
	"github.com/kurtosis-tech/kurtosis/cli/cli/out"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	sleepTimeBeforeFinalOutput = 100 * time.Millisecond
	messageChanBufferSize      = 100
)

type ParallelExecutionPrinter struct {
	lock *sync.Mutex

	isStarted bool

	// bubbletea library components for rendering the TUI
	bubbleteaModel   *ExecutionModel
	bubbleteaProgram *tea.Program
	bubbleteaMsgChan chan tea.Msg

	errorMsgOutput string        // gets populated if error occurs during execution
	finalOutput    string        // gets populated after execution is complete
	programDone    chan struct{} // used to wait for the bubbletea program to finish and signal when to print the final output
}

func NewParallelExecutionPrinter() *ParallelExecutionPrinter {
	return &ParallelExecutionPrinter{
		lock:             &sync.Mutex{},
		isStarted:        false,
		bubbleteaModel:   nil,
		bubbleteaProgram: nil,
		bubbleteaMsgChan: make(chan tea.Msg, messageChanBufferSize),
		errorMsgOutput:   "",
		finalOutput:      "",
		programDone:      make(chan struct{}),
	}
}

func (printer *ParallelExecutionPrinter) Start() error {
	if printer.isStarted {
		return stacktrace.NewError("printer already started")
	}
	printer.isStarted = true
	if !interactive_terminal_decider.IsInteractiveTerminal() {
		logrus.Infof("Kurtosis CLI is running in a non interactive terminal. Everything will work but progress information and the progress bar will not be displayed.")
		return nil
	}

	printer.bubbleteaModel = NewExecutionModel()
	printer.bubbleteaProgram = tea.NewProgram(
		printer.bubbleteaModel,
		tea.WithMouseCellMotion(),
		tea.WithAltScreen(),
	)

	go func() {
		if _, err := printer.bubbleteaProgram.Run(); err != nil {
			logrus.Errorf("Error running bubbletea program: %v", err)
		}
		close(printer.programDone)
	}()

	// Send messages to the bubbletea program
	go func() {
		for msg := range printer.bubbleteaMsgChan {
			if printer.bubbleteaProgram != nil {
				printer.bubbleteaProgram.Send(msg)
			}
		}
	}()

	return nil
}

func (printer *ParallelExecutionPrinter) Stop() {
	if !printer.isStarted {
		return
	}
	if !interactive_terminal_decider.IsInteractiveTerminal() {
		printer.isStarted = false
		return
	}

	// Give a moment for final messages to be processed and rendered
	time.Sleep(sleepTimeBeforeFinalOutput)

	close(printer.bubbleteaMsgChan)

	printer.bubbleteaProgram.Quit()

	<-printer.programDone

	if printer.errorMsgOutput != "" {
		out.PrintOutLn(printer.errorMsgOutput)
	}

	if printer.finalOutput != "" {
		out.PrintOutLn(printer.finalOutput)
	}

	printer.isStarted = false
}

// PrintKurtosisExecutionResponseLineToStdOut formats the instruction output and sends it to the bubbletea model to render the instruction output to the terminal UI
func (printer *ParallelExecutionPrinter) PrintKurtosisExecutionResponseLineToStdOut(responseLine *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, verbosity run.Verbosity, dryRun bool) error {
	printer.lock.Lock()
	defer printer.lock.Unlock()

	if !printer.isStarted {
		return stacktrace.NewError("Cannot print with a non started printer")
	}

	var msg tea.Msg
	if responseLine.GetInstruction() != nil && verbosity != run.OutputOnly {
		instruction := responseLine.GetInstruction()
		instructionId := instruction.GetInstructionId()
		msg = InstructionStartedMsg{
			ID:   instructionId,
			Name: formatInstruction(instruction, verbosity),
		}
	} else if responseLine.GetInstructionResult() != nil {
		result := responseLine.GetInstructionResult()
		instructionId := result.GetInstructionId()
		msg = InstructionCompletedMsg{
			ID:     instructionId,
			Result: formatInstructionResult(result, verbosity),
		}
	} else if responseLine.GetError() != nil {
		var errorMsg string
		var instructionId string

		if responseLine.GetError().GetInterpretationError() != nil {
			errorMsg = fmt.Sprintf("There was an error interpreting Starlark code \n%v", responseLine.GetError().GetInterpretationError().GetErrorMessage())
			instructionId = interpretationInstructionId
		} else if responseLine.GetError().GetValidationError() != nil {
			errorMsg = fmt.Sprintf("There was an error validating Starlark code \n%v", responseLine.GetError().GetValidationError().GetErrorMessage())
			instructionId = validationInstructionId
		} else if responseLine.GetError().GetExecutionError() != nil {
			errorMsgWithStackTrace := errors.New(responseLine.GetError().GetExecutionError().GetErrorMessage())
			cleanedErrorFromStarlark := out.GetErrorMessageToBeDisplayedOnCli(errorMsgWithStackTrace)
			errorMsg = fmt.Sprintf("There was an error executing Starlark code \n%v", cleanedErrorFromStarlark)
			instructionId = executionInstructionId
		}

		formattedError := FormatError(errorMsg)
		printer.errorMsgOutput = formattedError
		msg = InstructionFailedMsg{
			ID:    instructionId,
			Error: formattedError,
		}
	} else if responseLine.GetProgressInfo() != nil {
		progress := responseLine.GetProgressInfo()
		progressRatio := float64(progress.GetCurrentStepNumber()) / float64(progress.GetTotalSteps())
		instructionId := progress.GetInstructionId()
		msg = InstructionProgressMsg{
			ID:       instructionId,
			Progress: progressRatio,
			Message:  formatProgressMessage(progress.GetCurrentStepInfo()),
		}
	} else if responseLine.GetRunFinishedEvent() != nil {
		runFinished := responseLine.GetRunFinishedEvent()
		output := formatRunOutput(runFinished, dryRun, verbosity)
		printer.finalOutput = output
		msg = ExecutionCompleteMsg{
			ID:      executionInstructionId,
			Result:  output,
			Success: runFinished.GetIsRunSuccessful(),
			Error:   nil,
		}
	} else if responseLine.GetWarning() != nil {
		warning := responseLine.GetWarning()
		instructionId := executionInstructionId
		msg = InstructionWarningMsg{
			ID:      instructionId,
			Warning: formatWarning(warning.GetWarningMessage()),
		}
	} else if responseLine.GetInfo() != nil {
		info := responseLine.GetInfo()
		instructionId := executionInstructionId
		msg = InstructionInfoMsg{
			ID:   instructionId,
			Info: formatInfo(info.GetInfoMessage()),
		}
	}
	if msg != nil {
		select {
		case printer.bubbleteaMsgChan <- msg:
		default:
			logrus.Warnf("Message channel full, dropping message")
		}
	}

	return nil
}
