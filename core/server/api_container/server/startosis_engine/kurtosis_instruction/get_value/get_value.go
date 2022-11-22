package get_value

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/recipe_executor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	DefineFactBuiltinName = "get_value"

	recipeArgName = "recipe"
)

func GenerateGetValueBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, recipeExecutor *recipe_executor.RecipeExecutor, serviceNetwork service_network.ServiceNetwork) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
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
		resultUuid := recipeExecutor.CreateValue(httpRequestRecipe)
		returnValue := recipe_executor.CreateStarlarkDictFromHttpRequestRuntimeValue(starlark.String(fmt.Sprintf(shared_helpers.RuntimeValueReplacementPlaceholderFormat, resultUuid, "body")), starlark.String(fmt.Sprintf(shared_helpers.RuntimeValueReplacementPlaceholderFormat, resultUuid, "code")))
		getValueInstruction := NewGetValueInstruction(serviceNetwork, *shared_helpers.GetCallerPositionFromThread(thread), recipeExecutor, resultUuid)
		*instructionsQueue = append(*instructionsQueue, getValueInstruction)
		return returnValue, nil
	}
}

type GetValueInstruction struct {
	serviceNetwork service_network.ServiceNetwork

	position       kurtosis_instruction.InstructionPosition
	recipeExecutor *recipe_executor.RecipeExecutor
	resultUuid     string
}

func NewGetValueInstruction(serviceNetwork service_network.ServiceNetwork, position kurtosis_instruction.InstructionPosition, recipeExecutor *recipe_executor.RecipeExecutor, resultUuid string) *GetValueInstruction {
	return &GetValueInstruction{
		serviceNetwork: serviceNetwork,
		position:       position,
		recipeExecutor: recipeExecutor,
		resultUuid:     resultUuid,
	}
}

func (instruction *GetValueInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return &instruction.position
}

func (instruction *GetValueInstruction) GetCanonicalInstruction() string {
	return shared_helpers.MultiLineCanonicalizer.CanonicalizeInstruction(DefineFactBuiltinName, kurtosis_instruction.NoArgs, instruction.getKwargs(), &instruction.position)
}

func (instruction *GetValueInstruction) Execute(ctx context.Context) (*string, error) {
	err := instruction.recipeExecutor.ExecuteValue(ctx, instruction.serviceNetwork, instruction.resultUuid)
	return nil, err
}

func (instruction *GetValueInstruction) String() string {
	return shared_helpers.SingleLineCanonicalizer.CanonicalizeInstruction(DefineFactBuiltinName, kurtosis_instruction.NoArgs, instruction.getKwargs(), &instruction.position)
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
