package get_value

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/recipe"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/recipe_executor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	GetValueBuiltinName = "get_value"

	recipeArgName = "recipe"
)

func GenerateGetValueBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, recipeExecutor *recipe_executor.RuntimeValueStore, serviceNetwork service_network.ServiceNetwork) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		recipeConfig, interpretationError := parseStartosisArgs(b, args, kwargs)
		if interpretationError != nil {
			return nil, interpretationError
		}
		httpRequestRecipe, interpretationErr := kurtosis_instruction.ParseHttpRequestRecipe(recipeConfig)
		if interpretationErr != nil {
			return nil, interpretationErr
		}
		resultUuid := recipeExecutor.CreateValue()
		returnValue := recipe.CreateStarlarkReturnValueFromHttpRequestRecipe(resultUuid)
		getValueInstruction := NewGetValueInstruction(serviceNetwork, *shared_helpers.GetCallerPositionFromThread(thread), recipeExecutor, httpRequestRecipe, resultUuid)
		*instructionsQueue = append(*instructionsQueue, getValueInstruction)
		return returnValue, nil
	}
}

type GetValueInstruction struct {
	serviceNetwork service_network.ServiceNetwork

	position          kurtosis_instruction.InstructionPosition
	recipeExecutor    *recipe_executor.RuntimeValueStore
	httpRequestRecipe *recipe.HttpRequestRecipe
	resultUuid        string
}

func NewGetValueInstruction(serviceNetwork service_network.ServiceNetwork, position kurtosis_instruction.InstructionPosition, recipeExecutor *recipe_executor.RuntimeValueStore, httpRequestRecipe *recipe.HttpRequestRecipe, resultUuid string) *GetValueInstruction {
	return &GetValueInstruction{
		serviceNetwork:    serviceNetwork,
		position:          position,
		recipeExecutor:    recipeExecutor,
		httpRequestRecipe: httpRequestRecipe,
		resultUuid:        resultUuid,
	}
}

func (instruction *GetValueInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return &instruction.position
}

func (instruction *GetValueInstruction) GetCanonicalInstruction() string {
	return shared_helpers.CanonicalizeInstruction(GetValueBuiltinName, kurtosis_instruction.NoArgs, instruction.getKwargs())
}

func (instruction *GetValueInstruction) Execute(ctx context.Context) (*string, error) {
	result, err := instruction.httpRequestRecipe.Execute(ctx, instruction.serviceNetwork)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error executing http recipe")
	}
	instruction.recipeExecutor.SetValue(instruction.resultUuid, result)
	return nil, err
}

func (instruction *GetValueInstruction) String() string {
	return instruction.GetCanonicalInstruction()
}

func (instruction *GetValueInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	return nil
}

func (instruction *GetValueInstruction) getKwargs() starlark.StringDict {
	return starlark.StringDict{}
}

func parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (*starlarkstruct.Struct, *startosis_errors.InterpretationError) {

	var recipeConfigArg *starlarkstruct.Struct

	if err := starlark.UnpackArgs(b.Name(), args, kwargs, recipeArgName, &recipeConfigArg); err != nil {
		return nil, startosis_errors.NewInterpretationError(err.Error())
	}

	return recipeConfigArg, nil
}
