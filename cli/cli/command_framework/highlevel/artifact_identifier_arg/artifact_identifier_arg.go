package artifact_identifier_arg

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/stacktrace"
	"strings"
)

const (
	emptyArtifactIdentifier = ""
)

func NewArtifactIdentifierArg(
	argKey string,
	enclaveArgKey string,
	isOptional bool,
	isGreedy bool,
) *args.ArgConfig {
	return &args.ArgConfig{
		Key:                   argKey,
		IsOptional:            isOptional,
		DefaultValue:          emptyArtifactIdentifier,
		IsGreedy:              isGreedy,
		ArgCompletionProvider: args.NewManualCompletionsProvider(getCompletionFunc(enclaveArgKey)),
		ValidationFunc:        getValidationFunc(argKey),
	}
}

func getValidationFunc(artifactIdentifierArgKey string) func(context.Context, *flags.ParsedFlags, *args.ParsedArgs) error {
	return func(_ context.Context, _ *flags.ParsedFlags, args *args.ParsedArgs) error {
		artifactIdentifier, err := args.GetNonGreedyArg(artifactIdentifierArgKey)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting the identifier to validate using key '%v'", artifactIdentifier)
		}

		if strings.TrimSpace(artifactIdentifier) == emptyArtifactIdentifier {
			return stacktrace.NewError("Artifact identifier cannot be an empty string")
		}

		return nil
	}
}

func getCompletionFunc(enclaveArgKey string) func(ctx context.Context, _ *flags.ParsedFlags, previousArgs *args.ParsedArgs) ([]string, error) {
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

		fileArtifacts, err := enclave.GetAllFilesArtifactNamesAndUuids(ctx)
		if err != nil {
			return nil, stacktrace.Propagate(
				err,
				"An error occurred getting the file artifacts",
			)
		}
		fileArtifactNames := []string{}
		for _, fileArtifact := range fileArtifacts {
			fileName := fileArtifact.GetFileName()
			fileArtifactNames = append(fileArtifactNames, fileName)
		}

		return fileArtifactNames, nil
	}
}
