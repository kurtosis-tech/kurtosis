package centralized_logs

import "fmt"

type LokiLineFilter struct {
	operator    lokiLineFilterOperator
	textPattern string
}

func NewDoesContainTextLokiLineFilter(text string) *LokiLineFilter {
	operator := lokiLineFilterOperatorDoesContainText
	return &LokiLineFilter{operator: operator, textPattern: text}
}

func NewDoesNotContainTextLokiLineFilter(text string) *LokiLineFilter {
	operator := lokiLineFilterOperatorDoesNotContainText
	return &LokiLineFilter{operator: operator, textPattern: text}
}

//Loki accepts re2 regex syntax type, more here: https://github.com/google/re2/wiki/Syntax
func NewDoesContainMatchRegexLokiLineFilter(regex string) *LokiLineFilter {
	operator := lokiLineFilterOperatorDoesContainMatchRegex
	return &LokiLineFilter{operator: operator, textPattern: regex}
}

//Loki accepts re2 regex syntax type, more here: https://github.com/google/re2/wiki/Syntax
func NewDoesNotContainMatchRegexLokiLineFilter(regex string) *LokiLineFilter {
	operator := lokiLineFilterOperatorDoesNotContainMatchRegex
	return &LokiLineFilter{operator: operator, textPattern: regex}
}

func (lineFilter *LokiLineFilter) String() string {
	return fmt.Sprintf(`%s "%s"`, lineFilter.operator, lineFilter.textPattern)
}
