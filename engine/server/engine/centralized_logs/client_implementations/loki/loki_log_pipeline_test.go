package loki

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	containTextOperatorStr              = "|="
	doesNotContainTextOperatorStr       = "!="
	containMatchRegexOperatorStr        = "|~"
	doesNotContainMatchRegexOperatorStr = "!~"

	containTextStr              = "contains text"
	doesNotContainTextStr       = "does not contain text"
	containAnotherTextStr       = "contain this other text"
	containsMatchRegexStr       = "start-with.*finish-with"
	doesNotContainMatchRegexStr = `error=\w+`
)

func TestNewValidLokiLogPipeline(t *testing.T) {

	expectLogPipelineStr := fmt.Sprintf(
		`%s "%s" %s "%s" %s "%s" %s "%s" %s "%s"`,
		containTextOperatorStr,
		containTextStr,
		doesNotContainTextOperatorStr,
		doesNotContainTextStr,
		containTextOperatorStr,
		containAnotherTextStr,
		containMatchRegexOperatorStr,
		containsMatchRegexStr,
		doesNotContainMatchRegexOperatorStr,
		doesNotContainMatchRegexStr,
	)

	lineFilterOne := NewDoesContainTextLokiLineFilter(containTextStr)
	lineFilterTwo := NewDoesNotContainTextLokiLineFilter(doesNotContainTextStr)
	lineFilterThree := NewDoesContainTextLokiLineFilter(containAnotherTextStr)
	lineFilterFourth := NewDoesContainMatchRegexLokiLineFilter(containsMatchRegexStr)
	lineFilterFifth := NewDoesNotContainMatchRegexLokiLineFilter(doesNotContainMatchRegexStr)

	lineFilters := []LokiLineFilter{
		*lineFilterOne,
		*lineFilterTwo,
		*lineFilterThree,
		*lineFilterFourth,
		*lineFilterFifth,
	}

	logPipeline := NewLokiLogPipeline(lineFilters)
	require.Equal(t, expectLogPipelineStr, logPipeline.GetConjunctiveLogLineFiltersString())
}

func TestNewLokiLogPipelineWithEmptyFilters(t *testing.T) {

	expectedEmptyLogPipeline := ""

	emptyFilters := []LokiLineFilter{}

	logPipeline := NewLokiLogPipeline(emptyFilters)
	require.Equal(t, expectedEmptyLogPipeline, logPipeline.GetConjunctiveLogLineFiltersString())
}

func TestNewLokiLogPipelineWithNilFilters(t *testing.T) {

	expectedEmptyLogPipeline := ""

	logPipeline := NewLokiLogPipeline(nil)
	require.Equal(t, expectedEmptyLogPipeline, logPipeline.GetConjunctiveLogLineFiltersString())
}
