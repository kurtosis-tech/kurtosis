package inspect

import (
	"context"

	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/container_status_stringifier"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/user_services"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	userServiceUUIDColHeader   = "UUID"
	userServiceNameColHeader   = "Name"
	userServicePortsColHeader  = "Ports"
	userServiceStatusColHeader = "Status"
)

func inspectUserServices(ctx context.Context, _ *kurtosis_context.KurtosisContext, enclaveInfo *kurtosis_engine_rpc_api_bindings.EnclaveInfo, showFullUuids bool, isAPIContainerRunning bool) ([]UserService, error) {
	userServices := map[string]*kurtosis_core_rpc_api_bindings.ServiceInfo{}
	if isAPIContainerRunning {
		var err error
		allServicesMap := map[string]bool{}
		userServices, err = user_services.GetUserServiceInfoMapFromAPIContainer(ctx, enclaveInfo, allServicesMap)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to get service info from API container in enclave '%v'", enclaveInfo.GetEnclaveUuid())
		}
	}
	var services []UserService
	sortedUserServices := user_services.GetSortedUserServiceSliceFromUserServiceMap(userServices)
	for _, userService := range sortedUserServices {
		serviceIdStr := userService.GetName()
		uuidStr := userService.GetServiceUuid()
		uuidToPrint := userService.GetShortenedUuid()
		if showFullUuids {
			uuidToPrint = uuidStr
		}

		containerStatus := userService.GetContainer().GetStatus()
		containerStatusStr := container_status_stringifier.ContainerStatusStringifier(containerStatus)

		portBindingLines, err := user_services.GetUserServicePortBindingStrings(userService)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting the port binding strings")
		}
		services = append(services, UserService{
			UUID:   uuidToPrint,
			Name:   serviceIdStr,
			Status: containerStatusStr,
			Ports:  portBindingLines,
		})
	}

	return services, nil
}
