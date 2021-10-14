/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package repl

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/repl/install"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/repl/new"
	"github.com/spf13/cobra"
)

var REPLCmd = &cobra.Command{
	Use:   "repl",
	Short: "Manage REPL",
	RunE:  nil,
}

func init() {
	REPLCmd.AddCommand(new.NewCmd)
	REPLCmd.AddCommand(install.InstallCmd)
}
