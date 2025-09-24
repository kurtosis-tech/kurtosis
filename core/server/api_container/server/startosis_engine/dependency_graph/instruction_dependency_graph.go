package dependency_graph

import (
	"fmt"
	"slices"

	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers/magic_string_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/types"
	"github.com/sirupsen/logrus"
)

// InstructionDependencyGraph tracks dependencies between Kurtosis instructions.
//
// Dependencies can be:
// - **Implicit**: An instruction uses outputs (e.g. service info, file artifacts, runtime values) from a prior instruction.
// - **Explicit**: An instruction lists another in its `depends_on` field.
//
// The graph is built by iterating through plan instructions, each of which calls:
//   kurtosisInstruction.UpdateDependencyGraph(instructionUuid, dependencyGraph)
//
// Each instruction can:
// 1. `StoreOutput` — Register outputs it produces.
// 2. `DependsOnOutput` — Declare dependencies on outputs from earlier instructions.
// 3. `DependsOnInstruction` — Add explicitly declared dependencies (`depends_on`).
//
// TODO: Implement `depends_on` for all instructions.

// For example,
// TODO: add an example here
// What does the another developor have to know about storing outputs?
// there is a link between the output and the depends on format
// need to find a way to explain a) when something is an output (e.g.) give an exhaustive list: Files Artifacts, Service Information, Runtime Values
type InstructionDependencyGraph struct {
	instructionsDependencies map[types.ScheduledInstructionUuid]map[types.ScheduledInstructionUuid]bool

	// the following data structures tracks artifacts produced by instructions and consumed by downstream instructions
	outputsToInstructionMap map[string]types.ScheduledInstructionUuid

	instructionShortDescriptors map[types.ScheduledInstructionUuid]string

	instructionsSequence []types.ScheduledInstructionUuid
}

func NewInstructionDependencyGraph(instructionsSequence []types.ScheduledInstructionUuid) *InstructionDependencyGraph {
	return &InstructionDependencyGraph{
		instructionsDependencies:    map[types.ScheduledInstructionUuid]map[types.ScheduledInstructionUuid]bool{},
		outputsToInstructionMap:     map[string]types.ScheduledInstructionUuid{},
		instructionShortDescriptors: map[types.ScheduledInstructionUuid]string{},
		instructionsSequence:        instructionsSequence,
	}
}

func (graph *InstructionDependencyGraph) ProducesService(instruction types.ScheduledInstructionUuid, serviceName string) {
	graph.addInstruction(instruction)
	logrus.Infof("Storing service %s for instruction %s", serviceName, instruction)
	graph.outputsToInstructionMap[serviceName] = instruction
}

func (graph *InstructionDependencyGraph) ProducesFilesArtifact(instruction types.ScheduledInstructionUuid, filesArtifactName string) {
	graph.addInstruction(instruction)
	logrus.Infof("Storing service %s for instruction %s", filesArtifactName, instruction)
	graph.outputsToInstructionMap[filesArtifactName] = instruction
}

func (graph *InstructionDependencyGraph) ProducesRuntimeValue(instruction types.ScheduledInstructionUuid, runtimeValue string) {
	graph.addInstruction(instruction)
	logrus.Infof("Storing runtime value %s for instruction %s", runtimeValue, instruction)
	graph.outputsToInstructionMap[runtimeValue] = instruction
}

func (graph *InstructionDependencyGraph) ConsumesService(instruction types.ScheduledInstructionUuid, serviceName string) {
	logrus.Infof("Attempting to consume service %s for instruction %s", serviceName, instruction)
	instructionThatProducedService, ok := graph.outputsToInstructionMap[serviceName]
	if !ok {
		panic(fmt.Sprintf("No instruction found that output service %s.", serviceName))
	}
	graph.addDependency(instruction, instructionThatProducedService)
}

func (graph *InstructionDependencyGraph) ConsumesFilesArtifact(instruction types.ScheduledInstructionUuid, filesArtifactName string) {
	logrus.Infof("Attempting to consume files artifact %s for instruction %s", filesArtifactName, instruction)
	instructionThatProducedFilesArtifact, ok := graph.outputsToInstructionMap[filesArtifactName]
	if !ok {
		panic(fmt.Sprintf("No instruction found that output files artifact %s.", filesArtifactName))
	}
	graph.addDependency(instruction, instructionThatProducedFilesArtifact)
}

func (graph *InstructionDependencyGraph) ConsumesAnyRuntimeValuesInString(instruction types.ScheduledInstructionUuid, stringPotentiallyContainingRuntimeValues string) {
	if runtimeValues, ok := magic_string_helper.ContainsRuntimeValue(stringPotentiallyContainingRuntimeValues); ok {
		for _, runtimeValue := range runtimeValues {
			graph.consumesRuntimeValue(instruction, runtimeValue)
		}
	}
}

func (graph *InstructionDependencyGraph) ConsumesAnyRuntimeValuesInList(instruction types.ScheduledInstructionUuid, listPotentiallyContainingRuntimeValues []string) {
	for _, wordPotentiallyContainingRuntimeValues := range listPotentiallyContainingRuntimeValues {
		if runtimeValues, ok := magic_string_helper.ContainsRuntimeValue(wordPotentiallyContainingRuntimeValues); ok {
			for _, runtimeValue := range runtimeValues {
				graph.consumesRuntimeValue(instruction, runtimeValue)
			}
		}
	}
}

func (graph *InstructionDependencyGraph) consumesRuntimeValue(instruction types.ScheduledInstructionUuid, runtimeValue string) {
	logrus.Infof("Attempting to consume runtime value %s for instruction %s", runtimeValue, instruction)
	instructionThatProducedRuntimeValue, ok := graph.outputsToInstructionMap[runtimeValue]
	if !ok {
		panic(fmt.Sprintf("No instruction found that output runtime value %s.", runtimeValue))
	}
	graph.addDependency(instruction, instructionThatProducedRuntimeValue)
}

// AddPrintInstruction manually adds a dependency between a print and the instruction that comes right before it in the instructions sequence.
// TODO: explain why this is important
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
	// idempotently add the instruction and the dependency to the graph
	if instruction == dependency {
		return
	}
	graph.addInstruction(instruction)
	graph.addInstruction(dependency)
	graph.instructionsDependencies[instruction][dependency] = true
}

// if instruction is not in the graph yet, add it with an empty dependency list
func (graph *InstructionDependencyGraph) addInstruction(instruction types.ScheduledInstructionUuid) {
	if _, ok := graph.instructionsDependencies[instruction]; !ok {
		graph.instructionsDependencies[instruction] = map[types.ScheduledInstructionUuid]bool{}
	}
}

func (graph *InstructionDependencyGraph) GetDependencyGraph() map[types.ScheduledInstructionUuid][]types.ScheduledInstructionUuid {
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
