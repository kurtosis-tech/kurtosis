package inspect

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	"sort"
)

const (
	fileUuidsHeader = "UUID"
	fileNameHeader  = "Name"
)

func printFilesArtifacts(ctx context.Context, kurtosisCtx *kurtosis_context.KurtosisContext, enclaveInfo *kurtosis_engine_rpc_api_bindings.EnclaveInfo, showFullUuids bool, _ bool) error {
	enclaveContext, err := kurtosisCtx.GetEnclaveContext(ctx, enclaveInfo.GetName())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while fetching enclave with name '%v'", enclaveInfo.GetName())
	}

	filesArtifactsNamesAndUuids, err := enclaveContext.GetAllFilesArtifactNamesAndUuids(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while fetching files artifacts name and uuids for enclave '%v'", enclaveContext.GetEnclaveName())
	}

	sortedFilesNamesAndUuids := sortFileNamesAndUuids(filesArtifactsNamesAndUuids)

	tablePrinter := output_printers.NewTablePrinter(
		fileUuidsHeader,
		fileNameHeader,
	)

	for _, filesArtifactNameAndUuid := range sortedFilesNamesAndUuids {
		uuid := filesArtifactNameAndUuid.GetFileUuid()
		if !showFullUuids {
			uuid = uuid_generator.ShortenedUUIDString(uuid)
		}
		fileName := filesArtifactNameAndUuid.GetFileName()
		if err := tablePrinter.AddRow(uuid, fileName); err != nil {
			return stacktrace.Propagate(err, "An error occurred while adding row with uuid '%v' and file name '%v'; This is a bug in Kurtosis", uuid, fileName)
		}
	}

	tablePrinter.Print()
	return nil
}

// we sort this in ascending order so that the user finds the table easy to read
func sortFileNamesAndUuids(fileNamesAndUuids []*kurtosis_core_rpc_api_bindings.FilesArtifactNameAndUuid) []*kurtosis_core_rpc_api_bindings.FilesArtifactNameAndUuid {
	sort.Slice(fileNamesAndUuids, func(i, j int) bool {
		firstFilesArtifactNameAndUuid := fileNamesAndUuids[i]
		secondFilesArtifactNameAndUuid := fileNamesAndUuids[j]
		return firstFilesArtifactNameAndUuid.GetFileName() < secondFilesArtifactNameAndUuid.GetFileName()
	})

	return fileNamesAndUuids
}
