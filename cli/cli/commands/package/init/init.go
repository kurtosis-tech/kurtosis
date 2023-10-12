package init

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/host_machine_directories"
	"github.com/kurtosis-tech/kurtosis/cli/cli/out"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	packageNameArgKey = 
)

var InitCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:               command_str_consts.InitCmdStr,
	ShortDescription:         "Creates a new Kurtosis package",
	LongDescription:          "This command initializes the current directory to be a Kurtosis package by creating a `kurtosis.yml` with the given package name.",
	Flags:                    nil,
	Args: []*args.ArgConfig{
		{
			Key:            fileOrDirToLintArgKey,
			DefaultValue:   fileOrDirToLintDefaultValue,
			IsOptional:     fileOrDirToLintArgKeyIsOptional,
			IsGreedy:       fileOrDirToLintArgKeyIsGreedy,
			ValidationFunc: validateFileOrDirToLintArg,
		},
	},
	PreValidationAndRunFunc:  nil,
	RunFunc:                  run,
	PostValidationAndRunFunc: nil,
}

func run(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {
	configFilepath, err := host_machine_directories.GetKurtosisConfigYAMLFilepath()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the Kurtosis config filepath")
	}
	out.PrintOutLn(configFilepath)
	return nil
}
