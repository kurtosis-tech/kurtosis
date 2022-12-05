package centralized_logs

import (
	"fmt"
)

type lokiLogPipeline struct {
	lineFilters []*LokiLineFilter
}

func NewLokiLogPipeline(lineFilters []*LokiLineFilter) *lokiLogPipeline {
	return &lokiLogPipeline{lineFilters: lineFilters}
}

func (logPipeline *lokiLogPipeline) PipeLineStringify() string{
	var logPipelineStr string
	for _, lineFilter := range logPipeline.lineFilters {
		logPipelineStr = fmt.Sprintf("%s %s",logPipelineStr, lineFilter)
	}
	return logPipelineStr
}
