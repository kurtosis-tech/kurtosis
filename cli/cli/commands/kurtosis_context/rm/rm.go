package rm

import (
	"context"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/highlevel/context_id_arg"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_str_consts"
	"github.com/dzobbe/PoTE-kurtosis/contexts-config-store/store"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	contextIdentifiersArgKey      = "context"
	contextIdentifiersArgIsGreedy = true
)

var ContextRmCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:       command_str_consts.ContextRmCmdStr,
	ShortDescription: "Removes a Kurtosis contexts",
	LongDescription:  "Removes a Kurtosis context currently configured for this installation",
	Flags:            []*flags.FlagConfig{},
	Args: []*args.ArgConfig{
		context_id_arg.NewContextIdentifierArg(store.GetContextsConfigStore(), contextIdentifiersArgKey, contextIdentifiersArgIsGreedy),
	},
	PreValidationAndRunFunc:  nil,
	RunFunc:                  run,
	PostValidationAndRunFunc: nil,
}

func run(_ context.Context, _ *flags.ParsedFlags, args *args.ParsedArgs) error {
	contextIdentifiers, err := args.GetGreedyArg(contextIdentifiersArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for greedy context identifiers arg '%v' but none was found; this is a bug in the Kurtosis CLI!", contextIdentifiersArgKey)
	}

	contextsConfigStore := store.GetContextsConfigStore()
	contextUuids, err := context_id_arg.GetContextUuidForContextIdentifier(contextsConfigStore, contextIdentifiers)
	if err != nil {
		return stacktrace.Propagate(err, "Error finding contexts matching the provided identifiers")
	}

	logrus.Info("Removing contexts...")
	successfullyDeleted := 0
	for _, contextUuid := range contextUuids {
		if err = contextsConfigStore.RemoveContext(contextUuid); err != nil {
			logrus.Errorf("Error removing context with UUID: '%s'. Error was: \n%v", contextUuid.GetValue(), err.Error())
		} else {
			successfullyDeleted += 1
		}
	}
	if successfullyDeleted != len(contextUuids) {
		return stacktrace.NewError("Some contexts could not be removed. See logs above.")
	}
	logrus.Info("Contexts successfully removed")
	return nil
}
