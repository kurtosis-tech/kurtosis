/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package main

import (
	"errors"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands"
	"github.com/kurtosis-tech/kurtosis/cli/cli/out"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

const (
	successExitCode = 0
	errorExitCode   = 1

	forceColors   = true
	fullTimestamp = true

	separator                      = "---"
	removeTrailingSpecialCharacter = "\n"
)

func main() {
	// NOTE: we'll want to change the ForceColors to false if we ever want structured logging
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:               forceColors,
		DisableColors:             false,
		ForceQuote:                false,
		DisableQuote:              false,
		EnvironmentOverrideColors: false,
		DisableTimestamp:          false,
		FullTimestamp:             fullTimestamp,
		TimestampFormat:           "",
		DisableSorting:            false,
		SortingFunc:               nil,
		DisableLevelTruncation:    false,
		PadLevelText:              false,
		QuoteEmptyFields:          false,
		FieldMap:                  nil,
		CallerPrettyfier:          nil,
	})

	if err := commands.RootCmd.Execute(); err != nil {
		maybeCleanedError := getErrorMessageToBeDisplayedOnCli(err)
		// cobra uses this method underneath as well to print errors
		// so just used it directly.
		commands.RootCmd.PrintErrln("Error: ", maybeCleanedError.Error())
		os.Exit(errorExitCode)
	}
	os.Exit(successExitCode)
}

func getErrorMessageToBeDisplayedOnCli(errorWithStacktrace error) error {
	// if we are running in the debug mode, just return the error with stack-traces back to the client
	if logrus.GetLevel() == logrus.DebugLevel {
		return errorWithStacktrace
	}

	// silently catch the file logger error and print it in the debug mode
	// users should not worry about this error
	// downside is that we may lose stack-traces during file logger failures
	fileLogger, err := out.GetFileLogger()
	if err != nil {
		logrus.Debugf("Error occurred while getting the file logger %+v", err)
	} else {
		fileLogger.Errorln(errorWithStacktrace.Error())
	}

	errorMessage := errorWithStacktrace.Error()
	cleanError := removeFilePathFromErrorMessage(errorMessage)
	return cleanError
}

// this method removes the file path from the error
func removeFilePathFromErrorMessage(errorMessage string) error {
	errorMessageConvertedInList := strings.Split(errorMessage, separator)
	// safe to assume that the error was not generated using stacktrace package
	if len(errorMessageConvertedInList) == 1 {
		return errors.New(errorMessage)
	}

	// only the even numbered elements needs to be picked.
	var cleanErrorList []string
	for index, line := range errorMessageConvertedInList {
		if index%2 == 0 {
			cleanErrorList = append(cleanErrorList, strings.Trim(line, removeTrailingSpecialCharacter))
		}
	}

	return errors.New(strings.Join(cleanErrorList, ""))
}
