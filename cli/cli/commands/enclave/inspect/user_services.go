package inspect

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/enclave_liveness_validator"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/kurtosis-core/api/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-core/api/golang/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis-engine-server/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/stacktrace"
	"google.golang.org/grpc"
	"sort"
	"strings"
)

const (
	userServiceGUIDColHeader      = "GUID"
	userServiceIDColHeader        = "ID"
	userServicePortsColHeader     = "Ports"
	defaultEmptyIPAddrForServices = ""

	missingPortPlaceholder = "<none>"
)

func printUserServices(ctx context.Context, kurtosisBackend backend_interface.KurtosisBackend, enclaveInfo kurtosis_engine_rpc_api_bindings.EnclaveInfo, isAPIContainerRunning bool) error {
	enclaveIdStr := enclaveInfo.GetEnclaveId()
	enclaveId := enclave.EnclaveID(enclaveIdStr)
	userServiceFilters := &service.ServiceFilters{}

	userServices, err := kurtosisBackend.GetUserServices(ctx, enclaveId, userServiceFilters)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting user services in enclave '%v' using filters '%+v'", enclaveId, userServiceFilters)
	}

	// Pull service info from API container if it is running
	serviceInfoMapFromAPIC := map[string]*kurtosis_core_rpc_api_bindings.ServiceInfo{}
	if isAPIContainerRunning {
		serviceInfoMapFromAPIC, err = getUserServiceInfoMapFromAPIContainer(ctx, enclaveInfo)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to get service info from API container in enclave '%v'", enclaveInfo.GetEnclaveId())
		}
	}

	tablePrinter := output_printers.NewTablePrinter(
		userServiceGUIDColHeader,
		userServiceIDColHeader,
		userServicePortsColHeader,
	)
	sortedUserServices := getSortedUserServiceSliceFromUserServiceMap(userServices)
	for _, userService := range sortedUserServices {
		serviceIdStr := string(userService.GetRegistration().GetID())
		guidStr := string(userService.GetRegistration().GetGUID())

		// Look for public port and IP information in API container map
		maybePublicPortMapFromAPIC := map[string]*kurtosis_core_rpc_api_bindings.Port{}
		maybePublicIpAddrFromAPIC := defaultEmptyIPAddrForModules
		serviceInfoFromAPIC, found := serviceInfoMapFromAPIC[serviceIdStr]
		if found {
			// Set public port from API container information
			maybePublicPortMapFromAPIC = serviceInfoFromAPIC.GetMaybePublicPorts()
			// Set public IP address from API container information
			maybePublicIpAddrFromAPIC = serviceInfoFromAPIC.GetMaybePublicIpAddr()
		}

		portBindingLines, err := getUserServicePortBindingStrings(userService, maybePublicPortMapFromAPIC, maybePublicIpAddrFromAPIC)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting the port binding strings")
		}
		firstPortBindingLine := portBindingLines[0]
		additionalPortBindingLines := portBindingLines[1:]

		if err := tablePrinter.AddRow(guidStr, serviceIdStr, firstPortBindingLine); err != nil {
			return stacktrace.Propagate(
				err,
				"An error occurred adding row for user service with GUID '%v' to the table printer",
				guidStr,
			)
		}

		for _, additionalPortBindingLine := range additionalPortBindingLines {
			if err := tablePrinter.AddRow("", "", additionalPortBindingLine); err != nil {
				return stacktrace.Propagate(
					err,
					"An error occurred adding additional port binding row '%v' for user service with GUID '%v' to the table printer",
					additionalPortBindingLine,
					guidStr,
				)
			}
		}
	}
	tablePrinter.Print()

	return nil
}

func getSortedUserServiceSliceFromUserServiceMap(userServices map[service.ServiceGUID]*service.Service) []*service.Service {
	userServicesResult := make([]*service.Service, 0, len(userServices))
	for _, userService := range userServices {
		userServicesResult = append(userServicesResult, userService)
	}

	sort.Slice(userServicesResult, func(i, j int) bool {
		firstService := userServicesResult[i]
		secondService := userServicesResult[j]
		return firstService.GetRegistration().GetGUID() < secondService.GetRegistration().GetGUID()
	})

	return userServicesResult
}

// Guaranteed to have at least one entry
func getUserServicePortBindingStrings(userService *service.Service,
	maybePublicPortMapFromAPIC map[string]*kurtosis_core_rpc_api_bindings.Port,
	maybePublicIpAddrFromAPIC string) ([]string, error) {
	privatePorts := userService.GetPrivatePorts()
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
	if maybePublicIpAddrFromAPIC != defaultEmptyIPAddrForServices && len(maybePublicPortMapFromAPIC) > 0 {
		publicIpAddr := maybePublicIpAddrFromAPIC
		publicPorts := maybePublicPortMapFromAPIC
		for portId := range privatePorts {
			publicPortSpec, found := publicPorts[portId]
			// With Kubernetes, it's now possible for a private port to not have a corresponding public port
			// We only expose TCP ports through the gateway
			if !found {
				continue
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

func getUserServiceInfoMapFromAPIContainer(ctx context.Context, enclaveInfo kurtosis_engine_rpc_api_bindings.EnclaveInfo) (map[string]*kurtosis_core_rpc_api_bindings.ServiceInfo, error) {
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

	getAllServicesMap := map[string]bool{}
	getAllServicesArgs := binding_constructors.NewGetServicesArgs(getAllServicesMap)
	allServicesResponse, err := apiContainerClient.GetServices(ctx, getAllServicesArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get service information for all services in enclave '%v'", enclaveInfo.GetEnclaveId())
	}
	serviceInfoMapFromAPIC := allServicesResponse.GetServiceInfo()
	return serviceInfoMapFromAPIC, nil
}
