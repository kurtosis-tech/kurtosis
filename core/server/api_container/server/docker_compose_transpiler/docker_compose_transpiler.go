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

	builtImageSuffix = "-image"

	// eg. plan.upload_files(src = "./data/project", name = "web-volume0")
	uploadFilesLinesFmtStr = "plan.upload_files(src = \"%s\", name = \"%s\")"

	// eg. plan.add_service(name="web", config=ServiceConfig(...))
	addServiceLinesFmtStr = "plan.add_service(name = \"%s\", config = %s)"

	defRunStr = "def run(plan):\n"

	newStarlarkLineFmtStr = "   %s\n"
)

type ComposeService types.ServiceConfig

var dockerPortProtosToKurtosisPortProtos = map[string]port_spec.TransportProtocol{
	"tcp":  port_spec.TransportProtocol_TCP,
	"udp":  port_spec.TransportProtocol_UDP,
	"sctp": port_spec.TransportProtocol_SCTP,
}

// TODO: Make this return an interpretation error
func TranspileDockerComposePackageToStarlark(packageAbsDirpath string, composeRelativeFilepath string) (string, error) {
	composeAbsFilepath := path.Join(packageAbsDirpath, composeRelativeFilepath)

	// Useful for logging to prevent leaking internals of APIC
	composeFilename := path.Base(composeRelativeFilepath)

	composeBytes, err := os.ReadFile(composeAbsFilepath)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred reading Compose file '%v'", composeFilename)
	}

	// Use env vars file next to Compose if it exists
	envVarsFilepath := path.Join(packageAbsDirpath, envVarsFilename)
	var envVars map[string]string
	envVarsInFile, err := godotenv.Read(envVarsFilepath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return "", stacktrace.Propagate(err, "An %v file was found in the package, but an error occurred reading it.", envVarsFilename)
		}
		envVarsInFile = map[string]string{}
	}
	envVars = envVarsInFile

	starlarkScript, err := convertComposeToStarlark(composeBytes, envVars)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred converting Compose file '%v' to a Starlark script.", composeFilename)
	}
	return starlarkScript, nil
}

// ====================================================================================================
//
//	Private Helper Functions
//
// ====================================================================================================
func convertComposeToStarlark(composeBytes []byte, envVars map[string]string) (string, error) {
	composeStruct, err := convertComposeBytesToComposeStruct(composeBytes, envVars)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred converting compose bytes into a struct.")
	}

	serviceNameToStarlarkServiceConfig, perServiceDependencies, pathsToUpload, err := convertComposeServicesToStarlarkServiceConfigs(composeStruct.Services)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred converting compose services to starlark service configs.")
	}

	// Assemble Starlark script
	starlarkLines := []string{}

	// Add upload_files instructions
	for relativePath, filesArtifactName := range pathsToUpload {
		uploadFilesLine := fmt.Sprintf(uploadFilesLinesFmtStr, relativePath, filesArtifactName)
		starlarkLines = append(starlarkLines, uploadFilesLine)
	}

	// Add add_service instructions in an order that respects `depends_on` in Compose
	sortedServices, err := sortServicesBasedOnDependencies(perServiceDependencies)
	if err != nil {
		return "", err // no need to wrap err
	}

	for _, serviceName := range sortedServices {
		starlarkServiceConfig := serviceNameToStarlarkServiceConfig[serviceName]
		addServiceLine := fmt.Sprintf(addServiceLinesFmtStr, serviceName, starlarkServiceConfig.String())
		starlarkLines = append(starlarkLines, addServiceLine)
	}

	script := defRunStr
	for _, line := range starlarkLines {
		script += fmt.Sprintf(newStarlarkLineFmtStr, line)
	}
	return script, nil
}

func convertComposeBytesToComposeStruct(composeBytes []byte, envVars map[string]string) (*types.Project, error) {
	composeParseConfig := types.ConfigDetails{ //nolint:exhaustruct
		// Note that we might be able to use the WorkingDir property instead, to parse the entire directory
		ConfigFiles: []types.ConfigFile{{
			Content: composeBytes,
		}},
		Environment: envVars,
	}
	setOptionsFunc := func(options *loader.Options) {
		options.SetProjectName(composeProjectName, shouldOverrideComposeYamlKeyProjectName)

		// Don't resolve paths as they should be resolved by package content provider
		options.ResolvePaths = false

		// Don't convert Windows paths to Linux paths as APIC runs on Linux
		options.ConvertWindowsPaths = true
	}
	compose, err := loader.Load(composeParseConfig, setOptionsFunc)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred parsing compose based on provided parsing config and set options function.")
	}
	return compose, nil
}

// Turns DockerCompose Service into Kurtosis ServiceConfig + metadata needed for creating starlark script
// Returns:
// Map of service names to Kurtosis Service Configs
// A graph of service dependencies based on depends_on key -> determines order in which to add services
// Map of relative paths to files artifacts names that need to get uploaded -> determines files artifacts that need to be uploaded
func convertComposeServicesToStarlarkServiceConfigs(composeServices types.Services) (
	map[string]*kurtosis_type_constructor.KurtosisValueTypeDefault,
	map[string]map[string]bool,
	map[string]string,
	error) {
	serviceNameToStarlarkServiceConfig := map[string]*kurtosis_type_constructor.KurtosisValueTypeDefault{}
	perServiceDependencies := map[string]map[string]bool{}
	pathsToUpload := map[string]string{}

	for _, service := range composeServices {
		composeService := ComposeService(service)
		serviceConfigKwargs := []starlark.Tuple{}

		// NAME
		serviceName := composeService.Name

		// IMAGE
		if composeService.Image != "" {
			serviceConfigKwargs = appendKwarg(
				serviceConfigKwargs,
				service_config.ImageAttr,
				starlark.String(composeService.Image),
			)
		}

		// IMAGE BUILD SPEC
		if composeService.Build != nil {
			imageBuildSpec, err := getStarlarkImageBuildSpec(composeService.Build, serviceName)
			if err != nil {
				return nil, nil, nil, err
			}
			serviceConfigKwargs = appendKwarg(
				serviceConfigKwargs,
				service_config.ImageAttr,
				imageBuildSpec,
			)
		}

		// PORTS
		if composeService.Ports != nil {
			portSpecsDict, err := getStarlarkPortSpecs(composeService.Ports)
			if err != nil {
				return nil, nil, nil, stacktrace.Propagate(err, "An error occurred creating the port specs dict for service '%s'", serviceName)
			}
			serviceConfigKwargs = appendKwarg(
				serviceConfigKwargs,
				service_config.PortsAttr,
				portSpecsDict,
			)
		}

		// ENTRYPOINT
		if composeService.Entrypoint != nil {
			entrypointList := getStarlarkEntrypoint(composeService.Entrypoint)
			serviceConfigKwargs = appendKwarg(
				serviceConfigKwargs,
				service_config.EntrypointAttr,
				entrypointList,
			)
		}

		// CMD
		if composeService.Command != nil {
			commandList := getStarlarkCommand(composeService.Command)
			serviceConfigKwargs = appendKwarg(
				serviceConfigKwargs,
				service_config.CmdAttr,
				commandList,
			)
		}

		// ENV VARS
		if composeService.Environment != nil {
			envVarsDict, err := getStarlarkEnvVars(composeService.Environment)
			if err != nil {
				return nil, nil, nil, stacktrace.Propagate(err, "An error occurred creating the env vars dict for service '%s'", serviceName)
			}
			serviceConfigKwargs = appendKwarg(
				serviceConfigKwargs,
				service_config.EnvVarsAttr,
				envVarsDict,
			)
		}

		// VOLUMES -> FILES ARTIFACTS
		if composeService.Volumes != nil {
			filesDict, err := getStarlarkFilesArtifacts(composeService.Volumes, serviceName, pathsToUpload)
			if err != nil {
				return nil, nil, nil, stacktrace.Propagate(err, "An error occurred creating the files dict for service '%s'", serviceName)
			}
			serviceConfigKwargs = appendKwarg(
				serviceConfigKwargs,
				service_config.FilesAttr,
				filesDict,
			)
		}

		if composeService.Deploy != nil {
			// MIN MEMORY
			memMinLimit := getStarlarkMinMemory(composeService.Deploy)
			serviceConfigKwargs = appendKwarg(
				serviceConfigKwargs,
				service_config.MinMemoryMegaBytesAttr,
				memMinLimit)

			// MIN CPU
			cpuMinLimit := getStarlarkMinCpus(composeService.Deploy)
			serviceConfigKwargs = appendKwarg(
				serviceConfigKwargs,
				service_config.MinCpuMilliCoresAttr,
				cpuMinLimit)
		}

		// DEPENDS ON
		dependencyServiceNames := map[string]bool{}
		for dependencyName := range composeService.DependsOn {
			dependencyServiceNames[dependencyName] = true
		}
		perServiceDependencies[serviceName] = dependencyServiceNames

		// Finally, create Starlark Service Config object based on kwargs
		argumentValuesSet, interpretationErr := builtin_argument.CreateNewArgumentValuesSet(
			service_config.ServiceConfigTypeName,
			service_config.NewServiceConfigType().KurtosisBaseBuiltin.Arguments,
			[]starlark.Value{},
			serviceConfigKwargs,
		)
		if interpretationErr != nil {
			// TODO HANDLE THIS! interpretionerror vs go error
			return nil, nil, nil, interpretationErr
		}
		serviceConfigKurtosisType, interpretationErr := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(service_config.ServiceConfigTypeName, argumentValuesSet)
		if interpretationErr != nil {
			// TODO HANDLE THIS! interpretionerror vs go error
			return nil, nil, nil, interpretationErr
		}
		serviceNameToStarlarkServiceConfig[serviceName] = serviceConfigKurtosisType
	}

	return serviceNameToStarlarkServiceConfig, perServiceDependencies, pathsToUpload, nil
}

func getStarlarkImageBuildSpec(composeBuild *types.BuildConfig, serviceName string) (starlark.Value, error) {
	var imageBuildSpecKwargs []starlark.Tuple

	builtImageName := serviceName + builtImageSuffix
	imageNameKwarg := []starlark.Value{
		starlark.String(service_config.BuiltImageNameAttr),
		starlark.String(builtImageName),
	}
	imageBuildSpecKwargs = append(imageBuildSpecKwargs, imageNameKwarg)
	if composeBuild.Context != "" {
		contextDirKwarg := []starlark.Value{
			starlark.String(service_config.BuildContextAttr),
			starlark.String(composeBuild.Context),
		}
		imageBuildSpecKwargs = append(imageBuildSpecKwargs, contextDirKwarg)
	}
	if composeBuild.Target != "" {
		targetStageKwarg := []starlark.Value{
			starlark.String(service_config.TargetStageAttr),
			starlark.String(composeBuild.Target),
		}
		imageBuildSpecKwargs = append(imageBuildSpecKwargs, targetStageKwarg)
	}

	imageBuildSpecArgumentValuesSet, interpretationErr := builtin_argument.CreateNewArgumentValuesSet(
		service_config.ImageBuildSpecTypeName,
		service_config.NewImageBuildSpecType().KurtosisBaseBuiltin.Arguments,
		[]starlark.Value{},
		imageBuildSpecKwargs,
	)
	if interpretationErr != nil {
		// TODO: interpretation err vs. golang err
		return nil, interpretationErr
	}
	imageBuildSpecKurtosisType, interpretationErr := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(service_config.ImageBuildSpecTypeName, imageBuildSpecArgumentValuesSet)
	if interpretationErr != nil {
		// TODO: interpretation err vs. golang err
		return nil, interpretationErr
	}
	return imageBuildSpecKurtosisType, nil
}

func getStarlarkPortSpecs(composePorts []types.ServicePortConfig) (*starlark.Dict, error) {
	portSpecs := starlark.NewDict(len(composePorts))

	for portIdx, dockerPort := range composePorts {
		portName := fmt.Sprintf("port%d", portIdx)

		dockerProto := dockerPort.Protocol
		kurtosisProto, found := dockerPortProtosToKurtosisPortProtos[strings.ToLower(dockerProto)]
		if !found {
			return nil, stacktrace.NewError("Port #%d has unsupported protocol '%v'", portIdx, dockerProto)
		}

		portSpec, interpretationErr := port_spec_starlark.CreatePortSpecUsingGoValues(
			uint16(dockerPort.Target),
			kurtosisProto,
			nil, // Application protocol (which Compose doesn't have)
			"",  // Wait timeout (which Compose doesn't have a way to override)
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

func getStarlarkEntrypoint(composeEntrypoint types.ShellCommand) *starlark.List {
	entrypointSLStrs := make([]starlark.Value, len(composeEntrypoint))
	for idx, entrypointFragment := range composeEntrypoint {
		entrypointSLStrs[idx] = starlark.String(entrypointFragment)
	}
	return starlark.NewList(entrypointSLStrs)
}

func getStarlarkCommand(composeCommand types.ShellCommand) *starlark.List {
	commandSLStrs := make([]starlark.Value, len(composeCommand))
	for idx, commandFragment := range composeCommand {
		commandSLStrs[idx] = starlark.String(commandFragment)
	}
	return starlark.NewList(commandSLStrs)
}

func getStarlarkEnvVars(composeEnvironment types.MappingWithEquals) (*starlark.Dict, error) {
	enVarsSLDict := starlark.NewDict(len(composeEnvironment))
	for key, value := range composeEnvironment {
		if value == nil {
			continue
		}
		if err := enVarsSLDict.SetKey(
			starlark.String(key),
			starlark.String(*value),
		); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred setting key '%s' in environment variables Starlark dict.", key)
		}
	}
	return enVarsSLDict, nil
}

// The 'volumes:' compose key supports named volumes and bind mounts
// Named volumes are currently not supported TODO: Support named volumes https://docs.docker.com/storage/volumes/
// bind mount semantics:
// <rel path on host>:<path on container> := upload a files artifacts of <rel path on host>, mount the files artifacts on the container at <path on container>
// <abs path on host>:<path on container>:= create a persistent directory on container at <path on container>
// <abs path on host> := create a persistent directory on container at <abs path on host>
// <rel path on host> := create a persistent directory on container at <abs path on host>
func getStarlarkFilesArtifacts(composeVolumes []types.ServiceVolumeConfig, serviceName string, pathsToUpload map[string]string) (starlark.Value, error) {
	filesArgSLDict := starlark.NewDict(len(composeVolumes))

	for volumeIdx, volume := range composeVolumes {
		volumeType := volume.Type

		var shouldPersist bool
		switch volumeType {
		case types.VolumeTypeBind:
			// Assume that if an absolute path is specified, user wants to use volume as a persistence layer
			// Additionally, assume relative paths are read-only
			shouldPersist = path.IsAbs(volume.Source)
		case types.VolumeTypeVolume:
			shouldPersist = true
		}

		var filesDictValue starlark.Value
		if shouldPersist {
			persistenceKey := fmt.Sprintf("volume%d", volumeIdx)
			persistentDirectory, err := getStarlarkPersistentDirectory(persistenceKey)
			if err != nil {
				return nil, stacktrace.Propagate(err, "An error occurred creating persistent directory with key '%s' for volume #%d.", persistenceKey, volumeIdx)
			}
			filesDictValue = persistentDirectory
		} else {
			// If not persistent, do an upload_files
			filesArtifactName := fmt.Sprintf("%s--volume%d", serviceName, volumeIdx)
			pathsToUpload[volume.Source] = filesArtifactName
			filesDictValue = starlark.String(filesArtifactName)
		}

		if err := filesArgSLDict.SetKey(starlark.String(volume.Target), filesDictValue); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred setting volume mountpoint '%s' in the files Starlark dict.", volume.Target)
		}
	}

	return filesArgSLDict, nil
}

func getStarlarkPersistentDirectory(persistenceKey string) (starlark.Value, error) {
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

// TODO: Support max allocation
func getStarlarkMinMemory(composeDeployConfig *types.DeployConfig) starlark.Int {
	reservation := 0
	if composeDeployConfig.Resources.Reservations != nil {
		reservation = int(composeDeployConfig.Resources.Reservations.MemoryBytes) / bytesToMegabytes
	}
	return starlark.MakeInt(reservation)
}

func getStarlarkMinCpus(composeDeployConfig *types.DeployConfig) starlark.Int {
	reservation := 0
	if composeDeployConfig.Resources.Reservations != nil {
		reservationParsed, err := strconv.ParseFloat(composeDeployConfig.Resources.Reservations.NanoCPUs, float64BitWidth)
		if err == nil {
			// Despite being called 'nano CPUs', they actually refer to a float representing percentage of one CPU
			reservation = int(reservationParsed * cpuToMilliCpuConstant)
		} else {
			logrus.Warnf("Could not convert CPU reservation '%v' to integer, limits reservation", composeDeployConfig.Resources.Reservations.NanoCPUs)
		}
	}
	return starlark.MakeInt(reservation)
}

func appendKwarg(kwargs []starlark.Tuple, argName string, argValue starlark.Value) []starlark.Tuple {
	tuple := []starlark.Value{
		starlark.String(argName),
		argValue,
	}
	return append(kwargs, tuple)
}

// Returns list of service names in an order that respects dependencies by performing a topological sort (simple bfs)
// Returns error if cyclical dependency is detected
// TODO: make this determinitic with a tie breaker
func sortServicesBasedOnDependencies(perServiceDependencies map[string]map[string]bool) ([]string, error) {
	dependencyCount := map[string]int{}
	for _, dependencies := range perServiceDependencies {
		for dependency := range dependencies {
			dependencyCount[dependency]++
		}
	}

	queue := make([]string, 0)
	for service := range perServiceDependencies {
		if dependencyCount[service] == 0 {
			queue = append(queue, service)
		}
	}

	sortedServices := []string{}
	for len(queue) > 0 {
		dequeuedService := queue[0]
		queue = queue[1:]

		sortedServices = append(sortedServices, dequeuedService)

		for service := range perServiceDependencies[dequeuedService] {
			dependencyCount[service]--

			if dependencyCount[service] == 0 {
				queue = append(queue, service)
			}
		}
	}

	// Check for cycles
	for _, incoming := range dependencyCount {
		if incoming > 0 {
			return nil, stacktrace.NewError("A cycle was found in the service dependency graph.")
		}
	}

	// Reverse the result slice
	for i, j := 0, len(sortedServices)-1; i < j; i, j = i+1, j-1 {
		sortedServices[i], sortedServices[j] = sortedServices[j], sortedServices[i]
	}

	return sortedServices, nil
}
