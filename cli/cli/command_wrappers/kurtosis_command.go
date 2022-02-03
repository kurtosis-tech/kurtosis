package command_wrappers

import (
	"fmt"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"strings"
)

const (
	// It's pretty weird that this flag - which is a global flag that gets set on the root command - lives here
	//  rather than in the root package. Unfortunately, the root package will need to import commands that import
	//  this package, so if this package depends on the root to get this constant then we get an import cycle.
	// That said, it *does* make some amount of sense to live here: this flag will be set on all commands, and all
	//  commands will have completion, and we need to check this flag to see if we log completion errors to STDERR.
	CLILogLevelStrFlag = "cli-log-level"
)

// TODO Maybe better to make several different types of flags here - one for each type of value
type FlagConfig struct {
	Key string

	// TODO Add flag defaults!!!

	// TODO Use this!
	ValidationFunc func(string) error
}

type ArgConfig struct {
	Key string

	// If set to true, this argument can be ommitted
	IsOptional bool

	// For non-greedy args, this must be a string
	// For greedy args, this must be a []string
	// Has no effect in non-optional arguments!
	DefaultValue interface{}

	// If set to true, this arg will consume all the remaining arg strings
	IsGreedy bool

	// This function is for generating the valid completions that the shell can use for this arg, and can be dynamically
	//  modified using previous arg values
	// The previousArgs will only contain values for the args that come before this one (since completion doesn't make
	//  sense in the middle of an entry
	CompletionsFunc func(flags *ParsedFlags, previousArgs *ParsedArgs) ([]string, error)

	// Will be run after the user presses ENTER and before we start actually running the command
	ValidationFunc func(flags *ParsedFlags, args *ParsedArgs) error
}

// This is a struct that can take higher-level, Kurtosis-specific information (e.g. "this command takes in an enclave ID")
//  and generate the low-level cobra.Command corresponding to that information (with autogenerated usage information, etc.)
type KurtosisCommand struct {
	// The string for the command (e.g. "inspect" or "ls")
	CommandStr string

	ShortDescription string

	LongDescription string

	// Order isn't important here
	Flags []*FlagConfig

	// Order IS important
	Args []*ArgConfig

	RunFunc func(flags *ParsedFlags, args *ParsedArgs) error
}

type ParsedFlags struct {
	cmdFlagsSet *pflag.FlagSet
}
func (flags *ParsedFlags) GetString(name string) (string, error) {
	value, err := flags.cmdFlagsSet.GetString(name)
	if err != nil {
		return "", stacktrace.Propagate(
			err,
			"An error occurred getting string flag '%v'; this is a bug in Kurtosis!",
			name,
		)
	}
	return value, nil
}
func (flags *ParsedFlags) GetUint32(name string) (uint32, error) {
	value, err := flags.cmdFlagsSet.GetUint32(name)
	if err != nil {
		return 0, stacktrace.Propagate(err,
			"An error occurred getting uint32 flag '%v'; this is a bug in Kurtosis!",
			name,
		)
	}
	return value, nil
}

// WILL do validation on args
type ParsedArgs struct {
	nonGreedyArgs map[string]string
	greedyArgs    map[string][]string
}
// Non-greedy args only have a single value
func (args *ParsedArgs) GetNonGreedyArg(key string) (string, error) {
	result, found := args.nonGreedyArgs[key]
	if !found {
		return "", stacktrace.NewError("No non-greedy arg with key '%v'", key)
	}
	return result, nil
}
// Greedy args can have multiple values
func (args *ParsedArgs) GetGreedyArg(key string) ([]string, error) {
	elems, found := args.greedyArgs[key]
	if !found {
		return nil, stacktrace.NewError("No greedy arg with key '%v'", key)
	}
	return elems, nil
}

// Gets the Cobra command, and panics if there's an error because this is intended to be run inside the init() during
//  initialization of the code
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
	}

	// Verify that we don't have any invalid positional arg combinations, e.g.:
	//  - Any arg after an optional arg (the parser wouldn't know whether you want the optional arg or the one after it)
	//  - Any arg after an arg that consumes N args (since the CLI couldn't know where
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
		shouldPrintDebuggingMessagesToStderr := false

		// This will be set by the root function as a persistent flag
		// We need to check this value manually - rather than relying on whether logrus.GetLevel is debug - because
		//  the logrus debug level is only set in the root command's PersistentPreRunE function, which is set only when
		//  executing the commands - NOT when doing completion
		cliLogLevelStr, err := cmd.Flags().GetString(CLILogLevelStrFlag)
		if err != nil {
			// TODO Give them a link to file on our Github!
			cobra.CompErrorln(fmt.Sprintf(
				"An error occurred getting the value of the CLI log level flag '%v' to check if we should print further completion messages to STDERR; this is a bug in Kurtosis!\n%v",
				CLILogLevelStrFlag,
				err,
			))
		} else {
			shouldPrintDebuggingMessagesToStderr = cliLogLevelStr == logrus.DebugLevel.String()
		}

		parsedFlags := &ParsedFlags{
			cmdFlagsSet: cmd.Flags(),
		}

		parsedArgs, argToComplete := parseArgsForCompletion(kurtosisCmd.Args, previousArgStrs)
		if argToComplete == nil {
			// NOTE: We can't just use logrus because anything printed to STDOUT will be interpreted as a completion
			// See:
			//  https://github.com/spf13/cobra/blob/master/shell_completions.md#:~:text=the%20RunE%20function.-,Debugging,ShellCompDirectiveNoFileComp%20%23%20This%20is%20on%20stderr
			cobra.CompDebugln("Not completing because no argument needs completion", shouldPrintDebuggingMessagesToStderr)
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
				shouldPrintDebuggingMessagesToStderr,
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
				shouldPrintDebuggingMessagesToStderr,
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
		DisableFlagsInUseLine: true, // Not needed since we manually add the string in the uasge string
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
// This means that the returned ParsedArgs object is guaranteed to have all the args that were passed in
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