package wait

import (
	"context"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/assert"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/recipe"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"time"
)

const (
	WaitBuiltinName = "wait"

	recipeArgName           = "recipe"
	targetKeyArgName        = "field"
	assertionArgName        = "assertion"
	targetArgName           = "target_value"
	optionalIntervalArgName = "interval?"
	optionalTimeoutArgName  = "timeout?"
)

func GenerateWaitBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, recipeExecutor *runtime_value_store.RuntimeValueStore, serviceNetwork service_network.ServiceNetwork) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, builtin *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		instructionPosition := shared_helpers.GetCallerPositionFromThread(thread)
		waitInstruction := newEmptyWaitInstructionInstruction(serviceNetwork, instructionPosition, recipeExecutor)
		if interpretationError := waitInstruction.parseStartosisArgs(builtin, args, kwargs); interpretationError != nil {
			return nil, interpretationError
		}
		waitInstruction.resultUuid = recipeExecutor.CreateValue()
		returnValue := waitInstruction.httpRequestRecipe.CreateStarlarkReturnValue(waitInstruction.resultUuid)
		*instructionsQueue = append(*instructionsQueue, waitInstruction)
		return returnValue, nil
	}
}

type WaitInstruction struct {
	serviceNetwork service_network.ServiceNetwork

	position       *kurtosis_instruction.InstructionPosition
	starlarkKwargs starlark.StringDict

	recipeExecutor    *runtime_value_store.RuntimeValueStore
	httpRequestRecipe *recipe.HttpRequestRecipe
	recipeConfigArg   *starlarkstruct.Struct
	resultUuid        string
	targetKey         string
	assertion         string
	target            starlark.Comparable
	backoff           *backoff.ExponentialBackOff
}

func newEmptyWaitInstructionInstruction(serviceNetwork service_network.ServiceNetwork, position *kurtosis_instruction.InstructionPosition, recipeExecutor *runtime_value_store.RuntimeValueStore) *WaitInstruction {
	return &WaitInstruction{
		serviceNetwork:    serviceNetwork,
		position:          position,
		recipeExecutor:    recipeExecutor,
		httpRequestRecipe: nil,
		recipeConfigArg:   nil,
		resultUuid:        "",
		starlarkKwargs:    nil,
		targetKey:         "",
		assertion:         "",
		target:            nil,
		backoff:           backoff.NewExponentialBackOff(),
	}
}

func (instruction *WaitInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return instruction.position
}

func (instruction *WaitInstruction) GetCanonicalInstruction() *kurtosis_core_rpc_api_bindings.StarlarkInstruction {
	args := []*kurtosis_core_rpc_api_bindings.StarlarkInstructionArg{
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[recipeArgName]), recipeArgName, kurtosis_instruction.Representative),
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[targetKeyArgName]), targetKeyArgName, kurtosis_instruction.Representative),
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[assertionArgName]), assertionArgName, kurtosis_instruction.Representative),
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[targetArgName]), targetArgName, kurtosis_instruction.Representative),
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[optionalIntervalArgName]), optionalIntervalArgName, kurtosis_instruction.NotRepresentative),
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[optionalTimeoutArgName]), optionalTimeoutArgName, kurtosis_instruction.NotRepresentative),
	}
	return binding_constructors.NewStarlarkInstruction(instruction.position.ToAPIType(), WaitBuiltinName, instruction.String(), args)
}

func (instruction *WaitInstruction) Execute(ctx context.Context) (*string, error) {
	var (
		requestErr error
		assertErr  error
	)
	lastResult := map[string]starlark.Comparable{}
	for {
		backoffDuration := instruction.backoff.NextBackOff()
		if backoffDuration == backoff.Stop {
			break
		}
		lastResult, requestErr = instruction.httpRequestRecipe.Execute(ctx, instruction.serviceNetwork)
		if requestErr != nil {
			time.Sleep(backoffDuration)
			continue
		}
		instruction.recipeExecutor.SetValue(instruction.resultUuid, lastResult)
		value, found := lastResult[instruction.targetKey]
		if !found {
			return nil, stacktrace.NewError("Error grabbing value from key '%v'", instruction.targetKey)
		}
		assertErr = assert.Assert(value, instruction.assertion, instruction.target)
		if assertErr != nil {
			time.Sleep(backoffDuration)
			continue
		}
		break
	}
	if requestErr != nil {
		return nil, stacktrace.Propagate(requestErr, "Error executing HTTP recipe")
	}
	if assertErr != nil {
		return nil, stacktrace.Propagate(requestErr, "Error asserting HTTP recipe on '%v'", WaitBuiltinName)
	}
	instructionResult := fmt.Sprintf("Value obtained '%v'", lastResult)
	return &instructionResult, nil
}

func (instruction *WaitInstruction) String() string {
	return shared_helpers.CanonicalizeInstruction(WaitBuiltinName, kurtosis_instruction.NoArgs, instruction.starlarkKwargs)
}

func (instruction *WaitInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	return nil
}

func (instruction *WaitInstruction) parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) *startosis_errors.InterpretationError {
	var (
		recipeConfigArg  *starlarkstruct.Struct
		targetKeyArg     starlark.String
		assertionArg     starlark.String
		targetArg        starlark.Comparable
		optionalInterval starlark.String = ""
		optionalTimeout  starlark.String = ""
	)

	if err := starlark.UnpackArgs(b.Name(), args, kwargs, recipeArgName, &recipeConfigArg, targetKeyArgName, &targetKeyArg, assertionArgName, &assertionArg, targetArgName, &targetArg, optionalIntervalArgName, &optionalInterval, optionalTimeoutArgName, &optionalTimeout); err != nil {
		return startosis_errors.NewInterpretationError(err.Error())
	}
	instruction.starlarkKwargs = starlark.StringDict{
		recipeArgName:           recipeConfigArg,
		targetKeyArgName:        targetKeyArg,
		assertionArgName:        assertionArg,
		targetArgName:           targetArg,
		optionalIntervalArgName: optionalInterval,
		optionalTimeoutArgName:  optionalTimeout,
	}
	instruction.starlarkKwargs.Freeze()

	var err *startosis_errors.InterpretationError
	instruction.httpRequestRecipe, err = kurtosis_instruction.ParseHttpRequestRecipe(recipeConfigArg)
	if err != nil {
		return err
	}
	instruction.assertion = string(assertionArg)
	instruction.target = targetArg
	instruction.targetKey = string(targetKeyArg)
	if optionalInterval != "" {
		interval, parseErr := time.ParseDuration(optionalInterval.GoString())
		if parseErr != nil {
			return startosis_errors.WrapWithInterpretationError(parseErr, "An error occurred when parsing interval '%v'", optionalInterval.GoString())
		}
		instruction.backoff.InitialInterval = interval
	}
	if optionalTimeout != "" {
		timeout, parseErr := time.ParseDuration(optionalTimeout.GoString())
		if parseErr != nil {
			return startosis_errors.NewInterpretationError("An error occurred when parsing timeout '%v'", optionalTimeout)
		}
		instruction.backoff.MaxElapsedTime = timeout
	}

	if _, found := assert.StringTokenToComparisonStarlarkToken[instruction.assertion]; !found && instruction.assertion != "IN" && instruction.assertion != "NOT_IN" {
		return startosis_errors.NewInterpretationError("'%v' is not a valid assertion", assertionArg)
	}
	if _, ok := instruction.target.(*starlark.List); (instruction.assertion == "IN" || instruction.assertion == "NOT_IN") && !ok {
		return startosis_errors.NewInterpretationError("'%v' assertion requires list, got '%v'", assertionArg, targetArg.Type())
	}
	return nil
}
