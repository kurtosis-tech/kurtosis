package startosis_engine

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/sirupsen/logrus"
)

type StartosisRunner struct {
	startosisInterpreter *StartosisInterpreter

	startosisValidator *StartosisValidator

	startosisExecutor *StartosisExecutor
}

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
		kurtosisExecutionResponseLines <- binding_constructors.NewKurtosisExecutionResponseLineFromProgressInfo(
			"Interpreting Starlark code - execution will begin shortly", 0, 0)

		instructionsList, interpretationError := runner.startosisInterpreter.Interpret(ctx, moduleId, serializedStartosis, serializedParams)
		if interpretationError != nil {
			kurtosisExecutionResponseLines <- binding_constructors.NewKurtosisExecutionResponseLineFromInterpretationError(interpretationError)
			return
		}
		totalNumberOfInstructions := uint32(len(instructionsList))
		logrus.Debugf("Successfully interpreted Startosis script into a series of Kurtosis instructions: \n%v",
			instructionsList)

		// Validation starts > send progress info
		kurtosisExecutionResponseLines <- binding_constructors.NewKurtosisExecutionResponseLineFromProgressInfo(
			"Pre-validating Starlark code and downloading docker images - execution will begin shortly", 0, totalNumberOfInstructions)

		validationErrorsChan := runner.startosisValidator.Validate(ctx, instructionsList)
		forwardKurtosisResponseLineChannelUntilSourceIsClosed(validationErrorsChan, kurtosisExecutionResponseLines)
		logrus.Debugf("Successfully validated Startosis script")

		// Execution starts > send progress info. This will soon be overridden byt the first instruction execution
		kurtosisExecutionResponseLines <- binding_constructors.NewKurtosisExecutionResponseLineFromProgressInfo(
			"Starting execution", 0, totalNumberOfInstructions)

		executionResponseLinesChan := runner.startosisExecutor.Execute(ctx, dryRun, instructionsList)
		forwardKurtosisResponseLineChannelUntilSourceIsClosed(executionResponseLinesChan, kurtosisExecutionResponseLines)
		logrus.Debugf("Successfully executed the list of %d Kurtosis instructions", len(instructionsList))
	}()
	return kurtosisExecutionResponseLines
}

func forwardKurtosisResponseLineChannelUntilSourceIsClosed(sourceChan <-chan *kurtosis_core_rpc_api_bindings.KurtosisExecutionResponseLine, destChan chan<- *kurtosis_core_rpc_api_bindings.KurtosisExecutionResponseLine) {
	for executionResponseLine := range sourceChan {
		logrus.Debugf("Received kurtosis execution line Kurtosis:\n%v", executionResponseLine)
		destChan <- executionResponseLine
	}
	logrus.Debug("Kurtosis instructions stream was closed. Exiting execution loop")
}
