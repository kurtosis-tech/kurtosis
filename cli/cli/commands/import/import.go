package _import

import (
	"context"
	"fmt"
	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
	"github.com/joho/godotenv"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	enclave_consts "github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/enclave"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/file_system_path_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/service/add"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/name_generator"
	metrics_client "github.com/kurtosis-tech/metrics-library/golang/lib/client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

const (
	enclaveNameFlagKey        = "enclave"
	pathArgKey                = "file-path"
	dotEnvPathFlagKey         = "env"
	isPathArgOptional         = false
	defaultPathArg            = ""
	defaultDotEnvPathFlag     = ".env"
	emptyPrivateIpPlaceholder = ""
	defaultMainFunction       = ""
	noStarlarkParams          = "{}"
	cpuToMilliCpuConstant     = 1024
	bytesToMegabytes          = 1024 * 1024
	float64BitWidth           = 64

	// Signifies that an enclave name should be auto-generated
	autogenerateEnclaveNameKeyword = ""

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"
)

var ImportCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.ImportCmdStr,
	ShortDescription:          "Import external workflows into Kurtosis",
	LongDescription:           "Import external workflow into Kurtosis (currently only supports Docker Compose)",
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	Flags: []*flags.FlagConfig{
		{
			Key:       enclaveNameFlagKey,
			Shorthand: "n",
			Default:   autogenerateEnclaveNameKeyword,
			Usage: fmt.Sprintf(
				"The enclave name to give the new enclave, which must match regex '%v' "+
					"(emptystring will autogenerate an enclave name)",
				enclave_consts.AllowedEnclaveNameCharsRegexStr,
			),
			Type: flags.FlagType_String,
		},
		{
			Key:       dotEnvPathFlagKey,
			Shorthand: "e",
			Default:   defaultDotEnvPathFlag,
			Usage:     "The .env file path to be loaded into docker compose",
			Type:      flags.FlagType_String,
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
	RunFunc: run,
}

func run(
	ctx context.Context,
	_ backend_interface.KurtosisBackend,
	_ kurtosis_engine_rpc_api_bindings.EngineServiceClient,
	_ metrics_client.MetricsClient,
	flags *flags.ParsedFlags,
	args *args.ParsedArgs) error {
	path, err := args.GetNonGreedyArg(pathArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "Path arg '%v' is missing", pathArgKey)
	}

	dotEnvPath, err := flags.GetString(dotEnvPathFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Dot env path arg '%v' is missing", dotEnvPath)
	}

	dotEnvMap, err := godotenv.Read(dotEnvPath)
	if err != nil {
		logrus.Debugf("No dotenv file was found: %v", err)
		dotEnvMap = map[string]string{}
	}
	logrus.Infof("Enviroment loaded: %v", dotEnvMap)

	script, artifacts, err := convertComposeFileToStarlark(path, dotEnvMap)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to convert compose to starlark")
	}
	// TODO(victor.colombo): Make this as pretty as run is
	logrus.Debugf("Generated starlark:\n%s", script)

	enclaveName, err := flags.GetString(enclaveNameFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Couldn't find enclave name flag '%v'", enclaveNameFlagKey)
	}
	enclaveCtx, err := createEnclave(ctx, enclaveName)
	if err != nil {
		return stacktrace.Propagate(err, "Couldn't create enclave")
	}
	defer output_printers.PrintEnclaveName(enclaveCtx.GetEnclaveName())
	err = uploadArtifacts(enclaveCtx, artifacts)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to upload all required artifacts for execution")
	}
	err = runStarlark(ctx, enclaveCtx, script)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to run generated starlark from compose")
	}
	return nil
}

func convertComposeFileToStarlark(path string, dotEnvMap map[string]string) (string, map[string]string, error) {
	project, err := loader.Load(types.ConfigDetails{ //nolint:exhaustruct
		ConfigFiles: []types.ConfigFile{{Filename: path}},
		Environment: dotEnvMap,
	})
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "Error parsing docker compose")
	}
	script, artifacts, err := convertComposeProjectToStarlark(project)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "Error translating docker compose to Starlark")
	}
	return script, artifacts, nil
}

func uploadArtifacts(enclaveCtx *enclaves.EnclaveContext, artifactUploadMap map[string]string) error {
	for source, artifactName := range artifactUploadMap {
		_, _, err := enclaveCtx.UploadFiles(source, artifactName)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to upload path '%v' as artifact '%s'", source, artifactName)
		}
	}
	return nil
}

// TODO(victor.colombo): Have a better UX letting people know ports have been remapped
func convertComposeProjectToStarlark(compose *types.Project) (string, map[string]string, error) {
	serviceStarlarks := map[string]string{}
	requiredFileUploads := map[string]string{}
	for _, serviceConfig := range compose.Services {
		artifactsPiecesStr := []string{}
		for _, volume := range serviceConfig.Volumes {
			if volume.Type != types.VolumeTypeBind {
				return "", nil, stacktrace.NewError("Volume type '%v' is not supported", volume.Type)
			}
			if _, ok := requiredFileUploads[volume.Source]; !ok {
				requiredFileUploads[volume.Source] = name_generator.GenerateNatureThemeNameForFileArtifacts()
			}
			artifactsPiecesStr = append(artifactsPiecesStr, fmt.Sprintf("%s:%s", volume.Target, requiredFileUploads[volume.Source]))
		}
		portPiecesStr := []string{}
		for _, port := range serviceConfig.Ports {
			portStr := fmt.Sprintf("docker-%s=%d", port.Published, port.Target)
			if port.Protocol != "" {
				portStr += fmt.Sprintf("/%s", port.Protocol)
			}
			portPiecesStr = append(portPiecesStr, portStr)
		}
		envvarsPiecesStr := []string{}
		for envKey, envValue := range serviceConfig.Environment {
			envValueStr := ""
			if envValue != nil {
				envValueStr = *envValue
			}
			envvarsPiecesStr = append(envvarsPiecesStr, fmt.Sprintf("%s=%s", envKey, envValueStr))
		}
		memMinLimit := getMemoryMegabytesReservation(serviceConfig.Deploy)
		cpuMinLimit := getMilliCpusReservation(serviceConfig.Deploy)
		starlarkConfig, err := add.GetServiceConfigStarlark(
			serviceConfig.Image,
			strings.Join(portPiecesStr, ","),
			serviceConfig.Command,
			serviceConfig.Entrypoint,
			strings.Join(envvarsPiecesStr, ","),
			strings.Join(artifactsPiecesStr, ","),
			0,
			0,
			cpuMinLimit,
			memMinLimit,
			emptyPrivateIpPlaceholder)
		if err != nil {
			return "", nil, stacktrace.Propagate(err, "Error getting service config starlark for '%v'", serviceConfig)
		}
		serviceStarlarks[serviceConfig.Name] = starlarkConfig
	}
	script := "def run(plan):\n"
	for serviceName, serviceConfig := range serviceStarlarks {
		script += fmt.Sprintf("\tplan.add_service(name = '%s', config = %s)\n", serviceName, serviceConfig)
	}
	return script, requiredFileUploads, nil
}

func getMemoryMegabytesReservation(deployConfig *types.DeployConfig) int {
	if deployConfig == nil {
		return 0
	}
	reservation := 0
	if deployConfig.Resources.Reservations != nil {
		reservation = int(deployConfig.Resources.Reservations.MemoryBytes) / bytesToMegabytes
		logrus.Debugf("Converted '%v' bytes to '%v' megabytes", deployConfig.Resources.Reservations.MemoryBytes, reservation)
	}
	return reservation
}

func getMilliCpusReservation(deployConfig *types.DeployConfig) int {
	if deployConfig == nil {
		return 0
	}
	reservation := 0
	if deployConfig.Resources.Reservations != nil {
		reservationParsed, err := strconv.ParseFloat(deployConfig.Resources.Reservations.NanoCPUs, float64BitWidth)
		if err == nil {
			// Despite being called 'nano CPUs', they actually refer to a float representing percentage of one CPU
			reservation = int(reservationParsed * cpuToMilliCpuConstant)
			logrus.Debugf("Converted '%v' CPUs to '%v' milli CPUs", deployConfig.Resources.Reservations.NanoCPUs, reservation)
		} else {
			logrus.Warnf("Could not convert CPU reservation '%v' to integer, limits reservation", deployConfig.Resources.Reservations.NanoCPUs)
		}
	}
	return reservation
}

func createEnclave(ctx context.Context, enclaveName string) (*enclaves.EnclaveContext, error) {
	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred connecting to the local Kurtosis engine")
	}

	enclaveCtx, err := kurtosisCtx.CreateEnclave(ctx, enclaveName, false)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating an enclave '%v'", enclaveName)
	}

	return enclaveCtx, nil
}

// TODO(victor.colombo): This should be part of the SDK, since we implement this over and over again
func runStarlark(ctx context.Context, enclaveCtx *enclaves.EnclaveContext, starlarkScript string) error {
	starlarkRunResult, err := enclaveCtx.RunStarlarkScriptBlocking(ctx, defaultMainFunction, starlarkScript, noStarlarkParams, false, 1, []kurtosis_core_rpc_api_bindings.KurtosisFeatureFlag{})
	if err != nil {
		return stacktrace.Propagate(err, "An error has occurred when running Starlark to add service")
	}
	// TODO(victor.colombo): Make this as pretty as run is
	logrus.Infof("Enclave was built with following output:\n%s", starlarkRunResult.RunOutput)
	return nil
}
