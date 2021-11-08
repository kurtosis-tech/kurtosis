package inspect

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/kurtosis-core/commons/schema"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	interactiveReplGUIDColHeader                                    = "GUID"
	shouldShowStoppedContainersWhenGettingInteractiveREPLContainers = true
)

func printInteractiveRepls(ctx context.Context, dockerManager *docker_manager.DockerManager, enclaveId string) error {
	interactiveReplLabels := getLabelsForListInteractiveRepls(enclaveId)

	containers, err := dockerManager.GetContainersByLabels(ctx, interactiveReplLabels, shouldShowStoppedContainersWhenGettingInteractiveREPLContainers)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting interactive REPL containers by labels: '%+v'", interactiveReplLabels)
	}

	tablePrinter := output_printers.NewTablePrinter(interactiveReplGUIDColHeader)

	sortedContainers, err := sortContainersByGUID(containers)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred sorting interactive REPL containers by GUID")
	}
	for _, container := range sortedContainers {
		containerGuid, found := container.GetLabels()[schema.GUIDLabel]
		if !found {
			return stacktrace.NewError("No '%v' container label was found in container ID '%v' with labels '%+v'", schema.GUIDLabel, container.GetId(), container.GetLabels())
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
	labels[schema.ContainerTypeLabel] = schema.ContainerTypeInteractiveREPL
	labels[schema.EnclaveIDContainerLabel] = enclaveId
	return labels
}
