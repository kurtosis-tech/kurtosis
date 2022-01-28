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
	"github.com/kurtosis-tech/kurtosis-cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/kurtosis-cli/commons/positional_arg_parser"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/emptypb"
	"sort"
	"strings"
	"unicode/utf8"
)

const (
	enclaveIdArg = "enclave-id"

	enclaveIdTitleName          = "Enclave ID"
	enclaveDataDirpathTitleName = "Data Directory"
	enclaveStatusTitleName      = "Enclave Status"
	apiContainerStatusTitleName = "API Container Status"
	apiContainerHostPortTitle   = "API Container Host Port"

	headerWidthChars = 100
	headerPadChar    = "="

	shouldExamineStoppedContainersWhenPrintingEnclaveStatus = true
)

var defaultKurtosisLogLevel = logrus.InfoLevel.String()
var positionalArgs = []string{
	enclaveIdArg,
}

var enclaveObjectPrintingFuncs = map[string]func(ctx context.Context, dockerManager *docker_manager.DockerManager, enclaveId string) error{
	"Interactive REPLs": printInteractiveRepls,
	"User Services":     printUserServices,
	"Kurtosis Modules":  printModules,
}

var InspectCmd = &cobra.Command{
	Use:                   command_str_consts.EnclaveInspectCmdStr + " [flags] " + strings.Join(positionalArgs, " "),
	DisableFlagsInUseLine: true,
	Short:                 "Lists detailed information about an enclave",
	RunE:                  run,
}

var kurtosisLogLevelStr string

func init() {
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
	enclaveId := parsedPositionalArgs[enclaveIdArg]

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
