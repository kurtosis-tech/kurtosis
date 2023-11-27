/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/enclave/add"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/enclave/connect"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/enclave/dump"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/enclave/inspect"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/enclave/ls"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/enclave/rm"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/enclave/stop"
	"github.com/spf13/cobra"
)

// EnclaveCmd Suppressing exhaustruct requirement because this struct has ~40 properties
// nolint: exhaustruct
var EnclaveCmd = &cobra.Command{
	Use:   command_str_consts.EnclaveCmdStr,
	Short: "Manage enclaves",
	RunE:  nil,
}

func init() {
	EnclaveCmd.AddCommand(ls.EnclaveLsCmd.MustGetCobraCommand())
	EnclaveCmd.AddCommand(inspect.EnclaveInspectCmd.MustGetCobraCommand())
	EnclaveCmd.AddCommand(add.EnclaveAddCmd.MustGetCobraCommand())
	EnclaveCmd.AddCommand(stop.EnclaveStopCmd.MustGetCobraCommand())
	EnclaveCmd.AddCommand(rm.EnclaveRmCmd.MustGetCobraCommand())
	EnclaveCmd.AddCommand(dump.EnclaveDumpCmd.MustGetCobraCommand())
	EnclaveCmd.AddCommand(connect.EnclaveConnectCmd.MustGetCobraCommand())
}
