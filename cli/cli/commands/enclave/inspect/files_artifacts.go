package inspect

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	"sort"
)

const (
	fileUuidsHeader = "UUID"
	fileNameHeader  = "Name"
)

func printFilesArtifacts(ctx context.Context, kurtosisCtx *kurtosis_context.KurtosisContext, _ backend_interface.KurtosisBackend, enclaveInfo *kurtosis_engine_rpc_api_bindings.EnclaveInfo, showFullUuids bool, _ bool) error {
	enclaveContext, err := kurtosisCtx.GetEnclaveContext(ctx, enclaveInfo.GetName())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while fetching enclave with name '%v'", enclaveInfo.GetName())
	}

	filesArtifactsNamesAndUuids, err := enclaveContext.GetAllFilesArtifactNamesAndUuids(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while fetching files artifacts name and uuids for enclave '%v'", enclaveContext.GetEnclaveName())
	}

	sort.Slice(filesArtifactsNamesAndUuids, func(i, j int) bool {
		firstFilesArtifactNameAndUuid := filesArtifactsNamesAndUuids[i]
		secondFilesArtifactNameAndUuid := filesArtifactsNamesAndUuids[i]
		return firstFilesArtifactNameAndUuid.GetFileName() < secondFilesArtifactNameAndUuid.GetFileName()
	})

	tablePrinter := output_printers.NewTablePrinter(
		fileUuidsHeader,
		fileNameHeader,
	)

	for _, filesArtifactNameAndUuid := range filesArtifactsNamesAndUuids {
		uuid := filesArtifactNameAndUuid.GetFileUuid()
		if !showFullUuids {
			uuid = uuid_generator.ShortenedUUIDString(uuid)
		}
		fileName := filesArtifactNameAndUuid.GetFileName()
		tablePrinter.AddRow(uuid, fileName)
	}

	tablePrinter.Print()
	return nil
}
