package user_services

import (
	"context"
	"fmt"
	"github.com/dzobbe/PoTE-kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/dzobbe/PoTE-kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/dzobbe/PoTE-kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/helpers/enclave_liveness_validator"
	"github.com/kurtosis-tech/stacktrace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"sort"
	"strings"
)

const (
	defaultEmptyIPAddrForServices = ""
	emptyApplicationProtocol      = ""
	missingPortPlaceholder        = "<none>"
	linkDelimeter                 = "://"
)

func GetSortedUserServiceSliceFromUserServiceMap(userServices map[string]*kurtosis_core_rpc_api_bindings.ServiceInfo) []*kurtosis_core_rpc_api_bindings.ServiceInfo {
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
func GetUserServicePortBindingStrings(userService *kurtosis_core_rpc_api_bindings.ServiceInfo) ([]string, error) {
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

func GetUserServiceInfoMapFromAPIContainer(ctx context.Context, enclaveInfo *kurtosis_engine_rpc_api_bindings.EnclaveInfo, filterServiceIdentifiers map[string]bool) (map[string]*kurtosis_core_rpc_api_bindings.ServiceInfo, error) {
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

	getAllServicesArgs := binding_constructors.NewGetServicesArgs(filterServiceIdentifiers)
	allServicesResponse, err := apiContainerClient.GetServices(ctx, getAllServicesArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get service information for all services in enclave '%v'", enclaveInfo.GetEnclaveUuid())
	}
	serviceInfoMapFromAPIC := allServicesResponse.GetServiceInfo()
	return serviceInfoMapFromAPIC, nil
}
