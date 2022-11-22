package define_fact

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
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
		serviceId, commandArgs, factRecipe, interpretationError := parseStartosisArgs(b, args, kwargs)
		if interpretationError != nil {
			return nil, interpretationError
		}
		instructionPosition := shared_helpers.GetCallerPositionFromThread(thread)
		defineFactInstruction := NewDefineFactInstruction(factsEngine, instructionPosition, serviceId, commandArgs, factRecipe)
		*instructionsQueue = append(*instructionsQueue, defineFactInstruction)
		return starlark.None, nil
	}
}

type DefineFactInstruction struct {
	factsEngine *facts_engine.FactsEngine

	position   *kurtosis_instruction.InstructionPosition
	serviceId  kurtosis_backend_service.ServiceID
	factName   string
	factRecipe *kurtosis_core_rpc_api_bindings.FactRecipe
}

func NewDefineFactInstruction(factsEngine *facts_engine.FactsEngine, position *kurtosis_instruction.InstructionPosition, serviceId kurtosis_backend_service.ServiceID, factName string, factRecipe *kurtosis_core_rpc_api_bindings.FactRecipe) *DefineFactInstruction {
	return &DefineFactInstruction{
		factsEngine: factsEngine,
		position:    position,
		serviceId:   serviceId,
		factName:    factName,
		factRecipe:  factRecipe,
	}
}

func (instruction *DefineFactInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return instruction.position
}

func (instruction *DefineFactInstruction) GetCanonicalInstruction() string {
	return shared_helpers.MultiLineCanonicalizer.CanonicalizeInstruction(DefineFactBuiltinName, kurtosis_instruction.NoArgs, instruction.getKwargs(), instruction.position)
}

func (instruction *DefineFactInstruction) Execute(_ context.Context) (*string, error) {
	err := instruction.factsEngine.PushRecipe(instruction.factRecipe)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to wait for fact '%v' on service '%v'", instruction.factName, instruction.serviceId)
	}
	return nil, nil
}

func (instruction *DefineFactInstruction) String() string {
	return shared_helpers.SingleLineCanonicalizer.CanonicalizeInstruction(DefineFactBuiltinName, kurtosis_instruction.NoArgs, instruction.getKwargs(), instruction.position)
}

func (instruction *DefineFactInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	if !environment.DoesServiceIdExist(instruction.serviceId) {
		return stacktrace.NewError("There was an error validating exec with service ID '%v' that does not exist", instruction.serviceId)
	}
	// TODO(victor.colombo): Add fact validation
	return nil
}

func parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (service.ServiceID, string, *kurtosis_core_rpc_api_bindings.FactRecipe, *startosis_errors.InterpretationError) {

	var serviceIdArg starlark.String
	var factNameArg starlark.String
	var recipeConfigArg *starlarkstruct.Struct

	if err := starlark.UnpackArgs(b.Name(), args, kwargs, serviceIdArgName, &serviceIdArg, factNameArgName, &factNameArg, recipeArgName, &recipeConfigArg); err != nil {
		return "", "", nil, startosis_errors.NewInterpretationError(err.Error())
	}

	serviceId, interpretationErr := kurtosis_instruction.ParseServiceId(serviceIdArg)
	if interpretationErr != nil {
		return "", "", nil, interpretationErr
	}

	factName, interpretationErr := kurtosis_instruction.ParseFactName(factNameArg)
	if interpretationErr != nil {
		return "", "", nil, interpretationErr
	}

	factRecipe, interpretationErr := kurtosis_instruction.ParseHttpRequestFactRecipe(recipeConfigArg)
	if interpretationErr != nil {
		return "", "", nil, interpretationErr
	}

	return serviceId, factName, binding_constructors.NewHttpRequestFactRecipeWithDefaultRefresh(string(serviceId), factName, factRecipe), nil
}

func (instruction *DefineFactInstruction) getKwargs() starlark.StringDict {
	return starlark.StringDict{
		serviceIdArgName: starlark.String(instruction.serviceId),
		factNameArgName:  starlark.String(instruction.factName),
	}
}
