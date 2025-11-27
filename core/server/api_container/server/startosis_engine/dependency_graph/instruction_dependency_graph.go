package dependency_graph

import (
	"slices"

	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers/magic_string_helper"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/types"
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

	instructionShortDescriptors map[types.ScheduledInstructionUuid]string
	printInstructionUuids       map[types.ScheduledInstructionUuid]bool
	waitInstructionUuids        []types.ScheduledInstructionUuid

	// Right now, Services, Files Artifacts, and Runtime Values are all represented as strings for simplicity
	// In the future, we may add types to represent each output but not needed for now
	outputsToInstructionUuids map[string]types.ScheduledInstructionUuid
}

type InstructionWithDependencies struct {
	InstructionUuid    types.ScheduledInstructionUuid   `yaml:"instructionUuid"`
	ShortDescriptor    string                           `yaml:"shortDescriptor"`
	Dependencies       []types.ScheduledInstructionUuid `yaml:"dependencies"`
	IsPrintInstruction bool                             `yaml:"isPrintInstruction"`
}

func NewInstructionDependencyGraph(instructionsSequence []types.ScheduledInstructionUuid) *InstructionDependencyGraph {
	instructionsDependencies := make(map[types.ScheduledInstructionUuid]map[types.ScheduledInstructionUuid]bool)
	for _, instruction := range instructionsSequence {
		instructionsDependencies[instruction] = make(map[types.ScheduledInstructionUuid]bool)
	}

	printInstructionUuids := make(map[types.ScheduledInstructionUuid]bool)
	for _, instruction := range instructionsSequence {
		printInstructionUuids[instruction] = false
	}

	instructionShortDescriptors := make(map[types.ScheduledInstructionUuid]string)
	for _, instruction := range instructionsSequence {
		instructionShortDescriptors[instruction] = string(instruction) // initialize instruction descriptors to their uuid, but will be updated later
	}

	return &InstructionDependencyGraph{
		instructionsDependencies:    instructionsDependencies,
		outputsToInstructionUuids:   map[string]types.ScheduledInstructionUuid{},
		instructionsSequence:        instructionsSequence,
		instructionShortDescriptors: instructionShortDescriptors,
		printInstructionUuids:       printInstructionUuids,
		waitInstructionUuids:        []types.ScheduledInstructionUuid{},
	}
}

func (graph *InstructionDependencyGraph) ProducesService(instruction types.ScheduledInstructionUuid, serviceName string) {
	graph.outputsToInstructionUuids[serviceName] = instruction
}

func (graph *InstructionDependencyGraph) ProducesFilesArtifact(instruction types.ScheduledInstructionUuid, filesArtifactName string) {
	graph.outputsToInstructionUuids[filesArtifactName] = instruction
}

func (graph *InstructionDependencyGraph) ProducesRuntimeValue(instruction types.ScheduledInstructionUuid, runtimeValue string) {
	graph.outputsToInstructionUuids[runtimeValue] = instruction
}

func (graph *InstructionDependencyGraph) ConsumesService(instruction types.ScheduledInstructionUuid, serviceName string) {
	instructionThatProducedService, ok := graph.outputsToInstructionUuids[serviceName]
	if ok {
		graph.addDependency(instruction, instructionThatProducedService)
	}
}

func (graph *InstructionDependencyGraph) ConsumesFilesArtifact(instruction types.ScheduledInstructionUuid, filesArtifactName string) {
	instructionThatProducedFilesArtifact, ok := graph.outputsToInstructionUuids[filesArtifactName]
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
	instructionThatProducedRuntimeValue, ok := graph.outputsToInstructionUuids[runtimeValue]
	if ok {
		graph.addDependency(instruction, instructionThatProducedRuntimeValue)
	}
}

// AddPrintInstruction manually adds a dependency between a print and the instruction that comes right before it in the instructions sequence.
func (graph *InstructionDependencyGraph) AddPrintInstruction(instruction types.ScheduledInstructionUuid) {
	graph.printInstructionUuids[instruction] = true
	for i := 1; i < len(graph.instructionsSequence); i++ {
		if graph.instructionsSequence[i] == instruction {
			dependency := graph.instructionsSequence[i-1]
			graph.addDependency(instruction, dependency)
			return
		}
	}
}

// AddWaitInstruction tracks a wait instruction in the dependency graph.
// All instructions sequentially downstream from a wait instruction must depend on the wait instruction
func (graph *InstructionDependencyGraph) AddWaitInstruction(instruction types.ScheduledInstructionUuid) {
	graph.waitInstructionUuids = append(graph.waitInstructionUuids, instruction)
}

func (graph *InstructionDependencyGraph) ensureInstructionDependsOnWaitInstructions(instruction types.ScheduledInstructionUuid) {
	for _, waitInstruction := range graph.waitInstructionUuids {
		if _, ok := graph.instructionsDependencies[instruction][waitInstruction]; !ok {
			if instruction == waitInstruction {
				continue
			}
			graph.instructionsDependencies[instruction][waitInstruction] = true
		}
	}
}

// AddRemoveServiceInstruction tracks a remove service instruction in the dependency graph.
// remove_service instructions must come after all operations on the service have been completed
func (graph *InstructionDependencyGraph) AddRemoveServiceInstruction(instruction types.ScheduledInstructionUuid, serviceName string) {
	instructionThatProducedService, ok := graph.outputsToInstructionUuids[serviceName]
	if !ok {
		// it could be the case the remove_service exists in a script that doesn't contain the add_service instruction that produces the service
		// in which case the service won't exist in the outputs map so we just skip adding dependencies
		return
	}
	for maybeInstructionDependingOnService, dependencies := range graph.instructionsDependencies {
		// if a instruction depends on the service that the remove_service instruction is removing, add a dependency between the remove_service instruction and the instruction
		if _, ok := dependencies[instructionThatProducedService]; ok {
			graph.addDependency(instruction, maybeInstructionDependingOnService)
		}
	}
}

func (graph *InstructionDependencyGraph) UpdateInstructionShortDescriptor(instruction types.ScheduledInstructionUuid, shortDescriptor string) {
	graph.instructionShortDescriptors[instruction] = shortDescriptor
}

func (graph *InstructionDependencyGraph) addDependency(instruction types.ScheduledInstructionUuid, dependency types.ScheduledInstructionUuid) {
	if instruction == dependency {
		return
	}
	graph.instructionsDependencies[instruction][dependency] = true

	graph.ensureInstructionDependsOnWaitInstructions(instruction)
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

func (graph *InstructionDependencyGraph) GenerateInstructionsWithDependencies() []InstructionWithDependencies {
	instructionsWithDependencies := []InstructionWithDependencies{}
	for _, instruction := range graph.instructionsSequence {
		dependencies := []types.ScheduledInstructionUuid{}
		for dependency := range graph.instructionsDependencies[instruction] {
			dependencies = append(dependencies, dependency)
		}
		instructionsWithDependencies = append(instructionsWithDependencies, InstructionWithDependencies{
			InstructionUuid:    instruction,
			ShortDescriptor:    graph.instructionShortDescriptors[instruction],
			IsPrintInstruction: graph.printInstructionUuids[instruction],
			Dependencies:       dependencies,
		})
	}
	return instructionsWithDependencies
}
