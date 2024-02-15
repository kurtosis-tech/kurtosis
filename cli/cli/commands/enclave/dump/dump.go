package dump

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/enclave_id_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/file_system_path_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/files"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"os"
	"path"
)

const (
	enclaveIdentifierArgKey = "enclave"
	isEnclaveIdArgOptional  = false
	isEnclaveIdArgGreedy    = false

	outputDirpathArg = "output-dirpath"

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"

	defaultEnclaveDumpDir = "kurtosis-dump"
	enclaveDumpSeparator  = "--"
	outputDirIsOptional   = true

	filesArtifactDestinationDirPermission = 0o777
	filesArtifactFolderName               = "files"
)

var EnclaveDumpCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.EnclaveDumpCmdStr,
	ShortDescription:          "Dumps information about an enclave to disk",
	LongDescription:           "Dumps all information about the enclave to the given directory",
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	Flags:                     nil,
	Args: []*args.ArgConfig{
		enclave_id_arg.NewEnclaveIdentifierArg(
			enclaveIdentifierArgKey,
			engineClientCtxKey,
			isEnclaveIdArgOptional,
			isEnclaveIdArgGreedy,
		),
		file_system_path_arg.NewDirpathArg(
			outputDirpathArg,
			outputDirIsOptional,
			defaultEnclaveDumpDir,
			file_system_path_arg.BypassDefaultValidationFunc,
		),
	},
	RunFunc: run,
}

func run(
	ctx context.Context,
	kurtosisBackend backend_interface.KurtosisBackend,
	_ kurtosis_engine_rpc_api_bindings.EngineServiceClient,
	_ metrics_client.MetricsClient,
	_ *flags.ParsedFlags,
	args *args.ParsedArgs,
) error {
	enclaveIdentifier, err := args.GetNonGreedyArg(enclaveIdentifierArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting enclave identifier using arg key '%v'", enclaveIdentifierArgKey)
	}
	enclaveOutputDirpath, err := args.GetNonGreedyArg(outputDirpathArg)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting output dirpath using arg key '%v'", outputDirpathArg)
	}

	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating Kurtosis Context from local engine")
	}

	enclaveInfo, err := kurtosisCtx.GetEnclave(ctx, enclaveIdentifier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while getting enclave for identifier '%v'", enclaveIdentifier)
	}

	enclaveUuid := enclaveInfo.GetEnclaveUuid()
	if enclaveOutputDirpath == defaultEnclaveDumpDir {
		enclaveName := enclaveInfo.GetName()
		enclaveOutputDirpath = fmt.Sprintf("%s%s%s", enclaveName, enclaveDumpSeparator, enclaveUuid)
	}

	if err = kurtosisBackend.DumpEnclave(ctx, enclave.EnclaveUUID(enclaveUuid), enclaveOutputDirpath); err != nil {
		return stacktrace.Propagate(err, "An error occurred dumping enclave '%v' to '%v'", enclaveIdentifier, enclaveOutputDirpath)
	}

	if enclaveInfo.ApiContainerStatus != kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus_EnclaveAPIContainerStatus_RUNNING {
		logrus.Debugf("Couldn't dump file information as the enclave '%v' is not running", enclaveIdentifier)
		logrus.Infof("Dumped enclave '%v' to directory '%v'", enclaveIdentifier, enclaveOutputDirpath)
		return nil
	}

	enclaveCtx, err := kurtosisCtx.GetEnclaveContext(ctx, enclaveIdentifier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while retrieving enclave context for enclave with identifier '%v'", enclaveIdentifier)
	}

	filesInEnclave, err := enclaveCtx.GetAllFilesArtifactNamesAndUuids(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while fetching files artifact in enclave '%v'", enclaveIdentifier)
	}

	if len(filesInEnclave) == 0 {
		logrus.Infof("Dumped enclave '%v' to directory '%v'", enclaveIdentifier, enclaveOutputDirpath)
		return nil
	}

	filesDownloadFolder := path.Join(enclaveOutputDirpath, filesArtifactFolderName)
	if err = os.Mkdir(filesDownloadFolder, filesArtifactDestinationDirPermission); err != nil {
		return stacktrace.Propagate(err, "An error occurred while creating a folder '%v' to download files to", filesArtifactFolderName)
	}

	for _, fileNameAndUuid := range filesInEnclave {
		fileDownloadPath := path.Join(filesDownloadFolder, fileNameAndUuid.GetFileName())
		if err = os.Mkdir(fileDownloadPath, filesArtifactDestinationDirPermission); err != nil {
			return stacktrace.Propagate(err, "An error occurred while creating directory '%v' to write files artifact '%v'", fileDownloadPath, fileNameAndUuid.GetFileName())
		}
		if err = files.DownloadAndExtractFilesArtifact(ctx, enclaveCtx, fileNameAndUuid.GetFileName(), fileDownloadPath); err != nil {
			return stacktrace.Propagate(err, "An error occurred while downloading and extracting file '%v'", fileNameAndUuid.GetFileName())
		}
	}

	logrus.Infof("Dumped enclave '%v' to directory '%v'", enclaveIdentifier, enclaveOutputDirpath)
	return nil
}
