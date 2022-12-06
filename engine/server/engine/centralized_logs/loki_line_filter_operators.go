package centralized_logs

//We weren't able to use github.com/dmarkham/enumer enum system here because we can't do this "LokiLineFilterOperator_|="

type lokiLineFilterOperator int

const (
	// Remember to upgrade allLokiLineFilterOperators var if you add a new value here
	lokiLineFilterOperatorUndefined lokiLineFilterOperator = iota
	lokiLineFilterOperatorDoesContain
	lokiLineFilterOperatorDoesNotContain

	//These magic string are Loki's operators, you can read more about it here: https://grafana.com/docs/loki/latest/logql/log_queries/
	doesContainValueStr    = "|="
	doesNotContainValueStr = "!="
	unknownValueStr        = "unknown"
)

func (filterOperator lokiLineFilterOperator) String() string {
	switch filterOperator {
	case lokiLineFilterOperatorDoesContain:
		return doesContainValueStr
	case lokiLineFilterOperatorDoesNotContain:
		return doesNotContainValueStr
	}
	return unknownValueStr
}

func (filterOperator lokiLineFilterOperator) IsDefined() bool {
	return filterOperator != lokiLineFilterOperatorUndefined
}
