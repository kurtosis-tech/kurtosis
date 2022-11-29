package assert

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers/magic_string_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

const (
	AssertBuiltinName = "assert"

	runtimeValueArgName = "value"
	assertionArgName    = "assertion"
	targetArgName       = "target_value"
)

var stringTokenToComparisonStarlarkToken = map[string]syntax.Token{
	"==": syntax.EQL,
	"!=": syntax.NEQ,
	">=": syntax.GE,
	">":  syntax.GT,
	"<=": syntax.LE,
	"<":  syntax.LT,
}

func GenerateAssertBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, recipeExecutor *runtime_value_store.RuntimeValueStore, serviceNetwork service_network.ServiceNetwork) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		instructionPosition := shared_helpers.GetCallerPositionFromThread(thread)
		instruction := NewEmptyAssertInstruction(instructionPosition, recipeExecutor)
		if interpretationError := instruction.parseStartosisArgs(b, args, kwargs); interpretationError != nil {
			return nil, interpretationError
		}
		*instructionsQueue = append(*instructionsQueue, instruction)
		return starlark.None, nil
	}
}

type AssertInstruction struct {
	position       *kurtosis_instruction.InstructionPosition
	starlarkKwargs starlark.StringDict

	recipeExecutor *runtime_value_store.RuntimeValueStore
	runtimeValue   string
	assertion      string
	target         starlark.Comparable
}

func NewAssertInstruction(position *kurtosis_instruction.InstructionPosition, recipeExecutor *runtime_value_store.RuntimeValueStore, runtimeValue string, assertion string, target starlark.Comparable, starlarkKwargs starlark.StringDict) *AssertInstruction {
	return &AssertInstruction{
		position:       position,
		recipeExecutor: recipeExecutor,
		runtimeValue:   runtimeValue,
		assertion:      assertion,
		target:         target,
		starlarkKwargs: starlarkKwargs,
	}
}

func NewEmptyAssertInstruction(position *kurtosis_instruction.InstructionPosition, recipeExecutor *runtime_value_store.RuntimeValueStore) *AssertInstruction {
	return &AssertInstruction{
		position:       position,
		recipeExecutor: recipeExecutor,
		runtimeValue:   "",
		assertion:      "",
		target:         nil,
		starlarkKwargs: nil,
	}
}

func (instruction *AssertInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return instruction.position
}

func (instruction *AssertInstruction) GetCanonicalInstruction() *kurtosis_core_rpc_api_bindings.StarlarkInstruction {
	args := []*kurtosis_core_rpc_api_bindings.StarlarkInstructionArg{
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[runtimeValueArgName]), runtimeValueArgName, kurtosis_instruction.Representative),
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[assertionArgName]), assertionArgName, kurtosis_instruction.Representative),
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[targetArgName]), targetArgName, kurtosis_instruction.Representative),
	}
	return binding_constructors.NewStarlarkInstruction(instruction.position.ToAPIType(), AssertBuiltinName, instruction.String(), args)
}

func (instruction *AssertInstruction) Execute(ctx context.Context) (*string, error) {
	currentValue, err := magic_string_helper.GetRuntimeValueFromString(instruction.runtimeValue, instruction.recipeExecutor)
	if err != nil {
		return nil, err
	}
	if comparisonToken, found := stringTokenToComparisonStarlarkToken[instruction.assertion]; found {
		if currentValue.Type() != instruction.target.Type() {
			return nil, stacktrace.NewError("Assert failed because '%v' is type '%v' and '%v' is type '%v'", currentValue, currentValue.Type(), instruction.target, instruction.target.Type())
		}
		result, err := currentValue.CompareSameType(comparisonToken, instruction.target, 1)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Assert comparison failed '%v' '%v' '%v'", currentValue, instruction.assertion, instruction.target)
		}
		if !result {
			return nil, stacktrace.NewError("Assertion failed '%v' '%v' '%v'", currentValue, instruction.assertion, instruction.target)
		}
	} else {
		listTarget, ok := instruction.target.(*starlark.List)
		if !ok {
			return nil, stacktrace.NewError("Assertion failed, expected list but got '%v'", instruction.target.Type())
		}
		inList := false
		for i := 0; i < listTarget.Len(); i++ {
			if listTarget.Index(i) == currentValue {
				inList = true
				break
			}
		}
		switch instruction.assertion {
		case "IN":
			if !inList {
				return nil, stacktrace.NewError("Assertion failed '%v' '%v' '%v'", currentValue, instruction.assertion, instruction.target)
			}
		case "NOT_IN":
			if inList {
				return nil, stacktrace.NewError("Assertion failed '%v' '%v' '%v'", currentValue, instruction.assertion, instruction.target)
			}
		}
	}
	instructionResult := fmt.Sprintf("Assertion succeeded. Value is '%s'.", currentValue)
	return &instructionResult, nil
}

func (instruction *AssertInstruction) String() string {
	return shared_helpers.CanonicalizeInstruction(AssertBuiltinName, kurtosis_instruction.NoArgs, instruction.starlarkKwargs)
}

func (instruction *AssertInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	return nil
}

func (instruction *AssertInstruction) parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) *startosis_errors.InterpretationError {

	var (
		runtimeValueArg starlark.String
		assertionArg    starlark.String
		targetArg       starlark.Comparable
	)

	if err := starlark.UnpackArgs(b.Name(), args, kwargs, runtimeValueArgName, &runtimeValueArg, assertionArgName, &assertionArg, targetArgName, &targetArg); err != nil {
		return startosis_errors.NewInterpretationError(err.Error())
	}

	instruction.assertion = string(assertionArg)
	instruction.runtimeValue = string(runtimeValueArg)
	instruction.target = targetArg

	instruction.starlarkKwargs = starlark.StringDict{
		runtimeValueArgName: starlark.String(instruction.runtimeValue),
		assertionArgName:    starlark.String(instruction.assertion),
		targetArgName:       instruction.target,
	}
	instruction.starlarkKwargs.Freeze()

	if _, found := stringTokenToComparisonStarlarkToken[instruction.assertion]; !found && instruction.assertion != "IN" && instruction.assertion != "NOT_IN" {
		return startosis_errors.NewInterpretationError("'%v' is not a valid assertion", assertionArg)
	}
	if _, ok := instruction.target.(*starlark.List); (instruction.assertion == "IN" || instruction.assertion == "NOT_IN") && !ok {
		return startosis_errors.NewInterpretationError("'%v' assertion requires list, got '%v'", assertionArg, targetArg.Type())
	}

	return nil
}
