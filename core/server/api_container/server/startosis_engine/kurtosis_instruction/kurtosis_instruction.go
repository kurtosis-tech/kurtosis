package kurtosis_instruction

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_plan_persistence"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
)

type KurtosisInstruction interface {
	GetPositionInOriginalScript() *kurtosis_starlark_framework.KurtosisBuiltinPosition

	GetCanonicalInstruction(isSkipped bool) *kurtosis_core_rpc_api_bindings.StarlarkInstruction

	Execute(ctx context.Context) (*string, error)

	// String is only for easy printing in logs and error messages.
	// Most of the time it will just call GetCanonicalInstruction()
	String() string

	// ValidateAndUpdateEnvironment validates if the instruction can be applied to an environment, and mutates that
	// environment to reflect how Kurtosis would look like after this instruction is successfully executed.
	ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error

	// TryResolveWith assesses whether the instruction can be resolved with the one passed as an argument.
	TryResolveWith(other *enclave_plan_persistence.EnclavePlanInstruction, enclaveComponents *enclave_structure.EnclaveComponents) enclave_structure.InstructionResolutionStatus

	// GetPersistableAttributes returns the instruction attributes that needs to be persisted in the enclave plan
	// to power idempotent runs. This right now is composed of:
	// - The instruction type (i.e. add_service, exec, wait, assert, etc)
	// - The starlark code of the instruction to be able to check for instruction equality
	// - The list of service names this instruction affects (adds, removes, updates)
	// - The list of files artifact names this instruction affects (adds, removes, updates)
	// - Optionally the list of files artifact MD5 this instruction affects, corresponding to the files artifacts names, in order
	GetPersistableAttributes() (string, string, []string, []string, [][]byte)
}
