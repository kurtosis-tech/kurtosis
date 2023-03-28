package out

import (
	"fmt"
	"io"

	"github.com/sirupsen/logrus"
)

const logFileName = "kurtosis-cli.log"

// this is needed to silently log fire failures to the cli on --debug level
// the default logger has the hook attached to it. Using the same logger in the fire method
// for the hook would result in recursive condition
var silentFailureLogger = logrus.New()

// Hook is a hook that writes logs of specified LogLevels to specified Writer
type Hook struct {
	writer    io.Writer
	logLevels []logrus.Level
	formatter logrus.Formatter
}

func NewHook(writer io.Writer, logLevels []logrus.Level, formatter logrus.Formatter) Hook {
	return Hook{
		writer:    writer,
		logLevels: logLevels,
		formatter: formatter,
	}
}

func (hook *Hook) Fire(entry *logrus.Entry) error {
	line, err := hook.formatter.Format(entry)
	if err != nil {
		silentFailureLogger.Debug(fmt.Sprintf("Error occurred while formatting log message for: %+v", entry))
		return nil
	}
	_, err = hook.writer.Write(line)
	if err != nil {
		silentFailureLogger.Debug(fmt.Sprintf("Error occurred writing the log message to the file: %v for: %+v", logFileName, entry))
	}
	return nil
}

// Levels define on which log levels this hook would trigger
func (hook *Hook) Levels() []logrus.Level {
	return hook.logLevels
}
