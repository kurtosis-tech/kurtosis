/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave

import (
	"github.com/kurtosis-tech/kurtosis/cli/commands/enclave/ls"
	"github.com/spf13/cobra"
)

var EnclaveCmd = &cobra.Command{
	Use:   "enclave",
	Short: "Group all enclave commands",
	RunE:  run,
}

func init() {
	EnclaveCmd.AddCommand(ls.LsCmd)
}

func run(cmd *cobra.Command, args []string) error {
	return nil
}
