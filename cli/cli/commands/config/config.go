/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package config

import (
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_str_consts"
	config_path "github.com/dzobbe/PoTE-kurtosis/cli/cli/commands/config/path"
	"github.com/spf13/cobra"
)

// ConfigCmd Suppressing exhaustruct requirement because this struct has ~40 properties
// nolint: exhaustruct
var ConfigCmd = &cobra.Command{
	Use:   command_str_consts.ConfigCmdStr,
	Short: "Manage configurations",
	RunE:  nil,
}

func init() {
	ConfigCmd.AddCommand(config_path.PathCmd.MustGetCobraCommand())
}
