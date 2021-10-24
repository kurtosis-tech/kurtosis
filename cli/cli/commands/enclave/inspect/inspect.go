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
	enclave_status_from_container_status_retriever2 "github.com/kurtosis-tech/kurtosis-cli/cli/helpers/enclave_status_from_container_status_retriever"
	enclave_statuses2 "github.com/kurtosis-tech/kurtosis-cli/cli/helpers/enclave_statuses"
	logrus_log_levels2 "github.com/kurtosis-tech/kurtosis-cli/cli/helpers/logrus_log_levels"
	output_printers2 "github.com/kurtosis-tech/kurtosis-cli/cli/helpers/output_printers"
	positional_arg_parser2 "github.com/kurtosis-tech/kurtosis-cli/cli/helpers/positional_arg_parser"
	"github.com/kurtosis-tech/kurtosis-core/commons/enclave_object_labels"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"sort"
	"strings"
	"unicode/utf8"
)

const (
	kurtosisLogLevelArg = "kurtosis-log-level"
	enclaveIdArg        = "enclave-id"

	enclaveIdTitleName     = "Enclave ID"
	enclaveStatusTitleName = "Status"

	headerWidthChars = 100
	headerPadChar = "="

	shouldExamineStoppedContainersWhenPrintingEnclaveStatus = true
)

var defaultKurtosisLogLevel = logrus.InfoLevel.String()
var positionalArgs = []string{
	enclaveIdArg,
}

var enclaveObjectPrintingFuncs = map[string]func(ctx context.Context, dockerManager *docker_manager.DockerManager, enclaveId string) error {
	"Interactive REPLs": printInteractiveRepls,
	"User Services": printUserServices,
}

var InspectCmd = &cobra.Command{
	Use:   command_str_consts.EnclaveInspectCmdStr + " [flags] " + strings.Join(positionalArgs, " "),
	DisableFlagsInUseLine: true,
	Short: "Lists detailed information about an enclave",
	RunE:  run,
}

var kurtosisLogLevelStr string

func init() {
	InspectCmd.Flags().StringVarP(
		&kurtosisLogLevelStr,
		kurtosisLogLevelArg,
		"l",
		defaultKurtosisLogLevel,
		fmt.Sprintf(
			"The log level that Kurtosis itself should log at (%v)",
			strings.Join(logrus_log_levels2.GetAcceptableLogLevelStrs(), "|"),
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

	parsedPositionalArgs, err := positional_arg_parser2.ParsePositionalArgsAndRejectEmptyStrings(positionalArgs, args)
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

	enclaveStatus, err := getEnclaveStatus(ctx, dockerManager, enclaveId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred determining the status of the enclave from its containers' statuses")
	}

	keyValuePrinter := output_printers2.NewKeyValuePrinter()
	keyValuePrinter.AddPair(enclaveIdTitleName, enclaveId)
	keyValuePrinter.AddPair(enclaveStatusTitleName, string(enclaveStatus))
	keyValuePrinter.Print()
	fmt.Fprintln(logrus.StandardLogger().Out, "")

	headersWithPrintErrs := []string{}
	for header, printingFunc := range enclaveObjectPrintingFuncs {
		numRunesInHeader := utf8.RuneCountInString(header) + 2	// 2 because there will be a space before and after the header
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
func getEnclaveStatus(ctx context.Context, dockerManager *docker_manager.DockerManager, enclaveId string) (enclave_statuses2.EnclaveStatus, error) {
	searchLabels := map[string]string{
		enclave_object_labels.EnclaveIDContainerLabel: enclaveId,
	}
	// TODO Replace with a call to the engine server!
	enclaveContainers, err := dockerManager.GetContainersByLabels(ctx, searchLabels, shouldExamineStoppedContainersWhenPrintingEnclaveStatus)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the enclave containers by labels '%+v'", searchLabels)
	}

	enclaveStatus, err := enclave_status_from_container_status_retriever2.GetEnclaveStatus(enclaveContainers)
	if err != nil {
		return "", stacktrace.Propagate(
			err,
			"An error occurred getting the status of enclave '%v' from its containers' statuses",
			enclaveId,
		)
	}
	return enclaveStatus, nil
}

func sortContainersByGUID(containers []*types.Container) ([]*types.Container, error) {
	containersSet := map[string]*types.Container{}
	for _, container := range containers {
		if container != nil {
			containerGUID, found := container.GetLabels()[enclave_object_labels.GUIDLabel]
			if !found {
				return nil, stacktrace.NewError("No '%v' container label was found in container ID '%v' with labels '%+v'", enclave_object_labels.GUIDLabel, container.GetId(), container.GetLabels())
			}
			containersSet[containerGUID] = container
		}
	}

	containersResult := make([]*types.Container, 0, len(containersSet))
	for _, container := range containersSet {
		containersResult = append(containersResult, container)
	}

	sort.Slice(containersResult, func(i, j int) bool {
		return containersResult[i].GetLabels()[enclave_object_labels.GUIDLabel] < containersResult[j].GetLabels()[enclave_object_labels.GUIDLabel]
	})

	return containersResult, nil
}
