/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package shell

import (
	"bufio"
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/highlevel/enclave_id_arg"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-sdk/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/stacktrace"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"os"
)

const (
	enclaveIdArgKey   = "enclave-id"
	isEnclaveIdArgOptional = false
	isEnclaveIdArgGreedy = false

	serviceGuidArgKey = "service-guid"

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey  = "engine-client"
)

var ServiceShellCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.ServiceShellCmdStr,
	ShortDescription:          "Gets a shell on a service",
	LongDescription:           "Starts a shell on the specified service",
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	Args: []*args.ArgConfig{
		enclave_id_arg.NewEnclaveIDArg(
			enclaveIdArgKey,
			engineClientCtxKey,
			isEnclaveIdArgOptional,
			isEnclaveIdArgGreedy,
		),
		// TODO Create a NewServiceIDArg that adds autocomplete
		{
			Key:             serviceGuidArgKey,
		},
	},
	RunFunc:                   run,
}

func run(
	ctx context.Context,
	kurtosisBackend backend_interface.KurtosisBackend,
	_ kurtosis_engine_rpc_api_bindings.EngineServiceClient,
	flags *flags.ParsedFlags,
	args *args.ParsedArgs,
) error {
	enclaveIdStr, err := args.GetNonGreedyArg(enclaveIdArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave ID using arg key '%v'", enclaveIdArgKey)
	}
	enclaveId := enclave.EnclaveID(enclaveIdStr)

	serviceGuidStr, err := args.GetNonGreedyArg(serviceGuidArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the service GUID using arg key '%v'", serviceGuidArgKey)
	}
	serviceGuid := service.ServiceGUID(serviceGuidStr)

	conn, err := kurtosisBackend.GetConnectionWithUserService(ctx, enclaveId, serviceGuid)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting connection with user service with GUID '%v' in enclave '%v'", serviceGuid, enclaveId)
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
