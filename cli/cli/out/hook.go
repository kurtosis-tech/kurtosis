package out

import (
	"io"

	log "github.com/sirupsen/logrus"
)

// Hook is a hook that writes logs of specified LogLevels to specified Writer
type Hook struct {
	Writer    io.Writer
	LogLevels []log.Level
	Formatter log.Formatter
}

func (hook *Hook) Fire(entry *log.Entry) error {
	line, err := hook.Formatter.Format(entry)
	if err != nil {
		return err
	}
	_, err = hook.Writer.Write(line)
	return err
}

// Levels define on which log levels this hook would trigger
func (hook *Hook) Levels() []log.Level {
	return hook.LogLevels
}
