/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package config

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	init_config "github.com/kurtosis-tech/kurtosis/cli/cli/commands/config/init"
	config_path "github.com/kurtosis-tech/kurtosis/cli/cli/commands/config/path"
	"github.com/spf13/cobra"
)

var ConfigCmd = &cobra.Command{
	Use:                    command_str_consts.ConfigCmdStr,
	Aliases:                nil,
	SuggestFor:             nil,
	Short:                  "Manage configurations",
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
	ConfigCmd.AddCommand(init_config.InitCmd.MustGetCobraCommand())
	ConfigCmd.AddCommand(config_path.PathCmd.MustGetCobraCommand())
}
