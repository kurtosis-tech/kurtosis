package kurtosis_command

import (
	"github.com/stretchr/testify/require"
	"testing"
)

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
				Key: flag1Key,
			},
		},
	}

	require.Panics(
		t,
		func() { kurtosisCmd.MustGetCobraCommand() },
		"Expected a panic when trying to supply two flags with the same key",
	)
}

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
