package command_wrappers

import "github.com/spf13/cobra"

// Phases:
// Parse: the user presses tab-complete (so no error-reporting)
// Validation: the user has pressed ENTER


// I want the KUrtosis dev to be able to:
// - Define autocomplete and validation in the same spot
// - Be able to reuse autocomplete/validation components (e.g. enclave ID is supercommon)
// - Should be free to define
// - Their run function should be: func(flags ParsedFlags, positionalArgs ParsedPositionalArgs) error
	// - ParsedFlags will have GetBoolFlag, GetStringFlag, GetIntFlag, etc.
	// - ParsedPositionalArgs will have GetSingleElemArg() string and GetNElemsArgs() []string
// - Before their run function runs, all the validations will be applied

// I don't care right now about:
// - Flag 
// - Flag value autocompletion

// Options:
// 1) constantly pass off args

// Validations that should be genericized:
// - enclave ID
// - service ID (requires enclave ID)
// - REPL ID (requires enclave ID)
// - module ID

// A command that operates on a Kurtosis enclave
type KurtosisEnclaveCommand struct {
	// E.g. "enclave-id"
	EnclaveIDArgKey string



	RunFunc func(enclaveId string, args []string) error
}

func GetCobraCommand() *cobra.Command {

}
