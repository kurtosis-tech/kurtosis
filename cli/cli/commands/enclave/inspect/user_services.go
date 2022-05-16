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
	"strings"
)

const (
	userServiceGUIDColHeader                                    = "GUID"
	userServiceIDColHeader    = "ID"
	userServicePortsColHeader = "Ports"

	missingPortPlaceholder = "<none>"
)

func printUserServices(ctx context.Context, kurtosisBackend backend_interface.KurtosisBackend, enclaveId enclave.EnclaveID) error {

	userServiceFilters := &service.ServiceFilters{}

	// TODO Switch to using the API container API once it can show *all* services (not just running ones)
	userServices, err := kurtosisBackend.GetUserServices(ctx, enclaveId, userServiceFilters)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting user services in enclave '%v' using filters '%+v'", enclaveId, userServiceFilters)
	}

	tablePrinter := output_printers.NewTablePrinter(
		userServiceGUIDColHeader,
		userServiceIDColHeader,
		userServicePortsColHeader,
	)
	sortedUserServices:= getSortedUserServiceSliceFromUserServiceMap(userServices)
	for _, userService := range sortedUserServices {
		idStr := string(userService.GetRegistration().GetID())
		guidStr := string(userService.GetRegistration().GetGUID())

		portBindingLines, err := getPortBindingStrings(userService)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting the port binding strings")
		}
		firstPortBindingLine := portBindingLines[0]
		additionalPortBindingLines := portBindingLines[1:]

		if err := tablePrinter.AddRow(guidStr, idStr, firstPortBindingLine); err != nil {
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
func getPortBindingStrings(userService *service.Service) ([]string, error) {
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
	if userService.GetMaybePublicIP() != nil && userService.GetMaybePublicPorts() != nil {
		publicIpAddr := userService.GetMaybePublicIP()
		publicPorts := userService.GetMaybePublicPorts()
		for portId := range privatePorts {
			publicPortSpec, found := publicPorts[portId]
			if !found {
				return nil, stacktrace.NewError(
					"Private port '%v' was declared on service '%v' and the container is running, but no corresponding public port " +
						"was found; this is very strange!",
					portId,
					userService.GetRegistration().GetGUID(),
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
