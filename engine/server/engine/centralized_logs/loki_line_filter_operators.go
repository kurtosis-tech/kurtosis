package centralized_logs

//We weren't able to use github.com/dmarkham/enumer enum system here because we can't do this "LokiLineFilterOperator_|="

type LokiLineFilterOperator int

const (
	// Remember to upgrade allLoLineFilterOperators var if you add a new value here
	lokiLineFilterOperatorUndefined LokiLineFilterOperator = iota
	LokiLineFilterOperatorContains
	LokiLineFilterOperatorDoesNotContains

	containsValueStr = "|="
	doesNotContainsValueStr = "!="
	unknownValueStr = "unknown"
)

var allLoLineFilterOperators = map[LokiLineFilterOperator]bool{
	LokiLineFilterOperatorContains: true,
	LokiLineFilterOperatorDoesNotContains: true,
}

func (filterOperator LokiLineFilterOperator) String() string {
	switch filterOperator {
	case LokiLineFilterOperatorContains:
		return containsValueStr
	case LokiLineFilterOperatorDoesNotContains:
		return doesNotContainsValueStr
	}
	return unknownValueStr
}

func (filterOperator LokiLineFilterOperator) IsDefined() bool {
	return  filterOperator != lokiLineFilterOperatorUndefined
}
