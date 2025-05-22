/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package inspect

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/enclave_id_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/enclave_status_stringifier"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/kurtosis/cli/cli/out"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	enclaveIdentifierArgKey = "enclave"
	isEnclaveIdArgOptional  = false
	isEnclaveIdArgGreedy    = false

	enclaveUUIDTitleName         = "UUID"
	enclaveNameTitleName         = "Name"
	enclaveStatusTitleName       = "Status"
	enclaveCreationTimeTitleName = "Creation Time"
	flagsTitleName               = "Flags"

	jsonOutput             = "json"
	fullUuidsFlagKey       = "full-uuids"
	fullUuidFlagKeyDefault = "false"

	headerWidthChars = 100
	headerPadChar    = "="

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"

	userServicesArtifactsHeader = "User Services"
	filesArtifactsHeader        = "Files Artifacts"

	productionEnclaveFlagStr = "production"
)

type FilesArtifact struct {
	UUID string
	Name string
}

type UserService struct {
	UUID   string
	Name   string
	Ports  []string
	Status string
}

type InspectOutput struct {
	Basic    map[string]string `json:"Basic"`
	Files    []FilesArtifact   `json:"FilesArtifacts"`
	Services []UserService     `json:"UserServices"`
}

var EnclaveInspectCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.EnclaveInspectCmdStr,
	ShortDescription:          "Inspect an enclave",
	LongDescription:           "List information about the enclave's status and contents",
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	Flags: []*flags.FlagConfig{
		{
			Key:     fullUuidsFlagKey,
			Usage:   "If true then Kurtosis prints full UUIDs instead of shortened UUIDs. Default false.",
			Type:    flags.FlagType_Bool,
			Default: fullUuidFlagKeyDefault,
		},
		{
			Key:     jsonOutput,
			Usage:   "If true then Kurtosis prints the output in JSON format. Default false.",
			Type:    flags.FlagType_Bool,
			Default: "false",
		},
	},
	Args: []*args.ArgConfig{
		enclave_id_arg.NewEnclaveIdentifierArg(
			enclaveIdentifierArgKey,
			engineClientCtxKey,
			isEnclaveIdArgOptional,
			isEnclaveIdArgGreedy,
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

	showFullUuids, err := flags.GetBool(fullUuidsFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for the '%v' flag but failed to get it", fullUuidsFlagKey)
	}

	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating Kurtosis Context from local engine")
	}

	inspectInfo, err := inspect(ctx, kurtosisCtx, enclaveIdentifier, showFullUuids)
	if err != nil {
		// this is already wrapped up
		return err
	}
	isjsonOutput, _ := flags.GetBool(jsonOutput)
	if isjsonOutput {
		tmp, _ := json.Marshal(inspectInfo)
		fmt.Println(string(tmp))
	} else {
		tablePrint(inspectInfo)
	}
	return nil
}

func tablePrint(info *InspectOutput) {
	basicPrinter := output_printers.NewKeyValuePrinter()
	basicPrinter.AddPair(enclaveNameTitleName, info.Basic[enclaveNameTitleName])
	basicPrinter.AddPair(enclaveUUIDTitleName, info.Basic[enclaveUUIDTitleName])
	basicPrinter.AddPair(enclaveStatusTitleName, info.Basic[enclaveStatusTitleName])
	basicPrinter.AddPair(enclaveCreationTimeTitleName, info.Basic[enclaveCreationTimeTitleName])
	basicPrinter.AddPair(flagsTitleName, info.Basic[flagsTitleName])
	basicPrinter.Print()
	out.PrintOutLn("")
	numRunesInHeader := utf8.RuneLen(' ') + utf8.RuneCountInString(filesArtifactsHeader) + utf8.RuneLen(' ') // there will be a space before and after the header
	numPadChars := (headerWidthChars - numRunesInHeader) / 2                                                 //nolint:mnd
	padStr := strings.Repeat(headerPadChar, numPadChars)
	fmt.Printf("%v %v %v\n", padStr, filesArtifactsHeader, padStr)

	tablePrinter := output_printers.NewTablePrinter(
		fileUuidsHeader,
		fileNameHeader,
	)

	for _, item := range info.Files {
		tablePrinter.AddRow(item.UUID, item.Name)
	}
	tablePrinter.Print()
	out.PrintOutLn("")

	numRunesInHeader = utf8.RuneLen(' ') + utf8.RuneCountInString(userServicesArtifactsHeader) + utf8.RuneLen(' ') // there will be a space before and after the header
	numPadChars = (headerWidthChars - numRunesInHeader) / 2                                                        //nolint:mnd
	padStr = strings.Repeat(headerPadChar, numPadChars)
	fmt.Printf("%v %v %v\n", padStr, userServicesArtifactsHeader, padStr)

	userServicePrinter := output_printers.NewTablePrinter(
		userServiceUUIDColHeader,
		userServiceNameColHeader,
		userServicePortsColHeader,
		userServiceStatusColHeader,
	)

	for _, item := range info.Services {
		userServicePrinter.AddRow(item.UUID, item.Name, item.Ports[0], item.Status)
		for i := 1; i < len(item.Ports); i++ {
			userServicePrinter.AddRow("", "", item.Ports[i], "")
		}
	}
	userServicePrinter.Print()
	out.PrintOutLn("")
}

func baseInfo(info *kurtosis_engine_rpc_api_bindings.EnclaveInfo, showFullUuids bool) (map[string]string, error) {

	// Add UUID row
	uuid := info.GetShortenedUuid()
	if showFullUuids {
		uuid = info.GetEnclaveUuid()
	}

	// Add status row
	enclaveContainersStatusStr, err := enclave_status_stringifier.EnclaveContainersStatusStringifier(info.GetContainersStatus())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred when stringify enclave containers status")
	}

	// Add creation time row
	enclaveCreationTime := info.GetCreationTime()
	if enclaveCreationTime == nil {
		return nil, stacktrace.Propagate(err, "Expected to get the enclave creation time from the enclave info received but it was not received, this is a bug in Kurtosis")
	}
	enclaveCreationTimeStr := enclaveCreationTime.AsTime().Local().Format(time.RFC1123)

	res := map[string]string{
		enclaveNameTitleName:         info.GetName(),
		enclaveUUIDTitleName:         uuid,
		enclaveStatusTitleName:       enclaveContainersStatusStr,
		enclaveCreationTimeTitleName: enclaveCreationTimeStr,
		flagsTitleName:               getAllEnclaveFlagsStr(info),
	}
	return res, nil
}

func PrintEnclaveInspect(ctx context.Context, kurtosisCtx *kurtosis_context.KurtosisContext, enclaveIdentifier string, showFullUuids bool) error {
	inspectInfo, err := inspect(ctx, kurtosisCtx, enclaveIdentifier, showFullUuids)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave inspect info")
	} else {
		tablePrint(inspectInfo)
	}
	return nil
}

func inspect(ctx context.Context, kurtosisCtx *kurtosis_context.KurtosisContext, enclaveIdentifier string, showFullUuids bool) (*InspectOutput, error) {
	var inspectInfo = InspectOutput{}
	enclaveInfo, err := kurtosisCtx.GetEnclave(ctx, enclaveIdentifier)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the enclave for identifier '%v'", enclaveIdentifier)
	}

	inspectInfo.Basic, err = baseInfo(enclaveInfo, showFullUuids)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the base info for enclave '%v'", enclaveIdentifier)
	}

	isApiContainerRunning := enclaveInfo.GetApiContainerStatus() == kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus_EnclaveAPIContainerStatus_RUNNING

	inspectInfo.Files, err = inspectFilesArtifacts(ctx, kurtosisCtx, enclaveInfo, showFullUuids, isApiContainerRunning)
	if err != nil {
		return nil, err
	}

	inspectInfo.Services, err = inspectUserServices(ctx, kurtosisCtx, enclaveInfo, showFullUuids, isApiContainerRunning)
	if err != nil {
		return nil, err
	}

	return &inspectInfo, nil
}

func getAllEnclaveFlagsStr(enclaveInfo *kurtosis_engine_rpc_api_bindings.EnclaveInfo) string {
	allEnclaveFragsStr := ""

	// we only one enclave flag added so far, but we could have more in the future, so we should add them here
	// and return all of them together in just one string
	currentModeStr := strings.ToLower(enclaveInfo.GetMode().String())

	if currentModeStr == productionEnclaveFlagStr {
		allEnclaveFragsStr = currentModeStr
	}

	return allEnclaveFragsStr
}
