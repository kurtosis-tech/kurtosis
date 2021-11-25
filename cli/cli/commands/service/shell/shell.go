/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package shell

import (
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	docker_manager_types "github.com/kurtosis-tech/container-engine-lib/lib/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/logrus_log_levels"
	labelsHelper "github.com/kurtosis-tech/kurtosis-cli/cli/helpers/service_containers_labels_by_enclaveID"
	"github.com/kurtosis-tech/kurtosis-cli/commons/positional_arg_parser"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"os"
	"strings"
)

const (
	kurtosisLogLevelArg = "kurtosis-log-level"
	enclaveIDArg        = "enclave-id"
	serviceIDArg        = "service-id"
)

var defaultKurtosisLogLevel = logrus.InfoLevel.String()
var positionalArgs = []string{
	enclaveIDArg,
	serviceIDArg,
}

var ShellCmd = &cobra.Command{
	Use:                   command_str_consts.ShellCmdStr + " [flags] " + strings.Join(positionalArgs, " "),
	DisableFlagsInUseLine: true,
	Short:                 "Get access to the service shell",
	RunE:                  run,
}

var kurtosisLogLevelStr string

func init() {
	ShellCmd.Flags().StringVarP(
		&kurtosisLogLevelStr,
		kurtosisLogLevelArg,
		"l",
		defaultKurtosisLogLevel,
		fmt.Sprintf(
			"The log level that Kurtosis itself should log at (%v)",
			strings.Join(logrus_log_levels.GetAcceptableLogLevelStrs(), "|"),
		),
	)

}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	kurtosisLogLevel, err := logrus.ParseLevel(kurtosisLogLevelStr)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing Kurtosis loglevel string '%v' to a log level object", kurtosisLogLevelStr)
	}
	logrus.SetLevel(kurtosisLogLevel)

	parsedPositionalArgs, err := positional_arg_parser.ParsePositionalArgsAndRejectEmptyStrings(positionalArgs, args)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the positional args")
	}
	enclaveID := parsedPositionalArgs[enclaveIDArg]
	serviceID := parsedPositionalArgs[serviceIDArg]

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker client")
	}
	dockerManager := docker_manager.NewDockerManager(
		logrus.StandardLogger(),
		dockerClient,
	)

	labels := labelsHelper.GetUserServiceContainerLabelsWithEnclaveId(enclaveID)

	containers, err := dockerManager.GetContainersByLabels(ctx, labels, true)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting containers by labels: '%+v'", labels)
	}

	if containers == nil || len(containers) == 0 {
		logrus.Errorf("There is not any service container with enclave ID '%v'", enclaveID)
		return nil
	}

	var containersWithSearchedServiceID = []*docker_manager_types.Container{}
	for _, container := range containers {
		labelsMap := container.GetLabels()
		containerID, found := labelsMap[schema.GUIDLabel]
		if found && containerID == serviceID {
			containersWithSearchedServiceID = append(containersWithSearchedServiceID, container)
		}
	}

	if len(containersWithSearchedServiceID) == 0 {
		logrus.Errorf("There is not any service container with GUID '%v'", serviceID)
		return nil
	}

	if len(containersWithSearchedServiceID) > 1 {
		return stacktrace.NewError("Only one container with enclave-id '%v' and GUID '%v' should exist but there are '%v' containers with these properties", enclaveID, serviceID, len(containers))
	}

	serviceContainer := containersWithSearchedServiceID[0]

	config := types.ExecConfig{
		AttachStdin: true,
		Tty:          true,
		AttachStderr: true,
		AttachStdout: true,
		Detach:       false,
		Cmd:          []string{"sh"},
	}

	response, err := dockerClient.ContainerExecCreate(ctx, serviceContainer.GetId(), config)

	if err != nil {
		return stacktrace.Propagate(err, "an error occurred while creating the ContainerExec")
	}

	execID := response.ID
	if execID == "" {
		return errors.New("the Exec ID was empty")
	}

	es := types.ExecStartCheck{
		Detach: false,
		Tty:    true,
	}

	hijackedResponse, err := dockerClient.ContainerExecAttach(ctx, execID, es)
	if err != nil {
		return stacktrace.Propagate(err, "there was an error while attaching to the ContainerExec")
	}
	defer hijackedResponse.Close()

	// From this point on down, I don't know why it works.... but it does
	// I just followed the solution here: https://stackoverflow.com/questions/58732588/accept-user-input-os-stdin-to-container-using-golang-docker-sdk-interactive-co
	// This channel is being used to know the user exited the ContainerExec
	finishChan := make(chan bool)
	go func() {
		io.Copy(os.Stdout, hijackedResponse.Reader)
		finishChan <- true
	}()
	go io.Copy(os.Stderr, hijackedResponse.Reader)
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

	_ =  <- finishChan

	terminal.Restore(stdinFd, oldState)

	return nil
}


