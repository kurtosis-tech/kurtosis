package startosis_engine

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers/magic_string_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/enclave_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"sync"
)

const (
	// we limit the output to 64k characters
	// TODO(tedi) get rid of this in favor of streaming
	outputSizeLimit          = 64 * 1024
	outputLimitReachedSuffix = "..."
	progressMsg              = "Execution in progress"
)

var (
	skippedInstructionOutput = "SKIPPED - This instruction has already been run in this enclave"
)

type StartosisExecutor struct {
	mutex                            *sync.Mutex
	enclavePlan                      *instructions_plan.InstructionsPlan
	runtimeValueStore                *runtime_value_store.RuntimeValueStore
	enclavePlanInstructionRepository *enclave_plan_instruction.EnclavePlanInstructionRepository
}

type ExecutionError struct {
	Error string
}

func NewStartosisExecutor(
	runtimeValueStore *runtime_value_store.RuntimeValueStore,
	enclavePlanInstructionRepository *enclave_plan_instruction.EnclavePlanInstructionRepository,
) *StartosisExecutor {
	return &StartosisExecutor{
		mutex:                            &sync.Mutex{},
		enclavePlan:                      instructions_plan.NewInstructionsPlan(),
		runtimeValueStore:                runtimeValueStore,
		enclavePlanInstructionRepository: enclavePlanInstructionRepository,
	}
}

// Execute executes the list of Kurtosis instructions _asynchronously_ against the Kurtosis backend
// Consumers of this method should read the response lines channel and return as soon as one it is closed
//
// The channel of KurtosisExecutionResponseLine can contain three kinds of line:
// - A regular KurtosisInstruction that was successfully executed
// - A KurtosisExecutionError if the execution failed
// - A ProgressInfo to update the current "state" of the execution
func (executor *StartosisExecutor) Execute(ctx context.Context, dryRun bool, parallelism int, indexOfFirstInstructionInEnclavePlan int, instructionsSequence []*instructions_plan.ScheduledInstruction, serializedScriptOutput string) <-chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine {
	executor.mutex.Lock()
	starlarkRunResponseLineStream := make(chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine)
	ctxWithParallelism := context.WithValue(ctx, startosis_constants.ParallelismParam, parallelism)
	go func() {
		defer func() {
			executor.mutex.Unlock()
			close(starlarkRunResponseLineStream)
		}()

		// TODO: for now the plan is append only, as each Starlark run happens on top of whatever exists in the enclave
		logrus.Debugf("Current enclave plan contains %d instuctions. About to process a new plan with %d instructions starting at index %d (dry-run: %v)",
			executor.enclavePlan.Size(), len(instructionsSequence), indexOfFirstInstructionInEnclavePlan, dryRun)

		var err error
		executor.enclavePlan, err = executor.enclavePlan.PartialClone(indexOfFirstInstructionInEnclavePlan)
		if err != nil {
			sendErrorAndFail(starlarkRunResponseLineStream, err, "An error occurred keeping the enclave plan up to date with the current enclave state")
		}

		logrus.Debugf("Transfered %d instructions from previous enclave plan to keep the enclave state consistent", executor.enclavePlan.Size())

		totalNumberOfInstructions := uint32(len(instructionsSequence))
		for index, scheduledInstruction := range instructionsSequence {
			instructionNumber := uint32(index + 1)
			progress := binding_constructors.NewStarlarkRunResponseLineFromSinglelineProgressInfo(
				progressMsg, instructionNumber, totalNumberOfInstructions)
			starlarkRunResponseLineStream <- progress

			instruction := scheduledInstruction.GetInstruction()
			canonicalInstruction := binding_constructors.NewStarlarkRunResponseLineFromInstruction(instruction.GetCanonicalInstruction(scheduledInstruction.IsExecuted()))
			starlarkRunResponseLineStream <- canonicalInstruction

			if !dryRun {
				var err error
				var instructionOutput *string
				var isExecuted bool

				instructionStr := scheduledInstruction.GetInstruction().String()
				enclavePlanInstruction, err := executor.enclavePlanInstructionRepository.Get(instructionStr)
				if err != nil {
					sendErrorAndFail(starlarkRunResponseLineStream, err, "An error occurred checking if there is an enclave instruction plan")
				}

				if enclavePlanInstruction != nil {
					isExecuted = enclavePlanInstruction.IsExecuted()
				}

				if isExecuted {
					// instruction already executed within this enclave. Do not run it
					instructionOutput = &skippedInstructionOutput
				} else {
					instructionOutput, err = instruction.Execute(ctxWithParallelism)
				}
				if err != nil {
					sendErrorAndFail(starlarkRunResponseLineStream, err, "An error occurred executing instruction (number %d) at %v:\n%v", instructionNumber, instruction.GetPositionInOriginalScript().String(), instruction.String())
					return
				}
				if instructionOutput != nil {
					instructionOutputStr := *instructionOutput
					if len(instructionOutputStr) > outputSizeLimit {
						instructionOutputStr = fmt.Sprintf("%s%s", instructionOutputStr[0:outputSizeLimit], outputLimitReachedSuffix)
					}
					starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromInstructionResult(instructionOutputStr)
				}
				// mark the instruction as executed and add it to the current instruction plan
				executor.enclavePlan.AddScheduledInstruction(scheduledInstruction).Executed(true)

				if enclavePlanInstruction == nil {
					enclavePlanInstruction = enclave_plan_instruction.NewEnclavePlanInstructionImpl(scheduledInstruction.GetInstruction().String(), scheduledInstruction.GetInstruction().GetCapabilites().GetEnclavePlanCapabilities())
					enclavePlanInstruction.Executed(true)
					if err := executor.enclavePlanInstructionRepository.SaveIfNotExist(enclavePlanInstruction); err != nil {
						sendErrorAndFail(starlarkRunResponseLineStream, err, "An error occurred while saving enclave instruction plan with UUID '%v'", scheduledInstruction.GetUuid())
					}
				} else {
					if err := executor.enclavePlanInstructionRepository.Executed(scheduledInstruction.GetInstruction().String(), true); err != nil {
						sendErrorAndFail(starlarkRunResponseLineStream, err, "An error occurred while setting enclave instruction plan with UUID '%v' has executed", scheduledInstruction.GetUuid())
					}
				}

			}
		}

		if !dryRun {
			logrus.Debugf("Serialized script output before runtime value replace: '%v'", serializedScriptOutput)
			scriptWithValuesReplaced, err := magic_string_helper.ReplaceRuntimeValueInString(serializedScriptOutput, executor.runtimeValueStore)
			if err != nil {
				sendErrorAndFail(starlarkRunResponseLineStream, err, "An error occurred while replacing the runtime values in the output of the script")
				return
			}
			starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromRunSuccessEvent(scriptWithValuesReplaced)
			logrus.Debugf("Current enclave plan has been updated. It now contains %d instructions", executor.enclavePlan.Size())
		} else {
			logrus.Debugf("Current enclave plan remained the same as the it was a dry-run. It contains %d instructions", executor.enclavePlan.Size())
		}
	}()
	return starlarkRunResponseLineStream
}

func (executor *StartosisExecutor) GetCurrentEnclavePLan() *instructions_plan.InstructionsPlan {
	return executor.enclavePlan
}

func sendErrorAndFail(starlarkRunResponseLineStream chan<- *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, err error, msg string, msgArgs ...interface{}) {
	propagatedErr := stacktrace.Propagate(err, msg, msgArgs...)
	serializedError := binding_constructors.NewStarlarkExecutionError(propagatedErr.Error())
	starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromExecutionError(serializedError)
	starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromRunFailureEvent()
}
