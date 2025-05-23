package startosis_engine

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/benchmark"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan/resolver"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/starlark_warning"
	"github.com/sirupsen/logrus"
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
	startingInterpretationMsg = "Interpreting plan - execution will begin shortly"
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
	packageReplaceOptions map[string]string,
	mainFunctionName string,
	relativePathToMainFile string,
	serializedStartosis string,
	serializedParams string,
	imageDownloadMode image_download_mode.ImageDownloadMode,
	nonBlockingMode bool,
	experimentalFeatures []kurtosis_core_rpc_api_bindings.KurtosisFeatureFlag,
) <-chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine {
	runner.mutex.Lock()
	starlark_warning.Clear()
	defer runner.mutex.Unlock()

	benchmark := benchmark.StartosisBenchmark{
		TimeToRunStartosisScript:    time.Duration(0),
		TimeToExecuteInstructions:   time.Duration(0),
		TimeToValidateInstructions:  time.Duration(0),
		TimeToInterpretInstructions: time.Duration(0),
	}

	beforeRunStartosisScript := time.Now()
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

			starlarkRunResponseLines <- binding_constructors.NewStarlarkRunResponseLineFromInfoMsg("⭐ us on GitHub - https://github.com/kurtosis-tech/kurtosis")
			close(starlarkRunResponseLines)
		}()

		var experimentalFeaturesStr []string
		for _, experimentalFeature := range experimentalFeatures {
			experimentalFeaturesStr = append(experimentalFeaturesStr, experimentalFeature.String())
		}
		logrus.Infof("Executing Starlark package '%s' with the following parameters: dry-run: '%v', parallelism: '%d', experimental features: '%s', main function name: '%s', params: '%s'",
			packageId,
			dryRun,
			parallelism,
			strings.Join(experimentalFeaturesStr, ", "),
			mainFunctionName,
			serializedParams)

		// Interpretation starts > send progress info (this line will be invisible as interpretation is super quick)
		progressInfo := binding_constructors.NewStarlarkRunResponseLineFromSinglelineProgressInfo(
			startingInterpretationMsg, defaultCurrentStepNumber, defaultTotalStepsNumber)
		starlarkRunResponseLines <- progressInfo

		beforeInterpretation := time.Now()
		// TODO: once we have feature flags, add a switch here to call InterpretAndOptimizePlan if the feature flag is
		//  turned on
		var serializedScriptOutput string
		var instructionsPlan *instructions_plan.InstructionsPlan
		var interpretationError *kurtosis_core_rpc_api_bindings.StarlarkInterpretationError
		if doesFeatureFlagsContain(experimentalFeatures, kurtosis_core_rpc_api_bindings.KurtosisFeatureFlag_NO_INSTRUCTIONS_CACHING) {
			serializedScriptOutput, instructionsPlan, interpretationError = runner.startosisInterpreter.Interpret(
				ctx,
				packageId,
				mainFunctionName,
				packageReplaceOptions,
				relativePathToMainFile,
				serializedStartosis,
				serializedParams,
				nonBlockingMode,
				enclave_structure.NewEnclaveComponents(),
				resolver.NewInstructionsPlanMask(0),
				imageDownloadMode,
			)
		} else {
			serializedScriptOutput, instructionsPlan, interpretationError = runner.startosisInterpreter.InterpretAndOptimizePlan(
				ctx,
				packageId,
				packageReplaceOptions,
				mainFunctionName,
				relativePathToMainFile,
				serializedStartosis,
				serializedParams,
				nonBlockingMode,
				runner.startosisExecutor.enclavePlan,
				imageDownloadMode,
			)
		}
		benchmark.TimeToInterpretInstructions = time.Since(beforeInterpretation)

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

		beforeValidation := time.Now()
		validationErrorsChan := runner.startosisValidator.Validate(ctx, instructionsSequence, imageDownloadMode)
		if isRunFinished, isRunSuccessful := forwardKurtosisResponseLineChannelUntilSourceIsClosed(validationErrorsChan, starlarkRunResponseLines); isRunFinished {
			if !isRunSuccessful {
				logrus.Warnf("An error occurred validating the sequence of Kurtosis instructions. See logs above for more details")
			}
			return
		}
		logrus.Debugf("Successfully validated Starlark script")
		benchmark.TimeToValidateInstructions = time.Since(beforeValidation)

		// Execution starts > send progress info. This will soon be overridden byt the first instruction execution
		progressInfo = binding_constructors.NewStarlarkRunResponseLineFromSinglelineProgressInfo(
			startingExecutionMsg, defaultCurrentStepNumber, totalNumberOfInstructions)
		starlarkRunResponseLines <- progressInfo

		beforeExecution := time.Now()
		executionResponseLinesChan := runner.startosisExecutor.Execute(ctx, dryRun, parallelism, instructionsPlan.GetIndexOfFirstInstruction(), instructionsSequence, serializedScriptOutput)
		if isRunFinished, isRunSuccessful := forwardKurtosisResponseLineChannelUntilSourceIsClosed(executionResponseLinesChan, starlarkRunResponseLines); !isRunFinished {
			logrus.Warnf("Execution finished but no 'RunFinishedEvent' was received through the stream. This is unexpected as every execution should be terminal.")
		} else if !isRunSuccessful {
			logrus.Warnf("An error occurred executing the sequence of Kurtosis instructions. See logs above for more details")
		} else {
			logrus.Debugf("Successfully executed Kurtosis plan composed of %d Kurtosis instructions", totalNumberOfInstructions)
		}
		benchmark.TimeToExecuteInstructions = time.Since(beforeExecution)
		benchmark.TimeToRunStartosisScript = time.Since(beforeRunStartosisScript)
		benchmark.OutputToFile()
	}()

	return starlarkRunResponseLines
}

func forwardKurtosisResponseLineChannelUntilSourceIsClosed(sourceChan <-chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, destChan chan<- *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine) (bool, bool) {
	isSuccessful := false
	isStarlarkRunFinished := false
	for executionResponseLine := range sourceChan {
		logrus.Debugf("Received kurtosis execution line Kurtosis:\n%v", executionResponseLine)
		if executionResponseLine.GetRunFinishedEvent() != nil {
			isStarlarkRunFinished = true
			isSuccessful = executionResponseLine.GetRunFinishedEvent().GetIsRunSuccessful()
		}
		destChan <- executionResponseLine
	}
	logrus.Debugf("Kurtosis instructions stream was closed. Exiting execution loop. Run finished: '%v'", isStarlarkRunFinished)
	return isStarlarkRunFinished, isSuccessful
}

func doesFeatureFlagsContain(featureFlags []kurtosis_core_rpc_api_bindings.KurtosisFeatureFlag, requestedFeatureFlag kurtosis_core_rpc_api_bindings.KurtosisFeatureFlag) bool {
	for _, featureFlag := range featureFlags {
		if featureFlag == requestedFeatureFlag {
			return true
		}
	}
	return false
}
