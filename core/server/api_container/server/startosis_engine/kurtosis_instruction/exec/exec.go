package exec

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
)

const (
	ExecBuiltinName = "exec"

	recipeArgName = "recipe"
)

func GenerateExecBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, serviceNetwork service_network.ServiceNetwork, runtimeValueStore *runtime_value_store.RuntimeValueStore) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		instructionPosition := shared_helpers.GetCallerPositionFromThread(thread)
		execInstruction := newEmptyExecInstruction(serviceNetwork, instructionPosition, runtimeValueStore)
		if interpretationError := execInstruction.parseStartosisArgs(b, args, kwargs, runtimeValueStore); interpretationError != nil {
			return nil, interpretationError
		}
		resultUuid, err := runtimeValueStore.CreateValue()
		if err != nil {
			return nil, startosis_errors.NewInterpretationError("An error occurred while generating uuid for future reference for %v instruction", ExecBuiltinName)
		}
		execInstruction.resultUuid = resultUuid
		returnValue, interpretationErr := execInstruction.execRecipe.CreateStarlarkReturnValue(execInstruction.resultUuid)
		if interpretationErr != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "An error occurred while generating return value for %v instruction", ExecBuiltinName)
		}
		*instructionsQueue = append(*instructionsQueue, execInstruction)
		return returnValue, nil
	}
}

type ExecInstruction struct {
	serviceNetwork service_network.ServiceNetwork

	position       *kurtosis_instruction.InstructionPosition
	starlarkKwargs starlark.StringDict

	runtimeValueStore *runtime_value_store.RuntimeValueStore
	resultUuid        string
	execRecipe        *recipe.ExecRecipe
}

func newEmptyExecInstruction(serviceNetwork service_network.ServiceNetwork, position *kurtosis_instruction.InstructionPosition, runtimeValueStore *runtime_value_store.RuntimeValueStore) *ExecInstruction {
	return &ExecInstruction{
		serviceNetwork:    serviceNetwork,
		position:          position,
		starlarkKwargs:    starlark.StringDict{},
		resultUuid:        "",
		execRecipe:        nil,
		runtimeValueStore: runtimeValueStore,
	}
}

func (instruction *ExecInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return instruction.position
}

func (instruction *ExecInstruction) GetCanonicalInstruction() *kurtosis_core_rpc_api_bindings.StarlarkInstruction {
	args := []*kurtosis_core_rpc_api_bindings.StarlarkInstructionArg{
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[recipeArgName]), recipeArgName, kurtosis_instruction.Representative),
	}
	return binding_constructors.NewStarlarkInstruction(instruction.position.ToAPIType(), ExecBuiltinName, instruction.String(), args)
}

func (instruction *ExecInstruction) Execute(ctx context.Context) (*string, error) {
	result, err := instruction.execRecipe.Execute(ctx, instruction.serviceNetwork, instruction.runtimeValueStore)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error executing exec recipe")
	}
	instruction.runtimeValueStore.SetValue(instruction.resultUuid, result)
	instructionResult := instruction.execRecipe.ResultMapToString(result)
	return &instructionResult, err
}

func (instruction *ExecInstruction) String() string {
	return shared_helpers.CanonicalizeInstruction(ExecBuiltinName, kurtosis_instruction.NoArgs, instruction.starlarkKwargs)
}

func (instruction *ExecInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	// TODO: validate recipe
	return nil
}

func (instruction *ExecInstruction) parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple, runtimeValueStore *runtime_value_store.RuntimeValueStore) *startosis_errors.InterpretationError {
	var recipeConfigExecRecipe *recipe.ExecRecipe

	if err := starlark.UnpackArgs(b.Name(), args, kwargs, recipeArgName, &recipeConfigExecRecipe); err != nil {
		return startosis_errors.NewInterpretationError(fmt.Sprintf("Error occurred while parsing recipe: %v", err.Error()))
	}

	instruction.starlarkKwargs = starlark.StringDict{
		recipeArgName: recipeConfigExecRecipe,
	}
	instruction.starlarkKwargs.Freeze()
	instruction.execRecipe = recipeConfigExecRecipe
	return nil
}
