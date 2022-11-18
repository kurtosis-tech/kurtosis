package args

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/spf13/cobra"
)

const (
	shellDirectiveForManualCompletionProvider = cobra.ShellCompDirectiveNoFileComp
)

type ArgCompletionProvider interface {
	// Returns an argument completion func
	GetCompletionFunction() func(ctx context.Context, flags *flags.ParsedFlags, previousArgs *ParsedArgs) ([]string, cobra.ShellCompDirective, error)
}

// Only ONE of these fields will be set at a time!
type argCompletionProviderImpl struct {
	completionOptions []string

	shellCompletionDirective cobra.ShellCompDirective
}

func (impl *argCompletionProviderImpl) GetCompletionFunction() {

	//There is an error on this case, error loudly
	if both are nil

	if impl.completionOptions != nil {
		return impl.manualCompletions, shellDirectiveForManualCompletionProvider
	}

	if impl.shellCompletionDirective != nil {
		return
	}


}

func NewManualCompletionsProvider([]func) {

}

func NewShellCompletionsProvider(cobra.ShellCompletion) {
	emptyCompletionOptions := []string{}
}

