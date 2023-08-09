package inspect

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/enclave_liveness_validator"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/stacktrace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"sort"
	"strings"
)

const (
	userServiceUUIDColHeader      = "UUID"
	userServiceNameColHeader      = "Name"
	userServicePortsColHeader     = "Ports"
	userServiceStatusColHeader    = "Status"
	defaultEmptyIPAddrForServices = ""
	emptyApplicationProtocol      = ""
	missingPortPlaceholder        = "<none>"
	linkDelimeter                 = "://"
	defaultEmptyIPAddrForAPIC     = ""
)

var (
	colorizeRunning = color.New(color.FgGreen).SprintFunc()
	colorizeStopped = color.New(color.FgYellow).SprintFunc()
)

func printUserServices(ctx context.Context, _ *kurtosis_context.KurtosisContext, enclaveInfo *kurtosis_engine_rpc_api_bindings.EnclaveInfo, showFullUuids bool, isAPIContainerRunning bool) error {
	userServices := map[string]*kurtosis_core_rpc_api_bindings.ServiceInfo{}
	if isAPIContainerRunning {
		var err error
		userServices, err = getUserServiceInfoMapFromAPIContainer(ctx, enclaveInfo)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to get service info from API container in enclave '%v'", enclaveInfo.GetEnclaveUuid())
		}
	}

	tablePrinter := output_printers.NewTablePrinter(
		userServiceUUIDColHeader,
		userServiceNameColHeader,
		userServicePortsColHeader,
		userServiceStatusColHeader,
	)
	sortedUserServices := getSortedUserServiceSliceFromUserServiceMap(userServices)
	for _, userService := range sortedUserServices {
		serviceIdStr := userService.GetName()
		uuidStr := userService.GetServiceUuid()
		uuidToPrint := userService.GetShortenedUuid()
		if showFullUuids {
			uuidToPrint = uuidStr
		}

		serviceStatus := userService.GetServiceStatus()
		serviceStatusStr := colorizeServiceStatus(serviceStatus)

		portBindingLines, err := getUserServicePortBindingStrings(userService)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting the port binding strings")
		}
		firstPortBindingLine := portBindingLines[0]
		additionalPortBindingLines := portBindingLines[1:]

		if err := tablePrinter.AddRow(uuidToPrint, serviceIdStr, firstPortBindingLine, serviceStatusStr); err != nil {
			return stacktrace.Propagate(
				err,
				"An error occurred adding row for user service with UUID '%v' to the table printer",
				uuidStr,
			)
		}

		for _, additionalPortBindingLine := range additionalPortBindingLines {
			if err := tablePrinter.AddRow("", "", additionalPortBindingLine, ""); err != nil {
				return stacktrace.Propagate(
					err,
					"An error occurred adding additional port binding row '%v' for user service with UUID '%v' to the table printer",
					additionalPortBindingLine,
					uuidStr,
				)
			}
		}
	}
	tablePrinter.Print()

	return nil
}

func getSortedUserServiceSliceFromUserServiceMap(userServices map[string]*kurtosis_core_rpc_api_bindings.ServiceInfo) []*kurtosis_core_rpc_api_bindings.ServiceInfo {
	userServicesResult := make([]*kurtosis_core_rpc_api_bindings.ServiceInfo, 0, len(userServices))
	for _, userService := range userServices {
		userServicesResult = append(userServicesResult, userService)
	}

	sort.Slice(userServicesResult, func(i, j int) bool {
		firstService := userServicesResult[i]
		secondService := userServicesResult[j]
		return firstService.GetName() < secondService.GetName()
	})

	return userServicesResult
}

// Guaranteed to have at least one entry
func getUserServicePortBindingStrings(userService *kurtosis_core_rpc_api_bindings.ServiceInfo) ([]string, error) {
	privatePorts := userService.GetPrivatePorts()
	if len(privatePorts) == 0 {
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
			strings.ToLower(privatePortSpec.GetTransportProtocol().String()),
		)
		resultLines[portId] = line
	}

	maybePublicIpAddr := userService.GetMaybePublicIpAddr()
	maybePublicPortMap := userService.GetMaybePublicPorts()

	// If the container is running, add host machine port binding information
	if maybePublicIpAddr != defaultEmptyIPAddrForServices && len(maybePublicPortMap) > 0 {
		publicIpAddr := maybePublicIpAddr
		publicPorts := maybePublicPortMap
		for portId := range privatePorts {
			publicPortSpec, found := publicPorts[portId]
			// With Kubernetes, it's now possible for a private port to not have a corresponding public port
			// We only expose TCP ports through the gateway
			if !found {
				continue
			}
			currentPortLine := resultLines[portId]

			applicationProtocol := emptyApplicationProtocol
			privatePortSpec, found := privatePorts[portId]
			if !found {
				return nil, stacktrace.NewError("port spec associated with %v is not found", portId)
			}
			if privatePortSpec.GetMaybeApplicationProtocol() != "" {
				applicationProtocol = fmt.Sprintf("%v%v", privatePorts[portId].GetMaybeApplicationProtocol(), linkDelimeter)
			}
			resultLines[portId] = currentPortLine + fmt.Sprintf(" -> %v%v:%v", applicationProtocol, publicIpAddr, publicPortSpec.GetNumber())
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

func getUserServiceInfoMapFromAPIContainer(ctx context.Context, enclaveInfo *kurtosis_engine_rpc_api_bindings.EnclaveInfo) (map[string]*kurtosis_core_rpc_api_bindings.ServiceInfo, error) {
	apicHostMachineIp, apicHostMachineGrpcPort, err := enclave_liveness_validator.ValidateEnclaveLiveness(enclaveInfo)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred verifying that the enclave was running")
	}

	apiContainerHostGrpcUrl := fmt.Sprintf(
		"%v:%v",
		apicHostMachineIp,
		apicHostMachineGrpcPort,
	)
	conn, err := grpc.Dial(apiContainerHostGrpcUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred connecting to the API container grpc port at '%v' in enclave '%v'",
			apiContainerHostGrpcUrl,
			enclaveInfo.EnclaveUuid,
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
		return nil, stacktrace.Propagate(err, "Failed to get service information for all services in enclave '%v'", enclaveInfo.GetEnclaveUuid())
	}
	serviceInfoMapFromAPIC := allServicesResponse.GetServiceInfo()
	return serviceInfoMapFromAPIC, nil
}

func colorizeServiceStatus(serviceStatus kurtosis_core_rpc_api_bindings.ServiceStatus) string {
	serviceStatusStr := kurtosis_core_rpc_api_bindings.ServiceStatus_name[int32(serviceStatus)]
	switch serviceStatus {
	case kurtosis_core_rpc_api_bindings.ServiceStatus_STOPPED:
		return colorizeStopped(serviceStatusStr)
	case kurtosis_core_rpc_api_bindings.ServiceStatus_RUNNING:
		return colorizeRunning(serviceStatusStr)
	default:
		return serviceStatusStr
	}
}
