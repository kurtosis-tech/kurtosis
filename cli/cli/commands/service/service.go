/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package service

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/service/add"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/service/logs"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/service/pause"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/service/rm"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/service/shell"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/service/unpause"
	"github.com/spf13/cobra"
)

var ServiceCmd = &cobra.Command{
	Use:   command_str_consts.ServiceCmdStr,
	Short: "Manage services",
	RunE:  nil,
}

func init() {
	ServiceCmd.AddCommand(add.ServiceAddCmd.MustGetCobraCommand())
	ServiceCmd.AddCommand(logs.LogsCmd)
	ServiceCmd.AddCommand(rm.ServiceRmCmd.MustGetCobraCommand())
	ServiceCmd.AddCommand(shell.ShellCmd)
	ServiceCmd.AddCommand(pause.PauseCmd)
	ServiceCmd.AddCommand(unpause.UnpauseCmd)
}
