/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave

import (
	"github.com/kurtosis-tech/kurtosis-core/cli/commands/enclave/inspect"
	"github.com/kurtosis-tech/kurtosis-core/cli/commands/enclave/ls"
	"github.com/spf13/cobra"
)

var EnclaveCmd = &cobra.Command{
	Use:   "enclave",
	Short: "Manage enclaves",
	RunE:  nil,
}

func init() {
	EnclaveCmd.AddCommand(ls.LsCmd)
	EnclaveCmd.AddCommand(inspect.InspectCmd)
}
