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

func (runner *StartosisRunner) Run(ctx context.Context, dryRun bool, moduleId string, serializedStartosis string, serializedParams string) <-chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine {
	// TODO(gb): add metric tracking maybe?
	kurtosisExecutionResponseLines := make(chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine)

	go func() {
		defer close(kurtosisExecutionResponseLines)

		// Interpretation starts > send progress info (this line will be invisible as interpretation is super quick)
		progressInfo := binding_constructors.NewStarlarkRunResponseLineFromProgressInfo(
			startingInterpretationMsg, defaultCurrentStepNumber, defaultTotalStepsNumber)
		kurtosisExecutionResponseLines <- progressInfo

		instructionsList, interpretationError := runner.startosisInterpreter.Interpret(ctx, moduleId, serializedStartosis, serializedParams)
		if interpretationError != nil {
			interpretationError = maybeMakeMissingRunMethodErrorFriendlier(interpretationError, moduleId)
			kurtosisExecutionResponseLines <- binding_constructors.NewStarlarkRunResponseLineFromInterpretationError(interpretationError)
			return
		}
		totalNumberOfInstructions := uint32(len(instructionsList))
		logrus.Debugf("Successfully interpreted Starlark script into a series of Kurtosis instructions: \n%v",
			instructionsList)

		// Validation starts > send progress info
		progressInfo = binding_constructors.NewStarlarkRunResponseLineFromProgressInfo(
			startingValidationMsg, defaultCurrentStepNumber, totalNumberOfInstructions)
		kurtosisExecutionResponseLines <- progressInfo

		validationErrorsChan := runner.startosisValidator.Validate(ctx, instructionsList)
		if messagesWereReceived := forwardKurtosisResponseLineChannelUntilSourceIsClosed(validationErrorsChan, kurtosisExecutionResponseLines); messagesWereReceived {
			return
		}
		logrus.Debugf("Successfully validated Starlark script")

		// Execution starts > send progress info. This will soon be overridden byt the first instruction execution
		progressInfo = binding_constructors.NewStarlarkRunResponseLineFromProgressInfo(
			startingExecutionMsg, defaultCurrentStepNumber, totalNumberOfInstructions)
		kurtosisExecutionResponseLines <- progressInfo

		executionResponseLinesChan := runner.startosisExecutor.Execute(ctx, dryRun, instructionsList)
		forwardKurtosisResponseLineChannelUntilSourceIsClosed(executionResponseLinesChan, kurtosisExecutionResponseLines)
		logrus.Debugf("Successfully executed the list of %d Kurtosis instructions", len(instructionsList))
	}()
	return kurtosisExecutionResponseLines
}

func forwardKurtosisResponseLineChannelUntilSourceIsClosed(sourceChan <-chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, destChan chan<- *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine) bool {
	messagesWereReceived := false
	for executionResponseLine := range sourceChan {
		logrus.Debugf("Received kurtosis execution line Kurtosis:\n%v", executionResponseLine)
		destChan <- executionResponseLine
		messagesWereReceived = true
	}
	logrus.Debug("Kurtosis instructions stream was closed. Exiting execution loop")
	return messagesWereReceived
}

func maybeMakeMissingRunMethodErrorFriendlier(originalError *kurtosis_core_rpc_api_bindings.StarlarkInterpretationError, packageId string) *kurtosis_core_rpc_api_bindings.StarlarkInterpretationError {
	if originalError.GetErrorMessage() == missingRunMethodErrorFromStarlark {
		return binding_constructors.NewStarlarkInterpretationError(fmt.Sprintf("No 'run' function found in file '%v/main.star'; a 'run' entrypoint function is required in the main.star file of any Kurtosis package", packageId))
	}
	return originalError
}
