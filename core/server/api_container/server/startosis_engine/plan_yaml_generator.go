package startosis_engine

import (
	"github.com/go-yaml/yaml"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/add_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/remove_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/tasks"
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
		case remove_service.RemoveServiceBuiltinName:
			pyg.updatePlanYamlFromRemoveService(scheduledInstruction)
		case tasks.RunShBuiltinName:
			pyg.updatePlanYamlFromRunSh(scheduledInstruction)
		case tasks.RunPythonBuiltinName:
			pyg.updatePlanYamlFromRunPython(scheduledInstruction)
		default:
			return nil, nil
		}
	}

	// at the very end, convert the plan yaml representation into a yaml
	return convertPlanYamlToYaml(pyg.planYaml)
}

// is there anyway i can make this type more specific for type safety
func (pyg *PlanYamlGeneratorImpl) updatePlanYamlFromAddService(addServiceInstruction *instructions_plan.ScheduledInstruction) {
	// TODO: update the plan yaml based on an add_service
}

func (pyg *PlanYamlGeneratorImpl) updatePlanYamlFromRemoveService(RemoveServiceInstruction *instructions_plan.ScheduledInstruction) {
	// TODO: update the plan yaml based on an add_service
}

func (pyg *PlanYamlGeneratorImpl) updatePlanYamlFromRunSh(addServiceInstruction *instructions_plan.ScheduledInstruction) {
	// TODO: update the plan yaml based on an add_service
}

func (pyg *PlanYamlGeneratorImpl) updatePlanYamlFromRunPython(addServiceInstruction *instructions_plan.ScheduledInstruction) {
	// TODO: update the plan yaml based on an add_service
}

func (pyg *PlanYamlGeneratorImpl) updatePlanYamlFromUploadFiles(addServiceInstruction *instructions_plan.ScheduledInstruction) {
	// TODO: update the plan yaml based on an add_service
}

func (pyg *PlanYamlGeneratorImpl) updatePlanYamlFromRenderTemplates(addServiceInstruction *instructions_plan.ScheduledInstruction) {
	// TODO: update the plan yaml based on an add_service
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
