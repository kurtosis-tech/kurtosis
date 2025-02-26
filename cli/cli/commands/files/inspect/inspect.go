package inspect

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/artifact_identifier_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/enclave_id_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/out"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/xlab/treeprint"
	"golang.org/x/exp/slices"
)

const (
	enclaveIdentifierArgKey = "enclave"
	isEnclaveIdArgOptional  = false
	isEnclaveIdArgGreedy    = false

	artifactIdentifierArgKey        = "artifact-identifier"
	isArtifactIdentifierArgOptional = false
	isArtifactIdentifierArgGreedy   = false

	filePathArgKey        = "file-path"
	emptyFilePath         = ""
	isFilePathArgOptional = true
	isFilePathArgGreedy   = false

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"
	emptyFileStr          = ""
	rootLevelFileStr      = ""
	byteGroup             = 1024
)

var sizeSuffix = []byte{'K', 'M', 'G', 'T', 'P'}

var FilesInspectCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.FilesInspectCmdStr,
	ShortDescription:          "Inspect files of an enclave",
	LongDescription:           "Inspect the requested file artifact, returning the file tree, metadata and a preview, if available",
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	Flags:                     []*flags.FlagConfig{},
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
		{
			Key:                   filePathArgKey,
			IsOptional:            isFilePathArgOptional,
			IsGreedy:              isFilePathArgGreedy,
			DefaultValue:          emptyFilePath,
			ArgCompletionProvider: args.NewManualCompletionsProvider(getCompletionFunc(enclaveIdentifierArgKey, artifactIdentifierArgKey)),
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
		out.PrintErrLn(fmt.Sprintf("Artifact '%v' contents:\n", artifactIdentifierName))
		out.PrintOutLn(buildTree(fileDescriptions))
		return nil
	}
	index := slices.IndexFunc(fileDescriptions, func(desc *kurtosis_core_rpc_api_bindings.FileArtifactContentsFileDescription) bool {
		return desc.GetPath() == filePath
	})
	if index == -1 {
		return stacktrace.NewError("An error finding file '%v' on artifact identifier '%v', from '%v'", filePath, artifactIdentifierName, enclaveIdentifier)
	}
	out.PrintErrLn("File contents:")
	out.PrintOutLn(fileDescriptions[index].GetTextPreview())
	return nil
}

// This structure helps assemble a file tree compatible with treeprint lib
type treeMap struct {
	internalMap map[string]*treeMap
	subtree     treeprint.Tree
}

func (nm *treeMap) addBranchIfNotPresent(s string) *treeMap {
	if value, ok := nm.internalMap[s]; ok {
		return value
	}
	nm.internalMap[s] = &treeMap{
		map[string]*treeMap{},
		nm.subtree.AddBranch(s),
	}
	return nm.internalMap[s]
}

func (nm *treeMap) addNodeIfNotPresent(s string) *treeMap {
	if value, ok := nm.internalMap[s]; ok {
		return value
	}
	nm.internalMap[s] = &treeMap{
		map[string]*treeMap{},
		nm.subtree.AddNode(s),
	}
	return nm.internalMap[s]
}

// Assembles a file tree string
func buildTree(fileDescritions []*kurtosis_core_rpc_api_bindings.FileArtifactContentsFileDescription) string {
	tree := treeprint.NewWithRoot("")
	tMap := &treeMap{map[string]*treeMap{}, tree}
	for _, fileDescription := range fileDescritions {
		dir, file := filepath.Split(fileDescription.GetPath())
		curTree := tMap
		if dir != rootLevelFileStr {
			subdirs := strings.Split(filepath.Clean(dir), string(filepath.Separator))
			for _, subdir := range subdirs {
				curTree = curTree.addBranchIfNotPresent(color.CyanString(subdir))
			}
		}
		if file != emptyFileStr {
			curTree.addNodeIfNotPresent(fmt.Sprintf("%v [%s]", file, humanReadableSize(fileDescription.GetSize())))
		}
	}
	return tree.String()
}

func humanReadableSize(size uint64) string {
	if size < byteGroup {
		return fmt.Sprintf("%4d", size)
	}
	suffixIdx := 0
	fSize := float64(size) / byteGroup
	for ; fSize >= byteGroup; suffixIdx++ {
		fSize /= byteGroup
	}
	return fmt.Sprintf("%3.1f%c", fSize, sizeSuffix[suffixIdx])
}

func getCompletionFunc(enclaveArgKey string, artifactArgKey string) func(ctx context.Context, _ *flags.ParsedFlags, previousArgs *args.ParsedArgs) ([]string, error) {
	return func(ctx context.Context, _ *flags.ParsedFlags, previousArgs *args.ParsedArgs) ([]string, error) {
		kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
		if err != nil {
			return nil, stacktrace.Propagate(
				err,
				"An error occurred connecting to the Kurtosis engine for retrieving the names for tab completion",
			)
		}

		enclaveIdentifier, err := previousArgs.GetNonGreedyArg(enclaveArgKey)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting the enclave ID using key '%v'", enclaveArgKey)
		}
		enclave, err := kurtosisCtx.GetEnclaveContext(ctx, enclaveIdentifier)
		if err != nil {
			return nil, stacktrace.Propagate(
				err,
				"An error occurred getting the enclave identifiers",
			)
		}

		artifactIdentifier, err := previousArgs.GetNonGreedyArg(artifactArgKey)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting the artifact ID using key '%v'", artifactArgKey)
		}
		fileArtifactContents, err := enclave.InspectFilesArtifact(ctx, services.FileArtifactName(artifactIdentifier))
		if err != nil {
			return nil, stacktrace.Propagate(
				err,
				"An error occurred getting the file artifacts",
			)
		}
		fileArtifactContentPaths := []string{}
		for _, fileArtifactDescription := range fileArtifactContents.GetFileDescriptions() {
			fileArtifactContentPath := fileArtifactDescription.GetPath()
			if fileArtifactDescription.TextPreview != nil {
				fileArtifactContentPaths = append(fileArtifactContentPaths, fileArtifactContentPath)
			}
		}

		return fileArtifactContentPaths, nil
	}
}
