package startosis_engine

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/stacktrace"
	"sync"
)

type StartosisExecutor struct {
	mutex *sync.Mutex
}

type ExecutionError struct {
	Error string
}

func NewStartosisExecutor() *StartosisExecutor {
	return &StartosisExecutor{
		mutex: &sync.Mutex{},
	}
}

// Execute executes the list of Kurtosis instructions _asynchronously_ against the Kurtosis backend
// Consumers of this method should read the 3 channels in a select/case loop and return as soon as one is closed (they are all closed at the same time)
//
// The channels are the following:
// - One channel of instruction output. Everytime an instruction is executed and returns a non-nil output, it is written to this channel
// - One channel of serialized kurtosis instructions. Every successfully executed instruction will be serialized and sent to this channel
// - One channel for potential error. As soon as an instruction fails to be executed, an error will be sent to this channel and the execution will stop
func (executor *StartosisExecutor) Execute(ctx context.Context, dryRun bool, instructions []kurtosis_instruction.KurtosisInstruction) (<-chan *kurtosis_core_rpc_api_bindings.KurtosisInstruction, <-chan *kurtosis_core_rpc_api_bindings.KurtosisExecutionError) {
	executor.mutex.Lock()
	kurtosisInstructionsStream := make(chan *kurtosis_core_rpc_api_bindings.KurtosisInstruction)
	errorChan := make(chan *kurtosis_core_rpc_api_bindings.KurtosisExecutionError)

	go func() {
		defer func() {
			executor.mutex.Unlock()
			close(kurtosisInstructionsStream)
			close(errorChan)
		}()

		for index, instruction := range instructions {
			if !dryRun {
				instructionOutput, err := instruction.Execute(ctx)
				if err != nil {
					errorChan <- binding_constructors.NewKurtosisExecutionError(stacktrace.Propagate(err, "An error occurred executing instruction (number %d): \n%v", index+1, instruction.String()).Error())
					return
				}
				kurtosisInstructionsStream <- binding_constructors.AddResultToKurtosisInstruction(instruction.GetCanonicalInstruction(), instructionOutput)
			} else {
				kurtosisInstructionsStream <- instruction.GetCanonicalInstruction()
			}
		}
	}()

	return kurtosisInstructionsStream, errorChan
}
