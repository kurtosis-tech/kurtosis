package upload

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
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

const (
	enclaveIdentifierArgKey = "enclave-identifier"
	isEnclaveIdArgOptional  = false
	isEnclaveIdArgGreedy    = false

	pathArgKey = "path"

	nameFlagKey = "name"
	defaultName = ""

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"

	artifactNamePrefix = "cli-uploaded-artifact-%v"
)

var FilesUploadCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.FilesUploadCmdStr,
	ShortDescription:          "Uploads files to an enclave",
	LongDescription:           "Uploads the requested files to the enclave so they can be used by services within the enclave",
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
			Key:            pathArgKey,
			ValidationFunc: validatePathArg,
		},
	},
	RunFunc: run,
}

func run(
	ctx context.Context,
	kurtosisBackend backend_interface.KurtosisBackend,
	engineClient kurtosis_engine_rpc_api_bindings.EngineServiceClient,
	flags *flags.ParsedFlags,
	args *args.ParsedArgs,
) error {
	enclaveIdentifier, err := args.GetNonGreedyArg(enclaveIdentifierArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave ID using key '%v'", enclaveIdentifierArgKey)
	}

	path, err := args.GetNonGreedyArg(pathArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the path to upload using key '%v'", pathArgKey)
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

	filesArtifactUuid, err := enclaveCtx.UploadFiles(path, artifactName)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred uploading files at path '%v' to enclave '%v'", path, enclaveIdentifier)
	}
	logrus.Infof("Files package '%v' uploaded with UUID: %v", artifactName, filesArtifactUuid)
	return nil
}

func validatePathArg(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {
	path, err := args.GetNonGreedyArg(pathArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the path to validate using key '%v'", pathArgKey)
	}

	if _, err := os.Stat(path); err != nil {
		return stacktrace.Propagate(err, "An error occurred verifying path '%v' exists and is readable", path)
	}
	return nil
}
