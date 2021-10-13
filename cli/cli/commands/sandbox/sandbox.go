/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package sandbox

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/best_effort_image_puller"
	"github.com/kurtosis-tech/kurtosis-cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis-cli/cli/enclave_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/enclave_manager/enclave_context"
	"github.com/kurtosis-tech/kurtosis-cli/cli/execution_ids"
	"github.com/kurtosis-tech/kurtosis-cli/cli/logrus_log_levels"
	"github.com/kurtosis-tech/kurtosis-client/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-client/golang/kurtosis_core_rpc_api_consts"
	"github.com/kurtosis-tech/kurtosis-core/commons/current_time_str_provider"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/sys/unix"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"io"
	"net"
	"os"
	"strings"
)

const (
	enclaveDataVolMountpointOnReplContainer = "/kurtosis-enclave-data"


	shouldPublishPorts = true

	isPartitioningEnabled = true

	apiContainerImageArg = "kurtosis-api-image"
	javascriptReplImageArg = "javascript-repl-image"
	kurtosisLogLevelArg = "kurtosis-log-level"

	// This is the directory in which the node REPL is running inside the REPL container, which is where
	//  we'll bind-mount the host machine's current directory into the container so the user can access
	//  files on their host machine
	workingDirpathInsideReplContainer = "/repl"

	replContainerSuccessExitCode = 0

	// WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING
	// vvvvvvvvvvvvvvv If you change these, update the REPL Dockerfile!!! vvvvvvvvvvvv
	replContainerKurtosisSocketEnvVar = "KURTOSIS_API_SOCKET"
	replContainerEnclaveDataVolMountpointEnvVar = "ENCLAVE_DATA_VOLUME_MOUNTPOINT"
	// ^^^^^^^^^^^^^^^ If you change these, update the REPL Dockerfile!!! ^^^^^^^^^^^^
	// WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING
)
var defaultKurtosisLogLevel = logrus.InfoLevel.String()

var SandboxCmd = &cobra.Command{
	Use:   "sandbox",
	Short: "Creates a new Kurtosis enclave and attaches a REPL for manipulating it",
	RunE:  run,
}

var kurtosisLogLevelStr string
var apiContainerImage string
var jsReplImage string


func init() {
	SandboxCmd.Flags().StringVarP(
		&kurtosisLogLevelStr,
		kurtosisLogLevelArg,
		"l",
		defaultKurtosisLogLevel,
		fmt.Sprintf(
			"The log level that Kurtosis itself should log at (%v)",
			strings.Join(logrus_log_levels.GetAcceptableLogLevelStrs(), "|"),
		),
	)

	SandboxCmd.Flags().StringVarP(
		&apiContainerImage,
		apiContainerImageArg,
		"a",
		defaults.DefaultApiContainerImage,
		"The image of the Kurtosis API container to use inside the enclave",
	)

	SandboxCmd.Flags().StringVarP(
		&jsReplImage,
		javascriptReplImageArg,
		"r",
		defaults.DefaultJavascriptReplImage,
		"The image of the Javascript REPL to connect to the enclave with",
	)
}

func run(cmd *cobra.Command, args []string) error {
	kurtosisLogLevel, err := logrus.ParseLevel(kurtosisLogLevelStr)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing Kurtosis loglevel string '%v' to a log level object", kurtosisLogLevelStr)
	}
	logrus.SetLevel(kurtosisLogLevel)

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker client")
	}
	dockerManager := docker_manager.NewDockerManager(
		logrus.StandardLogger(),
		dockerClient,
	)

	best_effort_image_puller.PullImageBestEffort(context.Background(), dockerManager, apiContainerImage)
	best_effort_image_puller.PullImageBestEffort(context.Background(), dockerManager, jsReplImage)

	enclaveId := execution_ids.GetExecutionID()

	enclaveManager := enclave_manager.NewEnclaveManager(dockerClient, apiContainerImage)

	enclaveCtx, err := enclaveManager.CreateEnclave(
		context.Background(),
		logrus.StandardLogger(),
		kurtosisLogLevel,
		enclaveId,
		isPartitioningEnabled,
		shouldPublishPorts,
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating an enclave")
	}
	defer func() {
		// Ensure we don't leak enclaves
		logrus.Info("Removing enclave...")
		if err := enclaveManager.DestroyEnclave(context.Background(), logrus.StandardLogger(), enclaveCtx); err != nil {
			logrus.Errorf("An error occurred destroying enclave '%v' that the interactive environment was connected to:", enclaveId)
			fmt.Fprintln(logrus.StandardLogger().Out, err)
			logrus.Errorf("ACTION REQUIRED: You'll need to clean this up manually!!!!")
		} else {
			logrus.Info("Enclave removed")
		}
	}()

	logrus.Debug("Running REPL...")
	if err := runReplContainer(dockerManager, enclaveCtx, jsReplImage); err != nil {
		return stacktrace.Propagate(err, "An error occurred running the REPL container")
	}
	logrus.Debug("REPL exited")

	return nil
}

func runReplContainer(
	dockerManager *docker_manager.DockerManager,
	enclaveCtx *enclave_context.EnclaveContext,
	javascriptReplImage string,
) error {
	enclaveId := enclaveCtx.GetEnclaveID()
	networkId := enclaveCtx.GetNetworkID()
	apiContainerHostMachinePortBinding := enclaveCtx.GetAPIContainerHostPortBinding()
	kurtosisApiContainerIpAddr := enclaveCtx.GetAPIContainerIPAddr()
	enclaveObjNameProvider := enclaveCtx.GetObjectNameProvider()
	enclaveObjLabelProvider := enclaveCtx.get

	apiContainerUrlOnHostMachine := fmt.Sprintf(
		"%v:%v",
		apiContainerHostMachinePortBinding.HostIP,
		apiContainerHostMachinePortBinding.HostPort,
	)
	conn, err := grpc.Dial(apiContainerUrlOnHostMachine, grpc.WithInsecure())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred dialling the API container via its host machine port binding")
	}
	defer conn.Close()
	apiContainerClient := kurtosis_core_rpc_api_bindings.NewApiContainerServiceClient(conn)

	startRegistrationResp, err := apiContainerClient.StartExternalContainerRegistration(context.Background(), &emptypb.Empty{})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred starting the registration of the interactive REPL container")
	}
	replContainerRegistrationKey := startRegistrationResp.RegistrationKey
	replContainerIpAddrStr := startRegistrationResp.IpAddr
	replContainerIpAddr := net.ParseIP(replContainerIpAddrStr)
	if replContainerIpAddr == nil {
		return stacktrace.NewError(
			"Received an IP, '%v', from the API container to give the interactive REPL container, but it wasn't parseable to an IP",
			replContainerIpAddr,
		)
	}

	stdoutFd := int(os.Stdout.Fd())
	windowSize, err := unix.IoctlGetWinsize(stdoutFd, unix.TIOCGWINSZ)
	if err != nil {
		return stacktrace.NewError("An error occurred getting the current terminal window size")
	}
	interactiveModeTtySize := &docker_manager.InteractiveModeTtySize{
		Height: uint(windowSize.Row),
		Width:  uint(windowSize.Col),
	}

	// Map of host_path -> path_on_container that will mounted
	var bindMounts map[string]string
	hostMachineWorkingDirpath, err := os.Getwd()
	if err != nil {
		logrus.Warn("Couldn't get the current working directory; local files will not be available")
		bindMounts = map[string]string{}
	} else {
		bindMounts = map[string]string{
			hostMachineWorkingDirpath: workingDirpathInsideReplContainer,
		}
	}

	interactiveReplGuid := current_time_str_provider.GetCurrentTimeStr()


	kurtosisApiContainerSocket := fmt.Sprintf("%v:%v", kurtosisApiContainerIpAddr, kurtosis_core_rpc_api_consts.ListenPort)
	containerName := enclaveObjNameProvider.ForInteractiveREPLContainer(interactiveReplGuid)
	labels := enclaveObjLa
	// TODO Add interactive labels!!!
	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		javascriptReplImage,
		containerName,
		networkId,
	).WithInteractiveModeTtySize(
		interactiveModeTtySize,
	).WithStaticIP(
		replContainerIpAddr,
	).WithEnvironmentVariables(map[string]string{
		replContainerKurtosisSocketEnvVar:            kurtosisApiContainerSocket,
		replContainerEnclaveDataVolMountpointEnvVar: enclaveDataVolMountpointOnReplContainer,
	}).WithBindMounts(
		bindMounts,
	).WithVolumeMounts(map[string]string{
		enclaveId: enclaveDataVolMountpointOnReplContainer,
	}).Build()
	replContainerId, _, err := dockerManager.CreateAndStartContainer(context.Background(), createAndStartArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred starting the REPL container")
	}
	defer func() {
		// Safeguard to ensure we don't leak a container
		if err := dockerManager.KillContainer(context.Background(), replContainerId); err != nil {
			logrus.Errorf("An error occurred killing the REPL container:")
			fmt.Fprintln(logrus.StandardLogger().Out, err)
		}
	}()

	finishRegistrationArgs := &kurtosis_core_rpc_api_bindings.FinishExternalContainerRegistrationArgs{
		RegistrationKey: replContainerRegistrationKey,
		ContainerId:     replContainerId,
	}
	if _, err := apiContainerClient.FinishExternalContainerRegistration(context.Background(), finishRegistrationArgs); err != nil {
		return stacktrace.Propagate(err, "An error occurred finishing the registration of the interactive REPL container")
	}

	hijackedResponse, err := dockerManager.AttachToContainer(context.Background(), replContainerId)
	if err != nil {
		return stacktrace.Propagate(err, "Couldn't attack to the REPL container")
	}
	defer hijackedResponse.Close()

	// From this point on down, I don't know why it works.... but it does
	// I just followed the solution here: https://stackoverflow.com/questions/58732588/accept-user-input-os-stdin-to-container-using-golang-docker-sdk-interactive-co
	go io.Copy(os.Stderr, hijackedResponse.Reader)
	go io.Copy(os.Stdout, hijackedResponse.Reader)
	go io.Copy(hijackedResponse.Conn, os.Stdin)

	stdinFd := int(os.Stdin.Fd())
	var oldState *terminal.State
	if terminal.IsTerminal(stdinFd) {
		oldState, err = terminal.MakeRaw(stdinFd)
		if err != nil {
			// print error
			return stacktrace.Propagate(err, "An error occurred making STDIN stream raw")
		}
		defer terminal.Restore(stdinFd, oldState)
	}

	exitCode, err := dockerManager.WaitForExit(context.Background(), replContainerId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred waiting for the REPL container to exit")
	}
	if exitCode != replContainerSuccessExitCode {
		logrus.Warnf("The REPL container exited with a non-%v exit code '%v'", replContainerSuccessExitCode, exitCode)
	}

	terminal.Restore(stdinFd, oldState)

	return nil
}
