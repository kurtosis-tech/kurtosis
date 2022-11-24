package extract

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers/magic_string_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/recipe"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/recipe_executor"
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

func GenerateExtractInstructionBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, recipeExecutor *recipe_executor.RuntimeValueStore, serviceNetwork service_network.ServiceNetwork) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		runtimeValue, fieldExtractor, interpretationError := parseStartosisArgs(b, args, kwargs)
		if interpretationError != nil {
			return nil, interpretationError
		}
		extractRecipe := recipe.NewExtractRecipe(fieldExtractor)
		resultUuid := recipeExecutor.CreateValue()
		returnValue := recipe.CreateStarlarkReturnValueFromExtractRuntimeValue(resultUuid)
		extractInstruction := NewExtractInstruction(serviceNetwork, *shared_helpers.GetCallerPositionFromThread(thread), recipeExecutor, resultUuid, runtimeValue, extractRecipe)
		*instructionsQueue = append(*instructionsQueue, extractInstruction)
		return returnValue, nil
	}
}

type ExtractInstruction struct {
	serviceNetwork service_network.ServiceNetwork

	position       kurtosis_instruction.InstructionPosition
	recipeExecutor *recipe_executor.RuntimeValueStore
	resultUuid     string
	extractRecipe  *recipe.ExtractRecipe
	runtimeValue   string
}

func NewExtractInstruction(serviceNetwork service_network.ServiceNetwork, position kurtosis_instruction.InstructionPosition, recipeExecutor *recipe_executor.RuntimeValueStore, resultUuid string, runtimeValue string, extractRecipe *recipe.ExtractRecipe) *ExtractInstruction {
	return &ExtractInstruction{
		serviceNetwork: serviceNetwork,
		position:       position,
		recipeExecutor: recipeExecutor,
		runtimeValue:   runtimeValue,
		extractRecipe:  extractRecipe,
		resultUuid:     resultUuid,
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
	if _, ok := runtimeValueCurrent.(starlark.String); !ok {
		return nil, stacktrace.Propagate(err, "Only string values are supported, got '%v'", runtimeValueCurrent)
	}
	result, err := instruction.extractRecipe.Execute(runtimeValueCurrent.(starlark.String).GoString())
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error executing get_value")
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
	return starlark.StringDict{}
}

func parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (string, string, *startosis_errors.InterpretationError) {

	var (
		runtimeValueArg   starlark.String
		fieldExtractorArg starlark.String
	)

	if err := starlark.UnpackArgs(b.Name(), args, kwargs, runtimeValueArgName, &runtimeValueArg, fieldExtractorArgName, &fieldExtractorArg); err != nil {
		return "", "", startosis_errors.NewInterpretationError(err.Error())
	}

	return string(runtimeValueArg), string(fieldExtractorArg), nil
}
