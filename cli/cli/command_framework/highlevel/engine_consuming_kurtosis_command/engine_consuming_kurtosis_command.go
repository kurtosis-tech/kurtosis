package engine_consuming_kurtosis_command

import (
	"context"
	portal_constructors "github.com/kurtosis-tech/kurtosis-portal/api/golang/constructors"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/metrics_client_factory"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/portal_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/contexts-config-store/store"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type EngineContextKey string

const (
	engineClientCloseFuncCtxKey     EngineContextKey = "engine-client-close-func"
	metricsClientKey                EngineContextKey = "metrics-client-key"
	metricsClientClosingFunctionKey EngineContextKey = "metrics-client-closing-func"
)

// This is a convenience KurtosisCommand for commands that interact with the engine
type EngineConsumingKurtosisCommand struct {
	// The string for the command (e.g. "inspect" or "ls")
	CommandStr string

	// Will be used when displaying the command for tab completion
	ShortDescription string

	LongDescription string

	// The name of the key that will be set during PreValidationAndRun where the KurtosisBackend can be found
	KurtosisBackendContextKey EngineContextKey

	// TODO Replace with KurtosisContext!!! This will:
	//  1) be easier to work with and
	//  2) force us to use the same SDK we give to users, so there's no "secret" or "private" API, which will force
	//     us to improve the SDK
	// The name of the key that will be set during PreValidationAndRun where the engine client will be made available
	EngineClientContextKey EngineContextKey

	// Order isn't important here
	Flags []*flags.FlagConfig

	Args []*args.ArgConfig

	RunFunc func(
		ctx context.Context,
		// TODO This is a hack that's only here temporarily because we have commands that use KurtosisBackend directly (they
		//  should not), and EngineConsumingKurtosisCommand therefore needs to provide them with a KurtosisBackend. Once all our
		//  commands only access the Kurtosis APIs, we can remove this.
		kurtosisBackend backend_interface.KurtosisBackend,
		engineClient kurtosis_engine_rpc_api_bindings.EngineServiceClient,
		client metrics_client.MetricsClient,
		flags *flags.ParsedFlags,
		args *args.ParsedArgs,
	) error
}

func (cmd *EngineConsumingKurtosisCommand) MustGetCobraCommand() *cobra.Command {
	// Validation
	if cmd.KurtosisBackendContextKey == engineClientCloseFuncCtxKey {
		panic(stacktrace.NewError(
			"Kurtosis backend context key '%v' on command '%v' is equal to engine client close function context key '%v'; this is a bug in Kurtosis!",
			cmd.KurtosisBackendContextKey,
			cmd.CommandStr,
			engineClientCloseFuncCtxKey,
		))
	}

	if cmd.EngineClientContextKey == engineClientCloseFuncCtxKey {
		panic(stacktrace.NewError(
			"Engine client context key '%v' on command '%v' is equal to engine client close function context key '%v'; this is a bug in Kurtosis!",
			cmd.EngineClientContextKey,
			cmd.CommandStr,
			engineClientCloseFuncCtxKey,
		))
	}
	if cmd.KurtosisBackendContextKey == cmd.EngineClientContextKey {
		panic(stacktrace.NewError(
			"Kurtosis backend context key '%v' on command '%v' is equal to engine client close function context key '%v'; this is a bug in Kurtosis!",
			cmd.KurtosisBackendContextKey,
			cmd.CommandStr,
			cmd.EngineClientContextKey,
		))
	}

	if cmd.KurtosisBackendContextKey == metricsClientClosingFunctionKey {
		panic(stacktrace.NewError(
			"Kurtosis backend context key '%v' on command '%v' is equal to metrics client close function context key '%v'; this is a bug in Kurtosis!",
			cmd.KurtosisBackendContextKey,
			cmd.CommandStr,
			metricsClientClosingFunctionKey,
		))
	}

	if cmd.EngineClientContextKey == metricsClientClosingFunctionKey {
		panic(stacktrace.NewError(
			"Engine client context key '%v' on command '%v' is equal to metrics client close function context key '%v'; this is a bug in Kurtosis!",
			cmd.EngineClientContextKey,
			cmd.CommandStr,
			metricsClientClosingFunctionKey,
		))
	}

	if cmd.KurtosisBackendContextKey == metricsClientKey {
		panic(stacktrace.NewError(
			"Kurtosis backend context key '%v' on command '%v' is equal to metrics client context key '%v'; this is a bug in Kurtosis!",
			cmd.KurtosisBackendContextKey,
			cmd.CommandStr,
			metricsClientKey,
		))
	}

	if cmd.EngineClientContextKey == metricsClientKey {
		panic(stacktrace.NewError(
			"Engine client context key '%v' on command '%v' is equal to metrics client context key '%v'; this is a bug in Kurtosis!",
			cmd.EngineClientContextKey,
			cmd.CommandStr,
			metricsClientKey,
		))
	}

	lowlevelCmd := &lowlevel.LowlevelKurtosisCommand{
		CommandStr:               cmd.CommandStr,
		ShortDescription:         cmd.ShortDescription,
		LongDescription:          cmd.LongDescription,
		Flags:                    cmd.Flags,
		Args:                     cmd.Args,
		PreValidationAndRunFunc:  cmd.getSetupFunc(),
		RunFunc:                  cmd.getRunFunc(),
		PostValidationAndRunFunc: cmd.getTeardownFunc(),
	}

	return lowlevelCmd.MustGetCobraCommand()
}

func (cmd *EngineConsumingKurtosisCommand) getSetupFunc() func(context.Context) (context.Context, error) {
	return func(ctx context.Context) (context.Context, error) {
		result := ctx

		currentContext, err := store.GetContextsConfigStore().GetCurrentContext()
		if err == nil {
			if store.IsRemote(currentContext) {
				portalManager := portal_manager.NewPortalManager()
				if !portalManager.IsReachable() {
					return nil, stacktrace.NewError("Kurtosis is setup to use the remote context '%s' but Kurtosis "+
						"Portal is unreachable for this context. Make sure Kurtosis Portal is running locally with 'kurtosis "+
						"%s %s' and potentially 'kurtosis %s %s'. If it is, make sure the remote server is running and healthy",
						currentContext.GetName(), command_str_consts.PortalCmdStr, command_str_consts.PortalStatusCmdStr,
						command_str_consts.PortalCmdStr, command_str_consts.PortalStartCmdStr)
				}
				// Forward the remote engine port to the local machine
				portalClient := portalManager.GetClient()
				forwardEnginePortArgs := portal_constructors.NewForwardPortArgs(uint32(kurtosis_context.DefaultGrpcEngineServerPortNum), uint32(kurtosis_context.DefaultGrpcEngineServerPortNum), kurtosis_context.EngineRemoteEndpointType, &kurtosis_context.EnginePortTransportProtocol, &kurtosis_context.ForwardPortWaitUntilReady)
				if _, err := portalClient.ForwardPort(ctx, forwardEnginePortArgs); err != nil {
					return nil, stacktrace.Propagate(err, "Unable to forward the remote engine port to the local machine")
				}
			}
		} else {
			logrus.Warnf("Unable to retrieve current Kurtosis context. This is not critical, it will assume using Kurtosis default context for now.")
		}

		metricsClient, metricsClientCloser, err := metrics_client_factory.GetMetricsClient()
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred while creating metrics client")
		}

		engineManager, err := engine_manager.NewEngineManager(ctx)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting an engine manager.")
		}

		// TODO This is a hack that's only here temporarily because we have commands that use KurtosisBackend directly (they
		//  should not), and EngineConsumingKurtosisCommand therefore needs to provide them with a KurtosisBackend. Once all our
		//  commands only access the Kurtosis APIs, we can remove this.
		kurtosisBackend := engineManager.GetKurtosisBackend()

		dontRestartAPIContainers := false
		engineClient, closeClientFunc, err := engineManager.StartEngineIdempotentlyWithDefaultVersion(ctx, defaults.DefaultEngineLogLevel, defaults.DefaultEngineEnclavePoolSize, defaults.DefaultGitHubAuthTokenOverride, dontRestartAPIContainers, defaults.DefaultDomain, "")
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating a new Kurtosis engine client")
		}
		result = context.WithValue(result, cmd.EngineClientContextKey, engineClient)
		result = context.WithValue(result, engineClientCloseFuncCtxKey, closeClientFunc)
		result = context.WithValue(result, cmd.KurtosisBackendContextKey, kurtosisBackend)
		result = context.WithValue(result, metricsClientKey, metricsClient)
		result = context.WithValue(result, metricsClientClosingFunctionKey, metricsClientCloser)

		return result, nil
	}
}

func (cmd *EngineConsumingKurtosisCommand) getRunFunc() func(context.Context, *flags.ParsedFlags, *args.ParsedArgs) error {
	// Do the gruntwork necessary to give a Kurtosis dev the Docker manager & engine client without them
	// needing to think about how they should get it
	return func(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {
		uncastedEngineClient := ctx.Value(cmd.EngineClientContextKey)
		if uncastedEngineClient == nil {
			return stacktrace.NewError("Expected an engine client to have been stored in the context under key '%v', but none was found; this is a bug in Kurtosis!", cmd.EngineClientContextKey)
		}
		engineClient, ok := uncastedEngineClient.(kurtosis_engine_rpc_api_bindings.EngineServiceClient)
		if !ok {
			return stacktrace.NewError("Found an object that should be the engine client stored in the context under key '%v', but this object wasn't of the correct type", cmd.EngineClientContextKey)
		}

		uncastedKurtosisBackend := ctx.Value(cmd.KurtosisBackendContextKey)
		if uncastedKurtosisBackend == nil {
			return stacktrace.NewError("Expected Kurtosis backend to have been stored in the context under key '%v', but none was found; this is a bug in Kurtosis!", cmd.KurtosisBackendContextKey)
		}
		kurtosisBackend, ok := uncastedKurtosisBackend.(backend_interface.KurtosisBackend)
		if !ok {
			return stacktrace.NewError("Found an object that should be the Kurtosis backend stored in the context under key '%v', but this object wasn't of the correct type", cmd.KurtosisBackendContextKey)
		}

		uncastedMetricsClient := ctx.Value(metricsClientKey)
		if uncastedMetricsClient == nil {
			return stacktrace.NewError("Expected Metrics Client to have been stored in the context under key '%v', but none was found; this is a bug in Kurtosis", metricsClientKey)
		}
		metricsClient, ok := uncastedMetricsClient.(metrics_client.MetricsClient)
		if !ok {
			return stacktrace.NewError("Found an object that should be the metrics client stored in the context under key '%v', but this object wasn't of the correct type", metricsClientKey)
		}

		currentContext, err := store.GetContextsConfigStore().GetCurrentContext()
		if err != nil {
			return stacktrace.Propagate(err, "Error fetching current context information")
		}

		portalManager := portal_manager.NewPortalManager()
		if store.IsRemote(currentContext) && !portalManager.IsReachable() {
			// TODO: add command to start it when it's implemented
			return stacktrace.NewError("Selected context is a remote context but Kurtosis Portal daemon is " +
				"not reachable. Make sure it is started and re-run the command")
		}

		if err = cmd.RunFunc(ctx, kurtosisBackend, engineClient, metricsClient, flags, args); err != nil {
			return stacktrace.Propagate(
				err,
				"An error occurred calling the run function for command '%v'",
				cmd.CommandStr,
			)
		}

		return nil
	}
}

func (cmd *EngineConsumingKurtosisCommand) getTeardownFunc() func(ctx context.Context) {
	return func(ctx context.Context) {
		uncastedEngineClientCloseFunc := ctx.Value(engineClientCloseFuncCtxKey)
		if uncastedEngineClientCloseFunc != nil {
			engineClientCloseFunc, ok := uncastedEngineClientCloseFunc.(func() error)
			if ok {
				if err := engineClientCloseFunc(); err != nil {
					logrus.Warnf("We tried to close the engine client after we're done using it, but doing so threw an error:\n%v", err)
				}
			} else {
				logrus.Errorf("Expected the object at context key '%v' to be an engine client close function, but it wasn't; this is a bug in Kurtosis!", engineClientCloseFuncCtxKey)
			}
		} else {
			logrus.Errorf(
				"Expected to find an engine client close function during teardown at context key '%v', but none was found; this is a bug in Kurtosis!",
				engineClientCloseFuncCtxKey,
			)
		}

		uncastedMetricsClientCloser := ctx.Value(metricsClientClosingFunctionKey)
		if uncastedMetricsClientCloser != nil {
			metricsClientCloser, ok := uncastedMetricsClientCloser.(func() error)
			if ok {
				if err := metricsClientCloser(); err != nil {
					logrus.Warnf("An error occurred while closing the metrics client\n%s", err)
				}
			} else {
				logrus.Errorf("Expected the object at context key '%v' to be a metrics client close function, but it wasn't; this is a bug in Kurtosis!", metricsClientClosingFunctionKey)
			}
		} else {
			logrus.Errorf(
				"Expected to metrics client close function during teardown at context key '%v', but none was found; this is a bug in Kurtosis!",
				metricsClientClosingFunctionKey,
			)
		}
	}
}
