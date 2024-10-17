package kurtosis_plan_instruction

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_plan_persistence"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/plan_yaml"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"go.starlark.net/starlark"
)

type kurtosisPlanInstructionInternal struct {
	*kurtosis_starlark_framework.KurtosisBaseBuiltinInternal

	capabilities KurtosisPlanInstructionCapabilities

	defaultDisplayArguments map[string]bool
}

func newKurtosisPlanInstructionInternal(internalBuiltin *kurtosis_starlark_framework.KurtosisBaseBuiltinInternal, capabilities KurtosisPlanInstructionCapabilities, defaultDisplayArguments map[string]bool) *kurtosisPlanInstructionInternal {
	return &kurtosisPlanInstructionInternal{
		KurtosisBaseBuiltinInternal: internalBuiltin,

		capabilities: capabilities,

		defaultDisplayArguments: defaultDisplayArguments,
	}
}

func (builtin *kurtosisPlanInstructionInternal) GetCanonicalInstruction(isSkipped bool) *kurtosis_core_rpc_api_bindings.StarlarkInstruction {
	args := make([]*kurtosis_core_rpc_api_bindings.StarlarkInstructionArg, len(builtin.GetArguments().GetDefinition()))
	for idx, argument := range builtin.GetArguments().GetDefinition() {
		name := argument.Name
		value, err := builtin_argument.ExtractArgumentValue[starlark.Value](builtin.GetArguments(), name)
		if err != nil {
			// should never happen
			continue
		}
		shouldBeDisplayed := false
		if _, found := builtin.defaultDisplayArguments[name]; found {
			shouldBeDisplayed = true
		}
		args[idx] = binding_constructors.NewStarlarkInstructionKwarg(builtin_argument.StringifyArgumentValue(value), name, shouldBeDisplayed)
	}
	return binding_constructors.NewStarlarkInstruction(builtin.GetPosition().ToAPIType(), builtin.GetName(), builtin.String(), args, isSkipped, builtin.capabilities.Description())
}

// GetPositionInOriginalScript is here to implement the KurtosisInstruction interface. Remove it when it's not needed anymore
func (builtin *kurtosisPlanInstructionInternal) GetPositionInOriginalScript() *kurtosis_starlark_framework.KurtosisBuiltinPosition {
	position := builtin.GetPosition().ToAPIType()
	return kurtosis_starlark_framework.NewKurtosisBuiltinPosition(position.GetFilename(), position.GetLine(), position.GetColumn())
}

// ValidateAndUpdateEnvironment is here to ease transition to the new framework and to implement the KurtosisInstruction interface.
// Remove it when it's not needed anymore
func (builtin *kurtosisPlanInstructionInternal) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	validationErr := builtin.Validate(environment)
	if validationErr != nil {
		return validationErr
	}
	return nil
}

func (builtin *kurtosisPlanInstructionInternal) Validate(validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	return builtin.capabilities.Validate(builtin.GetArguments(), validatorEnvironment)
}

func (builtin *kurtosisPlanInstructionInternal) Execute(ctx context.Context) (*string, error) {
	result, err := builtin.capabilities.Execute(ctx, builtin.GetArguments())
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (builtin *kurtosisPlanInstructionInternal) TryResolveWith(other *enclave_plan_persistence.EnclavePlanInstruction, enclaveComponents *enclave_structure.EnclaveComponents) enclave_structure.InstructionResolutionStatus {
	isAnAbortAllInstruction := builtin.capabilities.TryResolveWith(false, nil, enclaveComponents) == enclave_structure.InstructionIsNotResolvableAbort
	if isAnAbortAllInstruction {
		return enclave_structure.InstructionIsNotResolvableAbort
	}

	if other == nil {
		return enclave_structure.InstructionIsUnknown
	}

	instructionsAreEqual := builtin.String() == other.StarlarkCode
	return builtin.capabilities.TryResolveWith(instructionsAreEqual, other, enclaveComponents)
}

func (builtin *kurtosisPlanInstructionInternal) GetPersistableAttributes() *enclave_plan_persistence.EnclavePlanInstructionBuilder {
	enclavePlaneInstructionBuilder := enclave_plan_persistence.NewEnclavePlanInstructionBuilder()
	builtin.capabilities.FillPersistableAttributes(enclavePlaneInstructionBuilder)
	return enclavePlaneInstructionBuilder.SetStarlarkCode(builtin.String())
}

func (builtin *kurtosisPlanInstructionInternal) UpdatePlan(plan *plan_yaml.PlanYamlGenerator) error {
	return builtin.capabilities.UpdatePlan(plan)
}

func (builtin *kurtosisPlanInstructionInternal) interpret() (starlark.Value, *startosis_errors.InterpretationError) {
	result, interpretationErr := builtin.capabilities.Interpret(builtin.GetPosition().GetFilename(), builtin.GetArguments())
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return result, nil
}
