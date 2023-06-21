package startosis_engine

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers/magic_string_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"sync"
)

const (
	progressMsg      = "Execution in progress"
	ParallelismParam = "PARALLELISM"
)

var (
	skippedInstructionOutput = "SKIPPED"
)

type StartosisExecutor struct {
	mutex                   *sync.Mutex
	currentInstructionsPlan *instructions_plan.InstructionsPlan
	runtimeValueStore       *runtime_value_store.RuntimeValueStore
}

type ExecutionError struct {
	Error string
}

func NewStartosisExecutor(runtimeValueStore *runtime_value_store.RuntimeValueStore) *StartosisExecutor {
	return &StartosisExecutor{
		mutex:                   &sync.Mutex{},
		currentInstructionsPlan: instructions_plan.NewInstructionsPlan(),
		runtimeValueStore:       runtimeValueStore,
	}
}

// Execute executes the list of Kurtosis instructions _asynchronously_ against the Kurtosis backend
// Consumers of this method should read the response lines channel and return as soon as one it is closed
//
// The channel of KurtosisExecutionResponseLine can contain three kinds of line:
// - A regular KurtosisInstruction that was successfully executed
// - A KurtosisExecutionError if the execution failed
// - A ProgressInfo to update the current "state" of the execution
func (executor *StartosisExecutor) Execute(ctx context.Context, dryRun bool, parallelism int, scheduledInstructions []*instructions_plan.ScheduledInstruction, serializedScriptOutput string) <-chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine {
	executor.mutex.Lock()
	starlarkRunResponseLineStream := make(chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine)
	ctxWithParallelism := context.WithValue(ctx, ParallelismParam, parallelism)
	go func() {
		defer func() {
			executor.mutex.Unlock()
			close(starlarkRunResponseLineStream)
		}()

		logrus.Debugf("Current enclave plan contains %d instuctions. About to process a new plan with %d", executor.currentInstructionsPlan.Size(), len(scheduledInstructions))
		if !dryRun {
			// TODO: for now it's fine to start by clearing the current state and rebuilding it instruction after
			//  instructions. Once we start refining how we update the enclave, we will need to update this heuristic
			executor.currentInstructionsPlan = instructions_plan.NewInstructionsPlan()
		}
		totalNumberOfInstructions := uint32(len(scheduledInstructions))
		for index, scheduledInstruction := range scheduledInstructions {
			instructionNumber := uint32(index + 1)
			progress := binding_constructors.NewStarlarkRunResponseLineFromSinglelineProgressInfo(
				progressMsg, instructionNumber, totalNumberOfInstructions)
			starlarkRunResponseLineStream <- progress

			if scheduledInstruction.IsImportedFromPreviousEnclavePlan() {
				// do not execute this instruction as it has already been executed, but store it to the new enclave plan
				executor.currentInstructionsPlan.AddScheduledInstruction(scheduledInstruction)
				continue
			}

			instruction := scheduledInstruction.GetInstruction()
			canonicalInstruction := binding_constructors.NewStarlarkRunResponseLineFromInstruction(instruction.GetCanonicalInstruction())
			starlarkRunResponseLineStream <- canonicalInstruction

			if !dryRun {
				var err error
				var instructionOutput *string
				if scheduledInstruction.IsExecuted() {
					// instruction already executed within this enclave. Do not run it
					instructionOutput = &skippedInstructionOutput
				} else {
					instructionOutput, err = instruction.Execute(ctxWithParallelism)
				}
				if err != nil {
					sendErrorAndFail(starlarkRunResponseLineStream, err,
						"An error occurred executing instruction (number %d) at %v:\n%v",
						instructionNumber,
						instruction.GetPositionInOriginalScript().String(),
						instruction.String())
					return
				}

				if instructionOutput != nil {
					starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromInstructionResult(*instructionOutput)
				}

				// mark the instruction as executed and add it to the current instruction plan
				executor.currentInstructionsPlan.AddScheduledInstruction(scheduledInstruction).Executed(true)
			}
		}

		if !dryRun {
			scriptWithValuesReplaced, err := magic_string_helper.ReplaceRuntimeValueInString(serializedScriptOutput, executor.runtimeValueStore)
			if err != nil {
				sendErrorAndFail(starlarkRunResponseLineStream, err, "An error occurred while replacing the runtime values in the output of the script")
				return
			}
			starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromRunSuccessEvent(scriptWithValuesReplaced)
			logrus.Debugf("Current enclave plan has been updated. It now contains %d instructions", executor.currentInstructionsPlan.Size())
		} else {
			logrus.Debugf("Current enclave plan remained the same as the it was a dry-run. It contains %d instructions", executor.currentInstructionsPlan.Size())
		}
	}()
	return starlarkRunResponseLineStream
}

func (executor *StartosisExecutor) GetCurrentEnclavePlan() *instructions_plan.InstructionsPlan {
	return executor.currentInstructionsPlan
}

func sendErrorAndFail(starlarkRunResponseLineStream chan<- *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, err error, msg string, msgArgs ...interface{}) {
	propagatedErr := stacktrace.Propagate(err, msg, msgArgs...)
	serializedError := binding_constructors.NewStarlarkExecutionError(propagatedErr.Error())
	starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromExecutionError(serializedError)
	starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromRunFailureEvent()
}
