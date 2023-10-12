package init_cmd

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/shared_utils"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_package"
	"github.com/kurtosis-tech/stacktrace"
	"os"
)

const (
	packageNameArgKey          = "package-name"
	packageNameArgDefaultValue = ""
	packageNameArgIsOptional   = false
	packageNameArgIsGreedy     = false
)

var InitCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:       command_str_consts.InitCmdStr,
	ShortDescription: "Creates a new Kurtosis package",
	LongDescription:  "This command initializes the current directory to be a Kurtosis package by creating a `kurtosis.yml` with the given package name.",
	Flags:            nil,
	Args: []*args.ArgConfig{
		{
			Key:            packageNameArgKey,
			DefaultValue:   packageNameArgDefaultValue,
			IsOptional:     packageNameArgIsOptional,
			IsGreedy:       packageNameArgIsGreedy,
			ValidationFunc: validatePackageNameArg,
		},
	},
	PreValidationAndRunFunc:  nil,
	RunFunc:                  run,
	PostValidationAndRunFunc: nil,
}

func run(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {
	packageNameArg, err := args.GetNonGreedyArg(packageNameArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "an error occurred getting the value of argument with key '%v'", packageNameArgKey)
	}

	packageDestinationDirpath, err := os.Getwd()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the current working directory for creating the Kurtosis package")
	}

	if err := kurtosis_package.InitializeKurtosisPackage(packageDestinationDirpath, packageNameArg); err != nil {
		return stacktrace.Propagate(err, "An error occurred initializing the Kurtosis package '%s' in '%s'", packageNameArg, packageDestinationDirpath)
	}

	return nil
}

func validatePackageNameArg(_ context.Context, _ *flags.ParsedFlags, args *args.ParsedArgs) error {
	packageNameArg, err := args.GetNonGreedyArg(packageNameArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "an error occurred getting the value of argument with key '%v'", packageNameArgKey)
	}

	if _, err := shared_utils.ParseGitURL(packageNameArg); err != nil {
		return stacktrace.Propagate(err, "An erro occurred validating package name '%v', invalid GitHub URL", packageNameArg)
	}

	return nil
}
