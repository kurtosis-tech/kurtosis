/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package shell

import (
	"bufio"
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/backend_creator"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/logrus_log_levels"
	"github.com/kurtosis-tech/kurtosis-cli/commons/positional_arg_parser"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"os"
	"strings"
)

const (
	kurtosisLogLevelArg                    = "kurtosis-log-level"
	enclaveIDArg                           = "enclave-id"
	guidArg                                = "guid"
)

var defaultKurtosisLogLevel = logrus.InfoLevel.String()
var positionalArgs = []string{
	enclaveIDArg,
	guidArg,
}

// We'll try to use the nicer-to-use shells first before we drop down to the lower shells
var commandToRun = []string{
	"sh",
	"-c",
	"if command -v 'bash' > /dev/null; then echo \"Found bash on container; creating bash shell...\"; bash; else echo \"No bash found on container; dropping down to sh shell...\"; sh; fi",
}

var ShellCmd = &cobra.Command{
	Use:                   command_str_consts.ShellCmdStr + " [flags] " + strings.Join(positionalArgs, " "),
	DisableFlagsInUseLine: true,
	Short:                 "Start a shell on the specified service",
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

	parsedPositionalArgs, err := positional_arg_parser.ParsePositionalArgsAndRejectEmptyStrings(positionalArgs, args)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the positional args")
	}
	enclaveIdStr := parsedPositionalArgs[enclaveIDArg]
	enclaveId := enclave.EnclaveID(enclaveIdStr)
	guidStr := parsedPositionalArgs[guidArg]
	guid := service.ServiceGUID(guidStr)

	// TODO REFACTOR: we should get this backend from the config!!
	var apiContainerModeArgs *backend_creator.APIContainerModeArgs = nil  // Not an API container
	kurtosisBackend, err := backend_creator.GetLocalDockerKurtosisBackend(apiContainerModeArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting local Docker Kurtosis backend")
	}

	conn, err := kurtosisBackend.GetConnectionWithUserService(ctx, enclaveId, guid)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting connection with user service with GUID '%v' in enclave '%v'", guid, enclaveId)
	}
	defer conn.Close()

	newReader := bufio.NewReader(conn)

	// From this point on down, I don't know why it works.... but it does
	// I just followed the solution here: https://stackoverflow.com/questions/58732588/accept-user-input-os-stdin-to-container-using-golang-docker-sdk-interactive-co
	// This channel is being used to know the user exited the ContainerExec
	finishChan := make(chan bool)
	go func() {
		io.Copy(os.Stdout, newReader)
		finishChan <- true
	}()
	go io.Copy(os.Stderr, newReader)
	go io.Copy(conn, os.Stdin)

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

	_ = <-finishChan

	terminal.Restore(stdinFd, oldState)

	return nil
}
