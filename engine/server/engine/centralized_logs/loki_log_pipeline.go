package centralized_logs

import (
	"fmt"
	"github.com/kurtosis-tech/stacktrace"
)

type lokiLogPipeline struct {
	lineFilters []*lokiLineFilter
}

func NewLokiLogPipeline(lineFilters []*lokiLineFilter) (*lokiLogPipeline, error) {
	for lineFilterIndex, lineFilter := range lineFilters{
		lineFilterOperator := lineFilter.GetOperator()
		if !lineFilterOperator.IsDefined(){
			lineFilterPosition := lineFilterIndex + 1
			return nil, stacktrace.NewError(
				"New Loki log line with line filters '%+v' can't be created because the operator '%v' is not defined for the line filter in position '%v' and with text '%v'",
				lineFilters,
				lineFilterOperator,
				lineFilterPosition,
				lineFilter.GetText(),
			)
		}
	}

	return &lokiLogPipeline{lineFilters: lineFilters}, nil
}

func (logPipeline *lokiLogPipeline) PipeLineStringify() string{
	var logPipelineStr string
	for _, lineFilter := range logPipeline.lineFilters {
		logPipelineStr = fmt.Sprintf("%s %s",logPipelineStr, lineFilter)
	}
	return logPipelineStr
}
