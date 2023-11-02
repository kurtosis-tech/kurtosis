package logline

import (
	"github.com/kurtosis-tech/stacktrace"
	"strings"
	"time"
)

const (
	newlineChar = "\n"
)

type LogLine struct {
	content string

	timestamp time.Time
}

func NewLogLine(content string, timestamp time.Time) *LogLine {
	contentWithoutNewLine := strings.TrimSuffix(content, newlineChar)
	return &LogLine{content: contentWithoutNewLine, timestamp: timestamp}
}

func (logLine LogLine) GetContent() string {
	return logLine.content
}

func (logLine LogLine) GetTimestamp() time.Time {
	return logLine.timestamp
}

func (logLine LogLine) IsValidLogLineBaseOnFilters(
	conjunctiveLogLinesFiltersWithRegex []LogLineFilterWithRegex,
) (bool, error) {

	shouldReturnIt := true

	for _, logLineFilter := range conjunctiveLogLinesFiltersWithRegex {
		operator := logLineFilter.GetOperator()

		logLineContent := logLine.GetContent()
		logLineContentLowerCase := strings.ToLower(logLineContent)
		textPatternLowerCase := strings.ToLower(logLineFilter.GetTextPattern())

		switch operator {
		case LogLineOperator_DoesContainText:
			if !strings.Contains(logLineContentLowerCase, textPatternLowerCase) {
				shouldReturnIt = false
			}
		case LogLineOperator_DoesNotContainText:
			if strings.Contains(logLineContentLowerCase, textPatternLowerCase) {
				shouldReturnIt = false
			}
		case LogLineOperator_DoesContainMatchRegex:
			if !logLineFilter.compiledRegexPattern.MatchString(logLineContent) {
				shouldReturnIt = false
			}
		case LogLineOperator_DoesNotContainMatchRegex:
			if logLineFilter.compiledRegexPattern.MatchString(logLineContent) {
				shouldReturnIt = false
			}
		default:
			return false, stacktrace.NewError("Unrecognized log line filter operator '%v' in filter '%v'; this is a bug in Kurtosis", operator, logLineFilter)
		}
		if !shouldReturnIt {
			break
		}
	}

	return shouldReturnIt, nil
}
