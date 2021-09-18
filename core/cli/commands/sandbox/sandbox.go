/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package sandbox

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-client/golang/kurtosis_core_rpc_api_consts"
	"github.com/kurtosis-tech/kurtosis/commons/enclave_manager"
	"github.com/kurtosis-tech/kurtosis/commons/enclave_manager/enclave_context"
	"github.com/kurtosis-tech/kurtosis/commons/logrus_log_levels"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/sys/unix"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"
)

const (
	enclaveDataVolMountpointOnReplContainer = "/kurtosis-enclave-data"

	// TODO These defaults aren't great - it will just start a Kurtosis interactive with the latest
	//  of both images, which may or may not be compatible - what we really need is a system that
	//  detects what version of the API container/REPL to start based off the Kurt Core API version
	// TODO It's also not great that these are hardcoded - they should be hooked into the build system,
	//  to guarantee that they're compatible with each other
	defaultApiContainerImage = "kurtosistech/kurtosis-core_api"
	defaultJavascriptReplImage = "kurtosistech/javascript-interactive-repl"

	shouldPublishPorts = true

	kurtosisInteractiveIdentifier = "KTI"
	// TODO centralize this between the Bash wrapper script and this!!
	// YYYY-MM-DDTHH.MM.SS
	enclaveIdTimestampFormat = "2006-01-02T15.04.05"

	isPartitioningEnabled = true

	apiContainerImageArg = "kurtosis-api-image"
	javascriptReplImageArg = "javascript-repl-image"
	kurtosisLogLevelArg = "kurtosis-log-level"

	replContainerSuccessExitCode = 0
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
		defaultApiContainerImage,
		"The image of the Kurtosis API container to use inside the enclave",
	)

	SandboxCmd.Flags().StringVarP(
		&jsReplImage,
		javascriptReplImageArg,
		"r",
		defaultJavascriptReplImage,
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

	pullImageBestEffort(dockerManager, apiContainerImage)
	pullImageBestEffort(dockerManager, jsReplImage)

	enclaveId := getEnclaveId()

	enclaveManager := enclave_manager.NewEnclaveManager(dockerClient, apiContainerImage)

	enclaveCtx, err := enclaveManager.CreateEnclave(
		context.Background(),
		logrus.StandardLogger(),
		kurtosisLogLevel,
		map[string]bool{},
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
	javascriptReplImage string) error {
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
		"",
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
			"KURTOSIS_API_SOCKET":            kurtosisApiContainerSocket,
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
	if exitCode != replContainerSuccessExitCode {
		logrus.Warnf("The REPL container exited with a non-%v exit code '%v'", replContainerSuccessExitCode, exitCode)
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

func pullImageBestEffort(dockerManager *docker_manager.DockerManager, image string) {
	if err := dockerManager.PullImage(context.Background(), image); err != nil {
		logrus.Warnf("Failed to pull the latest version of image '%v'; you may be running an out-of-date version", image)
	}
}
