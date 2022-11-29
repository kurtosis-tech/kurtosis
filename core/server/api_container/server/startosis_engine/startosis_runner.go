package startosis_engine

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/sirupsen/logrus"
)

type StartosisRunner struct {
	startosisInterpreter *StartosisInterpreter

	startosisValidator *StartosisValidator

	startosisExecutor *StartosisExecutor
}

const (
	defaultCurrentStepNumber  = 0
	defaultTotalStepsNumber   = 0
	startingInterpretationMsg = "Interpreting Starlark code - execution will begin shortly"
	startingValidationMsg     = "Pre-validating Starlark code and downloading docker images - execution will begin shortly"
	startingExecutionMsg      = "Starting execution"

	missingRunMethodErrorFromStarlark = "Evaluation error: module has no .run field or method\n\tat [3:12]: <toplevel>"
)

func NewStartosisRunner(interpreter *StartosisInterpreter, validator *StartosisValidator, executor *StartosisExecutor) *StartosisRunner {
	return &StartosisRunner{
		startosisInterpreter: interpreter,
		startosisValidator:   validator,
		startosisExecutor:    executor,
	}
}

func (runner *StartosisRunner) Run(ctx context.Context, dryRun bool, moduleId string, serializedStartosis string, serializedParams string) <-chan *kurtosis_core_rpc_api_bindings.KurtosisExecutionResponseLine {
	// TODO(gb): add metric tracking maybe?
	kurtosisExecutionResponseLines := make(chan *kurtosis_core_rpc_api_bindings.KurtosisExecutionResponseLine)

	go func() {
		defer close(kurtosisExecutionResponseLines)

		// Interpretation starts > send progress info (this line will be invisible as interpretation is super quick)
		progressInfo := binding_constructors.NewKurtosisExecutionResponseLineFromProgressInfo(
			startingInterpretationMsg, defaultCurrentStepNumber, defaultTotalStepsNumber)
		kurtosisExecutionResponseLines <- progressInfo

		instructionsList, interpretationError := runner.startosisInterpreter.Interpret(ctx, moduleId, serializedStartosis, serializedParams)
		if interpretationError != nil {
			interpretationError = maybeMakeMissingRunMethodErrorFriendlier(interpretationError, moduleId)
			kurtosisExecutionResponseLines <- binding_constructors.NewKurtosisExecutionResponseLineFromInterpretationError(interpretationError)
			return
		}
		totalNumberOfInstructions := uint32(len(instructionsList))
		logrus.Debugf("Successfully interpreted Startosis script into a series of Kurtosis instructions: \n%v",
			instructionsList)

		// Validation starts > send progress info
		progressInfo = binding_constructors.NewKurtosisExecutionResponseLineFromProgressInfo(
			startingValidationMsg, defaultCurrentStepNumber, totalNumberOfInstructions)
		kurtosisExecutionResponseLines <- progressInfo

		validationErrorsChan := runner.startosisValidator.Validate(ctx, instructionsList)
		if messagesWereReceived := forwardKurtosisResponseLineChannelUntilSourceIsClosed(validationErrorsChan, kurtosisExecutionResponseLines); messagesWereReceived {
			return
		}

		logrus.Debugf("Successfully validated Startosis script")

		// Execution starts > send progress info. This will soon be overridden byt the first instruction execution
		progressInfo = binding_constructors.NewKurtosisExecutionResponseLineFromProgressInfo(
			startingExecutionMsg, defaultCurrentStepNumber, totalNumberOfInstructions)
		kurtosisExecutionResponseLines <- progressInfo

		executionResponseLinesChan := runner.startosisExecutor.Execute(ctx, dryRun, instructionsList)
		forwardKurtosisResponseLineChannelUntilSourceIsClosed(executionResponseLinesChan, kurtosisExecutionResponseLines)
		logrus.Debugf("Successfully executed the list of %d Kurtosis instructions", len(instructionsList))
	}()
	return kurtosisExecutionResponseLines
}

func forwardKurtosisResponseLineChannelUntilSourceIsClosed(sourceChan <-chan *kurtosis_core_rpc_api_bindings.KurtosisExecutionResponseLine, destChan chan<- *kurtosis_core_rpc_api_bindings.KurtosisExecutionResponseLine) bool {
	messagesWereReceived := false
	for executionResponseLine := range sourceChan {
		logrus.Debugf("Received kurtosis execution line Kurtosis:\n%v", executionResponseLine)
		destChan <- executionResponseLine
		messagesWereReceived = true
	}
	logrus.Debug("Kurtosis instructions stream was closed. Exiting execution loop")
	return messagesWereReceived
}

func maybeMakeMissingRunMethodErrorFriendlier(originalError *kurtosis_core_rpc_api_bindings.KurtosisInterpretationError, moduleId string) *kurtosis_core_rpc_api_bindings.KurtosisInterpretationError {
	if originalError.GetErrorMessage() == missingRunMethodErrorFromStarlark {
		return binding_constructors.NewKurtosisInterpretationError(fmt.Sprintf("No 'run' function found in file '%v/main.star'; a 'run' entrypoint function is required in the main.star file of any Kurtosis package", moduleId))
	}
	return originalError
}
