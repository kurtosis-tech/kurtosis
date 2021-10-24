package repl_runner

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-client/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-core/commons/current_time_str_provider"
	"github.com/kurtosis-tech/kurtosis-core/commons/object_labels_providers"
	"github.com/kurtosis-tech/kurtosis-core/commons/object_name_providers"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/sys/unix"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"io"
	"net"
	"os"
)

const (
)

const (
	enclaveDataVolMountpointOnReplContainer = "/kurtosis-enclave-data"

	// The dirpath inside the REPL container where the user's current directory will be bind-mounted, so the user
	//  can access files on their local system within the REPL
	workingDirectoryBindMountDirpathInsideReplContainer = "/local"

	replContainerSuccessExitCode = 0

	// WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING
	// vvvvvvvvvvvvvvv If you change these, update the REPL Dockerfile!!! vvvvvvvvvvvv
	replContainerKurtosisSocketEnvVar = "KURTOSIS_API_SOCKET"
	replContainerEnclaveDataVolMountpointEnvVar = "ENCLAVE_DATA_VOLUME_MOUNTPOINT"
	// ^^^^^^^^^^^^^^^ If you change these, update the REPL Dockerfile!!! ^^^^^^^^^^^^
	// WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING
)

// Launches a REPL container and attaches to it, blocking until the REPL container exits
func RunREPL(
	enclaveId string,
	networkId string,
	apiContainerIpInsideEnclave string,
	apiContainerPortInsideEnclave uint32,
	apiContainerIpOnHostMachine string,
	apiContainerPortOnHostMachine uint32,
	javascriptReplImage string,
	dockerManager *docker_manager.DockerManager,
) error {
	apiContainerUrlOnHostMachine := fmt.Sprintf(
		"%v:%v",
		apiContainerIpOnHostMachine,
		apiContainerPortOnHostMachine,
	)

	enclaveObjNameProvider := object_name_providers.NewEnclaveObjectNameProvider(enclaveId)
	enclaveObjLabelsProvider := object_labels_providers.NewEnclaveObjectLabelsProvider(enclaveId)

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
	postContainerStartLogMsg := "REPL container started"
	hostMachineWorkingDirpath, err := os.Getwd()
	if err != nil {
		logrus.Warn("Couldn't get the current working directory; local files will not be available inside the REPL")
		bindMounts = map[string]string{}
	} else {
		bindMounts = map[string]string{
			hostMachineWorkingDirpath: workingDirectoryBindMountDirpathInsideReplContainer,
		}
		postContainerStartLogMsg = postContainerStartLogMsg + fmt.Sprintf(
			", and the files in your current directory (%v) are available at path '%v' within the REPL",
			hostMachineWorkingDirpath,
			workingDirectoryBindMountDirpathInsideReplContainer,
		)
	}

	interactiveReplGuid := current_time_str_provider.GetCurrentTimeStr()


	kurtosisApiContainerSocket := fmt.Sprintf("%v:%v", apiContainerIpInsideEnclave, apiContainerPortInsideEnclave)
	containerName := enclaveObjNameProvider.ForInteractiveREPLContainer(interactiveReplGuid)
	labels := enclaveObjLabelsProvider.ForInteractiveREPLContainer(interactiveReplGuid)
	// TODO Replace all this with a call to the engine server!!!
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
	}).WithLabels(
		labels,
	).Build()
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
	logrus.Info(postContainerStartLogMsg)

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
