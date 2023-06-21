package startosis_engine

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan"
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

	starlarkRunResponseLines := make(chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine)
	go func() {
		defer func() {
			warnings := starlark_warning.GetContentFromWarningSet()

			if len(warnings) > 0 {
				for _, warning := range warnings {
					starlarkRunResponseLines <- binding_constructors.NewStarlarkRunResponseLineFromWarning(warning)
				}
			}

			close(starlarkRunResponseLines)
		}()

		// Interpretation starts > send progress info (this line will be invisible as interpretation is super quick)
		progressInfo := binding_constructors.NewStarlarkRunResponseLineFromSinglelineProgressInfo(
			startingInterpretationMsg, defaultCurrentStepNumber, defaultTotalStepsNumber)
		starlarkRunResponseLines <- progressInfo

		serializedScriptOutput, optimizedInstructionsPlan, interpretationError := runner.interpretPackageAndOptimizePlan(
			ctx,
			packageId,
			mainFunctionName,
			serializedStartosis,
			serializedParams,
			runner.startosisExecutor.GetCurrentEnclavePlan(),
		)
		if interpretationError != nil {
			starlarkRunResponseLines <- binding_constructors.NewStarlarkRunResponseLineFromInterpretationError(interpretationError)
			starlarkRunResponseLines <- binding_constructors.NewStarlarkRunResponseLineFromRunFailureEvent()
			return
		}

		optimizedInstructionsPlanSequence, interpretationErr := optimizedInstructionsPlan.GeneratePlan()
		if interpretationErr != nil {
			starlarkRunResponseLines <- binding_constructors.NewStarlarkRunResponseLineFromInterpretationError(interpretationErr.ToAPIType())
			starlarkRunResponseLines <- binding_constructors.NewStarlarkRunResponseLineFromRunFailureEvent()
			return
		}
		totalNumberOfInstructions := uint32(len(optimizedInstructionsPlanSequence))
		logrus.Debugf("Successfully interpreted Starlark script into a series of '%d' Kurtosis instructions",
			totalNumberOfInstructions)

		// Validation starts > send progress info
		progressInfo = binding_constructors.NewStarlarkRunResponseLineFromSinglelineProgressInfo(
			startingValidationMsg, defaultCurrentStepNumber, totalNumberOfInstructions)
		starlarkRunResponseLines <- progressInfo

		validationErrorsChan := runner.startosisValidator.Validate(ctx, optimizedInstructionsPlanSequence)
		if isRunFinished := forwardKurtosisResponseLineChannelUntilSourceIsClosed(validationErrorsChan, starlarkRunResponseLines); isRunFinished {
			return
		}
		logrus.Debugf("Successfully validated Starlark script")

		// Execution starts > send progress info. This will soon be overridden byt the first instruction execution
		progressInfo = binding_constructors.NewStarlarkRunResponseLineFromSinglelineProgressInfo(
			startingExecutionMsg, defaultCurrentStepNumber, totalNumberOfInstructions)
		starlarkRunResponseLines <- progressInfo

		executionResponseLinesChan := runner.startosisExecutor.Execute(ctx, dryRun, parallelism, optimizedInstructionsPlanSequence, serializedScriptOutput)
		if isRunFinished := forwardKurtosisResponseLineChannelUntilSourceIsClosed(executionResponseLinesChan, starlarkRunResponseLines); !isRunFinished {
			logrus.Warnf("Execution finished but no 'RunFinishedEvent' was received through the stream. This is unexpected as every execution should be terminal.")
		}
		logrus.Debugf("Successfully executed the list of %d Kurtosis instructions", totalNumberOfInstructions)

	}()
	return starlarkRunResponseLines
}

func (runner *StartosisRunner) interpretPackageAndOptimizePlan(
	ctx context.Context,
	packageId string,
	mainFunctionName string,
	serializedStartosis string,
	serializedParams string,
	currentEnclavePlan *instructions_plan.InstructionsPlan,
) (
	string,
	*instructions_plan.InstructionsPlan,
	*kurtosis_core_rpc_api_bindings.StarlarkInterpretationError,
) {
	// run interpretation with no mask at all to generate the list of instructions as if the enclave was empty
	emptyPlanInstructionsMask := resolver.NewInstructionsPlanMask(0)
	naiveInstructionsPlanSerializedScriptOutput, naiveInstructionsPlan, interpretationErrorApi := runner.startosisInterpreter.Interpret(ctx, packageId, mainFunctionName, serializedStartosis, serializedParams, emptyPlanInstructionsMask)
	if interpretationErrorApi != nil {
		return "", nil, interpretationErrorApi
	}

	naiveInstructionsPlanSequence, interpretationErr := naiveInstructionsPlan.GeneratePlan()
	if interpretationErr != nil {
		return "", nil, interpretationErr.ToAPIType()
	}
	logrus.Debugf("First interpretation of package generated %d instructions", len(naiveInstructionsPlanSequence))

	currentEnclavePlanSequence, interpretationErr := currentEnclavePlan.GeneratePlan()
	if interpretationErr != nil {
		return "", nil, interpretationErr.ToAPIType()
	}
	logrus.Debugf("Current enclave state contains %d instructions", len(currentEnclavePlanSequence))
	logrus.Debugf("Starting iterations to find the best plan to execute given the current state of the enclave")

	// We're going to iterate this way:
	// 1. Find an instruction in the current enclave plan matching the first instruction of the new plan
	// 2. Recopy all instructions prior to the match into the optimized plan
	// 3. Recopy all following instruction from the current enclave plan into an Instructions Plan Mask -> the reason
	//    we're naively recopying all the following instructions, not just the ones that depends on this instruction
	//    is because right now, we don't have the ability to know which instructions depends on which. We assume that
	//    all instructions executed AFTER this one will depend on it, to stay on the safe side
	// 4. Run the interpretation with the mask.
	//     - If it's successful, then we've found the optimized plan
	//     - if it's not successful, then the mask is not compatible with the package. Go back to step 1
	firstPossibleIndexForMatchingInstruction := 0
	for {
		// initialize an empty optimized plan and an empty the mask
		potentialMask := resolver.NewInstructionsPlanMask(len(naiveInstructionsPlanSequence))
		optimizedPlan := instructions_plan.NewInstructionsPlan()

		// find the index of an instruction in the current enclave plan matching the FIRST instruction of our instructions plan generated by the first interpretation
		matchingInstructionIdx := findFirstEqualInstructionPastIndex(currentEnclavePlanSequence, naiveInstructionsPlanSequence, firstPossibleIndexForMatchingInstruction)
		if matchingInstructionIdx >= 0 {
			logrus.Debugf("Found an instruction in enclave state at index %d which matches the first instruction of the new instructions plan", matchingInstructionIdx)
			// we found a match
			// -> First recopy all enclave state instructions prior to this match to the optimized plan. Those won't
			// be executed, but they need to be part of the plan to keep the state of the enclave accurate
			logrus.Debugf("Copying %d instructions from current enclave plan to new plan. Those instructions won't be executed but need to be kept in the enclave plan", matchingInstructionIdx)
			for i := 0; i < matchingInstructionIdx; i++ {
				optimizedPlan.AddScheduledInstruction(currentEnclavePlanSequence[i]).ImportedFromPreviousEnclavePlan(true)
			}
			// -> Then recopy all instructions past this match from the enclave state to the mask
			// Those instructions are the instructions that will mask the instructions for the newly submitted plan
			numberOfInstructionCopiedToMask := 0
			for copyIdx := matchingInstructionIdx; copyIdx < len(currentEnclavePlanSequence); copyIdx++ {
				if numberOfInstructionCopiedToMask >= potentialMask.Size() {
					// the mask is already full, can't recopy more instructions, stop here
					break
				}
				potentialMask.InsertAt(numberOfInstructionCopiedToMask, currentEnclavePlanSequence[copyIdx])
				numberOfInstructionCopiedToMask += 1
			}
			logrus.Debugf("Writing %d instruction at the beginning of the plan mask, leaving %d empty at the end", numberOfInstructionCopiedToMask, potentialMask.Size()-numberOfInstructionCopiedToMask)
		} else {
			// We cannot find any more instructions inside the enclave state matching the first instruction of the plan
			for _, currentPlanInstruction := range currentEnclavePlanSequence {
				optimizedPlan.AddScheduledInstruction(currentPlanInstruction).ImportedFromPreviousEnclavePlan(true)
			}
			for _, newPlanInstruction := range naiveInstructionsPlanSequence {
				optimizedPlan.AddScheduledInstruction(newPlanInstruction)
			}
			logrus.Debugf("Exhausted all possibilities. Concatenated the previous enclave plan with the new plan to obtain a %d instructions plan", optimizedPlan.Size())
			return naiveInstructionsPlanSerializedScriptOutput, optimizedPlan, nil
		}

		// Now that we have a potential plan mask, try running interpretation again using this plan mask
		attemptSerializedScriptOutput, attemptInstructionsPlan, interpretationErrorApi := runner.startosisInterpreter.Interpret(ctx, packageId, mainFunctionName, serializedStartosis, serializedParams, potentialMask)
		if interpretationErrorApi != nil {
			logrus.Debug("Interpreting the package again with the plan mask failed. Ignoring this mask")
			// if it throws an error, we know this is because of the mask because the first interpretation with no
			// mask succeeded. It therefore means the mask is invalid, we need to try another one
			firstPossibleIndexForMatchingInstruction += 1
			continue
		}

		// no error happened, it seems we found a good mask
		// -> recopy all instructions from the interpretation to the optimized plan
		attemptInstructionsPlanSequence, interpretationErr := attemptInstructionsPlan.GeneratePlan()
		if interpretationErr != nil {
			return "", nil, interpretationErr.ToAPIType()
		}

		logrus.Debugf("Interpreting the package again with the plan mask succeeded and generated %d new instructions. Adding them to the new optimized plan", attemptInstructionsPlan.Size())
		for _, scheduledInstruction := range attemptInstructionsPlanSequence {
			optimizedPlan.AddScheduledInstruction(scheduledInstruction)
		}

		// there might be still be instructions in the current enclave plan that have not been imported to the
		// optimized plan
		// for now, we support this only if no new instructions will be executed for this run. If that's not the case
		// continue the loop in the hope of finding another mask
		if len(currentEnclavePlanSequence) > matchingInstructionIdx+optimizedPlan.Size() {
			logrus.Debugf("There are %d instructions remaining in the current state that have not been transferred to the new plan. Transferring them now", len(currentEnclavePlanSequence)-matchingInstructionIdx+optimizedPlan.Size())
			atLeastOneInstructionWillBeExecuted := false
			for _, instructionThatWillPotentiallyBeRun := range attemptInstructionsPlanSequence {
				if !instructionThatWillPotentiallyBeRun.IsExecuted() {
					atLeastOneInstructionWillBeExecuted = true
				}
			}
			if atLeastOneInstructionWillBeExecuted {
				logrus.Debugf("The remaining instructions in the current enclave plan cannot be transferred to the new plan because this plan contains instructions that will be executed." +
					"The remaining instructions might depend on those and Kurtosis cannot re-run them (this is unsupported for now)")
				continue
			}
			// recopy all remaining instructions into the optimized plan
			for _, remainingInstructionFromCurrentEnclaveState := range currentEnclavePlanSequence[matchingInstructionIdx+optimizedPlan.Size():] {
				optimizedPlan.AddScheduledInstruction(remainingInstructionFromCurrentEnclaveState).ImportedFromPreviousEnclavePlan(true)
			}
		}
		// finally we can return the optimized plan as well as the serialized script output returned by the last
		// interpretation attempt
		return attemptSerializedScriptOutput, optimizedPlan, nil
	}
}

func findFirstEqualInstructionPastIndex(currentEnclaveInstructionsList []*instructions_plan.ScheduledInstruction, naiveInstructionsList []*instructions_plan.ScheduledInstruction, minIndex int) int {
	if len(naiveInstructionsList) == 0 {
		return -1 // no result as the naiveInstructionsList is empty
	}
	for i := minIndex; i < len(currentEnclaveInstructionsList); i++ {
		if currentEnclaveInstructionsList[i].GetInstruction().String() == naiveInstructionsList[0].GetInstruction().String() {
			return i
		}
	}
	return -1 // no match
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
