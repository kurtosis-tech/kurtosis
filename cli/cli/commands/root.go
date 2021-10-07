/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package commands

import (
	"github.com/kurtosis-tech/kurtosis/cli/commands/enclave"
	"github.com/kurtosis-tech/kurtosis/cli/commands/lambda"
	"github.com/kurtosis-tech/kurtosis/cli/commands/sandbox"
	"github.com/kurtosis-tech/kurtosis/cli/commands/service"
	"github.com/kurtosis-tech/kurtosis/cli/commands/test"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	// Leaving out the "use" will auto-use os.Args[0]
	Use:                        "",
	Short: "A CLI for interacting with the Kurtosis engine",

	// Cobra will print usage whenever _any_ error occurs, including ones we throw in Kurtosis
	// This doesn't make sense in 99% of the cases, so just turn them off entirely
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(sandbox.SandboxCmd)
	RootCmd.AddCommand(test.TestCmd)
	RootCmd.AddCommand(enclave.EnclaveCmd)
	RootCmd.AddCommand(service.ServiceCmd)
	RootCmd.AddCommand(lambda.LambdaCmd)
}
