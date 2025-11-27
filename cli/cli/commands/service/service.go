/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package service

import (
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_str_consts"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/commands/service/add"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/commands/service/exec"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/commands/service/inspect"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/commands/service/logs"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/commands/service/rm"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/commands/service/shell"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/commands/service/start"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/commands/service/stop"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/commands/service/update"
	"github.com/spf13/cobra"
)

// ServiceCmd Suppressing exhaustruct requirement because this struct has ~40 properties
// nolint: exhaustruct
var ServiceCmd = &cobra.Command{
	Use:   command_str_consts.ServiceCmdStr,
	Short: "Manage services",
	RunE:  nil,
}

func init() {
	ServiceCmd.AddCommand(add.ServiceAddCmd.MustGetCobraCommand())
	ServiceCmd.AddCommand(exec.ServiceShellCmd.MustGetCobraCommand())
	ServiceCmd.AddCommand(logs.ServiceLogsCmd.MustGetCobraCommand())
	ServiceCmd.AddCommand(rm.ServiceRmCmd.MustGetCobraCommand())
	ServiceCmd.AddCommand(shell.ServiceShellCmd.MustGetCobraCommand())
	ServiceCmd.AddCommand(start.ServiceStartCmd.MustGetCobraCommand())
	ServiceCmd.AddCommand(stop.ServiceStopCmd.MustGetCobraCommand())
	ServiceCmd.AddCommand(inspect.ServiceInspectCmd.MustGetCobraCommand())
	ServiceCmd.AddCommand(update.ServiceUpdateCmd.MustGetCobraCommand())
}
