package logging

import (
	"github.com/sirupsen/logrus"
)


const (
	trace = "trace"
	debug = "debug"
	info = "info"
	warn = "warn"
	error = "error"
	fatal = "fatal"
)

/*
These are the strings forming a pseudo-"enum" of log levels that the Ava E2E tests initializer and the controller will
	accept, with a mapping to the logrus log levels that will be used to actually set the logger's level
*/
var acceptableLogLevels = map[string]logrus.Level{
	trace: logrus.TraceLevel,
	debug: logrus.DebugLevel,
	info:  logrus.InfoLevel,
	warn:  logrus.WarnLevel,
	error: logrus.ErrorLevel,
	fatal: logrus.FatalLevel,
}

/*
Gets a logrus log level from the given string (which likely comes from a CLI argument), verifying that the string is
	in our list of accepted log levels
 */
func LevelFromString(str string) *logrus.Level {
	level, found := acceptableLogLevels[str]
	if !found {
		return nil
	}
	return &level
}

/*
Gets the list of all acceptable strings representing our log levels (used for validation)
 */
func GetAcceptableStrings() []string {
	return []string {
		trace,
		debug,
		info,
		warn,
		error,
		fatal,
	}
}
