package command_wrappers

import "github.com/spf13/cobra"


// A command that operates on a Kurtosis enclave
type KurtosisEnclaveCommand struct {
	// E.g. "enclave-id"
	EnclaveIDArgKey string

	RunFunc func(enclaveId string, args []string) error
}

func GetCobraCommand() *cobra.Command {

}
