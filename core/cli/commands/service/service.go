/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package service

import (
	"github.com/kurtosis-tech/kurtosis/cli/commands/service/logs"
	"github.com/spf13/cobra"
)

var ServiceCmd = &cobra.Command{
	Use:   "service",
	Short: "Manage services",
	RunE:  nil,
}

func init() {
	ServiceCmd.AddCommand(logs.LogsCmd)
}
