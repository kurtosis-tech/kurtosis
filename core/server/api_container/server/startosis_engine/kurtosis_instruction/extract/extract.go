package extract

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers/magic_string_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/recipe"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
)

const (
	ExtractBuiltinName = "extract"

	runtimeValueArgName   = "input"
	fieldExtractorArgName = "extractor"
)

func GenerateExtractInstructionBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, recipeExecutor *runtime_value_store.RuntimeValueStore, serviceNetwork service_network.ServiceNetwork) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		instructionPosition := shared_helpers.GetCallerPositionFromThread(thread)
		instruction := newEmptyExtractInstruction(serviceNetwork, instructionPosition, recipeExecutor)
		if interpretationError := instruction.parseStartosisArgs(b, args, kwargs); interpretationError != nil {
			return nil, interpretationError
		}
		instruction.resultUuid = recipeExecutor.CreateValue()
		returnValue := recipe.CreateStarlarkReturnValueFromExtractRuntimeValue(instruction.resultUuid)
		*instructionsQueue = append(*instructionsQueue, instruction)
		return returnValue, nil
	}
}

type ExtractInstruction struct {
	serviceNetwork service_network.ServiceNetwork
	recipeExecutor *runtime_value_store.RuntimeValueStore

	position       *kurtosis_instruction.InstructionPosition
	starlarkKwargs starlark.StringDict

	resultUuid    string
	extractRecipe *recipe.ExtractRecipe
	runtimeValue  string
}

func NewExtractInstruction(serviceNetwork service_network.ServiceNetwork, position *kurtosis_instruction.InstructionPosition, recipeExecutor *runtime_value_store.RuntimeValueStore, resultUuid string, runtimeValue string, extractRecipe *recipe.ExtractRecipe, starlarkKwargs starlark.StringDict) *ExtractInstruction {
	return &ExtractInstruction{
		serviceNetwork: serviceNetwork,
		position:       position,
		recipeExecutor: recipeExecutor,
		starlarkKwargs: starlarkKwargs,
		extractRecipe:  extractRecipe,
		runtimeValue:   runtimeValue,
		resultUuid:     resultUuid,
	}
}

func newEmptyExtractInstruction(serviceNetwork service_network.ServiceNetwork, position *kurtosis_instruction.InstructionPosition, recipeExecutor *runtime_value_store.RuntimeValueStore) *ExtractInstruction {
	return &ExtractInstruction{
		serviceNetwork: serviceNetwork,
		position:       position,
		recipeExecutor: recipeExecutor,
		starlarkKwargs: nil,
		extractRecipe:  nil,
		runtimeValue:   "",
		resultUuid:     "",
	}
}

func (instruction *ExtractInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return instruction.position
}

func (instruction *ExtractInstruction) GetCanonicalInstruction() *kurtosis_core_rpc_api_bindings.StarlarkInstruction {
	args := []*kurtosis_core_rpc_api_bindings.StarlarkInstructionArg{
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[runtimeValueArgName]), runtimeValueArgName, kurtosis_instruction.Representative),
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[fieldExtractorArgName]), fieldExtractorArgName, kurtosis_instruction.NotRepresentative),
	}
	return binding_constructors.NewStarlarkInstruction(instruction.position.ToAPIType(), ExtractBuiltinName, instruction.String(), args)
}

func (instruction *ExtractInstruction) Execute(ctx context.Context) (*string, error) {
	runtimeValueCurrent, err := magic_string_helper.GetRuntimeValueFromString(instruction.runtimeValue, instruction.recipeExecutor)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error fetching value")
	}
	castedRuntimeValue, ok := runtimeValueCurrent.(starlark.String)
	if !ok {
		return nil, stacktrace.Propagate(err, "Only string values are supported, got '%v'", runtimeValueCurrent)
	}
	result, err := instruction.extractRecipe.Execute(castedRuntimeValue.GoString())
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error executing extract recipe")
	}
	instruction.recipeExecutor.SetValue(instruction.resultUuid, result)
	instructionResult := fmt.Sprintf("Value '%s' extracted", castedRuntimeValue.GoString())
	return &instructionResult, err
}

func (instruction *ExtractInstruction) String() string {
	return shared_helpers.CanonicalizeInstruction(ExtractBuiltinName, kurtosis_instruction.NoArgs, instruction.starlarkKwargs)
}

func (instruction *ExtractInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	return nil
}

func (instruction *ExtractInstruction) parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) *startosis_errors.InterpretationError {

	var (
		runtimeValueArg   starlark.String
		fieldExtractorArg starlark.String
	)

	if err := starlark.UnpackArgs(b.Name(), args, kwargs, runtimeValueArgName, &runtimeValueArg, fieldExtractorArgName, &fieldExtractorArg); err != nil {
		return startosis_errors.NewInterpretationError(err.Error())
	}

	instruction.starlarkKwargs = starlark.StringDict{
		runtimeValueArgName:   runtimeValueArg,
		fieldExtractorArgName: fieldExtractorArg,
	}
	instruction.starlarkKwargs.Freeze()

	instruction.runtimeValue = string(runtimeValueArg)
	instruction.extractRecipe = recipe.NewExtractRecipe(string(fieldExtractorArg))
	return nil
}
