package add

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/cloud"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	cloudhelper "github.com/kurtosis-tech/kurtosis/cli/cli/helpers/cloud"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/metrics_user_id_store"
	api "github.com/kurtosis-tech/kurtosis/cloud/api/golang/kurtosis_backend_server_rpc_api_bindings"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

var AddCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:               command_str_consts.CloudAddCmdStr,
	ShortDescription:         "Create a new Kurtosis Cloud instance",
	LongDescription:          "Create a new remote Kurtosis Cloud instance",
	Flags:                    []*flags.FlagConfig{},
	Args:                     nil,
	PreValidationAndRunFunc:  nil,
	RunFunc:                  run,
	PostValidationAndRunFunc: nil,
}

func run(ctx context.Context, _ *flags.ParsedFlags, _ *args.ParsedArgs) error {
	logrus.Info("Creating a new remote Kurtosis Cloud instance")
	apiKey, err := cloudhelper.LoadApiKey()
	if err != nil {
		return stacktrace.Propagate(err, "Could not load an API Key. Check that it's defined using the "+
			"%s env var and it's a valid (active) key", cloudhelper.KurtosisCloudApiKeyEnvVarArg)
	}

	// Use metrics id for now until we replace with a proper auth'd id:
	metricsUserIdStore := metrics_user_id_store.GetMetricsUserIDStore()
	metricsUserId, err := metricsUserIdStore.GetUserID()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while getting the user's id")
	}

	cloudConfig, err := cloudhelper.GetCloudConfig()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while loading the Cloud Config")
	}
	// Create the connection
	connectionStr := fmt.Sprintf("%s:%d", cloudConfig.ApiUrl, cloudConfig.Port)
	client, err := cloud.CreateCloudClient(connectionStr, cloudConfig.CertificateChain)
	if err != nil {
		return stacktrace.Propagate(err, "Error building client for Kurtosis Cloud")
	}

	getConfigArgs := &api.CreateCloudInstanceConfigArgs{
		ApiKey: *apiKey,
		UserId: metricsUserId,
	}
	result, err := client.CreateCloudInstance(ctx, getConfigArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while calling the Kurtosis Cloud API")
	}

	instanceId := result.GetInstanceId()
	logrus.Infof("Success! The Kurtosis Cloud instance is being created with id: %s", instanceId)
	logrus.Infof("The Kurtosis cloud instance is currently being created and it should take about 2-3 mins to become ready. "+
		"Once ready, load the instance by calling: kurtosis %s %s %s",
		command_str_consts.CloudCmdStr,
		command_str_consts.CloudLoadCmdStr,
		instanceId,
	)
	return nil
}
