/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package commands

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/enclave"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/engine"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/module"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/repl"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/sandbox"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/service"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/test"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/version"
	logrus_log_levels2 "github.com/kurtosis-tech/kurtosis-cli/cli/helpers/logrus_log_levels"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

const (
	logLevelStrArg = "cli-log-level"
)

var logLevelStr string
var defaultLogLevelStr = logrus.InfoLevel.String()

var RootCmd = &cobra.Command{
	// Leaving out the "use" will auto-use os.Args[0]
	Use:                        "",
	Short: "A CLI for interacting with the Kurtosis engine",

	// Cobra will print usage whenever _any_ error occurs, including ones we throw in Kurtosis
	// This doesn't make sense in 99% of the cases, so just turn them off entirely
	SilenceUsage: true,
	PersistentPreRunE: globalSetup,
}

func init() {
	RootCmd.PersistentFlags().StringVar(
		&logLevelStr,
		logLevelStrArg,
		defaultLogLevelStr,
		"Sets the level that the CLI will log at (" + strings.Join(logrus_log_levels2.GetAcceptableLogLevelStrs(), "|") + ")",
	)

	RootCmd.AddCommand(sandbox.SandboxCmd)
	RootCmd.AddCommand(test.TestCmd)
	RootCmd.AddCommand(enclave.EnclaveCmd)
	RootCmd.AddCommand(service.ServiceCmd)
	RootCmd.AddCommand(module.ModuleCmd)
	RootCmd.AddCommand(repl.REPLCmd)
	RootCmd.AddCommand(engine.EngineCmd)
	RootCmd.AddCommand(version.VersionCmd)

	// TODO Add global flag to set the CLI's log level
}

func globalSetup(cmd *cobra.Command, args []string) error {
	logLevel, err := logrus.ParseLevel(logLevelStr)
	if err != nil {
		return stacktrace.Propagate(err, "Could not parse log level string '%v'", logLevelStr)
	}
	logrus.SetOutput(cmd.OutOrStdout())
	logrus.SetLevel(logLevel)
	return nil
}