package inspect

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/enclave_id_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	metrics_client "github.com/kurtosis-tech/metrics-library/golang/lib/client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/xlab/treeprint"
	"golang.org/x/exp/slices"
	"path/filepath"
	"strings"
)

const (
	enclaveIdentifierArgKey = "enclave"
	isEnclaveIdArgOptional  = false
	isEnclaveIdArgGreedy    = false

	artifactIdentifierArgKey        = "artifact-identifier"
	emptyArtifactIdentifier         = ""
	isArtifactIdentifierArgOptional = false
	isArtifactIdentifierArgGreedy   = false

	filePathArgKey        = "file-path"
	emptyFilePath         = ""
	isFilePathArgOptional = true
	isFilePathArgGreedy   = false

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"
)

var FilesInspectCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.FilesInspectCmdStr,
	ShortDescription:          "Inspect files of an enclave",
	LongDescription:           "Inspect the requested file artifact, returning the file tree, metadata and a preview, if avaliable",
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	Flags:                     []*flags.FlagConfig{
		//{
		//	Key:     nameFlagKey,
		//	Usage:   "The name to be given to the produced of the artifact, auto generated if not passed",
		//	Type:    flags.FlagType_String,
		//	Default: defaultName,
		//},
	},
	Args: []*args.ArgConfig{
		enclave_id_arg.NewEnclaveIdentifierArg(
			enclaveIdentifierArgKey,
			engineClientCtxKey,
			isEnclaveIdArgOptional,
			isEnclaveIdArgGreedy,
		),
		{
			Key:                   artifactIdentifierArgKey,
			ValidationFunc:        validateArtifactIdentifier,
			IsOptional:            isArtifactIdentifierArgOptional,
			IsGreedy:              isArtifactIdentifierArgGreedy,
			DefaultValue:          nil,
			ArgCompletionProvider: nil,
		},
		{
			Key:                   filePathArgKey,
			IsOptional:            isFilePathArgOptional,
			IsGreedy:              isFilePathArgGreedy,
			DefaultValue:          emptyFilePath,
			ArgCompletionProvider: nil,
		},
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

	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred connecting to the local Kurtosis engine")
	}

	enclaveCtx, err := kurtosisCtx.GetEnclaveContext(ctx, enclaveIdentifier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave context for enclave '%v'", enclaveIdentifier)
	}

	artifactIdentifierName, err := args.GetNonGreedyArg(artifactIdentifierArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the artifact identifier")
	}

	filePath, err := args.GetNonGreedyArg("file-path")
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the file path")
	}

	filesInspectResponse, err := enclaveCtx.InspectFilesArtifact(ctx, services.FileArtifactName(artifactIdentifierName))
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred inspecting files from artifact identifier '%v', enclave '%v'", artifactIdentifierName, enclaveIdentifier)
	}
	fileDescriptions := filesInspectResponse.GetFileDescriptions()

	if filePath == "" {
		logrus.Infof("Artifact '%v' contents:\n%v", artifactIdentifierName, buildTree(artifactIdentifierName, fileDescriptions))
		return nil
	}
	index := slices.IndexFunc(fileDescriptions, func(desc *kurtosis_core_rpc_api_bindings.FileArtifactContentsFileDescription) bool {
		return desc.GetPath() == filePath
	})
	if index == -1 {
		return stacktrace.NewError("An error finding file '%v' on artifact identifier '%v', from '%v'", filePath, artifactIdentifierName, enclaveIdentifier)
	}
	logrus.Infof("File contents:\n%v", fileDescriptions[index].GetTextPreview())
	return nil
}

func buildTree(artifactIdentifierName string, fileDescritions []*kurtosis_core_rpc_api_bindings.FileArtifactContentsFileDescription) string {
	// to add a custom root name use `treeprint.NewWithRoot()` instead
	tree := treeprint.NewWithRoot(artifactIdentifierName)
	for _, fileDescription := range fileDescritions {
		dir, file := filepath.Split(fileDescription.GetPath())
		subdirs := filepath.SplitList(dir)
		cp := tree
		for _, subdir := range subdirs {
			cp = cp.AddBranch(subdir)
		}
		if file != "" {
			if fileDescription.GetDescription() == "" {
				cp.AddNode(file)
			} else {
				cp.AddMetaNode(fileDescription.GetDescription(), file)
			}
		}
	}
	return tree.String()
}

// TODO(vcolombo): Dedupe this
func validateArtifactIdentifier(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {
	artifactIdentifier, err := args.GetNonGreedyArg(artifactIdentifierArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the identifier to validate using key '%v'", artifactIdentifier)
	}

	if strings.TrimSpace(artifactIdentifier) == emptyArtifactIdentifier {
		return stacktrace.NewError("Artifact identifier cannot be an empty string")
	}

	return nil
}
