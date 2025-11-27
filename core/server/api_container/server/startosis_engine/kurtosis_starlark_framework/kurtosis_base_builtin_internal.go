package kurtosis_starlark_framework

import (
	"fmt"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/starlark_warning"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/stacktrace"
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

	if err := printWarningForArguments(arguments, baseBuiltin); err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "An error occurred printing the deprecation warning messages")
	}
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

func printWarningForArguments(argumentSet *builtin_argument.ArgumentValuesSet, builtin *KurtosisBaseBuiltin) error {
	// if instruction is deprecated, print the deprecated warning for the instruction
	// ignore the warnings associated with arguments
	arguments := argumentSet.GetDefinition()
	if builtin.Deprecation != nil {
		warningMessage := getFormattedWarningMessageForInstruction(builtin.Deprecation, builtin.Name)
		starlark_warning.PrintOnceAtTheEndOfExecutionf("%v %v", starlark_warning.WarningConstant, warningMessage)
	} else {
		// print if arguments for this builtIn is deprecated.
		for _, argument := range arguments {
			if argumentSet.IsSet(argument.Name) && argument.IsDeprecated() {
				shouldShownDeprecationNotice := true
				maybeShouldShownDeprecationNoticeFunc := argument.Deprecation.GetMaybeShouldShowDeprecationNoticeBaseOnArgumentValueFunc()
				if maybeShouldShownDeprecationNoticeFunc != nil {
					argumentValue := argument.ZeroValueProvider()
					if err := argumentSet.ExtractArgumentValue(argument.Name, &argumentValue); err != nil {
						return stacktrace.Propagate(err, "An error occurred extracting argument '%v' value", argument.Name)
					}
					shouldShownDeprecationNotice = maybeShouldShownDeprecationNoticeFunc(argumentValue)
				}

				if shouldShownDeprecationNotice {
					warningMessage := getFormattedWarningMessageForArgument(argument.Deprecation, builtin.Name, argument.Name)
					starlark_warning.PrintOnceAtTheEndOfExecutionf("%v %v", starlark_warning.WarningConstant, warningMessage)
				}
			}
		}
	}
	return nil
}

func getFormattedWarningMessageForInstruction(deprecation *starlark_warning.DeprecationNotice, builtinName string) string {
	deprecationDateStr := deprecation.GetDeprecatedDate()
	deprecationReason := deprecation.GetMitigation()
	return fmt.Sprintf("%q instruction will be deprecated by %v. %v", builtinName, deprecationDateStr, deprecationReason)
}

func getFormattedWarningMessageForArgument(deprecation *starlark_warning.DeprecationNotice, builtinName string, argumentName string) string {
	deprecationDateStr := "soon"
	if deprecation.IsDeprecatedDateScheduled() {
		deprecationDateStr = fmt.Sprintf("by %v", deprecation.GetDeprecatedDate())
	}

	deprecationReason := deprecation.GetMitigation()
	return fmt.Sprintf(
		"%q field for %q will be deprecated %v. %v",
		argumentName,
		builtinName,
		deprecationDateStr,
		deprecationReason,
	)
}
