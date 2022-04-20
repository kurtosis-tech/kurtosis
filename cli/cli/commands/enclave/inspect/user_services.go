package inspect

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/stacktrace"
	"sort"
)

const (
	userServiceGUIDColHeader                                    = "GUID"
	userServiceIDColHeader                                      = "ID"
	userServiceHostMachinePortBindingsColHeader                 = "LocalPortBindings"

	noUserServiceHostPortBindingsPlaceholder = "<none>"
)

func printUserServices(ctx context.Context, kurtosisBackend backend_interface.KurtosisBackend, enclaveId enclave.EnclaveID) error {

	userServiceFilters := &service.ServiceFilters{
		EnclaveIDs: map[enclave.EnclaveID]bool{
			enclaveId: true,
		},
	}

	userServices, err := kurtosisBackend.GetUserServices(ctx, userServiceFilters)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting user services using filters '%+v'", userServiceFilters)
	}

	tablePrinter := output_printers.NewTablePrinter(userServiceGUIDColHeader, userServiceIDColHeader, userServiceHostMachinePortBindingsColHeader)
	sortedUserServices:= getSortedUserServiceSliceFromUserServiceMap(userServices)
	for _, userService := range sortedUserServices {

		hostPortBindingsStrings := getContainerHostPortBindingStrings(userService)
		firstHostPortBindingStr := noUserServiceHostPortBindingsPlaceholder
		if hostPortBindingsStrings != nil {
			firstHostPortBindingStr = hostPortBindingsStrings[0]
			hostPortBindingsStrings = hostPortBindingsStrings[1:]
		}
		guidStr := string(userService.GetGUID())
		idStr := string(userService.GetID())

		if err := tablePrinter.AddRow(guidStr, idStr, firstHostPortBindingStr); err != nil {
			return stacktrace.Propagate(
				err,
				"An error occurred adding row for user service with GUID '%v' to the table printer",
				guidStr,
			)
		}

		for _, additionalHostPortBindingStr := range hostPortBindingsStrings {
			if err := tablePrinter.AddRow("", "", additionalHostPortBindingStr); err != nil {
				return stacktrace.Propagate(
					err,
					"An error occurred adding additional host port binding '%v' row for user service with GUID '%v' to the table printer",
					additionalHostPortBindingStr,
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
		return userServicesResult[i].GetGUID() < userServicesResult[j].GetGUID()
	})

	return userServicesResult
}

func getContainerHostPortBindingStrings(userService *service.Service) []string {
	var allHostPortBindings []string
	publicPorts := userService.GetMaybePublicPorts()
	if len(publicPorts) > 0 {
		//IF the service has at least one public port it will have set the public IP address
		publicIp := userService.GetMaybePublicIP()
		for portID, portSpec := range publicPorts {
			hostPortBindingString := fmt.Sprintf("%v -> %v:%v", portID, publicIp, portSpec.GetNumber())
			allHostPortBindings = append(allHostPortBindings, hostPortBindingString)
		}
	}

	return allHostPortBindings
}
