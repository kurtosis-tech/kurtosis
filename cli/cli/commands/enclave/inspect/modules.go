package inspect

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/module"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/enclave_liveness_validator"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/stacktrace"
	"google.golang.org/grpc"
	"sort"
	"strings"
)

const (
	moduleGUIDColHeader  = "GUID"
	modulePortsColHeader         = "Ports"
	defaultEmptyIPAddrForModules = ""

	grpcPortId = "grpc"
)

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

	// Pull module info from API container if it is running
	moduleInfoMapFromAPIC := map[string]*kurtosis_core_rpc_api_bindings.ModuleInfo{}
	if isAPIContainerRunning {
		moduleInfoMapFromAPIC, err = getModuleInfoMapFromAPIContainer(ctx, enclaveInfo)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to get module info from API container in enclave '%v'", enclaveInfo.GetEnclaveId())
		}
	}

	tablePrinter := output_printers.NewTablePrinter(moduleGUIDColHeader, modulePortsColHeader)
	sortedModules := getSortedModuleSliceFromModulesMap(modules)

	for _, moduleObj := range sortedModules {
		moduleGuidStr := string(moduleObj.GetGUID())
		moduleIdStr := string(moduleObj.GetID())

		// Look for public port and IP information in API container map
		var maybePublicPortFromAPIC uint32
		maybePublicIpAddrFromAPIC := ""
		moduleInfoFromAPIC, found := moduleInfoMapFromAPIC[moduleIdStr]
		if found {
			// Set public port from API container information
			maybePublicPortGRPC := moduleInfoFromAPIC.GetMaybePublicGrpcPort()
			if maybePublicPortGRPC != nil {
				maybePublicPortFromAPIC = maybePublicPortGRPC.GetNumber()
			}
			// Set public IP address from API container information
			maybePublicIpAddrFromAPIC = moduleInfoFromAPIC.GetMaybePublicIpAddr()
		}

		portString := getModulePortBindingString(moduleObj, &maybePublicPortFromAPIC, maybePublicIpAddrFromAPIC)

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
	for _, moduleObj := range modules {
		modulesResult = append(modulesResult, moduleObj)
	}

	sort.Slice(modulesResult, func(i, j int) bool {
		return modulesResult[i].GetGUID() < modulesResult[j].GetGUID()
	})

	return modulesResult
}

func getModulePortBindingString(module *module.Module, maybePublicPortFromAPIC *uint32, maybePublicIpAddrFromAPIC string) string {
	privatePort := module.GetPrivatePort()
	line := fmt.Sprintf(
		"%v: %v/%v",
		grpcPortId,
		privatePort.GetNumber(),
		strings.ToLower(privatePort.GetProtocol().String()),
	)

	// If the container is running, add host machine port binding information

	// If API container returned public port information, use that.
	var publicPort *uint32
	publicPortObjFromBackend := module.GetMaybePublicPort()
	if maybePublicPortFromAPIC != nil {
		publicPort = maybePublicPortFromAPIC
	} else if publicPortObjFromBackend != nil {
		publicPortUint32 := uint32(publicPortObjFromBackend.GetNumber())
		publicPort = &publicPortUint32
	}

	// If API container returned public IP information, use that.
	publicIpAddr := defaultEmptyIPAddrForModules
	publicIpAddrFromBackend := module.GetMaybePublicIP()
	if maybePublicIpAddrFromAPIC != defaultEmptyIPAddrForModules {
		publicIpAddr = maybePublicIpAddrFromAPIC
	} else if publicIpAddrFromBackend.String() != defaultEmptyIPAddrForModules {
		publicIpAddr = publicIpAddrFromBackend.String()
	}

	if publicIpAddr != defaultEmptyIPAddrForModules && publicPort != nil {
		line = line + fmt.Sprintf(" -> %v:%v", publicIpAddr, publicPort)
	}
	return line
}

func getModuleInfoMapFromAPIContainer(ctx context.Context, enclaveInfo kurtosis_engine_rpc_api_bindings.EnclaveInfo) (map[string]*kurtosis_core_rpc_api_bindings.ModuleInfo, error) {
	apicHostMachineIp, apicHostMachineGrpcPort, err := enclave_liveness_validator.ValidateEnclaveLiveness(&enclaveInfo)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred verifying that the enclave was running")
	}

	apiContainerHostGrpcUrl := fmt.Sprintf(
		"%v:%v",
		apicHostMachineIp,
		apicHostMachineGrpcPort,
	)
	conn, err := grpc.Dial(apiContainerHostGrpcUrl, grpc.WithInsecure())
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred connecting to the API container grpc port at '%v' in enclave '%v'",
			apiContainerHostGrpcUrl,
			enclaveInfo.EnclaveId,
		)
	}
	defer func() {
		conn.Close()
	}()
	apiContainerClient := kurtosis_core_rpc_api_bindings.NewApiContainerServiceClient(conn)

	getAllModulesMap := map[string]bool{}
	getAllModulesArgs := binding_constructors.NewGetModulesArgs(getAllModulesMap)
	allModulesResponse, err := apiContainerClient.GetModules(ctx, getAllModulesArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get service information for all services in enclave '%v'", enclaveInfo.GetEnclaveId())
	}
	moduleInfoMapFromAPIC := allModulesResponse.GetModuleInfo()
	return moduleInfoMapFromAPIC, nil
}
