package kurtosis_plan_instruction

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"go.starlark.net/starlark"
)

type KurtosisPlanInstructionInternal struct {
	*kurtosis_starlark_framework.KurtosisBaseBuiltinInternal

	capabilities KurtosisPlanInstructionCapabilities

	defaultDisplayArguments map[string]bool
}

func newKurtosisPlanInstructionInternal(internalBuiltin *kurtosis_starlark_framework.KurtosisBaseBuiltinInternal, capabilities KurtosisPlanInstructionCapabilities, defaultDisplayArguments map[string]bool) *KurtosisPlanInstructionInternal {
	return &KurtosisPlanInstructionInternal{
		KurtosisBaseBuiltinInternal: internalBuiltin,

		capabilities: capabilities,

		defaultDisplayArguments: defaultDisplayArguments,
	}
}

func (builtin *KurtosisPlanInstructionInternal) GetCanonicalInstruction() *kurtosis_core_rpc_api_bindings.StarlarkInstruction {
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
		args[idx] = binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(value), name, shouldBeDisplayed)
	}
	return binding_constructors.NewStarlarkInstruction(builtin.GetPosition().ToAPIType(), builtin.GetName(), builtin.String(), args)
}

// GetPositionInOriginalScript is here to implement the KurtosisInstruction interface. Remove it when it's not needed anymore
func (builtin *KurtosisPlanInstructionInternal) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	position := builtin.GetPosition().ToAPIType()
	return kurtosis_instruction.NewInstructionPosition(position.GetLine(), position.GetColumn(), position.GetFilename())
}

// ValidateAndUpdateEnvironment is here to ease transition to the new framework and to implement the KurtosisInstruction interface.
// Remove it when it's not needed anymore
func (builtin *KurtosisPlanInstructionInternal) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	validationErr := builtin.Validate(environment)
	if validationErr != nil {
		return validationErr
	}
	return nil
}

func (builtin *KurtosisPlanInstructionInternal) Validate(validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	return builtin.capabilities.Validate(builtin.GetArguments(), validatorEnvironment)
}

func (builtin *KurtosisPlanInstructionInternal) Execute(ctx context.Context) (*string, error) {
	result, err := builtin.capabilities.Execute(ctx, builtin.GetArguments())
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (builtin *KurtosisPlanInstructionInternal) interpret() (starlark.Value, *startosis_errors.InterpretationError) {
	result, interpretationErr := builtin.capabilities.Interpret(builtin.GetArguments())
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return result, nil
}
