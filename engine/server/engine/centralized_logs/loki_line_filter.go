package centralized_logs

import "fmt"

type LokiLineFilter struct {
	operator lokiLineFilterOperator
	text     string
}

func NewDoesContainLokiLineFilter(text string) *LokiLineFilter {
	operator := lokiLineFilterOperatorContains
	return &LokiLineFilter{operator: operator, text: text}
}

func NewDoesNotContainLokiLineFilter(text string) *LokiLineFilter {
	operator := lokiLineFilterOperatorDoesNotContains
	return &LokiLineFilter{operator: operator, text: text}
}


func (lineFilter *LokiLineFilter) GetText() string {
	return lineFilter.text
}

func (lineFilter *LokiLineFilter) GetOperator() lokiLineFilterOperator {
	return lineFilter.operator
}

func (lineFilter *LokiLineFilter) String() string {
	return fmt.Sprintf(`%s "%s"`, lineFilter.operator, lineFilter.text)
}
