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
	"strings"
)

const (
	fileOrDirToLintArgKey           = "file-or-dir"
	fileOrDirToLintArgKeyIsOptional = true
	fileOrDirToLintArgKeyIsGreedy   = true
	cmdArgsSeparator                = " "

	formatFlagKey          = "format"
	formatFlagShortKey     = "f"
	formatFlagDefaultValue = "false"

	pyBlackDockerImage    = "pyfound/black:23.9.1"
	dockerRunCmd          = "run"
	removeContainerOnExit = "--rm"
	dockerBinary          = "docker"
	lintVolumeName        = "lint"
)

var fileOrDirToLintDefaultValue = []string{"."}

var dockerRunPrefix = []string{dockerRunCmd, removeContainerOnExit, "-v"}
var dockerRunSuffix = []string{"--workdir", "/" + lintVolumeName, pyBlackDockerImage, "black", ".", "--include", "\\.star?$"}

// LintCmd we only fill in the required struct fields, hence the others remain nil
// nolint: exhaustruct
var LintCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:       command_str_consts.KurtosisLintCmdStr,
	ShortDescription: "Lints the Kurtosis package or file",
	LongDescription: fmt.Sprintf(
		"Lints the Kurtosis package or file"),

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
	if !formatFlag {
		dockerRunSuffix = append(dockerRunSuffix, "--check")
	}

	logrus.Infof("The first run might take a few seconds as we depend on the '%v' image and have to download it", pyBlackDockerImage)

	if _, err := exec.LookPath(dockerBinary); err != nil {
		return stacktrace.Propagate(err, "'%v' uses '%v' underneath in order to use the '%v' image but it couldn't find '%v' in path", command_str_consts.KurtosisLintCmdStr, dockerBinary, pyBlackDockerImage, dockerBinary)
	}

	for _, fileOrDirToLint := range fileOrDirToLintArg {
		commandArgs := append(dockerRunPrefix, fileOrDirToLint+":/"+lintVolumeName)
		commandArgs = append(commandArgs, dockerRunSuffix...)
		cmd := exec.Command(dockerBinary, commandArgs...)
		cmdOutput, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println(string(cmdOutput))
			return stacktrace.Propagate(err, "An error occurred while running the command '%v'", strings.Join(cmd.Args, cmdArgsSeparator))
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
