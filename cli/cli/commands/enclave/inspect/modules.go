package inspect

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/kurtosis-core/commons/enclave_object_labels"
	"github.com/palantir/stacktrace"
)

const (
	moduleGUIDColHeader                                    = "GUID"
	moduleHostMachinePortBindingsColHeader                 = "LocalPortBindings"
	shouldShowStoppedContainersWhenGettingModuleContainers = true
)

func printModules(ctx context.Context, dockerManager *docker_manager.DockerManager, enclaveId string) error {
	moduleLabels := getLabelsForListEnclaveModules(enclaveId)

	containers, err := dockerManager.GetContainersByLabels(ctx, moduleLabels, shouldShowStoppedContainersWhenGettingModuleContainers)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting module containers by labels: '%+v'", moduleLabels)
	}

	tablePrinter := output_printers.NewTablePrinter(moduleGUIDColHeader, moduleHostMachinePortBindingsColHeader)
	sortedContainers, err := sortContainersByGUID(containers)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred sorting module containers by GUID")
	}
	for _, container := range sortedContainers {
		containerGuid, found := container.GetLabels()[enclave_object_labels.GUIDLabel]
		if !found {
			return stacktrace.NewError("No '%v' container label was found in container ID '%v' with labels '%+v'", enclave_object_labels.GUIDLabel, container.GetId(), container.GetLabels())
		}
		hostPortBindingsStrings := getContainerHostPortBindingStrings(container)

		firstHostPortBindingStr := ""
		if hostPortBindingsStrings != nil {
			firstHostPortBindingStr = hostPortBindingsStrings[0]
			hostPortBindingsStrings = hostPortBindingsStrings[1:]
		}
		if err := tablePrinter.AddRow(containerGuid, firstHostPortBindingStr); err != nil {
			return stacktrace.NewError(
				"An error occurred adding row for module container '%v' to the table printer",
				containerGuid,
			)
		}

		for _, additionalHostPortBindingStr := range hostPortBindingsStrings {
			if err := tablePrinter.AddRow("", additionalHostPortBindingStr); err != nil {
				return stacktrace.NewError(
					"An error occurred adding additional host port binding '%v' row for module container '%v' to the table printer",
					additionalHostPortBindingStr,
					containerGuid,
				)
			}
		}
	}
	tablePrinter.Print()

	return nil
}

func getLabelsForListEnclaveModules(enclaveId string) map[string]string {
	labels := map[string]string{}
	labels[enclave_object_labels.ContainerTypeLabel] = enclave_object_labels.ContainerTypeModuleContainer
	labels[enclave_object_labels.EnclaveIDContainerLabel] = enclaveId
	return labels
}
