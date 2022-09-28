package inspect

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/module"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/kurtosis-sdk/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/stacktrace"
	"sort"
	"strings"
)

const (
	moduleGUIDColHeader  = "GUID"
	moduleIDColHeader = "ID"
	modulePortsColHeader         = "Ports"
	defaultEmptyIPAddrForModules = ""

	grpcPortId = "grpc"
)

// TODO TODO When gateway binds public ports for modules, use isAPIContainerRunning to know to query for public port bindings.
func printModules(ctx context.Context, kurtosisBackend backend_interface.KurtosisBackend, enclaveInfo kurtosis_engine_rpc_api_bindings.EnclaveInfo, isAPIContainerRunning bool) error {
	enclaveIdStr := enclaveInfo.GetEnclaveId()
	enclaveId := enclave.EnclaveID(enclaveIdStr)
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

	tablePrinter := output_printers.NewTablePrinter(moduleGUIDColHeader, moduleIDColHeader, modulePortsColHeader)
	sortedModules := getSortedModuleSliceFromModulesMap(modules)

	for _, moduleObj := range sortedModules {
		moduleGuidStr := string(moduleObj.GetGUID())
		moduleIdStr := string(moduleObj.GetID())

		portString := getModulePortBindingString(moduleObj)

		if err := tablePrinter.AddRow(moduleGuidStr, moduleIdStr, portString); err != nil {
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
	for _, moduleObj := range modules {
		modulesResult = append(modulesResult, moduleObj)
	}

	sort.Slice(modulesResult, func(i, j int) bool {
		return modulesResult[i].GetGUID() < modulesResult[j].GetGUID()
	})

	return modulesResult
}

func getModulePortBindingString(module *module.Module) string {
	privatePort := module.GetPrivatePort()
	line := fmt.Sprintf(
		"%v: %v/%v",
		grpcPortId,
		privatePort.GetNumber(),
		strings.ToLower(privatePort.GetProtocol().String()),
	)

	// If the container is running, add host machine port binding information
	publicIpAddr := module.GetMaybePublicPort()
	publicPort := module.GetMaybePublicPort()
	if publicIpAddr != nil && publicPort != nil {
		line = line + fmt.Sprintf(" -> %v:%v", publicIpAddr, publicPort.GetNumber())
	}
	return line
}