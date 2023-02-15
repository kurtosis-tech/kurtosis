package logline

import "strings"

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
