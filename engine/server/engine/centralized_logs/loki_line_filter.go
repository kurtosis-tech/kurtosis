package centralized_logs

import "fmt"

type LokiLineFilter struct {
	operator lokiLineFilterOperator
	text     string
}

func NewDoesContainLokiLineFilter(text string) *LokiLineFilter {
	operator := lokiLineFilterOperatorDoesContain
	return &LokiLineFilter{operator: operator, text: text}
}

func NewDoesNotContainLokiLineFilter(text string) *LokiLineFilter {
	operator := lokiLineFilterOperatorDoesNotContain
	return &LokiLineFilter{operator: operator, text: text}
}

func (lineFilter *LokiLineFilter) String() string {
	return fmt.Sprintf(`%s "%s"`, lineFilter.operator, lineFilter.text)
}
