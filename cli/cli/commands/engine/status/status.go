package status

import (
	"context"
	portal_constructors "github.com/kurtosis-tech/kurtosis-portal/api/golang/constructors"
	portal_api "github.com/kurtosis-tech/kurtosis-portal/api/golang/generated"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/portal_manager"
	"github.com/kurtosis-tech/kurtosis/contexts-config-store/store"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	enginePortTransportProtocol = portal_api.TransportProtocol_TCP
)

// StatusCmd Suppressing exhaustruct requirement because this struct has ~40 properties
// nolint: exhaustruct
var StatusCmd = &cobra.Command{
	Use:   command_str_consts.EngineStatusCmdStr,
	Short: "Reports the status of the Kurtosis engine",
	RunE:  run,
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	currentContext, err := store.GetContextsConfigStore().GetCurrentContext()
	if err == nil {
		if store.IsRemote(currentContext) {
			portalManager := portal_manager.NewPortalManager()
			if portalManager.IsReachable() {
				// Forward the remote engine port to the local machine
				portalClient := portalManager.GetClient()
				forwardEnginePortArgs := portal_constructors.NewForwardPortArgs(uint32(kurtosis_context.DefaultGrpcEngineServerPortNum), uint32(kurtosis_context.DefaultGrpcEngineServerPortNum), &enginePortTransportProtocol)
				if _, err := portalClient.ForwardPort(ctx, forwardEnginePortArgs); err != nil {
					return stacktrace.Propagate(err, "Unable to forward the remote engine port to the local machine")
				}
			}
		}
	} else {
		logrus.Warnf("Unable to retrieve current Kurtosis context. This is not critical, it will assume using Kurtosis default context for now.")
	}

	engineManager, err := engine_manager.NewEngineManager(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating an engine manager")
	}

	status, _, maybeApiVersion, err := engineManager.GetEngineStatus(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the Kurtosis engine status")
	}
	prettyPrintingStatusVisitor := newPrettyPrintingEngineStatusVisitor(maybeApiVersion)
	if err := status.Accept(prettyPrintingStatusVisitor); err != nil {
		return stacktrace.Propagate(err, "An error occurred printing the engine status")
	}

	return nil
}
