package load

import (
	"context"
	"encoding/base64"
	api "github.com/kurtosis-tech/kurtosis-cloud-backend/api/golang/kurtosis_backend_server_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/cloud"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/instance_id_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/kurtosis_context/add"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/kurtosis_context/context_switch"
	"github.com/kurtosis-tech/kurtosis/contexts-config-store/store"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"os"
)

const (
	instanceIdentifierArgKey      = "instance-id"
	instanceIdentifierArgIsGreedy = false
)

var LoadCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:       command_str_consts.CloudLoadCmdStr,
	ShortDescription: "Load a Kurtosis Cloud instance",
	LongDescription: "Load a remote Kurtosis Cloud instance into the local context by providing the instance id." +
		"Note, the remote instance must be in a running state for this operation to complete successfully",
	Flags: []*flags.FlagConfig{},
	Args: []*args.ArgConfig{
		instance_id_arg.InstanceIdentifierArg(instanceIdentifierArgKey, instanceIdentifierArgIsGreedy),
	},
	PreValidationAndRunFunc:  nil,
	RunFunc:                  run,
	PostValidationAndRunFunc: nil,
}

func run(ctx context.Context, _ *flags.ParsedFlags, args *args.ParsedArgs) error {
	instanceID, err := args.GetNonGreedyArg(instanceIdentifierArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for instance id arg '%v' but none was found; "+
			"this is a bug in the Kurtosis CLI!", instanceIdentifierArgKey)
	}
	logrus.Infof("Loading cloud instance %s", instanceID)

	// TODO: READ from some settings
	connectionStr := "localhost:8080"
	ApiKeyArg := "KURTOSIS_CLOUD_API_KEY"
	apiKey := os.Getenv(ApiKeyArg)
	if len(apiKey) < 1 {
		return stacktrace.NewError("No API Key was found. An API Key must be provided as env var %s", ApiKeyArg)
	}
	logrus.Info("Loaded API Key...")

	client, err := cloud.CreateCloudClient(connectionStr)
	if err != nil {
		return stacktrace.Propagate(err, "Error building client for Kurtosis Cloud")
	}

	getConfigArgs := &api.GetCloudInstanceConfigArgs{
		ApiKey:     apiKey,
		InstanceId: instanceID,
	}
	result, err := client.GetCloudInstanceConfig(ctx, getConfigArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while calling the Kurtosis Cloud API")
	}
	decodedConfigBytes, err := base64.StdEncoding.DecodeString(result.ContextConfig)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to base64 decode context config")
	}

	parsedContext, err := add.ParseContextData(decodedConfigBytes)
	if err != nil {
		return stacktrace.Propagate(err, "Unable to decode context config")
	}

	contextsConfigStore := store.GetContextsConfigStore()
	// We first have to remove the context incase it's already loaded
	err = contextsConfigStore.RemoveContext(parsedContext.Uuid)
	if err != nil {
		return stacktrace.Propagate(err, "While attempting to reload the context with uuid %s an error occurred while removing it from the context store", parsedContext.Uuid)
	}
	err = add.AddContext(parsedContext)
	if err != nil {
		return stacktrace.Propagate(err, "Unable to add context to context store")
	}
	contextIdentifier := parsedContext.GetName()
	return context_switch.SwitchContext(ctx, contextIdentifier)
}
