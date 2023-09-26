package lint

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
)

const (
	fileOrDirToLintArgKey           = "file-or-dir"
	fileOrDirToLintArgKeyIsOptional = true
	fileOrDirToLintArgKeyIsGreedy   = true

	formatFlagKey          = "format"
	formatFlagShortKey     = "f"
	formatFlagDefaultValue = "false"

	pyBlackDockerImage      = "pyfound/black:23.9.1"
	dockerRunCmd            = "run"
	removeContainerOnExit   = "--rm"
	dockerBinary            = "docker"
	lintVolumeName          = "/lint"
	dockerVolumeFlag        = "-v"
	dockerWorkDirFlag       = "--workdir"
	blackBinaryName         = "black"
	includeFlagForBlack     = "--include"
	presentWorkingDirectory = "."
	checkFlagForBlack       = "--check"
	allStarlarkFilesMatch   = "\\.star?$"
	dirVolumeSeparator      = ":"

	linterFailedAsThingsNeedToBeReformattedExitCode = 1
	linterFailedWithInternalErrorsExitCode          = 123
)

var fileOrDirToLintDefaultValue = []string{"."}

var dockerRunPrefix = []string{dockerRunCmd, removeContainerOnExit, dockerVolumeFlag}
var dockerRunSuffix = []string{dockerWorkDirFlag, lintVolumeName, pyBlackDockerImage, blackBinaryName, presentWorkingDirectory, includeFlagForBlack, allStarlarkFilesMatch}

// LintCmd we only fill in the required struct fields, hence the others remain nil
// nolint: exhaustruct
var LintCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:       command_str_consts.KurtosisLintCmdStr,
	ShortDescription: "Lints the Kurtosis package or file",
	LongDescription:  "Lints the Kurtosis package or file",

	Args: []*args.ArgConfig{
		{
			Key:            fileOrDirToLintArgKey,
			DefaultValue:   fileOrDirToLintDefaultValue,
			IsOptional:     fileOrDirToLintArgKeyIsOptional,
			IsGreedy:       fileOrDirToLintArgKeyIsGreedy,
			ValidationFunc: validateFileOrDirToLintArg,
		},
	},

	Flags: []*flags.FlagConfig{
		{
			Key:       formatFlagKey,
			Usage:     "Use this flag to edit files in place instead of just verifying whether the formatting is correct",
			Shorthand: formatFlagShortKey,
			Type:      flags.FlagType_Bool,
			Default:   formatFlagDefaultValue,
		},
	},
	RunFunc: run,
}

func run(_ context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {
	fileOrDirToLintArg, err := args.GetGreedyArg(fileOrDirToLintArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "an error occurred getting the value of argument with key '%v'", fileOrDirToLintArgKey)
	}

	formatFlag, err := flags.GetBool(formatFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "an error occurred getting the value of flag '%v'", formatFlag)
	}
	if !formatFlag {
		dockerRunSuffix = append(dockerRunSuffix, checkFlagForBlack)
	}

	logrus.Infof("This depends on '%v'; first run may take a while as we might have to download it", pyBlackDockerImage)

	if _, err := exec.LookPath(dockerBinary); err != nil {
		return stacktrace.Propagate(err, "'%v' uses '%v' underneath in order to use the '%v' image but it couldn't find '%v' in path", command_str_consts.KurtosisLintCmdStr, dockerBinary, pyBlackDockerImage, dockerBinary)
	}

	for _, fileOrDirToLint := range fileOrDirToLintArg {
		logrus.Infof("Linting '%v'", fileOrDirToLint)
		commandArgs := append(dockerRunPrefix, fileOrDirToLint+dirVolumeSeparator+lintVolumeName)
		commandArgs = append(commandArgs, dockerRunSuffix...)
		cmd := exec.Command(dockerBinary, commandArgs...)
		cmdOutput, err := cmd.CombinedOutput()
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				switch exitError.ExitCode() {
				case linterFailedAsThingsNeedToBeReformattedExitCode:
					fmt.Println(string(cmdOutput))
					return stacktrace.NewError("linting failed, this means that there are some files that need to be formatted, run this command with the '--%v' flag", formatFlagKey)
				case linterFailedWithInternalErrorsExitCode:
					fmt.Println(string(cmdOutput))
					return stacktrace.NewError("linting failed with an internal error please look at the output to see why; usually this happens if there's a mix of spaces & tabs")
				default:
					return stacktrace.Propagate(err, "linting failed with an unexpected exit code '%v'; This is a bug in Kurtosis", exitError.ExitCode())
				}
			}
			return stacktrace.Propagate(err, "Linting failed and we couldn't get an exit code out of the err; This is a bug in Kurtosis")
		}
		fmt.Println(string(cmdOutput))
	}

	return nil
}

func validateFileOrDirToLintArg(_ context.Context, _ *flags.ParsedFlags, args *args.ParsedArgs) error {
	fileOrDirToLintArg, err := args.GetGreedyArg(fileOrDirToLintArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "an error occurred getting the value of argument with key '%v'", fileOrDirToLintArgKey)
	}

	for _, fileOrDirToLint := range fileOrDirToLintArg {
		if _, err := os.Stat(fileOrDirToLint); err != nil {
			return stacktrace.Propagate(err, "an error occurred validating whether supplied path '%v' was valid", fileOrDirToLint)
		}
	}

	return nil
}
