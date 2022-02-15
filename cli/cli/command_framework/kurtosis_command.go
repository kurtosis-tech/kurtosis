package command_framework

import "github.com/spf13/cobra"

// A Kurtosis command is a command that wraps a Cobra command to make it easier to work with
// There are many implementations, and some wrap the logic of others
type KurtosisCommand interface {
	MustGetCobraCommand() *cobra.Command
}