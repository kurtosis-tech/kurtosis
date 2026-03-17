package engine_functions

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/stacktrace"
	"os"
	"path"
	"strings"
)

const (
	engineLogsSubDirpathSuffix = "engines"
	createdDirPerms            = 0755
	enclavesSubDirpathFragment = "enclaves"
	enclaveNameUuidSeparator   = "--"
	errorSeparator             = "\n\n"
)

var allEnclavesFilter = &enclave.EnclaveFilters{UUIDs: nil, Statuses: nil}

func DumpKurtosis(ctx context.Context, outputDirpath string, backend backend_interface.KurtosisBackend) error {

	allEnclaves, err := backend.GetEnclaves(ctx, allEnclavesFilter)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while getting a list of enclaves registered with the underlying engine")
	}

	// Check if the directory exists and if it does exit immediately
	// Note os.IsNotExist doesn't throw if the err is nil
	if _, err = os.Stat(outputDirpath); !os.IsNotExist(err) {
		return stacktrace.NewError("Cannot create output directory at '%v'; directory already exists", outputDirpath)
	}
	// If it doesn't we create the directory; if the directory fails on creation we exit
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

	allEnclaveDumpErrors := map[string]string{}
	for enclaveUuid, enclave := range allEnclaves {
		subDirForEnclaveBeingDumped := fmt.Sprintf("%v%v%v", enclave.GetName(), enclaveNameUuidSeparator, string(enclaveUuid))
		specificEnclaveOutputDir := path.Join(outputDirpath, enclavesSubDirpathFragment, subDirForEnclaveBeingDumped)
		if err = backend.DumpEnclave(ctx, enclaveUuid, specificEnclaveOutputDir); err != nil {
			allEnclaveDumpErrors[string(enclaveUuid)] = err.Error()
		}
	}

	if len(allEnclaveDumpErrors) > 0 {
		allIndexedEnclaveErrors := []string{}
		for enclaveUuidStr, errStr := range allEnclaveDumpErrors {
			indexedEnclaveErrorStr := fmt.Sprintf(">>>>>>>>>>>>>>>>> ERROR dumping enclave with UUID '%v' <<<<<<<<<<<<<<<<<\n%v", enclaveUuidStr, errStr)
			allIndexedEnclaveErrors = append(allIndexedEnclaveErrors, indexedEnclaveErrorStr)
		}

		return fmt.Errorf("errors occurred while dumping information for some enclaves :\n'%v'", strings.Join(allIndexedEnclaveErrors, errorSeparator))
	}

	return nil
}
