package install

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/positional_arg_parser"
	"github.com/kurtosis-tech/kurtosis-core/commons/enclave_object_labels"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

const (
	enclaveIdArg        = "enclave-id"
	replGuidArg         = "repl-guid"
	packageIdentifierArg = "package-identifier"

	shouldFetchStoppedContainers = false

	installCmdSuccessExitCode = 0

	// vvvvvvvvvvvvvvvvvvvvvvvvvvv WARNING vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv
	// This must be added to the NODE_PATH in the Docker image, so if this location
	//  changes then the REPL Dockerfile must change!
	javascriptReplImageModulesDirpath = "/preinstalled-node-modules"
	// ^^^^^^^^^^^^^^^^^^^^^^^^^^^ WARNING ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
)

var positionalArgs = []string{
	enclaveIdArg,
	replGuidArg,
	packageIdentifierArg,
}

var InstallCmd = &cobra.Command{
	Use:   "install [flags] " + strings.Join(positionalArgs, " "),
	DisableFlagsInUseLine: true,
	Short: "Installs packages (identified by the same string as your package manager, e.g. 'web3@1.6.0' for Javascript) into the given REPL container so they'll be available in the REPL",
	RunE:  run,
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	parsedPositionalArgs, err := positional_arg_parser.ParsePositionalArgs(positionalArgs, args)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the positional args")
	}
	enclaveId := parsedPositionalArgs[enclaveIdArg]
	replGuid := parsedPositionalArgs[replGuidArg]
	packageIdentifier := parsedPositionalArgs[packageIdentifierArg]

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker client")
	}
	dockerManager := docker_manager.NewDockerManager(
		logrus.StandardLogger(),
		dockerClient,
	)

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

	cmdToExec := []string{
		"bash",
		"-c",
		fmt.Sprintf(
			"cd %v && npm install %v",
			javascriptReplImageModulesDirpath,
			packageIdentifier,
		),
	}

	logrus.Infof("Installing package '%v'...", packageIdentifier)
	cmdOutputBuffer := &bytes.Buffer{}
	exitCode, err := dockerManager.RunExecCommand(ctx, containerId, cmdToExec, cmdOutputBuffer)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred running install command '%v' on the REPL container",
			strings.Join(cmdToExec, " "),
		)
	}
	if exitCode != installCmdSuccessExitCode {
		return stacktrace.NewError(
			"Install command '%v' exited with non-%v exit code '%v' and the following logs:\n%v",
			strings.Join(cmdToExec, " "),
			installCmdSuccessExitCode,
			exitCode,
			cmdOutputBuffer.String(),
		)
	}
	logrus.Infof("Successfully installed package '%v'", packageIdentifier)

	return nil
}