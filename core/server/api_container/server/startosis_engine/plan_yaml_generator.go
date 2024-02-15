package startosis_engine

import "github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan"

const (
	TCP ApplicationProtocol = "TCP"
	UDP ApplicationProtocol = "UDP"

	SHELL  TaskType = "sh"
	PYTHON TaskType = "python"
)

// we need the package id and the args, the args need to be filled in
// how to represent dependencies within the yaml???
// say a service config refers to another files artifact
// the conversions are
// add_service -> use service config and returned info to create a ServiceObject
// remove_service -> remove that from the plan representation
// upload_files -> FilesArtifact
// render_template -> FilesArtifact
// run_sh -> Task but returns a files artifact so create that
// run_python -> Task but returns a files artifact so create that

// go through all the kurtosis builtins and figure out which ones we need to accommodate for
//
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
	//
	//
	//
	//
	GenerateYaml(plan instructions_plan.InstructionsPlan) ([]byte, error)
}

type PlanYamlGeneratorImpl struct {
	plan *instructions_plan.InstructionsPlan

	componentIndex map[string]bool

	services []*Service

	filesArtifacts []*FilesArtifact

	tasks []*Task
}

func NewPlanYamlGenerator(plan *instructions_plan.InstructionsPlan) *PlanYamlGeneratorImpl {
	return &PlanYamlGeneratorImpl{
		plan: plan,
	}
}

func (pyg *PlanYamlGeneratorImpl) GenerateYaml() ([]byte, error) {
	// first thing: get list of instructions
	// second thing: iterate through list of instructions
	// depending on type of instruction, update the plan
	// at the very end, convert the plan into a yaml
	return []byte{}, nil
}

type Service struct {
	uuid    string
	name    string
	image   string
	envVars []*EnvironmentVariable
	ports   []*Port
	files   []*FileMount
}

type FilesArtifact struct {
	uuid  string
	name  string
	files map[string]string
}

type EnvironmentVariable struct {
	key   string
	value string
}

type Port struct {
	portName          string
	portNum           uint16
	transportProtocol ApplicationProtocol
}

type ApplicationProtocol string

type FileMount struct {
	mountPath string

	filesArtifactUuid string
	filesArtifactName string
}

type Task struct {
	taskType   TaskType
	name       string
	command    string
	image      string
	envVar     []*EnvironmentVariable
	files      []*FileMount
	store      []string
	shouldWait bool
	wait       string
}

type TaskType string
