package extract

import (
	"context"
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
	DefineExtractBuiltinName = "extract"

	runtimeValueArgName   = "input"
	fieldExtractorArgName = "extractor"
)

func GenerateExtractInstructionBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, recipeExecutor *runtime_value_store.RuntimeValueStore, serviceNetwork service_network.ServiceNetwork) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		runtimeValueArg, fieldExtractorArg, interpretationError := parseStartosisArgs(b, args, kwargs)
		if interpretationError != nil {
			return nil, interpretationError
		}
		extractRecipe := recipe.NewExtractRecipe(string(fieldExtractorArg))
		resultUuid := recipeExecutor.CreateValue()
		returnValue := recipe.CreateStarlarkReturnValueFromExtractRuntimeValue(resultUuid)
		extractInstruction := NewExtractInstruction(serviceNetwork, *shared_helpers.GetCallerPositionFromThread(thread), recipeExecutor, resultUuid, runtimeValueArg, fieldExtractorArg, extractRecipe)
		*instructionsQueue = append(*instructionsQueue, extractInstruction)
		return returnValue, nil
	}
}

type ExtractInstruction struct {
	serviceNetwork service_network.ServiceNetwork

	position       kurtosis_instruction.InstructionPosition
	recipeExecutor *runtime_value_store.RuntimeValueStore
	resultUuid     string
	extractRecipe  *recipe.ExtractRecipe
	runtimeValue   string

	runtimeValueArg   starlark.String
	fieldExtractorArg starlark.String
}

func NewExtractInstruction(serviceNetwork service_network.ServiceNetwork, position kurtosis_instruction.InstructionPosition, recipeExecutor *runtime_value_store.RuntimeValueStore, resultUuid string, runtimeValueArg starlark.String, fieldExtractorArg starlark.String, extractRecipe *recipe.ExtractRecipe) *ExtractInstruction {
	return &ExtractInstruction{
		serviceNetwork:    serviceNetwork,
		position:          position,
		recipeExecutor:    recipeExecutor,
		runtimeValue:      string(runtimeValueArg),
		runtimeValueArg:   runtimeValueArg,
		fieldExtractorArg: fieldExtractorArg,
		extractRecipe:     extractRecipe,
		resultUuid:        resultUuid,
	}
}

func (instruction *ExtractInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return &instruction.position
}

func (instruction *ExtractInstruction) GetCanonicalInstruction() string {
	return shared_helpers.CanonicalizeInstruction(DefineExtractBuiltinName, kurtosis_instruction.NoArgs, instruction.getKwargs())
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
	return nil, err
}

func (instruction *ExtractInstruction) String() string {
	return instruction.GetCanonicalInstruction()
}

func (instruction *ExtractInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	return nil
}

func (instruction *ExtractInstruction) getKwargs() starlark.StringDict {
	return starlark.StringDict{
		runtimeValueArgName:   instruction.runtimeValueArg,
		fieldExtractorArgName: instruction.fieldExtractorArg,
	}
}

func parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.String, starlark.String, *startosis_errors.InterpretationError) {

	var (
		runtimeValueArg   starlark.String
		fieldExtractorArg starlark.String
	)

	if err := starlark.UnpackArgs(b.Name(), args, kwargs, runtimeValueArgName, &runtimeValueArg, fieldExtractorArgName, &fieldExtractorArg); err != nil {
		return "", "", startosis_errors.NewInterpretationError(err.Error())
	}

	return runtimeValueArg, fieldExtractorArg, nil
}
