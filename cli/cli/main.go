/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package main

import (
	"errors"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands"
	"github.com/kurtosis-tech/kurtosis/cli/cli/out"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

const (
	successExitCode = 0
	errorExitCode   = 1

	forceColors   = true
	fullTimestamp = true

	errorPrefix     = "Error: "
	commandNotFound = "unknown command"
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
		if !displayErrorMessageToCli(err) {
			os.Exit(errorExitCode)
		}

		maybeCleanedError := out.GetErrorMessageToBeDisplayedOnCli(err)
		errorMessageFromCli := maybeCleanedError.Error()

		commands.RootCmd.PrintErrln(errorPrefix, errorMessageFromCli)
		// if unknown command is entered - display help command
		if strings.Contains(errorMessageFromCli, commandNotFound) {
			commands.RootCmd.PrintErrf("Run '%v --help' for usage.\n", commands.RootCmd.CommandPath())
		}
		os.Exit(errorExitCode)
	}
	os.Exit(successExitCode)
}

func displayErrorMessageToCli(err error) bool {
	rootCause := stacktrace.RootCause(err)
	return !errors.Is(rootCause, command_str_consts.ErrorMessageDueToStarlarkFailure)
}
