package centralized_logs

import "fmt"

type lokiLineFilter struct {
	operator LokiLineFilterOperator
	text string
}

func NewLokiLineFilter(operator LokiLineFilterOperator, text string) *lokiLineFilter {
	return &lokiLineFilter{operator: operator, text: text}
}

func (lineFilter *lokiLineFilter) GetText() string {
	return lineFilter.text
}

func (lineFilter *lokiLineFilter) GetOperator() LokiLineFilterOperator {
	return lineFilter.operator
}

func (lineFilter *lokiLineFilter) String() string {
	return fmt.Sprintf(`%s "%s"`, lineFilter.operator, lineFilter.text)
}
