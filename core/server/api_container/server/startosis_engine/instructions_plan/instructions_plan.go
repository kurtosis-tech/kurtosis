package instructions_plan

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/dependency_graph"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/plan_yaml"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/types"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
)

// InstructionsPlan is the object to store a sequence of instructions which forms a "plan" for the enclave.
// Right now, the object is fairly simple in the sense of it just stores literally the sequence of instructions, and
// a bit of metadata about each instruction (i.e. whether it has been executed of not, for example)
// The plan is "append-only", i.e. when an instruction is added, it cannot be removed.
// The only read method is GeneratePlan unwraps the plan into an actual list of instructions that can be submitted to
// the executor.
type InstructionsPlan struct {
	indexOfFirstInstruction int

	scheduledInstructionsIndex map[types.ScheduledInstructionUuid]*ScheduledInstruction

	instructionsSequence []types.ScheduledInstructionUuid

	// list of package names that this instructions plan relies on
	packageDependencies map[string]bool

	uuidGenerator types.ScheduledInstructionUuidGenerator
}

func NewInstructionsPlan() *InstructionsPlan {
	return &InstructionsPlan{
		indexOfFirstInstruction:    0,
		scheduledInstructionsIndex: map[types.ScheduledInstructionUuid]*ScheduledInstruction{},
		instructionsSequence:       []types.ScheduledInstructionUuid{},
		packageDependencies:        map[string]bool{},
		uuidGenerator:              types.NewScheduledInstructionUuidGenerator(),
	}
}

func NewInstructionsPlanForDependencyGraphTests() *InstructionsPlan {
	return &InstructionsPlan{
		indexOfFirstInstruction:    0,
		scheduledInstructionsIndex: map[types.ScheduledInstructionUuid]*ScheduledInstruction{},
		instructionsSequence:       []types.ScheduledInstructionUuid{},
		packageDependencies:        map[string]bool{},
		uuidGenerator:              types.NewScheduledInstructionUuidGeneratorForTests(),
	}
}

func (plan *InstructionsPlan) SetIndexOfFirstInstruction(indexOfFirstInstruction int) {
	plan.indexOfFirstInstruction = indexOfFirstInstruction
}

func (plan *InstructionsPlan) GetIndexOfFirstInstruction() int {
	return plan.indexOfFirstInstruction
}

func (plan *InstructionsPlan) AddInstruction(instruction kurtosis_instruction.KurtosisInstruction, returnedValue starlark.Value) error {
	generatedUuid, err := plan.uuidGenerator.GenerateUUIDString()
	if err != nil {
		return stacktrace.Propagate(err, "Unable to generate a random UUID for instruction '%s' to add it to the plan", instruction.String())
	}

	scheduledInstructionUuid := types.ScheduledInstructionUuid(generatedUuid)
	scheduledInstruction := NewScheduledInstruction(scheduledInstructionUuid, instruction, returnedValue)

	plan.scheduledInstructionsIndex[scheduledInstructionUuid] = scheduledInstruction
	plan.instructionsSequence = append(plan.instructionsSequence, scheduledInstructionUuid)
	return nil
}

func (plan *InstructionsPlan) AddScheduledInstruction(scheduledInstruction *ScheduledInstruction) *ScheduledInstruction {
	newScheduledInstructionUuid := scheduledInstruction.uuid
	newScheduledInstruction := NewScheduledInstruction(newScheduledInstructionUuid, scheduledInstruction.kurtosisInstruction, scheduledInstruction.returnedValue)
	newScheduledInstruction.Executed(scheduledInstruction.IsExecuted())

	plan.scheduledInstructionsIndex[newScheduledInstructionUuid] = newScheduledInstruction
	plan.instructionsSequence = append(plan.instructionsSequence, newScheduledInstructionUuid)
	return newScheduledInstruction
}

// GeneratePlan unwraps the plan into a list of instructions
func (plan *InstructionsPlan) GeneratePlan() ([]*ScheduledInstruction, *startosis_errors.InterpretationError) {
	var generatedPlan []*ScheduledInstruction
	for _, instructionUuid := range plan.instructionsSequence {
		instruction, found := plan.scheduledInstructionsIndex[instructionUuid]
		if !found {
			return nil, startosis_errors.NewInterpretationError("Unexpected error generating the Kurtosis Instructions plan. Instruction with UUID '%s' was scheduled but could not be found in Kurtosis instruction index", instructionUuid)
		}
		generatedPlan = append(generatedPlan, instruction)
	}
	return generatedPlan, nil
}

func (plan *InstructionsPlan) GenerateInstructionsDependencyGraph() (map[types.ScheduledInstructionUuid][]types.ScheduledInstructionUuid, *startosis_errors.InterpretationError) {
	instructionsDependencies := dependency_graph.NewInstructionDependencyGraph(plan.instructionsSequence)
	for _, instructionUuid := range plan.instructionsSequence {
		instruction, found := plan.scheduledInstructionsIndex[instructionUuid]
		if !found {
			return nil, startosis_errors.NewInterpretationError("An error occurred updating the Instructions Dependency Graph. Instruction with UUID '%s' was scheduled but could not be found in Kurtosis instruction index.", instructionUuid)
		}
		logrus.Infof("Updating dependency graph with instruction: %v", instruction.kurtosisInstruction.String())
		err := instruction.kurtosisInstruction.UpdateDependencyGraph(instructionUuid, instructionsDependencies)
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "An error occurred updating the dependency graph with instruction: %v.", instructionUuid)
		}
	}

	instructionsDependencies.OutputDependencyGraphVisualWithShortDescriptors("/tmp")

	return instructionsDependencies.GenerateDependencyGraph(), nil
}

// GenerateYaml takes in an existing planYaml (usually empty) and returns a yaml string containing the effects of the plan
func (plan *InstructionsPlan) GenerateYaml(planYaml *plan_yaml.PlanYamlGenerator) (string, error) {
	for _, instructionUuid := range plan.instructionsSequence {
		instruction, found := plan.scheduledInstructionsIndex[instructionUuid]
		if !found {
			return "", startosis_errors.NewInterpretationError("Unexpected error generating the Kurtosis Instructions plan. Instruction with UUID '%s' was scheduled but could not be found in Kurtosis instruction index", instructionUuid)
		}
		err := instruction.kurtosisInstruction.UpdatePlan(planYaml)
		if err != nil {
			return "", startosis_errors.WrapWithInterpretationError(err, "An error occurred updating the plan with instruction: %v.", instructionUuid)
		}
	}
	planYaml.AddPackageDependencies(plan.packageDependencies)
	return planYaml.GenerateYaml()
}

func (plan *InstructionsPlan) AddPackageDependency(packageDependency string) {
	plan.packageDependencies[packageDependency] = true
}

func (plan *InstructionsPlan) Size() int {
	return len(plan.instructionsSequence)
}
