package dependency_graph

import (
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"

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

	instructionShortDescriptors map[ScheduledInstructionUuid]string
}

func NewInstructionsDependencyGraph() *InstructionsDependencyGraph {
	return &InstructionsDependencyGraph{
		instructionsDependencies:    map[ScheduledInstructionUuid]map[ScheduledInstructionUuid]bool{},
		outputsToInstructionMap:     map[string]ScheduledInstructionUuid{},
		instructionShortDescriptors: map[ScheduledInstructionUuid]string{},
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

func (graph *InstructionsDependencyGraph) getInvertedDependencyGraph() map[ScheduledInstructionUuid][]ScheduledInstructionUuid {
	invertedDependencyGraph := map[ScheduledInstructionUuid][]ScheduledInstructionUuid{}
	for instruction := range graph.instructionsDependencies {
		invertedDependencyGraph[instruction] = []ScheduledInstructionUuid{}
	}
	for instruction, dependencies := range graph.instructionsDependencies {
		for dependency := range dependencies {
			invertedDependencyGraph[dependency] = append(invertedDependencyGraph[dependency], instruction)
		}
		slices.Sort(invertedDependencyGraph[instruction])
	}
	return invertedDependencyGraph
}

// if instruction is not in the graph yet, add it with an empty dependency list
func (graph *InstructionsDependencyGraph) addInstruction(instruction ScheduledInstructionUuid) {
	if _, ok := graph.instructionsDependencies[instruction]; !ok {
		graph.instructionsDependencies[instruction] = map[ScheduledInstructionUuid]bool{}
	}
}

// AddInstructionShortDescriptor adds a human-readable description for an instruction.
// This description will be used as the node label in the visual graph.
// Example: AddInstructionShortDescriptor("instruction-1", "add_service(database)")
func (graph *InstructionsDependencyGraph) AddInstructionShortDescriptor(instruction ScheduledInstructionUuid, shortDescriptor string) {
	graph.instructionShortDescriptors[instruction] = shortDescriptor
}

// GetInstructionShortDescriptor returns the short descriptor for an instruction.
// Returns the instruction UUID if no descriptor is set.
func (graph *InstructionsDependencyGraph) GetInstructionShortDescriptor(instruction ScheduledInstructionUuid) string {
	if descriptor, exists := graph.instructionShortDescriptors[instruction]; exists {
		return descriptor
	}
	return string(instruction)
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

func (graph *InstructionsDependencyGraph) GetInstructionShortDescriptors() map[ScheduledInstructionUuid]string {
	return graph.instructionShortDescriptors
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

func (graph *InstructionsDependencyGraph) OutputDependencyGraphVisualWithShortDescriptors(path string) {
	g := simple.NewDirectedGraph()

	// Create a mapping from instruction UUID to node ID
	instructionToNodeID := make(map[ScheduledInstructionUuid]int64)
	nodeIDToDescriptor := make(map[int64]string)

	// First pass: create all nodes with their descriptors
	dependencyGraph := graph.getInvertedDependencyGraph()
	logrus.Infof("dependencyGraph: %v", dependencyGraph)
	shortDescriptors := graph.GetInstructionShortDescriptors()

	// Add all instructions as nodes
	for instruction := range dependencyGraph {
		nextNodeID, err := strconv.ParseInt(string(instruction), 10, 64)
		if err != nil {
			panic(err)
		}
		instructionToNodeID[instruction] = nextNodeID

		// Get descriptor, fallback to instruction UUID if not found
		descriptor, exists := shortDescriptors[instruction]
		if !exists {
			descriptor = string(instruction)
		}
		nodeIDToDescriptor[nextNodeID] = descriptor

		g.AddNode(simple.Node(nextNodeID))
	}
	logrus.Infof("nodeIDToDescriptor: %v", nodeIDToDescriptor)

	// Second pass: add edges
	for to, fromList := range dependencyGraph {
		toNodeID := instructionToNodeID[to]
		for _, from := range fromList {
			fromNodeID := instructionToNodeID[from]
			g.SetEdge(g.NewEdge(simple.Node(toNodeID), simple.Node(fromNodeID)))
		}
	}

	// Create DOT representation with node labels
	dotContent := "digraph InstructionsDependencyGraph {\n"
	dotContent += "  rankdir=TB;\n"
	dotContent += "  node [shape=box, style=filled, fillcolor=lightblue];\n\n"

	// Add nodes with descriptions
	for nodeID, descriptor := range nodeIDToDescriptor {
		// Escape quotes in descriptor for DOT format
		escapedDescriptor := strings.ReplaceAll(descriptor, "\"", "\\\"")
		dotContent += fmt.Sprintf("  %d [label=\"%d - %s\"];\n", nodeID, nodeID, escapedDescriptor)
	}

	dotContent += "\n"

	// Add edges
	for to, fromList := range dependencyGraph {
		toNodeID := instructionToNodeID[to]
		for _, from := range fromList {
			fromNodeID := instructionToNodeID[from]
			dotContent += fmt.Sprintf("  %d -> %d;\n", toNodeID, fromNodeID)
		}
	}

	dotContent += "}\n"

	// Write DOT file
	if err := os.WriteFile(fmt.Sprintf("%s/dependency.dot", path), []byte(dotContent), 0644); err != nil {
		panic(err)
	}

	// Generate PNG
	cmd := exec.Command("dot", "-Tpng", fmt.Sprintf("%s/dependency.dot", path), "-o", fmt.Sprintf("%s/graph.png", path))
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}
