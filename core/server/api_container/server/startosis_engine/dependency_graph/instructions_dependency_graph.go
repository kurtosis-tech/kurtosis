package dependency_graph

// TODO: This mirror the ScheduledInstructionUuid in instructions_plan.go
// It should be merged into a single type by refactoring the instructions_plan package to avoid circular dependencies
type ScheduledInstructionUuid string

type InstructionsDependencyGraph struct {
	instructionsDependencies map[ScheduledInstructionUuid][]ScheduledInstructionUuid

	// the following data structures tracks artifacts produced by instructions and consumed by downstream instructions

	// maps the name of a files artifact to the instruction that produced it
	filesArtifactToIntructionMap map[string]ScheduledInstructionUuid

	// maps the name of a service to the add_service instruciton that created it
	serviceNamesToInstructionMap map[string]ScheduledInstructionUuid

	// maps a future reference to the instruction that created it
	// future references are output from instructions like exec, run_sh, run_python
	futureReferencesToInstructionMap map[string]ScheduledInstructionUuid
}

func NewInstructionsDependencyGraph() *InstructionsDependencyGraph {
	return &InstructionsDependencyGraph{
		instructionsDependencies:         map[ScheduledInstructionUuid][]ScheduledInstructionUuid{},
		filesArtifactToIntructionMap:     map[string]ScheduledInstructionUuid{},
		serviceNamesToInstructionMap:     map[string]ScheduledInstructionUuid{},
		futureReferencesToInstructionMap: map[string]ScheduledInstructionUuid{},
	}
}

func (graph *InstructionsDependencyGraph) StoreFutureReference(futureReference string, instruction ScheduledInstructionUuid) {
	graph.futureReferencesToInstructionMap[futureReference] = instruction
}

func (graph *InstructionsDependencyGraph) StoreService(serviceName string, instruction ScheduledInstructionUuid) {
	graph.serviceNamesToInstructionMap[serviceName] = instruction
}

func (graph *InstructionsDependencyGraph) StoreFileArtifact(fileArtifactName string, instruction ScheduledInstructionUuid) {
	graph.filesArtifactToIntructionMap[fileArtifactName] = instruction
}

func (graph *InstructionsDependencyGraph) AddDependency(instruction ScheduledInstructionUuid, dependency ScheduledInstructionUuid) {
	if _, ok := graph.instructionsDependencies[instruction]; !ok {
		graph.instructionsDependencies[instruction] = []ScheduledInstructionUuid{}
	}
	graph.instructionsDependencies[instruction] = append(graph.instructionsDependencies[instruction], dependency)
}

func (graph *InstructionsDependencyGraph) GetDependencyGraph() map[ScheduledInstructionUuid][]ScheduledInstructionUuid {
	return graph.instructionsDependencies
}
