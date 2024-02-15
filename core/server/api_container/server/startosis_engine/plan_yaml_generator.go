package startosis_engine

import (
	"github.com/go-yaml/yaml"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/add_service"
)

const (
	TCP ApplicationProtocol = "TCP"
	UDP ApplicationProtocol = "UDP"

	SHELL  TaskType = "sh"
	PYTHON TaskType = "python"
)

// We need the package id and the args, the args need to be filled in
// the instructions likely come with the args filled in already, but what if no args are passed in? are they left as variables?

// How to represent dependencies within the yaml???
// say a service config refers to another files artifact

// some conversions are:
// add_service -> use service config and returned info to create a ServiceObject
// remove_service -> remove that from the plan representation
// upload_files -> FilesArtifact
// render_template -> FilesArtifact
// run_sh -> Task but returns a files artifact so create that
// run_python -> Task but returns a files artifact so create that
//
// go through all the kurtosis builtins and figure out which ones we need to accommodate for and which ones we don't need to accomodate for

// PlanYamlGenerator generates a yaml representation of a [plan].
type PlanYamlGenerator interface {
	// GenerateYaml converts [plan] into a byte array that represents a yaml with information in the plan.
	// The format of the yaml in the byte array is as such:
	//
	//
	//
	// packageId: github.com/kurtosis-tech/postgres-package
	//
	// services:
	// 	- uuid:
	//	- name:
	//     service_config:
	//			image:
	//			env_var:
	//			...
	//
	//
	// files_artifacts:
	//
	//
	//
	//
	//
	//
	// tasks:
	//
	//

	GenerateYaml(plan instructions_plan.InstructionsPlan) ([]byte, error)
}

type PlanYamlGeneratorImpl struct {
	// Plan generetated by an interpretation of a starlark script of package
	plan *instructions_plan.InstructionsPlan

	// Representation of plan in yaml the plan is being processed, the yaml gets updated
	planYaml *PlanYaml
}

func NewPlanYamlGenerator(plan *instructions_plan.InstructionsPlan) *PlanYamlGeneratorImpl {
	return &PlanYamlGeneratorImpl{
		plan:     plan,
		planYaml: &PlanYaml{},
	}
}

func (pyg *PlanYamlGeneratorImpl) GenerateYaml() ([]byte, error) {
	instructionsSequence, err := pyg.plan.GeneratePlan()
	if err != nil {
		return nil, err
	}

	// iterate over the sequence of instructions
	for _, scheduledInstruction := range instructionsSequence {
		// based on the instruction, update the plan yaml representation accordingly
		switch getBuiltinNameFromInstruction(scheduledInstruction) {
		case add_service.AddServiceBuiltinName:
			pyg.updatePlanYamlFromAddService(scheduledInstruction)
		default:
			return nil, nil
		}
	}

	// at the very end, convert the plan yaml representation into a yaml
	return convertPlanYamlToYaml(pyg.planYaml)
}

func convertPlanYamlToYaml(planYaml *PlanYaml) ([]byte, error) {
	yamlBytes, err := yaml.Marshal(planYaml)
	if err != nil {
		return []byte{}, err
	}
	return yamlBytes, nil
}

func getBuiltinNameFromInstruction(instruction *instructions_plan.ScheduledInstruction) string {
	return instruction.GetInstruction().GetCanonicalInstruction(false).GetInstructionName()
}

func (pyg *PlanYamlGeneratorImpl) updatePlanYamlFromAddService(addServiceInstruction *instructions_plan.ScheduledInstruction) {

}

type PlanYaml struct {
	PackageId      string           `yaml:"packageId,omitempty"`
	Services       []*Service       `yaml:"services,omitempty"`
	FilesArtifacts []*FilesArtifact `yaml:"filesArtifacts,omitempty"`
	Tasks          []*Task          `yaml:"tasks,omitempty"`
}

// Service represents a service in the system.
type Service struct {
	Uuid    string                 `yaml:"uuid,omitempty"`
	Name    string                 `yaml:"name,omitempty"`
	Image   string                 `yaml:"image,omitempty"`
	EnvVars []*EnvironmentVariable `yaml:"envVars,omitempty"`
	Ports   []*Port                `yaml:"ports,omitempty"`
	Files   []*FileMount           `yaml:"files,omitempty"`
}

// FilesArtifact represents a collection of files.
type FilesArtifact struct {
	Uuid  string            `yaml:"uuid,omitempty"`
	Name  string            `yaml:"name,omitempty"`
	Files map[string]string `yaml:"files,omitempty"`
}

// EnvironmentVariable represents an environment variable.
type EnvironmentVariable struct {
	Key   string `yaml:"key,omitempty"`
	Value string `yaml:"value,omitempty"`
}

// Port represents a port.
type Port struct {
	TransportProtocol ApplicationProtocol `yaml:"transportProtocol,omitempty"`

	PortName string `yaml:"portName,omitempty"`
	PortNum  uint16 `yaml:"portNum,omitempty"`
}

// ApplicationProtocol represents the application protocol used.
type ApplicationProtocol string

// FileMount represents a mount point for files.
type FileMount struct {
	MountPath         string `yaml:"mountPath,omitempty"`
	FilesArtifactUuid string `yaml:"filesArtifactUuid,omitempty"`
	FilesArtifactName string `yaml:"filesArtifactName,omitempty"`
}

// Task represents a task to be executed.
type Task struct {
	TaskType   TaskType               `yaml:"taskType,omitempty"`
	Name       string                 `yaml:"name,omitempty"`
	Command    string                 `yaml:"command,omitempty"`
	Image      string                 `yaml:"image,omitempty"`
	EnvVars    []*EnvironmentVariable `yaml:"envVar,omitempty"`
	Files      []*FileMount           `yaml:"files,omitempty"`
	Store      []string               `yaml:"store,omitempty"`
	ShouldWait bool                   `yaml:"shouldWait,omitempty"`
	Wait       string                 `yaml:"wait,omitempty"`
}

// TaskType represents the type of task.
type TaskType string
