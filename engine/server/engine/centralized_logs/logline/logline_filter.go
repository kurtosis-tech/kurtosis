package logline

type ConjunctiveLogLineFilters []LogLineFilter

type LogLineFilter struct {
	operator    logLineOperator
	textPattern string
}

func NewDoesContainTextLogLineFilter(text string) *LogLineFilter {
	return &LogLineFilter{operator: LogLineOperator_DoesContainText, textPattern: text}
}

func NewDoesNotContainTextLogLineFilter(text string) *LogLineFilter {
	return &LogLineFilter{operator: LogLineOperator_DoesNotContainText, textPattern: text}
}

func NewDoesContainMatchRegexLogLineFilter(regex string) *LogLineFilter {
	return &LogLineFilter{operator: LogLineOperator_DoesContainMatchRegex, textPattern: regex}
}

func NewDoesNotContainMatchRegexLogLineFilter(regex string) *LogLineFilter {
	return &LogLineFilter{operator: LogLineOperator_DoesNotContainMatchRegex, textPattern: regex}
}

func (logLineFilter *LogLineFilter) GetOperator() logLineOperator {
	return logLineFilter.operator
}

func (logLineFilter *LogLineFilter) GetTextPattern() string {
	return logLineFilter.textPattern
}

func (logLineFilter *LogLineFilter) IsRegexFilter() bool {
	return logLineFilter.operator == LogLineOperator_DoesContainMatchRegex || logLineFilter.operator == LogLineOperator_DoesNotContainMatchRegex
}
