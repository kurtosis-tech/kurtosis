package snapshots

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/docker/docker/client"
	api_services "github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/starlark_script_creator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
)

// BEFORE THIS FUNCTION IS CALLED
// - create the persistent directory docker managed volumes
// - service network registers all the services beforehand so they get started with correct ips
// NOTES:
//   - uses add_services to parallelize services that can be parallelized?
func GetMainScriptToExecuteFromSnapshotPackage(packageRootPathOnDisk string) (string, error) {
	ctx := context.Background()
	// 	- upload the files artifacts into the enclave
	perServiceFilesArtifactsToUpload := map[string]map[string]string{}

	// 	- recreates the service configs of all the snapshotted services
	// 		- configs use snapshotted images
	// 		- same entrypoint and cmd as original services
	// 		- same env vars and ports as original service
	// 		- only mounts persistent directories
	orderedServiceList, serviceDependencyGraph, err := getOrderedServiceListAndDependencies(packageRootPathOnDisk)
	if err != nil {
		return "", err // already wrapped
	}

	serviceNameToStarlarkServiceConfig := map[string]starlark_script_creator.StarlarkServiceConfig{}
	for _, serviceName := range orderedServiceList {
		serviceConfigPath := fmt.Sprintf(serviceConfigPathFmtSpecifier, serviceName, serviceConfigFileName)

		serviceConfigBytes, err := os.ReadFile(path.Join(packageRootPathOnDisk, serviceConfigPath))
		if err != nil {
			return "", stacktrace.Propagate(err, "An error occurred reading the service config file at path: %v", serviceConfigPath)
		}

		serviceConfig := api_services.ServiceConfig{}
		if err := json.Unmarshal(serviceConfigBytes, &serviceConfig); err != nil {
			return "", stacktrace.Propagate(err, "An error occurred unmarshalling the service config file at path: %v", serviceConfigPath)
		}

		// Convert service config into Starlark Service Config
		serviceConfigKwargs := []starlark.Tuple{}

		// IMAGE
		dockerManager, err := docker_manager.CreateDockerManager([]client.Opt{})
		if err != nil {
			return "", stacktrace.Propagate(err, "An error occurred creating a docker manager")
		}

		imageName := fmt.Sprintf(snapshottedImageNameFmtSpecifier, serviceName)
		serviceImagePath := fmt.Sprintf(serviceImagePathFmtSpecifier, serviceName, imageName)
		logrus.Infof("Loading image for service '%v' from path: %v", serviceName, serviceImagePath)
		err = dockerManager.LoadImage(ctx, path.Join(packageRootPathOnDisk, serviceImagePath))
		if err != nil {
			return "", stacktrace.Propagate(err, "An error occurred loading the image for service '%v'", serviceName)
		}
		logrus.Infof("Successfully loaded image for service '%v'", serviceName)

		serviceConfigKwargs = starlark_script_creator.AppendKwarg(
			serviceConfigKwargs,
			service_config.ImageAttr,
			starlark.String(imageName),
		)

		// cmd

		// env vars

		// Finally, create Starlark Service Config object based on kwargs
		argumentValuesSet, interpretationErr := builtin_argument.CreateNewArgumentValuesSet(
			service_config.ServiceConfigTypeName,
			service_config.NewServiceConfigType().Arguments,
			[]starlark.Value{},
			serviceConfigKwargs,
		)
		if interpretationErr != nil {
			return "", stacktrace.Propagate(interpretationErr, "An starlark interpretation error was detected while attempting to create argument values for service config for service '%v'.", serviceName)
		}
		serviceConfigKurtosisType, interpretationErr := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(service_config.ServiceConfigTypeName, argumentValuesSet)
		if interpretationErr != nil {
			return "", stacktrace.Propagate(interpretationErr, "An starlark interpretation error was detected while attempting to create a service config for service '%v'.", serviceName)
		}
		serviceNameToStarlarkServiceConfig[serviceName] = serviceConfigKurtosisType
	}

	snapshotStarlarkScript, err := starlark_script_creator.CreateStarlarkScript(serviceNameToStarlarkServiceConfig, serviceDependencyGraph, perServiceFilesArtifactsToUpload)
	if err != nil {
		return "", err
	}
	return snapshotStarlarkScript, nil
}

// func getStarlarkImageBuildSpec(composeBuild *types.BuildConfig, serviceName string) (starlark.Value, error) {
// 	var imageBuildSpecKwargs []starlark.Tuple

// 	builtImageName := serviceName + builtImageSuffix
// 	imageNameKwarg := []starlark.Value{
// 		starlark.String(service_config.BuiltImageNameAttr),
// 		starlark.String(builtImageName),
// 	}
// 	imageBuildSpecKwargs = append(imageBuildSpecKwargs, imageNameKwarg)
// 	if composeBuild.Context != "" {
// 		contextDirKwarg := []starlark.Value{
// 			starlark.String(service_config.BuildContextAttr),
// 			starlark.String(composeBuild.Context),
// 		}
// 		imageBuildSpecKwargs = append(imageBuildSpecKwargs, contextDirKwarg)
// 	}
// 	if composeBuild.Target != "" {
// 		targetStageKwarg := []starlark.Value{
// 			starlark.String(service_config.TargetStageAttr),
// 			starlark.String(composeBuild.Target),
// 		}
// 		imageBuildSpecKwargs = append(imageBuildSpecKwargs, targetStageKwarg)
// 	}

// 	imageBuildSpecArgumentValuesSet, interpretationErr := builtin_argument.CreateNewArgumentValuesSet(
// 		service_config.ImageBuildSpecTypeName,
// 		service_config.NewImageBuildSpecType().Arguments,
// 		[]starlark.Value{},
// 		imageBuildSpecKwargs,
// 	)
// 	if interpretationErr != nil {
// 		return nil, stacktrace.Propagate(interpretationErr, "An starlark interpretation error was detected while attempting to create argument values for image build spec for service '%v'.", serviceName)
// 	}
// 	imageBuildSpecKurtosisType, interpretationErr := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(service_config.ImageBuildSpecTypeName, imageBuildSpecArgumentValuesSet)
// 	if interpretationErr != nil {
// 		return nil, stacktrace.Propagate(interpretationErr, "An starlark interpretation error was detected while attempting to create an image build spec for '%v'.", serviceName)
// 	}
// 	return imageBuildSpecKurtosisType, nil
// }

// // TODO: Support public ports
// func getStarlarkPortSpecs(serviceName string, composePorts []types.ServicePortConfig) (*starlark.Dict, error) {
// 	portSpecs := starlark.NewDict(len(composePorts))

// 	for portIdx, dockerPort := range composePorts {
// 		portName := fmt.Sprintf("port%d", portIdx)

// 		dockerProto := dockerPort.Protocol
// 		kurtosisProto, found := dockerPortProtosToKurtosisPortProtos[strings.ToLower(dockerProto)]
// 		if !found {
// 			return nil, stacktrace.NewError("Port #%d has unsupported protocol '%v'", portIdx, dockerProto)
// 		}

// 		var applicationProtocol string
// 		if slices.Contains(possibleHttpPorts, dockerPort.Target) {
// 			applicationProtocol = httpProtocol
// 		}

// 		portSpec, interpretationErr := port_spec_starlark.CreatePortSpecUsingGoValues(
// 			serviceName,
// 			uint16(dockerPort.Target),
// 			kurtosisProto,
// 			&applicationProtocol, // Application protocol (which Compose doesn't have)
// 			"",                   // Wait timeout (which Compose doesn't have a way to override)
// 			nil,                  // No way to change the URL for the port
// 		)
// 		if interpretationErr != nil {
// 			return nil, stacktrace.Propagate(interpretationErr, "An error occurred creating a %s object from port #%d", port_spec_starlark.PortSpecTypeName, portIdx)
// 		}
// 		if err := portSpecs.SetKey(starlark.String(portName), portSpec); err != nil {
// 			return nil, stacktrace.Propagate(err, "An error occurred putting port #%d in Starlark dict", portIdx)
// 		}
// 	}

// 	return portSpecs, nil
// }

// func getStarlarkEntrypoint(composeEntrypoint types.ShellCommand) *starlark.List {
// 	entrypointSLStrs := make([]starlark.Value, len(composeEntrypoint))
// 	for idx, entrypointFragment := range composeEntrypoint {
// 		entrypointSLStrs[idx] = starlark.String(entrypointFragment)
// 	}
// 	return starlark.NewList(entrypointSLStrs)
// }

// func getStarlarkCommand(composeCommand types.ShellCommand) *starlark.List {
// 	commandSLStrs := make([]starlark.Value, len(composeCommand))
// 	for idx, commandFragment := range composeCommand {
// 		commandSLStrs[idx] = starlark.String(commandFragment)
// 	}
// 	return starlark.NewList(commandSLStrs)
// }

// func getStarlarkEnvVars(composeEnvironment types.MappingWithEquals, envFiles types.StringList, packageAbsDirPath string) (*starlark.Dict, error) {
// 	// make iteration order of [composeEnvironment] deterministic by getting the keys and sorting them
// 	envVarKeys := []string{}
// 	for key := range composeEnvironment {
// 		envVarKeys = append(envVarKeys, key)
// 	}
// 	sort.Strings(envVarKeys)
// 	envVarsSLDict := starlark.NewDict(len(composeEnvironment))
// 	for _, key := range envVarKeys {
// 		value := composeEnvironment[key]
// 		if value == nil {
// 			continue
// 		}
// 		if err := envVarsSLDict.SetKey(
// 			starlark.String(key),
// 			starlark.String(*value),
// 		); err != nil {
// 			return nil, stacktrace.Propagate(err, "An error occurred setting key '%s' in environment variables Starlark dict.", key)
// 		}
// 	}

// 	// if env file is specified, manually parse the env file at the location it is inside the package on the APIC
// 	for _, envFilePath := range envFiles {
// 		serviceEnvFilePath := path.Join(packageAbsDirPath, envFilePath)
// 		envFileReader, err := os.Open(serviceEnvFilePath)
// 		if err != nil {
// 			return nil, stacktrace.Propagate(err, "An error occurred opening env file at path: %v", serviceEnvFilePath)
// 		}
// 		envVars, err := envparse.Parse(envFileReader)
// 		if err != nil {
// 			return nil, stacktrace.Propagate(err, "An error occurred parsing env file.")
// 		}
// 		for key, value := range envVars {
// 			if err := envVarsSLDict.SetKey(
// 				starlark.String(key),
// 				starlark.String(value),
// 			); err != nil {
// 				return nil, stacktrace.Propagate(err, "An error occurred setting key '%s' in environment variables Starlark dict.", key)
// 			}
// 		}
// 	}

// 	return envVarsSLDict, nil
// }

// // The 'volumes:' compose key supports named volumes and bind mounts https://docs.docker.com/storage/volumes/
// // bind mount semantics for starlark:
// // <rel path on host>:<path on container> := upload a files artifacts of <rel path on host>, mount the files artifacts on the container at <path on container>
// // <abs path on host>:<path on container>:= create a persistent directory on container at <path on container>
// // <abs path on host> := create a persistent directory on container at <abs path on host>
// // <rel path on host> := create a persistent directory on container at <rel path on host>
// // Named volumes are treated https://docs.docker.com/storage/volumes/ as absolute paths persistence layers, and thus a persistent directory is created
// func getStarlarkFilesArtifacts(composeVolumes []types.ServiceVolumeConfig, serviceName string, packageAbsDirPath string) (starlark.Value, map[string]string, *starlark.Dict, error) {
// 	filesArgSLDict := starlark.NewDict(len(composeVolumes))
// 	filesArtifactsToUpload := map[string]string{}

// 	filesToBeMoved := starlark.NewDict(len(composeVolumes))

// 	for volumeIdx, volume := range composeVolumes {
// 		volumeType := volume.Type

// 		var shouldPersist bool
// 		switch volumeType {
// 		// if an absolute path is specified, assume user wants to use volume as a persistence layer and create a Persistent Directory
// 		// if path is relative, assume it's read only and do an upload files
// 		case types.VolumeTypeBind:
// 			volumePath := cleanFilePath(volume.Source)
// 			shouldPersist = path.IsAbs(volumePath)
// 		// if named volume is provided, assume user wants to use volume as a persistence layer and create a Persistent Directory
// 		case types.VolumeTypeVolume:
// 			shouldPersist = true
// 		default:
// 			shouldPersist = false
// 		}

// 		var filesDictValue starlark.Value
// 		targetDirectoryForFilesArtifact := volume.Target
// 		if shouldPersist {
// 			persistenceKey := fmt.Sprintf("%s--volume%d", serviceName, volumeIdx)
// 			persistentDirectory, err := getStarlarkPersistentDirectory(persistenceKey)
// 			if err != nil {
// 				return nil, nil, nil, stacktrace.Propagate(err, "An error occurred creating persistent directory with key '%s' for volume #%d.", persistenceKey, volumeIdx)
// 			}
// 			filesDictValue = persistentDirectory
// 		} else {
// 			// If not persistent, do an upload_files
// 			filesArtifactName := fmt.Sprintf("%s--volume%d", serviceName, volumeIdx)
// 			filesArtifactsToUpload[volume.Source] = filesArtifactName
// 			filesDictValue = starlark.String(filesArtifactName)

// 			// TODO: update files artifact expansion to handle mounting files, not only directories so files_to_be_moved hack can be removed
// 			// if the volume is referencing a file, use files_to_be_moved
// 			maybeFileOrDirVolume, err := os.Stat(path.Join(packageAbsDirPath, volume.Source))
// 			if err != nil {
// 				return nil, nil, nil, stacktrace.Propagate(err, "An error occurred checking is the volume path existed in the package on disc: %v.", volume.Source)
// 			}
// 			if !maybeFileOrDirVolume.IsDir() {
// 				sourcePathNameEnd := path.Base(volume.Source)
// 				targetDirectoryForFilesArtifact = path.Join("/tmp", filesArtifactName)
// 				targetToMovePath := path.Join(targetDirectoryForFilesArtifact, sourcePathNameEnd)
// 				if err := filesToBeMoved.SetKey(starlark.String(targetToMovePath), starlark.String(volume.Target)); err != nil {
// 					return nil, nil, nil, stacktrace.Propagate(err, "An error occurred setting files to be moved for targetDirectoryForFilesArtifact '%v'", volume.Target)
// 				}
// 			}
// 		}
// 		if err := filesArgSLDict.SetKey(starlark.String(targetDirectoryForFilesArtifact), filesDictValue); err != nil {
// 			return nil, nil, nil, stacktrace.Propagate(err, "An error occurred setting volume mountpoint '%s' in the files Starlark dict.", volume.Target)
// 		}
// 	}

// 	return filesArgSLDict, filesArtifactsToUpload, filesToBeMoved, nil
// }

// func getStarlarkPersistentDirectory(persistenceKey string) (starlark.Value, error) {
// 	directoryKwargs := []starlark.Tuple{}
// 	directoryKwargs = appendKwarg(
// 		directoryKwargs,
// 		directory.PersistentKeyAttr,
// 		starlark.String(persistenceKey),
// 	)

// 	argumentValuesSet, interpretationErr := builtin_argument.CreateNewArgumentValuesSet(
// 		directory.DirectoryTypeName,
// 		directory.NewDirectoryType().Arguments,
// 		[]starlark.Value{},
// 		directoryKwargs,
// 	)
// 	if interpretationErr != nil {
// 		return nil, stacktrace.Propagate(interpretationErr, "An starlark interpretation error was detected while attempting to create argument values for persistent directory.")
// 	}
// 	directoryKurtosisType, interpretationErr := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(directory.DirectoryTypeName, argumentValuesSet)
// 	if interpretationErr != nil {
// 		return nil, stacktrace.Propagate(interpretationErr, "An starlark interpretation error was detected while attempting to create a persistent directory.")
// 	}

// 	return directoryKurtosisType, nil
// }

// // TODO: Support max allocation
// func getStarlarkMinMemory(composeDeployConfig *types.DeployConfig) starlark.Int {
// 	reservation := 0
// 	if composeDeployConfig.Resources.Reservations != nil {
// 		reservation = int(composeDeployConfig.Resources.Reservations.MemoryBytes) / bytesToMegabytes
// 	}
// 	return starlark.MakeInt(reservation)
// }

// func getStarlarkMinCpus(composeDeployConfig *types.DeployConfig) starlark.Int {
// 	reservation := 0
// 	if composeDeployConfig.Resources.Reservations != nil {
// 		reservationParsed, err := strconv.ParseFloat(composeDeployConfig.Resources.Reservations.NanoCPUs, float64BitWidth)
// 		if err == nil {
// 			// Despite being called 'nano CPUs', they actually refer to a float representing percentage of one CPU
// 			reservation = int(reservationParsed * cpuToMilliCpuConstant)
// 		} else {
// 			logrus.Warnf("Could not convert CPU reservation '%v' to integer, limits reservation", composeDeployConfig.Resources.Reservations.NanoCPUs)
// 		}
// 	}
// 	return starlark.MakeInt(reservation)
// }

func getOrderedServiceListAndDependencies(packageRootPathOnDisk string) ([]string, map[string]map[string]bool, error) {
	serviceStartupOrderBytes, err := os.ReadFile(path.Join(packageRootPathOnDisk, serviceStartupOrderFileName))
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred reading the service startup order file at path: %v", path.Join(packageRootPathOnDisk, fmt.Sprintf("%s/%s", packageRootPathOnDisk, serviceStartupOrderFileName)))
	}
	serviceStartupOrder := strings.Split(string(serviceStartupOrderBytes), "\n")

	serviceDependencyGraph := map[string]map[string]bool{}
	for idx, serviceName := range serviceStartupOrder {
		serviceDependencyGraph[serviceName] = map[string]bool{}
		if idx > 0 {
			serviceDependencyGraph[serviceName][serviceStartupOrder[idx-1]] = true
		}
	}
	return serviceStartupOrder, serviceDependencyGraph, nil
}
