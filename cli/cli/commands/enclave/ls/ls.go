/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package ls

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/engine_manager"
	logrus_log_levels2 "github.com/kurtosis-tech/kurtosis-cli/cli/helpers/logrus_log_levels"
	output_printers2 "github.com/kurtosis-tech/kurtosis-cli/cli/helpers/output_printers"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/emptypb"
	"sort"
	"strings"
)

const (
	kurtosisLogLevelArg = "kurtosis-log-level"

	enclaveIdColumnHeader = "EnclaveID"
	enclaveStatusColumnHeader = "Status"

	shouldExamineStoppedContainersForListingEnclaves = true
)

var kurtosisLogLevelStr string

var defaultKurtosisLogLevel = logrus.InfoLevel.String()

var LsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List Kurtosis enclaves",
	RunE:  run,
}

func init() {
	LsCmd.Flags().StringVarP(
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

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker client")
	}
	dockerManager := docker_manager.NewDockerManager(
		logrus.StandardLogger(),
		dockerClient,
	)

	engineClient, closeClientFunc, err := engine_manager.GetEngineClient(ctx, dockerManager)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating a new Kurtosis engine client")
	}
	defer closeClientFunc()

	response, err := engineClient.GetEnclaves(ctx, &emptypb.Empty{})
	if err != nil {
		return stacktrace.Propagate(err,"An error occurred getting enclaves")
	}
	enclaveInfoMap := response.GetEnclaveInfo()

	orderedEnclaveIds := []string{}
	enclaveStatuses := map[string]string{}
	for enclaveId, enclaveInfo := range enclaveInfoMap {
		orderedEnclaveIds = append(orderedEnclaveIds, enclaveId)
		//TODO refactor in order to print users friendly status strings and not the enum value
		enclaveStatuses[enclaveId] = enclaveInfo.GetContainersStatus().String()
	}
	sort.Strings(orderedEnclaveIds)

	tablePrinter := output_printers2.NewTablePrinter(enclaveIdColumnHeader, enclaveStatusColumnHeader)
	for _, enclaveId := range orderedEnclaveIds {
		enclaveStatus, found := enclaveStatuses[enclaveId]
		if !found {
			return stacktrace.NewError("We're about to print enclave '%v', but it doesn't have a status; this is a bug in Kurtosis!", enclaveId)
		}
		if err := tablePrinter.AddRow(enclaveId, string(enclaveStatus)); err != nil {
			return stacktrace.NewError("An error occurred adding row for enclave '%v' to the table printer", enclaveId)
		}
	}
	tablePrinter.Print()

	return nil
}
