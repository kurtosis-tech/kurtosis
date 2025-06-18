package dependency_graph

// TODO: This mirror the ScheduledInstructionUuid in instructions_plan.go
// It should be merged into a single type by refactoring the instructions_plan package to avoid circular dependencies
type ScheduledInstructionUuid string

type InstructionsDependencyGraph struct {
	instructionsDependencies map[ScheduledInstructionUuid][]ScheduledInstructionUuid

	// the following data structures tracks artifacts produced by instructions and consumed by downstream instructions
	outputsToInstructionMap map[string]ScheduledInstructionUuid

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
		outputsToInstructionMap:          map[string]ScheduledInstructionUuid{},
		filesArtifactToIntructionMap:     map[string]ScheduledInstructionUuid{},
		serviceNamesToInstructionMap:     map[string]ScheduledInstructionUuid{},
		futureReferencesToInstructionMap: map[string]ScheduledInstructionUuid{},
	}
}

func (graph *InstructionsDependencyGraph) StoreOutput(instruction ScheduledInstructionUuid, output string) {
	graph.addInstruction(instruction)
	graph.outputsToInstructionMap[output] = instruction
}

func (graph *InstructionsDependencyGraph) StoreFutureReference(futureReference string, instruction ScheduledInstructionUuid) {
	graph.addInstruction(instruction)
	graph.futureReferencesToInstructionMap[futureReference] = instruction
}

func (graph *InstructionsDependencyGraph) StoreServiceName(serviceName string, instruction ScheduledInstructionUuid) {
	graph.addInstruction(instruction)
	graph.serviceNamesToInstructionMap[serviceName] = instruction
}

func (graph *InstructionsDependencyGraph) StoreFileArtifact(filesArtifactName string, instruction ScheduledInstructionUuid) {
	graph.addInstruction(instruction)
	graph.filesArtifactToIntructionMap[filesArtifactName] = instruction
}

func (graph *InstructionsDependencyGraph) DependsOnOutput(instruction ScheduledInstructionUuid, output string) {
	instructionThatOutputOutput, ok := graph.outputsToInstructionMap[output]
	if !ok {
		panic("smth went wrong")
	}
	graph.addDependency(instruction, instructionThatOutputOutput)
}

// if [instruction] depends on [filesArtifactName] then the instruction that outputs [fileArtifactName] is stored as a dependency of [instruction]
func (graph *InstructionsDependencyGraph) DependsOnFilesArtifact(instruction ScheduledInstructionUuid, filesArtifactName string) {
	instructionThatOutputFileArtifact, ok := graph.filesArtifactToIntructionMap[filesArtifactName]
	if !ok {
		panic("smth went wrong")
	}
	graph.addDependency(instruction, instructionThatOutputFileArtifact)
}

// if [instruction] depends on [futureReference] then the instruction that outputs [futureReference] is stored as a dependency of [instruction]
func (graph *InstructionsDependencyGraph) DependsOnFutureReference(instruction ScheduledInstructionUuid, futureReference string) {
	instructionThatOutputFutureReference, ok := graph.futureReferencesToInstructionMap[futureReference]
	if !ok {
		panic("smth went wrong")
	}
	graph.addDependency(instruction, instructionThatOutputFutureReference)
}

// if [instruction] depends on [serviceName] then the instruction that outputs [serviceName] is stored as a dependency of [instruction]
func (graph *InstructionsDependencyGraph) DependsOnService(instruction ScheduledInstructionUuid, serviceName string) {
	instructionThatOutputService, ok := graph.serviceNamesToInstructionMap[serviceName]
	if !ok {
		panic("smth went wrong")
	}
	graph.addDependency(instruction, instructionThatOutputService)
}

// if [instruction] depends on [dependency] then [dependency] is stored as a dependency of [instruction]
func (graph *InstructionsDependencyGraph) DependsOnInstruction(instruction ScheduledInstructionUuid, dependency ScheduledInstructionUuid) {
	graph.addDependency(instruction, dependency)
}

func (graph *InstructionsDependencyGraph) addDependency(instruction ScheduledInstructionUuid, dependency ScheduledInstructionUuid) {
	graph.addInstruction(instruction)
	graph.addInstruction(dependency)
	graph.instructionsDependencies[instruction] = append(graph.instructionsDependencies[instruction], dependency)
}

func (graph *InstructionsDependencyGraph) addInstruction(instruction ScheduledInstructionUuid) {
	if _, ok := graph.instructionsDependencies[instruction]; !ok {
		graph.instructionsDependencies[instruction] = []ScheduledInstructionUuid{}
	}
}

func (graph *InstructionsDependencyGraph) GetDependencyGraph() map[ScheduledInstructionUuid][]ScheduledInstructionUuid {
	return graph.instructionsDependencies
}
