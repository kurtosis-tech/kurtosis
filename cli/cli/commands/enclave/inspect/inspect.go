/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package inspect

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/kurtosis_command"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/kurtosis_command/args"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/kurtosis_command/flags"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/prebuilt_command_components/enclave_id_arg"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
	"sort"
	"strings"
	"unicode/utf8"
)

const (
	enclaveIdArgKey = "enclave-id"

	isEnclaveIdArgOptional = false
	isEnclaveIdArgGreedy  = false

	// To avoid duplicating work, we'll instantiate the docker manager & engine client in the command's pre-validation-and-run function,
	//  and then pass it to both validation & run functions
	// This is the key where the engine client will be stored in the context
	dockerManagerCtxKey = "docker-manager"
	engineClientCtxKey = "engine-client"
	engineClientCloseFuncCtxKey = "engine-client-close-func"

	enclaveIdTitleName          = "Enclave ID"
	enclaveDataDirpathTitleName = "Data Directory"
	enclaveStatusTitleName      = "Enclave Status"
	apiContainerStatusTitleName = "API Container Status"
	apiContainerHostPortTitle   = "API Container Host Port"

	headerWidthChars = 100
	headerPadChar    = "="

	shouldExamineStoppedContainersWhenPrintingEnclaveStatus = true
)

var enclaveObjectPrintingFuncs = map[string]func(ctx context.Context, dockerManager *docker_manager.DockerManager, enclaveId string) error{
	"Interactive REPLs": printInteractiveRepls,
	"User Services":     printUserServices,
	"Kurtosis Modules":  printModules,
}

var EnclaveInspectCmd = &kurtosis_command.KurtosisCommand{
	CommandStr:       command_str_consts.EnclaveInspectCmdStr,
	ShortDescription: "Lists detailed information about an enclave",
	Args:             []*args.ArgConfig{
		enclave_id_arg.NewEnclaveIDArg(
			enclaveIdArgKey,
			engineClientCtxKey,
			isEnclaveIdArgOptional,
			isEnclaveIdArgGreedy,
		),
	},
	PreValidationAndRunFunc: setup,
	RunFunc:          run,
	PostValidationAndRunFunc: teardown,
}

func setup(ctx context.Context) (context.Context, error) {
	result := ctx

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the Docker client")
	}
	dockerManager := docker_manager.NewDockerManager(
		logrus.StandardLogger(),
		dockerClient,
	)
	result = context.WithValue(result, dockerManagerCtxKey, dockerManager)

	engineManager := engine_manager.NewEngineManager(dockerManager)
	objAttrsProvider := schema.GetObjectAttributesProvider()
	engineClient, closeClientFunc, err := engineManager.StartEngineIdempotentlyWithDefaultVersion(ctx, objAttrsProvider, defaults.DefaultEngineLogLevel)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a new Kurtosis engine client")
	}
	result = context.WithValue(result, engineClientCtxKey, engineClient)
	result = context.WithValue(result, engineClientCloseFuncCtxKey, closeClientFunc())

	return result, nil
}

func run(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {
	uncastedEngineClient := ctx.Value(engineClientCtxKey)
	if uncastedEngineClient == nil {
		return stacktrace.NewError("Expected an engine client to have been stored in the context under key '%v', but none was found; this is a bug in Kurtosis!", engineClientCtxKey)
	}
	engineClient, ok := uncastedEngineClient.(kurtosis_engine_rpc_api_bindings.EngineServiceClient)
	if !ok {
		return stacktrace.NewError("Found an object that should be the engine client stored in the context under key '%v', but this object wasn't of the correct type", engineClientCtxKey)
	}

	// TODO GET RID OF THIS!!! Everything should be doable through the engine client
	uncastedDockerManager := ctx.Value(dockerManagerCtxKey)
	if uncastedDockerManager == nil {
		return stacktrace.NewError("Expected a Docker manager to have been stored in the context under key '%v', but none was found; this is a bug in Kurtosis!", dockerManagerCtxKey)
	}
	dockerManager, ok := uncastedDockerManager.(*docker_manager.DockerManager)
	if !ok {
		return stacktrace.NewError("Found an object that should be the Docker manager stored in the context under key '%v', but this object wasn't of the correct type", dockerManagerCtxKey)
	}

	enclaveId, err := args.GetNonGreedyArg(enclaveIdArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for arg '%v' but didn't find one", enclaveIdArgKey)
	}


	getEnclavesResp, err := engineClient.GetEnclaves(ctx, &emptypb.Empty{})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting enclaves, which is necessary to display the state for enclave '%v'", enclaveId)
	}

	enclaveInfo, found := getEnclavesResp.EnclaveInfo[enclaveId]
	if !found {
		return stacktrace.NewError("No enclave with ID '%v' exists", enclaveId)
	}

	enclaveDataDirpath := enclaveInfo.GetEnclaveDataDirpathOnHostMachine()
	enclaveContainersStatus := enclaveInfo.ContainersStatus
	enclaveApiContainerStatus := enclaveInfo.ApiContainerStatus

	keyValuePrinter := output_printers.NewKeyValuePrinter()
	keyValuePrinter.AddPair(enclaveIdTitleName, enclaveId)
	keyValuePrinter.AddPair(enclaveDataDirpathTitleName, enclaveDataDirpath)
	// TODO Refactor these to use a user-friendly string and not the enum name
	keyValuePrinter.AddPair(enclaveStatusTitleName, enclaveContainersStatus.String())
	keyValuePrinter.AddPair(apiContainerStatusTitleName, enclaveApiContainerStatus.String())
	if enclaveApiContainerStatus == kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus_EnclaveAPIContainerStatus_RUNNING {
		apiContainerHostInfo := enclaveInfo.GetApiContainerHostMachineInfo()
		apiContainerHostPortInfoStr := fmt.Sprintf(
			"%v:%v",
			apiContainerHostInfo.GetIpOnHostMachine(),
			apiContainerHostInfo.GetPortOnHostMachine(),
		)
		keyValuePrinter.AddPair(apiContainerHostPortTitle, apiContainerHostPortInfoStr)
	}
	keyValuePrinter.Print()
	fmt.Fprintln(logrus.StandardLogger().Out, "")

	sortedEnclaveObjHeaders := []string{}
	for header := range enclaveObjectPrintingFuncs {
		sortedEnclaveObjHeaders = append(sortedEnclaveObjHeaders, header)
	}
	sort.Strings(sortedEnclaveObjHeaders)

	headersWithPrintErrs := []string{}
	for _, header := range sortedEnclaveObjHeaders {
		printingFunc, found := enclaveObjectPrintingFuncs[header]
		if !found {
			return stacktrace.NewError("No printing function found for enclave object '%v'; this is a bug in Kurtosis!", header)
		}

		numRunesInHeader := utf8.RuneCountInString(header) + 2 // 2 because there will be a space before and after the header
		numPadChars := (headerWidthChars - numRunesInHeader) / 2
		padStr := strings.Repeat(headerPadChar, numPadChars)
		fmt.Println(fmt.Sprintf("%v %v %v", padStr, header, padStr))

		if err := printingFunc(ctx, dockerManager, enclaveId); err != nil {
			logrus.Error(err)
			headersWithPrintErrs = append(headersWithPrintErrs, header)
		}
		fmt.Println("")
	}

	if len(headersWithPrintErrs) > 0 {
		return stacktrace.NewError(
			"Errors occurred printing the following enclave elements: %v",
			strings.Join(headersWithPrintErrs, ", "),
		)
	}

	return nil
}

func teardown(ctx context.Context) {
	uncastedEngineClientCloseFunc := ctx.Value(engineClientCloseFuncCtxKey)
	if uncastedEngineClientCloseFunc != nil {
		engineClientCloseFunc, ok := uncastedEngineClientCloseFunc.(func() error)
		if ok {
			if err := engineClientCloseFunc(); err != nil {
				logrus.Warnf("We tried to close the engine client after we're done using it, but doing so threw an error:\n%v", err)
			}
		} else {
			logrus.Errorf("Expected the object at context key '%v' to be an engine client close function, but it wasn't; this is a bug in Kurtosis!", engineClientCloseFuncCtxKey),
		}
	} else {
		logrus.Errorf(
			"Expected to find an engine client close function during teardown at context key '%v', but none was found; this is a bug in Kurtosis!",
			engineClientCloseFuncCtxKey,
		)
	}
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================
func sortContainersByGUID(containers []*types.Container) ([]*types.Container, error) {
	containersSet := map[string]*types.Container{}
	for _, container := range containers {
		if container != nil {
			containerGUID, found := container.GetLabels()[schema.GUIDLabel]
			if !found {
				return nil, stacktrace.NewError("No '%v' container label was found in container ID '%v' with labels '%+v'", schema.GUIDLabel, container.GetId(), container.GetLabels())
			}
			containersSet[containerGUID] = container
		}
	}

	containersResult := make([]*types.Container, 0, len(containersSet))
	for _, container := range containersSet {
		containersResult = append(containersResult, container)
	}

	sort.Slice(containersResult, func(i, j int) bool {
		return containersResult[i].GetLabels()[schema.GUIDLabel] < containersResult[j].GetLabels()[schema.GUIDLabel]
	})

	return containersResult, nil
}
