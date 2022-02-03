package kurtosis_command

import (
	"github.com/stretchr/testify/require"
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

var validArgsConfig = []*ArgConfig{
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

func TestMustGetCobraCommand_DuplicateArgsCausePanic(t *testing.T) {
	kurtosisCmd := &KurtosisCommand{
		CommandStr:       "test",
		ShortDescription: "Short description",
		LongDescription:  "This is a very long description",
		Args:             []*ArgConfig{
			{
				Key:             arg1Key,
			},
			{
				Key:             arg1Key,
			},
		},
	}

	require.Panics(
		t,
		func() { kurtosisCmd.MustGetCobraCommand() },
		"Expected a panic when trying to supply two args with the same key",
	)
}

func TestMustGetCobraCommand_DuplicateFlagsCausePanic(t *testing.T) {
	kurtosisCmd := &KurtosisCommand{
		CommandStr:       "test",
		ShortDescription: "Short description",
		LongDescription:  "This is a very long description",
		Flags: []*FlagConfig{
			{
				Key: flag1Key,
			},
			{
				Key: flag2Key,
			},
		},
	}

	require.Panics(
		t,
		func() { kurtosisCmd.MustGetCobraCommand() },
		"Expected a panic when trying to supply two flags with the same key",
	)
}

func TestMustGetCobraCommand_FlagsWithMismatchedDefaulValuesCausePanic(t *testing.T) {
	illegalFlagVariants := []*FlagConfig{
		{
			Key:       flag1Key,
			Type:      FlagType_Bool,
			Default:   "123",
		},
		{
			Key:       flag1Key,
			Type:      FlagType_Uint32,
			Default:   "true",
		},
	}
	for _, illegalFlag := range illegalFlagVariants {
		kurtosisCmd := &KurtosisCommand{
			CommandStr:       "test",
			ShortDescription: "Short description",
			LongDescription:  "This is a very long description",
			Flags: []*FlagConfig{
				illegalFlag,
			},
		}

		require.Panics(
			t,
			func() { kurtosisCmd.MustGetCobraCommand() },
			"Expected a panic when trying to set a flag with type '%v' and default value string '%v' that doesn't match the type",
			illegalFlag.Type.AsString(),
			illegalFlag.Default,
		)
	}
}

func TestMustGetCobraCommand_FlagWithNonsenseTypePanics(t *testing.T) {
	kurtosisCmd := &KurtosisCommand{
		CommandStr:       "test",
		ShortDescription: "Short description",
		LongDescription:  "This is a very long description",
		Flags: []*FlagConfig{
			{
				Key:       flag1Key,
				Type:      FlagType{
					// Technically this shouldn't even be possible since this is private, but we simulate it anywaysa
					typeStr: "NONSENSE TYPE WILL NEVER EXIST",
				},
			},
		},
	}

	require.Panics(
		t,
		func() { kurtosisCmd.MustGetCobraCommand() },
		"Expected a panic when a flag has a nonsense type",
	)
}

// TODO Add test to verify that no two flags have the same shorthand
// TODO Add test to verify that shorthands are always a single letter
// TODO Add flags to verify that an unrecognized flag type throws a panic
// Add some tests to verify that the flag type assertions work

func TestMustGetCobraCommand_EmptyArgKeyCausesPanic(t *testing.T) {
	kurtosisCmd := &KurtosisCommand{
		CommandStr:       "test",
		ShortDescription: "Short description",
		LongDescription:  "This is a very long description",
		Args: []*ArgConfig{
			{
				Key: "  ",
			},
		},
	}

	require.Panics(
		t,
		func() { kurtosisCmd.MustGetCobraCommand() },
		"Expected a panic when trying to set an arg with an empty key",
	)
}

func TestMustGetCobraCommand_EmptyFlagKeyCausesPanic(t *testing.T) {
	kurtosisCmd := &KurtosisCommand{
		CommandStr:       "test",
		ShortDescription: "Short description",
		LongDescription:  "This is a very long description",
		Flags: []*FlagConfig{
			{
				Key: "  ",
			},
		},
	}

	require.Panics(
		t,
		func() { kurtosisCmd.MustGetCobraCommand() },
		"Expected a panic when trying to set a flag with an empty key",
	)
}

func TestMustGetCobraCommand_TwoOptionalArgumentsCausePanic(t *testing.T) {
	kurtosisCmd := &KurtosisCommand{
		CommandStr:       "test",
		ShortDescription: "Short description",
		LongDescription:  "This is a very long description",
		Args:             []*ArgConfig{
			{
				Key:             arg1Key,
			},
			{
				Key:             arg2Key,
				IsOptional: true,
			},
			{
				Key:             arg3Key,
				IsOptional: true,
			},
		},
	}

	require.Panics(
		t,
		func() { kurtosisCmd.MustGetCobraCommand() },
		"Expected a panic when trying to supply two optional commands, but none happened",
	)
}

func TestMustGetCobraCommand_MiddleGreedyArgCausesPanic(t *testing.T) {
	kurtosisCmd := &KurtosisCommand{
		CommandStr:       "test",
		ShortDescription: "Short description",
		LongDescription:  "This is a very long description",
		Args:             []*ArgConfig{
			{
				Key:             arg1Key,
			},
			{
				Key:             arg2Key,
				IsGreedy: true,
			},
			{
				Key:             arg3Key,
			},
		},
	}

	require.Panics(
		t,
		func() { kurtosisCmd.MustGetCobraCommand() },
		"Expected a panic when trying to supply a greedy argument with another argument after it",
	)
}
