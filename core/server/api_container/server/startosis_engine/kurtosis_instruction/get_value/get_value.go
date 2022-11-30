package get_value

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/recipe"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
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

func GenerateGetValueBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, recipeExecutor *runtime_value_store.RuntimeValueStore, serviceNetwork service_network.ServiceNetwork) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		position := shared_helpers.GetCallerPositionFromThread(thread)
		instruction := newEmptyGetValueInstruction(serviceNetwork, position, recipeExecutor)
		if interpretationError := instruction.parseStartosisArgs(b, args, kwargs); interpretationError != nil {
			return nil, interpretationError
		}
		instruction.resultUuid = recipeExecutor.CreateValue()
		returnValue := recipe.CreateStarlarkReturnValueFromHttpRequestRecipe(instruction.resultUuid)
		*instructionsQueue = append(*instructionsQueue, instruction)
		return returnValue, nil
	}
}

type GetValueInstruction struct {
	serviceNetwork service_network.ServiceNetwork

	position       *kurtosis_instruction.InstructionPosition
	starlarkKwargs starlark.StringDict

	recipeExecutor    *runtime_value_store.RuntimeValueStore
	httpRequestRecipe *recipe.HttpRequestRecipe
	recipeConfigArg   *starlarkstruct.Struct
	resultUuid        string
}

func NewGetValueInstruction(serviceNetwork service_network.ServiceNetwork, position *kurtosis_instruction.InstructionPosition, recipeExecutor *runtime_value_store.RuntimeValueStore, httpRequestRecipe *recipe.HttpRequestRecipe, recipeConfigArg *starlarkstruct.Struct, resultUuid string, starlarkKwargs starlark.StringDict) *GetValueInstruction {
	return &GetValueInstruction{
		serviceNetwork:    serviceNetwork,
		position:          position,
		recipeExecutor:    recipeExecutor,
		httpRequestRecipe: httpRequestRecipe,
		recipeConfigArg:   recipeConfigArg,
		resultUuid:        resultUuid,
		starlarkKwargs:    starlarkKwargs,
	}
}

func newEmptyGetValueInstruction(serviceNetwork service_network.ServiceNetwork, position *kurtosis_instruction.InstructionPosition, recipeExecutor *runtime_value_store.RuntimeValueStore) *GetValueInstruction {
	return &GetValueInstruction{
		serviceNetwork:    serviceNetwork,
		position:          position,
		recipeExecutor:    recipeExecutor,
		httpRequestRecipe: nil,
		recipeConfigArg:   nil,
		resultUuid:        "",
		starlarkKwargs:    nil,
	}
}

func (instruction *GetValueInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return instruction.position
}

func (instruction *GetValueInstruction) GetCanonicalInstruction() *kurtosis_core_rpc_api_bindings.StarlarkInstruction {
	args := []*kurtosis_core_rpc_api_bindings.StarlarkInstructionArg{
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[recipeArgName]), recipeArgName, kurtosis_instruction.Representative),
	}
	return binding_constructors.NewStarlarkInstruction(instruction.position.ToAPIType(), GetValueBuiltinName, instruction.String(), args)

}

func (instruction *GetValueInstruction) Execute(ctx context.Context) (*string, error) {
	result, err := instruction.httpRequestRecipe.Execute(ctx, instruction.serviceNetwork)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error executing http recipe")
	}
	instruction.recipeExecutor.SetValue(instruction.resultUuid, result)
	instructionResult := fmt.Sprintf("Value obtained with status code '%d'", result[recipe.StatusCodeKey])
	return &instructionResult, err
}

func (instruction *GetValueInstruction) String() string {
	return shared_helpers.CanonicalizeInstruction(GetValueBuiltinName, kurtosis_instruction.NoArgs, instruction.starlarkKwargs)
}

func (instruction *GetValueInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	return nil
}

func (instruction *GetValueInstruction) parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) *startosis_errors.InterpretationError {

	var recipeConfigArg *starlarkstruct.Struct

	if err := starlark.UnpackArgs(b.Name(), args, kwargs, recipeArgName, &recipeConfigArg); err != nil {
		return startosis_errors.NewInterpretationError(err.Error())
	}
	instruction.starlarkKwargs = starlark.StringDict{
		"recipe": recipeConfigArg,
	}
	instruction.starlarkKwargs.Freeze()

	var err *startosis_errors.InterpretationError
	instruction.httpRequestRecipe, err = kurtosis_instruction.ParseHttpRequestRecipe(recipeConfigArg)
	if err != nil {
		return err
	}
	return nil
}
