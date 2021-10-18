package inspect

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/table_printer"
	"github.com/kurtosis-tech/kurtosis-core/commons/enclave_object_labels"
	"github.com/palantir/stacktrace"
)

const (
	interactiveReplGUIDColHeader = "GUID"
)

func printInteractiveRepls(ctx context.Context, dockerManager *docker_manager.DockerManager, enclaveId string) error {
	interactiveReplLabels := getLabelsForListInteractiveRepls(enclaveId)

	containers, err := dockerManager.GetContainersByLabels(ctx, interactiveReplLabels, true)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting interactive REPL containers by labels: '%+v'", interactiveReplLabels)
	}

	tablePrinter := table_printer.NewTablePrinter(interactiveReplGUIDColHeader)

	sortedContainers, err := getContainersSortedByGUID(containers)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred sorting interactive REPL containers by GUID")
	}
	for _, container := range sortedContainers {
		containerGuid, found := container.GetLabels()[enclave_object_labels.GUIDLabel]
		if !found {
			return stacktrace.NewError("No '%v' container label was found in container ID '%v' with labels '%+v'", enclave_object_labels.GUIDLabel, container.GetId(), container.GetLabels())
		}
		if err := tablePrinter.AddRow(containerGuid); err != nil {
			return stacktrace.Propagate(err, "An error occurred writing interactive REPL row for container '%v' to the table printer", containerGuid)
		}
	}
	tablePrinter.Print()

	return nil
}

func getLabelsForListInteractiveRepls(enclaveId string) map[string]string {
	labels := map[string]string{}
	labels[enclave_object_labels.ContainerTypeLabel] = enclave_object_labels.ContainerTypeInteractiveREPL
	labels[enclave_object_labels.EnclaveIDContainerLabel] = enclaveId
	return labels
}
