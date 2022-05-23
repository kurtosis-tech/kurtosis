package inspect

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
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
	modulePortsColHeader = "Ports"

	grpcPortId = "grpc"
)

func printModules(ctx context.Context, enclaveInfo kurtosis_engine_rpc_api_bindings.EnclaveInfo) error {
	enclaveIdStr := enclaveInfo.EnclaveId
	enclaveId := enclave.EnclaveID(enclaveIdStr)
	apicHostMachineIp, apicHostMachineGrpcPort, err := enclave_liveness_validator.ValidateEnclaveLiveness(&enclaveInfo)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred verifying that the enclave was running")
	}

	apiContainerHostGrpcUrl := fmt.Sprintf(
		"%v:%v",
		apicHostMachineIp,
		apicHostMachineGrpcPort,
	)
	conn, err := grpc.Dial(apiContainerHostGrpcUrl, grpc.WithInsecure())
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred connecting to the API container grpc port at '%v' in enclave '%v'",
			apiContainerHostGrpcUrl,
			enclaveIdStr,
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
		return stacktrace.Propagate(err, "Failed to get service information for all services in enclave '%v'", enclaveId)
	}
	moduleInfoMap := allModulesResponse.GetModuleInfo()

	tablePrinter := output_printers.NewTablePrinter(moduleGUIDColHeader, modulePortsColHeader)
	sortedModules := getSortedModuleSliceFromModuleInfoMap(moduleInfoMap)

	for _, moduleInfo := range sortedModules {
		portString := getModulePortBindingString(moduleInfo)

		moduleGuidStr := moduleInfo.GetGuid()

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

func getSortedModuleSliceFromModuleInfoMap(moduleInfoMap map[string]*kurtosis_core_rpc_api_bindings.ModuleInfo) []*kurtosis_core_rpc_api_bindings.ModuleInfo {
	modulesResult := make([]*kurtosis_core_rpc_api_bindings.ModuleInfo, 0, len(moduleInfoMap))
	for _, moduleInfo := range moduleInfoMap {
		modulesResult = append(modulesResult, moduleInfo)
	}

	sort.Slice(modulesResult, func(i, j int) bool {
		return modulesResult[i].GetGuid() < modulesResult[j].GetGuid()
	})

	return modulesResult
}

func getModulePortBindingString(moduleInfo *kurtosis_core_rpc_api_bindings.ModuleInfo) string {
	privatePort := moduleInfo.GetPrivateGrpcPort()
	line := fmt.Sprintf(
		"%v: %v/%v",
		grpcPortId,
		privatePort.GetNumber(),
		strings.ToLower(privatePort.GetProtocol().String()),
	)

	// If the container is running, add host machine port binding information
	publicIpAddr := moduleInfo.GetMaybePublicIpAddr()
	publicPort := moduleInfo.GetMaybePublicGrpcPort()
	if publicIpAddr != "" && publicPort != nil {
		line = line + fmt.Sprintf(" -> %v:%v", publicIpAddr, publicPort.GetNumber())
	}
	return line
}
