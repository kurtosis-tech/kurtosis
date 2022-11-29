package kurtosis_print

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers/magic_string_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
	"strings"
)

type separator string
type end string

const (
	PrintBuiltinName = "print"

	separatorArgName = "sep"
	defaultSeparator = separator(" ")

	endArgName = "end"
	defaultEnd = end("")
)

func GeneratePrintBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, recipeExecutor *runtime_value_store.RuntimeValueStore) func(thread *starlark.Thread, builtin *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		args, separatorStr, endStr, interpretationError := parseStartosisArg(b, args, kwargs)
		if interpretationError != nil {
			return nil, interpretationError
		}
		instructionPosition := shared_helpers.GetCallerPositionFromThread(thread)
		printInstruction := NewPrintInstruction(instructionPosition, args, separatorStr, endStr, recipeExecutor)
		*instructionsQueue = append(*instructionsQueue, printInstruction)
		return starlark.None, nil
	}
}

type PrintInstruction struct {
	position       *kurtosis_instruction.InstructionPosition
	args           []starlark.Value
	separator      separator
	end            end
	recipeExecutor *runtime_value_store.RuntimeValueStore
}

func NewPrintInstruction(position *kurtosis_instruction.InstructionPosition, args []starlark.Value, separatorStr separator, endStr end, recipeExecutor *runtime_value_store.RuntimeValueStore) *PrintInstruction {
	return &PrintInstruction{
		position:       position,
		args:           args,
		separator:      separatorStr,
		end:            endStr,
		recipeExecutor: recipeExecutor,
	}
}

func (instruction *PrintInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return instruction.position
}

func (instruction *PrintInstruction) GetCanonicalInstruction() *kurtosis_core_rpc_api_bindings.StarlarkInstruction {
	args := make([]*kurtosis_core_rpc_api_bindings.StarlarkInstructionArg, len(instruction.args))
	for idx, arg := range instruction.args {
		args[idx] = binding_constructors.NewStarlarkInstructionArg(shared_helpers.CanonicalizeArgValue(arg), kurtosis_instruction.Representative)
	}
	args = append(args, binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(starlark.String(instruction.separator)), separatorArgName, kurtosis_instruction.NotRepresentative))
	args = append(args, binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(starlark.String(instruction.end)), endArgName, kurtosis_instruction.NotRepresentative))
	return binding_constructors.NewStarlarkInstruction(instruction.position.ToAPIType(), PrintBuiltinName, instruction.String(), args)
}

func (instruction *PrintInstruction) Execute(_ context.Context) (*string, error) {
	serializedArgs := make([]string, len(instruction.args))
	for idx, genericArg := range instruction.args {
		if genericArg == nil {
			return nil, stacktrace.NewError("'%s' statement with nil argument. This is unexpected: '%v'", PrintBuiltinName, instruction.args)
		}
		switch arg := genericArg.(type) {
		case starlark.String:
			serializedArgs[idx] = arg.GoString()
		default:
			serializedArgs[idx] = arg.String()
		}
		newValue, err := magic_string_helper.ReplaceRuntimeValueInString(serializedArgs[idx], instruction.recipeExecutor)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error replacing runtime value '%v'", serializedArgs[idx])
		}
		serializedArgs[idx] = newValue
	}
	instructionOutput := fmt.Sprintf("%s%s", strings.Join(serializedArgs, string(instruction.separator)), instruction.end)

	return &instructionOutput, nil
}

func (instruction *PrintInstruction) String() string {
	return shared_helpers.CanonicalizeInstruction(PrintBuiltinName, instruction.args, instruction.getKwargs())
}

func (instruction *PrintInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	// nothing to do for now
	// TODO(gb): maybe in the future validate that if we're using a magic string, it points to something real
	return nil
}

// parseStartosisArg is specific because in python we can pass anything to a print statement. The contract is the following:
// - as many positional arg as wanted
// - those positional args can be of any type: string, int, array, dictionary, etc.
// - all those arguments will be stringified using the String() function and concatenated using the separator optionally passed in (see below)
// - an optional `sep` named argument which must be a string, representing the separator that will be used to concatenate the positional args (defaults to a blank space if absent)
// - an optional `end` named argument which must be a string, representing the endOfLine that will be used (defaults to '\n')
func parseStartosisArg(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) ([]starlark.Value, separator, end, *startosis_errors.InterpretationError) {
	// read positional args first
	argsList := make([]starlark.Value, len(args))
	copy(argsList, args)

	// read kwargs - throw if it's something different than the separator
	separatorKwarg := defaultSeparator
	endKwarg := defaultEnd
	for _, kwarg := range kwargs {
		switch kwarg.Index(0) {
		case starlark.String(separatorArgName):
			separatorKwargStr, interpretationError := kurtosis_instruction.ParseNonEmptyString(separatorArgName, kwarg.Index(1))
			if interpretationError != nil {
				return nil, "", "", interpretationError
			}
			separatorKwarg = separator(separatorKwargStr)
		case starlark.String(endArgName):
			endKwargStr, interpretationError := kurtosis_instruction.ParseNonEmptyString(endArgName, kwarg.Index(1))
			if interpretationError != nil {
				return nil, "", "", interpretationError
			}
			endKwarg = end(endKwargStr)
		default:
			// This is the python default error message
			return nil, "", "", startosis_errors.NewInterpretationError("%s: unexpected keyword argument '%'", b.Name(), kwarg.Index(0))
		}
	}
	return argsList, separatorKwarg, endKwarg, nil
}

func (instruction *PrintInstruction) getKwargs() starlark.StringDict {
	nonDefaultKwargs := starlark.StringDict{}
	if instruction.separator != defaultSeparator {
		nonDefaultKwargs[separatorArgName] = starlark.String(instruction.separator)
	}
	if instruction.end != defaultEnd {
		nonDefaultKwargs[endArgName] = starlark.String(instruction.end)
	}
	return nonDefaultKwargs
}
