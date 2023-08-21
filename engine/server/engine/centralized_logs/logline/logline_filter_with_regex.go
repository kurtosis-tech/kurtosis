package logline

import (
	"github.com/kurtosis-tech/stacktrace"
	"regexp"
)

type LogLineFilterWithRegex struct {
	LogLineFilter
	compiledRegexPattern *regexp.Regexp
}

func NewLogLineFilterWithRegex(logLineFilter LogLineFilter, compiledRegexPattern *regexp.Regexp) *LogLineFilterWithRegex {
	return &LogLineFilterWithRegex{LogLineFilter: logLineFilter, compiledRegexPattern: compiledRegexPattern}
}

func NewConjunctiveLogFiltersWithRegex(conjunctiveLogLineFilters ConjunctiveLogLineFilters) ([]LogLineFilterWithRegex, error) {
	conjunctiveLogFiltersWithRegex := []LogLineFilterWithRegex{}
	for _, logLineFilter := range conjunctiveLogLineFilters {
		logLineFilterWithRegex := NewLogLineFilterWithRegex(logLineFilter, nil)

		if logLineFilter.IsRegexFilter() {
			filterRegexPattern := logLineFilter.GetTextPattern()
			logLineRegexPattern, err := regexp.Compile(filterRegexPattern)
			if err != nil {
				return nil, stacktrace.Propagate(err, "An error occurred compiling regex string '%v' for log line filter '%+v'", filterRegexPattern, logLineFilter)
			}
			logLineFilterWithRegex.compiledRegexPattern = logLineRegexPattern
		}
		conjunctiveLogFiltersWithRegex = append(conjunctiveLogFiltersWithRegex, *logLineFilterWithRegex)
	}

	return conjunctiveLogFiltersWithRegex, nil
}
