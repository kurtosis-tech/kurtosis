package user_service_functions

import (
	"bufio"
	"context"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	terminal "golang.org/x/term"
	"io"
	"os"
)

// We'll try to use the nicer-to-use shells first before we drop down to the lower shells
var commandToRunWhenCreatingUserServiceShell = []string{
	"sh",
	"-c",
	`if command -v 'bash' > /dev/null; then
		echo "Found bash on container; creating bash shell..."; bash; 
       else 
		echo "No bash found on container; dropping down to sh shell..."; sh; 
	fi`,
}

func GetShellOnUserService(ctx context.Context, enclaveId enclave.EnclaveUUID, serviceUuid service.ServiceUUID, dockerManager *docker_manager.DockerManager) error {
	_, serviceDockerResources, err := getSingleUserServiceObjAndResourcesNoMutex(ctx, enclaveId, serviceUuid, dockerManager)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting service object and Docker resources for service '%v' in enclave '%v'", serviceUuid, enclaveId)
	}
	container := serviceDockerResources.ServiceContainer

	hijackedResponse, err := dockerManager.CreateContainerExec(ctx, container.GetId(), commandToRunWhenCreatingUserServiceShell)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting a shell on user service with UUID '%v' in enclave '%v'", serviceUuid, enclaveId)
	}

	newConnection := hijackedResponse.Conn

	newReader := bufio.NewReader(newConnection)

	// From this point on down, I don't know why it works.... but it does
	// I just followed the solution here: https://stackoverflow.com/questions/58732588/accept-user-input-os-stdin-to-container-using-golang-docker-sdk-interactive-co
	// This channel is being used to know the user exited the ContainerExec
	finishChan := make(chan bool)
	// TODO(victor.colombo): Decide what to do with errors that happen inside these go routines
	go func() {
		//nolint:errcheck
		io.Copy(os.Stdout, newReader)
		finishChan <- true
	}()

	//nolint:errcheck
	go io.Copy(os.Stderr, newReader)
	//nolint:errcheck
	go io.Copy(newConnection, os.Stdin)

	stdinFd := int(os.Stdin.Fd())
	var oldState *terminal.State
	if terminal.IsTerminal(stdinFd) {
		oldState, err = terminal.MakeRaw(stdinFd)
		if err != nil {
			// print error
			return stacktrace.Propagate(err, "An error occurred making STDIN stream raw")
		}
		//nolint:errcheck
		defer terminal.Restore(stdinFd, oldState)
	}

	<-finishChan

	return nil
}
