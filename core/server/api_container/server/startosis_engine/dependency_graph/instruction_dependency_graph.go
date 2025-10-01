package dependency_graph

import (
	"slices"

	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers/magic_string_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/types"
)

// InstructionDependencyGraph tracks dependencies between Starlark instructions in an instruction sequence.
//
// Currently, dependencies between instructions can be established in three ways:
// 1. Services
// 2. Files Artifacts
// 3. Runtime Values (any Starlark future reference output by instructions such as `exec`, `request`, `run_python`, `run_sh`, etc.)
//
// Instructions define their dependencies in kurtosisInstruction.UpdateDependencyGraph(instructionUuid, dependencyGraph) by informing the dependency graph which of the above they produce and which they consume.
//
// See instruction_dependency_graph_test.go for examples on how Starlark scripts/instruction sequences are represented as an InstructionDependencyGraph.
type InstructionDependencyGraph struct {
	instructionsSequence []types.ScheduledInstructionUuid

	instructionsDependencies map[types.ScheduledInstructionUuid]map[types.ScheduledInstructionUuid]bool

	// Right now, Services, Files Artifacts, and Runtime Values are all represented as strings for simplicity
	// In the future, we may add types to represent each output but not needed for now
	outputsToInstructionMap map[string]types.ScheduledInstructionUuid

	shortDescriptors map[types.ScheduledInstructionUuid]string
}

func NewInstructionDependencyGraph(instructionsSequence []types.ScheduledInstructionUuid) *InstructionDependencyGraph {
	instructionsDependencies := make(map[types.ScheduledInstructionUuid]map[types.ScheduledInstructionUuid]bool)
	for _, instruction := range instructionsSequence {
		instructionsDependencies[instruction] = make(map[types.ScheduledInstructionUuid]bool)
	}

	return &InstructionDependencyGraph{
		instructionsDependencies: instructionsDependencies,
		outputsToInstructionMap:  map[string]types.ScheduledInstructionUuid{},
		instructionsSequence:     instructionsSequence,
		shortDescriptors:         map[types.ScheduledInstructionUuid]string{},
	}
}

func (graph *InstructionDependencyGraph) ProducesService(instruction types.ScheduledInstructionUuid, serviceName string) {
	graph.outputsToInstructionMap[serviceName] = instruction
}

func (graph *InstructionDependencyGraph) ProducesFilesArtifact(instruction types.ScheduledInstructionUuid, filesArtifactName string) {
	graph.outputsToInstructionMap[filesArtifactName] = instruction
}

func (graph *InstructionDependencyGraph) ProducesRuntimeValue(instruction types.ScheduledInstructionUuid, runtimeValue string) {
	graph.outputsToInstructionMap[runtimeValue] = instruction
}

func (graph *InstructionDependencyGraph) ConsumesService(instruction types.ScheduledInstructionUuid, serviceName string) {
	instructionThatProducedService, ok := graph.outputsToInstructionMap[serviceName]
	if ok {
		graph.addDependency(instruction, instructionThatProducedService)
	}
}

func (graph *InstructionDependencyGraph) ConsumesFilesArtifact(instruction types.ScheduledInstructionUuid, filesArtifactName string) {
	instructionThatProducedFilesArtifact, ok := graph.outputsToInstructionMap[filesArtifactName]
	if ok {
		graph.addDependency(instruction, instructionThatProducedFilesArtifact)
	}
}

func (graph *InstructionDependencyGraph) ConsumesAnyRuntimeValuesInString(instruction types.ScheduledInstructionUuid, stringPotentiallyContainingRuntimeValues string) {
	for _, runtimeValue := range magic_string_helper.GetRuntimeValuesFromString(stringPotentiallyContainingRuntimeValues) {
		graph.consumesRuntimeValue(instruction, runtimeValue)
	}
}

func (graph *InstructionDependencyGraph) ConsumesAnyRuntimeValuesInList(instruction types.ScheduledInstructionUuid, listPotentiallyContainingRuntimeValues []string) {
	for _, wordPotentiallyContainingRuntimeValues := range listPotentiallyContainingRuntimeValues {
		for _, runtimeValue := range magic_string_helper.GetRuntimeValuesFromString(wordPotentiallyContainingRuntimeValues) {
			graph.consumesRuntimeValue(instruction, runtimeValue)
		}
	}
}

func (graph *InstructionDependencyGraph) consumesRuntimeValue(instruction types.ScheduledInstructionUuid, runtimeValue string) {
	instructionThatProducedRuntimeValue, ok := graph.outputsToInstructionMap[runtimeValue]
	if ok {
		graph.addDependency(instruction, instructionThatProducedRuntimeValue)
	}
}

// AddPrintInstruction manually adds a dependency between a print and the instruction that comes right before it in the instructions sequence.
func (graph *InstructionDependencyGraph) AddPrintInstruction(instruction types.ScheduledInstructionUuid) {
	for i := 1; i < len(graph.instructionsSequence); i++ {
		if graph.instructionsSequence[i] == instruction {
			dependency := graph.instructionsSequence[i-1]
			graph.addDependency(instruction, dependency)
			return
		}
	}
}

func (graph *InstructionDependencyGraph) addDependency(instruction types.ScheduledInstructionUuid, dependency types.ScheduledInstructionUuid) {
	if instruction == dependency {
		return
	}
	graph.instructionsDependencies[instruction][dependency] = true
}

func (graph *InstructionDependencyGraph) GenerateDependencyGraph() map[types.ScheduledInstructionUuid][]types.ScheduledInstructionUuid {
	dependencyGraph := map[types.ScheduledInstructionUuid][]types.ScheduledInstructionUuid{}
	for instruction, dependencies := range graph.instructionsDependencies {
		instructionDependencies := []types.ScheduledInstructionUuid{}
		for dependency := range dependencies {
			instructionDependencies = append(instructionDependencies, dependency)
		}
		slices.Sort(instructionDependencies)
		dependencyGraph[instruction] = instructionDependencies
	}
	return dependencyGraph
}

func (graph *InstructionDependencyGraph) getInvertedDependencyGraph() map[types.ScheduledInstructionUuid][]types.ScheduledInstructionUuid {
	invertedDependencyGraph := map[types.ScheduledInstructionUuid][]types.ScheduledInstructionUuid{}
	for instruction := range graph.instructionsDependencies {
		invertedDependencyGraph[instruction] = []types.ScheduledInstructionUuid{}
	}
	for instruction, dependencies := range graph.instructionsDependencies {
		for dependency := range dependencies {
			invertedDependencyGraph[dependency] = append(invertedDependencyGraph[dependency], instruction)
		}
		slices.Sort(invertedDependencyGraph[instruction])
	}
	return invertedDependencyGraph
}
