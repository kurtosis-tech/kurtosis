package logs

import (
	"github.com/dzobbe/PoTE-kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDefiningLogLineFilterFromFlags_doNotFilter(t *testing.T) {
	matchTextStr := ""
	matchRegexStr := ""
	invertMatch := false

	logLineFilter, err := getLogLineFilterFromFilterFlagValues(matchTextStr, matchRegexStr, invertMatch)
	require.Nil(t, logLineFilter)
	require.Nil(t, err)
}

func TestDefiningLogLineFilterFromFlags_isNotAllowedSendBothMatch(t *testing.T) {
	matchTextStr := "something"
	matchRegexStr := "something else"
	invertMatch := false
	errContainStr := "Both filter match have being sent"

	logLineFilter, err := getLogLineFilterFromFilterFlagValues(matchTextStr, matchRegexStr, invertMatch)
	require.Nil(t, logLineFilter)
	require.Error(t, err)
	require.ErrorContains(t, err, errContainStr)
}

func TestDefiningLogLineFilterFromFlags_validMatchText(t *testing.T) {
	matchTextStr := "something"
	matchRegexStr := ""
	invertMatch := false
	expectedLogLineFilter := kurtosis_context.NewDoesContainTextLogLineFilter(matchTextStr)

	logLineFilter, err := getLogLineFilterFromFilterFlagValues(matchTextStr, matchRegexStr, invertMatch)
	require.NotNil(t, logLineFilter)
	require.NoError(t, err)
	require.Equal(t, expectedLogLineFilter, logLineFilter)
}

func TestDefiningLogLineFilterFromFlags_validInvertMatchText(t *testing.T) {
	matchTextStr := "something"
	matchRegexStr := ""
	invertMatch := true
	expectedLogLineFilter := kurtosis_context.NewDoesNotContainTextLogLineFilter(matchTextStr)

	logLineFilter, err := getLogLineFilterFromFilterFlagValues(matchTextStr, matchRegexStr, invertMatch)
	require.NotNil(t, logLineFilter)
	require.NoError(t, err)
	require.Equal(t, expectedLogLineFilter, logLineFilter)
}

func TestDefiningLogLineFilterFromFlags_validMatchRegex(t *testing.T) {
	matchTextStr := ""
	matchRegexStr := "my.*regex"
	invertMatch := false
	expectedLogLineFilter := kurtosis_context.NewDoesContainMatchRegexLogLineFilter(matchRegexStr)

	logLineFilter, err := getLogLineFilterFromFilterFlagValues(matchTextStr, matchRegexStr, invertMatch)
	require.NotNil(t, logLineFilter)
	require.NoError(t, err)
	require.Equal(t, expectedLogLineFilter, logLineFilter)
}

func TestDefiningLogLineFilterFromFlags_validInvertMatchRegex(t *testing.T) {
	matchTextStr := ""
	matchRegexStr := "my.*regex"
	invertMatch := true
	expectedLogLineFilter := kurtosis_context.NewDoesNotContainMatchRegexLogLineFilter(matchRegexStr)

	logLineFilter, err := getLogLineFilterFromFilterFlagValues(matchTextStr, matchRegexStr, invertMatch)
	require.NotNil(t, logLineFilter)
	require.NoError(t, err)
	require.Equal(t, expectedLogLineFilter, logLineFilter)
}
