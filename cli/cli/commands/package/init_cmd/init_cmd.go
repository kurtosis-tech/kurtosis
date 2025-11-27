package init_cmd

import (
	"context"
	"github.com/dzobbe/PoTE-kurtosis/api/golang/core/lib/shared_utils"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_str_consts"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/kurtosis_package"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/user_support_constants"
	"github.com/kurtosis-tech/stacktrace"
	"os"
)

const (
	packageNameArgKey          = "package-name"
	packageNameArgDefaultValue = "github.com/example-org/example-package"
	packageNameArgIsOptional   = true
	packageNameArgIsGreedy     = false

	alwaysCreateExecutablePackage = true

	validPackageNameExample = "github.com/ethpandaops/ethereum-package"
)

// InitCmd we only fill in the required struct fields, hence the others remain nil
// nolint: exhaustruct
var InitCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:       command_str_consts.InitCmdStr,
	ShortDescription: "Creates a new Kurtosis package",
	LongDescription:  "This command initializes the current directory to be a Kurtosis package by creating a `kurtosis.yml` with the given package name.",
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

	if err := kurtosis_package.InitializeKurtosisPackage(packageDestinationDirpath, packageNameArg, alwaysCreateExecutablePackage); err != nil {
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
		return stacktrace.Propagate(err, "An error occurred validating package name '%v', invalid GitHub URL, the package name has to be a valid GitHub URL like '%s'. You can see more here: '%s' ", packageNameArg, validPackageNameExample, user_support_constants.StarlarkPackagesReferenceURL)
	}

	return nil
}
