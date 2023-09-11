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

	// GetPersistableAttributes returns an EnclavePlanInstruction object which contains all the persistable attributes
	// for this instruction. Persistable attributes are what will be written to the enclave database so that even
	// if the APIC is restarted, idempotent runs will continue to work.
	GetPersistableAttributes() *enclave_plan_persistence.EnclavePlanInstructionBuilder
}
