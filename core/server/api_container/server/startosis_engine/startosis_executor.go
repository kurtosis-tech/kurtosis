package startosis_engine

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/dependency_graph"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_plan_persistence"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers/magic_string_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
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
	mutex              *sync.Mutex
	starlarkValueSerde *kurtosis_types.StarlarkValueSerde
	enclavePlan        *enclave_plan_persistence.EnclavePlan
	enclaveDb          *enclave_db.EnclaveDB
	runtimeValueStore  *runtime_value_store.RuntimeValueStore
}

type ExecutionError struct {
	Error string
}

func NewStartosisExecutor(starlarkValueSerde *kurtosis_types.StarlarkValueSerde, runtimeValueStore *runtime_value_store.RuntimeValueStore, enclavePlan *enclave_plan_persistence.EnclavePlan, enclaveDb *enclave_db.EnclaveDB) *StartosisExecutor {
	return &StartosisExecutor{
		mutex:              &sync.Mutex{},
		starlarkValueSerde: starlarkValueSerde,
		enclaveDb:          enclaveDb,
		enclavePlan:        enclavePlan,
		runtimeValueStore:  runtimeValueStore,
	}
}

// Execute executes the list of Kurtosis instructions _asynchronously_ against the Kurtosis backend
// Consumers of this method should read the response lines channel and return as soon as one it is closed
//
// The channel of KurtosisExecutionResponseLine can contain three kinds of line:
// - A regular KurtosisInstruction that was successfully executed
// - A KurtosisExecutionError if the execution failed
// - A ProgressInfo to update the current "state" of the execution
func (executor *StartosisExecutor) Execute(ctx context.Context, dryRun bool, parallelism int, indexOfFirstInstructionInEnclavePlan int, instructionsSequence []*instructions_plan.ScheduledInstruction, serializedScriptOutput string, instructionDependencyGraph map[instructions_plan.ScheduledInstructionUuid][]instructions_plan.ScheduledInstructionUuid) <-chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine {
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

		executor.enclavePlan = executor.enclavePlan.PartialDeepClone(indexOfFirstInstructionInEnclavePlan)

		defer func() {
			// TODO: we now perist the plan at the end of the execution. We could persist it everytime an instruction
			//  is executed, to be resilient to the APIC being stopped in the middle of a Starlark script execution
			//  Or we could even persist it only when the APIC is stopped. This seems to be a good middle ground, but
			//  can be tuned according the our future needs
			logrus.Infof("Persisting enclave plan composed of %d instructions into enclave database", executor.enclavePlan.Size())
			if err := executor.enclavePlan.Persist(executor.enclaveDb); err != nil {
				logrus.Errorf("An error occurred persisting the enclave plan at the end of the execution of the" +
					"package. The enclave will continue to run, but next runs of Starlark package might not be executed" +
					"as expected.")
			}
		}()

		logrus.Debugf("Transfered %d instructions from previous enclave plan to keep the enclave state consistent", executor.enclavePlan.Size())

		totalNumberOfInstructions := uint32(len(instructionsSequence))
		totalExecutionDuration := time.Duration(0)
		// instructionNumToDuration := make(map[int]time.Duration)

		for _, scheduledInstruction := range instructionsSequence {
			// add the instruction into the current enclave plan
			enclavePlanInstruction, err := scheduledInstruction.GetInstruction().GetPersistableAttributes().SetUuid(
				string(scheduledInstruction.GetUuid()),
			).SetReturnedValue(
				executor.starlarkValueSerde.Serialize(scheduledInstruction.GetReturnedValue()),
			).Build()
			if err != nil {
				sendErrorAndFail(starlarkRunResponseLineStream, totalExecutionDuration, err, "An error occurred persisting instruction (number %d) at %v after it's been executed:\n%v", instructionNumber, instruction.GetPositionInOriginalScript().String(), instruction.String())
			}
			executor.enclavePlan.AppendInstruction(enclavePlanInstruction)
		}

		wgSenders := sync.WaitGroup{}
		for index, scheduledInstruction := range instructionsSequence {
			wgSenders.Add(1)
			go func() {
				defer wgSenders.Done()

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
					var duration time.Duration
					if scheduledInstruction.IsExecuted() {
						logrus.Infof("INSTRUCTION %d ALREADY EXECUTED", instructionNumber)
						// instruction already executed within this enclave. Do not run it
						instructionOutput = &skippedInstructionOutput
					} else {
						// startTime := time.Now()
						instructionOutput, err = instruction.Execute(ctxWithParallelism)
						// duration = time.Since(startTime)
						// totalExecutionDuration += duration
						// instructionNumToDuration[index+1] = duration
					}
					if err != nil {
						sendErrorAndFail(starlarkRunResponseLineStream, totalExecutionDuration, err, "An error occurred executing instruction (number %d) at %v:\n%v", instructionNumber, instruction.GetPositionInOriginalScript().String(), instruction.String())
						return
					}
					if instructionOutput != nil {
						instructionOutputStr := *instructionOutput
						if len(instructionOutputStr) > outputSizeLimit {
							instructionOutputStr = fmt.Sprintf("%s%s", instructionOutputStr[0:outputSizeLimit], outputLimitReachedSuffix)
						}
						starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromInstructionResult(instructionOutputStr, duration)
					}
					// add the instruction into the current enclave plan
					// enclavePlanInstruction, err := scheduledInstruction.GetInstruction().GetPersistableAttributes().SetUuid(
					// 	string(scheduledInstruction.GetUuid()),
					// ).SetReturnedValue(
					// 	executor.starlarkValueSerde.Serialize(scheduledInstruction.GetReturnedValue()),
					// ).Build()
					// if err != nil {
					// 	sendErrorAndFail(starlarkRunResponseLineStream, totalExecutionDuration, err, "An error occurred persisting instruction (number %d) at %v after it's been executed:\n%v", instructionNumber, instruction.GetPositionInOriginalScript().String(), instruction.String())
					// }
					// executor.enclavePlan.AppendInstruction(enclavePlanInstruction)
				}

				instructionsDependencyGraph := make(map[dependency_graph.ScheduledInstructionUuid][]dependency_graph.ScheduledInstructionUuid)
				for instructionUuid, dependencies := range instructionDependencyGraph {
					instructionsDependencyGraph[dependency_graph.ScheduledInstructionUuid(instructionUuid)] = make([]dependency_graph.ScheduledInstructionUuid, len(dependencies))
					for i, dependency := range dependencies {
						instructionsDependencyGraph[dependency_graph.ScheduledInstructionUuid(instructionUuid)][i] = dependency_graph.ScheduledInstructionUuid(dependency)
					}
				}

				// logrus.Infof("Computing parallel execution time for instructionsDependencyGraph")
				totalParallelExecutionDuration := time.Duration(0)
				// totalParallelExecutionDuration := dependency_graph.ComputeParallelExecutionTime(instructionsDependencyGraph, instructionNumToDuration)
				// logrus.Infof("totalParallelExecutionDuration: %v", totalParallelExecutionDuration)
				// printInstructionToDuration(instructionNumToDuration)

				if !dryRun {
					logrus.Debugf("Serialized script output before runtime value replace: '%v'", serializedScriptOutput)
					scriptWithValuesReplaced, err := magic_string_helper.ReplaceRuntimeValueInString(serializedScriptOutput, executor.runtimeValueStore)
					if err != nil {
						sendErrorAndFail(starlarkRunResponseLineStream, totalExecutionDuration, err, "An error occurred while replacing the runtime values in the output of the script")
						return
					}
					starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromRunSuccessEvent(scriptWithValuesReplaced, totalExecutionDuration, totalParallelExecutionDuration)
					logrus.Debugf("Current enclave plan has been updated. It now contains %d instructions", executor.enclavePlan.Size())
				} else {
					logrus.Debugf("Current enclave plan remained the same as the it was a dry-run. It contains %d instructions", executor.enclavePlan.Size())
				}
			}()
		}

		wgSenders.Wait()
	}()

	return starlarkRunResponseLineStream
}

func (executor *StartosisExecutor) GetCurrentEnclavePLan() *enclave_plan_persistence.EnclavePlan {
	return executor.enclavePlan
}

func sendErrorAndFail(starlarkRunResponseLineStream chan<- *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, totalExecutionDuration time.Duration, err error, msg string, msgArgs ...interface{}) {
	propagatedErr := stacktrace.Propagate(err, msg, msgArgs...)
	serializedError := binding_constructors.NewStarlarkExecutionError(propagatedErr.Error())
	starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromExecutionError(serializedError)
	starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromRunFailureEventWithDuration(totalExecutionDuration)
}

func printInstructionToDuration(instructionNumToDuration map[int]time.Duration) {
	// Convert durations to string representations for JSON marshaling
	stringDurations := make(map[int]string)
	for num, duration := range instructionNumToDuration {
		stringDurations[num] = duration.String()
	}

	jsonBytes, err := json.MarshalIndent(stringDurations, "", "  ")
	if err != nil {
		logrus.Errorf("Failed to marshal instruction durations to JSON: %v", err)
		return
	}

	err = os.WriteFile("/tmp/instruction_durations.json", jsonBytes, 0644)
	if err != nil {
		logrus.Errorf("Failed to write instruction durations to file: %v", err)
		return
	}
	logrus.Infof("Wrote instruction durations to /tmp/instruction_durations.json")
}
