package kurtosis_backend

import (
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/logline"
	"regexp"
)

type LogLineFilterWithRegex struct {
	logline.LogLineFilter
	compiledRegexPattern *regexp.Regexp
}

func newLogLineFilterWithRegex(logLineFilter logline.LogLineFilter, compiledRegexPattern *regexp.Regexp) *LogLineFilterWithRegex {
	return &LogLineFilterWithRegex{LogLineFilter: logLineFilter, compiledRegexPattern: compiledRegexPattern}
}
