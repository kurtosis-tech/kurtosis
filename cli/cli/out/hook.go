package out

import (
	"fmt"
	"io"

	log "github.com/sirupsen/logrus"
)

const logFileName = "kurtosis-cli.log"

// Hook is a hook that writes logs of specified LogLevels to specified Writer
type Hook struct {
	writer    io.Writer
	logLevels []log.Level
	formatter log.Formatter
}

func NewHook(writer io.Writer, logLevels []log.Level, formatter log.Formatter) Hook {
	return Hook{
		writer:    writer,
		logLevels: logLevels,
		formatter: formatter,
	}
}

func (hook *Hook) Fire(entry *log.Entry) error {
	line, err := hook.formatter.Format(entry)
	if err != nil {
		GetFileLogger().Debug(fmt.Sprintf("Error occurred while formatting log message for: %+v", entry))
		return nil
	}
	_, err = hook.writer.Write(line)
	if err != nil {
		GetFileLogger().Debug(fmt.Sprintf("Error occurred writing the log message to the file: %v for: %+v", logFileName, entry))
	}
	return nil
}

// Levels define on which log levels this hook would trigger
func (hook *Hook) Levels() []log.Level {
	return hook.logLevels
}
