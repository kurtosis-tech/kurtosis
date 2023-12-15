package service_config

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_build_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service_directory"
	"github.com/kurtosis-tech/kurtosis/core/files_artifacts_expander/args"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/directory"
	starlark_port_spec "github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/port_spec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/starlark_warning"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"go.starlark.net/starlark"
	"math"
	"path"
)

const (
	ServiceConfigTypeName = "ServiceConfig"

	ImageAttr                       = "image"
	PortsAttr                       = "ports"
	PublicPortsAttr                 = "public_ports"
	FilesAttr                       = "files"
	EntrypointAttr                  = "entrypoint"
	CmdAttr                         = "cmd"
	EnvVarsAttr                     = "env_vars"
	PrivateIpAddressPlaceholderAttr = "private_ip_address_placeholder"
	CpuAllocationAttr               = "cpu_allocation"
	MemoryAllocationAttr            = "memory_allocation"
	ReadyConditionsAttr             = "ready_conditions"
	MinCpuMilliCoresAttr            = "min_cpu"
	MinMemoryMegaBytesAttr          = "min_memory"
	MaxCpuMilliCoresAttr            = "max_cpu"
	MaxMemoryMegaBytesAttr          = "max_memory"
	LabelsAttr                      = "labels"

	DefaultPrivateIPAddrPlaceholder = "KURTOSIS_IP_ADDR_PLACEHOLDER"

	filesArtifactExpansionDirsParentDirpath string = "/files-artifacts"
	// TODO This should be populated from the build flow that builds the files-artifacts-expander Docker image
	filesArtifactsExpanderImage string = "kurtosistech/files-artifacts-expander"

	minimumMemoryAllocationMegabytes = 6
)

func NewServiceConfigType() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: ServiceConfigTypeName,

			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              ImageAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Value],
					// TODO: add validation for image build spec
					Validator: nil,
				},
				{
					Name:              PortsAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.Dict],
					Validator:         nil,
				},
				{
					Name:              PublicPortsAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.Dict],
					Validator:         nil,
				},
				{
					Name:              FilesAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.Dict],
					Validator:         nil,
				},
				{
					Name:              EntrypointAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.List],
					Validator:         nil,
				},
				{
					Name:              CmdAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.List],
					Validator:         nil,
				},
				{
					Name:              EnvVarsAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.Dict],
					Validator:         nil,
				},
				{
					Name:              PrivateIpAddressPlaceholderAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, PrivateIpAddressPlaceholderAttr)
					},
				},
				{
					Name:              CpuAllocationAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Int],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.Uint64InRange(value, CpuAllocationAttr, 0, math.MaxUint64)
					},
					Deprecation: starlark_warning.Deprecation(
						starlark_warning.DeprecationDate{
							Day: 25, Year: 2023, Month: 6, //nolint:gomnd
						},
						"This field is being deprecated in favour of `max_cpu` to set a maximum cpu a container can use",
						nil,
					),
				},
				{
					Name:              MemoryAllocationAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Int],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.Uint64InRange(value, MemoryAllocationAttr, minimumMemoryAllocationMegabytes, math.MaxUint64)
					},
					Deprecation: starlark_warning.Deprecation(
						starlark_warning.DeprecationDate{
							Day: 25, Year: 2023, Month: 6, //nolint:gomnd
						},
						"This field is being deprecated in favour of `max_memory` to set maximum memory a container can use",
						nil,
					),
				},
				{
					Name:              MaxCpuMilliCoresAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Int],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.Uint64InRange(value, CpuAllocationAttr, 0, math.MaxUint64)
					},
				},
				{
					Name:              MinCpuMilliCoresAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Int],
					Validator:         nil,
				},
				{
					Name:              MaxMemoryMegaBytesAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Int],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.Uint64InRange(value, MemoryAllocationAttr, minimumMemoryAllocationMegabytes, math.MaxUint64)
					},
				},
				{
					Name:              MinMemoryMegaBytesAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Int],
					Validator:         nil,
				},
				{
					Name:              ReadyConditionsAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*ReadyCondition],
					Validator:         nil,
				},
				{
					Name:              LabelsAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.Dict],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.ServiceConfigLabels(value, LabelsAttr)
					},
				},
			},
		},

		Instantiate: instantiateServiceConfig,
	}
}

func instantiateServiceConfig(arguments *builtin_argument.ArgumentValuesSet) (builtin_argument.KurtosisValueType, *startosis_errors.InterpretationError) {
	kurtosisValueType, err := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(ServiceConfigTypeName, arguments)
	if err != nil {
		return nil, err
	}
	return &ServiceConfig{
		KurtosisValueTypeDefault: kurtosisValueType,
	}, nil
}

// ServiceConfig is a starlark.Value that represents a service config used in the add_service instruction
type ServiceConfig struct {
	*kurtosis_type_constructor.KurtosisValueTypeDefault
}

func (config *ServiceConfig) Copy() (builtin_argument.KurtosisValueType, error) {
	copiedValueType, err := config.KurtosisValueTypeDefault.Copy()
	if err != nil {
		return nil, err
	}
	return &ServiceConfig{
		KurtosisValueTypeDefault: copiedValueType,
	}, nil
}

func (config *ServiceConfig) ToKurtosisType(
	serviceNetwork service_network.ServiceNetwork,
	locatorOfModuleInWhichThisBuiltInIsBeingCalled string,
	packageId string,
	packageContentProvider startosis_packages.PackageContentProvider,
	packageReplaceOptions map[string]string,
) (*service.ServiceConfig, *startosis_errors.InterpretationError) {
	var ok bool

	var imageName string
	var imageBuildSpec *image_build_spec.ImageBuildSpec
	rawImageAttrValue, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.Value](config.KurtosisValueTypeDefault, ImageAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if !found {
		return nil, startosis_errors.NewInterpretationError("Required attribute '%s' could not be found on type '%s'", ImageAttr, ServiceConfigTypeName)
	}
	imageName, imageBuildSpec, interpretationErr = convertImageAttr(
		rawImageAttrValue,
		locatorOfModuleInWhichThisBuiltInIsBeingCalled,
		packageId,
		packageContentProvider,
		packageReplaceOptions)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	privatePorts := map[string]*port_spec.PortSpec{}
	privatePortsStarlark, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.Dict](config.KurtosisValueTypeDefault, PortsAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if found && privatePortsStarlark.Len() > 0 {
		for _, portItem := range privatePortsStarlark.Items() {
			portKey, portValue, interpretationError := convertPortMapEntry(PortsAttr, portItem[0], portItem[1], privatePortsStarlark)
			if interpretationError != nil {
				return nil, interpretationError
			}
			privatePorts[portKey] = portValue
		}
	}

	publicPorts := map[string]*port_spec.PortSpec{}
	publicPortsStarlark, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.Dict](config.KurtosisValueTypeDefault, PublicPortsAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if found && publicPortsStarlark.Len() > 0 {
		for _, portItem := range publicPortsStarlark.Items() {
			portKey, portValue, interpretationError := convertPortMapEntry(PublicPortsAttr, portItem[0], portItem[1], publicPortsStarlark)
			if interpretationError != nil {
				return nil, interpretationError
			}
			publicPorts[portKey] = portValue
		}
	}

	var filesArtifactExpansions *service_directory.FilesArtifactsExpansion
	var persistentDirectories *service_directory.PersistentDirectories
	filesStarlark, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.Dict](config.KurtosisValueTypeDefault, FilesAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if found {
		var filesArtifactsMountDirpathsMap map[string]string
		filesArtifactsMountDirpathsMap, persistentDirectoriesDirpathsMap, interpretationErr := convertFilesArguments(FilesAttr, filesStarlark)
		if interpretationErr != nil {
			return nil, interpretationErr
		}
		filesArtifactExpansions, interpretationErr = ConvertFilesArtifactsMounts(filesArtifactsMountDirpathsMap, serviceNetwork)
		if interpretationErr != nil {
			return nil, interpretationErr
		}
		persistentDirectories = convertPersistentDirectoryMounts(persistentDirectoriesDirpathsMap)
	}

	var entryPointArgs []string
	entrypointStarlark, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.List](config.KurtosisValueTypeDefault, EntrypointAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if found && entrypointStarlark.Len() > 0 {
		entryPointArgs, interpretationErr = kurtosis_types.SafeCastToStringSlice(entrypointStarlark, EntrypointAttr)
		if interpretationErr != nil {
			return nil, interpretationErr
		}
	}

	var cmdArgs []string
	cmdStarlark, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.List](config.KurtosisValueTypeDefault, CmdAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if found && cmdStarlark.Len() > 0 {
		cmdArgs, interpretationErr = kurtosis_types.SafeCastToStringSlice(cmdStarlark, CmdAttr)
		if interpretationErr != nil {
			return nil, interpretationErr
		}
	}

	envVars := map[string]string{}
	envVarsStarlark, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.Dict](config.KurtosisValueTypeDefault, EnvVarsAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if found && envVarsStarlark.Len() > 0 {
		envVars, interpretationErr = kurtosis_types.SafeCastToMapStringString(envVarsStarlark, EnvVarsAttr)
		if interpretationErr != nil {
			return nil, interpretationErr
		}
	}

	var privateIpAddressPlaceholder string
	privateIpAddressPlaceholderStarlark, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](config.KurtosisValueTypeDefault, PrivateIpAddressPlaceholderAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if found && privateIpAddressPlaceholderStarlark.GoString() != "" {
		privateIpAddressPlaceholder = privateIpAddressPlaceholderStarlark.GoString()
	} else {
		privateIpAddressPlaceholder = DefaultPrivateIPAddrPlaceholder
	}

	var maxCpu uint64
	maxCpuStarlark, foundMaxCpu, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.Int](config.KurtosisValueTypeDefault, MaxCpuMilliCoresAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if foundMaxCpu {
		maxCpu, ok = maxCpuStarlark.Uint64()
		if !ok {
			return nil, startosis_errors.NewInterpretationError("An error occurred parsing field '%v' with value '%v' to uint64", MaxCpuMilliCoresAttr, maxCpuStarlark)
		}
	}

	if !foundMaxCpu {
		cpuAllocationStarlark, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.Int](config.KurtosisValueTypeDefault, CpuAllocationAttr)
		if interpretationErr != nil {
			return nil, interpretationErr
		}
		if found {
			maxCpu, ok = cpuAllocationStarlark.Uint64()
			if !ok {
				return nil, startosis_errors.NewInterpretationError("An error occurred parsing field '%v' with value '%v' to uint64", CpuAllocationAttr, cpuAllocationStarlark)
			}
		}
	}

	var maxMemory uint64
	maxMemoryStarlark, foundMaxMemory, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.Int](config.KurtosisValueTypeDefault, MaxMemoryMegaBytesAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if foundMaxMemory {
		maxMemory, ok = maxMemoryStarlark.Uint64()
		if !ok {
			return nil, startosis_errors.NewInterpretationError("An error occurred parsing field '%v' with value '%v' to uint64", MaxMemoryMegaBytesAttr, maxMemoryStarlark)
		}
	}

	if !foundMaxMemory {
		memoryAllocationStarlark, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.Int](config.KurtosisValueTypeDefault, MemoryAllocationAttr)
		if interpretationErr != nil {
			return nil, interpretationErr
		}
		if found {
			maxMemory, ok = memoryAllocationStarlark.Uint64()
			if !ok {
				return nil, startosis_errors.NewInterpretationError("An error occurred parsing field '%v' with value '%v' to uint64", MemoryAllocationAttr, memoryAllocationStarlark)
			}
		}
	}

	var minCpu uint64
	minCpuStarlark, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.Int](config.KurtosisValueTypeDefault, MinCpuMilliCoresAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if found {
		minCpu, ok = minCpuStarlark.Uint64()
		if !ok {
			return nil, startosis_errors.NewInterpretationError("An error occurred parsing field '%v' with value '%v' to uint64", MinCpuMilliCoresAttr, minCpuStarlark)
		}
	}

	var minMemory uint64
	minMemoryStarlark, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.Int](config.KurtosisValueTypeDefault, MinMemoryMegaBytesAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if found {
		minMemory, ok = minMemoryStarlark.Uint64()
		if !ok {
			return nil, startosis_errors.NewInterpretationError("An error occurred parsing field '%v' with value '%v' to uint64", MinMemoryMegaBytesAttr, minMemoryStarlark)
		}
	} else {
		minMemory = 0
	}

	labels := map[string]string{}
	labelsStarlark, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.Dict](config.KurtosisValueTypeDefault, LabelsAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if found && labelsStarlark.Len() > 0 {
		labels, interpretationErr = kurtosis_types.SafeCastToMapStringString(labelsStarlark, LabelsAttr)
		if interpretationErr != nil {
			return nil, interpretationErr
		}
	}

	serviceConfig, err := service.CreateServiceConfig(
		imageName,
		imageBuildSpec,
		privatePorts,
		publicPorts,
		entryPointArgs,
		cmdArgs,
		envVars,
		filesArtifactExpansions,
		persistentDirectories,
		maxCpu,
		maxMemory,
		privateIpAddressPlaceholder,
		minCpu,
		minMemory,
		labels,
	)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "An error occurred creating a service config")
	}
	return serviceConfig, nil
}

func (config *ServiceConfig) GetReadyCondition() (*ReadyCondition, *startosis_errors.InterpretationError) {
	readyConditions, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*ReadyCondition](config.KurtosisValueTypeDefault, ReadyConditionsAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if !found {
		return nil, nil
	}

	return readyConditions, nil
}

func ConvertFilesArtifactsMounts(filesArtifactsMountDirpathsMap map[string]string, serviceNetwork service_network.ServiceNetwork) (*service_directory.FilesArtifactsExpansion, *startosis_errors.InterpretationError) {
	filesArtifactsExpansions := []args.FilesArtifactExpansion{}
	serviceDirpathsToArtifactIdentifiers := map[string]string{}
	expanderDirpathToUserServiceDirpathMap := map[string]string{}
	for mountpointOnUserService, filesArtifactIdentifier := range filesArtifactsMountDirpathsMap {
		dirpathToExpandTo := path.Join(filesArtifactExpansionDirsParentDirpath, filesArtifactIdentifier)
		expansion := args.FilesArtifactExpansion{
			FilesIdentifier:   filesArtifactIdentifier,
			DirPathToExpandTo: dirpathToExpandTo,
		}
		filesArtifactsExpansions = append(filesArtifactsExpansions, expansion)
		serviceDirpathsToArtifactIdentifiers[mountpointOnUserService] = filesArtifactIdentifier
		expanderDirpathToUserServiceDirpathMap[dirpathToExpandTo] = mountpointOnUserService
	}

	// TODO: Sad that we need the service network here to get the APIC info. This is wrong, we should fix this by
	//  passing the APIC info DOWN to the backend and have the backend create the expander itself.
	//  Here writing those info into each service config is dumb
	apiContainerInfo := serviceNetwork.GetApiContainerInfo()
	filesArtifactsExpanderArgs, err := args.NewFilesArtifactsExpanderArgs(
		apiContainerInfo.GetIpAddress().String(),
		apiContainerInfo.GetGrpcPortNum(),
		filesArtifactsExpansions,
	)
	if err != nil {
		return nil, startosis_errors.NewInterpretationError("An error occurred creating files artifacts expander args")
	}

	expanderEnvVars, err := args.GetEnvFromArgs(filesArtifactsExpanderArgs)
	if err != nil {
		return nil, startosis_errors.NewInterpretationError("An error occurred getting files artifacts expander environment variables using args: %+v", filesArtifactsExpanderArgs)
	}

	expanderImageAndTag := fmt.Sprintf(
		"%v:%v",
		filesArtifactsExpanderImage,
		apiContainerInfo.GetVersion(),
	)

	return &service_directory.FilesArtifactsExpansion{
		ExpanderImage:                        expanderImageAndTag,
		ExpanderEnvVars:                      expanderEnvVars,
		ServiceDirpathsToArtifactIdentifiers: serviceDirpathsToArtifactIdentifiers,
		ExpanderDirpathsToServiceDirpaths:    expanderDirpathToUserServiceDirpathMap,
	}, nil
}

func convertPersistentDirectoryMounts(persistentDirectoriesMap map[string]service_directory.PersistentDirectory) *service_directory.PersistentDirectories {
	return service_directory.NewPersistentDirectories(persistentDirectoriesMap)
}

func convertPortMapEntry(attrNameForLogging string, key starlark.Value, value starlark.Value, dictForLogging *starlark.Dict) (string, *port_spec.PortSpec, *startosis_errors.InterpretationError) {
	keyStr, ok := key.(starlark.String)
	if !ok {
		return "", nil, startosis_errors.NewInterpretationError("Unable to convert key of '%s' dictionary '%v' to string", attrNameForLogging, dictForLogging)
	}
	valuePortSpec, ok := value.(*starlark_port_spec.PortSpec)
	if !ok {
		return "", nil, startosis_errors.NewInterpretationError("Unable to convert value of '%s' dictionary '%v' to a port object", attrNameForLogging, dictForLogging)
	}
	servicePortSpec, interpretationErr := valuePortSpec.ToKurtosisType()
	if interpretationErr != nil {
		return "", nil, interpretationErr
	}
	return keyStr.GoString(), servicePortSpec, nil
}

func convertFilesArguments(attrNameForLogging string, filesDict *starlark.Dict) (map[string]string, map[string]service_directory.PersistentDirectory, *startosis_errors.InterpretationError) {
	filesArtifacts := map[string]string{}
	persistentDirectories := map[string]service_directory.PersistentDirectory{}
	for _, fileItem := range filesDict.Items() {
		rawDirPath := fileItem[0]
		dirPath, ok := rawDirPath.(starlark.String)
		if !ok {
			return nil, nil, startosis_errors.NewInterpretationError("Unable to convert key of '%s' dictionary '%v' to string", attrNameForLogging, filesDict)
		}

		var interpretationErr *startosis_errors.InterpretationError
		rawDirectoryObj := fileItem[1]
		directoryObj, isDirectoryArg := rawDirectoryObj.(*directory.Directory)
		if !isDirectoryArg {
			// we're also supporting raw strings as well and transform them into files artifact name.
			directoryObjAsStr, isSimpleStringArg := rawDirectoryObj.(starlark.String)
			if !isSimpleStringArg {
				return nil, nil, startosis_errors.NewInterpretationError("Unable to convert value of '%s' dictionary '%v' to a Directory object", attrNameForLogging, filesDict)
			}
			directoryObj, interpretationErr = directory.CreateDirectoryFromFilesArtifact(directoryObjAsStr.GoString())
			if interpretationErr != nil {
				return nil, nil, interpretationErr
			}
		}
		artifactName, artifactNameSet, interpretationErr := directoryObj.GetArtifactNameIfSet()
		if interpretationErr != nil {
			return nil, nil, interpretationErr
		}
		persistentKey, persistentKeySet, interpretationErr := directoryObj.GetPersistentKeyIfSet()
		if interpretationErr != nil {
			return nil, nil, interpretationErr
		}
		persistentDirectorySize, interpretationErr := directoryObj.GetSizeOrDefault()
		if interpretationErr != nil {
			return nil, nil, interpretationErr
		}
		if artifactNameSet == persistentKeySet {
			// this condition is a XOR
			return nil, nil, startosis_errors.NewInterpretationError("Parameter '%s' and '%s' cannot be set on the same '%s' object: '%s'",
				directory.ArtifactNameAttr, directory.PersistentKeyAttr, directory.DirectoryTypeName, directoryObj.String())
		}
		if artifactNameSet {
			filesArtifacts[dirPath.GoString()] = artifactName
		} else {
			// persistentKey is necessarily set since we checked the exclusivity above
			persistentDirectories[dirPath.GoString()] = service_directory.PersistentDirectory{
				PersistentKey: service_directory.DirectoryPersistentKey(persistentKey),
				Size:          service_directory.DirectoryPersistentSize(persistentDirectorySize),
			}
		}
	}
	return filesArtifacts, persistentDirectories, nil
}

// If [rawImageAttrValue] is a string, returns the image name with no image build spec (image will be fetched from local cache or remote)
// If [rawImageAttrValue] is an ImageBuildSpec type, name for the image to build and ImageBuildSpec converted to KurtosisType is returned (image will be built)
func convertImageAttr(
	rawImageAttrValue starlark.Value,
	locatorOfModuleInWhichThisBuiltInIsBeingCalled string,
	packageId string,
	packageContentProvider startosis_packages.PackageContentProvider,
	packageReplaceOptions map[string]string) (string, *image_build_spec.ImageBuildSpec, *startosis_errors.InterpretationError) {
	imageName, interpretationErr := kurtosis_types.SafeCastToString(rawImageAttrValue, ImageAttr)
	if interpretationErr == nil {
		return imageName, nil, nil
	} else {
		imageBuildSpecStarlarkType, isImageBuildSpecStarlarkType := rawImageAttrValue.(*ImageBuildSpec)
		if !isImageBuildSpecStarlarkType {
			return "", nil, startosis_errors.NewInterpretationError("Failed to cast '%v' to an image build spec object.", rawImageAttrValue)
		}
		imageBuildSpec, interpretationErr := imageBuildSpecStarlarkType.ToKurtosisType(locatorOfModuleInWhichThisBuiltInIsBeingCalled, packageId, packageContentProvider, packageReplaceOptions)
		if interpretationErr != nil {
			return "", nil, interpretationErr
		}
		imageName, interpretationErr = imageBuildSpecStarlarkType.GetImageName()
		if interpretationErr != nil {
			return "", nil, interpretationErr
		}
		return imageName, imageBuildSpec, nil
	}
}
