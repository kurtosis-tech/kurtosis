package startosis_engine

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers/magic_string_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_optimizer"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_optimizer/graph"
	"github.com/kurtosis-tech/stacktrace"
	"sync"
)

const (
	progressMsg      = "Execution in progress"
	ParallelismParam = "PARALLELISM"
)

var (
	skippedInstructionResult = "SKIPPED"
)

type StartosisExecutor struct {
	mutex             *sync.Mutex
	runtimeValueStore *runtime_value_store.RuntimeValueStore
	enclaveState      *graph.InstructionGraph
}

type ExecutionError struct {
	Error string
}

func NewStartosisExecutor(runtimeValueStore *runtime_value_store.RuntimeValueStore) *StartosisExecutor {
	return &StartosisExecutor{
		mutex:             &sync.Mutex{},
		runtimeValueStore: runtimeValueStore,
		enclaveState:      graph.NewInstructionGraph(),
	}
}

// Execute executes the list of Kurtosis instructions _asynchronously_ against the Kurtosis backend
// Consumers of this method should read the response lines channel and return as soon as one it is closed
//
// The channel of KurtosisExecutionResponseLine can contain three kinds of line:
// - A regular KurtosisInstruction that was successfully executed
// - A KurtosisExecutionError if the execution failed
// - A ProgressInfo to update the current "state" of the execution
func (executor *StartosisExecutor) Execute(ctx context.Context, dryRun bool, parallelism int, instructions []startosis_optimizer.PlannedInstruction, serializedScriptOutput string) <-chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine {
	executor.mutex.Lock()
	starlarkRunResponseLineStream := make(chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine)
	ctxWithParallelism := context.WithValue(ctx, ParallelismParam, parallelism)
	go func() {
		defer func() {
			executor.mutex.Unlock()
			close(starlarkRunResponseLineStream)
		}()

		var parentInstructionUuid graph.NodeUuid
		totalNumberOfInstructions := uint32(len(instructions))
		for index, plannedInstruction := range instructions {
			if plannedInstruction.IsOutOfScope() {
				parentInstructionUuid = plannedInstruction.GetUuid()
				continue
			}
			instruction := plannedInstruction.GetInstruction()
			instructionNumber := uint32(index + 1)
			progress := binding_constructors.NewStarlarkRunResponseLineFromSinglelineProgressInfo(
				progressMsg, instructionNumber, totalNumberOfInstructions)
			starlarkRunResponseLineStream <- progress

			canonicalInstruction := binding_constructors.NewStarlarkRunResponseLineFromInstruction(instruction.GetCanonicalInstruction())
			starlarkRunResponseLineStream <- canonicalInstruction

			if !dryRun {
				if plannedInstruction.IsSkipped() {
					parentInstructionUuid = plannedInstruction.GetUuid()
					starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromInstructionResult(skippedInstructionResult)
					continue
				}
				instructionOutput, err := instruction.Execute(ctxWithParallelism)
				if err != nil {
					propagatedError := stacktrace.Propagate(err, "An error occurred executing instruction (number %d) at %v:\n%v", instructionNumber, instruction.GetPositionInOriginalScript().String(), instruction.String())
					serializedError := binding_constructors.NewStarlarkExecutionError(propagatedError.Error())
					starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromExecutionError(serializedError)
					starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromRunFailureEvent()
					return
				}

				if instructionOutput != nil {
					starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromInstructionResult(*instructionOutput)
					if parentInstructionUuid == "" {
						parentInstructionUuid, err = executor.enclaveState.AddNode(instruction)
					} else {
						parentInstructionUuid, err = executor.enclaveState.AddNode(instruction, parentInstructionUuid)
					}
					if err != nil {
						propagatedError := stacktrace.Propagate(err, "An error occurred persisting state of enclave after instruction number %d (at %v) was successfully executed. The execution cannot continue.", instructionNumber, instruction.GetPositionInOriginalScript().String())
						serializedError := binding_constructors.NewStarlarkExecutionError(propagatedError.Error())
						starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromExecutionError(serializedError)
						starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromRunFailureEvent()
					}
				}
			}
		}

		if !dryRun {
			scriptWithValuesReplaced, err := magic_string_helper.ReplaceRuntimeValueInString(serializedScriptOutput, executor.runtimeValueStore)
			if err != nil {
				propagatedErr := stacktrace.Propagate(err, "An error occurred while replacing the runtime values in the output of the script")
				serializedError := binding_constructors.NewStarlarkExecutionError(propagatedErr.Error())
				starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromExecutionError(serializedError)
				starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromRunFailureEvent()
				return
			}
			starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromRunSuccessEvent(scriptWithValuesReplaced)
		}
	}()
	return starlarkRunResponseLineStream
}

func (executor *StartosisExecutor) GetEnclaveState() *graph.InstructionGraph {
	return executor.enclaveState
}
