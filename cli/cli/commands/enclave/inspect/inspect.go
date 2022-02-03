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
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_wrappers"
	"github.com/kurtosis-tech/kurtosis-cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/lib/kurtosis_context"
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

var EnclaveInspectCmd = &command_wrappers.KurtosisCommand{
	CommandStr:       command_str_consts.EnclaveInspectCmdStr,
	ShortDescription: "Lists detailed information about an enclave",
	Args:             []*command_wrappers.ArgConfig{
		{
			Key:             enclaveIdArgKey,
			CompletionsFunc: getCompletions,
			ValidationFunc:  validate,
		},
	},
	RunFunc:          run,
}

/*
var InspectCmd = &cobra.Command{
	Use:                   command_str_consts.EnclaveInspectCmdStr + " [flags] " + strings.Join(positionalArgs, " "),
	DisableFlagsInUseLine: true,
	Short:                 "Lists detailed information about an enclave",
	RunE:                  run,
	ValidArgsFunction:     getValidArgs,
}

func init() {
}
 */

func getCompletions(flags *command_wrappers.ParsedFlags, previousArgs *command_wrappers.ParsedArgs) ([]string, error) {
	ctx := context.Background()

	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred connecting to the Kurtosis engine for retrieving the enclave IDs for tab completion",
		)
	}

	enclaves, err := kurtosisCtx.GetEnclaves(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting the enclaves retrieving for enclave ID tab completion",
		)
	}

	result := []string{}
	for enclaveId := range enclaves {
		result = append(result, string(enclaveId))
	}
	sort.Strings(result)

	return result, nil
}

func validate(flags *command_wrappers.ParsedFlags, args *command_wrappers.ParsedArgs) error {
	enclaveId, err := args.GetNonGreedyArg(enclaveIdArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for arg '%v' but didn't find one", enclaveIdArgKey)
	}

	ctx := context.Background()

	// TODO It seems really bad to do this twice!! We do it here, and in the 'run' method as well
	//  It makes for some really clean validation-vs-run code, but maybe that's being too ambitious
	//  and it's better to have it all rolled into one so you can reuse the engineManager that gets created here
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker client")
	}
	dockerManager := docker_manager.NewDockerManager(
		logrus.StandardLogger(),
		dockerClient,
	)

	engineManager := engine_manager.NewEngineManager(dockerManager)
	objAttrsProvider := schema.GetObjectAttributesProvider()
	engineClient, closeClientFunc, err := engineManager.StartEngineIdempotentlyWithDefaultVersion(ctx, objAttrsProvider, defaults.DefaultEngineLogLevel)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating a new Kurtosis engine client")
	}
	defer closeClientFunc()

	getEnclavesResp, err := engineClient.GetEnclaves(ctx, &emptypb.Empty{})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting enclaves, which is necessary to display the state for enclave '%v'", enclaveId)
	}

	if _, found := getEnclavesResp.EnclaveInfo[enclaveId]; !found {
		return stacktrace.Propagate(err, "No enclave found with ID '%v'", enclaveId)
	}
	return nil
}

func run(flags *command_wrappers.ParsedFlags, args *command_wrappers.ParsedArgs) error {
	enclaveId, err := args.GetNonGreedyArg(enclaveIdArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for arg '%v' but didn't find one", enclaveIdArgKey)
	}

	ctx := context.Background()

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker client")
	}
	dockerManager := docker_manager.NewDockerManager(
		logrus.StandardLogger(),
		dockerClient,
	)

	engineManager := engine_manager.NewEngineManager(dockerManager)
	objAttrsProvider := schema.GetObjectAttributesProvider()
	engineClient, closeClientFunc, err := engineManager.StartEngineIdempotentlyWithDefaultVersion(ctx, objAttrsProvider, defaults.DefaultEngineLogLevel)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating a new Kurtosis engine client")
	}
	defer closeClientFunc()

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
