package engine_functions

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/stacktrace"
	"os"
	"path"
)

const (
	engineLogsSubDirpathSuffix = "engine-logs"
	createdDirPerms            = 0755
	enclavesSubDirpathFragment = "enclaves"
	enclaveNameUuidSeparator   = "--"
)

var allEnclavesFilter = &enclave.EnclaveFilters{}

func DumpKurtosis(ctx context.Context, outputDirpath string, backend backend_interface.KurtosisBackend) error {

	allEnclaves, err := backend.GetEnclaves(ctx, allEnclavesFilter)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while getting a list of enclaves registered with the underlying engine")
	}

	// Create the main output directory
	if _, err = os.Stat(outputDirpath); !os.IsNotExist(err) {
		return stacktrace.NewError("Cannot create output directory at '%v'; directory already exists", outputDirpath)
	}
	if err = os.Mkdir(outputDirpath, createdDirPerms); err != nil {
		return stacktrace.Propagate(err, "An error occurred creating output directory at '%v'", outputDirpath)
	}

	engineOutputDir := path.Join(outputDirpath, engineLogsSubDirpathSuffix)
	if err = backend.GetEngineLogs(ctx, engineOutputDir); err != nil {
		return stacktrace.Propagate(err, "An error occurred while dumping engine logs to dir '%v'", err)
	}

	allEnclavesOutputSubdir := path.Join(outputDirpath, enclavesSubDirpathFragment)
	if err = os.Mkdir(allEnclavesOutputSubdir, createdDirPerms); err != nil {
		return stacktrace.Propagate(err, "An error occurred creating output directory for all enclaves at '%v'", outputDirpath)
	}

	// perhaps dont error immediately and dump what you can before sending errors
	for enclaveUuid, enclave := range allEnclaves {
		subDirForEnclaveBeingDumped := fmt.Sprintf("%v%v%v", enclave.GetName(), enclaveNameUuidSeparator, string(enclaveUuid))
		specificEnclaveOutputDir := path.Join(outputDirpath, enclavesSubDirpathFragment, subDirForEnclaveBeingDumped)
		if err = backend.DumpEnclave(ctx, enclaveUuid, specificEnclaveOutputDir); err != nil {
			return stacktrace.Propagate(err, "An error occurred while dumping enclave with uuid '%v'", enclaveUuid)
		}
	}

	return nil
}
