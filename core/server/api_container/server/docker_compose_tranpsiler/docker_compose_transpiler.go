package docker_compose_tranpsiler

import (
	"errors"
	"fmt"
	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
	"github.com/joho/godotenv"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	port_spec_starlark "github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/port_spec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
	"os"
	"path"
	"strconv"
	"strings"
)

const (
	cpuToMilliCpuConstant = 1024
	bytesToMegabytes      = 1024 * 1024
	float64BitWidth       = 64

	// Look for an environment variables file at the package root, and if present use the values found there
	// to fill out the Compose
	envVarsFilename = ".env"

	// Every Compose project needs a project name
	// This is the one we give by default, but we let it be overridden if the Compose file specifies a 'name' stanza
	composeProjectName = "root-compose-project"

	// Our project name should cede to the project name in the Compose
	shouldOverrideComposeYamlKeyProjectName = false
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

var dockerPortProtosToKurtosisPortProtos = map[string]port_spec.TransportProtocol{
	"tcp":  port_spec.TransportProtocol_TCP,
	"udp":  port_spec.TransportProtocol_UDP,
	"sctp": port_spec.TransportProtocol_SCTP,
}

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
	composeParseConfig := types.ConfigDetails{ //nolint:exhaustruct
		// Note that we might be able to use the WorkingDir property instead, to parse the entire directory
		ConfigFiles: []types.ConfigFile{{
			Content: composeBytes,
		}},
		Environment: envVars,
	}

	setProjectNameOpt := func(options *loader.Options) {
		options.SetProjectName(composeProjectName, shouldOverrideComposeYamlKeyProjectName)
	}

	compose, err := loader.Load(composeParseConfig, setProjectNameOpt)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred parsing the Compose file in preparation for Starlark transpilation")
	}

	serviceStarlarks := map[string]string{}
	for _, serviceConfig := range compose.Services {
		serviceName := serviceConfig.Name

		serviceConfigKwargs := []starlark.Tuple{}

		/*
			artifactsPiecesStr := []string{}
			for volumeIdx, volume := range serviceConfig.Volumes {
				if volume.Type != types.VolumeTypeBind {
					return "", stacktrace.NewError("Volume #%v on service '%v' has type '%v', which isn't supported", serviceName, volumeIdx, volume.Type)
				}
				if _, ok := requiredFileUploads[volume.Source]; !ok {
					requiredFileUploads[volume.Source] = name_generator.GenerateNatureThemeNameForFileArtifacts()
				}
				artifactsPiecesStr = append(artifactsPiecesStr, fmt.Sprintf("%s:%s", volume.Target, requiredFileUploads[volume.Source]))
			}

		*/

		// Image
		serviceConfigKwargs = appendKwarg(
			serviceConfigKwargs,
			service_config.ImageAttr,
			starlark.String(serviceConfig.Image),
		)

		// Ports
		portSpecsSLDict, err := getPortSpecsSLDict(serviceConfig)
		if err != nil {
			return "", stacktrace.Propagate(err, "An error occurred creating the port specs dict for service '%s'", serviceName)
		}
		serviceConfigKwargs = appendKwarg(
			serviceConfigKwargs,
			service_config.PortsAttr,
			portSpecsSLDict,
		)

		// ENTRYPOINT
		if serviceConfig.Entrypoint != nil {
			entrypointSLStrs := make([]starlark.Value, len(serviceConfig.Entrypoint))
			for idx, entrypointFragment := range serviceConfig.Entrypoint {
				entrypointSLStrs[idx] = starlark.String(entrypointFragment)
			}

			serviceConfigKwargs = appendKwarg(
				serviceConfigKwargs,
				service_config.EnvVarsAttr,
				starlark.NewList(entrypointSLStrs),
			)
		}

		// CMD
		if serviceConfig.Command != nil {
			commandSLStrs := make([]starlark.Value, len(serviceConfig.Command))
			for idx, commandFragment := range serviceConfig.Command {
				commandSLStrs[idx] = starlark.String(commandFragment)
			}

			serviceConfigKwargs = appendKwarg(
				serviceConfigKwargs,
				service_config.EnvVarsAttr,
				starlark.NewList(commandSLStrs),
			)
		}

		// Env vars
		if serviceConfig.Environment != nil {
			enVarsSLDict := starlark.NewDict(len(serviceConfig.Environment))
			for key, value := range serviceConfig.Environment {
				if value == nil {
					continue
				}
				if err := enVarsSLDict.SetKey(
					starlark.String(key),
					starlark.String(*value),
				); err != nil {
					return "", stacktrace.Propagate(err, "An error occurred setting key '%s' in environment variables Starlark dict", key)
				}
			}

			serviceConfigKwargs = appendKwarg(
				serviceConfigKwargs,
				service_config.EnvVarsAttr,
				enVarsSLDict,
			)
		}

		// TODO uncomment
		/*
			memMinLimit := getMemoryMegabytesReservation(serviceConfig.Deploy)
			cpuMinLimit := getMilliCpusReservation(serviceConfig.Deploy)
		*/

		argumentValuesSet, interpretationErr := builtin_argument.CreateNewArgumentValuesSet(
			service_config.ServiceConfigTypeName,
			service_config.NewServiceConfigType().KurtosisBaseBuiltin.Arguments,
			[]starlark.Value{},
			serviceConfigKwargs,
		)
		if interpretationErr != nil {
			// TODO HANDLE THIS! interpretionerror vs go error
			return "", interpretationErr
		}

		serviceConfigKurtosisType, interpretationErr := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(service_config.ServiceConfigTypeName, argumentValuesSet)
		if interpretationErr != nil {
			// TODO HANDLE THIS! interpretionerror vs go error
			return "", interpretationErr
		}

		serviceStarlarks[serviceName] = serviceConfigKurtosisType.String()
	}

	script := "def run(plan):\n"
	for serviceName, serviceConfig := range serviceStarlarks {
		script += fmt.Sprintf("    plan.add_service(name = '%s', config = %s)\n", serviceName, serviceConfig)
	}
	return script, nil
}

func getPortSpecsSLDict(
	serviceConfig types.ServiceConfig,
) (*starlark.Dict, error) {
	portSpecs := starlark.NewDict(len(serviceConfig.Ports))
	for portIdx, dockerPort := range serviceConfig.Ports {
		portName := fmt.Sprintf("port%d", portIdx)

		dockerProto := dockerPort.Protocol
		kurtosisProto, found := dockerPortProtosToKurtosisPortProtos[strings.ToLower(dockerProto)]
		if !found {
			return nil, stacktrace.NewError("Port #%d has unsupported protocol '%v'", portIdx, dockerProto)
		}

		portSpec, interpretationErr := port_spec_starlark.CreatePortSpecUsingGoValues(
			uint16(dockerPort.Target),
			kurtosisProto,
			nil, // Application protocol (which Compose doesn't have). Maybe we could guess it in the future?
			"",  // Wait timeout (Compose doesn't have a way to override this)
		)
		if interpretationErr != nil {
			logrus.Debugf(
				"Interpretation error that occurred when creating a %s object from port #%d:\n%s",
				port_spec_starlark.PortSpecTypeName,
				portIdx,
				interpretationErr.Error(),
			)
			return nil, stacktrace.NewError(
				"An error occurred creating a %s object from port #%d",
				port_spec_starlark.PortSpecTypeName,
				portIdx,
			)
		}
		if err := portSpecs.SetKey(
			starlark.String(portName),
			portSpec,
		); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred putting port #%d in Starlark dict", portIdx)
		}

		// TODO public ports??
	}

	return portSpecs, nil
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

func appendKwarg(kwargs []starlark.Tuple, argName string, argValue starlark.Value) []starlark.Tuple {
	tuple := []starlark.Value{
		starlark.String(argName),
		argValue,
	}
	return append(kwargs, tuple)
}
