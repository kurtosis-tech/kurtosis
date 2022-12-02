package centralized_logs

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	containTextStr = "contains text"
	doesNotContainTextStr = "does not contain text"
	containAnotherTextStr = "contain this other text"
)

func TestNewValidLokiLogPipeline(t *testing.T) {

	expectLogPipeLineStr := fmt.Sprintf(
		` |= "%s" != "%s" |= "%s"`,
		containTextStr,
		doesNotContainTextStr,
		containAnotherTextStr,
	)

	lineFilterOne := NewLokiLineFilter(LokiLineFilterOperatorContains, containTextStr)
	lineFilterTwo := NewLokiLineFilter(LokiLineFilterOperatorDoesNotContains, doesNotContainTextStr)
	lineFilterThree := NewLokiLineFilter(LokiLineFilterOperatorContains, containAnotherTextStr)

	lineFilters := []*lokiLineFilter{
		lineFilterOne,
		lineFilterTwo,
		lineFilterThree,
	}

	logPipeLine, err := NewLokiLogPipeline(lineFilters)
	require.NoError(t, err)
	require.Equal(t, expectLogPipeLineStr, logPipeLine.String())
}

func TestNewNotValidLokiLogPipeline(t *testing.T) {

	var undefinedLogLineOperator LokiLineFilterOperator

	lineFilterOne := NewLokiLineFilter(LokiLineFilterOperatorContains, containTextStr)
	lineFilterTwo := NewLokiLineFilter(undefinedLogLineOperator, doesNotContainTextStr)

	lineFilters := []*lokiLineFilter{
		lineFilterOne,
		lineFilterTwo,
	}

	_, err := NewLokiLogPipeline(lineFilters)
	require.Error(t, err)
}
