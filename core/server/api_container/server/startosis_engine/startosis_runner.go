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

func NewStartosisRunner(insterpreter *StartosisInterpreter, validator *StartosisValidator, executor *StartosisExecutor) *StartosisRunner {
	return &StartosisRunner{
		startosisInterpreter: insterpreter,
		startosisValidator:   validator,
		startosisExecutor:    executor,
	}
}

func (runner *StartosisRunner) Run(ctx context.Context, dryRun bool, moduleId string, serializedStartosis string, serializedParams string) <-chan *kurtosis_core_rpc_api_bindings.KurtosisResponseLine {
	// TODO(gb): add metric tracking maybe?
	responseLineChan := make(chan *kurtosis_core_rpc_api_bindings.KurtosisResponseLine)

	go func() {
		defer close(responseLineChan)
		generatedInstructionsList, interpretationError := runner.startosisInterpreter.Interpret(ctx, moduleId, serializedStartosis, serializedParams)
		if interpretationError != nil {
			responseLineChan <- binding_constructors.NewKurtosisResponseLineFromInterpretationError(interpretationError)
			return
		}
		logrus.Debugf("Successfully interpreted Startosis script into a series of Kurtosis instructions: \n%v",
			generatedInstructionsList)

		validationErrors := runner.startosisValidator.Validate(ctx, generatedInstructionsList)
		if validationErrors != nil {
			// TODO(gb): push this logic down to the validator
			for _, validationError := range validationErrors.GetErrors() {
				responseLineChan <- binding_constructors.NewKurtosisResponseLineFromValidationError(validationError)
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
				responseLineChan <- binding_constructors.NewKurtosisResponseLineFromInstruction(executedKurtosisInstruction)
			case executionError, isChanOpen := <-errChan:
				if !isChanOpen {
					logrus.Debug("Kurtosis execution error channel was closed. Exiting execution loop")
					break ReadChannelsLoop
				}
				responseLineChan <- binding_constructors.NewKurtosisResponseLineFromExecutionError(executionError)
			}
		}
		logrus.Debugf("Successfully executed the list of %d Kurtosis instructions", len(generatedInstructionsList))
	}()
	return responseLineChan
}
