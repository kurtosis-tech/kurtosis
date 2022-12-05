package centralized_logs

//We weren't able to use github.com/dmarkham/enumer enum system here because we can't do this "LokiLineFilterOperator_|="

type lokiLineFilterOperator int

const (
	// Remember to upgrade allLokiLineFilterOperators var if you add a new value here
	lokiLineFilterOperatorUndefined lokiLineFilterOperator = iota
	lokiLineFilterOperatorContains
	lokiLineFilterOperatorDoesNotContains

	containsValueStr = "|="
	doesNotContainsValueStr = "!="
	unknownValueStr = "unknown"
)

func (filterOperator lokiLineFilterOperator) String() string {
	switch filterOperator {
	case lokiLineFilterOperatorContains:
		return containsValueStr
	case lokiLineFilterOperatorDoesNotContains:
		return doesNotContainsValueStr
	}
	return unknownValueStr
}

func (filterOperator lokiLineFilterOperator) IsDefined() bool {
	return  filterOperator != lokiLineFilterOperatorUndefined
}
