package startosis_engine

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/core/api/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/stacktrace"
)

type StartosisCompiler struct {
	enclaveId enclave.EnclaveID
}

func NewStartosisCompiler(enclaveId enclave.EnclaveID) *StartosisCompiler {
	// TODO(gb): build the bindings to populate an instruction list on compile
	return &StartosisCompiler{
		enclaveId: enclaveId,
	}
}

func (compiler *StartosisCompiler) Compile(script *kurtosis_core_rpc_api_bindings.SerializedStartosisScript) ([]kurtosis_instruction.KurtosisInstruction, error) {
	// TODO(gb): implement
	return nil, stacktrace.NewError("not implemented")
}
