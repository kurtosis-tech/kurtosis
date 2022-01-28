package inspect

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/object-attributes-schema-lib/forever_constants"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	userServiceGUIDColHeader                                    = "GUID"
	userServiceIDColHeader                                      = "ID"
	userServiceHostMachinePortBindingsColHeader                 = "LocalPortBindings"
	shouldShowStoppedContainersWhenGettingUserServiceContainers = true
)

func printUserServices(ctx context.Context, dockerManager *docker_manager.DockerManager, enclaveId string) error {
	userServiceLabels := getLabelsForListEnclaveUserServices(enclaveId)

	containers, err := dockerManager.GetContainersByLabels(ctx, userServiceLabels, shouldShowStoppedContainersWhenGettingUserServiceContainers)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting user service containers by labels: '%+v'", userServiceLabels)
	}

	tablePrinter := output_printers.NewTablePrinter(userServiceGUIDColHeader, userServiceIDColHeader, userServiceHostMachinePortBindingsColHeader)
	sortedContainers, err := sortContainersByGUID(containers)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred sorting user service containers by GUID")
	}
	for _, container := range sortedContainers {
		serviceGuid, found := container.GetLabels()[schema.GUIDLabel]
		if !found {
			return stacktrace.NewError("No '%v' container label was found in container ID '%v' with labels '%+v'", schema.GUIDLabel, container.GetId(), container.GetLabels())
		}
		serviceId, found := container.GetLabels()[schema.IDLabel]
		if !found {
			return stacktrace.NewError("No '%v' container label was found in container ID '%v' with labels '%+v'", schema.IDLabel, container.GetId(), container.GetLabels())
		}

		hostPortBindingsStrings := getContainerHostPortBindingStrings(container)
		firstHostPortBindingStr := ""
		if hostPortBindingsStrings != nil {
			firstHostPortBindingStr = hostPortBindingsStrings[0]
			hostPortBindingsStrings = hostPortBindingsStrings[1:]
		}
		if err := tablePrinter.AddRow(serviceGuid, serviceId, firstHostPortBindingStr); err != nil {
			return stacktrace.Propagate(
				err,
				"An error occurred adding row for user service container '%v' to the table printer",
				serviceGuid,
			)
		}

		for _, additionalHostPortBindingStr := range hostPortBindingsStrings {
			if err := tablePrinter.AddRow("", "", additionalHostPortBindingStr); err != nil {
				return stacktrace.Propagate(
					err,
					"An error occurred adding additional host port binding '%v' row for user service container '%v' to the table printer",
					additionalHostPortBindingStr,
					serviceGuid,
				)
			}
		}
	}
	tablePrinter.Print()

	return nil
}

func getContainerHostPortBindingStrings(container *types.Container) []string {
	var allHosPortBindings []string
	hostPortBindings := container.GetHostPortBindings()
	for hostPortBindingKey, hostPortBinding := range hostPortBindings {
		hostPortBindingString := fmt.Sprintf("%v -> %v:%v", hostPortBindingKey, hostPortBinding.HostIP, hostPortBinding.HostPort)
		allHosPortBindings = append(allHosPortBindings, hostPortBindingString)
	}
	return allHosPortBindings
}

func getLabelsForListEnclaveUserServices(enclaveId string) map[string]string {
	labels := map[string]string{}
	labels[forever_constants.ContainerTypeLabel] = schema.ContainerTypeUserServiceContainer
	labels[schema.EnclaveIDContainerLabel] = enclaveId
	return labels
}
