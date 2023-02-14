package loki

import (
	"strings"
)

const (
	separatorCharacter = " "
)

type lokiLogPipeline struct {
	lineFilters []LokiLineFilter
}

func NewLokiLogPipeline(lineFilters []LokiLineFilter) *lokiLogPipeline {
	return &lokiLogPipeline{lineFilters: lineFilters}
}

func (logPipeline *lokiLogPipeline) GetConjunctiveLogLineFiltersString() string {
	var lineFiltersStr []string
	for _, lineFilter := range logPipeline.lineFilters {
		lineFiltersStr = append(lineFiltersStr, lineFilter.String())
	}
	logPipelineStr := strings.Join(lineFiltersStr, separatorCharacter)

	return logPipelineStr
}
