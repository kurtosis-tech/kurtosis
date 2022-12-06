package startosis_engine

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/sirupsen/logrus"
	"regexp"
	"strings"
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
	startingValidationMsg     = "Starting validation"
	startingExecutionMsg      = "Starting execution"

	missingRunMethodErrorPrefixFromStarlarkPackage = "Evaluation error: module has no .run field or method\n\tat"
	missingRunMethodErrorSuffixFromStarlarkPackage = "3:32]: <toplevel>"
	missingRunMethodErrorFromStarlarkScriptPattern = "Multiple errors caught interpreting the Starlark script. Listing each of them below.\n\tat \\[\\d+:1\\]: undefined: run"
)

var (
	missingRunMethodErrorFromStarlarkScriptRegex = regexp.MustCompile(missingRunMethodErrorFromStarlarkScriptPattern)
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
	starlarkRunResponseLines := make(chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine)

	go func() {
		defer close(starlarkRunResponseLines)

		// Interpretation starts > send progress info (this line will be invisible as interpretation is super quick)
		progressInfo := binding_constructors.NewStarlarkRunResponseLineFromSinglelineProgressInfo(
			startingInterpretationMsg, defaultCurrentStepNumber, defaultTotalStepsNumber)
		starlarkRunResponseLines <- progressInfo

		serializedScriptOutput, instructionsList, interpretationError := runner.startosisInterpreter.Interpret(ctx, moduleId, serializedStartosis, serializedParams)
		if interpretationError != nil {
			interpretationError = maybeMakeMissingRunMethodErrorFriendlier(interpretationError, moduleId)
			starlarkRunResponseLines <- binding_constructors.NewStarlarkRunResponseLineFromInterpretationError(interpretationError)
			starlarkRunResponseLines <- binding_constructors.NewStarlarkRunResponseLineFromRunFailureEvent()
			return
		}
		totalNumberOfInstructions := uint32(len(instructionsList))
		logrus.Debugf("Successfully interpreted Starlark script into a series of Kurtosis instructions: \n%v",
			instructionsList)

		// Validation starts > send progress info
		progressInfo = binding_constructors.NewStarlarkRunResponseLineFromSinglelineProgressInfo(
			startingValidationMsg, defaultCurrentStepNumber, totalNumberOfInstructions)
		starlarkRunResponseLines <- progressInfo

		validationErrorsChan := runner.startosisValidator.Validate(ctx, instructionsList)
		if isRunFinished := forwardKurtosisResponseLineChannelUntilSourceIsClosed(validationErrorsChan, starlarkRunResponseLines); isRunFinished {
			return
		}
		logrus.Debugf("Successfully validated Starlark script")

		// Execution starts > send progress info. This will soon be overridden byt the first instruction execution
		progressInfo = binding_constructors.NewStarlarkRunResponseLineFromSinglelineProgressInfo(
			startingExecutionMsg, defaultCurrentStepNumber, totalNumberOfInstructions)
		starlarkRunResponseLines <- progressInfo

		executionResponseLinesChan := runner.startosisExecutor.Execute(ctx, dryRun, instructionsList, serializedScriptOutput)
		if isRunFinished := forwardKurtosisResponseLineChannelUntilSourceIsClosed(executionResponseLinesChan, starlarkRunResponseLines); !isRunFinished {
			logrus.Warnf("Execution finished but no 'RunFinishedEvent' was received through the stream. This is unexpected as every execution should be terminal.")
		}
		logrus.Debugf("Successfully executed the list of %d Kurtosis instructions", len(instructionsList))
	}()
	return starlarkRunResponseLines
}

func forwardKurtosisResponseLineChannelUntilSourceIsClosed(sourceChan <-chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, destChan chan<- *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine) bool {
	isStarlarkRunFinished := false
	for executionResponseLine := range sourceChan {
		logrus.Debugf("Received kurtosis execution line Kurtosis:\n%v", executionResponseLine)
		if executionResponseLine.GetRunFinishedEvent() != nil {
			isStarlarkRunFinished = true
		}
		destChan <- executionResponseLine
	}
	logrus.Debugf("Kurtosis instructions stream was closed. Exiting execution loop. Run finishedL '%v'", isStarlarkRunFinished)
	return isStarlarkRunFinished
}

func maybeMakeMissingRunMethodErrorFriendlier(originalError *kurtosis_core_rpc_api_bindings.StarlarkInterpretationError, packageId string) *kurtosis_core_rpc_api_bindings.StarlarkInterpretationError {
	if strings.HasPrefix(originalError.GetErrorMessage(), missingRunMethodErrorPrefixFromStarlarkPackage) && strings.HasSuffix(originalError.GetErrorMessage(), missingRunMethodErrorSuffixFromStarlarkPackage) {
		return binding_constructors.NewStarlarkInterpretationError(fmt.Sprintf("No 'run' function found in file '%v/main.star'; a 'run' entrypoint function is required in the main.star file of any Kurtosis package", packageId))
	}

	if missingRunMethodErrorFromStarlarkScriptRegex.MatchString(originalError.GetErrorMessage()) {
		return binding_constructors.NewStarlarkInterpretationError("No 'run' function found in the script; a 'run' entrypoint function with the signature `run(args)` is required in any Kurtosis script")
	}

	return originalError
}
