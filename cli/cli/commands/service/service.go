/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package service

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/service/add"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/service/logs"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/service/pause"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/service/rm"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/service/shell"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/service/unpause"
	"github.com/spf13/cobra"
)

var ServiceCmd = &cobra.Command{
	Use:                    command_str_consts.ServiceCmdStr,
	Aliases:                nil,
	SuggestFor:             nil,
	Short:                  "Manage services",
	Long:                   "",
	Example:                "",
	ValidArgs:              nil,
	ValidArgsFunction:      nil,
	Args:                   nil,
	ArgAliases:             nil,
	BashCompletionFunction: "",
	Deprecated:             "",
	Annotations:            nil,
	Version:                "",
	PersistentPreRun:       nil,
	PersistentPreRunE:      nil,
	PreRun:                 nil,
	PreRunE:                nil,
	Run:                    nil,
	RunE:                   nil,
	PostRun:                nil,
	PostRunE:               nil,
	PersistentPostRun:      nil,
	PersistentPostRunE:     nil,
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: false,
	},
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd:   false,
		DisableNoDescFlag:   false,
		DisableDescriptions: false,
	},
	TraverseChildren:           false,
	Hidden:                     false,
	SilenceErrors:              false,
	SilenceUsage:               false,
	DisableFlagParsing:         false,
	DisableAutoGenTag:          false,
	DisableFlagsInUseLine:      false,
	DisableSuggestions:         false,
	SuggestionsMinimumDistance: 0,
}

func init() {
	ServiceCmd.AddCommand(add.ServiceAddCmd.MustGetCobraCommand())
	ServiceCmd.AddCommand(logs.ServiceLogsCmd.MustGetCobraCommand())
	ServiceCmd.AddCommand(rm.ServiceRmCmd.MustGetCobraCommand())
	ServiceCmd.AddCommand(shell.ServiceShellCmd.MustGetCobraCommand())
	ServiceCmd.AddCommand(pause.PauseCmd.MustGetCobraCommand())
	ServiceCmd.AddCommand(unpause.UnpauseCmd.MustGetCobraCommand())
}
