package docker_compose_transpiler

import (
	"errors"
	"fmt"
	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
	"github.com/joho/godotenv"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/directory"
	port_spec_starlark "github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/port_spec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
	"os"
	"path"
	"sort"
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

	builtImageSuffx = "-image"
)

var dockerPortProtosToKurtosisPortProtos = map[string]port_spec.TransportProtocol{
	"tcp":  port_spec.TransportProtocol_TCP,
	"udp":  port_spec.TransportProtocol_UDP,
	"sctp": port_spec.TransportProtocol_SCTP,
}

// TODO Make this return an interpretation error?
func TranspileDockerComposePackageToStarlark(packageAbsDirpath string, composeRelativeFilepath string) (string, error) {
	composeAbsFilepath := path.Join(packageAbsDirpath, composeRelativeFilepath)

	// Useful for logging, to not leak internals of APIC
	composeFilename := path.Base(composeRelativeFilepath)

	composeBytes, err := os.ReadFile(composeAbsFilepath)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred reading Compose file '%v'", composeFilename)
	}

	// Use the env vars file next to the Compose if it exists
	envVarsFilepath := path.Join(packageAbsDirpath, envVarsFilename)
	var envVars map[string]string
	envVarsInFile, err := godotenv.Read(envVarsFilepath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return "", stacktrace.Propagate(err, "Failed to transpile Docker Compose package to Starlark; a %v file was detected in the package, but an error occurred reading it.", envVarsFilename)
		}
		envVarsInFile = map[string]string{}
	}
	envVars = envVarsInFile

	starlarkScript, err := convertComposeToStarlark(composeBytes, envVars)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred transpiling Compose file '%v' to Starlark", composeFilename)
	}
	return starlarkScript, nil
}

// ====================================================================================================
//
//	Private Helper Functions
//
// ====================================================================================================
func convertComposeToStarlark(composeBytes []byte, envVars map[string]string) (string, error) {
	// Convert composeBytes into compose data structure
	composeParseConfig := types.ConfigDetails{ //nolint:exhaustruct
		// Note that we might be able to use the WorkingDir property instead, to parse the entire directory
		ConfigFiles: []types.ConfigFile{{
			Content: composeBytes,
		}},
		Environment: envVars,
	}
	setOptionsFunc := func(options *loader.Options) {
		options.SetProjectName(composeProjectName, shouldOverrideComposeYamlKeyProjectName)

		// We don't want to resolve paths; these should get resolved by our package content provider instead
		options.ResolvePaths = false

		// We do want to convert Windows paths into Linux paths, since the APIC runs on Linux
		options.ConvertWindowsPaths = true
	}
	compose, err := loader.Load(composeParseConfig, setOptionsFunc)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred parsing the Compose file in preparation for Starlark transpilation")
	}

	// Convert each Compose Service into corresponding Starlark ServiceConfig object
	perServiceDependencies := map[string]map[string]bool{} // Mapping of services -> services they depend on
	perServiceLines := map[string][]string{}               // Mapping of services -> lines of Starlark that they need to run
	numTotalServiceLines := 0
	sortedServiceNames := []string{} // List of sorted service names to make sure depends_on is processed  correctly
	for _, serviceConfig := range compose.Services {
		serviceConfigKwargs := []starlark.Tuple{}

		// NAME
		serviceName := serviceConfig.Name
		sortedServiceNames = append(sortedServiceNames, serviceName)

		// IMAGE: use either serviceConfig.Image(docker compose) or ImageBuildSpec
		if serviceConfig.Image != "" {
			serviceConfigKwargs = appendKwarg(
				serviceConfigKwargs,
				service_config.ImageAttr,
				starlark.String(serviceConfig.Image),
			)
		} else {
			// Create ImageBuildSpec
			imageBuildSpecKwargs, err := getImageBuildSpecKwargs(serviceName, serviceConfig)
			if err != nil {
				return "", err
			}
			imageBuildSpecArgumentValuesSet, interpretationErr := builtin_argument.CreateNewArgumentValuesSet(
				service_config.ImageBuildSpecTypeName,
				service_config.NewImageBuildSpecType().KurtosisBaseBuiltin.Arguments,
				[]starlark.Value{},
				imageBuildSpecKwargs,
			)
			if interpretationErr != nil {
				// TODO: interpretation err vs. golang err
				return "", interpretationErr
			}
			imageBuildSpecKurtosisType, interpretationErr := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(service_config.ImageBuildSpecTypeName, imageBuildSpecArgumentValuesSet)
			if interpretationErr != nil {
				// TODO: interpretation err vs. golang err
				return "", interpretationErr
			}
			serviceConfigKwargs = appendKwarg(
				serviceConfigKwargs,
				service_config.ImageAttr,
				imageBuildSpecKurtosisType,
			)
		}

		// PORTS
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
				service_config.EntrypointAttr,
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
				service_config.CmdAttr,
				starlark.NewList(commandSLStrs),
			)
		}

		// ENV VARS
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
					return "", stacktrace.Propagate(err, "An error occurred setting key '%s' in environment variables Starlark dict for service '%s'", key, serviceName)
				}
			}

			serviceConfigKwargs = appendKwarg(
				serviceConfigKwargs,
				service_config.EnvVarsAttr,
				enVarsSLDict,
			)
		}

		// VOLUMES -> Files Artifacts
		pathsToUpload := make(map[string]string) // Mapping of relative_path_to_upload -> files_artifact_name
		if serviceConfig.Volumes != nil {
			filesArgSLDict := starlark.NewDict(len(serviceConfig.Volumes))

			for volumeIdx, volume := range serviceConfig.Volumes {
				volumeType := volume.Type

				var shouldPersist bool
				switch volumeType {
				case types.VolumeTypeBind:
					source := volume.Source

					// We guess that when the user specifies an absolute (not relative) path, they want to use the volume
					// as a persistence layer. We further guess that relative paths are just read-only.
					shouldPersist = path.IsAbs(source)
				case types.VolumeTypeVolume:
					shouldPersist = true
				}

				var filesDictValue starlark.Value
				if shouldPersist {
					persistenceKey := fmt.Sprintf("volume%d", volumeIdx)
					persistentDirectory, err := createPersistentDirectoryKurtosisType(persistenceKey)
					if err != nil {
						return "", stacktrace.Propagate(err, "An error occurred creating persistent directory with key '%s' for volume #%d on service '%s'", persistenceKey, volumeIdx, serviceName)
					}
					filesDictValue = persistentDirectory
				} else {
					// If not persistent, do an upload_files
					filesArtifactName := fmt.Sprintf("%s--volume%d", serviceName, volumeIdx)
					pathsToUpload[volume.Source] = filesArtifactName
					filesDictValue = starlark.String(filesArtifactName)
				}

				if err := filesArgSLDict.SetKey(
					starlark.String(volume.Target),
					filesDictValue,
				); err != nil {
					return "", stacktrace.Propagate(err, "An error occurred setting volume mountpoint '%s' in the files Starlark dict for service '%s'", volume.Target, serviceName)
				}
			}

			serviceConfigKwargs = appendKwarg(
				serviceConfigKwargs,
				service_config.FilesAttr,
				filesArgSLDict,
			)
		}

		// DEPENDS ON
		// TODO(kevin): handle the dependencyType
		dependencyServiceNames := map[string]bool{}
		for dependencyName := range serviceConfig.DependsOn {
			dependencyServiceNames[dependencyName] = true
		}
		perServiceDependencies[serviceName] = dependencyServiceNames

		// TODO uncomment (why was it commented?)
		// CPU allocation?
		//memMinLimit := getMemoryMegabytesReservation(serviceConfig.Deploy)
		//cpuMinLimit := getMilliCpusReservation(serviceConfig.Deploy)
		//memory allocations?

		// Whats
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
		serviceConfigStr := serviceConfigKurtosisType.String()

		linesForService := []string{}

		for relativePath, filesArtifactName := range pathsToUpload {
			// TODO SWITCH FROM HARDCODING THESE TO DYNAMIC CONSTS
			uploadFilesLine := fmt.Sprintf("plan.upload_files(src = \"%s\", name = \"%s\")", relativePath, filesArtifactName)
			linesForService = append(linesForService, uploadFilesLine)
		}

		// TODO SWITCH FROM HARDCODING THESE TO DYNAMIC CONSTS
		addServiceLine := fmt.Sprintf("plan.add_service(name = \"%s\", config = %s)", serviceName, serviceConfigStr)
		linesForService = append(linesForService, addServiceLine)

		perServiceLines[serviceName] = linesForService
		numTotalServiceLines += len(linesForService)
	}

	sort.Strings(sortedServiceNames)

	// TODO(kevin) SWITCH THIS TO BE A PROPER DAG!!! This doesn't catch circular dependencies
	// This is a super janky, inefficient (but deterministic) topological sort
	starlarkLines := make([]string, 0, numTotalServiceLines)
	alreadyProcessedServices := map[string]bool{} // "Set" of service lines that we've already written
	for len(alreadyProcessedServices) < len(perServiceLines) {
		// Important to iterate over the sorted version, to have a deterministic topological sort
		var serviceToProcess string
		for _, serviceName := range sortedServiceNames {
			//
			if _, found := alreadyProcessedServices[serviceName]; found {
				continue
			}

			// Check if all dependencies have already been processed
			allDependenciesProcessed := true
			for dependencyName := range perServiceDependencies[serviceName] {
				if _, found := alreadyProcessedServices[dependencyName]; !found {
					allDependenciesProcessed = false
					break
				}
			}
			if !allDependenciesProcessed {
				continue
			}

			// We've found a service that can be processed now
			serviceToProcess = serviceName
			break
		}

		linesForService := perServiceLines[serviceToProcess]
		starlarkLines = append(starlarkLines, linesForService...)
		alreadyProcessedServices[serviceToProcess] = true
	}

	// TODO SWITCH FROM HARDCODING THESE TO DYNAMIC CONSTS
	script := "def run(plan):\n"
	for _, line := range starlarkLines {
		script += fmt.Sprintf("    %s\n", line)
	}
	return script, nil
}

func getImageBuildSpecKwargs(
	serviceName string,
	serviceConfig types.ServiceConfig,
) ([]starlark.Tuple, error) {
	var imageBuildSpecKwargs []starlark.Tuple

	builtImageName := serviceName + builtImageSuffx
	imageNameKwarg := []starlark.Value{
		starlark.String(service_config.BuiltImageNameAttr),
		starlark.String(builtImageName),
	}
	imageBuildSpecKwargs = append(imageBuildSpecKwargs, imageNameKwarg)

	if serviceConfig.Build != nil && serviceConfig.Build.Context != "" {
		contextDirKwarg := []starlark.Value{
			starlark.String(service_config.BuildContextAttr),
			starlark.String(serviceConfig.Build.Context),
		}
		imageBuildSpecKwargs = append(imageBuildSpecKwargs, contextDirKwarg)
	}

	if serviceConfig.Build != nil && serviceConfig.Build.Target != "" {
		targetStageKwarg := []starlark.Value{
			starlark.String(service_config.TargetStageAttr),
			starlark.String(serviceConfig.Build.Target),
		}
		imageBuildSpecKwargs = append(imageBuildSpecKwargs, targetStageKwarg)
	}

	return imageBuildSpecKwargs, nil
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

func createPersistentDirectoryKurtosisType(persistenceKey string) (starlark.Value, error) {
	directoryKwargs := []starlark.Tuple{}

	directoryKwargs = appendKwarg(
		directoryKwargs,
		directory.PersistentKeyAttr,
		starlark.String(persistenceKey),
	)

	argumentValuesSet, interpretationErr := builtin_argument.CreateNewArgumentValuesSet(
		directory.DirectoryTypeName,
		directory.NewDirectoryType().KurtosisBaseBuiltin.Arguments,
		[]starlark.Value{},
		directoryKwargs,
	)
	if interpretationErr != nil {
		// TODO HANDLE THIS! interpretionerror vs go error
		return nil, interpretationErr
	}

	directoryKurtosisType, interpretationErr := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(directory.DirectoryTypeName, argumentValuesSet)
	if interpretationErr != nil {
		// TODO FIX THIS! INTERPRETATION ERROR VS GO ERROR
		return nil, interpretationErr
	}

	return directoryKurtosisType, nil
}
