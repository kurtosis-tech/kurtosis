package kurtosis_command

import (
	"fmt"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"strings"
)

const (
	shouldLogCompletionDebugMessagesToStderr = true
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
	RunFunc func(flags *ParsedFlags, args *ParsedArgs) error
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
		parsedFlags := &ParsedFlags{
			cmdFlagsSet: cmd.Flags(),
		}

		parsedArgs, argToComplete := parseArgsForCompletion(kurtosisCmd.Args, previousArgStrs)
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
		parsedFlags := &ParsedFlags{
			cmdFlagsSet: cmd.Flags(),
		}

		parsedArgs, err := parseArgsForValidation(kurtosisCmd.Args, args)
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
func parseFlags(flagConfigs []*FlagConfig, cobraFlags *pflag.FlagSet) (*ParsedFlags, error) {

	resultUint32Flags := map[string]uint32{}
	resultStringFlags := map[string]string{}
	resultBoolFlags := map[string]bool{}

	for _, config := range flagConfigs {
		key := config.Key
		typeStr := config.Type.typeStr
		switch typeStr {
		case FlagType_Uint32.typeStr:
			value, err := cobraFlags.GetUint32(key)
			if err != nil {
				return nil, stacktrace.Propagate(err, "An error occurred getting uint32 flag '%v' from the underlying Cobra flag set", key)
			}
			resultUint32Flags[key] = value
		case FlagType_String.typeStr:
			value, err := cobraFlags.GetString(key)
			if err != nil {
				return nil, stacktrace.Propagate(err, "An error occurred getting string flag '%v' from the underlying Cobra flag set", key)
			}
			resultStringFlags[key] = value
		case FlagType_Bool.typeStr:
			value, err := cobraFlags.GetBool(key)
			if err != nil {
				return nil, stacktrace.Propagate(err, "An error occurred getting bool flag '%v' from the underlying Cobra flag set", key)
			}
			resultBoolFlags[key] = value
		default:
			return nil, stacktrace.NewError("Flag '%v' has unrecognized type string '%v'; this is a bug in Kurtosis!", typeStr)
		}
	}
	// TODO Verify that the length of the resulting map == length of flag configs
	numParsedFlagValues := len(resultUint32Flags)
}

// Takes in the currently-entered arg strings, categorizes them according to the arg configs defined, and
//  returns the ArgConfig whose completion function should be used
// NOTES:
//  - If the user presses TAB in the middle of several args (e.g. "arg1 arg2  TAB   arg3"), then `input` will only contain
//     the previous args (which is actually good behaviour)
//  - If the input isn't long enough, the resulting ParsedArgs object won't have arg strings for all the args
//  - A nil value for the returned ArgConfig indicates that no completion should be used
func parseArgsForCompletion(argConfigs []*ArgConfig, input []string) (*ParsedArgs, *ArgConfig) {
	nonGreedyArgValues := map[string]string{}
	greedyArgValues := map[string][]string{}

	var nextArg *ArgConfig = nil
	if len(argConfigs) > 0 {
		nextArg = argConfigs[0]
	}

	configIdx := 0
	inputIdx := 0
	for configIdx < len(argConfigs) && inputIdx < len(input) {
		config := argConfigs[configIdx]
		key := config.Key

		// Greedy case (arg must always be last)
		if config.IsGreedy {
			greedyArgValues[key] = input[inputIdx:]
			inputIdx += len(input) - inputIdx
			nextArg = config  // Greedy args must always be at the end, so they'll be infinitely-completable
			break
		}

		// Non-greedy case
		nonGreedyArgValues[key] = input[inputIdx]
		configIdx += 1
		inputIdx += 1
		if configIdx >= len(argConfigs) {
			// If there's not another ArgConfig (indicating we've used them all) then we return nil to indicate
			//  that tab completion shouldn't be done (since no more args are needed)
			nextArg = nil
		} else {
			nextArg = argConfigs[configIdx]
		}
	}
	result := &ParsedArgs{
		nonGreedyArgs: nonGreedyArgValues,
		greedyArgs:    greedyArgValues,
	}
	return result, nextArg
}

// Parses all the args, guaranteeing that the required args are filled out and that default values for non-optional arguments get applied
// This means that if no error was returned, the returned ParsedArgs object is guaranteed to have all the args that were
//  passed in
func parseArgsForValidation(argConfigs []*ArgConfig, input []string) (*ParsedArgs, error) {
	nonGreedyArgValues := map[string]string{}
	greedyArgValues := map[string][]string{}
	inputIdx := 0
	for configIdx, config := range argConfigs {
		key := config.Key
		if config.IsOptional && configIdx < len(argConfigs) - 1 {
			return nil, stacktrace.NewError("Arg '%v' is marked as optional, but isn't the last argument; this is a bug in Kurtosis!", key)
		}
		if inputIdx >= len(input) {
			if !config.IsOptional {
				return nil, stacktrace.NewError("Missing required arg '%v'", config.Key)
			}
			if config.IsGreedy {
				 defaultVal, ok := config.DefaultValue.([]string)
				 if !ok {
					  return nil, stacktrace.NewError("Greedy arg '%v' wasn't provided, but the default value wasn't a string list; this is a bug in Kurtosis!", key)
				 }
				 greedyArgValues[key] = defaultVal
			} else {
				 defaultVal, ok := config.DefaultValue.(string)
				 if !ok {
					  return nil, stacktrace.NewError("Greedy arg '%v' wasn't provided, but the default value wasn't a string; this is a bug in Kurtosis!", key)
				 }
				 nonGreedyArgValues[key] = defaultVal
			}
			continue
		}

		// Greedy case (arg must always be last)
		if config.IsGreedy {
			greedyArgValues[key] = input[inputIdx:]
			inputIdx += len(input) - inputIdx
			break
		}

		nonGreedyArgValues[key] = input[inputIdx]
		inputIdx += 1
	}

	numArgTokenGroupings := len(greedyArgValues) + len(nonGreedyArgValues)
	if numArgTokenGroupings != len(argConfigs) {
		return nil, stacktrace.NewError(
			"Expected '%v' arg token groups, but got '%v'",
			len(argConfigs),
			numArgTokenGroupings,
		)
	}


	result := &ParsedArgs{
		nonGreedyArgs: nonGreedyArgValues,
		greedyArgs:    greedyArgValues,
	}
	return result, nil
}

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