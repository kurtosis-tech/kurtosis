/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package logrus_log_levels

import (
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"sort"
)


const (
	Trace = "trace"
	Debug = "debug"
	Info  = "info"
	Warn  = "warn"
	Error = "error"
	Fatal = "fatal"
)

/*
These are the strings forming a pseudo-"enum" of log levels that the Ava E2E tests initializer and the controller will
	accept, with a mapping to the logrus log levels that will be used to actually set the logger's level
*/
var acceptableLogLevels = map[string]logrus.Level{
	Trace: logrus.TraceLevel,
	Debug: logrus.DebugLevel,
	Info:  logrus.InfoLevel,
	Warn:  logrus.WarnLevel,
	Error: logrus.ErrorLevel,
	Fatal: logrus.FatalLevel,
}

/*
Gets a logrus log level from the given string (which likely comes from a CLI argument), verifying that the string is
	in our list of accepted log levels
 */
func LevelFromString(str string) (logrus.Level, error) {
	level, found := acceptableLogLevels[str]
	if !found {
		return 0, stacktrace.NewError("No log level found for string '%v'", str)
	}
	return level, nil
}

/*
Gets the list of all acceptable strings representing our log levels (used for validation)
 */
func GetAcceptableStrings() []string {
	result := []string{}
	for levelStr, _ := range acceptableLogLevels {
		result = append(result, levelStr)
	}
	sort.Strings(result)
	return result
}
