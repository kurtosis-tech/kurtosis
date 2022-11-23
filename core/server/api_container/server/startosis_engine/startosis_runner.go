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
		generatedInstructionsList, interpretationError := runner.startosisInterpreter.Interpret(ctx, moduleId, serializedStartosis, serializedParams)
		if interpretationError != nil {
			kurtosisExecutionResponseLines <- binding_constructors.NewKurtosisExecutionResponseLineFromInterpretationError(interpretationError)
			return
		}
		logrus.Debugf("Successfully interpreted Startosis script into a series of Kurtosis instructions: \n%v",
			generatedInstructionsList)

		validationErrors := runner.startosisValidator.Validate(ctx, generatedInstructionsList)
		if validationErrors != nil {
			// TODO(gb): push this logic down to the validator
			for _, validationError := range validationErrors.GetErrors() {
				kurtosisExecutionResponseLines <- binding_constructors.NewKurtosisExecutionResponseLineFromValidationError(validationError)
			}
			return
		}
		logrus.Debugf("Successfully validated Startosis script")

		kurtosisInstructionsStream, errChan := runner.startosisExecutor.Execute(ctx, dryRun, generatedInstructionsList)

	ReadChannelsLoop:
		for {
			select {
			case executedKurtosisInstruction, isChanOpen := <-kurtosisInstructionsStream:
				if !isChanOpen {
					logrus.Debug("Kurtosis instructions stream was closed. Exiting execution loop")
					break ReadChannelsLoop
				}
				logrus.Debugf("Received serialized Kurtosis instruction:\n%v", executedKurtosisInstruction.GetExecutableInstruction())
				kurtosisExecutionResponseLines <- binding_constructors.NewKurtosisExecutionResponseLineFromInstruction(executedKurtosisInstruction)
			case executionError, isChanOpen := <-errChan:
				if !isChanOpen {
					logrus.Debug("Kurtosis execution error channel was closed. Exiting execution loop")
					break ReadChannelsLoop
				}
				kurtosisExecutionResponseLines <- binding_constructors.NewKurtosisExecutionResponseLineFromExecutionError(executionError)
			}
		}
		logrus.Debugf("Successfully executed the list of %d Kurtosis instructions", len(generatedInstructionsList))
	}()
	return kurtosisExecutionResponseLines
}
