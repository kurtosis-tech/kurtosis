package kurtosis_starlark_framework

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/starlark_warning"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
)

// KurtosisBaseBuiltinInternal can be seen as an instance of a given Kurtosis builtin. It is composed of the name of
// the builtin, its position in the starlark script being interpreted and the arguments passed to the builtin (including
// its schema).
//
// It implements several functions like String() or GetPosition() that each builtin will naturally inherit.
type KurtosisBaseBuiltinInternal struct {
	builtinName string

	position *KurtosisBuiltinPosition

	arguments *builtin_argument.ArgumentValuesSet
}

func newKurtosisBaseBuiltinInternal(builtinName string, position *KurtosisBuiltinPosition, arguments *builtin_argument.ArgumentValuesSet) *KurtosisBaseBuiltinInternal {
	return &KurtosisBaseBuiltinInternal{
		builtinName: builtinName,
		position:    position,
		arguments:   arguments,
	}
}

func WrapKurtosisBaseBuiltin(baseBuiltin *KurtosisBaseBuiltin, thread *starlark.Thread, args starlark.Tuple, kwargs []starlark.Tuple) (*KurtosisBaseBuiltinInternal, *startosis_errors.InterpretationError) {
	// First store the argument values passed to the builtin
	arguments, interpretationErr := builtin_argument.CreateNewArgumentValuesSet(baseBuiltin.Name, baseBuiltin.Arguments, args, kwargs)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	printWarningForArguments(arguments.GetDefinition(), baseBuiltin)

	// Second store the position at which the builtin is called within the source script
	callFrame := thread.CallStack().At(1)
	position := NewKurtosisBuiltinPosition(callFrame.Pos.Filename(), callFrame.Pos.Line, callFrame.Pos.Col)

	return &KurtosisBaseBuiltinInternal{
		builtinName: baseBuiltin.Name,
		position:    position,
		arguments:   arguments,
	}, nil
}

func (builtin *KurtosisBaseBuiltinInternal) GetName() string {
	return builtin.builtinName
}

func (builtin *KurtosisBaseBuiltinInternal) String() string {
	return fmt.Sprintf("%s%s", builtin.GetName(), builtin.arguments.String())
}

func (builtin *KurtosisBaseBuiltinInternal) GetArguments() *builtin_argument.ArgumentValuesSet {
	return builtin.arguments
}

func (builtin *KurtosisBaseBuiltinInternal) GetPosition() *KurtosisBuiltinPosition {
	return builtin.position
}

func printWarningForArguments(arguments []*builtin_argument.BuiltinArgument, builtin *KurtosisBaseBuiltin) {
	// if instruction is deprecated, print the deprecated warning for the instruction
	// ignore the warnings associated with arguments
	if builtin.Deprecation != nil {
		warningMessage := getFormattedWarningMessageForInstruction(builtin.Deprecation, builtin.Name)
		starlark_warning.PrintOnceAtTheEndOfExecutionf("%v %v", starlark_warning.WarningConstant, warningMessage)
	} else {
		// print if arguments for this builtIn is deprecated.
		for _, argument := range arguments {
			if argument.IsDeprecated() {
				warningMessage := getFormattedWarningMessageForArgument(argument.Deprecation, builtin.Name, argument.Name)
				starlark_warning.PrintOnceAtTheEndOfExecutionf("%v %v", starlark_warning.WarningConstant, warningMessage)
			}
		}
	}
}

func getFormattedWarningMessageForInstruction(deprecation *starlark_warning.DeprecationNotice, builtinName string) string {
	deprecationDateStr := deprecation.GetDeprecatedDate()
	deprecationReason := deprecation.GetMitigation()
	return fmt.Sprintf("%q instruction will be deprecated by %v. %v", builtinName, deprecationDateStr, deprecationReason)
}

func getFormattedWarningMessageForArgument(deprecation *starlark_warning.DeprecationNotice, builtinName string, argumentName string) string {
	deprecationDateStr := deprecation.GetDeprecatedDate()
	deprecationReason := deprecation.GetMitigation()
	return fmt.Sprintf(
		"%q field for %q will be deprecated by %v. %v",
		argumentName,
		builtinName,
		deprecationDateStr,
		deprecationReason,
	)
}
