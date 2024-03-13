package plan_yaml

import (
	"fmt"
	"github.com/go-yaml/yaml"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	store_spec2 "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/store_spec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"golang.org/x/exp/slices"
	"strconv"
	"strings"
)

const (
	ipAddressFutureRefType = "ip_address"
	codeFutureRefType      = "code"
	hostnameFutureRefType  = "hostname"
	outputFutureRefType    = "output"
)

// PlanYaml is a yaml representation of the effect of an Instructions Plan or sequence of instructions on the state of the Enclave.
type PlanYaml struct {
	privatePlanYaml *privatePlanYaml

	futureReferenceIndex map[string]string
	filesArtifactIndex   map[string]*FilesArtifact
	latestUuid           int
}

func CreateEmptyPlan(packageId string) *PlanYaml {
	return &PlanYaml{
		privatePlanYaml: &privatePlanYaml{
			PackageId:      packageId,
			Services:       []*Service{},
			Tasks:          []*Task{},
			FilesArtifacts: []*FilesArtifact{},
		},
		futureReferenceIndex: map[string]string{},
		filesArtifactIndex:   map[string]*FilesArtifact{},
		latestUuid:           0,
	}
}

func (planYaml *PlanYaml) GenerateYaml() (string, error) {
	yamlBytes, err := yaml.Marshal(planYaml.privatePlanYaml)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred generating plan yaml.")
	}
	return string(yamlBytes), nil
}

func (planYaml *PlanYaml) AddService(
	serviceName service.ServiceName,
	serviceInfo *kurtosis_types.Service,
	serviceConfig *service.ServiceConfig,
	// image values might be empty depending on how the image is buitl
	imageBuildContextLocator string,
	imageTargetStage string,
	imageRegistryAddress string,
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
	planYaml.storeFutureReference(uuid, ipAddrFutureRef, ipAddressFutureRefType)
	planYaml.storeFutureReference(uuid, hostnameFutureRef, hostnameFutureRefType)

	// construct service yaml object for plan
	serviceYaml := &Service{} //nolint exhaustruct
	serviceYaml.Uuid = uuid

	serviceYaml.Name = planYaml.swapFutureReference(string(serviceName))

	imageYaml := &ImageSpec{} //nolint:exhaustruct
	imageYaml.ImageName = serviceConfig.GetContainerImageName()
	imageYaml.BuildContextLocator = imageBuildContextLocator
	imageYaml.TargetStage = imageTargetStage
	imageYaml.Registry = imageRegistryAddress
	serviceYaml.Image = imageYaml

	cmdArgs := []string{}
	for _, cmdArg := range serviceConfig.GetCmdArgs() {
		cmdArgs = append(cmdArgs, planYaml.swapFutureReference(cmdArg))
	}
	serviceYaml.Cmd = cmdArgs

	entryArgs := []string{}
	for _, entryArg := range serviceConfig.GetEntrypointArgs() {
		entryArgs = append(entryArgs, planYaml.swapFutureReference(entryArg))
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
			ApplicationProtocol: ApplicationProtocol(applicationProtocolStr),
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
	if serviceConfig.GetFilesArtifactsExpansion() != nil {
		for mountPath, artifactIdentifiers := range serviceConfig.GetFilesArtifactsExpansion().ServiceDirpathsToArtifactIdentifiers {
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
	codeVal, err := returnValue.Attr(codeFutureRefType)
	if err != nil {
		return err
	}
	codeFutureRef, interpErr := kurtosis_types.SafeCastToString(codeVal, "run sh "+codeFutureRefType)
	if interpErr != nil {
		return interpErr
	}
	planYaml.storeFutureReference(uuid, codeFutureRef, codeFutureRefType)
	outputVal, err := returnValue.Attr(outputFutureRefType)
	if err != nil {
		return err
	}
	outputFutureRef, interpErr := kurtosis_types.SafeCastToString(outputVal, "run sh "+outputFutureRefType)
	if interpErr != nil {
		return interpErr
	}
	planYaml.storeFutureReference(uuid, outputFutureRef, outputFutureRefType)

	// create task yaml object
	taskYaml := &Task{} //nolint exhaustruct
	taskYaml.Uuid = uuid
	taskYaml.TaskType = shell

	taskYaml.RunCmd = []string{planYaml.swapFutureReference(runCommand)}
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
	if serviceConfig.GetFilesArtifactsExpansion() != nil {
		for mountPath, fileArtifactNames := range serviceConfig.GetFilesArtifactsExpansion().ServiceDirpathsToArtifactIdentifiers {
			var filesArtifacts []*FilesArtifact
			for _, filesArtifactName := range fileArtifactNames {
				var filesArtifact *FilesArtifact
				// if there's already a files artifact that exists with this name from a previous instruction, reference that
				if filesArtifactToReference, ok := planYaml.filesArtifactIndex[filesArtifactName]; ok {
					filesArtifact = &FilesArtifact{
						Name:  filesArtifactToReference.Name,
						Uuid:  filesArtifactToReference.Uuid,
						Files: []string{},
					}
				} else {
					// otherwise create a new one
					// the only information we have about a files artifact that didn't already exist is the name
					// if it didn't already exist AND interpretation was successful, it MUST HAVE been passed in via args
					filesArtifact = &FilesArtifact{
						Name:  filesArtifactName,
						Uuid:  planYaml.generateUuid(),
						Files: []string{},
					}
					planYaml.addFilesArtifactYaml(filesArtifact)
				}
				filesArtifacts = append(filesArtifacts, filesArtifact)
			}

			taskYaml.Files = append(taskYaml.Files, &FileMount{
				MountPath:      mountPath,
				FilesArtifacts: filesArtifacts,
			})
		}
	}

	// for store
	// - all files artifacts product from store are new files artifact that are added to the plan
	//		- add them to files artifacts list
	// 		- add them to the store section of run sh
	var store []*FilesArtifact
	for _, storeSpec := range storeSpecList {
		var newFilesArtifactFromStoreSpec = &FilesArtifact{
			Uuid:  planYaml.generateUuid(),
			Name:  storeSpec.GetName(),
			Files: []string{storeSpec.GetSrc()},
		}
		planYaml.addFilesArtifactYaml(newFilesArtifactFromStoreSpec)

		store = append(store, &FilesArtifact{
			Uuid:  newFilesArtifactFromStoreSpec.Uuid,
			Name:  newFilesArtifactFromStoreSpec.Name,
			Files: []string{}, // don't want to repeat the files on a referenced files artifact
		})
	}
	taskYaml.Store = store

	planYaml.addTaskYaml(taskYaml)
	return nil
}

func (planYaml *PlanYaml) AddRunPython(
	runCommand string,
	returnValue *starlarkstruct.Struct,
	serviceConfig *service.ServiceConfig,
	storeSpecList []*store_spec2.StoreSpec,
	pythonArgs []string,
	pythonPackages []string) error {
	uuid := planYaml.generateUuid()

	// store future references
	codeVal, err := returnValue.Attr(codeFutureRefType)
	if err != nil {
		return err
	}
	codeFutureRef, interpErr := kurtosis_types.SafeCastToString(codeVal, "run python "+codeFutureRefType)
	if interpErr != nil {
		return interpErr
	}
	planYaml.storeFutureReference(uuid, codeFutureRef, codeFutureRefType)
	outputVal, err := returnValue.Attr(outputFutureRefType)
	if err != nil {
		return err
	}
	outputFutureRef, interpErr := kurtosis_types.SafeCastToString(outputVal, "run python "+outputFutureRefType)
	if interpErr != nil {
		return interpErr
	}
	planYaml.storeFutureReference(uuid, outputFutureRef, outputFutureRefType)

	// create task yaml object
	taskYaml := &Task{} //nolint exhaustruct
	taskYaml.Uuid = uuid
	taskYaml.TaskType = python

	taskYaml.RunCmd = []string{planYaml.swapFutureReference(runCommand)}
	taskYaml.Image = serviceConfig.GetContainerImageName()

	var envVars []*EnvironmentVariable
	for key, val := range serviceConfig.GetEnvVars() {
		envVars = append(envVars, &EnvironmentVariable{
			Key:   key,
			Value: planYaml.swapFutureReference(val),
		})
	}
	taskYaml.EnvVars = envVars

	// python args and python packages
	taskYaml.PythonArgs = append(taskYaml.PythonArgs, pythonArgs...)
	taskYaml.PythonPackages = append(taskYaml.PythonPackages, pythonPackages...)

	// for files:
	//	1. either the referenced files artifact already exists in the plan, in which case, look for it and reference it via instruction uuid
	// 	2. the referenced files artifact is new, in which case we add it to the plan
	if serviceConfig.GetFilesArtifactsExpansion() != nil {
		for mountPath, fileArtifactNames := range serviceConfig.GetFilesArtifactsExpansion().ServiceDirpathsToArtifactIdentifiers {
			var filesArtifacts []*FilesArtifact
			for _, filesArtifactName := range fileArtifactNames {
				var filesArtifact *FilesArtifact
				// if there's already a files artifact that exists with this name from a previous instruction, reference that
				if filesArtifactToReference, ok := planYaml.filesArtifactIndex[filesArtifactName]; ok {
					filesArtifact = &FilesArtifact{
						Name:  filesArtifactToReference.Name,
						Uuid:  filesArtifactToReference.Uuid,
						Files: []string{},
					}
				} else {
					// otherwise create a new one
					// the only information we have about a files artifact that didn't already exist is the name
					// if it didn't already exist AND interpretation was successful, it MUST HAVE been passed in via args
					filesArtifact = &FilesArtifact{
						Name:  filesArtifactName,
						Uuid:  planYaml.generateUuid(),
						Files: []string{},
					}
					planYaml.addFilesArtifactYaml(filesArtifact)
				}
				filesArtifacts = append(filesArtifacts, filesArtifact)
			}

			taskYaml.Files = append(taskYaml.Files, &FileMount{
				MountPath:      mountPath,
				FilesArtifacts: filesArtifacts,
			})
		}
	}

	// for store
	// - all files artifacts product from store are new files artifact that are added to the plan
	//		- add them to files artifacts list
	// 		- add them to the store section of run sh
	var store []*FilesArtifact
	for _, storeSpec := range storeSpecList {
		var newFilesArtifactFromStoreSpec = &FilesArtifact{
			Uuid:  planYaml.generateUuid(),
			Name:  storeSpec.GetName(),
			Files: []string{storeSpec.GetSrc()},
		}
		planYaml.addFilesArtifactYaml(newFilesArtifactFromStoreSpec)

		store = append(store, &FilesArtifact{
			Uuid:  newFilesArtifactFromStoreSpec.Uuid,
			Name:  newFilesArtifactFromStoreSpec.Name,
			Files: []string{}, // don't want to repeat the files on a referenced files artifact
		})
	}
	taskYaml.Store = store

	planYaml.addTaskYaml(taskYaml)
	return nil
}

func (planYaml *PlanYaml) AddExec(
	serviceName string,
	returnValue *starlark.Dict,
	cmdList []string,
	acceptableCodes []int64) error {
	uuid := planYaml.generateUuid()

	// store future references
	codeVal, found, err := returnValue.Get(starlark.String(codeFutureRefType))
	if err != nil {
		return err
	}
	if !found {
		return stacktrace.NewError("No code value found on exec dict")
	}
	codeFutureRef, interpErr := kurtosis_types.SafeCastToString(codeVal, "exec "+codeFutureRefType)
	if interpErr != nil {
		return interpErr
	}
	planYaml.storeFutureReference(uuid, codeFutureRef, codeFutureRefType)
	outputVal, found, err := returnValue.Get(starlark.String(outputFutureRefType))
	if err != nil {
		return err
	}
	if !found {
		return stacktrace.NewError("No code value found on exec dict")
	}
	outputFutureRef, interpErr := kurtosis_types.SafeCastToString(outputVal, "exec "+outputFutureRefType)
	if interpErr != nil {
		return interpErr
	}
	planYaml.storeFutureReference(uuid, outputFutureRef, outputFutureRefType)

	// create task yaml
	taskYaml := &Task{} //nolint exhaustruct
	taskYaml.Uuid = uuid
	taskYaml.TaskType = exec
	taskYaml.ServiceName = serviceName

	cmdListWithFutureRefsSwapped := []string{}
	for _, cmd := range cmdList {
		cmdListWithFutureRefsSwapped = append(cmdListWithFutureRefsSwapped, planYaml.swapFutureReference(cmd))
	}
	taskYaml.RunCmd = cmdListWithFutureRefsSwapped
	taskYaml.AcceptableCodes = acceptableCodes

	planYaml.privatePlanYaml.Tasks = append(planYaml.privatePlanYaml.Tasks, taskYaml)
	return nil
}

func (planYaml *PlanYaml) AddRenderTemplates(filesArtifactName string, filepaths []string) error {
	uuid := planYaml.generateUuid()
	filesArtifactYaml := &FilesArtifact{} //nolint exhaustruct
	filesArtifactYaml.Uuid = uuid
	filesArtifactYaml.Name = filesArtifactName
	filesArtifactYaml.Files = filepaths
	planYaml.addFilesArtifactYaml(filesArtifactYaml)
	return nil
}

func (planYaml *PlanYaml) AddUploadFiles(filesArtifactName, locator string) error {
	uuid := planYaml.generateUuid()
	filesArtifactYaml := &FilesArtifact{} //nolint exhauststruct
	filesArtifactYaml.Uuid = uuid
	filesArtifactYaml.Name = filesArtifactName
	filesArtifactYaml.Files = []string{locator}
	planYaml.addFilesArtifactYaml(filesArtifactYaml)
	return nil
}

func (planYaml *PlanYaml) AddStoreServiceFiles(filesArtifactName, locator string) error {
	uuid := planYaml.generateUuid()
	filesArtifactYaml := &FilesArtifact{} //nolint exhaustruct
	filesArtifactYaml.Uuid = uuid
	filesArtifactYaml.Name = filesArtifactName
	filesArtifactYaml.Files = []string{locator}
	planYaml.addFilesArtifactYaml(filesArtifactYaml)
	return nil
}

func (planYaml *PlanYaml) RemoveService(serviceName string) {
	for idx, service := range planYaml.privatePlanYaml.Services {
		if service.Name == serviceName {
			planYaml.privatePlanYaml.Services = slices.Delete(planYaml.privatePlanYaml.Services, idx, idx+1)
			return
		}
	}
}

func (planYaml *PlanYaml) addServiceYaml(service *Service) {
	planYaml.privatePlanYaml.Services = append(planYaml.privatePlanYaml.Services, service)
}

func (planYaml *PlanYaml) addFilesArtifactYaml(filesArtifact *FilesArtifact) {
	planYaml.filesArtifactIndex[filesArtifact.Name] = filesArtifact
	planYaml.privatePlanYaml.FilesArtifacts = append(planYaml.privatePlanYaml.FilesArtifacts, filesArtifact)
}

func (planYaml *PlanYaml) addTaskYaml(task *Task) {
	planYaml.privatePlanYaml.Tasks = append(planYaml.privatePlanYaml.Tasks, task)
}

// yaml future reference format: {{ kurtosis.<assigned uuid>.<future reference type }}
func (planYaml *PlanYaml) storeFutureReference(uuid, futureReference, futureReferenceType string) {
	planYaml.futureReferenceIndex[futureReference] = fmt.Sprintf("{{ kurtosis.%v.%v }}", uuid, futureReferenceType)
}

// swapFutureReference replaces all future references in s, if any exist, with the value required for the yaml format
func (planYaml *PlanYaml) swapFutureReference(s string) string {
	swappedString := s
	for futureRef, yamlFutureRef := range planYaml.futureReferenceIndex {
		swappedString = strings.Replace(swappedString, futureRef, yamlFutureRef, -1) // -1 to swap all instances of [futureRef]
	}
	return swappedString
}

func (planYaml *PlanYaml) generateUuid() string {
	planYaml.latestUuid++
	return strconv.Itoa(planYaml.latestUuid)
}
