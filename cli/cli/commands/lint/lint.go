package lint

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/dzobbe/PoTE-kurtosis-package-indexer/server/crawler"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	fileOrDirToLintArgKey           = "file-or-dir"
	fileOrDirToLintArgKeyIsOptional = true
	fileOrDirToLintArgKeyIsGreedy   = true

	formatFlagKey          = "format"
	formatFlagShortKey     = "f"
	formatFlagDefaultValue = "false"

	checkDocStringFlagKey      = "check-docstring"
	checkDocStringFlagShortKey = "c"
	checkDocStringDefaultValue = "false"

	pyBlackDockerImage      = "pyfound/black:23.9.1"
	dockerRunCmd            = "run"
	removeContainerOnExit   = "--rm"
	dockerBinary            = "docker"
	lintVolumeName          = "/lint"
	dockerVolumeFlag        = "-v"
	dockerWorkDirFlag       = "--workdir"
	blackBinaryName         = "black"
	includeFlagForBlack     = "--include"
	checkFlagForBlack       = "--check"
	allStarlarkFilesMatch   = "\\.star?$"
	dirVolumeSeparator      = ":"
	presentWorkingDirectory = "."
	versionArg              = "version"

	mainDotStarFilename = "main.star"

	linterFailedAsThingsNeedToBeReformattedExitCode = 1
	linterFailedWithInternalErrorsExitCode          = 123
)

var fileOrDirToLintDefaultValue = []string{"."}

var dockerRunPrefix = []string{dockerRunCmd, removeContainerOnExit, dockerVolumeFlag}
var dockerRunSuffix = []string{dockerWorkDirFlag, lintVolumeName, pyBlackDockerImage, blackBinaryName, includeFlagForBlack, allStarlarkFilesMatch}

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
		{
			Key:       checkDocStringFlagKey,
			Usage:     fmt.Sprintf("Use this flag to check whether the doc string is valid or not. This requires you to pass a single path to a '%v' or a directory containing a '%v'", mainDotStarFilename, mainDotStarFilename),
			Shorthand: checkDocStringFlagShortKey,
			Type:      flags.FlagType_Bool,
			Default:   checkDocStringDefaultValue,
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

	checkDocStringFlag, err := flags.GetBool(checkDocStringFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "an error occurred getting the value of the flag '%v'", checkDocStringFlagKey)
	}

	if checkDocStringFlag {
		if err = validateDocString(fileOrDirToLintArg); err != nil {
			return stacktrace.Propagate(err, "an error occurred while running the doc string validator")
		}
	}

	logrus.Infof("This depends on '%v'; first run may take a while as we might have to download it", pyBlackDockerImage)

	if _, err := exec.LookPath(dockerBinary); err != nil {
		return stacktrace.Propagate(err, "'%v' uses '%v' underneath in order to use the '%v' image but it couldn't find '%v' in path", command_str_consts.KurtosisLintCmdStr, dockerBinary, pyBlackDockerImage, dockerBinary)
	}

	versionCommand := exec.Command(dockerBinary, versionArg)
	if err := versionCommand.Run(); err != nil {
		return stacktrace.Propagate(err, "An error occurred checking Docker version. Please ensure Docker engine is running and try again.")
	}

	for _, fileOrDirToLint := range fileOrDirToLintArg {
		logrus.Infof("Linting '%v'", fileOrDirToLint)
		volumeToMount, pathToLint, err := getVolumeToMountAndPathToLint(fileOrDirToLint)
		if err != nil {
			return stacktrace.Propagate(err, "an error occurred while attempting to parse the volume to mount and file to lint for path '%v'", fileOrDirToLint)
		}
		commandArgs := append(dockerRunPrefix, volumeToMount+dirVolumeSeparator+lintVolumeName)
		dockerRunSuffix = append(dockerRunSuffix, pathToLint)
		commandArgs = append(commandArgs, dockerRunSuffix...)
		cmd := exec.Command(dockerBinary, commandArgs...)
		logrus.Debugf("Running command '%v'", cmd.String())
		cmdOutput, err := cmd.CombinedOutput()
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				fmt.Println(string(cmdOutput))
				switch exitError.ExitCode() {
				case linterFailedAsThingsNeedToBeReformattedExitCode:
					return stacktrace.NewError("linting failed, this means that there are some files that need to be formatted, run this command with the '--%v' flag", formatFlagKey)
				case linterFailedWithInternalErrorsExitCode:
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

func getVolumeToMountAndPathToLint(pathOfFileOrDirToLint string) (string, string, error) {
	fileInfo, err := os.Stat(pathOfFileOrDirToLint)
	if err != nil {
		return "", "", stacktrace.Propagate(err, "an error occurred while verifying that '%v' exist", pathOfFileOrDirToLint)
	}
	absolutePathForFileOrDirToLint, err := filepath.Abs(pathOfFileOrDirToLint)
	if err != nil {
		return "", "", stacktrace.Propagate(err, "tried to get absolute path for dir to lint '%v but failed'", absolutePathForFileOrDirToLint)
	}

	if fileInfo.IsDir() {
		return absolutePathForFileOrDirToLint, presentWorkingDirectory, nil
	} else {
		return path.Dir(absolutePathForFileOrDirToLint), path.Base(absolutePathForFileOrDirToLint), nil
	}
}

func validateDocString(fileOrDirToLintArg []string) error {
	if len(fileOrDirToLintArg) != 1 {
		return stacktrace.NewError("Doc string validation only works with one argument, either a full path to a '%v' file or a directory containing it got '%v' arguments", mainDotStarFilename, len(fileOrDirToLintArg))
	}

	fileOrDirToCheckForDocString := fileOrDirToLintArg[0]
	lastPartOfFileToCheck := path.Base(fileOrDirToCheckForDocString)

	if lastPartOfFileToCheck == mainDotStarFilename {
		return parseDocString(fileOrDirToCheckForDocString)
	}

	fileInfo, _ := os.Stat(fileOrDirToCheckForDocString)
	if !fileInfo.IsDir() {
		return stacktrace.NewError("Passed argument '%v' isn't a '%v' file nor is it a directory", fileOrDirToLintArg[0], mainDotStarFilename)
	}

	entries, err := os.ReadDir(fileOrDirToCheckForDocString)
	if err != nil {
		return stacktrace.Propagate(err, "Couldn't read directory '%v' for doc string validation", fileOrDirToCheckForDocString)
	}

	for _, entry := range entries {
		if entry.Name() == mainDotStarFilename {
			return parseDocString(path.Join(fileOrDirToCheckForDocString, entry.Name()))
		}
	}

	return stacktrace.Propagate(err, "Couldn't find a '%v' file in the passed directory '%v'", mainDotStarFilename, fileOrDirToCheckForDocString)
}

func parseDocString(mainStarFilepath string) error {
	contents, err := os.ReadFile(mainStarFilepath)
	if err != nil {
		return stacktrace.Propagate(err, "")
	}
	contentsAsStr := string(contents)
	currentOutput := logrus.StandardLogger().Out
	defer logrus.SetOutput(currentOutput)
	// we disable the output as the crawler pollutes the screen otherwise
	logrus.SetOutput(io.Discard)
	if _, err := crawler.ParseMainDotStarContent(contentsAsStr); err != nil {
		return stacktrace.Propagate(err, "an error occurred while parsing the doc string")
	}
	return nil
}
