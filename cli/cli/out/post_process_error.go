package out

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"regexp"
	"strings"
)

const (
	lineWithStacktraceRegex       = "^--- at\\s*(.*?):([\\d]*)\\s*\\((.*?)\\)\\s*---$"
	separator                     = "\n"
	indentation                   = "  "
	errorNotCreatedFromStacktrace = 1
)

var lineWithStacktrace = regexp.MustCompile(lineWithStacktraceRegex)

func GetErrorMessageToBeDisplayedOnCli(errorWithStacktrace error) error {
	// silently catch the file logger error and print it in the debug mode
	// users should not worry about this error
	// downside is that we may lose stack-traces during file logger failures
	loggerToFile, err := GetFileLogger()
	if err != nil {
		logrus.Warnf("Error occurred while getting the file logger %+v", err)
	} else {
		loggerToFile.Errorln(errorWithStacktrace.Error())
	}

	// if we are running in the debug mode, just return the error with stack-traces back to the client
	if logrus.GetLevel() == logrus.DebugLevel {
		return errorWithStacktrace
	}

	errorMessage := errorWithStacktrace.Error()
	cleanError := removeFilePathFromErrorMessage(errorMessage)
	return cleanError
}

// this method removes the file path from the error
func removeFilePathFromErrorMessage(errorMessage string) error {
	errorMessageConvertedInList := strings.Split(errorMessage, separator)
	// safe to assume that the error was not generated using stacktrace package
	if len(errorMessageConvertedInList) == errorNotCreatedFromStacktrace {
		return errors.New(errorMessage)
	}

	// only the even numbered elements needs to be picked.
	var cleanErrorList []string
	for idx, line := range errorMessageConvertedInList {
		// this only cleans spaces for the lines that contains the stack-trace information
		trimmedLine := strings.TrimSpace(line)
		if !lineWithStacktrace.MatchString(trimmedLine) {
			var maybeIndentedLine string
			if idx > 0 {
				maybeIndentedLine = fmt.Sprintf("%s%s", indentation, line)
			} else {
				maybeIndentedLine = line
			}
			cleanErrorList = append(cleanErrorList, maybeIndentedLine)
		}
	}

	cleanErrorMessage := strings.Join(cleanErrorList, "\n")
	return errors.New(cleanErrorMessage)
}
