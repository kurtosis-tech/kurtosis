package load

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/cloud"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/instance_id_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/kurtosis_context/add"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/kurtosis_context/set"
	cloudhelper "github.com/kurtosis-tech/kurtosis/cli/cli/helpers/cloud"
	api "github.com/kurtosis-tech/kurtosis/cloud/api/golang/kurtosis_backend_server_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/contexts-config-store/store"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	instanceIdentifierArgKey      = "instance-id"
	instanceIdentifierArgIsGreedy = false

	awsAccessKeyIdEnv      = "AWS_ACCESS_KEY_ID"
	awsSecretAccessKeyEnv  = "AWS_SECRET_ACCESS_KEY"
	awsBucketRegionEnv     = "AWS_BUCKET_REGION"
	awsBucketNameEnv       = "AWS_BUCKET_NAME"
	awsBucketUserFolderEnv = "AWS_BUCKET_USER_FOLDER"
)

var LoadCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:       command_str_consts.CloudLoadCmdStr,
	ShortDescription: "Load a Kurtosis Cloud instance",
	LongDescription: "Load a remote Kurtosis Cloud instance by providing the instance id." +
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

	contextsConfigStore := store.GetContextsConfigStore()
	currentContext, err := contextsConfigStore.GetCurrentContext()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while retrieving the current context")
	}

	if currentContext.Uuid.Value == instanceID {
		logrus.Infof("Cloud instance %s already loaded", instanceID)
		return nil
	}

	logrus.Infof("Loading cloud instance %s", instanceID)

	apiKey, err := cloudhelper.LoadApiKey()
	if err != nil {
		return stacktrace.Propagate(err, "Could not load an API Key. Check that it's defined using the "+
			"%s env var and it's a valid (active) key", cloudhelper.KurtosisCloudApiKeyEnvVarArg)
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

	getConfigArgs := &api.GetCloudInstanceConfigArgs{
		ApiKey:          apiKey,
		InstanceId:      &instanceID,
		InstanceShortId: nil,
	}
	result, err := client.GetCloudInstanceConfig(ctx, getConfigArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while calling the Kurtosis Cloud API")
	}

	// TODO: Create shared enums for instance states:
	if result.Status != "running" {
		logrus.Warnf("The Kurtosis Cloud instance is in state \"%s\" and cannot currently be loaded."+
			" Instance needs to be in state \"running\"", result.Status)
		return nil
	}

	decodedConfigBytes, err := base64.StdEncoding.DecodeString(result.ContextConfig)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to base64 decode context config")
	}

	parsedContext, err := add.ParseContextData(decodedConfigBytes)
	if err != nil {
		return stacktrace.Propagate(err, "Unable to decode context config")
	}

	// We first have to remove the context incase it's already loaded
	err = contextsConfigStore.RemoveContext(parsedContext.Uuid)
	if err != nil {
		return stacktrace.Propagate(err, "While attempting to reload the context with uuid %s an error occurred while removing it from the context store", parsedContext.Uuid)
	}
	if add.AddContext(parsedContext, assembleEnvVars(result), result.InstanceId, result.UserId) != nil {
		return stacktrace.Propagate(err, "Unable to add context to context store")
	}
	contextIdentifier := parsedContext.GetName()
	return set.SetContext(ctx, contextIdentifier)
}

func assembleEnvVars(cloudInstanceConfig *api.GetCloudInstanceConfigResponse) *string {
	var envVars string
	if cloudInstanceConfig.GetAwsEnvironment() != nil {
		awsEnvironment := cloudInstanceConfig.GetAwsEnvironment()
		envVars = fmt.Sprintf("{%q: %q, %q: %q, %q: %q, %q: %q, %q: %q}",
			awsAccessKeyIdEnv, awsEnvironment.GetAccessKeyId(),
			awsSecretAccessKeyEnv, awsEnvironment.GetSecretAccessKey(),
			awsBucketRegionEnv, awsEnvironment.GetBucketRegion(),
			awsBucketNameEnv, awsEnvironment.GetBucketName(),
			awsBucketUserFolderEnv, awsEnvironment.GetBucketUserFolder(),
		)
	} else if cloudInstanceConfig.GetUserKey() != nil {
		// TODO: for backward compatibility - some components are computed live as they are constant. it should be
		//  removed in favor of the above once the cloud BE is upgraded
		awsUserKey := cloudInstanceConfig.GetUserKey()
		envVars = fmt.Sprintf("{%q: %q, %q: %q, %q: %q, %q: %q, %q: %q}",
			awsAccessKeyIdEnv, awsUserKey.GetId(),
			awsSecretAccessKeyEnv, awsUserKey.GetSecret(),
			awsBucketRegionEnv, "us-east-1",
			awsBucketNameEnv, "kurtosis-cloud",
			awsBucketUserFolderEnv, fmt.Sprintf("cloud-user-%s", cloudInstanceConfig.GetUserId()),
		)
	}
	return &envVars
}
