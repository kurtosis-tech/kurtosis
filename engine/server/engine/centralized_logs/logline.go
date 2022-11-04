package centralized_logs

type LogLine struct{
	//lineTime time.Time //TODO add the time from loki logs result
	content string
}

func newLogLine(content string) *LogLine {
	return &LogLine{content: content}
}

func (logLine LogLine) GetContent() string {
	return logLine.content
}
