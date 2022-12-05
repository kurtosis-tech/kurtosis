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

	expectLogPipelineStr := fmt.Sprintf(
		`|= "%s" != "%s" |= "%s"`,
		containTextStr,
		doesNotContainTextStr,
		containAnotherTextStr,
	)

	lineFilterOne := NewDoesContainLokiLineFilter(containTextStr)
	lineFilterTwo := NewDoesNotContainLokiLineFilter(doesNotContainTextStr)
	lineFilterThree := NewDoesContainLokiLineFilter(containAnotherTextStr)

	lineFilters := []*LokiLineFilter{
		lineFilterOne,
		lineFilterTwo,
		lineFilterThree,
	}

	logPipeline := NewLokiLogPipeline(lineFilters)
	require.Equal(t, expectLogPipelineStr, logPipeline.PipelineStringify())
}
