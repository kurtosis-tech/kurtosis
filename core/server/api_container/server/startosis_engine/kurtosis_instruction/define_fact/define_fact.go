package define_fact

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	kurtosis_backend_service "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/facts_engine"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	DefineFactBuiltinName = "define_fact"

	serviceIdArgName = "service_id"
	factNameArgName  = "fact_name"
	recipeArgName    = "fact_recipe"
)

func GenerateDefineFactBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, factsEngine *facts_engine.FactsEngine) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		instructionPosition := shared_helpers.GetCallerPositionFromThread(thread)
		defineFactInstruction := newEmptyDefineFactInstruction(factsEngine, instructionPosition)
		if interpretationError := defineFactInstruction.parseStartosisArgs(b, args, kwargs); interpretationError != nil {
			return nil, interpretationError
		}
		*instructionsQueue = append(*instructionsQueue, defineFactInstruction)
		return starlark.None, nil
	}
}

type DefineFactInstruction struct {
	factsEngine *facts_engine.FactsEngine

	position       *kurtosis_instruction.InstructionPosition
	starlarkKwargs starlark.StringDict

	serviceId  kurtosis_backend_service.ServiceID
	factName   string
	factRecipe *kurtosis_core_rpc_api_bindings.FactRecipe
}

func newEmptyDefineFactInstruction(factsEngine *facts_engine.FactsEngine, position *kurtosis_instruction.InstructionPosition) *DefineFactInstruction {
	return &DefineFactInstruction{
		factsEngine:    factsEngine,
		position:       position,
		serviceId:      "",
		factName:       "",
		factRecipe:     nil,
		starlarkKwargs: starlark.StringDict{},
	}
}

func (instruction *DefineFactInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return instruction.position
}

func (instruction *DefineFactInstruction) GetCanonicalInstruction() *kurtosis_core_rpc_api_bindings.StarlarkInstruction {
	args := []*kurtosis_core_rpc_api_bindings.StarlarkInstructionArg{
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[serviceIdArgName]), serviceIdArgName, kurtosis_instruction.Representative),
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[factNameArgName]), factNameArgName, kurtosis_instruction.Representative),
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[recipeArgName]), recipeArgName, kurtosis_instruction.NotRepresentative),
	}
	return binding_constructors.NewStarlarkInstruction(instruction.position.ToAPIType(), DefineFactBuiltinName, instruction.String(), args)
}

func (instruction *DefineFactInstruction) Execute(_ context.Context) (*string, error) {
	err := instruction.factsEngine.PushRecipe(instruction.factRecipe)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to wait for fact '%v' on service '%v'", instruction.factName, instruction.serviceId)
	}
	return nil, nil
}

func (instruction *DefineFactInstruction) String() string {
	return shared_helpers.CanonicalizeInstruction(DefineFactBuiltinName, kurtosis_instruction.NoArgs, instruction.starlarkKwargs)
}

func (instruction *DefineFactInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	if !environment.DoesServiceIdExist(instruction.serviceId) {
		return startosis_errors.NewValidationError("There was an error validating exec with service ID '%v' that does not exist", instruction.serviceId)
	}
	// TODO(victor.colombo): Add fact validation
	return nil
}

func (instruction *DefineFactInstruction) parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) *startosis_errors.InterpretationError {

	var serviceIdArg starlark.String
	var factNameArg starlark.String
	var recipeConfigArg *starlarkstruct.Struct

	if err := starlark.UnpackArgs(b.Name(), args, kwargs, serviceIdArgName, &serviceIdArg, factNameArgName, &factNameArg, recipeArgName, &recipeConfigArg); err != nil {
		return startosis_errors.WrapWithInterpretationError(err, "Failed parsing arguments for function '%s' (unparsed arguments were: '%v' '%v')", DefineFactBuiltinName, args, kwargs)
	}

	instruction.starlarkKwargs[serviceIdArgName] = serviceIdArg
	instruction.starlarkKwargs[factNameArgName] = factNameArg
	instruction.starlarkKwargs[recipeArgName] = recipeConfigArg
	instruction.starlarkKwargs.Freeze()

	serviceId, interpretationErr := kurtosis_instruction.ParseServiceId(serviceIdArg)
	if interpretationErr != nil {
		return interpretationErr
	}

	factName, interpretationErr := kurtosis_instruction.ParseFactName(factNameArg)
	if interpretationErr != nil {
		return interpretationErr
	}

	factRecipe, interpretationErr := kurtosis_instruction.ParseHttpRequestFactRecipe(recipeConfigArg)
	if interpretationErr != nil {
		return interpretationErr
	}

	instruction.serviceId = serviceId
	instruction.factName = factName
	instruction.factRecipe = binding_constructors.NewHttpRequestFactRecipeWithDefaultRefresh(string(serviceId), factName, factRecipe)
	return nil
}
