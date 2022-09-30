package rendertemplate

import (
	"context"
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/enclave_id_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"os"
	"path"
)

const (
	enclaveIdArgKey        = "enclave-id"
	isEnclaveIdArgOptional = false
	isEnclaveIdArgGreedy   = false

	templateFilepathArgKey = "template-filepath"
	dataJSONFilepathArgKey = "data-json-filepath"
	destRelFilepathArgKey  = "destination-relative-filepath"

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"
)

var RenderTemplateCommand = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.FilesRenderTemplate,
	ShortDescription:          "Renders a template to an enclave.",
	LongDescription:           "Renders a Golang text/template to an enclave so that the output can be accessed by modules and services inside the enclave.",
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
			Key:            templateFilepathArgKey,
			ValidationFunc: validateTemplateFileArg,
		},
		{
			Key:            dataJSONFilepathArgKey,
			ValidationFunc: validateDataJSONFileArg,
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

	templateFilepath, err := args.GetNonGreedyArg(templateFilepathArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the template file using key '%v'", templateFilepathArgKey)
	}

	dataJSONFilepath, err := args.GetNonGreedyArg(dataJSONFilepathArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the data JSON file using key '%v'", dataJSONFilepathArgKey)
	}

	destRelFilepath, err := args.GetNonGreedyArg(destRelFilepathArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the destination relative filepath using key '%v'", destRelFilepathArgKey)
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

	dataJSONFile, err := os.Open(dataJSONFilepath)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred opening the data JSON file '%v'", dataJSONFilepath)
	}
	defer dataJSONFile.Close()

	// We use this so that the large integers in the data JSON get parsed as integers and not floats
	decoder := json.NewDecoder(dataJSONFile)
	decoder.UseNumber()

	var templateData interface{}
	err = decoder.Decode(&templateData)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while decoding the JSON file '%v'", dataJSONFilepath)
	}

	templateAndData := enclaves.NewTemplateAndData(templateFileContents, templateData)
	templateAndDataByDestRelFilepath := make(map[string]*enclaves.TemplateAndData)
	templateAndDataByDestRelFilepath[destRelFilepath] = templateAndData

	filesArtifactUuid, err := enclaveCtx.RenderTemplates(templateAndDataByDestRelFilepath)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred rendering the template file at path '%v' with data in the file at path '%v' to enclave '%v'", templateFilepath, dataJSONFilepath, enclaveId)
	}
	logrus.Infof("Files package UUID: %v", filesArtifactUuid)
	return nil
}

func validateTemplateFileArg(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {
	templateFilepath, err := args.GetNonGreedyArg(templateFilepathArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the template filepath to validate using key '%v'", templateFilepathArgKey)
	}

	if _, err := os.Stat(templateFilepath); err != nil {
		return stacktrace.Propagate(err, "An error occurred verifying that the template file '%v' exists and is readable", templateFilepath)
	}
	return nil
}

func validateDataJSONFileArg(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {
	dataJSONFilepath, err := args.GetNonGreedyArg(dataJSONFilepathArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the data JSON filepath to validate using key '%v'", dataJSONFilepathArgKey)
	}

	dataJSONFileContent, err := os.ReadFile(dataJSONFilepath)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred verifying data JSON '%v' exists and is readable", dataJSONFilepath)
	}

	if !json.Valid(dataJSONFileContent) {
		return stacktrace.NewError("The data file isn't valid JSON")
	}

	return nil
}

func validateDestRelFilePathArg(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {
	destRelFilepath, err := args.GetNonGreedyArg(destRelFilepathArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the destination relative filepath to validate using key '%v'", destRelFilepathArgKey)
	}

	if path.IsAbs(destRelFilepath) {
		return stacktrace.NewError("Expected a relative path but got an absolute path '%v'", destRelFilepath)
	}

	return nil
}
