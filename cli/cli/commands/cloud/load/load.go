package load

import (
	"context"
	"encoding/base64"
	"fmt"
	api "github.com/kurtosis-tech/kurtosis-cloud-backend/api/golang/kurtosis_backend_server_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/cloud"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/instance_id_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/kurtosis_context/add"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/kurtosis_context/context_switch"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/resolved_config"
	"github.com/kurtosis-tech/kurtosis/contexts-config-store/store"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"os"
)

const (
	instanceIdentifierArgKey      = "instance-id"
	instanceIdentifierArgIsGreedy = false
	kurtosisCloudApiKeyEnvVarArg  = "KURTOSIS_CLOUD_API_KEY"
	// TODO: Move the connection information out into a configuration file. Will happen in future work:
	defaultKurtosisCloudApiUrl  = resolved_config.DefaultCloudConfigApiUrl
	defaultKurtosisCloudApiPort = resolved_config.DefaultCloudConfigPort
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

	apiKey, err := loadApiKey()
	if err != nil {
		return stacktrace.Propagate(err, "Could not load an API Key. Check that it's defined using the "+
			"%s env var and it's a valid (active) key", kurtosisCloudApiKeyEnvVarArg)
	}

	cloudConfig, err := getCloudConfig()
	if err != nil {
		return stacktrace.Propagate(err, "An error occured while loading the Cloud Config")
	}
	// Create the connection
	connectionStr := fmt.Sprintf("%s:%d", cloudConfig.ApiUrl, cloudConfig.Port)
	client, err := cloud.CreateCloudClient(connectionStr, cloudConfig.CertificateChain)
	if err != nil {
		return stacktrace.Propagate(err, "Error building client for Kurtosis Cloud")
	}

	getConfigArgs := &api.GetCloudInstanceConfigArgs{
		ApiKey:     *apiKey,
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

func loadApiKey() (*string, error) {
	apiKey := os.Getenv(kurtosisCloudApiKeyEnvVarArg)
	if len(apiKey) < 1 {
		return nil, stacktrace.NewError("No API Key was found. An API Key must be provided as env var %s", kurtosisCloudApiKeyEnvVarArg)
	}
	logrus.Info("Successfully Loaded API Key...")
	return &apiKey, nil
}

func getCloudConfig() (*resolved_config.KurtosisCloudConfig, error) {
	// Get the configuration
	kurtosisConfigStore := kurtosis_config.GetKurtosisConfigStore()
	configProvider := kurtosis_config.NewKurtosisConfigProvider(kurtosisConfigStore)
	kurtosisConfig, err := configProvider.GetOrInitializeConfig()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get or initialize Kurtosis configuration")
	}
	if kurtosisConfig.GetCloudConfig() == nil {
		return nil, stacktrace.Propagate(err, "No cloud config was found. This is an internal Kurtosis error.")
	}
	cloudConfig := kurtosisConfig.GetCloudConfig()

	if cloudConfig.Port == 0 {
		cloudConfig.Port = resolved_config.DefaultCloudConfigPort
	}
	if len(cloudConfig.ApiUrl) < 1 {
		cloudConfig.ApiUrl = resolved_config.DefaultCloudConfigApiUrl
	}
	if len(cloudConfig.CertificateChain) < 1 {
		cloudConfig.CertificateChain = resolved_config.DefaultCertificateChain
	}

	return cloudConfig, nil
}
