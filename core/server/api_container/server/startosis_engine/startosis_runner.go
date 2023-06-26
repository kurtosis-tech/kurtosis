package startosis_engine

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan/resolver"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/starlark_warning"
	"github.com/sirupsen/logrus"
	"sync"
)

type StartosisRunner struct {
	startosisInterpreter *StartosisInterpreter

	startosisValidator *StartosisValidator

	startosisExecutor *StartosisExecutor

	mutex *sync.Mutex
}

const (
	defaultCurrentStepNumber  = 0
	defaultTotalStepsNumber   = 0
	startingInterpretationMsg = "Interpreting Starlark code - execution will begin shortly"
	startingValidationMsg     = "Starting validation"
	startingExecutionMsg      = "Starting execution"
)

func NewStartosisRunner(interpreter *StartosisInterpreter, validator *StartosisValidator, executor *StartosisExecutor) *StartosisRunner {
	return &StartosisRunner{
		startosisInterpreter: interpreter,
		startosisValidator:   validator,
		startosisExecutor:    executor,

		// we only expect one starlark package to run at a time against an enclave
		// this lock ensures that only warning set is accessed by one starlark run method
		mutex: &sync.Mutex{},
	}
}

func (runner *StartosisRunner) Run(
	ctx context.Context,
	dryRun bool,
	parallelism int,
	packageId string,
	mainFunctionName string,
	serializedStartosis string,
	serializedParams string,
) <-chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine {
	runner.mutex.Lock()
	starlark_warning.Clear()
	defer runner.mutex.Unlock()

	// TODO(gb): add metric tracking maybe?
	starlarkRunResponseLines := make(chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine)
	go func() {
		defer func() {
			warnings := starlark_warning.GetContentFromWarningSet()

			if len(warnings) > 0 {
				for _, warning := range warnings {
					// TODO: create a new binding_constructor for warning message
					starlarkRunResponseLines <- binding_constructors.NewStarlarkRunResponseLineFromWarning(warning)
				}
			}

			close(starlarkRunResponseLines)
		}()

		// Interpretation starts > send progress info (this line will be invisible as interpretation is super quick)
		progressInfo := binding_constructors.NewStarlarkRunResponseLineFromSinglelineProgressInfo(
			startingInterpretationMsg, defaultCurrentStepNumber, defaultTotalStepsNumber)
		starlarkRunResponseLines <- progressInfo

		// TODO: once we have feature flags, add a switch here to call InterpretAndOptimizePlan if the feature flag is
		//  turned on
		serializedScriptOutput, instructionsPlan, interpretationError := runner.startosisInterpreter.Interpret(
			ctx,
			packageId,
			mainFunctionName,
			serializedStartosis,
			serializedParams,
			resolver.NewInstructionsPlanMask(0),
		)
		if interpretationError != nil {
			starlarkRunResponseLines <- binding_constructors.NewStarlarkRunResponseLineFromInterpretationError(interpretationError)
			starlarkRunResponseLines <- binding_constructors.NewStarlarkRunResponseLineFromRunFailureEvent()
			return
		}
		totalNumberOfInstructions := uint32(instructionsPlan.Size())
		logrus.Debugf("Successfully interpreted Starlark script into a series of %d Kurtosis instructions",
			totalNumberOfInstructions)

		instructionsSequence, interpretationErr := instructionsPlan.GeneratePlan()
		if interpretationErr != nil {
			starlarkRunResponseLines <- binding_constructors.NewStarlarkRunResponseLineFromInterpretationError(interpretationErr.ToAPIType())
			starlarkRunResponseLines <- binding_constructors.NewStarlarkRunResponseLineFromRunFailureEvent()
			return
		}

		// Validation starts > send progress info
		progressInfo = binding_constructors.NewStarlarkRunResponseLineFromSinglelineProgressInfo(
			startingValidationMsg, defaultCurrentStepNumber, totalNumberOfInstructions)
		starlarkRunResponseLines <- progressInfo

		validationErrorsChan := runner.startosisValidator.Validate(ctx, instructionsSequence)
		if isRunFinished := forwardKurtosisResponseLineChannelUntilSourceIsClosed(validationErrorsChan, starlarkRunResponseLines); isRunFinished {
			return
		}
		logrus.Debugf("Successfully validated Starlark script")

		// Execution starts > send progress info. This will soon be overridden byt the first instruction execution
		progressInfo = binding_constructors.NewStarlarkRunResponseLineFromSinglelineProgressInfo(
			startingExecutionMsg, defaultCurrentStepNumber, totalNumberOfInstructions)
		starlarkRunResponseLines <- progressInfo

		executionResponseLinesChan := runner.startosisExecutor.Execute(ctx, dryRun, parallelism, instructionsSequence, serializedScriptOutput)
		if isRunFinished := forwardKurtosisResponseLineChannelUntilSourceIsClosed(executionResponseLinesChan, starlarkRunResponseLines); !isRunFinished {
			logrus.Warnf("Execution finished but no 'RunFinishedEvent' was received through the stream. This is unexpected as every execution should be terminal.")
		}
		logrus.Debugf("Successfully executed the list of %d Kurtosis instructions", totalNumberOfInstructions)

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
