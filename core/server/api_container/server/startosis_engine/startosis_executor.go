package startosis_engine

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/stacktrace"
)

type StartosisExecutor struct {
	enclaveId enclave.EnclaveID

	backend backend_interface.KurtosisBackend
}

func NewStartosisExecutor(enclaveId enclave.EnclaveID, backend backend_interface.KurtosisBackend) *StartosisExecutor {
	// TODO(gb): Implement the bindings to send instructions straight to the backend
	return &StartosisExecutor{
		enclaveId: enclaveId,
		backend:   backend,
	}
}

func (executor *StartosisExecutor) Execute(instructions []kurtosis_instruction.KurtosisInstruction) ([]kurtosis_instruction.KurtosisInstructionOutput, error) {
	// TODO(gb): implement
	return nil, stacktrace.NewError("not implemented")
}
