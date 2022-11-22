package define_recipe

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	kurtosis_backend_service "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/recipe_executor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	DefineFactBuiltinName = "define_recipe"

	serviceIdArgName = "service_id"
	factNameArgName  = "fact_name"
	recipeArgName    = "fact_recipe"
)

func GenerateDefineRecipeBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, recipeExecutor *recipe_executor.RecipeExecutor) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		factRecipe, interpretationError := parseStartosisArgs(b, args, kwargs)
		if interpretationError != nil {
			return nil, interpretationError
		}
		uuid := recipeExecutor.SaveRecipe(factRecipe)
		defineFactInstruction := NewDefineRecipeInstruction(*shared_helpers.GetCallerPositionFromThread(thread))
		*instructionsQueue = append(*instructionsQueue, defineFactInstruction)
		return starlark.String(fmt.Sprintf(shared_helpers.RecipeValueReplacementPlaceholderFormat, uuid)), nil
	}
}

type DefineRecipeInstruction struct {
	position   kurtosis_instruction.InstructionPosition
	serviceId  kurtosis_backend_service.ServiceID
	factRecipe *kurtosis_core_rpc_api_bindings.FactRecipe
}

func NewDefineRecipeInstruction(position kurtosis_instruction.InstructionPosition) *DefineRecipeInstruction {
	return &DefineRecipeInstruction{
		position: position,
	}
}

func (instruction *DefineRecipeInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return &instruction.position
}

func (instruction *DefineRecipeInstruction) GetCanonicalInstruction() string {
	return shared_helpers.MultiLineCanonicalizer.CanonicalizeInstruction(DefineFactBuiltinName, kurtosis_instruction.NoArgs, instruction.getKwargs(), &instruction.position)
}

func (instruction *DefineRecipeInstruction) Execute(ctx context.Context) (*string, error) {
	return nil, nil
}

func (instruction *DefineRecipeInstruction) String() string {
	return shared_helpers.SingleLineCanonicalizer.CanonicalizeInstruction(DefineFactBuiltinName, kurtosis_instruction.NoArgs, instruction.getKwargs(), &instruction.position)
}

func (instruction *DefineRecipeInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	return nil
}

func parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (*recipe_executor.HttpRequestRecipe, *startosis_errors.InterpretationError) {

	var recipeConfigArg *starlarkstruct.Struct

	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "http", &recipeConfigArg); err != nil {
		return nil, startosis_errors.NewInterpretationError(err.Error())
	}

	factRecipe, interpretationErr := kurtosis_instruction.ParseHttpRequestRecipe(recipeConfigArg)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	return factRecipe, nil
}

func (instruction *DefineRecipeInstruction) getKwargs() starlark.StringDict {
	return starlark.StringDict{}
}
