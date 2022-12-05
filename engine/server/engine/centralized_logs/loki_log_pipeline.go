package centralized_logs

import (
	"strings"
)

type lokiLogPipeline struct {
	lineFilters []*LokiLineFilter
}

func NewLokiLogPipeline(lineFilters []*LokiLineFilter) *lokiLogPipeline {
	return &lokiLogPipeline{lineFilters: lineFilters}
}

func (logPipeline *lokiLogPipeline) PipelineStringify() string{
	var lineFiltersStr []string
	for _, lineFilter := range logPipeline.lineFilters {
		lineFiltersStr = append(lineFiltersStr, lineFilter.String())
	}
	logPipelineStr := strings.Join(lineFiltersStr, " ")

	return logPipelineStr
}
