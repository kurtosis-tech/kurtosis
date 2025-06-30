package dependency_graph

import "github.com/sirupsen/logrus"

// TODO: This mirror the ScheduledInstructionUuid in instructions_plan.go
// It should be merged into a single type by refactoring the instructions_plan package to avoid circular dependencies
type ScheduledInstructionUuid string

type InstructionsDependencyGraph struct {
	instructionsDependencies map[ScheduledInstructionUuid]map[ScheduledInstructionUuid]bool

	// the following data structures tracks artifacts produced by instructions and consumed by downstream instructions
	outputsToInstructionMap map[string]ScheduledInstructionUuid
}

func NewInstructionsDependencyGraph() *InstructionsDependencyGraph {
	return &InstructionsDependencyGraph{
		instructionsDependencies: map[ScheduledInstructionUuid]map[ScheduledInstructionUuid]bool{},
		outputsToInstructionMap:  map[string]ScheduledInstructionUuid{},
	}
}

func (graph *InstructionsDependencyGraph) StoreOutput(instruction ScheduledInstructionUuid, output string) {
	graph.addInstruction(instruction)
	logrus.Infof("Storing output %s for instruction %s", output, instruction)
	graph.outputsToInstructionMap[output] = instruction
}

func (graph *InstructionsDependencyGraph) DependsOnOutput(instruction ScheduledInstructionUuid, output string) {
	instructionThatProducedOutput, ok := graph.outputsToInstructionMap[output]
	if !ok {
		panic("smth went wrong")
	}
	graph.addDependency(instruction, instructionThatProducedOutput)
}

// if [instruction] depends on [dependency] then [dependency] is stored as a dependency of [instruction]
func (graph *InstructionsDependencyGraph) DependsOnInstruction(instruction ScheduledInstructionUuid, dependency ScheduledInstructionUuid) {
	graph.addDependency(instruction, dependency)
}

func (graph *InstructionsDependencyGraph) addDependency(instruction ScheduledInstructionUuid, dependency ScheduledInstructionUuid) {
	// idempotently add the instruction and the dependency to the graph
	graph.addInstruction(instruction)
	graph.addInstruction(dependency)
	graph.instructionsDependencies[instruction][dependency] = true
}

// if instruction is not in the graph yet, add it with an empty dependency list
func (graph *InstructionsDependencyGraph) addInstruction(instruction ScheduledInstructionUuid) {
	if _, ok := graph.instructionsDependencies[instruction]; !ok {
		graph.instructionsDependencies[instruction] = map[ScheduledInstructionUuid]bool{}
	}
}

func (graph *InstructionsDependencyGraph) GetDependencyGraph() map[ScheduledInstructionUuid][]ScheduledInstructionUuid {
	dependencyGraph := map[ScheduledInstructionUuid][]ScheduledInstructionUuid{}
	for instruction, dependencies := range graph.instructionsDependencies {
		dependencyGraph[instruction] = []ScheduledInstructionUuid{}
		for dependency := range dependencies {
			dependencyGraph[instruction] = append(dependencyGraph[instruction], dependency)
		}
	}
	return dependencyGraph
}
