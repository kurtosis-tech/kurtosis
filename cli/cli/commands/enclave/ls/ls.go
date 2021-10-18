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
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis-cli/cli/enclave_manager/enclave_statuses"
	"github.com/kurtosis-tech/kurtosis-cli/cli/enclave_status_from_container_status_retriever"
	"github.com/kurtosis-tech/kurtosis-cli/cli/logrus_log_levels"
	"github.com/kurtosis-tech/kurtosis-cli/cli/output_printers"
	"github.com/kurtosis-tech/kurtosis-core/commons/enclave_object_labels"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker client")
	}
	dockerManager := docker_manager.NewDockerManager(
		logrus.StandardLogger(),
		dockerClient,
	)

	searchLabels := map[string]string{
		enclave_object_labels.AppIDLabel: enclave_object_labels.AppIDValue,
	}
	kurtosisContainers, err := dockerManager.GetContainersByLabels(ctx, searchLabels, shouldExamineStoppedContainersForListingEnclaves)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting Kurtosis containers by labels: '%+v'", searchLabels)
	}

	kurtosisContainersByEnclaveId := map[string][]*types.Container{}
	for _, container := range kurtosisContainers {
		containerLabels := container.GetLabels()
		enclaveId, found := containerLabels[enclave_object_labels.EnclaveIDContainerLabel]
		if !found {
			return stacktrace.NewError(
				"No '%v' label found for container '%v'; this is a bug in Kurtosis!",
				enclave_object_labels.EnclaveIDContainerLabel,
				container.GetId(),
			)
		}

		enclaveContainers, found := kurtosisContainersByEnclaveId[enclaveId]
		if !found {
			enclaveContainers = []*types.Container{}
		}
		enclaveContainers = append(enclaveContainers, container)

		kurtosisContainersByEnclaveId[enclaveId] = enclaveContainers
	}

	orderedEnclaveIds := []string{}
	enclaveStatuses := map[string]enclave_statuses.EnclaveStatus{}
	for enclaveId, enclaveContainers := range kurtosisContainersByEnclaveId {
		orderedEnclaveIds = append(orderedEnclaveIds, enclaveId)
		enclaveStatus, err := enclave_status_from_container_status_retriever.GetEnclaveStatus(enclaveContainers)
		if err != nil {
			return stacktrace.NewError(
				"An error occurred getting the status for enclave '%v' from the status of its containers",
				enclaveId,
			)
		}
		enclaveStatuses[enclaveId] = enclaveStatus
	}
	sort.Strings(orderedEnclaveIds)

	tablePrinter := output_printers.NewTablePrinter(enclaveIdColumnHeader, enclaveStatusColumnHeader)
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
