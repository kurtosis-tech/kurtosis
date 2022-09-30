package lowlevel

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

const (
	arg1Key = "arg1"
	arg2Key = "arg2"
	arg3Key = "arg3"

	arg1Value = "arg1Value"
	arg2Value = "arg2Value"
	arg3Value1 = "arg3Value1"
	arg3Value2 = "arg3Value2"
	arg3Value3 = "arg3Value3"

	flag1Key = "flag1"
	flag2Key = "flag2"
)

var validArgsConfig = []*args.ArgConfig{
	{
		Key: arg1Key,
	},
	{
		Key: arg2Key,
	},
	{
		Key:      arg3Key,
		IsGreedy: true,
	},
}
var validTokens = []string{
	arg1Value,
	arg2Value,
	arg3Value1,
	arg3Value2,
	arg3Value3,
}
var doNothingFunc = func(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {
	return nil
}

func TestMustGetCobraCommand_EmptyCommandStrCausesPanic(t *testing.T) {
	kurtosisCmd := &LowlevelKurtosisCommand{
		CommandStr:       "   ",
		ShortDescription: "Short description",
		LongDescription:  "This is a very long description",
		RunFunc: doNothingFunc,
	}

	requirePanicWithSubstring(
		t,
		"A Kurtosis command must have a command string",
		kurtosisCmd.MustGetCobraCommand,
		"Expected an empty command string to cause a panic",
	)
}

func TestMustGetCobraCommand_EmptyShortDescriptionCausesPanic(t *testing.T) {
	kurtosisCmd := &LowlevelKurtosisCommand{
		CommandStr:       "test",
		ShortDescription: "     ",
		LongDescription:  "Long description",
		RunFunc: doNothingFunc,
	}

	requirePanicWithSubstring(
		t,
		"A short description must be defined for command",
		kurtosisCmd.MustGetCobraCommand,
		"Expected an empty short description to cause a panic",
	)
}

func TestMustGetCobraCommand_EmptyLongDescriptionCausesPanic(t *testing.T) {
	kurtosisCmd := &LowlevelKurtosisCommand{
		CommandStr:       "test",
		ShortDescription: "Short description",
		LongDescription:  "     ",
		RunFunc: doNothingFunc,
	}

	requirePanicWithSubstring(
		t,
		"A long description must be defined for command",
		kurtosisCmd.MustGetCobraCommand,
		"Expected an empty long description to cause a panic",
	)
}

func TestMustGetCobraCommand_NilRunFunctionCausesPanic(t *testing.T) {
	kurtosisCmd := &LowlevelKurtosisCommand{
		CommandStr:       "test",
		ShortDescription: "Short description",
		LongDescription:  "This is a very long description",
		RunFunc: nil,
	}

	requirePanicWithSubstring(
		t,
		"A run function must be defined for command",
		kurtosisCmd.MustGetCobraCommand,
		"Expected a nil run function to cause a panic",
	)
}

func TestMustGetCobraCommand_DuplicateArgsCausePanic(t *testing.T) {
	kurtosisCmd := &LowlevelKurtosisCommand{
		CommandStr:       "test",
		ShortDescription: "Short description",
		LongDescription:  "This is a very long description",
		Args:             []*args.ArgConfig{
			{
				Key: arg1Key,
			},
			{
				Key: arg1Key,
			},
		},
		RunFunc: doNothingFunc,
	}

	requirePanicWithSubstring(
		t,
		"Found duplicate args with key",
		kurtosisCmd.MustGetCobraCommand,
		"Expected a panic when trying to supply two args with the same key",
	)
}

func TestMustGetCobraCommand_DuplicateFlagsCausePanic(t *testing.T) {
	kurtosisCmd := &LowlevelKurtosisCommand{
		CommandStr:       "test",
		ShortDescription: "Short description",
		LongDescription:  "This is a very long description",
		Flags: []*flags.FlagConfig{
			{
				Key: flag1Key,
			},
			{
				Key: flag1Key,
			},
		},
		RunFunc: doNothingFunc,
	}

	requirePanicWithSubstring(
		t,
		"Found duplicate flags",
		kurtosisCmd.MustGetCobraCommand,
		"Expected a panic when trying to supply two flags with the same key",
	)
}

func TestMustGetCobraCommand_FlagsWithMismatchedDefaulValuesCausePanic(t *testing.T) {
	illegalFlagVariants := []*flags.FlagConfig{
		{
			Key:     flag1Key,
			Type:    flags.FlagType_Bool,
			Default: "123",
		},
		{
			Key:     flag2Key,
			Type:    flags.FlagType_Uint32,
			Default: "true",
		},
	}
	for _, illegalFlag := range illegalFlagVariants {
		kurtosisCmd := &LowlevelKurtosisCommand{
			CommandStr:       "test",
			ShortDescription: "Short description",
			LongDescription:  "This is a very long description",
			Flags: []*flags.FlagConfig{
				illegalFlag,
			},
			RunFunc: doNothingFunc,
		}

		requirePanicWithSubstring(
			t,
			"An error occurred processing flag",
			kurtosisCmd.MustGetCobraCommand,
			"Expected a panic when trying to set a flag with type '%v' and default value string '%v' that doesn't match the type",
			illegalFlag.Type.String(),
			illegalFlag.Default,
		)
	}
}

func TestMustGetCobraCommand_FlagWithTypeDoesntPanic(t *testing.T) {
	kurtosisCmd := &LowlevelKurtosisCommand{
		CommandStr:       "test",
		ShortDescription: "Short description",
		LongDescription:  "This is a very long description",
		Flags: []*flags.FlagConfig{
			{
				Key:  flag1Key,
				Type: flags.FlagType_String,
			},
		},
		RunFunc: doNothingFunc,
	}

	require.NotPanics(
		t,
		func() { kurtosisCmd.MustGetCobraCommand() },
		"Expected no panic when a flag has a valid type",
	)
}

func TestMustGetCobraCommand_DuplicateFlagShorthandsPanic(t *testing.T) {
	dupedShorthandValue := "x"
	kurtosisCmd := &LowlevelKurtosisCommand{
		CommandStr:       "test",
		ShortDescription: "Short description",
		LongDescription:  "This is a very long description",
		Flags: []*flags.FlagConfig{
			{
				Key:       flag1Key,
				Shorthand: dupedShorthandValue,
			},
			{
				Key:       flag2Key,
				Shorthand: dupedShorthandValue,
			},
		},
		RunFunc: doNothingFunc,
	}

	requirePanicWithSubstring(
		t,
		"but this shorthand is already used by flag",
		kurtosisCmd.MustGetCobraCommand,
		"Expected a panic when setting two flags with the same shorthand value",
	)
}

func TestMustGetCobraCommand_ShorthandsGreaterThanOneLetterPanic(t *testing.T) {
	kurtosisCmd := &LowlevelKurtosisCommand{
		CommandStr:       "test",
		ShortDescription: "Short description",
		LongDescription:  "This is a very long description",
		Flags: []*flags.FlagConfig{
			{
				Key:       flag1Key,
				Shorthand: "this is way too long",
			},
		},
		RunFunc: doNothingFunc,
	}

	requirePanicWithSubstring(
		t,
		"that isn't exactly 1 letter",
		kurtosisCmd.MustGetCobraCommand,
		"Expected a panic when setting a flag whose shorthand is greater than one letter",
	)
}

func TestMustGetCobraCommand_TestEmptyShorthandsDontTriggerShorthandValidation(t *testing.T) {
	kurtosisCmd := &LowlevelKurtosisCommand{
		CommandStr:       "test",
		ShortDescription: "Short description",
		LongDescription:  "This is a very long description",
		Flags: []*flags.FlagConfig{
			{
				Key:       flag1Key,
				Shorthand: "",
			},
			{
				Key:       flag2Key,
				Shorthand: "",
			},
		},
		RunFunc: doNothingFunc,
	}

	require.NotPanics(
		t,
		func() { kurtosisCmd.MustGetCobraCommand() },
		"Expected duplicate emptystring shorthands to be allowed",
	)
}

func TestMustGetCobraCommand_EmptyArgKeyCausesPanic(t *testing.T) {
	kurtosisCmd := &LowlevelKurtosisCommand{
		CommandStr:       "test",
		ShortDescription: "Short description",
		LongDescription:  "This is a very long description",
		Args: []*args.ArgConfig{
			{
				Key: "  ",
			},
		},
		RunFunc: doNothingFunc,
	}

	requirePanicWithSubstring(
		t,
		"Empty arg key defined for command",
		kurtosisCmd.MustGetCobraCommand,
		"Expected a panic when trying to set an arg with an empty key",
	)
}

func TestMustGetCobraCommand_EmptyFlagKeyCausesPanic(t *testing.T) {
	kurtosisCmd := &LowlevelKurtosisCommand{
		CommandStr:       "test",
		ShortDescription: "Short description",
		LongDescription:  "This is a very long description",
		Flags: []*flags.FlagConfig{
			{
				Key: "  ",
			},
		},
		RunFunc: doNothingFunc,
	}

	requirePanicWithSubstring(
		t,
		"Empty flag key defined for command",
		kurtosisCmd.MustGetCobraCommand,
		"Expected a panic when trying to set a flag with an empty key",
	)
}

func TestMustGetCobraCommand_TwoOptionalArgumentsCausePanic(t *testing.T) {
	kurtosisCmd := &LowlevelKurtosisCommand{
		CommandStr:       "test",
		ShortDescription: "Short description",
		LongDescription:  "This is a very long description",
		Args:             []*args.ArgConfig{
			{
				Key: arg1Key,
			},
			{
				Key:        arg2Key,
				IsOptional: true,
			},
			{
				Key:        arg3Key,
				IsOptional: true,
			},
		},
		RunFunc: doNothingFunc,
	}

	requirePanicWithSubstring(
		t,
		"must be the last argument because it's either optional or greedy",
		kurtosisCmd.MustGetCobraCommand,
		"Expected a panic when trying to supply two optional commands, but none happened",
	)
}

func TestMustGetCobraCommand_MiddleGreedyArgCausesPanic(t *testing.T) {
	kurtosisCmd := &LowlevelKurtosisCommand{
		CommandStr:       "test",
		ShortDescription: "Short description",
		LongDescription:  "This is a very long description",
		Args:             []*args.ArgConfig{
			{
				Key: arg1Key,
			},
			{
				Key:      arg2Key,
				IsGreedy: true,
			},
			{
				Key: arg3Key,
			},
		},
		RunFunc: doNothingFunc,
	}

	requirePanicWithSubstring(
		t,
		"must be the last argument because it's either optional or greedy",
		kurtosisCmd.MustGetCobraCommand,
		"Expected a panic when trying to supply a greedy argument with another argument after it",
	)
}

func TestMustGetCobraCommand_WorkingFlagDefaultValueChecking(t *testing.T) {
	kurtosisCmd := &LowlevelKurtosisCommand{
		CommandStr:       "test",
		ShortDescription: "Short description",
		LongDescription:  "This is a very long description",
		Flags: []*flags.FlagConfig{
			{
				Key:     flag1Key,
				Type:    flags.FlagType_Uint32,
				Default: "0",
			},
			{
				Key:     flag2Key,
				Type:    flags.FlagType_Bool,
				Default: "false",
			},
		},
		RunFunc: doNothingFunc,
	}

	require.NotPanics(
		t,
		func() { kurtosisCmd.MustGetCobraCommand() },
		"Expected default value flag validation to work",
	)
}

func TestMustGetCobraCommand_OptionalArgsWithNilDefaultPanic(t *testing.T) {
	kurtosisCmd := &LowlevelKurtosisCommand{
		CommandStr:       "test",
		ShortDescription: "Short description",
		LongDescription:  "This is a very long description",
		Args: []*args.ArgConfig{
			{
				Key:          arg1Key,
				DefaultValue: nil,
				IsOptional:   true,
			},
		},
		RunFunc: doNothingFunc,
	}

	requirePanicWithSubstring(
		t,
		"is optional, but doesn't have a default value",
		kurtosisCmd.MustGetCobraCommand,
		"Expected an optional arg with a nil default to cause a panic",
	)
}

func TestMustGetCobraCommand_RequiredArgWithNilDefaultDoesntPanic(t *testing.T) {
	kurtosisCmd := &LowlevelKurtosisCommand{
		CommandStr:       "test",
		ShortDescription: "Short description",
		LongDescription:  "This is a very long description",
		Args: []*args.ArgConfig{
			{
				Key:          arg1Key,
				DefaultValue: nil,
				IsOptional:   false,
			},
		},
		RunFunc: doNothingFunc,
	}

	require.NotPanics(
		t,
		func() { kurtosisCmd.MustGetCobraCommand() },
		"Expected a required argument to skip default-value validation",
	)
}

func TestMustGetCobraCommand_OptionalNonGreedyArgWithWrongDefaultTypePanics(t *testing.T) {
	kurtosisCmd := &LowlevelKurtosisCommand{
		CommandStr:       "test",
		ShortDescription: "Short description",
		LongDescription:  "This is a very long description",
		Args: []*args.ArgConfig{
			{
				Key:          arg1Key,
				IsGreedy:     false,
				DefaultValue: []string{"foo", "bar"},
				IsOptional:   true,
			},
		},
		RunFunc: doNothingFunc,
	}

	requirePanicWithSubstring(
		t,
		"is optional, but the default value isn't a string",
		kurtosisCmd.MustGetCobraCommand,
		"Expected an optional non-greedy arg with a string slice default type to cause a panic",
	)
}

func TestMustGetCobraCommand_OptionalGreedyArgWithWrongDefaultTypePanics(t *testing.T) {
	kurtosisCmd := &LowlevelKurtosisCommand{
		CommandStr:       "test",
		ShortDescription: "Short description",
		LongDescription:  "This is a very long description",
		Args: []*args.ArgConfig{
			{
				Key:          arg1Key,
				IsGreedy:     true,
				DefaultValue: "foobar",
				IsOptional:   true,
			},
		},
		RunFunc: doNothingFunc,
	}

	requirePanicWithSubstring(
		t,
		"is optional, but the default value isn't a string array",
		kurtosisCmd.MustGetCobraCommand,
		"Expected an optional greedy arg with a string default type to cause a panic",
	)
}


// ====================================================================================================
//                                   Private Helper Functions
// ====================================================================================================
func requirePanicWithSubstring(t *testing.T, expectedSubstringInPanic string, toTest func() *cobra.Command, msgAndArgs ...interface{}) {
	didPanic := false
	var caughtValue interface{}
	func() {
		defer func() {
			if caughtValue = recover(); caughtValue != nil {
				didPanic = true
			}
		}()

		toTest()
	}()

	require.True(t, didPanic, msgAndArgs...)

	caughtError, ok := caughtValue.(error)
	require.True(t, ok, "Expected value '%v' caught during the panic to be an error, but it wasn't - this is very weird as MustGetCobraCommand should always return an error!", caughtValue)

	strValue := caughtError.Error()
	require.True(t, strings.Contains(strValue, expectedSubstringInPanic), msgAndArgs...)
}