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
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/module"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/enclave_liveness_validator"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/execution_ids"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/logrus_log_levels"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"io"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	loadParamsFlagKey      = "load-params"
	executeParamsFlagKey       = "execute-params"
	apiContainerVersionFlagKey  = "api-container-version"
	apiContainerLogLevelFlagKey = "api-container-log-level"
	enclaveIdFlagKey             = "enclave-id"
	isPartitioningEnabledFlagKey = "with-partitioning"

	moduleImageArgKey        = "module-image"

	defaultLoadParams            = "{}"
	defaultExecuteParams         = "{}"
	defaultEnclaveId             = ""
	defaultIsPartitioningEnabled = false

	allowedEnclaveIdCharsRegexStr = `^[-A-Za-z0-9.]{1,63}$`

	shouldFollowModuleLogs            = true

	netReadOpt = "read"

	netReadOptFailBecauseSourceIsUsedOrClosedErrorText = "use of closed network connection"

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey = "engine-client"
)

var positionalArgs = []string{
	moduleImageArgKey,
}

var ModuleExecCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.ModuleExecCmdStr,
	ShortDescription:          "Executes a module in an enclave",
	LongDescription:           "Creates a new enclave and loads & executes the given executable module inside it",
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	Flags: []*flags.FlagConfig{
		{
			Key:       loadParamsFlagKey,
			Usage:     "The serialized params that should be passed to the module when loading it",
			Type:      flags.FlagType_String,
			Default:   defaultLoadParams,
		},
		{
			Key:       executeParamsFlagKey,
			Usage:     "The serialized params that should be passed to the module when executing it",
			Type:      flags.FlagType_String,
			Default:   defaultExecuteParams,
		},
		{
			Key:       apiContainerVersionFlagKey,
			Usage:     "The image of the API container that should be started inside the enclave where the module will execute (blank will use the engine's default version)",
			Type:      flags.FlagType_String,
			Default:   defaults.DefaultAPIContainerVersion,
		},
		{
			Key:       apiContainerLogLevelFlagKey,
			Usage:     fmt.Sprintf(
				"The log level that the started API container should log at (%v)",
				strings.Join(logrus_log_levels.GetAcceptableLogLevelStrs(), "|"),
			),
			Shorthand: "l",
			Type:      flags.FlagType_String,
			Default:   defaults.DefaultApiContainerLogLevel.String(),
		},
		{
			Key:       enclaveIdFlagKey,
			Usage:     fmt.Sprintf(
				"The ID to give the enclave that will be created to execute the module inside, which must match regex '%v' (default: use the module image and the current Unix time)",
				// TODO Get this from the Kurtosis backend maybe????
				allowedEnclaveIdCharsRegexStr,
			),
			Type:      flags.FlagType_String,
			Default:   defaultEnclaveId,
		},
		{
			Key:       isPartitioningEnabledFlagKey,
			Usage:     "If set to true, the enclave that the module executes in will have partitioning enabled so network partitioning simulations can be run",
			Type:      flags.FlagType_Bool,
			Default:   strconv.FormatBool(defaultIsPartitioningEnabled),
		},
	},
	Args: []*args.ArgConfig{
		{
			Key:             moduleImageArgKey,
		},
	},
	RunFunc:                   run,
}

func run(
	ctx context.Context,
	// TODO This is a hack that's only here temporarily because we have commands that use KurtosisBackend directly (they
	//  should not), and EngineConsumingKurtosisCommand therefore needs to provide them with a KurtosisBackend. Once all our
	//  commands only access the Kurtosis APIs, we can remove this.
	kurtosisBackend backend_interface.KurtosisBackend,
	engineClient kurtosis_engine_rpc_api_bindings.EngineServiceClient,
	flags *flags.ParsedFlags,
	args *args.ParsedArgs,
) error {
	loadParamsStr, err := flags.GetString(loadParamsFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the module load params using flag key '%v'", loadParamsFlagKey)
	}
	executeParamsStr, err := flags.GetString(executeParamsFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the module execute params using flag key '%v'", executeParamsFlagKey)
	}
	userRequestedEnclaveId, err := flags.GetString(enclaveIdFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave ID using flag key '%v'", enclaveIdFlagKey)
	}
	apiContainerVersion, err := flags.GetString(apiContainerVersionFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the API container version using flag key '%v'", apiContainerVersionFlagKey)
	}
	apiContainerLogLevelStr, err := flags.GetString(apiContainerLogLevelFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the API container log level using flag key '%v'", apiContainerLogLevelFlagKey)
	}
	isPartitioningEnabled, err := flags.GetBool(isPartitioningEnabledFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the is-partitioning-enabled setting using flag key '%v'", isPartitioningEnabledFlagKey)
	}

	moduleImage, err := args.GetNonGreedyArg(moduleImageArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the module image using arg key '%v'", moduleImageArgKey)
	}

	imageNameWithUnixTimestamp := getImageNameWithUnixTimestamp(moduleImage)

	enclaveIdStr := userRequestedEnclaveId
	if enclaveIdStr == defaultEnclaveId {
		enclaveIdStr = imageNameWithUnixTimestamp
	}
	enclaveId := enclave.EnclaveID(enclaveIdStr)

	// TODO Push down into MetricsReportingKurtosisBackend
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
	// TODO
	conn, err := grpc.Dial(apiContainerHostGrpcUrl, grpc.WithInsecure(), grpc.WithBlock())
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
		GUIDs: map[module.ModuleGUID]bool{
			moduleGUID: true,
		},
	}

	//TODO replace with API Container call
	successfulModuleLogs, erroredModuleGuids, err := kurtosisBackend.GetModuleLogs(ctx, enclaveId, moduleFilters, shouldFollowModuleLogs)
	if err != nil {
		//We do not return because logs aren't mandatory, if it fails we continue executing the module without logs
		logrus.Errorf("The module containers logs won't be printed. An error occurred getting logs for module with ID '%v': \n%v", moduleId, err)
	}
	if len(successfulModuleLogs) == 0 {
		return stacktrace.NewError("Didn't find any module logs for newly-created GUID '%v'; this is a bug in Kurtosis", moduleGUID)
	}
	for _, readCloser := range successfulModuleLogs {
		defer readCloser.Close()
	}
	if len(erroredModuleGuids) > 0 {
		moduleLogErr, found := erroredModuleGuids[moduleGUID]
		if !found {
			//We do not return because logs aren't mandatory, if it fails we continue executing the module without logs
			logrus.Errorf("The module containers logs won't be printed. Expected to find an error for module with ID '%v' in error map '%+v' but was not found; this is a bug in Kurtosis", moduleId, erroredModuleGuids)
		}
		if moduleLogErr != nil {
			//We do not return because logs aren't mandatory, if it fails we continue executing the module without logs
			logrus.Errorf("The module containers logs won't be printed. An error occurred getting logs for module with ID '%v': \n%v", moduleId, moduleLogErr)
		}
	}
	readCloserLogs, found := successfulModuleLogs[moduleGUID]
	if !found {
		//We do not return because logs aren't mandatory, if it fails we continue executing the module without logs
		logrus.Errorf("The module containers logs won't be printed. Expected to find the read closer object for module with ID '%v' in successful module logs map '%+v' but was not found; this is a bug in Kurtosis", moduleId, successfulModuleLogs)
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
		"%v--%v",
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
