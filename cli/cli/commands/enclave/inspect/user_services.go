package inspect

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis-core/commons/enclave_object_labels"
	"github.com/palantir/stacktrace"
	"os"
	"sort"
	"text/tabwriter"
)

func printUserServices(ctx context.Context, dockerManager *docker_manager.DockerManager, enclaveId string) error {
	userServiceLabels := getLabelsForListEnclaveUserServices(enclaveId)

	containers, err := dockerManager.GetContainersByLabels(ctx, userServiceLabels, true)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting user service containers by labels: '%+v'", userServiceLabels)
	}

	if containers != nil {
		tabWriter := tabwriter.NewWriter(os.Stdout, tabWriterMinwidth, tabWriterTabwidth, tabWriterPadding, tabWriterPadchar, tabWriterFlags)
		fmt.Fprintln(tabWriter, guidHeader + "\t" + nameHeader + "\t" + hostPortBindingsHeader)
		sortedContainers, err := getContainersSortedByGUID(containers)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting user service containers sorted by GUID")
		}
		for _, container := range sortedContainers {
			containerGUIDLabel, found := container.GetLabels()[enclave_object_labels.GUIDLabel]
			if !found {
				return stacktrace.NewError("No '%v' container label was found in container ID '%v' with labels '%+v'", enclave_object_labels.GUIDLabel, container.GetId(), container.GetLabels())
			}
			hostPortBindingsStrings := getContainerHostPortBindingStrings(container)

			var firstHostPortBinding string
			if hostPortBindingsStrings != nil  {
				firstHostPortBinding = hostPortBindingsStrings[0]
				hostPortBindingsStrings = hostPortBindingsStrings[1:]
			}
			line := containerGUIDLabel + "\t" + container.GetName() + "\t" + firstHostPortBinding
			fmt.Fprintln(tabWriter, line)

			for _, hostPortBindingsString := range hostPortBindingsStrings {
				line = "\t\t" + hostPortBindingsString
				fmt.Fprintln(tabWriter, line)
			}
		}
		tabWriter.Flush()
	}

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
	labels[enclave_object_labels.ContainerTypeLabel] = enclave_object_labels.ContainerTypeUserServiceContainer
	labels[enclave_object_labels.EnclaveIDContainerLabel] = enclaveId
	return labels
}

func getContainersSortedByGUID(containers []*types.Container) ([]*types.Container, error) {
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
