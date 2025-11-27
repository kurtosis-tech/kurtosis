package set

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis-portal/api/golang/constructors"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/highlevel/context_id_arg"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_str_consts"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/defaults"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/helpers/engine_manager"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/helpers/portal_manager"
	"github.com/dzobbe/PoTE-kurtosis/contexts-config-store/store"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"strings"
)

const (
	contextIdentifierArgKey      = "context"
	contextIdentifierArgIsGreedy = false

	dockerDaemonIsNotRunningErrorSubStr     = "Is the docker daemon running?"
	cantConnectToLocalKubernetesErrorSubStr = "connect: connection refused"
)

var ContextSetCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:       command_str_consts.ContextSetCmdStr,
	ShortDescription: "Sets the active Kurtosis context",
	LongDescription: fmt.Sprintf("Sets the active Kurtosis context. The context needs to be added "+
		"first using the `%s` command. When setting a remote context, the connection will be established with "+
		"the remote Kurtosis server. Kurtosis Portal needs to be running for this. If the remote server can't be "+
		"reached, the context will remain unchanged.", command_str_consts.ContextAddCmdStr),
	Flags: []*flags.FlagConfig{},
	Args: []*args.ArgConfig{
		context_id_arg.NewContextIdentifierArg(store.GetContextsConfigStore(), contextIdentifierArgKey, contextIdentifierArgIsGreedy),
	},
	PreValidationAndRunFunc:  nil,
	RunFunc:                  run,
	PostValidationAndRunFunc: nil,
}

func run(ctx context.Context, _ *flags.ParsedFlags, args *args.ParsedArgs) error {
	contextIdentifier, err := args.GetNonGreedyArg(contextIdentifierArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for context identifier arg '%v' but none was found; this is a bug in the Kurtosis CLI!", contextIdentifierArgKey)
	}

	return SetContext(ctx, contextIdentifier)
}

func SetContext(
	ctx context.Context,
	contextIdentifier string,
) error {
	isContextSetSuccessful := false
	logrus.Info("Switching context...")

	contextsConfigStore := store.GetContextsConfigStore()
	contextPriorToSet, err := contextsConfigStore.GetCurrentContext()
	if err != nil {
		return stacktrace.NewError("An error occurred retrieving current context prior to setting to the new one '%s'", contextIdentifier)
	}

	if !store.IsRemote(contextPriorToSet) {
		engineManager, err := engine_manager.NewEngineManager(ctx)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred creating an engine manager.")
		}
		if stopLocalEngineErr := engineManager.StopEngineIdempotently(ctx); stopLocalEngineErr != nil {
			if !isDockerOrKubernetesNotRunningErr(stopLocalEngineErr) {
				return stacktrace.Propagate(stopLocalEngineErr, "An error occurred stopping the local engine. The local engine "+
					"needs to be stopped before the context can be set. The engine status can be obtained running "+
					"kurtosis %s %s and it can be stopped manually by running kurtosis %s %s.",
					command_str_consts.EngineCmdStr, command_str_consts.EngineStatusCmdStr,
					command_str_consts.EngineCmdStr, command_str_consts.EngineStopCmdStr)
			}

		}
	}

	contextsMatchingIdentifiers, err := context_id_arg.GetContextUuidForContextIdentifier(contextsConfigStore, []string{contextIdentifier})
	if err != nil {
		return stacktrace.Propagate(err, "Error searching for context matching context identifier: '%s'", contextIdentifier)
	}
	contextUuidToSet, found := contextsMatchingIdentifiers[contextIdentifier]
	if !found {
		return stacktrace.NewError("No context matching identifier '%s' could be found", contextIdentifier)
	}

	if contextUuidToSet.GetValue() == contextPriorToSet.GetUuid().GetValue() {
		logrus.Infof("Context '%s' already set", contextPriorToSet.GetName())
		return nil
	}

	if err = contextsConfigStore.SetContext(contextUuidToSet); err != nil {
		return stacktrace.Propagate(err, "An error occurred setting context '%s' with UUID '%s'", contextIdentifier, contextUuidToSet.GetValue())
	}
	defer func() {
		if isContextSetSuccessful {
			return
		}
		if err = contextsConfigStore.SetContext(contextPriorToSet.GetUuid()); err != nil {
			logrus.Errorf("An unexpected error occurred setting context '%s' with UUID "+
				"'%s'. Kurtosis tried to roll back to previous context '%s' with UUID '%s' but the roll back "+
				"failed. It is likely that the current context is still set to '%s' but it is not fully functional. "+
				"Try manually reverting to context '%s' to get back to a working state. Error was:\n%v",
				contextIdentifier, contextUuidToSet.GetValue(), contextPriorToSet.GetName(),
				contextPriorToSet.GetUuid(), contextIdentifier, contextPriorToSet.GetName(), err.Error())
		}
	}()

	currentContext, err := contextsConfigStore.GetCurrentContext()
	if err != nil {
		return stacktrace.Propagate(err, "Error retrieving context info for context '%s' after setting it", contextIdentifier)
	}

	portalManager := portal_manager.NewPortalManager()
	if store.IsRemote(currentContext) {
		if err := portalManager.StartRequiredVersion(ctx); err != nil {
			return stacktrace.Propagate(err, "An error occurred starting the portal")
		}
		portalDaemonClient := portalManager.GetClient()
		if portalDaemonClient != nil {
			switchContextArg := constructors.NewSwitchContextArgs()
			if _, err = portalDaemonClient.SwitchContext(ctx, switchContextArg); err != nil {
				return stacktrace.Propagate(err, "Error switching Kurtosis portal context")
			}
		}
	} else {
		// We stop the portal when the user switches back to the local context.
		// We do that to be consistent with the start above.
		// However, the portal is designed to also work with the local context with a client and server
		// running locally.
		if err := portalManager.StopExisting(ctx); err != nil {
			return stacktrace.Propagate(err, "An error occurred stopping Kurtosis Portal")
		}
	}

	logrus.Infof("Context set to '%s', Kurtosis engine will now be restarted", contextIdentifier)

	// Instantiate the engine manager after storing the new context so the manager can read it.
	engineManager, err := engine_manager.NewEngineManager(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating an engine manager for the new context.")
	}

	dontRestartAPIContainers := false
	_, engineClientCloseFunc, startEngineErr := engineManager.StartEngineIdempotentlyWithDefaultVersion(ctx, logrus.InfoLevel, defaults.DefaultEngineEnclavePoolSize, defaults.DefaultGitHubAuthTokenOverride, dontRestartAPIContainers, defaults.DefaultDomain, defaults.DefaultLogRetentionPeriod, defaults.DefaultLogsSinks)
	if startEngineErr != nil {
		logrus.Warnf("The context was successfully set to '%s' but Kurtosis failed to start an engine in "+
			"this new context. A new engine should be started manually with '%s %s %s'. The error was:\n%v",
			contextIdentifier, command_str_consts.KurtosisCmdStr, command_str_consts.EngineCmdStr, command_str_consts.EngineRestartCmdStr, startEngineErr)
	} else {
		defer func() {
			if err = engineClientCloseFunc(); err != nil {
				logrus.Warnf("Error closing connection to the engine running in context '%s':\n'%v'",
					contextIdentifier, err)
			}
		}()
		logrus.Info("Successfully set context")
	}

	isContextSetSuccessful = true
	return nil
}

func isDockerOrKubernetesNotRunningErr(err error) bool {
	rootCauseErrStr := stacktrace.RootCause(err).Error()
	if strings.Contains(rootCauseErrStr, dockerDaemonIsNotRunningErrorSubStr) ||
		strings.Contains(rootCauseErrStr, cantConnectToLocalKubernetesErrorSubStr) {
		return true
	}
	return false
}
