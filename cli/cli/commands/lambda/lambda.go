/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package lambda

import (
	"github.com/kurtosis-tech/kurtosis-core/cli/commands/lambda/exec"
	"github.com/spf13/cobra"
)

var LambdaCmd = &cobra.Command{
	Use:   "lambda",
	Short: "Manage Kurtosis Lambda",
	RunE:  nil,
}

func init() {
	LambdaCmd.AddCommand(exec.ExecCmd)
}
