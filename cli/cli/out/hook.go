package out

import (
	"fmt"
	"io"

	"github.com/sirupsen/logrus"
)

const logFileName = "kurtosis-cli.log"

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

// Fire This hook is called whenever default logrus is to log messages; therefore
// cannot use it logrus default logger in this method to avoid recursive behaviour
// therefore using logrus.StandardLogger to log failures with this method onto the
// terminal only during debug level.
func (hook *Hook) Fire(entry *logrus.Entry) error {
	line, err := hook.formatter.Format(entry)
	if err != nil {
		logrus.StandardLogger().Debug(fmt.Sprintf("Error occurred while formatting log message for: %+v", entry))
		return nil
	}
	_, err = hook.writer.Write(line)
	if err != nil {
		logrus.StandardLogger().Debug(fmt.Sprintf("Error occurred writing the log message to the file: %v for: %+v", logFileName, entry))
	}
	return nil
}

// Levels define on which log levels this hook would trigger
func (hook *Hook) Levels() []logrus.Level {
	return hook.logLevels
}
