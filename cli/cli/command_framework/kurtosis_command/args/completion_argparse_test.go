package args

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseArgsForCompletion_NoArgsNoTokens(t *testing.T) {
	args := []*ArgConfig{}
	tokens := []string{}
	parsedArgs, argToComplete := ParseArgsForCompletion(args, tokens)
	require.Nil(t, argToComplete)
	require.Equal(t, 0, len(parsedArgs.greedyArgs))
	require.Equal(t, 0, len(parsedArgs.nonGreedyArgs))
}

func TestParseArgsForCompletion_Missing1Token(t *testing.T) {
	args := validArgsConfig[:1]
	tokens := []string{}
	parsedArgs, argToComplete := ParseArgsForCompletion(args, tokens)
	require.Equal(t, validArgsConfig[0], argToComplete)
	require.Equal(t, 0, len(parsedArgs.greedyArgs))
	require.Equal(t, 0, len(parsedArgs.nonGreedyArgs))
}

func TestParseArgsForCompletion_NoTokens(t *testing.T) {
	tokens := []string{}
	parsedArgs, argToComplete := ParseArgsForCompletion(validArgsConfig, tokens)
	require.Equal(t, 0, len(parsedArgs.greedyArgs))
	require.Equal(t, 0, len(parsedArgs.nonGreedyArgs))
	require.Equal(t, validArgsConfig[0], argToComplete)
}

func TestParseArgsForCompletion_NoArgs(t *testing.T) {
	args := []*ArgConfig{}
	parsedArgs, argToComplete := ParseArgsForCompletion(args, validTokens)
	require.Equal(t, 0, len(parsedArgs.greedyArgs))
	require.Equal(t, 0, len(parsedArgs.nonGreedyArgs))
	require.Nil(t, argToComplete)
}

func TestParseArgsForCompletion_ArgsAndTokens(t *testing.T) {
	parsedArgs, argToComplete := ParseArgsForCompletion(validArgsConfig, validTokens)
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
	parsedArgs, argToComplete := ParseArgsForCompletion(validArgsConfig[:2], validTokens)
	require.Nil(t, argToComplete)

	actualArg1Value, err := parsedArgs.GetNonGreedyArg(arg1Key)
	require.NoError(t, err)
	require.Equal(t, arg1Value, actualArg1Value)

	actualArg2Value, err := parsedArgs.GetNonGreedyArg(arg2Key)
	require.NoError(t, err)
	require.Equal(t, arg2Value, actualArg2Value)
}
