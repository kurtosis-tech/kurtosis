/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package inspect

import (
	"context"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"strings"

	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/enclave_id_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/service_identifier_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/service_status_stringifier"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/user_services"
	"github.com/kurtosis-tech/kurtosis/cli/cli/out"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	enclaveIdentifierArgKey = "enclave"
	isEnclaveIdArgOptional  = false
	isEnclaveIdArgGreedy    = false

	serviceIdentifierArgKey        = "service"
	isServiceIdentifierArgOptional = false
	isServiceIdentifierArgGreedy   = false

	fullUuidFlagKey        = "full-uuid"
	fullUuidFlagKeyDefault = "false"

	outputFormatKey          = "output"
	outputFormatKeyShorthand = "o"
	outputFormatKeyDefault   = ""
	yamlOutputFormat         = "yaml"

	jsonOutputFormat = "json"

	ServiceNameTitleName                = "Name"
	ServiceUUIDTitleName                = "UUID"
	ServiceStatusTitleName              = "Status"
	ServiceImageTitleName               = "Image"
	ServicePortsTitleName               = "Ports"
	ServiceEntrypointArgsTitleName      = "ENTRYPOINT"
	ServiceCmdArgsTitleName             = "CMD"
	ServiceEnvVarsTitleName             = "ENV"
	ServiceFilesTitleName               = "Files"
	ServiceMaxCpuAllocationTitleName    = "MaxCpu"
	ServiceMinCpuAllocationTitleName    = "MinCpu"
	ServiceMemoryAllocationTitleName    = "MaxMemory"
	ServiceMinMemoryAllocationTitleName = "MinMemory"

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"
)

var ServiceInspectCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.ServiceInspectCmdStr,
	ShortDescription:          "Inspect a service",
	LongDescription:           "List information about the service's status and contents",
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	Flags: []*flags.FlagConfig{
		{
			Key:     fullUuidFlagKey,
			Usage:   "If true then Kurtosis prints the service full UUID instead of the shortened UUID. Default false.",
			Type:    flags.FlagType_Bool,
			Default: fullUuidFlagKeyDefault,
		},
		{
			Key:       outputFormatKey,
			Shorthand: outputFormatKeyShorthand,
			Usage:     "Format to output the result (yaml or json)",
			Type:      flags.FlagType_String,
			Default:   outputFormatKeyDefault,
		},
	},
	Args: []*args.ArgConfig{
		enclave_id_arg.NewEnclaveIdentifierArg(
			enclaveIdentifierArgKey,
			engineClientCtxKey,
			isEnclaveIdArgOptional,
			isEnclaveIdArgGreedy,
		),
		service_identifier_arg.NewServiceIdentifierArg(
			serviceIdentifierArgKey,
			enclaveIdentifierArgKey,
			isServiceIdentifierArgGreedy,
			isServiceIdentifierArgOptional,
		),
	},
	RunFunc: run,
}

func run(
	ctx context.Context,
	kurtosisBackend backend_interface.KurtosisBackend,
	_ kurtosis_engine_rpc_api_bindings.EngineServiceClient,
	_ metrics_client.MetricsClient,
	flags *flags.ParsedFlags,
	args *args.ParsedArgs,
) error {
	enclaveIdentifier, err := args.GetNonGreedyArg(enclaveIdentifierArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for non-greedy enclave identifier arg '%v' but none was found; this is a bug in the Kurtosis CLI!", enclaveIdentifierArgKey)
	}

	serviceIdentifier, err := args.GetNonGreedyArg(serviceIdentifierArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for non-greedy enclave identifier arg '%v' but none was found; this is a bug in the Kurtosis CLI!", serviceIdentifier)
	}

	showFullUuid, err := flags.GetBool(fullUuidFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for the '%v' flag but failed to get it", fullUuidFlagKey)
	}

	outputFormat, err := flags.GetString(outputFormatKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for the '%v' flag but failed to get it", outputFormatKey)
	}

	outputFormat = strings.ToLower(strings.TrimSpace(outputFormat))
	if outputFormat != "" && outputFormat != jsonOutputFormat && outputFormat != yamlOutputFormat {
		return stacktrace.NewError("Invalid output format '%s'; must be '%v' or '%v'", outputFormat, jsonOutputFormat, yamlOutputFormat)
	}

	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating Kurtosis Context from local engine")
	}

	if _, err = PrintServiceInspect(ctx, kurtosisBackend, kurtosisCtx, enclaveIdentifier, serviceIdentifier, showFullUuid, outputFormat); err != nil {
		// this is already wrapped up
		return err
	}
	return nil
}

func PrintServiceInspect(ctx context.Context, kurtosisBackend backend_interface.KurtosisBackend, kurtosisCtx *kurtosis_context.KurtosisContext, enclaveIdentifier string, serviceIdentifier string, showFullUuid bool, outputFormat string) (map[string]interface{}, error) {
	var jsonMap map[string]interface{}
	enclaveInfo, err := kurtosisCtx.GetEnclave(ctx, enclaveIdentifier)
	if err != nil {
		return jsonMap, stacktrace.Propagate(err, "An error occurred getting the enclave for identifier '%v'", enclaveIdentifier)
	}

	enclaveApiContainerStatus := enclaveInfo.ApiContainerStatus
	isApiContainerRunning := enclaveApiContainerStatus == kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus_EnclaveAPIContainerStatus_RUNNING

	userServices := map[string]*kurtosis_core_rpc_api_bindings.ServiceInfo{}
	if isApiContainerRunning {
		var err error
		serviceMap := map[string]bool{
			serviceIdentifier: true,
		}
		userServices, err = user_services.GetUserServiceInfoMapFromAPIContainer(ctx, enclaveInfo, serviceMap)
		if err != nil {
			return jsonMap, stacktrace.Propagate(err, "Failed to get service info from API container in enclave '%v'", enclaveInfo.GetEnclaveUuid())
		}
	}

	var userService *kurtosis_core_rpc_api_bindings.ServiceInfo
	for _, userServiceInfo := range userServices {
		userService = userServiceInfo
		break
	}

	if outputFormat != "" {
		jsonMap[ServiceImageTitleName] = userService.GetContainer().GetImageName()
		jsonMap[ServiceCmdArgsTitleName] = userService.GetContainer().GetCmdArgs()
		jsonMap[ServiceEntrypointArgsTitleName] = userService.GetContainer().GetEntrypointArgs()
		jsonMap[ServiceEnvVarsTitleName] = userService.GetContainer().GetEnvVars()
		jsonMap[ServicePortsTitleName] = userService.GetPrivatePorts()
		jsonMap["Files"] = ""
		// TODO: add support for files
		// How would we get files information?
		var marshaled []byte
		var err error
		switch outputFormat {
		case jsonOutputFormat:
			marshaled, err = json.MarshalIndent(jsonMap, "", "")
		case yamlOutputFormat:
			marshaled, err = yaml.Marshal(jsonMap)
		}
		if err != nil {
			return jsonMap, stacktrace.Propagate(err, "Failed to marshal service info to %s", outputFormat)
		}
		out.PrintOutLn(string(marshaled))
		return jsonMap, nil
	}

	out.PrintOutLn(fmt.Sprintf("%s: %s", ServiceNameTitleName, userService.GetName()))

	uuidStr := userService.GetShortenedUuid()
	if showFullUuid {
		uuidStr = userService.GetServiceUuid()
	}
	out.PrintOutLn(fmt.Sprintf("%s: %s", ServiceUUIDTitleName, uuidStr))

	serviceStatus := userService.GetServiceStatus()
	serviceStatusStr := service_status_stringifier.ServiceStatusStringifier(serviceStatus)
	out.PrintOutLn(fmt.Sprintf("%s: %s", ServiceStatusTitleName, serviceStatusStr))

	out.PrintOutLn(fmt.Sprintf("%s: %s", ServiceImageTitleName, userService.GetContainer().GetImageName()))

	out.PrintOutLn(fmt.Sprintf("%s:", ServicePortsTitleName))
	portBindingLines, err := user_services.GetUserServicePortBindingStrings(userService)
	if err != nil {
		return jsonMap, stacktrace.Propagate(err, "An error occurred getting the port binding strings")
	}
	for _, portBindingLine := range portBindingLines {
		out.PrintOutLn(fmt.Sprintf("  %s", portBindingLine))
	}

	out.PrintOutLn(fmt.Sprintf("%s:", ServiceEntrypointArgsTitleName))
	for _, entrypointArg := range userService.GetContainer().GetEntrypointArgs() {
		out.PrintOutLn(fmt.Sprintf("  %s", entrypointArg))
	}

	out.PrintOutLn(fmt.Sprintf("%s:", ServiceCmdArgsTitleName))
	for _, cmdArg := range userService.GetContainer().GetCmdArgs() {
		out.PrintOutLn(fmt.Sprintf("  %s", cmdArg))
	}

	out.PrintOutLn(fmt.Sprintf("%s:", ServiceEnvVarsTitleName))
	for envVarKey, envVarVal := range userService.GetContainer().GetEnvVars() {
		out.PrintOutLn(fmt.Sprintf("  %s: %s", envVarKey, envVarVal))
	}

	return jsonMap, nil
}
