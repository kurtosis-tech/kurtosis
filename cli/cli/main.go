/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package main

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands"
	"github.com/sirupsen/logrus"
	"os"
)

const (
	successExitCode = 0
	errorExitCode   = 1

	forceColors   = true
	fullTimestamp = true
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
		// We don't actually need to print the error because Cobra will do it for us
		os.Exit(errorExitCode)
	}
	os.Exit(successExitCode)
}
