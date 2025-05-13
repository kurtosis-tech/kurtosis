package docker_compose_transpiler

import (
	"errors"
	"fmt"
	"github.com/hashicorp/go-envparse"
	"golang.org/x/exp/slices"
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"

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

	// Don't resolve paths as they should be resolved by package content provider
	shouldResolvePaths = false

	// Convert Windows paths to Linux paths as APIC runs on Linux
	shouldConvertWindowsPathsToLinux = true

	builtImageSuffix = "-image"

	// eg. plan.upload_files(src = "./data/project", name = "web-volume0")
	uploadFilesLinesFmtStr = "plan.upload_files(src = \"%s\", name = \"%s\")"

	// eg. plan.add_service(name="web", config=ServiceConfig(...))
	addServiceLinesFmtStr = "plan.add_service(name = \"%s\", config = %s)"

	defRunStr = "def run(plan):\n"

	newStarlarkLineFmtStr = "    %s\n"

	httpProtocol = "http"

	alphanumericCharWithDashesRegexStr = `[^a-z0-9-]`
	consecutiveDashesRegexStr          = `-+`
)

var (
	alphanumericCharWithDashesRegex = regexp.MustCompile(alphanumericCharWithDashesRegexStr)
	consecutiveDashesRegex          = regexp.MustCompile(consecutiveDashesRegexStr)
	possibleHttpPorts               = []uint32{8080, 8000, 80, 443}
)

var DefaultComposeFilenames = []string{
	"compose.yml",
	"compose.yaml",
	"docker-compose.yml",
	"docker-compose.yaml",
	"docker_compose.yml",
	"docker_compose.yaml",
}

type ComposeService types.ServiceConfig

type StarlarkServiceConfig *kurtosis_type_constructor.KurtosisValueTypeDefault

var dockerPortProtosToKurtosisPortProtos = map[string]port_spec.TransportProtocol{
	"tcp":  port_spec.TransportProtocol_TCP,
	"udp":  port_spec.TransportProtocol_UDP,
	"sctp": port_spec.TransportProtocol_SCTP,
}

var CyclicalDependencyError = stacktrace.NewError("A cycle was detected in the service dependency graph.")

func TranspileDockerComposePackageToStarlark(packageAbsDirpath string, relativePathToComposeFile string) (string, error) {
	composeAbsFilepath := path.Join(packageAbsDirpath, relativePathToComposeFile)

	// Useful for logging to prevent leaking internals of APIC
	composeFilename := path.Base(relativePathToComposeFile)

	composeBytes, err := os.ReadFile(composeAbsFilepath)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred reading Compose file '%v'", composeFilename)
	}

	// Use env vars file next to Compose if it exists
	envVarsFilepath := path.Join(packageAbsDirpath, envVarsFilename)
	envVarsInFile, err := godotenv.Read(envVarsFilepath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return "", stacktrace.Propagate(err, "An %v file was found in the package, but an error occurred reading it.", envVarsFilename)
		}
		envVarsInFile = map[string]string{}
	}

	starlarkScript, err := convertComposeToStarlarkScript(composeBytes, envVarsInFile, packageAbsDirpath)
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
func convertComposeToStarlarkScript(composeBytes []byte, envVars map[string]string, packageAbsDirPath string) (string, error) {
	composeStruct, err := convertComposeBytesToComposeStruct(composeBytes, envVars)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred converting compose bytes into a struct.")
	}

	serviceNameToStarlarkServiceConfig, serviceDependencyGraph, perServiceFilesArtifactsToUpload, err := convertComposeServicesToStarlarkInfo(composeStruct.Services, packageAbsDirPath)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred converting compose services to starlark service configs.")
	}

	return createStarlarkScript(serviceNameToStarlarkServiceConfig, serviceDependencyGraph, perServiceFilesArtifactsToUpload)
}

func convertComposeBytesToComposeStruct(composeBytes []byte, envVars map[string]string) (*types.Project, error) {
	composeParseConfig := types.ConfigDetails{ //nolint:exhaustruct
		// Note that we might be able to use the WorkingDir property instead, to parse the entire directory
		// nolint: exhaustruct
		ConfigFiles: []types.ConfigFile{{
			Content: composeBytes,
		}},
		Environment: envVars,
	}
	setOptionsFunc := func(options *loader.Options) {
		options.SetProjectName(composeProjectName, shouldOverrideComposeYamlKeyProjectName)
		options.ResolvePaths = shouldResolvePaths
		options.ConvertWindowsPaths = shouldConvertWindowsPathsToLinux
	}
	compose, err := loader.Load(composeParseConfig, setOptionsFunc)
	// don't err if env file not found, transpiler will handle finding it
	// TODO: fork compose loader package and change the env file logic so we don't have to catch this err
	if err != nil && !isEnvFileNotFoundErr(err) {
		return nil, stacktrace.Propagate(err, "An error occurred parsing compose based on provided parsing config and set options function.")
	}
	return compose, nil
}

// Creates a starlark script based on starlark ServiceConfigs, the service dependency graph, and files artifacts to upload
func createStarlarkScript(
	serviceNameToStarlarkServiceConfig map[string]StarlarkServiceConfig,
	serviceDependencyGraph map[string]map[string]bool,
	servicesToFilesArtifactsToUpload map[string]map[string]string) (string, error) {
	starlarkLines := []string{}

	// Add add_service instructions in an order that respects [serviceDependencyGraph] determined by 'depends_on' keys in Compose
	sortedServices, err := sortServicesBasedOnDependencies(serviceDependencyGraph)
	if err != nil {
		return "", err
	}
	for _, serviceName := range sortedServices {
		// upload_files artifacts for service
		// get and sort keys first for deterministic order
		filesArtifactsToUpload := servicesToFilesArtifactsToUpload[serviceName]
		sortedRelativePaths := []string{}
		for relativePath := range filesArtifactsToUpload {
			sortedRelativePaths = append(sortedRelativePaths, relativePath)
		}
		sort.Strings(sortedRelativePaths)
		for _, relativePath := range sortedRelativePaths {
			filesArtifactName := filesArtifactsToUpload[relativePath]
			uploadFilesLine := fmt.Sprintf(uploadFilesLinesFmtStr, relativePath, filesArtifactName)
			starlarkLines = append(starlarkLines, uploadFilesLine)
		}

		// add_service
		starlarkServiceConfig := *serviceNameToStarlarkServiceConfig[serviceName]
		addServiceLine := fmt.Sprintf(addServiceLinesFmtStr, serviceName, starlarkServiceConfig.String())
		starlarkLines = append(starlarkLines, addServiceLine)
	}

	script := defRunStr
	for _, line := range starlarkLines {
		script += fmt.Sprintf(newStarlarkLineFmtStr, line)
	}
	return script, nil
}

// Turns DockerCompose Service into Kurtosis ServiceConfigs and returns info needed for creating a valid starlark script
func convertComposeServicesToStarlarkInfo(composeServices types.Services, packageAbsDirPath string) (
	map[string]StarlarkServiceConfig, // Map of service names to Kurtosis ServiceConfig's
	map[string]map[string]bool, // Graph of service dependencies based on depends_on key (determines order in which to add services)
	map[string]map[string]string, // Map of service names to map of relative paths to files artifacts names that need to get uploaded for the service (determines files artifacts that need to be uploaded)
	error) {
	serviceNameToStarlarkServiceConfig := map[string]StarlarkServiceConfig{}
	perServiceDependencies := map[string]map[string]bool{}
	servicesToFilesArtifactsToUpload := map[string]map[string]string{}

	serviceNameToContainerNameMap := map[string]string{}
	for _, s := range composeServices {
		if s.ContainerName != "" {
			serviceNameToContainerNameMap[s.Name] = s.ContainerName
		}
	}

	for _, service := range composeServices {
		composeService := ComposeService(service)
		serviceConfigKwargs := []starlark.Tuple{}

		serviceName := composeService.Name
		// hostname becomes container name if it's set in compose
		// in kurtosis service name becomes hostname, so set service name as container name (thus hostname) so other services can comm with it
		if containerName, ok := serviceNameToContainerNameMap[serviceName]; ok {
			serviceName = containerName
		}
		serviceName = convertToRFC1035(serviceName)

		// IMAGE
		imageName := composeService.Image
		// if build directive used, create an image build spec
		// otherwise, use image name
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
		} else {
			serviceConfigKwargs = appendKwarg(
				serviceConfigKwargs,
				service_config.ImageAttr,
				starlark.String(imageName),
			)
		}

		// PORTS
		if composeService.Ports != nil {
			portSpecsDict, err := getStarlarkPortSpecs(serviceName, composeService.Ports)
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
		if composeService.Environment != nil || composeService.EnvFile != nil {
			envVarsDict, err := getStarlarkEnvVars(composeService.Environment, composeService.EnvFile, packageAbsDirPath)
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
			filesDict, artifactsToUpload, filesToBeMoved, err := getStarlarkFilesArtifacts(composeService.Volumes, serviceName, packageAbsDirPath)
			if err != nil {
				return nil, nil, nil, stacktrace.Propagate(err, "An error occurred creating the files dict for service '%s'", serviceName)
			}
			serviceConfigKwargs = appendKwarg(
				serviceConfigKwargs,
				service_config.FilesAttr,
				filesDict,
			)
			if filesToBeMoved.Len() > 0 {
				serviceConfigKwargs = appendKwarg(serviceConfigKwargs, service_config.FilesToBeMovedAttr, filesToBeMoved)
			}
			servicesToFilesArtifactsToUpload[serviceName] = artifactsToUpload
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
			// do container name switch if it exists
			if containerName, ok := serviceNameToContainerNameMap[dependencyName]; ok {
				dependencyName = containerName
			}
			dependencyServiceNames[convertToRFC1035(dependencyName)] = true
		}
		perServiceDependencies[serviceName] = dependencyServiceNames

		// Finally, create Starlark Service Config object based on kwargs
		argumentValuesSet, interpretationErr := builtin_argument.CreateNewArgumentValuesSet(
			service_config.ServiceConfigTypeName,
			service_config.NewServiceConfigType().Arguments,
			[]starlark.Value{},
			serviceConfigKwargs,
		)
		if interpretationErr != nil {
			return nil, nil, nil, stacktrace.Propagate(interpretationErr, "An starlark interpretation error was detected while attempting to create argument values for service config for service '%v'.", serviceName)
		}
		serviceConfigKurtosisType, interpretationErr := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(service_config.ServiceConfigTypeName, argumentValuesSet)
		if interpretationErr != nil {
			return nil, nil, nil, stacktrace.Propagate(interpretationErr, "An starlark interpretation error was detected while attempting to create a service config for service '%v'.", serviceName)
		}
		serviceNameToStarlarkServiceConfig[serviceName] = serviceConfigKurtosisType
	}

	return serviceNameToStarlarkServiceConfig, perServiceDependencies, servicesToFilesArtifactsToUpload, nil
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
		service_config.NewImageBuildSpecType().Arguments,
		[]starlark.Value{},
		imageBuildSpecKwargs,
	)
	if interpretationErr != nil {
		return nil, stacktrace.Propagate(interpretationErr, "An starlark interpretation error was detected while attempting to create argument values for image build spec for service '%v'.", serviceName)
	}
	imageBuildSpecKurtosisType, interpretationErr := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(service_config.ImageBuildSpecTypeName, imageBuildSpecArgumentValuesSet)
	if interpretationErr != nil {
		return nil, stacktrace.Propagate(interpretationErr, "An starlark interpretation error was detected while attempting to create an image build spec for '%v'.", serviceName)
	}
	return imageBuildSpecKurtosisType, nil
}

// TODO: Support public ports
func getStarlarkPortSpecs(serviceName string, composePorts []types.ServicePortConfig) (*starlark.Dict, error) {
	portSpecs := starlark.NewDict(len(composePorts))

	for portIdx, dockerPort := range composePorts {
		portName := fmt.Sprintf("port%d", portIdx)

		dockerProto := dockerPort.Protocol
		kurtosisProto, found := dockerPortProtosToKurtosisPortProtos[strings.ToLower(dockerProto)]
		if !found {
			return nil, stacktrace.NewError("Port #%d has unsupported protocol '%v'", portIdx, dockerProto)
		}

		var applicationProtocol string
		if slices.Contains(possibleHttpPorts, dockerPort.Target) {
			applicationProtocol = httpProtocol
		}

		portSpec, interpretationErr := port_spec_starlark.CreatePortSpecUsingGoValues(
			serviceName,
			uint16(dockerPort.Target),
			kurtosisProto,
			&applicationProtocol, // Application protocol (which Compose doesn't have)
			"",                   // Wait timeout (which Compose doesn't have a way to override)
			nil,                  // No way to change the URL for the port
		)
		if interpretationErr != nil {
			return nil, stacktrace.Propagate(interpretationErr, "An error occurred creating a %s object from port #%d", port_spec_starlark.PortSpecTypeName, portIdx)
		}
		if err := portSpecs.SetKey(starlark.String(portName), portSpec); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred putting port #%d in Starlark dict", portIdx)
		}
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

func getStarlarkEnvVars(composeEnvironment types.MappingWithEquals, envFiles types.StringList, packageAbsDirPath string) (*starlark.Dict, error) {
	// make iteration order of [composeEnvironment] deterministic by getting the keys and sorting them
	envVarKeys := []string{}
	for key := range composeEnvironment {
		envVarKeys = append(envVarKeys, key)
	}
	sort.Strings(envVarKeys)
	envVarsSLDict := starlark.NewDict(len(composeEnvironment))
	for _, key := range envVarKeys {
		value := composeEnvironment[key]
		if value == nil {
			continue
		}
		if err := envVarsSLDict.SetKey(
			starlark.String(key),
			starlark.String(*value),
		); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred setting key '%s' in environment variables Starlark dict.", key)
		}
	}

	// if env file is specified, manually parse the env file at the location it is inside the package on the APIC
	for _, envFilePath := range envFiles {
		serviceEnvFilePath := path.Join(packageAbsDirPath, envFilePath)
		envFileReader, err := os.Open(serviceEnvFilePath)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred opening env file at path: %v", serviceEnvFilePath)
		}
		envVars, err := envparse.Parse(envFileReader)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred parsing env file.")
		}
		for key, value := range envVars {
			if err := envVarsSLDict.SetKey(
				starlark.String(key),
				starlark.String(value),
			); err != nil {
				return nil, stacktrace.Propagate(err, "An error occurred setting key '%s' in environment variables Starlark dict.", key)
			}
		}
	}

	return envVarsSLDict, nil
}

// The 'volumes:' compose key supports named volumes and bind mounts https://docs.docker.com/storage/volumes/
// bind mount semantics for starlark:
// <rel path on host>:<path on container> := upload a files artifacts of <rel path on host>, mount the files artifacts on the container at <path on container>
// <abs path on host>:<path on container>:= create a persistent directory on container at <path on container>
// <abs path on host> := create a persistent directory on container at <abs path on host>
// <rel path on host> := create a persistent directory on container at <rel path on host>
// Named volumes are treated https://docs.docker.com/storage/volumes/ as absolute paths persistence layers, and thus a persistent directory is created
func getStarlarkFilesArtifacts(composeVolumes []types.ServiceVolumeConfig, serviceName string, packageAbsDirPath string) (starlark.Value, map[string]string, *starlark.Dict, error) {
	filesArgSLDict := starlark.NewDict(len(composeVolumes))
	filesArtifactsToUpload := map[string]string{}

	filesToBeMoved := starlark.NewDict(len(composeVolumes))

	for volumeIdx, volume := range composeVolumes {
		volumeType := volume.Type

		var shouldPersist bool
		switch volumeType {
		// if an absolute path is specified, assume user wants to use volume as a persistence layer and create a Persistent Directory
		// if path is relative, assume it's read only and do an upload files
		case types.VolumeTypeBind:
			volumePath := cleanFilePath(volume.Source)
			shouldPersist = path.IsAbs(volumePath)
		// if named volume is provided, assume user wants to use volume as a persistence layer and create a Persistent Directory
		case types.VolumeTypeVolume:
			shouldPersist = true
		default:
			shouldPersist = false
		}

		var filesDictValue starlark.Value
		targetDirectoryForFilesArtifact := volume.Target
		if shouldPersist {
			persistenceKey := fmt.Sprintf("%s--volume%d", serviceName, volumeIdx)
			persistentDirectory, err := getStarlarkPersistentDirectory(persistenceKey)
			if err != nil {
				return nil, nil, nil, stacktrace.Propagate(err, "An error occurred creating persistent directory with key '%s' for volume #%d.", persistenceKey, volumeIdx)
			}
			filesDictValue = persistentDirectory
		} else {
			// If not persistent, do an upload_files
			filesArtifactName := fmt.Sprintf("%s--volume%d", serviceName, volumeIdx)
			filesArtifactsToUpload[volume.Source] = filesArtifactName
			filesDictValue = starlark.String(filesArtifactName)

			// TODO: update files artifact expansion to handle mounting files, not only directories so files_to_be_moved hack can be removed
			// if the volume is referencing a file, use files_to_be_moved
			maybeFileOrDirVolume, err := os.Stat(path.Join(packageAbsDirPath, volume.Source))
			if err != nil {
				return nil, nil, nil, stacktrace.Propagate(err, "An error occurred checking is the volume path existed in the package on disc: %v.", volume.Source)
			}
			if !maybeFileOrDirVolume.IsDir() {
				sourcePathNameEnd := path.Base(volume.Source)
				targetDirectoryForFilesArtifact = path.Join("/tmp", filesArtifactName)
				targetToMovePath := path.Join(targetDirectoryForFilesArtifact, sourcePathNameEnd)
				if err := filesToBeMoved.SetKey(starlark.String(targetToMovePath), starlark.String(volume.Target)); err != nil {
					return nil, nil, nil, stacktrace.Propagate(err, "An error occurred setting files to be moved for targetDirectoryForFilesArtifact '%v'", volume.Target)
				}
			}
		}
		if err := filesArgSLDict.SetKey(starlark.String(targetDirectoryForFilesArtifact), filesDictValue); err != nil {
			return nil, nil, nil, stacktrace.Propagate(err, "An error occurred setting volume mountpoint '%s' in the files Starlark dict.", volume.Target)
		}
	}

	return filesArgSLDict, filesArtifactsToUpload, filesToBeMoved, nil
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
		directory.NewDirectoryType().Arguments,
		[]starlark.Value{},
		directoryKwargs,
	)
	if interpretationErr != nil {
		return nil, stacktrace.Propagate(interpretationErr, "An starlark interpretation error was detected while attempting to create argument values for persistent directory.")
	}
	directoryKurtosisType, interpretationErr := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(directory.DirectoryTypeName, argumentValuesSet)
	if interpretationErr != nil {
		return nil, stacktrace.Propagate(interpretationErr, "An starlark interpretation error was detected while attempting to create a persistent directory.")
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

// Returns list of service names in an order that respects dependencies by performing a topological sort
// Returns error if cyclical dependency is detected
// o(n^2) but simpler variation of Kahns algorithm https://en.wikipedia.org/wiki/Topological_sorting#Kahn's_algorithm
// To ensure a deterministic sort, ties are broken lexicographically based on service name
func sortServicesBasedOnDependencies(serviceDependencyGraph map[string]map[string]bool) ([]string, error) {
	initServices := []string{} // start services with services that have no dependencies
	for service, dependencies := range serviceDependencyGraph {
		if len(dependencies) == 0 {
			initServices = append(initServices, service)
		}
	}

	sortedServices := []string{}
	queue := []string{}
	sort.Strings(initServices)
	queue = append(queue, initServices...)

	for len(queue) > 0 {
		processedService := queue[0]
		queue = queue[1:]
		sortedServices = append(sortedServices, processedService)
		delete(serviceDependencyGraph, processedService)

		servicesToQueue := []string{}
		for service, dependencies := range serviceDependencyGraph {
			// Remove processedService if it was as a dependency
			if dependencies[processedService] {
				delete(dependencies, processedService)

				// add service to queue if all of its dependencies have been processed
				if len(dependencies) == 0 {
					servicesToQueue = append(servicesToQueue, service)
				}
			}
		}

		sort.Strings(servicesToQueue)
		queue = append(queue, servicesToQueue...)
	}

	// If there are still dependencies that need to be processed, a cycle exists
	if len(serviceDependencyGraph) > 0 {
		return nil, CyclicalDependencyError
	}

	return sortedServices, nil
}

// Converts a string to a RFC 1035 compliant format
func convertToRFC1035(input string) string {
	// Remove leading and trailing spaces
	input = strings.TrimSpace(input)
	// Convert to lowercase
	input = strings.ToLower(input)
	// Replace non-alphanumeric characters with dashes
	input = alphanumericCharWithDashesRegex.ReplaceAllString(input, "-")
	// Remove consecutive dashes
	input = consecutiveDashesRegex.ReplaceAllString(input, "-")
	// Remove dashes at the beginning or end
	input = strings.Trim(input, "-")
	return input
}

// CleanFilePath cleans up a file path by removing '~', '..', and '_' characters
// while retaining absolute paths as absolute and relative paths as relative.
func cleanFilePath(filePath string) string {
	filePath = strings.ReplaceAll(filePath, "~", "")
	filePath = strings.ReplaceAll(filePath, "..", "")
	filePath = strings.ReplaceAll(filePath, "_", "")
	return filePath
}

func isEnvFileNotFoundErr(err error) bool {
	// Prob not the best way to check for this error but right now the only time anything's loaded by compose is for env_file
	return strings.Contains(err.Error(), "Failed to load")
}
