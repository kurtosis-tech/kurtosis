package docker_compose_tranpsiler

import (
	"errors"
	"fmt"
	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
	"github.com/joho/godotenv"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/service/add"
	"github.com/kurtosis-tech/kurtosis/name_generator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"strconv"
	"strings"
)

const (
	emptyPrivateIpPlaceholder = ""
	cpuToMilliCpuConstant     = 1024
	bytesToMegabytes          = 1024 * 1024
	float64BitWidth           = 64
	readWriteEveryone         = 0666

	// Look for an environment variables file at the package root, and if persent use the values found there
	// to fill out the Compose
	envVarsFilename = ".env"
)

// TODO remove this, and instead use the mainFileName that the user passes in!
var supportedComposeFilenames = []string{
	"compose.yml",
	"compose.yaml",
	"docker-compose.yml",
	"docker-compose.yaml",
	"docker_compose.yml",
	"docker_compose.yaml",
}

// TODO actually take in a Compose file
func TranspileDockerComposePackageToStarlark(packageAbsDirpath string) (string, error) {
	// Useful for logging, to not leak internals of APIC
	composeFilename, composeBytes, err := getComposeFilenameAndContent(packageAbsDirpath)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred reading the Compose file")
	}

	// Use the envvars file next to the Compose if it exists
	envVarsFilepath := path.Join(packageAbsDirpath, envVarsFilename)
	var envVars map[string]string
	envVarsInFile, err := godotenv.Read(envVarsFilepath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return "", stacktrace.Propagate(err, "Failed to transpile Docker Compose package to Starlark; a %v file was detected in the package, but an error occurred reading", envVarsFilename)
		}
		envVarsInFile = map[string]string{}
	}
	envVars = envVarsInFile

	script, err := convertComposeToStarlark(composeBytes, envVars)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred transpiling Compose file '%v' to Starlark", composeFilename)
	}
	return script, nil
}

// ====================================================================================================
//                                   Private Helper Functions
// ====================================================================================================

func getComposeFilenameAndContent(packageAbsDirpath string) (string, []byte, error) {
	for _, composeFilename := range supportedComposeFilenames {
		composeFilepath := path.Join(packageAbsDirpath, composeFilename)
		composeBytes, err := os.ReadFile(composeFilepath)
		if err != nil {
			continue
		}

		return composeFilename, composeBytes, nil
	}

	joinedComposeFilenames := strings.Join(supportedComposeFilenames, ", ")
	return "", nil, stacktrace.NewError("Failed to transpile Docker Compose package to Starlark because no Compose file was found at the package root after looking for the following files: %s", joinedComposeFilenames)
}

// TODO(victor.colombo): Have a better UX letting people know ports have been remapped
// NOTE: This returns Go errors, not
func convertComposeToStarlark(composeBytes []byte, envVars map[string]string) (string, error) {
	project, err := loader.Load(types.ConfigDetails{ //nolint:exhaustruct
		// Note that we might be able to use the WorkingDir property instead, to parse the entire directory
		ConfigFiles: []types.ConfigFile{{
			Content: composeBytes,
		}},
		Environment: envVars,
	})
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred parsing the Compose file in preparation for Stalrark transpilation")
	}

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
