/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis-client/golang/kurtosis_core_rpc_api_consts"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/kurtosis-tech/kurtosis/commons/enclave_manager"
	"github.com/kurtosis-tech/kurtosis/commons/enclave_manager/enclave_context"
	"github.com/kurtosis-tech/kurtosis/initializer/api_container_launcher"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/sys/unix"
	"io"
	"math/rand"
	"os"
	"time"
)

const (
	successExitCode = 0
	errorExitCode = 1

	enclaveDataVolMountpointOnReplContainer = "/kurtosis-enclave-data"
)

const (
	// TODO Read this from either:
	//  1) the Kurt Core version if we're inside a testsuite repo
	//  2) a global Kurtosis config if not
	apiContainerImage = "kurtosistech/kurtosis-core_api:mieubrisse_enclave-creation-cli"

	// TODO make this configurable somehow
	kurtosisLogLevel = logrus.DebugLevel

	// TODO make configurable
	javascriptReplImage = "test-repl-image"

	shouldPublishPorts = true

	kurtosisInteractiveIdentifier = "KTI"
	// TODO centralize this between the Bash wrapper script and this!!
	// YYYY-MM-DDTHH.MM.SS
	enclaveIdTimestampFormat = "2006-01-02T15.04.05"

	isPartitioningEnabled = true
)

func main() {
	// NOTE: we'll want to change the ForceColors to false if we ever want structured logging
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	// TODO figure out a way to set the loglevel for Kurtosis from here

	// TODO set log level???

	if err := runMain(); err != nil {
		fmt.Fprintln(logrus.StandardLogger().Out, err)
		os.Exit(errorExitCode)
	}
	os.Exit(successExitCode)
}

func runMain() error {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker client")
	}
	dockerManager := docker_manager.NewDockerManager(
		logrus.StandardLogger(),
		dockerClient,
	)

	enclaveId := getEnclaveId()

	apiContainerLauncher := api_container_launcher.NewApiContainerLauncher(
		apiContainerImage,
		kurtosisLogLevel,
		shouldPublishPorts,
	)

	enclaveManager := enclave_manager.NewEnclaveManager(dockerClient, apiContainerLauncher)

	enclaveCtx, err := enclaveManager.CreateEnclave(
		context.Background(),
		logrus.StandardLogger(),
		map[string]bool{},
		enclaveId,
		isPartitioningEnabled,
	)
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

	logrus.Info("Running REPL...")
	if err := runReplContainer(dockerManager, enclaveCtx); err != nil {
		return stacktrace.Propagate(err, "An error occurred running the REPL container")
	}
	logrus.Info("REPL exited")

	return nil
}

func runReplContainer(dockerManager *docker_manager.DockerManager, enclaveCtx *enclave_context.EnclaveContext) error {
	enclaveId := enclaveCtx.GetEnclaveID()
	networkId := enclaveCtx.GetNetworkID()
	kurtosisApiContainerIpAddr := enclaveCtx.GetAPIContainerIPAddr()
	replContainerIpAddr := enclaveCtx.GetREPLContainerIPAddr()

	stdoutFd := int(os.Stdout.Fd())
	windowSize, err := unix.IoctlGetWinsize(stdoutFd, unix.TIOCGWINSZ)
	if err != nil {
		return stacktrace.NewError("An error occurred getting the current terminal window size")
	}
	interactiveModeTtySize := &docker_manager.InteractiveModeTtySize{
		Height: uint(windowSize.Row),
		Width:  uint(windowSize.Col),
	}

	kurtosisApiContainerSocket := fmt.Sprintf("%v:%v", kurtosisApiContainerIpAddr, kurtosis_core_rpc_api_consts.ListenPort)
	replContainerId, _, err := dockerManager.CreateAndStartContainer(
		context.Background(),
		javascriptReplImage,
		enclaveId + "_INTERACTIVE",
		interactiveModeTtySize,  // REPL container needs to run in interactive mode
		networkId,
		replContainerIpAddr,
		map[docker_manager.ContainerCapability]bool{},
		docker_manager.DefaultNetworkMode,
		map[nat.Port]bool{},
		false,	// REPL container doesn't have any ports for publishing
		nil,
		nil,
		map[string]string{
			// TODO Extract to named constant
			"KURTOSIS_API_SOCKET": kurtosisApiContainerSocket,
			"ENCLAVE_DATA_VOLUME_MOUNTPOINT": enclaveDataVolMountpointOnReplContainer,
		},
		map[string]string{},	// TODO bind-mount a local directory so the user can give files to the REPL
		map[string]string{
			enclaveId: enclaveDataVolMountpointOnReplContainer,
		},
		false,	// The REPL doesn't need access to the host machine
	)
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
	if exitCode != successExitCode {
		logrus.Warnf("The REPL container exited with a non-%v exit code", exitCode)
	}

	terminal.Restore(stdinFd, oldState)

	return nil
}

// TODO Merge this with the Bash enclave ID generation so that it's standardized!!!!!
func getEnclaveId() string {
	rand.Seed(time.Now().UnixNano())
	// We make this uint16 to approximate Bash's RANDOM
	randomNumUint16Bytes := make([]byte, 2)
	rand.Read(randomNumUint16Bytes)
	randomNumUint16 := binary.BigEndian.Uint16(randomNumUint16Bytes)
	return fmt.Sprintf(
		"%v%v-%v",
		kurtosisInteractiveIdentifier,
		time.Now().Format(enclaveIdTimestampFormat),
		randomNumUint16,
	)
}
