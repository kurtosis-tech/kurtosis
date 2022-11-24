package wait

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
)

const (
	WaitBuiltinName = "wait"

	serviceIdArgName = "service_id"
	factNameArgName  = "fact_name"
)

func GenerateWaitBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, factsEngine *facts_engine.FactsEngine) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, builtin *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		instructionPosition := shared_helpers.GetCallerPositionFromThread(thread)
		waitInstruction := newEmptyWaitInstruction(factsEngine, instructionPosition)
		if interpretationError := waitInstruction.parseStartosisArgs(builtin, args, kwargs); interpretationError != nil {
			return nil, interpretationError
		}
		*instructionsQueue = append(*instructionsQueue, waitInstruction)
		returnValue := shared_helpers.MakeWaitInterpretationReturnValue(waitInstruction.serviceId, waitInstruction.factName)
		return returnValue, nil
	}
}

type WaitInstruction struct {
	factsEngine *facts_engine.FactsEngine

	position       *kurtosis_instruction.InstructionPosition
	starlarkKwargs starlark.StringDict

	serviceId kurtosis_backend_service.ServiceID
	factName  string
}

func newEmptyWaitInstruction(factsEngine *facts_engine.FactsEngine, position *kurtosis_instruction.InstructionPosition) *WaitInstruction {
	return &WaitInstruction{
		factsEngine:    factsEngine,
		position:       position,
		serviceId:      "",
		factName:       "",
		starlarkKwargs: starlark.StringDict{},
	}
}

func (instruction *WaitInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return instruction.position
}

func (instruction *WaitInstruction) GetCanonicalInstruction() *kurtosis_core_rpc_api_bindings.KurtosisInstruction {
	args := []*kurtosis_core_rpc_api_bindings.KurtosisInstructionArg{
		binding_constructors.NewKurtosisInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[serviceIdArgName]), serviceIdArgName, true),
		binding_constructors.NewKurtosisInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[factNameArgName]), factNameArgName, true),
	}
	return binding_constructors.NewKurtosisInstruction(instruction.position.ToAPIType(), WaitBuiltinName, instruction.String(), args)
}

func (instruction *WaitInstruction) Execute(_ context.Context) (*string, error) {
	_, err := instruction.factsEngine.WaitForValue(facts_engine.GetFactId(string(instruction.serviceId), instruction.factName))
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to wait for fact '%v' on service '%v'", instruction.factName, instruction.serviceId)
	}
	return nil, nil
}

func (instruction *WaitInstruction) String() string {
	return shared_helpers.CanonicalizeInstruction(WaitBuiltinName, kurtosis_instruction.NoArgs, instruction.starlarkKwargs)
}

func (instruction *WaitInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	if !environment.DoesServiceIdExist(instruction.serviceId) {
		return stacktrace.NewError("There was an error validating exec with service ID '%v' that does not exist", instruction.serviceId)
	}
	// TODO(victor.colombo): Add fact validation
	return nil
}

func (instruction *WaitInstruction) parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) *startosis_errors.InterpretationError {
	var serviceIdArg starlark.String
	var factNameArg starlark.String
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, serviceIdArgName, &serviceIdArg, factNameArgName, &factNameArg); err != nil {
		return startosis_errors.NewInterpretationError(err.Error())
	}
	instruction.starlarkKwargs[serviceIdArgName] = serviceIdArg
	instruction.starlarkKwargs[factNameArgName] = factNameArg
	instruction.starlarkKwargs.Freeze()

	serviceId, interpretationErr := kurtosis_instruction.ParseServiceId(serviceIdArg)
	if interpretationErr != nil {
		return interpretationErr
	}

	factName, interpretationErr := kurtosis_instruction.ParseFactName(factNameArg)
	if interpretationErr != nil {
		return interpretationErr
	}
	instruction.serviceId = serviceId
	instruction.factName = factName
	return nil
}
