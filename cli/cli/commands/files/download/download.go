package download

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/artifact_identifier_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/enclave_id_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/file_system_path_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/mholt/archiver"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"path/filepath"
)

const (
	enclaveIdentifierArgKey = "enclave"
	isEnclaveIdArgOptional  = false
	isEnclaveIdArgGreedy    = false

	artifactIdentifierArgKey        = "artifact-identifier"
	isArtifactIdentifierArgOptional = false
	isArtifactIdentifierArgGreedy   = false

	destinationPathArgKey        = "destination-path"
	isDestinationPathArgOptional = true
	emptyDestinationPathArg      = ""

	noExtractFlagKey          = "no-extract"
	noExtractFlagDefaultValue = "false"

	filesArtifactExtension                = ".tgz"
	filesArtifactPermission               = 0o744
	filesArtifactDestinationDirPermission = 0o777

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"

	defaultTmpDir = ""
	tmpDirPattern = "tmp-dir-for-download-*"
)

var FilesUploadCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.FilesDownloadCmdStr,
	ShortDescription:          "Download a files artifact from an enclave",
	LongDescription:           "Download the given files artifact from the given enclave to your machine. The files artifact and enclave are specified by identifier (name, UUID, or shortened UUID). Read more about identifiers here: https://docs.kurtosis.com/reference/resource-identifier",
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	Flags: []*flags.FlagConfig{
		{
			Key:     noExtractFlagKey,
			Usage:   "If true then the file won't be extracted. Default false.",
			Type:    flags.FlagType_Bool,
			Default: noExtractFlagDefaultValue,
		},
	},
	Args: []*args.ArgConfig{
		enclave_id_arg.NewEnclaveIdentifierArg(
			enclaveIdentifierArgKey,
			engineClientCtxKey,
			isEnclaveIdArgOptional,
			isEnclaveIdArgGreedy,
		),
		artifact_identifier_arg.NewArtifactIdentifierArg(
			artifactIdentifierArgKey,
			enclaveIdentifierArgKey,
			isArtifactIdentifierArgOptional,
			isArtifactIdentifierArgGreedy,
		),
		file_system_path_arg.NewDirpathArg(
			destinationPathArgKey,
			isDestinationPathArgOptional,
			emptyDestinationPathArg,
			file_system_path_arg.BypassDefaultValidationFunc,
		),
	},
	RunFunc: run,
}

func run(
	ctx context.Context,
	_ backend_interface.KurtosisBackend,
	_ kurtosis_engine_rpc_api_bindings.EngineServiceClient,
	_ metrics_client.MetricsClient,
	flags *flags.ParsedFlags,
	args *args.ParsedArgs,
) error {
	enclaveIdentifier, err := args.GetNonGreedyArg(enclaveIdentifierArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave ID using key '%v'", enclaveIdentifierArgKey)
	}

	artifactIdentifier, err := args.GetNonGreedyArg(artifactIdentifierArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the artifact identifier to download using key '%v'", artifactIdentifier)
	}

	destinationPath, err := args.GetNonGreedyArg(destinationPathArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the destination path to write downloaded file to using key '%v'", destinationPath)
	}
	if destinationPath == "" {
		destinationPath = fmt.Sprintf("./%v", artifactIdentifier)
	}
	absoluteDestinationPath, err := filepath.Abs(destinationPath)
	if err != nil {
		return stacktrace.NewError("An error occurred while getting absolute path for the passed destination path '%v'", destinationPath)
	}
	// check if the path doesn't exist, we create a directory if it doesn't
	// we have already checked if the passed path isn't a dir so we don't do it again
	_, err = os.Stat(absoluteDestinationPath)
	if os.IsNotExist(err) {
		err = os.Mkdir(absoluteDestinationPath, filesArtifactDestinationDirPermission)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while creating the target dir")
		}
	}

	shouldNotExtract, err := flags.GetBool(noExtractFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while getting the default value for the '%v' flag", noExtractFlagKey)
	}

	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred connecting to the local Kurtosis engine")
	}
	enclaveCtx, err := kurtosisCtx.GetEnclaveContext(ctx, enclaveIdentifier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave context for enclave '%v'", enclaveIdentifier)
	}

	artifactBytes, err := enclaveCtx.DownloadFilesArtifact(ctx, artifactIdentifier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred downloading files with identifier '%v' from enclave '%v'", artifactIdentifier, enclaveIdentifier)
	}

	fileNameToWriteTo := fmt.Sprintf("%v%v", artifactIdentifier, filesArtifactExtension)
	destinationPathToDownloadFileTo := path.Join(absoluteDestinationPath, fileNameToWriteTo)

	// if the user doesn't want to extract, we just download and return
	if shouldNotExtract {
		err = os.WriteFile(destinationPathToDownloadFileTo, artifactBytes, filesArtifactPermission)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while writing bytes to file '%v' with permission '%v'", destinationPathToDownloadFileTo, filesArtifactPermission)
		}
		logrus.Infof("File package with identifier '%v' downloaded to '%v'", artifactIdentifier, destinationPathToDownloadFileTo)
		return nil
	}

	tmpDirPath, err := os.MkdirTemp(defaultTmpDir, tmpDirPattern)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while creating a temporary directory to download the files artifact with identifier '%v' to", artifactIdentifier)
	}
	shouldCleanupTmpDir := false
	defer func() {
		if shouldCleanupTmpDir {
			os.RemoveAll(tmpDirPath)
		}
	}()
	tmpFileToWriteTo := path.Join(tmpDirPath, fileNameToWriteTo)
	err = os.WriteFile(tmpFileToWriteTo, artifactBytes, filesArtifactPermission)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while writing bytes to file '%v' with permission '%v'", tmpDirPath, filesArtifactPermission)
	}
	err = archiver.Unarchive(tmpFileToWriteTo, absoluteDestinationPath)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while extracting '%v' to '%v'", tmpFileToWriteTo, destinationPathToDownloadFileTo)
	}
	logrus.Infof("File package with identifier '%v' extracted to '%v'", artifactIdentifier, absoluteDestinationPath)

	shouldCleanupTmpDir = true
	return nil
}
