package lowlevel

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

const (
	shouldLogCompletionDebugMessagesToStderr = true
)

// LowlevelKurtosisCommand is the most configurable, lowest-level implementation of the KurtosisCommand interface
// This is a struct intended to abstract away much of the details of creating a Cobra command that does what we want,
//  so that Kurtosis devs can talk in higher-level notions
// E.g. simply by providing the flags and args, the usage string will be automatically generated for the Kurtosis dev
type LowlevelKurtosisCommand struct {
	// The string for the command (e.g. "inspect" or "ls")
	CommandStr string

	// Will be used when displaying the command for tab completion
	ShortDescription string

	LongDescription string

	// Order isn't important here
	Flags []*flags.FlagConfig

	// Order IS important here
	Args []*args.ArgConfig

	// Oftentimes, the validation logic and the run logic will require the same resources (e.g. a dockerClient, or
	//  dockerManager, or engineClient, etc.) This function will run before both validation & run, so that you can
	//  create resources and add them into the context that gets passed to the validation & run funcs
	PreValidationAndRunFunc func(ctx context.Context) (context.Context, error)

	// The actual logic that the command will run
	RunFunc func(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error

	// Function used to close resources opened in PreValidationAndRunFunc, which is guaranteed to run no matter the outcome
	//  of validation or run
	// This function should only be used for closing resources, and cannot change the return value of the command
	PostValidationAndRunFunc func(ctx context.Context)
}

// Gets a Cobra command represnting the LowlevelKurtosisCommand
// This function is intended to be run in an init() (i.e. before the program runs any logic), so it will panic if
//  any errors occur
func (kurtosisCmd *LowlevelKurtosisCommand) MustGetCobraCommand() *cobra.Command {
	// Verify basic things (e.g. command string & run function) are provided
	if strings.TrimSpace(kurtosisCmd.CommandStr) == "" {
		panic(stacktrace.NewError(
			"A Kurtosis command must have a command string",
		))
	}
	if strings.TrimSpace(kurtosisCmd.ShortDescription) == "" {
		panic(stacktrace.NewError(
			"A short description must be defined for command '%v'",
			kurtosisCmd.CommandStr,
		))
	}
	if strings.TrimSpace(kurtosisCmd.LongDescription) == "" {
		panic(stacktrace.NewError(
			"A long description must be defined for command '%v'",
			kurtosisCmd.CommandStr,
		))
	}
	if kurtosisCmd.RunFunc == nil {
		panic(stacktrace.NewError(
			"A run function must be defined for command '%v'",
			kurtosisCmd.CommandStr,
		))
	}


	// Verify no duplicate flag keys
	usedFlagKeys := map[string]bool{}
	for _, flagConfig := range kurtosisCmd.Flags {
		key := flagConfig.Key
		if len(strings.TrimSpace(key)) == 0 {
			panic(stacktrace.NewError(
				"Empty flag key defined for command '%v'",
				kurtosisCmd.CommandStr,
			))
		}
		if _, found := usedFlagKeys[key]; found {
			panic(stacktrace.NewError(
				"Found duplicate flags with key '%v' for command '%v'",
				key,
				kurtosisCmd.CommandStr,
			))
		}
		usedFlagKeys[key] = true
	}

	// Verify shorthands are unique and, if they exist, only one letter
	flagsByUsedShorthand := map[string]string{}
	for _, flagConfig := range kurtosisCmd.Flags {
		key := flagConfig.Key
		shorthand := flagConfig.Shorthand
		if len(shorthand) == 0 {
			continue
		}

		if len(shorthand) != 1 {
			panic(stacktrace.NewError(
				"Arg '%v' for command '%v' declares shorthand '%v' that isn't exactly 1 letter",
				key,
				kurtosisCmd.CommandStr,
				shorthand,
			))
		}

		if preexistingFlagKey, found := flagsByUsedShorthand[flagConfig.Shorthand]; found {
			panic(stacktrace.NewError(
				"Arg '%v' for command '%v' declares shorthand '%v', but this shorthand is already used by flag '%v'",
				key,
				kurtosisCmd.CommandStr,
				shorthand,
				preexistingFlagKey,
			))
		}
		flagsByUsedShorthand[shorthand] = key
	}

	// Verify no duplicate arg keys
	usedArgKeys := map[string]bool{}
	for _, argConfig := range kurtosisCmd.Args {
		key := argConfig.Key
		if len(strings.TrimSpace(key)) == 0 {
			panic(stacktrace.NewError(
				"Empty arg key defined for command '%v'",
				kurtosisCmd.CommandStr,
			))
		}
		if _, found := usedArgKeys[key]; found {
			panic(stacktrace.NewError(
				"Found duplicate args with key '%v' for command '%v'",
				key,
				kurtosisCmd.CommandStr,
			))
		}
		usedArgKeys[key] = true
	}

	// Verify that we don't have any invalid positional arg combinations, e.g.:
	//  - Any arg after an optional arg (the parser wouldn't know whether you want the optional arg or the one after it)
	//  - Any arg after an arg that consumes N args (since the CLI couldn't know where the greedy arg stops and the required arg begins)
	terminalArgKey := ""
	for _, argConfig := range kurtosisCmd.Args {
		key := argConfig.Key
		if terminalArgKey != "" {
			panic(stacktrace.NewError(
				"Arg '%v' for command '%v' must be the last argument because it's either optional or greedy, but arg '%v' was declared after it",
				terminalArgKey,
				kurtosisCmd.CommandStr,
				key,
			))
		}
		if argConfig.IsOptional || argConfig.IsGreedy {
			terminalArgKey = key
		}
	}

	// Verify all optional args have default values that match their type
	for _, argConfig := range kurtosisCmd.Args {
		if !argConfig.IsOptional {
			continue
		}

		key := argConfig.Key
		if argConfig.DefaultValue == nil {
			panic(stacktrace.NewError(
				"Arg '%v for command '%v' is optional, but doesn't have a default value",
				key,
				kurtosisCmd.CommandStr,
			))
		}
		if argConfig.IsGreedy {
			_, ok := argConfig.DefaultValue.([]string)
			if !ok {
				panic(stacktrace.NewError(
					"Greedy arg '%v for command '%v' is optional, but the default value isn't a string array",
					key,
					kurtosisCmd.CommandStr,
				))
			}
		} else {
			_, ok := argConfig.DefaultValue.(string)
			if !ok {
				panic(stacktrace.NewError(
					"Non-greedy arg '%v for command '%v' is optional, but the default value isn't a string",
					key,
					kurtosisCmd.CommandStr,
				))
			}
		}
	}

	// Based on digging through the Cobra source code, the toComplete string is theoretically the string that the user
	//  is in the process of typing when they press TAB. However, in my tests on Bash, the shell will automatically
	//  filter the results based off the partialStr without us needing to filter them ~ ktoday, 2022-02-02
	getCompletionsFunc := func(cmd *cobra.Command, previousArgStrs []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		ctx := context.Background()

		parsedFlags := flags.NewParsedFlags(cmd.Flags())

		parsedArgs, argToComplete := args.ParseArgsForCompletion(kurtosisCmd.Args, previousArgStrs)
		if argToComplete == nil {
			// NOTE: We can't just use logrus because anything printed to STDOUT will be interpreted as a completion
			// See:
			//  https://github.com/spf13/cobra/blob/master/shell_completions.md#:~:text=the%20RunE%20function.-,Debugging,ShellCompDirectiveNoFileComp%20%23%20This%20is%20on%20stderr
			cobra.CompDebugln("Not completing because no argument needs completion", shouldLogCompletionDebugMessagesToStderr)
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		completionFunc := argToComplete.CompletionsFunc
		if completionFunc == nil {
			// NOTE: We can't just use logrus because anything printed to STDOUT will be interpreted as a completion
			// See:
			//  https://github.com/spf13/cobra/blob/master/shell_completions.md#:~:text=the%20RunE%20function.-,Debugging,ShellCompDirectiveNoFileComp%20%23%20This%20is%20on%20stderr
			cobra.CompDebugln(
				fmt.Sprintf(
					"Not completing because arg needing completion '%v' doesn't have a custom completion function",
					argToComplete.Key,
				),
				shouldLogCompletionDebugMessagesToStderr,
			)
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		completions, err := argToComplete.CompletionsFunc(ctx, parsedFlags, parsedArgs)
		if err != nil {
			// NOTE: We can't just use logrus because anything printed to STDOUT will be interpreted as a completion
			// See:
			//  https://github.com/spf13/cobra/blob/master/shell_completions.md#:~:text=the%20RunE%20function.-,Debugging,ShellCompDirectiveNoFileComp%20%23%20This%20is%20on%20stderr
			cobra.CompDebugln(
				fmt.Sprintf(
					"An error occurred running the completions function with previous arg strs '%+v' and toComplete string '%v':\n%v",
					previousArgStrs,
					toComplete,
					err,
				),
				shouldLogCompletionDebugMessagesToStderr,
			)
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}

	// Prepare the run function to be slotted into the Cobra command, which will do both arg validation & logic execution
	cobraRunFunc := func(cmd *cobra.Command, allArgs []string) error {
		parsedFlags := flags.NewParsedFlags(cmd.Flags())

		parsedArgs, err := args.ParseArgsForValidation(kurtosisCmd.Args, allArgs)
		if err != nil {
			logrus.Debugf("An error occurred while parsing args '%+v':\n%v", allArgs, err)

			// NOTE: This is a VERY special instance where we don't wrap the error with stacktrace.Propagate, because
			//  the errors returned by this function will *only* be arg-parsing errors and the stacktrace just adds
			//  clutter & confusion to what the user sees without providing any useful information
			return err
		}

		ctx := context.Background()
		if kurtosisCmd.PreValidationAndRunFunc != nil {
			newCtx, err := kurtosisCmd.PreValidationAndRunFunc(ctx)
			if err != nil {
				return stacktrace.Propagate(err, "An error occurred running the pre-validation-and-run function")
			}
			ctx = newCtx
		}
		if kurtosisCmd.PostValidationAndRunFunc != nil {
			defer kurtosisCmd.PostValidationAndRunFunc(ctx)
		}

		// Validate all the args
		for _, config := range kurtosisCmd.Args {
			validationFunc := config.ValidationFunc
			if validationFunc == nil {
				continue
			}
			if err := validationFunc(ctx, parsedFlags, parsedArgs); err != nil {
				return stacktrace.Propagate(err, "An error occurred validating arg '%v'", config.Key)
			}
		}

		if err := kurtosisCmd.RunFunc(ctx, parsedFlags, parsedArgs); err != nil {
			return stacktrace.Propagate(err, "An error occurred running command '%v'", kurtosisCmd.CommandStr)
		}

		return nil
	}

	// Build usage string
	allArgUsageStrs := []string{}
	for _, argConfig := range kurtosisCmd.Args {
		argUsageStr := renderArgUsageStr(argConfig)
		allArgUsageStrs = append(allArgUsageStrs, argUsageStr)
	}
	usageStr := fmt.Sprintf(
		"%v [flags] %v",
		kurtosisCmd.CommandStr,
		strings.Join(allArgUsageStrs, " "),
	)

	result := &cobra.Command{
		Use:                   usageStr,
		DisableFlagsInUseLine: true, // Not needed since we manually add the string in the usage string
		Short:                 kurtosisCmd.ShortDescription,
		Long:                  kurtosisCmd.LongDescription,
		ValidArgsFunction:     getCompletionsFunc,
		RunE: cobraRunFunc,
	}

	// Validates that the default values for the declared flags match the declard types, and add them to the Cobra command
	// Verify all flag default values match their declared types
	resultFlags := result.Flags()
	for _, flagConfig := range kurtosisCmd.Flags {
		key := flagConfig.Key
		shorthand := flagConfig.Shorthand
		usage := flagConfig.Usage
		defaultValStr := flagConfig.Default
		flagType := flagConfig.Type

		processor, found := flags.AllFlagTypeProcessors[flagType]
		if !found {
			// Should never happen because we enforce completeness via unit test
			panic(stacktrace.NewError(
				"Flag '%v' on command '%v' has type '%v' which doesn't have a flag type processor defined; this means " +
					"that the flag type is invalid or a processor needs to be defined",
				key,
				kurtosisCmd.CommandStr,
				flagType.String(),
			))
		}

		if err := processor(key, shorthand, defaultValStr, usage, resultFlags); err != nil {
			panic(stacktrace.NewError(
				"An error occurred processing flag '%v' on command '%v' of type '%v'",
				key,
				kurtosisCmd.CommandStr,
				flagType.String(),
			))
		}
	}

	return result
}


// ====================================================================================================
//                                   Private Helper Functions
// ====================================================================================================

func renderArgUsageStr(arg *args.ArgConfig) string {
	result := arg.Key
	if arg.IsGreedy {
		result = result + "..."
	}
	if arg.IsOptional {
		result = "[" + result + "]"
	}
	return result
}