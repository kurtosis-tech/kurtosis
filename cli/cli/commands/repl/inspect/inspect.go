package inspect

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/version_checker"
	"github.com/kurtosis-tech/kurtosis-cli/commons/positional_arg_parser"
	"github.com/kurtosis-tech/kurtosis-cli/commons/repl_consts"
	"github.com/kurtosis-tech/kurtosis-core/commons/enclave_object_labels"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"path"
	"sort"
	"strings"
)

const (
	enclaveIdArg        = "enclave-id"
	replGuidArg         = "repl-guid"

	shouldFetchStoppedContainers = false

	listPackagesCmdExitCode = 0

	installedPackagesTitleKey = "Installed Packages"

	hiddenDirectoryPrefix = "."
)

var positionalArgs = []string{
	enclaveIdArg,
	replGuidArg,
}

var InspectCmd = &cobra.Command{
	Use:   command_str_consts.ReplInspectCmdStr + " [flags] " + strings.Join(positionalArgs, " "),
	DisableFlagsInUseLine: true,
	Short: "Lists detailed information about the given REPL",
	RunE:  run,
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	version_checker.CheckLatestVersion()

	parsedPositionalArgs, err := positional_arg_parser.ParsePositionalArgsAndRejectEmptyStrings(positionalArgs, args)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the positional args")
	}
	enclaveId := parsedPositionalArgs[enclaveIdArg]
	replGuid := parsedPositionalArgs[replGuidArg]

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker client")
	}
	dockerManager := docker_manager.NewDockerManager(
		logrus.StandardLogger(),
		dockerClient,
	)

	// TODO Replace with a better API for getting REPL containers, so that users don't have to construct the labels map manually
	replContainerLabels := map[string]string{
		enclave_object_labels.EnclaveIDContainerLabel: enclaveId,
		enclave_object_labels.ContainerTypeLabel: enclave_object_labels.ContainerTypeInteractiveREPL,
		enclave_object_labels.GUIDLabel: replGuid,
	}
	matchingContainers, err := dockerManager.GetContainersByLabels(ctx, replContainerLabels, shouldFetchStoppedContainers)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting interactive REPL containers matching labels: %+v", replContainerLabels)
	}
	if len(matchingContainers) == 0 {
		return stacktrace.NewError("No running REPL containers matched labels: %+v", replContainerLabels)
	}
	if len(matchingContainers) > 1 {
		return stacktrace.NewError("More than one running REPL container matched the following labels; this is a bug in Kurtosis: %+v", replContainerLabels)
	}
	matchingContainer := matchingContainers[0]

	containerId := matchingContainer.GetId()

	// TODO Get this from the REPL itself somehow - likely using labels!!
	replType := repl_consts.ReplType_Javascript

	installedPackagesDirpath, found := repl_consts.InstalledPackagesDirpath[replType]
	if !found {
		return stacktrace.NewError("No installed packages dirpath defined for REPL type '%v' - this is a bug in Kurtosis", replType)
	}

	cmdToExec := []string{
		"find",
		installedPackagesDirpath,
		"-type",
		"d",
		"-mindepth",
		"1",
		"-maxdepth",
		"1",
	}

	cmdOutputBuffer := &bytes.Buffer{}
	exitCode, err := dockerManager.RunExecCommand(ctx, containerId, cmdToExec, cmdOutputBuffer)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred running package-listing command '%v' on the REPL container",
			strings.Join(cmdToExec, " "),
		)
	}
	cmdOutputStr := cmdOutputBuffer.String()
	if exitCode != listPackagesCmdExitCode {
		return stacktrace.NewError(
			"Package-listing command '%v' exited with non-%v exit code '%v' and the following logs:\n%v",
			strings.Join(cmdToExec, " "),
			listPackagesCmdExitCode,
			exitCode,
			cmdOutputStr,
		)
	}

	allPackageNames := []string{}
	for _, packageFilepath := range strings.Split(cmdOutputStr, "\n") {
		packageName := path.Base(packageFilepath)
		if strings.HasPrefix(packageName, hiddenDirectoryPrefix) {
			continue
		}
		allPackageNames = append(allPackageNames, packageName)
	}
	sort.Strings(allPackageNames)

	fmt.Fprintln(logrus.StandardLogger().Out, installedPackagesTitleKey+ ":")
	for _, packageName := range allPackageNames {
		fmt.Fprintln(logrus.StandardLogger().Out, packageName)
	}

	return nil
}
