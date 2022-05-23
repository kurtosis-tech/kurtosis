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
	userServiceGUIDColHeader                                    = "GUID"
	userServiceIDColHeader    = "ID"
	userServicePortsColHeader = "Ports"

	missingPortPlaceholder = "<none>"
)

func printUserServices(ctx context.Context, enclaveInfo kurtosis_engine_rpc_api_bindings.EnclaveInfo) error {
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

	getAllServicesMap := map[string]bool{}
	getAllServicesArgs := binding_constructors.NewGetServicesArgs(getAllServicesMap)
	allServicesResponse, err := apiContainerClient.GetServices(ctx, getAllServicesArgs)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to get service information for all services in enclave '%v'", enclaveId)
	}
	userServiceInfoMap := allServicesResponse.GetServiceInfo()

	tablePrinter := output_printers.NewTablePrinter(
		userServiceGUIDColHeader,
		userServiceIDColHeader,
		userServicePortsColHeader,
	)

	sortedUserServiceInfoTuples := getSortedUserServiceSliceFromUserServiceInfoMap(userServiceInfoMap)

	for _, userServiceInfoTuple := range sortedUserServiceInfoTuples {
		serviceIdStr := userServiceInfoTuple.serviceId
		userServiceInfo := userServiceInfoTuple.serviceInfo
		serviceGuid := userServiceInfo.GetServiceGuid()
		serviceGuidStr := string(serviceGuid)

		portBindingLines, err := getUserServicePortBindingStringsFromServiceInfo(userServiceInfo)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to create port binding strings from service info.")
		}
		firstPortBindingLine := portBindingLines[0]
		additionalPortBindingLines := portBindingLines[1:]

		if err := tablePrinter.AddRow(serviceGuidStr, serviceIdStr, firstPortBindingLine); err != nil {
			return stacktrace.Propagate(
				err,
				"An error occurred adding row for user service with GUID '%v' to the table printer",
				serviceGuidStr,
			)
		}

		for _, additionalPortBindingLine := range additionalPortBindingLines {
			if err := tablePrinter.AddRow("", "", additionalPortBindingLine); err != nil {
				return stacktrace.Propagate(
					err,
					"An error occurred adding additional port binding row '%v' for user service with GUID '%v' to the table printer",
					additionalPortBindingLine,
					serviceGuidStr,
				)
			}
		}
	}
	tablePrinter.Print()
	return nil
}

type serviceIdServiceInfoTuple struct {
	serviceId string
	serviceInfo *kurtosis_core_rpc_api_bindings.ServiceInfo
}

func getSortedUserServiceSliceFromUserServiceInfoMap(userServiceInfoMap map[string]*kurtosis_core_rpc_api_bindings.ServiceInfo) []*serviceIdServiceInfoTuple {
	userServicesResult := make([]*serviceIdServiceInfoTuple, 0, len(userServiceInfoMap))
	for serviceId, userServiceInfo := range userServiceInfoMap {
		userServicesResult = append(userServicesResult, &serviceIdServiceInfoTuple{
			serviceId: serviceId,
			serviceInfo: userServiceInfo})
	}

	sort.Slice(userServicesResult, func(i, j int) bool {
		firstService := userServicesResult[i]
		secondService := userServicesResult[j]
		return firstService.serviceInfo.GetServiceGuid() < secondService.serviceInfo.GetServiceGuid()
	})

	return userServicesResult
}

// Guaranteed to have at least one entry
func getUserServicePortBindingStringsFromServiceInfo(serviceInfo *kurtosis_core_rpc_api_bindings.ServiceInfo) ([]string, error) {
	privatePorts := serviceInfo.GetPrivatePorts()
	if privatePorts == nil || len(privatePorts) == 0 {
		return []string{missingPortPlaceholder}, nil
	}

	portIds := []string{}
	resultLines := map[string]string{}
	for portId, privatePortSpec := range privatePorts {
		portIds = append(portIds, portId)
		line := fmt.Sprintf(
			"%v: %v/%v",
			portId,
			privatePortSpec.GetNumber(),
			strings.ToLower(privatePortSpec.GetProtocol().String()),
		)
		resultLines[portId] = line
	}

	// If the container is running, add host machine port binding information
	if serviceInfo.GetMaybePublicIpAddr() != "" && serviceInfo.GetMaybePublicPorts() != nil {
		publicIpAddr := serviceInfo.GetMaybePublicIpAddr()
		publicPorts := serviceInfo.GetMaybePublicPorts()
		for portId := range privatePorts {
			publicPortSpec, found := publicPorts[portId]
			if !found {
				return nil, stacktrace.NewError(
					"Private port '%v' was declared on service '%v' and the container is running, but no corresponding public port " +
						"was found; this is very strange!",
					portId,
					serviceInfo.GetServiceGuid(),
				)
			}
			currentPortLine := resultLines[portId]
			resultLines[portId] = currentPortLine + fmt.Sprintf(" -> %v:%v", publicIpAddr, publicPortSpec.GetNumber())
		}
	}

	// Finally, sort the resulting lines by port ID
	sort.Strings(portIds)
	result := []string{}
	for _, portId := range portIds {
		result = append(result, resultLines[portId])
	}

	return result, nil
}
