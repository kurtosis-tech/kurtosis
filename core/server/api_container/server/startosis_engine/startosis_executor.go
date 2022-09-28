package startosis_engine

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/stacktrace"
)

type StartosisExecutor struct {
}

type ExecutionError struct {
	Error string
}

func NewStartosisExecutor() *StartosisExecutor {
	// TODO(gb): Implement the bindings to send instructions straight to the backend
	return &StartosisExecutor{}
}

// Execute executes the list of Kurtosis instructions against the Kurtosis backend
// It returns a potential execution error if something went wrong.
// It returns a error if something unexpected happens outside the execution of the script
func (executor *StartosisExecutor) Execute(ctx context.Context, instructions []kurtosis_instruction.KurtosisInstruction) (*ExecutionError, error) {
	// TODO(gb): implement
	return nil, stacktrace.NewError("not implemented")
}
