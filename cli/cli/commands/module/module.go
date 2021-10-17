/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package module

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/module/exec"
	"github.com/spf13/cobra"
)

var ModuleCmd = &cobra.Command{
	Use:   "module",
	Short: "Manage Kurtosis modules",
	RunE:  nil,
}

func init() {
	ModuleCmd.AddCommand(exec.ExecCmd)
}
