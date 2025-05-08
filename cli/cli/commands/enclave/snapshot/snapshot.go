/*
 * Copyright (c) 2024 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package snapshot

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/enclave_id_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/file_system_path_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/shared_starlark_calls"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/mholt/archiver"
	"github.com/sirupsen/logrus"
)

const (
	enclaveIdentifierArgKey        = "enclave-identifier"
	enclaveIdentifierArgIsOptional = false
	enclaveIdentifierArgIsGreedy   = false

	destinationPathArgKey        = "output-path"
	isDestinationPathArgOptional = true
	emptyDestinationPathArg      = ""

	extractFlagKey          = "extract"
	extractFlagDefaultValue = "false"

	snapshotExtension                = ".tgz"
	snapshotDestinationDirPermission = 0o777

	defaultTmpDir = ""
	tmpDirPattern = "tmp-dir-for-download-*"

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"
)

var EnclaveSnapshotCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.EnclaveSnapshotCmdStr,
	ShortDescription:          "Takes a snapshot of an enclave",
	LongDescription:           "Takes a snapshot of the specified enclave, capturing its current state",
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	Flags: []*flags.FlagConfig{
		{
			Key:     extractFlagKey,
			Usage:   "If true then the file will be extracted. Default to false.",
			Type:    flags.FlagType_Bool,
			Default: extractFlagDefaultValue,
		},
	},
	Args: []*args.ArgConfig{
		enclave_id_arg.NewEnclaveIdentifierArg(
			enclaveIdentifierArgKey,
			engineClientCtxKey,
			false,
			false,
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
	kurtosisBackend backend_interface.KurtosisBackend,
	_ kurtosis_engine_rpc_api_bindings.EngineServiceClient,
	_ metrics_client.MetricsClient,
	_ *flags.ParsedFlags,
	args *args.ParsedArgs,
) error {
	enclaveIdentifier, err := args.GetNonGreedyArg(enclaveIdentifierArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for enclave identifier but none was found")
	}
	if enclaveIdentifier == "" {
		return stacktrace.NewError("Enclave identifier cannot be empty")
	}

	destinationPath, err := args.GetNonGreedyArg(destinationPathArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the destination path to write downloaded file to using key '%v'", destinationPath)
	}
	if destinationPath == "" {
		destinationPath = "./"
	}
	absoluteDestinationPath, err := filepath.Abs(destinationPath)
	if err != nil {
		return stacktrace.NewError("An error occurred while getting absolute path for the passed destination path '%v'", destinationPath)
	}
	// check if the path doesn't exist, we create a directory if it doesn't
	// we have already checked if the passed path isn't a dir so we don't do it again
	_, err = os.Stat(absoluteDestinationPath)
	if os.IsNotExist(err) {
		err = os.Mkdir(absoluteDestinationPath, snapshotDestinationDirPermission)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while creating the target dir")
		}
	}

	shouldExtract := false
	// shouldNotExtract, err := flags.GetBool(extractFlagKey)
	// if err != nil {
	// 	return stacktrace.Propagate(err, "An error occurred while getting the default value for the '%v' flag", extractFlagKey)
	// }

	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating kurtosis context from local engine.")
	}

	enclaveCtx, err := kurtosisCtx.GetEnclaveContext(ctx, enclaveIdentifier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting enclave context for enclave '%v'", enclaveIdentifier)
	}

	err = stopAllEnclaveServices(ctx, enclaveIdentifier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred stopping all services in enclave '%v'", enclaveIdentifier)
	}

	snapshotContentBytes, err := enclaveCtx.CreateSnapshot()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating snapshot for enclave '%v'", enclaveIdentifier)
	}
	logrus.Infof("Successfully created snapshot for enclave '%v'", enclaveIdentifier)

	// output to to path
	// if the user doesn't want to extract, we just download and return
	err = outputSnapshotToPath(snapshotContentBytes, enclaveIdentifier, absoluteDestinationPath, shouldExtract)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while outputting snapshot to path '%v'", absoluteDestinationPath)
	}
	logrus.Infof("Output snapshot to '%v'", absoluteDestinationPath)

	// err = kurtosisCtx.StopEnclave(ctx, enclaveIdentifier)
	// if err != nil {
	// 	return stacktrace.Propagate(err, "An error occurred stopping enclave '%v'", enclaveIdentifier)
	// }
	// logrus.Infof("Successfully stopped enclave '%v'", enclaveIdentifier)

	return nil
}

func stopAllEnclaveServices(ctx context.Context, enclaveIdentifier string) error {
	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating Kurtosis Context from local engine")
	}

	enclaveCtx, err := kurtosisCtx.GetEnclaveContext(ctx, enclaveIdentifier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting an enclave context from enclave info for enclave '%v'", enclaveIdentifier)
	}

	allEnclaveServices, err := enclaveCtx.GetServices()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting all enclave services")
	}

	for serviceName := range allEnclaveServices {
		if err := shared_starlark_calls.StopServiceStarlarkCommand(ctx, enclaveCtx, serviceName); err != nil {
			return stacktrace.Propagate(err, "An error occurred stopping service '%s'", serviceName)
		}
	}
	return nil
}

func outputSnapshotToPath(snapshotContentBytes []byte, enclaveIdentifier string, absoluteDestinationPath string, shouldExtract bool) error {
	fileNameToWriteTo := fmt.Sprintf("%v-%v%v", enclaveIdentifier, time.Now().Unix(), snapshotExtension)
	destinationPathToDownloadFileTo := path.Join(absoluteDestinationPath, fileNameToWriteTo)

	if shouldExtract {
		tmpDirPath, err := os.MkdirTemp(defaultTmpDir, tmpDirPattern)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while creating a temporary directory to download the snapshot with identifier '%v' to", enclaveIdentifier)
		}
		shouldCleanupTmpDir := false
		defer func() {
			if shouldCleanupTmpDir {
				os.RemoveAll(tmpDirPath)
			}
		}()

		tmpFileToWriteTo := path.Join(tmpDirPath, fileNameToWriteTo)
		err = os.WriteFile(tmpFileToWriteTo, snapshotContentBytes, snapshotDestinationDirPermission)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while writing bytes to file '%v' with permission '%v'", tmpDirPath, snapshotDestinationDirPermission)
		}

		err = archiver.Unarchive(tmpFileToWriteTo, absoluteDestinationPath)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while extracting '%v' to '%v'", tmpFileToWriteTo, absoluteDestinationPath)
		}

		shouldCleanupTmpDir = true
	} else {
		err := os.WriteFile(destinationPathToDownloadFileTo, snapshotContentBytes, snapshotDestinationDirPermission)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while writing bytes to file '%v' with permission '%v'", destinationPathToDownloadFileTo, snapshotDestinationDirPermission)
		}
	}

	return nil
}
