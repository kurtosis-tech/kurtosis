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
)

const (
	AssertBuiltinName = "assert"

	runtimeValueArgName = "value"
	assertionArgName    = "assertion"
	stringTargetArgName = "value"
)

func GenerateAssertBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, recipeExecutor *recipe_executor.RecipeExecutor, serviceNetwork service_network.ServiceNetwork) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		runtimeValue, assertion, stringTarget, interpretationError := parseStartosisArgs(b, args, kwargs)
		if interpretationError != nil {
			return nil, interpretationError
		}
		assertInstruction := NewAssertInstruction(*shared_helpers.GetCallerPositionFromThread(thread), recipeExecutor, runtimeValue, assertion, stringTarget)
		*instructionsQueue = append(*instructionsQueue, assertInstruction)
		return starlark.None, nil
	}
}

type AssertInstruction struct {
	position       kurtosis_instruction.InstructionPosition
	recipeExecutor *recipe_executor.RecipeExecutor
	runtimeValue   starlark.String
	assertion      starlark.String
	stringTarget   starlark.String
}

func NewAssertInstruction(position kurtosis_instruction.InstructionPosition, recipeExecutor *recipe_executor.RecipeExecutor, runtimeValue starlark.String, assertion starlark.String, stringTarget starlark.String) *AssertInstruction {
	return &AssertInstruction{
		position:       position,
		recipeExecutor: recipeExecutor,
		runtimeValue:   runtimeValue,
		assertion:      assertion,
		stringTarget:   stringTarget,
	}
}

func (instruction *AssertInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return &instruction.position
}

func (instruction *AssertInstruction) GetCanonicalInstruction() string {
	return shared_helpers.MultiLineCanonicalizer.CanonicalizeInstruction(AssertBuiltinName, kurtosis_instruction.NoArgs, instruction.getKwargs(), &instruction.position)
}

func (instruction *AssertInstruction) Execute(ctx context.Context) (*string, error) {
	currentValue, err := shared_helpers.ReplaceRuntimeValueInString(instruction.runtimeValue.String(), instruction.recipeExecutor)
	if err != nil {
		return nil, err
	}
	switch instruction.assertion {
	case "==":
		if !(currentValue == instruction.stringTarget.String()) {
			return nil, stacktrace.NewError("Assertion failed on %v %v %v", currentValue, instruction.assertion, instruction.stringTarget)
		}
	case "!=":
		if !(currentValue != instruction.stringTarget.String()) {
			return nil, stacktrace.NewError("Assertion failed on %v %v %v", currentValue, instruction.assertion, instruction.stringTarget)
		}
	}
	return nil, err
}

func (instruction *AssertInstruction) String() string {
	return shared_helpers.SingleLineCanonicalizer.CanonicalizeInstruction(AssertBuiltinName, kurtosis_instruction.NoArgs, instruction.getKwargs(), &instruction.position)
}

func (instruction *AssertInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	return nil
}

func (instruction *AssertInstruction) getKwargs() starlark.StringDict {
	return starlark.StringDict{}
}

func parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.String, starlark.String, starlark.String, *startosis_errors.InterpretationError) {

	var (
		runtimeValueArg starlark.String
		assertionArg    starlark.String
		stringTarget    starlark.String
	)

	if err := starlark.UnpackArgs(b.Name(), args, kwargs, runtimeValueArgName, &runtimeValueArg, assertionArgName, &assertionArg, stringTargetArgName, &stringTarget); err != nil {
		return "", "", "", startosis_errors.NewInterpretationError(err.Error())
	}

	return runtimeValueArg, assertionArg, stringTarget, nil
}
