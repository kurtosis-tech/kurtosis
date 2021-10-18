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
	"github.com/kurtosis-tech/kurtosis-cli/cli/enclave_manager/enclave_states"
	"github.com/kurtosis-tech/kurtosis-cli/cli/enclave_state_from_container_state_retriever"
	"github.com/kurtosis-tech/kurtosis-cli/cli/logrus_log_levels"
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
	kurtosisContainers, err := dockerManager.GetContainersByLabels(ctx, searchLabels, true)
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

	orderedEnclaveIds := make([]string, len(kurtosisContainersByEnclaveId))
	enclaveStates := map[string]enclave_states.EnclaveState{}
	for enclaveId, enclaveContainers := range kurtosisContainersByEnclaveId {
		orderedEnclaveIds = append(orderedEnclaveIds, enclaveId)
		enclaveState, err := enclave_state_from_container_state_retriever.GetEnclaveState(enclaveContainers)
		if err != nil {
			return stacktrace.NewError("An error occurred getting the state for enclave '%v' from the status of its containers")
		}
		enclaveStates[enclaveId] = enclaveState
	}
	sort.Strings(orderedEnclaveIds)



	fmt.Fprintln(logrus.StandardLogger().Out, enclaveIdColumnHeader)
	for _, enclaveId := range orderedEnclaveIds {

	}

	return nil
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================
func getLabelsForListEnclaves() map[string]string {
	labels := map[string]string{}
	labels[enclave_object_labels.ContainerTypeLabel] = enclave_object_labels.ContainerTypeAPIContainer
	return labels
}

func getContainersEnclaveIds(containers []*types.Container) []string{
	containersSet := map[string]*types.Container{}
	for _, container := range containers {
		if container != nil {
			containerId := container.GetId()
			containersSet[containerId] = container
		}
	}

	enclaveIds := []string{}
	for _, container := range containersSet {
		enclaveId := container.GetLabels()[enclave_object_labels.EnclaveIDContainerLabel]
		enclaveIds = append(enclaveIds, enclaveId)
	}

	sort.Strings(enclaveIds)

	return enclaveIds
}
