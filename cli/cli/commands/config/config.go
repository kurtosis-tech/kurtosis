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
	Use:   command_str_consts.ConfigCmdStr,
	Short: "Manage configurations",
	RunE:  nil,
}

func init() {
	ConfigCmd.AddCommand(init_config.InitCmd.MustGetCobraCommand())
	ConfigCmd.AddCommand(config_path.PathCmd.MustGetCobraCommand())
}

