/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/enclave/add"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/enclave/dump"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/enclave/inspect"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/enclave/ls"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/enclave/rm"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/enclave/stop"
	"github.com/spf13/cobra"
)

var EnclaveCmd = &cobra.Command{
	Use:                    command_str_consts.EnclaveCmdStr,
	Aliases:                nil,
	SuggestFor:             nil,
	Short:                  "Manage enclaves",
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
	EnclaveCmd.AddCommand(ls.EnclaveLsCmd.MustGetCobraCommand())
	EnclaveCmd.AddCommand(inspect.EnclaveInspectCmd.MustGetCobraCommand())
	EnclaveCmd.AddCommand(add.EnclaveAddCmd)
	EnclaveCmd.AddCommand(stop.EnclaveStopCmd.MustGetCobraCommand())
	EnclaveCmd.AddCommand(rm.EnclaveRmCmd.MustGetCobraCommand())
	EnclaveCmd.AddCommand(dump.EnclaveDumpCmd.MustGetCobraCommand())
}
