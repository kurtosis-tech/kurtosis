package centralized_logs

import "fmt"

type LokiLineFilter struct {
	operator LokiLineFilterOperator
	text string
}

func NewLokiLineFilter(operator LokiLineFilterOperator, text string) *LokiLineFilter {
	return &LokiLineFilter{operator: operator, text: text}
}

func (lineFilter *LokiLineFilter) GetText() string {
	return lineFilter.text
}

func (lineFilter *LokiLineFilter) GetOperator() LokiLineFilterOperator {
	return lineFilter.operator
}

func (lineFilter *LokiLineFilter) String() string {
	return fmt.Sprintf(`%s "%s"`, lineFilter.operator, lineFilter.text)
}
