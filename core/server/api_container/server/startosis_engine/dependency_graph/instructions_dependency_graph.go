package dependency_graph

import (
	"os"
	"os/exec"
	"slices"
	"strconv"

	"github.com/sirupsen/logrus"
	"gonum.org/v1/gonum/graph/encoding/dot"
	"gonum.org/v1/gonum/graph/simple"
)

// TODO: This mirrors the ScheduledInstructionUuid in instructions_plan.go
// It should be merged into a single type by refactoring the instructions_plan package to avoid circular dependencies
type ScheduledInstructionUuid string

// InstructionsDependencyGraph tracks dependencies between Kurtosis instructions.
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
	logrus.Infof("Attempting to depend instruction %s on output %s", instruction, output)
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
		instructionDependencies := []ScheduledInstructionUuid{}
		for dependency := range dependencies {
			instructionDependencies = append(instructionDependencies, dependency)
		}
		slices.Sort(instructionDependencies)
		dependencyGraph[instruction] = instructionDependencies
	}
	return dependencyGraph
}

func OutputDependencyGraphVisual(dependencyGraph map[ScheduledInstructionUuid][]ScheduledInstructionUuid) {
	g := simple.NewDirectedGraph()

	nodes := make(map[string]int64)

	for to, fromList := range dependencyGraph {
		if _, ok := nodes[string(to)]; !ok {
			nextID, err := strconv.ParseInt(string(to), 10, 64)
			if err != nil {
				panic(err)
			}
			nodes[string(to)] = nextID
			g.AddNode(simple.Node(nextID))
		}
		for _, from := range fromList {
			if _, ok := nodes[string(from)]; !ok {
				nextID, err := strconv.ParseInt(string(from), 10, 64)
				if err != nil {
					panic(err)
				}
				nodes[string(from)] = nextID
				g.AddNode(simple.Node(nextID))
			}
			g.SetEdge(g.NewEdge(simple.Node(nodes[string(to)]), simple.Node(nodes[string(from)])))
		}
	}

	b, err := dot.Marshal(g, "InstructionsDependencyGraph", "", "  ")
	if err != nil {
		panic(err)
	}

	// Write to file
	if err := os.WriteFile("/Users/tewodrosmitiku/craft/graphs/dependency.dot", b, 0644); err != nil {
		panic(err)
	}

	cmd := exec.Command("dot", "-Tpng", "/Users/tewodrosmitiku/craft/graphs/dependency.dot", "-o", "/Users/tewodrosmitiku/craft/graphs/graph.png")
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}
