package logline

import (
	"github.com/kurtosis-tech/stacktrace"
	"strings"
)

const (
	newlineChar = "\n"
)

type LogLine struct {
	//lineTime time.Time //TODO add the time from loki logs result
	content string
}

func NewLogLine(content string) *LogLine {
	contentWithoutNewLine := strings.TrimSuffix(content, newlineChar)
	return &LogLine{content: contentWithoutNewLine}
}

func (logLine LogLine) GetContent() string {
	return logLine.content
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
