package loki

//We weren't able to use github.com/dmarkham/enumer enum system here because we can't do this "LokiLineFilterOperator_|="

type lokiLineFilterOperator int

const (
	lokiLineFilterOperatorUndefined lokiLineFilterOperator = iota
	lokiLineFilterOperatorDoesContainText
	lokiLineFilterOperatorDoesNotContainText
	lokiLineFilterOperatorDoesContainMatchRegex
	lokiLineFilterOperatorDoesNotContainMatchRegex

	//These magic string are Loki's operators, you can read more about it here: https://grafana.com/docs/loki/latest/logql/log_queries/
	doesContainTextLokiOperatorStr          = "|="
	doesNotContainTextLokiOperatorStr       = "!="
	doesContainMatchRegexLokiOperatorStr    = "|~"
	doesNotContainMatchRegexLokiOperatorStr = "!~"
	unknownValueStr                         = "unknown"
)

func (filterOperator lokiLineFilterOperator) String() string {
	switch filterOperator {
	case lokiLineFilterOperatorDoesContainText:
		return doesContainTextLokiOperatorStr
	case lokiLineFilterOperatorDoesNotContainText:
		return doesNotContainTextLokiOperatorStr
	case lokiLineFilterOperatorDoesContainMatchRegex:
		return doesContainMatchRegexLokiOperatorStr
	case lokiLineFilterOperatorDoesNotContainMatchRegex:
		return doesNotContainMatchRegexLokiOperatorStr
	}
	return unknownValueStr
}

func (filterOperator lokiLineFilterOperator) IsDefined() bool {
	return filterOperator != lokiLineFilterOperatorUndefined
}
