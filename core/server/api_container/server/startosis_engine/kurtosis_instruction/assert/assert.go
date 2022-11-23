package assert

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/recipe_executor"
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
	targetArgName       = "value"
)

var stringTokenToStarlarkToken = map[string]syntax.Token{
	"==": syntax.EQ,
	"!=": syntax.NEQ,
	">=": syntax.GE,
	">":  syntax.GT,
	"<=": syntax.LE,
	"<":  syntax.LT,
}

func GenerateAssertBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, recipeExecutor *recipe_executor.RecipeExecutor, serviceNetwork service_network.ServiceNetwork) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
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
	recipeExecutor *recipe_executor.RecipeExecutor
	runtimeValue   starlark.String
	assertion      starlark.String
	target         starlark.Comparable
}

func NewAssertInstruction(position kurtosis_instruction.InstructionPosition, recipeExecutor *recipe_executor.RecipeExecutor, runtimeValue starlark.String, assertion starlark.String, target starlark.Comparable) *AssertInstruction {
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
	currentValue, err := shared_helpers.GetRuntimeValueFromString(instruction.runtimeValue.String(), instruction.recipeExecutor)
	if err != nil {
		return nil, err
	}
	result, err := currentValue.CompareSameType(stringTokenToStarlarkToken[string(instruction.assertion)], instruction.target, 1)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Assert comparison failed '%v' '%v' '%v'", currentValue, instruction.assertion, instruction.target)
	}
	if !result {
		return nil, stacktrace.Propagate(err, "Assertion failed '%v' '%v' '%v'", currentValue, instruction.assertion, instruction.target)
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
	return starlark.StringDict{}
}

func parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.String, starlark.String, starlark.Comparable, *startosis_errors.InterpretationError) {

	var (
		runtimeValueArg starlark.String
		assertionArg    starlark.String
		target          starlark.Comparable
	)

	if err := starlark.UnpackArgs(b.Name(), args, kwargs, runtimeValueArgName, &runtimeValueArg, assertionArgName, &assertionArg, targetArgName, &target); err != nil {
		return "", "", nil, startosis_errors.NewInterpretationError(err.Error())
	}

	return runtimeValueArg, assertionArg, target, nil
}
