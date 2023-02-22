package startosis_engine

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	metrics_client "github.com/kurtosis-tech/metrics-library/golang/lib/client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"sync"
)

const (
	progressMsg                   = "Execution in progress"
	ParallelismParam              = "PARALLELISM"
	executionFinishedSuccessfully = true
	executionFailed               = false
)

type StartosisExecutor struct {
	mutex          *sync.Mutex
	metricsClient  metrics_client.MetricsClient
	serviceNetwork service_network.ServiceNetwork
}

type ExecutionError struct {
	Error string
}

func NewStartosisExecutor(metricsClient metrics_client.MetricsClient, serviceNetwork service_network.ServiceNetwork) *StartosisExecutor {
	return &StartosisExecutor{
		mutex:          &sync.Mutex{},
		metricsClient:  metricsClient,
		serviceNetwork: serviceNetwork,
	}
}

// Execute executes the list of Kurtosis instructions _asynchronously_ against the Kurtosis backend
// Consumers of this method should read the response lines channel and return as soon as one it is closed
//
// The channel of KurtosisExecutionResponseLine can contain three kinds of line:
// - A regular KurtosisInstruction that was successfully executed
// - A KurtosisExecutionError if the execution failed
// - A ProgressInfo to update the current "state" of the execution
func (executor *StartosisExecutor) Execute(ctx context.Context, dryRun bool, parallelism int, instructions []kurtosis_instruction.KurtosisInstruction, serializedScriptOutput string, packageId string) <-chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine {
	executor.mutex.Lock()
	starlarkRunResponseLineStream := make(chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine)
	ctxWithParallelism := context.WithValue(ctx, ParallelismParam, parallelism)
	go func() {
		defer func() {
			executor.mutex.Unlock()
			close(starlarkRunResponseLineStream)
		}()

		totalNumberOfInstructions := uint32(len(instructions))
		for index, instruction := range instructions {
			instructionNumber := uint32(index + 1)
			progress := binding_constructors.NewStarlarkRunResponseLineFromSinglelineProgressInfo(
				progressMsg, instructionNumber, totalNumberOfInstructions)
			starlarkRunResponseLineStream <- progress

			canonicalInstruction := binding_constructors.NewStarlarkRunResponseLineFromInstruction(instruction.GetCanonicalInstruction())
			starlarkRunResponseLineStream <- canonicalInstruction

			if !dryRun {
				instructionOutput, err := instruction.Execute(ctxWithParallelism)
				if err != nil {
					propagatedError := stacktrace.Propagate(err, "An error occurred executing instruction (number %d): \n%v", instructionNumber, instruction.String())
					serializedError := binding_constructors.NewStarlarkExecutionError(propagatedError.Error())
					numServices := len(executor.serviceNetwork.GetServiceNames())
					if err := executor.metricsClient.TrackKurtosisRunFinishedEvent(packageId, numServices, executionFailed); err != nil {
						logrus.Errorf("An error occurred tracking kurtosis run finished event \n%s", err)
					}
					starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromExecutionError(serializedError)
					starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromRunFailureEvent()
					return
				}
				if instructionOutput != nil {
					starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromInstructionResult(*instructionOutput)
				}
			}
		}

		numServices := len(executor.serviceNetwork.GetServiceNames())
		if err := executor.metricsClient.TrackKurtosisRunFinishedEvent(packageId, numServices, executionFinishedSuccessfully); err != nil {
			logrus.Errorf("An error occurred tracking kurtosis run finished event \n%s", err)
		}
		// TODO(gb): we should run magic string replacement on the output
		starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromRunSuccessEvent(serializedScriptOutput)
	}()
	return starlarkRunResponseLineStream
}
