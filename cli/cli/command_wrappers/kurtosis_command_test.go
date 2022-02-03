package command_wrappers

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
)

var validArgsConfig = []*ArgConfig{
	{
		Key:             arg1Key,
	},
	{
		Key:             arg2Key,
	},
	{
		Key:             arg3Key,
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

// TODO test that two optional arguments throw an error
// TODO test that arguments after a greedy arg throws an error
// TODO verify that a couple normal arguments don't throw an error

func TestParseArgsForCompletion_NoArgsNoTokens(t *testing.T) {
	args := []*ArgConfig{}
	tokens := []string{}
	parsedArgs, argToComplete := parseArgsForCompletion(args, tokens)
	require.Nil(t, argToComplete)
	require.Equal(t, 0, len(parsedArgs.greedyArgs))
	require.Equal(t, 0, len(parsedArgs.nonGreedyArgs))
}

func TestParseArgsForCompletion_Missing1Token(t *testing.T) {
	args := validArgsConfig[:1]
	tokens := []string{}
	parsedArgs, argToComplete := parseArgsForCompletion(args, tokens)
	require.Equal(t, validArgsConfig[0], argToComplete)
	require.Equal(t, 0, len(parsedArgs.greedyArgs))
	require.Equal(t, 0, len(parsedArgs.nonGreedyArgs))
}

func TestParseArgsForCompletion_NoTokens(t *testing.T) {
	tokens := []string{}
	parsedArgs, argToComplete := parseArgsForCompletion(validArgsConfig, tokens)
	require.Equal(t, 0, len(parsedArgs.greedyArgs))
	require.Equal(t, 0, len(parsedArgs.nonGreedyArgs))
	require.Equal(t, validArgsConfig[0], argToComplete)
}

func TestParseArgsForCompletion_NoArgs(t *testing.T) {
	args := []*ArgConfig{}
	parsedArgs, argToComplete := parseArgsForCompletion(args, validTokens)
	require.Equal(t, 0, len(parsedArgs.greedyArgs))
	require.Equal(t, 0, len(parsedArgs.nonGreedyArgs))
	require.Nil(t, argToComplete)
}

func TestParseArgsForCompletion_ArgsAndTokens(t *testing.T) {
	parsedArgs, argToComplete := parseArgsForCompletion(validArgsConfig, validTokens)
	// Because the last arg is greedy, we should get completion for any more the user wants to add
	require.Equal(t, validArgsConfig[2], argToComplete)

	actualArg1Value, err := parsedArgs.GetNonGreedyArg(arg1Key)
	require.NoError(t, err)
	require.Equal(t, arg1Value, actualArg1Value)

	actualArg2Value, err := parsedArgs.GetNonGreedyArg(arg2Key)
	require.NoError(t, err)
	require.Equal(t, arg2Value, actualArg2Value)

	actualArg3Values, err := parsedArgs.GetGreedyArg(arg3Key)
	require.NoError(t, err)
	require.Equal(t, 3, len(actualArg3Values))
	require.Equal(t, arg3Value1, actualArg3Values[0])
	require.Equal(t, arg3Value2, actualArg3Values[1])
	require.Equal(t, arg3Value3, actualArg3Values[2])
}

func TestParseArgsForCompletion_NonGreedyLastArgMeansNoCompletion(t *testing.T) {
	parsedArgs, argToComplete := parseArgsForCompletion(validArgsConfig[:2], validTokens)
	require.Nil(t, argToComplete)

	actualArg1Value, err := parsedArgs.GetNonGreedyArg(arg1Key)
	require.NoError(t, err)
	require.Equal(t, arg1Value, actualArg1Value)

	actualArg2Value, err := parsedArgs.GetNonGreedyArg(arg2Key)
	require.NoError(t, err)
	require.Equal(t, arg2Value, actualArg2Value)
}

func TestParseArgsForValidation_MissingAllTokens(t *testing.T) {
	tokens := []string{}
	_, err := parseArgsForValidation(validArgsConfig, tokens)
	require.Error(t, err)
}

func TestParseArgsForValidation_Missing2Tokens(t *testing.T) {
	tokens := []string{arg1Value}
	_, err := parseArgsForValidation(validArgsConfig, tokens)
	require.Error(t, err)
}

func TestParseArgsForValidation_Missing1Token(t *testing.T) {
	tokens := []string{arg1Value, arg2Value}
	_, err := parseArgsForValidation(validArgsConfig, tokens)
	require.Error(t, err)
}

func TestParseArgsForValidation_AllTokensSupplied(t *testing.T) {
	parsedArgs, err := parseArgsForValidation(validArgsConfig, validTokens)
	require.NoError(t, err)

	actualArg1Value, err := parsedArgs.GetNonGreedyArg(arg1Key)
	require.NoError(t, err)
	require.Equal(t, arg1Value, actualArg1Value)

	actualArg2Value, err := parsedArgs.GetNonGreedyArg(arg2Key)
	require.NoError(t, err)
	require.Equal(t, arg2Value, actualArg2Value)

	actualArg3Values, err := parsedArgs.GetGreedyArg(arg3Key)
	require.NoError(t, err)
	require.Equal(t, 3, len(actualArg3Values))
	require.Equal(t, arg3Value1, actualArg3Values[0])
	require.Equal(t, arg3Value2, actualArg3Values[1])
	require.Equal(t, arg3Value3, actualArg3Values[2])
}

// Technically, the validation that we do in creating the KurtosisCommand should force the greedy arg to be last,
//  but even if that validation breaks we should still catch the issue here
func TestParseArgsForValidation_InappropriateGreedyArg(t *testing.T) {
	args := []*ArgConfig{
		{
			Key:             arg1Key,
		},
		// This one should never be allowed
		{
			Key:             arg2Key,
			IsGreedy:        true,
		},
		{
			Key: arg3Key,
		},
	}
	tokens := []string{
		"1",
		"2",
		"3",
		"4",
		"5",
	}
	_, err := parseArgsForValidation(args, tokens)
	require.Error(t, err)
}

// Technically, the validation that we do in creating the KurtosisCommand should force the optional arg to be last,
//  but even if that validation breaks we should still catch the issue here
func TestParseArgsForValidation_InappropriateOptionalArg(t *testing.T) {
	args := []*ArgConfig{
		{
			Key:             arg1Key,
		},
		// This one should never be allowed
		{
			Key:             arg2Key,
			IsOptional: true,
		},
		{
			Key: arg3Key,
		},
	}
	tokens := []string{
		"1",
		"2",
		"3",
	}
	_, err := parseArgsForValidation(args, tokens)
	require.Error(t, err)
}

func TestParseArgsForValidation_TestOptionalArgNoTokens(t *testing.T) {
	defaultValue := "NON GREEDY DEFAULT"
	args := []*ArgConfig{
		{
			Key: arg1Key,
			IsOptional: true,
			DefaultValue: defaultValue,
		},
	}
	tokens := []string{}

	parsedArgs, err := parseArgsForValidation(args, tokens)
	require.NoError(t, err)

	value, err := parsedArgs.GetNonGreedyArg(arg1Key)
	require.NoError(t, err)
	require.Equal(t, defaultValue, value)
}

func TestParseArgsForValidation_TestOptionalArgWithToken(t *testing.T) {
	defaultValue := "NON GREEDY DEFAULT"
	args := []*ArgConfig{
		{
			Key: arg1Key,
			IsOptional: true,
			DefaultValue: defaultValue,
		},
	}
	suppliedValue := "SUPPLIED VALUE"
	tokens := []string{
		suppliedValue,
	}

	parsedArgs, err := parseArgsForValidation(args, tokens)
	require.NoError(t, err)

	value, err := parsedArgs.GetNonGreedyArg(arg1Key)
	require.NoError(t, err)
	require.Equal(t, suppliedValue, value)
}

