package plan_yaml

import (
	"fmt"
	"github.com/go-yaml/yaml"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	store_spec2 "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/store_spec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"strconv"
	"strings"
)

const (
	HTTP ApplicationProtocol = "HTTP"
	UDP  TransportProtocol   = "UDP"
	TCP  TransportProtocol   = "TCP"

	SHELL  TaskType = "sh"
	PYTHON TaskType = "python"
	EXEC   TaskType = "exec"
)

// PlanYaml is a yaml representation of the effect of an "plan" or sequence of instructions on the state of the Enclave.
type PlanYaml struct {
	privatePlanYaml *privatePlanYaml

	futureReferenceIndex map[string]string
	filesArtifactIndex   map[string]*FilesArtifact
	serviceIndex         map[string]*Service
	latestUuid           int
}

// TODO: pass by value instead of pass by reference
type privatePlanYaml struct {
	PackageId      string           `yaml:"packageId,omitempty"`
	Services       []*Service       `yaml:"services,omitempty"`
	FilesArtifacts []*FilesArtifact `yaml:"filesArtifacts,omitempty"`
	Tasks          []*Task          `yaml:"tasks,omitempty"`
}

// Service represents a service in the system.
type Service struct {
	Uuid       string                 `yaml:"uuid,omitempty"`
	Name       string                 `yaml:"name,omitempty"`
	Image      *ImageSpec             `yaml:"image,omitempty"`
	Cmd        []string               `yaml:"command,omitempty"`
	Entrypoint []string               `yaml:"entrypoint,omitempty"`
	EnvVars    []*EnvironmentVariable `yaml:"envVars,omitempty"`
	Ports      []*Port                `yaml:"ports,omitempty"`
	Files      []*FileMount           `yaml:"files,omitempty"`
}

type ImageSpec struct {
	ImageName string `yaml:"name,omitempty"`

	// for built images
	BuildContextLocator string `yaml:"buildContextLocator,omitempty"`
	TargetStage         string `yaml:"targetStage,omitempty"`

	// for images from registry
	Registry string `yaml:"registry,omitempty"`
}

// FilesArtifact represents a collection of files.
type FilesArtifact struct {
	Uuid  string   `yaml:"uuid,omitempty"`
	Name  string   `yaml:"name,omitempty"`
	Files []string `yaml:"files,omitempty"`
}

// EnvironmentVariable represents an environment variable.
type EnvironmentVariable struct {
	Key   string `yaml:"key,omitempty"`
	Value string `yaml:"value,omitempty"`
}

// Port represents a port.
type Port struct {
	Name   string `yaml:"name,omitempty"`
	Number uint16 `yaml:"number,omitempty"`

	TransportProtocol   TransportProtocol   `yaml:"transportProtocol,omitempty"`
	ApplicationProtocol ApplicationProtocol `yaml:"applicationProtocol,omitempty"`
}

// ApplicationProtocol represents the application protocol used.
type ApplicationProtocol string

// TransportProtocol represents transport protocol used.
type TransportProtocol string

// FileMount represents a mount point for files.
type FileMount struct {
	MountPath      string           `yaml:"mountPath,omitempty"`
	FilesArtifacts []*FilesArtifact `yaml:"filesArtifacts,omitempty"` // TODO: support persistent directories
}

// Task represents a task to be executed.
type Task struct {
	Uuid     string           `yaml:"uuid,omitempty"`     // done
	Name     string           `yaml:"name,omitempty"`     // done
	TaskType TaskType         `yaml:"taskType,omitempty"` // done
	RunCmd   []string         `yaml:"command,omitempty"`  // done
	Image    string           `yaml:"image,omitempty"`    // done
	Files    []*FileMount     `yaml:"files,omitempty"`
	Store    []*FilesArtifact `yaml:"store,omitempty"`

	// only exists on SHELL tasks
	EnvVars []*EnvironmentVariable `yaml:"envVar,omitempty"` // done

	// only exists on PYTHON tasks
	PythonPackages []string `yaml:"pythonPackages,omitempty"`
	PythonArgs     []string `yaml:"pythonArgs,omitempty"`

	// service name
	ServiceName     string  `yaml:"serviceName,omitempty"`
	AcceptableCodes []int64 `yaml:"acceptableCodes,omitempty"`
}

// TaskType represents the type of task (either PYTHON or SHELL)
type TaskType string

func CreateEmptyPlan(packageId string) *PlanYaml {
	return &PlanYaml{
		privatePlanYaml: &privatePlanYaml{
			PackageId:      packageId,
			Services:       []*Service{},
			Tasks:          []*Task{},
			FilesArtifacts: []*FilesArtifact{},
		},
		serviceIndex:         map[string]*Service{},
		futureReferenceIndex: map[string]string{},
		filesArtifactIndex:   map[string]*FilesArtifact{},
	}
}

func (planYaml *PlanYaml) GenerateYaml() (string, error) {
	yamlBytes, err := yaml.Marshal(planYaml.privatePlanYaml)
	if err != nil {
		return "", err
	}
	return string(yamlBytes), nil
}

func (planYaml *PlanYaml) AddService(
	serviceName service.ServiceName,
	serviceInfo *kurtosis_types.Service,
	serviceConfig *service.ServiceConfig,
	imageValue starlark.Value,
) error {
	uuid := planYaml.generateUuid()

	// store future references of this service
	ipAddrFutureRef, err := serviceInfo.GetIpAddress()
	if err != nil {
		return err
	}
	hostnameFutureRef, err := serviceInfo.GetHostname()
	if err != nil {
		return err
	}
	planYaml.storeFutureReference(uuid, ipAddrFutureRef, "ip_address")
	planYaml.storeFutureReference(uuid, hostnameFutureRef, "hostname")

	// construct service yaml object for plan
	serviceYaml := &Service{}
	serviceYaml.Uuid = uuid

	serviceYaml.Name = planYaml.swapFutureReference(string(serviceName))

	image := &ImageSpec{ //nolint:exhaustruct
		ImageName: serviceConfig.GetContainerImageName(),
	}
	//imageBuildSpec := serviceConfig.GetImageBuildSpec()
	//if imageBuildSpec != nil {
	//	// Need the raw imageValue to get the build context locator
	//	switch starlarkImgVal := imageValue.(type) {
	//	case *service_config.ImageBuildSpec: // importing service_config.ImageBuildSpec causes a dependency issue figure that out later
	//		contextLocator, err := starlarkImgVal.GetBuildContextLocator()
	//		if err != nil {
	//			return err
	//		}
	//		image.BuildContextLocator = contextLocator
	//	default:
	//		return stacktrace.NewError("An image build spec was detected on the kurtosis type service config but the starlark image value was not an ImageBuildSpec type.")
	//	}
	//	image.TargetStage = imageBuildSpec.GetTargetStage()
	//}
	//imageSpec := serviceConfig.GetImageRegistrySpec()
	//if imageSpec != nil {
	//	image.Registry = imageSpec.GetRegistryAddr()
	//}
	serviceYaml.Image = image

	cmdArgs := []string{}
	for _, cmdArg := range serviceConfig.GetCmdArgs() {
		realCmdArg := planYaml.swapFutureReference(cmdArg)
		cmdArgs = append(cmdArgs, realCmdArg)
	}
	serviceYaml.Cmd = cmdArgs

	entryArgs := []string{}
	for _, entryArg := range serviceConfig.GetEntrypointArgs() {
		realEntryArg := planYaml.swapFutureReference(entryArg)
		entryArgs = append(entryArgs, realEntryArg)
	}
	serviceYaml.Entrypoint = entryArgs

	serviceYaml.Ports = []*Port{}
	for portName, configPort := range serviceConfig.GetPrivatePorts() { // TODO: support public ports
		var applicationProtocolStr string
		if configPort.GetMaybeApplicationProtocol() != nil {
			applicationProtocolStr = *configPort.GetMaybeApplicationProtocol()
		}
		port := &Port{
			TransportProtocol:   TransportProtocol(configPort.GetTransportProtocol().String()),
			ApplicationProtocol: ApplicationProtocol(applicationProtocolStr), // TODO: write a test for this, dereferencing config port is not a good idea
			Name:                portName,
			Number:              configPort.GetNumber(),
		}
		serviceYaml.Ports = append(serviceYaml.Ports, port)
	}

	serviceYaml.EnvVars = []*EnvironmentVariable{}
	for key, val := range serviceConfig.GetEnvVars() {
		envVar := &EnvironmentVariable{
			Key:   key,
			Value: planYaml.swapFutureReference(val),
		}
		serviceYaml.EnvVars = append(serviceYaml.EnvVars, envVar)
	}

	// file mounts have two cases:
	// 1. the referenced files artifact already exists in the plan, in which case add the referenced files artifact
	// 2. the referenced files artifact does not already exist in the plan, in which case the file MUST have been passed in via a top level arg OR is invalid
	// 	  for this case,
	// 	  - create new files artifact
	//	  - add it to the service's file mount accordingly
	//	  - add the files artifact to the plan
	serviceYaml.Files = []*FileMount{}
	serviceFilesArtifactExpansions := serviceConfig.GetFilesArtifactsExpansion()
	if serviceFilesArtifactExpansions != nil {
		for mountPath, artifactIdentifiers := range serviceFilesArtifactExpansions.ServiceDirpathsToArtifactIdentifiers {
			var serviceFilesArtifacts []*FilesArtifact
			for _, identifier := range artifactIdentifiers {
				var filesArtifact *FilesArtifact
				// if there's already a files artifact that exists with this name from a previous instruction, reference that
				if filesArtifactToReference, ok := planYaml.filesArtifactIndex[identifier]; ok {
					filesArtifact = &FilesArtifact{
						Name:  filesArtifactToReference.Name,
						Uuid:  filesArtifactToReference.Uuid,
						Files: []string{}, // leave empty because this is referencing an existing files artifact
					}
				} else {
					// otherwise create a new one
					// the only information we have about a files artifact that didn't already exist is the name
					// if it didn't already exist AND interpretation was successful, it MUST HAVE been passed in via args of run function
					filesArtifact = &FilesArtifact{
						Name:  identifier,
						Uuid:  planYaml.generateUuid(),
						Files: []string{}, // don't know at interpretation what files are on the artifact when passed in via args
					}
					planYaml.addFilesArtifactYaml(filesArtifact)
				}
				serviceFilesArtifacts = append(serviceFilesArtifacts, filesArtifact)
			}

			serviceYaml.Files = append(serviceYaml.Files, &FileMount{
				MountPath:      mountPath,
				FilesArtifacts: serviceFilesArtifacts,
			})
		}

	}

	planYaml.addServiceYaml(serviceYaml)
	return nil
}

func (planYaml *PlanYaml) AddRunSh(
	runCommand string,
	returnValue *starlarkstruct.Struct,
	serviceConfig *service.ServiceConfig,
	storeSpecList []*store_spec2.StoreSpec,
) error {
	uuid := planYaml.generateUuid()

	// store run sh future references
	starlarkCodeVal, err := returnValue.Attr("code")
	if err != nil {
		return err
	}
	starlarkCodeFutureRefStr, interpErr := kurtosis_types.SafeCastToString(starlarkCodeVal, "run sh code")
	if interpErr != nil {
		return interpErr
	}
	planYaml.storeFutureReference(uuid, starlarkCodeFutureRefStr, "code")
	starlarkOutputVal, err := returnValue.Attr("output")
	if err != nil {
		return err
	}
	starlarkOutputFutureRefStr, interpErr := kurtosis_types.SafeCastToString(starlarkOutputVal, "run sh code")
	if interpErr != nil {
		return interpErr
	}
	planYaml.storeFutureReference(uuid, starlarkOutputFutureRefStr, "output")

	// create task yaml object
	taskYaml := &Task{}
	taskYaml.Uuid = uuid

	taskYaml.RunCmd = []string{runCommand}
	taskYaml.Image = serviceConfig.GetContainerImageName()

	var envVars []*EnvironmentVariable
	for key, val := range serviceConfig.GetEnvVars() {
		envVars = append(envVars, &EnvironmentVariable{
			Key:   key,
			Value: planYaml.swapFutureReference(val),
		})
	}
	taskYaml.EnvVars = envVars

	// for files:
	//	1. either the referenced files artifact already exists in the plan, in which case, look for it and reference it via instruction uuid
	// 	2. the referenced files artifact is new, in which case we add it to the plan
	for mountPath, fileArtifactName := range serviceConfig.GetFilesArtifactsExpansion().ServiceDirpathsToArtifactIdentifiers {
		var filesArtifact *FilesArtifact
		// if there's already a files artifact that exists with this name from a previous instruction, reference that
		if potentialFilesArtifact, ok := planYaml.filesArtifactIndex[fileArtifactName]; ok {
			filesArtifact = &FilesArtifact{ //nolint:exhaustruct
				Name: potentialFilesArtifact.Name,
				Uuid: potentialFilesArtifact.Uuid,
			}
		} else {
			// otherwise create a new one
			// the only information we have about a files artifact that didn't already exist is the name
			// if it didn't already exist AND interpretation was successful, it MUST HAVE been passed in via args
			filesArtifact = &FilesArtifact{ //nolint:exhaustruct
				Name: fileArtifactName,
				Uuid: planYaml.generateUuid(),
			}
			// add to the index and append to the plan yaml
			planYaml.addFilesArtifactYaml(filesArtifact)
		}

		taskYaml.Files = append(taskYaml.Files, &FileMount{
			MountPath:      mountPath,
			FilesArtifacts: []*FilesArtifact{filesArtifact},
		})
	}

	// for store
	// - all files artifacts product from store are new files artifact that are added to the plan
	//		- add them to files artifacts list
	// 		- add them to the store section of run sh
	var store []*FilesArtifact
	for _, storeSpec := range storeSpecList {
		// add the FilesArtifact to list of all files artifacts and index
		uuid := planYaml.generateUuid()
		var newFilesArtifactFromStoreSpec = &FilesArtifact{
			Uuid:  uuid,
			Name:  storeSpec.GetName(),
			Files: []string{storeSpec.GetSrc()},
		}
		planYaml.addFilesArtifactYaml(newFilesArtifactFromStoreSpec)
		store = append(store, &FilesArtifact{
			Uuid:  uuid,
			Name:  storeSpec.GetName(),
			Files: []string{}, // don't want to repeat the files on a referenced files artifact
		})
	}
	taskYaml.Store = store

	// add task to index, do we even need a tasks index?
	planYaml.addTaskYaml(taskYaml)
	return nil
}

func (planYaml *PlanYaml) addServiceYaml(service *Service) {
	planYaml.serviceIndex[service.Name] = service
	planYaml.privatePlanYaml.Services = append(planYaml.privatePlanYaml.Services, service)
}

func (planYaml *PlanYaml) addFilesArtifactYaml(filesArtifact *FilesArtifact) {
	planYaml.filesArtifactIndex[filesArtifact.Name] = filesArtifact
	// do we need both map and list structures? what about just map then resolve at the end?
	planYaml.privatePlanYaml.FilesArtifacts = append(planYaml.privatePlanYaml.FilesArtifacts, filesArtifact)
}

func (planYaml *PlanYaml) addTaskYaml(task *Task) {
	planYaml.privatePlanYaml.Tasks = append(planYaml.privatePlanYaml.Tasks, task)
}

func (planYaml *PlanYaml) storeFutureReference(uuid, futureReference, futureReferenceType string) {
	planYaml.futureReferenceIndex[futureReference] = fmt.Sprintf("{{ kurtosis.%v.%v }}", uuid, futureReferenceType)
}

// swapFutureReference replaces all future references in s, if any exist, with the value required for the yaml format
func (planYaml *PlanYaml) swapFutureReference(s string) string {
	swappedString := s
	for futureRef, yamlFutureRef := range planYaml.futureReferenceIndex {
		if strings.Contains(s, futureRef) {
			swappedString = strings.Replace(s, futureRef, yamlFutureRef, -1) // -1 to swap all instances of [futureRef]
		}
	}
	return swappedString
}

func (planYaml *PlanYaml) generateUuid() string {
	planYaml.latestUuid++
	return strconv.Itoa(planYaml.latestUuid)
}
