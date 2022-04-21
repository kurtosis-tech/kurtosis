/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package exec

import (
	"context"
	"fmt"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/kurtosis-tech/container-engine-lib/lib"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/module"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/enclave_liveness_validator"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/execution_ids"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/logrus_log_levels"
	"github.com/kurtosis-tech/kurtosis-cli/commons/positional_arg_parser"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"io"
	"net"
	"regexp"
	"strings"
	"time"
)

const (
	loadParamsStrArg         = "load-params"
	executeParamsStrArg      = "execute-params"
	apiContainerVersionArg   = "api-container-version"
	apiContainerLogLevelArg  = "api-container-log-level"
	moduleImageArg           = "module-image"
	enclaveIdArg             = "enclave-id"
	isPartitioningEnabledArg = "with-partitioning"

	defaultLoadParams            = "{}"
	defaultExecuteParams         = "{}"
	defaultEnclaveId             = ""
	defaultIsPartitioningEnabled = false

	shouldPublishAllPorts = true

	allowedEnclaveIdCharsRegexStr = `^[-A-Za-z0-9.]{1,63}$`

	shouldFollowModuleLogs            = true

	netReadOpt = "read"

	netReadOptFailBecauseSourceIsUsedOrClosedErrorText = "use of closed network connection"
)

var positionalArgs = []string{
	moduleImageArg,
}

var loadParamsStr string
var executeParamsStr string
var apiContainerVersion string
var apiContainerLogLevelStr string
var userRequestedEnclaveId string
var isPartitioningEnabled bool

var ExecCmd = &cobra.Command{
	Use:                   command_str_consts.ModuleExecCmdStr + " [flags] " + strings.Join(positionalArgs, " "),
	DisableFlagsInUseLine: true,
	Short:                 "Creates a new enclave and loads & executes the given executable module inside it",
	RunE:                  run,
}

func init() {
	ExecCmd.Flags().StringVar(
		&loadParamsStr,
		loadParamsStrArg,
		defaultLoadParams,
		"The serialized params that should be passed to the module when loading it",
	)
	ExecCmd.Flags().StringVar(
		&executeParamsStr,
		executeParamsStrArg,
		defaultExecuteParams,
		"The serialized params that should be passed to the module when executing it",
	)
	ExecCmd.Flags().StringVar(
		&apiContainerVersion,
		apiContainerVersionArg,
		defaults.DefaultAPIContainerVersion,
		"The image of the API container that should be started inside the enclave where the module will execute (blank will use the engine's default version)",
	)
	ExecCmd.Flags().StringVarP(
		&apiContainerLogLevelStr,
		apiContainerLogLevelArg,
		"l",
		defaults.DefaultApiContainerLogLevel.String(),
		fmt.Sprintf(
			"The log level that the started API container should log at (%v)",
			strings.Join(logrus_log_levels.GetAcceptableLogLevelStrs(), "|"),
		),
	)
	ExecCmd.Flags().StringVar(
		&userRequestedEnclaveId,
		enclaveIdArg,
		defaultEnclaveId,
		fmt.Sprintf(
			"The ID to give the enclave that will be created to execute the module inside, which must match regex '%v' (default: use the module image and the current Unix time)",
			// TODO Get this from the Kurtosis backend maybe????
			allowedEnclaveIdCharsRegexStr,
		),
	)
	ExecCmd.Flags().BoolVar(
		&isPartitioningEnabled,
		isPartitioningEnabledArg,
		defaultIsPartitioningEnabled,
		"If set to true, the enclave that the module executes in will have partitioning enabled so network partitioning simulations can be run",
	)
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	parsedPositionalArgs, err := positional_arg_parser.ParsePositionalArgsAndRejectEmptyStrings(positionalArgs, args)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the positional args")
	}
	moduleImage := parsedPositionalArgs[moduleImageArg]

	imageNameWithUnixTimestamp := getImageNameWithUnixTimestamp(moduleImage)

	enclaveIdStr := userRequestedEnclaveId
	if enclaveIdStr == defaultEnclaveId {
		enclaveIdStr = imageNameWithUnixTimestamp
	}
	enclaveId := enclave.EnclaveID(enclaveIdStr)

	validEnclaveId, err := regexp.Match(allowedEnclaveIdCharsRegexStr, []byte(enclaveIdStr))
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred validating that enclave ID '%v' matches allowed enclave ID regex '%v'",
			enclaveIdStr,
			allowedEnclaveIdCharsRegexStr,
		)
	}
	if !validEnclaveId {
		return stacktrace.NewError(
			"Enclave ID '%v' doesn't match allowed enclave ID regex '%v'",
			enclaveIdStr,
			allowedEnclaveIdCharsRegexStr,
		)
	}

	kurtosisBackend, err := lib.GetLocalDockerKurtosisBackend()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting a Kurtosis backend connected to local Docker")
	}
	engineManager := engine_manager.NewEngineManager(kurtosisBackend)
	engineClient, closeClientFunc, err := engineManager.StartEngineIdempotentlyWithDefaultVersion(ctx, defaults.DefaultEngineLogLevel)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating a new Kurtosis engine client")
	}
	defer closeClientFunc()

	getEnclavesResp, err := engineClient.GetEnclaves(ctx, &emptypb.Empty{})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the existing enclaves")
	}

	enclaveInfo, foundExistingEnclave := getEnclavesResp.EnclaveInfo[enclaveIdStr]
	// If no enclave with the requested ID exists, create it
	didModuleExecutionCompleteSuccessfully := false
	if !foundExistingEnclave {
		logrus.Infof("Creating enclave '%v' for the module to execute inside...", enclaveIdStr)
		createEnclaveArgs := &kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs{
			EnclaveId:              enclaveIdStr,
			ApiContainerVersionTag: apiContainerVersion,
			ApiContainerLogLevel:   apiContainerLogLevelStr,
			IsPartitioningEnabled:  isPartitioningEnabled,
			ShouldPublishAllPorts:  shouldPublishAllPorts,
		}

		response, err := engineClient.CreateEnclave(ctx, createEnclaveArgs)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred creating an enclave with ID '%v'", enclaveIdStr)
		}
		didModuleExecutionCompleteSuccessfully = true
		defer func() {
			if !didModuleExecutionCompleteSuccessfully {
				logrus.Warnf(
					"NOTE: Even though the module execution didn't complete successfully, we've left the enclave we've created running so you can continue to debug; to stop this enclave and free its resources, run '%v %v %v %v'",
					command_str_consts.KurtosisCmdStr,
					command_str_consts.EnclaveCmdStr,
					command_str_consts.EnclaveStopCmdStr,
					enclaveIdStr,
				)
			}
		}()
		enclaveInfo = response.GetEnclaveInfo()
		logrus.Infof("Enclave '%v' created successfully", enclaveIdStr)
	}

	apicHostMachineIp, apicHostMachineGrpcPort, err := enclave_liveness_validator.ValidateEnclaveLiveness(enclaveInfo)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred verifying that the enclave was running")
	}

	apiContainerHostGrpcUrl := fmt.Sprintf(
		"%v:%v",
		apicHostMachineIp,
		apicHostMachineGrpcPort,
	)
	conn, err := grpc.Dial(apiContainerHostGrpcUrl, grpc.WithInsecure())
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred connecting to the API container grpc port at '%v' in enclave '%v'",
			apiContainerHostGrpcUrl,
			enclaveIdStr,
		)
	}
	defer func() {
		conn.Close()
	}()
	apiContainerClient := kurtosis_core_rpc_api_bindings.NewApiContainerServiceClient(conn)

	logrus.Infof(
		"Loading module '%v' with load params '%v' inside enclave '%v'...",
		moduleImage,
		loadParamsStr,
		enclaveIdStr,
	)
	moduleId := imageNameWithUnixTimestamp
	loadModuleArgs := binding_constructors.NewLoadModuleArgs(moduleId, moduleImage, loadParamsStr)

	loadModuleResponse, err := apiContainerClient.LoadModule(ctx, loadModuleArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred loading the module with image '%v'", moduleImage)
	}
	defer func() {
		if !didModuleExecutionCompleteSuccessfully {
			 logrus.Warnf(
				 "Module execution didn't complete successfully; we've left the module and the services it started inside enclave '%v' for debugging",
				 enclaveIdStr,
			 )
		}
	}()
	logrus.Info("Module loaded successfully")

	moduleGUIDStr := loadModuleResponse.GetGuid()
	moduleGUID := module.ModuleGUID(moduleGUIDStr)

	moduleFilters := &module.ModuleFilters{
		EnclaveIDs: map[enclave.EnclaveID]bool{
			enclaveId: true,
		},
		GUIDs: map[module.ModuleGUID]bool{
			moduleGUID: true,
		},
	}

	successfulModuleLogs, erroredModuleGuids, err := kurtosisBackend.GetModuleLogs(ctx, moduleFilters, shouldFollowModuleLogs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the module logs using filters '%+v'", moduleFilters)
	}
	if len(erroredModuleGuids) > 0 {
		moduleLogErr, found := erroredModuleGuids[moduleGUID]
		if !found {
			return stacktrace.NewError("Expected to find an error for module with GUID '%v' in error map '%+v' but was not found; this is a bug in Kurtosis", moduleGUID, erroredModuleGuids)
		}
		return stacktrace.Propagate(moduleLogErr, "An error occurred getting module logs for module with GUID '%v'", moduleGUID)
	}
	readCloserLogs, found := successfulModuleLogs[moduleGUID]
	if !found {
		return stacktrace.NewError("Expected to find the read closer object for module with GUID '%v' in successful module logs map '%+v' but was not found; this is a bug in Kurtosis", moduleGUID, successfulModuleLogs)
	}

	logrus.Infof("Executing the module with execute params '%v'...", executeParamsStr)
	if readCloserLogs != nil {
		go printModuleContainerLogs(readCloserLogs)
		logrus.Info("----------------------- MODULE LOGS ----------------------")
	}

	executeModuleArgs := binding_constructors.NewExecuteModuleArgs(moduleId, executeParamsStr)
	executeModuleResult, err := apiContainerClient.ExecuteModule(ctx, executeModuleArgs)

	//Stops printing logs
	if readCloserLogs != nil {
		readCloserLogs.Close()
		logrus.Info("--------------------- END MODULE LOGS --------------------")
	}
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred executing the module with params '%v'", executeParamsStr)
	}
	logrus.Infof(
		"Module executed successfully and returned the following result:\n%v",
		executeModuleResult.SerializedResult,
	)

	didModuleExecutionCompleteSuccessfully = true
	return nil
}

// ====================================================================================================
//                                      Private Helper Methods
// ====================================================================================================
func getImageNameWithUnixTimestamp(moduleImage string) string {
	enclaveId := execution_ids.GetExecutionID()
	parsedModuleImage, err := reference.Parse(moduleImage)
	if err != nil {
		logrus.Warnf("Couldn't parse the module image string '%v'; using enclave ID '%v'", moduleImage, enclaveId)
		return enclaveId
	}

	namedModuleImage, ok := parsedModuleImage.(reference.Named)
	if !ok {
		logrus.Warnf("Module image string '%v' couldn't be cast to a named reference; using enclave ID '%v'", moduleImage, enclaveId)
		return enclaveId
	}
	pathElement := reference.Path(namedModuleImage)

	return fmt.Sprintf(
		"%v.%v",
		pathElement,
		time.Now().Unix(),
	)
}

func printModuleContainerLogs(readCloserLogs io.ReadCloser) {
	if _, err := stdcopy.StdCopy(logrus.StandardLogger().Out, logrus.StandardLogger().Out, readCloserLogs); err != nil {
		opError, ok := err.(*net.OpError)
		if ok {
			//We ignore this type of error because it was generated when the main go routine closes the readCloserLogs object
			if opError.Op == netReadOpt && strings.Contains(opError.Error(), netReadOptFailBecauseSourceIsUsedOrClosedErrorText) {
				return
			}
		}
		logrus.Errorf("An error occurred copying the container logs to STDOUT: \n %v", err)
	}
}
