package lint

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/stacktrace"
	"os"
	"os/exec"
	"strings"
)

import _ "embed"

// go:embed black/*

const (
	fileOrDirToLintArgKey           = "file-or-dir"
	fileOrDirToLintArgKeyIsOptional = true
	fileOrDirToLintArgKeyIsGreedy   = true
	checkFlagKey                    = "check"
	checkFlagForBlack               = "--check"
	checkFlagShortKey               = "c"
	checkFlagDefaultValue           = "false"
	cmdArgsSeparator                = " "
)

var fileOrDirToLintDefaultValue = []string{"."}
var possiblePythonBinaries = []string{"python", "python3"}

var flagsForBlack = []string{"black/__main__.py", "--include='\\.star?$'"}

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
			Key:       checkFlagKey,
			Usage:     fmt.Sprintf("Passes the %v flag to black; doesn't change the files and just reports the status", checkFlagForBlack),
			Shorthand: checkFlagShortKey,
			Type:      flags.FlagType_Bool,
			Default:   checkFlagDefaultValue,
		},
	},
	RunFunc: run,
}

func run(_ context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {
	fileOrDirToLintArg, err := args.GetGreedyArg(fileOrDirToLintArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "an error occurred getting the value of argument with key '%v'", fileOrDirToLintArgKey)
	}

	checkFlag, err := flags.GetBool(checkFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "an error occurred getting the value of the '%v' flag", checkFlagKey)
	}

	var pythonBinaryToUse string
	foundPythonBinaryInPath := false
	for _, possiblePythonBinary := range possiblePythonBinaries {
		if _, err := exec.LookPath(possiblePythonBinary); err != nil {
			pythonBinaryToUse = possiblePythonBinary
			foundPythonBinaryInPath = true
		}
	}

	if !foundPythonBinaryInPath {
		return stacktrace.NewError("Tried looking for the following python binaries '%v' but found none; one of them has to exist for lint to work", possiblePythonBinaries)
	}

	if checkFlag {
		flagsForBlack = append(flagsForBlack, checkFlagForBlack)
	}

	for _, fileOrDirToLint := range fileOrDirToLintArg {
		flagsForBlackWithFile := append(flagsForBlack, fileOrDirToLint)
		cmd := exec.Command(pythonBinaryToUse, flagsForBlackWithFile...)
		if err = cmd.Run(); err != nil {
			return stacktrace.Propagate(err, "An error occurred while running the command '%v %v'", cmd.Path, strings.Join(cmd.Args, cmdArgsSeparator))
		}
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
