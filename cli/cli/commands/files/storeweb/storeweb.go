package storeweb

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/enclave_id_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	enclaveIdentifierArgKey = "enclave"
	isEnclaveIdArgOptional  = false
	isEnclaveIdArgGreedy    = false

	urlArgKey = "url"

	nameFlagKey = "name"
	defaultName = ""

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"

	artifactNamePrefix = "cli-stored-artifact-%v"
)

// TODO Maybe, instead of having 'storeweb' and 'storeservice' we could just have a flag that switches between the two??
var FilesStoreWebCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:       command_str_consts.FilesStoreWebCmdStr,
	ShortDescription: "Downloads files from a URL",
	LongDescription: fmt.Sprintf(
		"Instructs Kurtosis to download an archive file from the given URL and store it in the "+
			"enclave for later use (e.g. with '%v %v')",
		command_str_consts.ServiceCmdStr,
		command_str_consts.ServiceAddCmdStr,
	),
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	Flags: []*flags.FlagConfig{
		{
			Key:     nameFlagKey,
			Usage:   "The name to be given to the produced of the artifact, auto generated if not passed",
			Type:    flags.FlagType_String,
			Default: defaultName,
		},
	},
	Args: []*args.ArgConfig{
		enclave_id_arg.NewEnclaveIdentifierArg(
			enclaveIdentifierArgKey,
			engineClientCtxKey,
			isEnclaveIdArgOptional,
			isEnclaveIdArgGreedy,
		),
		{
			Key: urlArgKey,
		},
	},
	RunFunc: run,
}

func run(
	ctx context.Context,
	kurtosisBackend backend_interface.KurtosisBackend,
	engineClient kurtosis_engine_rpc_api_bindings.EngineServiceClient,
	_ metrics_client.MetricsClient,
	flags *flags.ParsedFlags,
	args *args.ParsedArgs,
) error {
	enclaveIdentifier, err := args.GetNonGreedyArg(enclaveIdentifierArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave ID using key '%v'", enclaveIdentifierArgKey)
	}

	url, err := args.GetNonGreedyArg(urlArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the URL to download from using key '%v'", urlArgKey)
	}

	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred connecting to the local Kurtosis engine")
	}
	enclaveCtx, err := kurtosisCtx.GetEnclaveContext(ctx, enclaveIdentifier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave context for enclave '%v'", enclaveIdentifier)
	}

	artifactName, err := flags.GetString(nameFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the name to be given to the produced artifact")
	}

	if artifactName == defaultName {
		artifactName = fmt.Sprintf(artifactNamePrefix, time.Now().Unix())
	}
	filesArtifactUuid, err := enclaveCtx.StoreWebFiles(ctx, url, artifactName)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred downloading the archive file from URL '%v' to enclave '%v'",
			url,
			enclaveIdentifier,
		)
	}
	logrus.Infof("Files package '%v' stored with UUID: %v", artifactName, filesArtifactUuid)
	return nil
}
