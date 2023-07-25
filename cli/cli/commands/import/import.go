package _import

import (
	"context"
	"fmt"
	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	enclave_consts "github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/enclave"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/file_system_path_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/service/add"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"strings"
)

const (
	enclaveNameFlagKey        = "enclave"
	pathArgKey                = "file-path"
	isPathArgOptional         = false
	defaultPathArg            = ""
	emptyPrivateIpPlaceholder = ""
	nonSupportedField         = ""
	defaultMainFunction       = ""
	noStarlarkParams          = "{}"

	// Signifies that an enclave name should be auto-generated
	autogenerateEnclaveNameKeyword = ""
)

var ImportCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:       command_str_consts.ImportCmdStr,
	ShortDescription: "Import external workflows into Kurtosis",
	LongDescription:  "Import external workflow into Kurtosis (currently only supports Docker Compose)",
	Flags: []*flags.FlagConfig{
		{
			Key:       enclaveNameFlagKey,
			Shorthand: "e",
			Default:   autogenerateEnclaveNameKeyword,
			Usage: fmt.Sprintf(
				"The enclave name to give the new enclave, which must match regex '%v' "+
					"(emptystring will autogenerate an enclave name)",
				enclave_consts.AllowedEnclaveNameCharsRegexStr,
			),
			Type: flags.FlagType_String,
		},
	},
	Args: []*args.ArgConfig{
		file_system_path_arg.NewFilepathOrDirpathArg(
			pathArgKey,
			isPathArgOptional,
			defaultPathArg,
			file_system_path_arg.DefaultValidationFunc,
		),
	},
	PreValidationAndRunFunc:  nil,
	RunFunc:                  run,
	PostValidationAndRunFunc: nil,
}

func run(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {
	path, err := args.GetNonGreedyArg(pathArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "Path arg '%v' is missing", pathArgKey)
	}
	file, err := os.Open(path)
	if err != nil {
		return stacktrace.Propagate(err, "File on '%v' was not found", path)
	}
	defer file.Close()

	// Read the content of the file into a []byte slice
	content, err := io.ReadAll(file)
	if err != nil {
		return stacktrace.Propagate(err, "Error reading file: %s\n", err)
	}

	script, err := convertComposeFileToStarlark(content)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to convert compose to starlark")
	}
	// TODO(victor.colombo): Make this as pretty as run is
	logrus.Debugf("Generated starlark:\n%s", script)

	enclaveName, err := flags.GetString(enclaveNameFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Couldn't find enclave name flag '%v'", enclaveNameFlagKey)
	}
	err = runStarlark(ctx, enclaveName, script)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to run generated starlark from compose")
	}
	return nil
}

func convertComposeFileToStarlark(content []byte) (string, error) {
	// Load the Compose configuration from the []byte slice
	config, err := loader.ParseYAML(content)
	if err != nil {
		return "", stacktrace.Propagate(err, "Error parsing YAML: %s\n", err)
	}

	// Convert the generic map to a structured compose.Config
	project, err := loader.Load(types.ConfigDetails{ //nolint:exhaustruct
		ConfigFiles: []types.ConfigFile{{Config: config}},
	})
	if err != nil {
		return "", stacktrace.Propagate(err, "Error parsing docker compose")
	}
	script, err := convertComposeProjectToStarlark(project)
	if err != nil {
		return "", stacktrace.Propagate(err, "Error translating docker compose to Starlark")
	}
	return script, nil
}

// TODO(victor.colombo): Have a better UX letting people know ports have been remapped
func convertComposeProjectToStarlark(compose *types.Project) (string, error) {
	serviceStarlarks := map[string]string{}
	for _, serviceConfig := range compose.Services {
		portPiecesStr := []string{}
		for _, port := range serviceConfig.Ports {
			portStr := fmt.Sprintf("docker-%s=%d", port.Published, port.Target)
			if port.Protocol != "" {
				portStr += fmt.Sprintf("/%s", port.Protocol)
			}
			portPiecesStr = append(portPiecesStr, portStr)
		}
		starlarkConfig, err := add.GetServiceConfigStarlark(serviceConfig.Image, strings.Join(portPiecesStr, ","), serviceConfig.Command, serviceConfig.Entrypoint, nonSupportedField, nonSupportedField, emptyPrivateIpPlaceholder)
		if err != nil {
			return "", stacktrace.Propagate(err, "Error getting service config starlark for '%v'", serviceConfig)
		}
		serviceStarlarks[serviceConfig.Name] = starlarkConfig
	}
	script := "def run(plan):\n"
	for serviceName, serviceConfig := range serviceStarlarks {
		script += fmt.Sprintf("\tplan.add_service(name = '%s', config = %s)\n", serviceName, serviceConfig)
	}
	return script, nil
}

// TODO(victor.colombo): This should be part of the SDK, since we implement this over and over again
func runStarlark(ctx context.Context, enclaveName string, starlarkScript string) error {
	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred connecting to the local Kurtosis engine")
	}

	enclaveCtx, err := kurtosisCtx.CreateEnclave(ctx, enclaveName, false)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating an enclave '%v'", enclaveName)
	}
	defer output_printers.PrintEnclaveName(enclaveName)

	starlarkRunResult, err := enclaveCtx.RunStarlarkScriptBlocking(ctx, defaultMainFunction, starlarkScript, noStarlarkParams, false, 1, []kurtosis_core_rpc_api_bindings.KurtosisFeatureFlag{})
	if err != nil {
		return stacktrace.Propagate(err, "An error has occurred when running Starlark to add service")
	}
	// TODO(victor.colombo): Make this as pretty as run is
	logrus.Infof("Enclave was built with following output:\n%s", starlarkRunResult.RunOutput)
	return nil
}
