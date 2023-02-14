package kurtosis_context

type LogLineFilter struct {
	operator    logLineOperator
	textPattern string
}

func NewDoesContainTextLogLineFilter(text string) *LogLineFilter {
	return &LogLineFilter{operator: logLineOperator_DoesContainText, textPattern: text}
}

func NewDoesNotContainTextLogLineFilter(text string) *LogLineFilter {
	return &LogLineFilter{operator: logLineOperator_DoesNotContainText, textPattern: text}
}

func NewDoesContainMatchRegexLogLineFilter(regex string) *LogLineFilter {
	return &LogLineFilter{operator: logLineOperator_DoesContainMatchRegex, textPattern: regex}
}

func NewDoesNotContainMatchRegexLogLineFilter(regex string) *LogLineFilter {
	return &LogLineFilter{operator: logLineOperator_DoesNotContainMatchRegex, textPattern: regex}
}
