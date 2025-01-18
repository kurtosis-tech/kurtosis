package inspect

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/container_status_stringifier"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/user_services"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	userServiceUUIDColHeader   = "UUID"
	userServiceNameColHeader   = "Name"
	userServicePortsColHeader  = "Ports"
	userServiceStatusColHeader = "Status"
)

func printUserServices(ctx context.Context, _ *kurtosis_context.KurtosisContext, enclaveInfo *kurtosis_engine_rpc_api_bindings.EnclaveInfo, showFullUuids bool, isAPIContainerRunning bool) error {
	userServices := map[string]*kurtosis_core_rpc_api_bindings.ServiceInfo{}
	if isAPIContainerRunning {
		var err error
		allServicesMap := map[string]bool{}
		userServices, err = user_services.GetUserServiceInfoMapFromAPIContainer(ctx, enclaveInfo, allServicesMap)
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
			return stacktrace.Propagate(err, "An error occurred getting the port binding strings")
		}
		firstPortBindingLine := portBindingLines[0]
		additionalPortBindingLines := portBindingLines[1:]

		if err := tablePrinter.AddRow(uuidToPrint, serviceIdStr, firstPortBindingLine, containerStatusStr); err != nil {
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
