package args

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/spf13/cobra"
)

const (
	shellDirectiveForManualCompletionProvider          = cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
	shellDirectiveForShellProvideDefaultFileCompletion = cobra.ShellCompDirectiveDefault
	defaultShellDirective                              = cobra.ShellCompDirectiveDefault
	noShellDirectiveDefined                            = 999999
)

var (
	noCustomCompletionFuncDefined *func(ctx context.Context, flags *flags.ParsedFlags, previousArgs *ParsedArgs) ([]string, error) = nil
)

type argCompletionProvider interface {
	// Runs the argument completion func
	RunCompletionFunction(ctx context.Context, flags *flags.ParsedFlags, previousArgs *ParsedArgs) ([]string, cobra.ShellCompDirective, error)
}

// Only ONE of these fields will be set at a time!
type argCompletionProviderImpl struct {
	customCompletionFunc *func(ctx context.Context, flags *flags.ParsedFlags, previousArgs *ParsedArgs) ([]string, error)

	shellCompletionDirective cobra.ShellCompDirective
}

func (impl *argCompletionProviderImpl) RunCompletionFunction(
	ctx context.Context,
	flags *flags.ParsedFlags,
	previousArgs *ParsedArgs,
) ([]string, cobra.ShellCompDirective, error) {

	if impl.customCompletionFunc != noCustomCompletionFuncDefined {
		completionFunc := *impl.customCompletionFunc
		completions, err := completionFunc(ctx, flags, previousArgs)
		return completions, shellDirectiveForManualCompletionProvider, err
	}

	if impl.shellCompletionDirective != noShellDirectiveDefined {
		return nil, shellDirectiveForShellProvideDefaultFileCompletion, nil
	}

	return nil, defaultShellDirective, stacktrace.NewError("The custom completion func and the shell completion directive are not defined, this should never happens; this is a bug in Kurtosis")
}

// Receive a custom completion function which should generate the completions list for the argument and return it
func NewManualCompletionsProvider(
	customCompletionFunc func(ctx context.Context, flags *flags.ParsedFlags, previousArgs *ParsedArgs) ([]string, error),
) argCompletionProvider {
	newManualCompletionProvider := &argCompletionProviderImpl{
		customCompletionFunc:     &customCompletionFunc,
		shellCompletionDirective: noShellDirectiveDefined,
	}
	return newManualCompletionProvider
}

// This argument completion provider enables the default shell file completion functionality for the argument
func NewDefaultShellFileCompletionProvider() argCompletionProvider {
	newDefaultShellFileCompletionProvider := &argCompletionProviderImpl{
		customCompletionFunc:     noCustomCompletionFuncDefined,
		shellCompletionDirective: shellDirectiveForShellProvideDefaultFileCompletion,
	}
	return newDefaultShellFileCompletionProvider
}
