package rendertemplate

import (
	"context"
	"encoding/json"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/highlevel/enclave_id_arg"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/lib/kurtosis_context"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"os"
	"path"
)

const (
	enclaveIdArgKey        = "enclave-id"
	isEnclaveIdArgOptional = false
	isEnclaveIdArgGreedy   = false

	templateFileArgKey    = "template-file"
	dataJsonFileArgKey    = "data-json-file"
	destRelFilepathArgKey = "destination-relative-filepath"

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"
)

var RenderTemplateCommand = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.FilesUploadCmdStr,
	ShortDescription:          "Renders golang templates to an enclave.",
	LongDescription:           "Renders golang templates to an enclave so that they can be accessed by modules and services inside the enclave.",
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	Args: []*args.ArgConfig{
		enclave_id_arg.NewEnclaveIDArg(
			enclaveIdArgKey,
			engineClientCtxKey,
			isEnclaveIdArgOptional,
			isEnclaveIdArgGreedy,
		),
		{
			Key:            templateFileArgKey,
			ValidationFunc: validateTemplateFileArg,
		},
		{
			Key:            dataJsonFileArgKey,
			ValidationFunc: validateDataJsonFileArg,
		},
		{
			Key:            destRelFilepathArgKey,
			ValidationFunc: validateDestRelFilePathArg,
		},
	},
	RunFunc: run,
}

func run(
	ctx context.Context,
	kurtosisBackend backend_interface.KurtosisBackend,
	engineClient kurtosis_engine_rpc_api_bindings.EngineServiceClient,
	flags *flags.ParsedFlags,
	args *args.ParsedArgs,
) error {
	enclaveIdStr, err := args.GetNonGreedyArg(enclaveIdArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave ID using key '%v'", enclaveIdArgKey)
	}
	enclaveId := enclaves.EnclaveID(enclaveIdStr)

	templateFilepath, err := args.GetNonGreedyArg(templateFileArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the template file using key '%v'", templateFileArgKey)
	}

	dataJsonFilepath, err := args.GetNonGreedyArg(dataJsonFileArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the destination relative filepath using key '%v'", destRelFilepathArgKey)
	}

	destRelFilepath, err := args.GetNonGreedyArg(dataJsonFileArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the data json file using key '%v'", dataJsonFileArgKey)
	}

	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred connecting to the local Kurtosis engine")
	}
	enclaveCtx, err := kurtosisCtx.GetEnclaveContext(ctx, enclaveId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave context for enclave '%v'", enclaveId)
	}

	templateFileBytes, err := os.ReadFile(templateFilepath)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred reading the template file '%v'", templateFilepath)
	}
	templateFileContents := string(templateFileBytes)

	dataJsonFile, err := os.Open(dataJsonFilepath)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred verifying data json '%v' exists and is readable", dataJsonFilepath)
	}
	defer dataJsonFile.Close()

	// We use this so that the large integers in the data json get parsed as integers and not floats
	decoder := json.NewDecoder(dataJsonFile)
	decoder.UseNumber()

	var templateData interface{}
	decoder.Decode(&templateData)

	templateAndData := enclaves.NewTemplateAndData(templateFileContents, templateData)
	templateAndDataByDestRelFilepath := make(map[string]*enclaves.TemplateAndData)
	templateAndDataByDestRelFilepath[destRelFilepath] = templateAndData

	filesArtifactUuid, err := enclaveCtx.RenderTemplates(templateAndDataByDestRelFilepath)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred rendering the template file at path '%v' with data in the file at path '%v' to enclave '%v'", templateFilepath, dataJsonFilepath, enclaveId)
	}
	logrus.Infof("Files package UUID: %v", filesArtifactUuid)
	return nil
}

func validateTemplateFileArg(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {
	templateFile, err := args.GetNonGreedyArg(templateFileArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the template file to validate using key '%v'", templateFileArgKey)
	}

	if _, err := os.Stat(templateFile); err != nil {
		return stacktrace.Propagate(err, "An error occurred verifying templateFileArgKey '%v' exists and is readable", templateFile)
	}
	return nil
}

func validateDataJsonFileArg(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {
	dataJsonFilepath, err := args.GetNonGreedyArg(dataJsonFileArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the data json file to validate using key '%v'", dataJsonFileArgKey)
	}

	dataJsonFileContent, err := os.ReadFile(dataJsonFilepath)

	if err != nil {
		return stacktrace.Propagate(err, "An error occurred verifying data json '%v' exists and is readable", dataJsonFilepath)
	}

	if !json.Valid(dataJsonFileContent) {
		return stacktrace.NewError("The data file isn't valid json")
	}

	return nil
}

func validateDestRelFilePathArg(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {
	destRelFilepath, err := args.GetNonGreedyArg(destRelFilepathArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the destination relative filepath to validate using key '%v'", destRelFilepathArgKey)
	}

	if path.IsAbs(destRelFilepath) {
		return stacktrace.NewError("Expected a relative path got an absolute path")
	}

	return nil
}
