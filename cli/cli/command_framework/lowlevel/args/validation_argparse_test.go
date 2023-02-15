package args

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseArgsForValidation_MissingAllTokens(t *testing.T) {
	tokens := []string{}
	_, err := ParseArgsForValidation(validArgsConfig, tokens)
	require.Error(t, err)
}

func TestParseArgsForValidation_Missing2Tokens(t *testing.T) {
	tokens := []string{arg1Value}
	_, err := ParseArgsForValidation(validArgsConfig, tokens)
	require.Error(t, err)
}

func TestParseArgsForValidation_Missing1Token(t *testing.T) {
	tokens := []string{arg1Value, arg2Value}
	_, err := ParseArgsForValidation(validArgsConfig, tokens)
	require.Error(t, err)
}

func TestParseArgsForValidation_AllTokensSupplied(t *testing.T) {
	parsedArgs, err := ParseArgsForValidation(validArgsConfig, validTokens)
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
			Key: arg1Key,
		},
		// This one should never be allowed
		{
			Key:      arg2Key,
			IsGreedy: true,
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
	_, err := ParseArgsForValidation(args, tokens)
	require.Error(t, err)
}

// Technically, the validation that we do in creating the KurtosisCommand should force the optional arg to be last,
//  but even if that validation breaks we should still catch the issue here
func TestParseArgsForValidation_InappropriateOptionalArg(t *testing.T) {
	args := []*ArgConfig{
		{
			Key: arg1Key,
		},
		// This one should never be allowed
		{
			Key:        arg2Key,
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
	_, err := ParseArgsForValidation(args, tokens)
	require.Error(t, err)
}

func TestParseArgsForValidation_OptionalArgNoTokens(t *testing.T) {
	defaultValue := "NON GREEDY DEFAULT"
	args := []*ArgConfig{
		{
			Key:          arg1Key,
			IsOptional:   true,
			DefaultValue: defaultValue,
		},
	}
	tokens := []string{}

	parsedArgs, err := ParseArgsForValidation(args, tokens)
	require.NoError(t, err)

	value, err := parsedArgs.GetNonGreedyArg(arg1Key)
	require.NoError(t, err)
	require.Equal(t, defaultValue, value)
}

func TestParseArgsForValidation_OptionalArgWithToken(t *testing.T) {
	defaultValue := "NON GREEDY DEFAULT"
	args := []*ArgConfig{
		{
			Key:          arg1Key,
			IsOptional:   true,
			DefaultValue: defaultValue,
		},
	}
	suppliedValue := "SUPPLIED VALUE"
	tokens := []string{
		suppliedValue,
	}

	parsedArgs, err := ParseArgsForValidation(args, tokens)
	require.NoError(t, err)

	value, err := parsedArgs.GetNonGreedyArg(arg1Key)
	require.NoError(t, err)
	require.Equal(t, suppliedValue, value)
}
