package assert

import (
	"context"
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
		runtimeValue, assertion, target, interpretationError := parseStartosisArgs(b, args, kwargs)
		if interpretationError != nil {
			return nil, interpretationError
		}
		assertInstruction := NewAssertInstruction(*shared_helpers.GetCallerPositionFromThread(thread), recipeExecutor, runtimeValue, assertion, target)
		*instructionsQueue = append(*instructionsQueue, assertInstruction)
		return starlark.None, nil
	}
}

type AssertInstruction struct {
	position       kurtosis_instruction.InstructionPosition
	recipeExecutor *runtime_value_store.RuntimeValueStore
	runtimeValue   string
	assertion      string
	target         starlark.Comparable
}

func NewAssertInstruction(position kurtosis_instruction.InstructionPosition, recipeExecutor *runtime_value_store.RuntimeValueStore, runtimeValue string, assertion string, target starlark.Comparable) *AssertInstruction {
	return &AssertInstruction{
		position:       position,
		recipeExecutor: recipeExecutor,
		runtimeValue:   runtimeValue,
		assertion:      assertion,
		target:         target,
	}
}

func (instruction *AssertInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return &instruction.position
}

func (instruction *AssertInstruction) GetCanonicalInstruction() string {
	return shared_helpers.CanonicalizeInstruction(AssertBuiltinName, kurtosis_instruction.NoArgs, instruction.getKwargs())
}

func (instruction *AssertInstruction) Execute(ctx context.Context) (*string, error) {
	currentValue, err := magic_string_helper.GetRuntimeValueFromString(instruction.runtimeValue, instruction.recipeExecutor)
	if err != nil {
		return nil, err
	}
	if currentValue.Type() != instruction.target.Type() {
		return nil, stacktrace.NewError("Assert failed because '%v' is type '%v' and '%v' is type '%v'", currentValue, currentValue.Type(), instruction.target, instruction.target.Type())
	}
	if comparisonToken, found := stringTokenToComparisonStarlarkToken[instruction.assertion]; found {
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
			return nil, stacktrace.NewError("Assertion failed, expected list but got '%v'", instruction.target)
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
	return nil, nil
}

func (instruction *AssertInstruction) String() string {
	return instruction.GetCanonicalInstruction()
}

func (instruction *AssertInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	return nil
}

func (instruction *AssertInstruction) getKwargs() starlark.StringDict {
	return starlark.StringDict{
		runtimeValueArgName: starlark.String(instruction.runtimeValue),
		assertionArgName:    starlark.String(instruction.assertion),
		targetArgName:       instruction.target,
	}
}

func parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (string, string, starlark.Comparable, *startosis_errors.InterpretationError) {

	var (
		runtimeValueArg starlark.String
		assertionArg    starlark.String
		targetArg       starlark.Comparable
	)

	if err := starlark.UnpackArgs(b.Name(), args, kwargs, runtimeValueArgName, &runtimeValueArg, assertionArgName, &assertionArg, targetArgName, &targetArg); err != nil {
		return "", "", nil, startosis_errors.NewInterpretationError(err.Error())
	}
	assertion := string(assertionArg)
	runtimeValue := string(runtimeValueArg)

	if _, found := stringTokenToComparisonStarlarkToken[assertion]; !found && assertionArg != "IN" && assertion != "NOT_IN" {
		return "", "", nil, startosis_errors.NewInterpretationError("'%v' is not a valid assertion", assertionArg)
	}
	if _, ok := targetArg.(*starlark.List); (assertionArg == "IN" || assertionArg == "NOT_IN") && !ok {
		return "", "", nil, startosis_errors.NewInterpretationError("'%v' assertion requires list, got '%v'", assertionArg, targetArg.Type())
	}

	return runtimeValue, assertion, targetArg, nil
}
