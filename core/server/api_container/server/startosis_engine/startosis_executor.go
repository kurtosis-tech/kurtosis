package startosis_engine

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
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

// Execute executes the list of Kurtosis instructions against the Kurtosis backend
// It returns a potential execution error if something went wrong.
// It returns a error if something unexpected happens outside the execution of the script
func (executor *StartosisExecutor) Execute(ctx context.Context, instructions []kurtosis_instruction.KurtosisInstruction) (*ExecutionError, error) {
	executor.mutex.Lock()
	defer executor.mutex.Unlock()
	for _, instruction := range instructions {
		err := instruction.Execute(ctx)
		if err != nil {
			return &ExecutionError{err.Error()}, nil
		}
	}
	return nil, nil
}
