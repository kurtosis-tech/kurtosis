package inspect

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/module"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/stacktrace"
	"sort"
	"strings"
)

const (
	moduleGUIDColHeader  = "GUID"
	modulePortsColHeader = "Ports"
)

func printModules(ctx context.Context, kurtosisBackend backend_interface.KurtosisBackend, enclaveId enclave.EnclaveID) error {
	moduleFilters := &module.ModuleFilters{
		Statuses: map[container_status.ContainerStatus]bool{
			container_status.ContainerStatus_Stopped: true,
			container_status.ContainerStatus_Running: true,
		},
	}

	modules, err := kurtosisBackend.GetModules(ctx, enclaveId, moduleFilters)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting modules using filters '%+v'", moduleFilters)
	}

	tablePrinter := output_printers.NewTablePrinter(moduleGUIDColHeader, modulePortsColHeader)
	sortedModules := getSortedModuleSliceFromModulesMap(modules)

	for _, module := range sortedModules {

		portString := fmt.Sprintf(
			"%v/%v",
			 module.GetPrivatePort().GetNumber(),
			 strings.ToLower(module.GetPrivatePort().GetProtocol().String()),
		)
		if module.GetStatus() == container_status.ContainerStatus_Running {
			portString = portString + fmt.Sprintf(
				" -> %v:%v",
				module.GetPublicIp(),
				module.GetPublicPort().GetNumber(),
			)
		}

		moduleGuidStr := string(module.GetGUID())

		if err := tablePrinter.AddRow(moduleGuidStr, portString); err != nil {
			return stacktrace.NewError(
				"An error occurred adding row for module GUID '%v' to the table printer",
				moduleGuidStr,
			)
		}

	}

	tablePrinter.Print()

	return nil
}

func getSortedModuleSliceFromModulesMap(modules map[module.ModuleGUID]*module.Module) []*module.Module {

	modulesResult := make([]*module.Module, 0, len(modules))
	for _, module := range modules {
		modulesResult = append(modulesResult, module)
	}

	sort.Slice(modulesResult, func(i, j int) bool {
		return modulesResult[i].GetGUID() < modulesResult[j].GetGUID()
	})

	return modulesResult
}
