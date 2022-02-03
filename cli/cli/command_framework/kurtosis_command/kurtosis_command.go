package kurtosis_command

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/kurtosis_command/parsed_args"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/kurtosis_command/parsed_flags"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strconv"
	"strings"
)

const (
	shouldLogCompletionDebugMessagesToStderr = true

	uintBase = 10
	uint32Bits = 32
)

// This is a struct intended to abstract away much of the details of creating a Cobra command that does what we want,
//  so that Kurtosis devs can talk in higher-level notions
// E.g. simply by providing the flags and args, the usage string will be automatically generated for the Kurtosis dev
type KurtosisCommand struct {
	// The string for the command (e.g. "inspect" or "ls")
	CommandStr string

	// Will be used when displaying the command for tab completion
	ShortDescription string

	LongDescription string

	// Order isn't important here
	Flags []*FlagConfig

	// Order IS important here
	Args []*ArgConfig

	// The actual logic that the command will run
	RunFunc func(flags *parsed_flags.ParsedFlags, args *parsed_args.ParsedArgs) error
}

// Gets a Cobra command represnting the KurtosisCommand
// This function is intended to be run in an init() (i.e. before the program runs any logic), so it will panic if
//  any errors occur
func (kurtosisCmd *KurtosisCommand) MustGetCobraCommand() *cobra.Command {
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

	// Verify all flag default values match their declared types
	for _, flagConfig := range kurtosisCmd.Flags {
		key := flagConfig.Key
		typeStr := flagConfig.Type.typeStr
		defaultValStr := flagConfig.Default
		defaultValueDoesntMatchType := false
		switch typeStr {
		case FlagType_String.typeStr:
			// Nothing to do
		case FlagType_Bool.typeStr:
			_, err := strconv.ParseBool(defaultValStr)
			defaultValueDoesntMatchType = err != nil
		case FlagType_Uint32.typeStr:
			_, err := strconv.ParseUint(defaultValStr, uintBase, uint32Bits)
			defaultValueDoesntMatchType = err != nil
		default:
			panic(stacktrace.NewError("Flag '%v' on command '%v' is of unrecognized type '%v'", key, kurtosisCmd.CommandStr, typeStr))
		}
		if defaultValueDoesntMatchType {
			panic(stacktrace.NewError(
				"Default value of flag '%v' on command '%v' is '%v', which doesn't match the flag's declared type of '%v'",
				key,
				kurtosisCmd.CommandStr,
				defaultValStr,
				typeStr,
			))
		}
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

	// Based on digging through the Cobra source code, the toComplete string is theoretically the string that the user
	//  is in the process of typing when they press TAB. However, in my tests on Bash, the shell will automatically
	//  filter the results based off the partialStr without us needing to filter them ~ ktoday, 2022-02-02
	getCompletionsFunc := func(cmd *cobra.Command, previousArgStrs []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		parsedFlags := parsed_flags.NewParsedFlags(cmd.Flags())

		parsedArgs, argToComplete := parsed_args.ParseArgsForCompletion(kurtosisCmd.Args, previousArgStrs)
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

		completions, err := argToComplete.CompletionsFunc(parsedFlags, parsedArgs)
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
	cobraRunFunc := func(cmd *cobra.Command, args []string) error {
		parsedFlags := parsed_flags.NewParsedFlags(cmd.Flags())

		parsedArgs, err := parsed_args.ParseArgsForValidation(kurtosisCmd.Args, args)
		if err != nil {
			logrus.Debugf("An error occurred while parsing args '%+v':\n%v", args, err)

			// NOTE: This is a VERY special instance where we don't wrap the error with stacktrace.Propagate, because
			//  the errors returned by this function will *only* be arg-parsing errors and the stacktrace just adds
			//  clutter & confusion to what the user sees without providing any useful information
			return err
		}

		// Validate all the args
		for _, config := range kurtosisCmd.Args {
			validationFunc := config.ValidationFunc
			if validationFunc == nil {
				continue
			}
			if err := validationFunc(parsedFlags, parsedArgs); err != nil {
				return stacktrace.Propagate(err, "An error occurred validating arg '%v'", config.Key)
			}
		}

		if err := kurtosisCmd.RunFunc(parsedFlags, parsedArgs); err != nil {
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
	
	return &cobra.Command{
		Use:                   usageStr,
		DisableFlagsInUseLine: true, // Not needed since we manually add the string in the usage string
		// TODO FLAGS!!!
		Short:                 kurtosisCmd.ShortDescription,
		Long:                  kurtosisCmd.LongDescription,
		ValidArgsFunction:     getCompletionsFunc,
		RunE: cobraRunFunc,
	}
}


// ====================================================================================================
//                                   Private Helper Functions
// ====================================================================================================

func renderArgUsageStr(arg *ArgConfig) string {
	result := arg.Key
	if arg.IsGreedy {
		result = result + "..."
	}
	if arg.IsOptional {
		result = "[" + result + "]"
	}
	return result
}