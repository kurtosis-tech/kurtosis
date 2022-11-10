package wait

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	kurtosis_backend_service "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/facts_engine"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_executor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
)

const (
	WaitBuiltinName = "wait"

	serviceIdArgName = "service_id"
	factNameArgName  = "fact_name"

	kurtosisNamespace                = "kurtosis"
	factReplacementPlaceholderFormat = "{{" + kurtosisNamespace + ":%v:%v.fact}}"
)

func GenerateWaitBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, factsEngine *facts_engine.FactsEngine) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		serviceId, factName, interpretationError := parseStartosisArgs(b, args, kwargs)
		if interpretationError != nil {
			return nil, interpretationError
		}
		waitInstruction := NewWaitInstruction(factsEngine, *shared_helpers.GetCallerPositionFromThread(thread), serviceId, factName)
		*instructionsQueue = append(*instructionsQueue, waitInstruction)
		returnValue, interpretationError := makeWaitInterpretationReturnValue(serviceId, factName)
		if interpretationError != nil {
			return nil, interpretationError
		}
		return returnValue, nil
	}
}

type WaitInstruction struct {
	factsEngine *facts_engine.FactsEngine

	position  kurtosis_instruction.InstructionPosition
	serviceId kurtosis_backend_service.ServiceID
	factName  string
}

func NewWaitInstruction(factsEngine *facts_engine.FactsEngine, position kurtosis_instruction.InstructionPosition, serviceId kurtosis_backend_service.ServiceID, factName string) *WaitInstruction {
	return &WaitInstruction{
		factsEngine: factsEngine,
		position:    position,
		serviceId:   serviceId,
		factName:    factName,
	}
}

func (instruction *WaitInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return &instruction.position
}

func (instruction *WaitInstruction) GetCanonicalInstruction() string {
	return shared_helpers.CanonicalizeInstruction(WaitBuiltinName, starlark.StringDict{
		serviceIdArgName: starlark.String(instruction.serviceId),
		factNameArgName:  starlark.String(instruction.factName),
	}, &instruction.position)
}

func (instruction *WaitInstruction) Execute(ctx context.Context, _ *startosis_executor.ExecutionEnvironment) error {
	_, err := instruction.factsEngine.WaitForValue(facts_engine.GetFactId(string(instruction.serviceId), instruction.factName))
	if err != nil {
		return stacktrace.Propagate(err, "Failed to wait for fact '%v' on service '%v'", instruction.factName, instruction.serviceId)
	}
	return nil
}

func (instruction *WaitInstruction) String() string {
	return instruction.GetCanonicalInstruction()
}

func (instruction *WaitInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	if !environment.DoesServiceIdExist(instruction.serviceId) {
		return stacktrace.NewError("There was an error validating exec with service ID '%v' that does not exist", instruction.serviceId)
	}
	// TODO(victor.colombo): Add fact validation
	return nil
}

func parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (service.ServiceID, string, *startosis_errors.InterpretationError) {

	var serviceIdArg starlark.String
	var factNameArg starlark.String
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, serviceIdArgName, &serviceIdArg, factNameArgName, &factNameArg); err != nil {
		return "", "", startosis_errors.NewInterpretationError(err.Error())
	}

	serviceId, interpretationErr := kurtosis_instruction.ParseServiceId(serviceIdArg)
	if interpretationErr != nil {
		return "", "", interpretationErr
	}

	factName, interpretationErr := kurtosis_instruction.ParseFactName(factNameArg)
	if interpretationErr != nil {
		return "", "", interpretationErr
	}

	return serviceId, factName, nil
}

func makeWaitInterpretationReturnValue(serviceId service.ServiceID, factName string) (starlark.String, *startosis_errors.InterpretationError) {
	fact := starlark.String(fmt.Sprintf(factReplacementPlaceholderFormat, serviceId, factName))
	return fact, nil
}
