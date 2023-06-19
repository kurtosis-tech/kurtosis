package instructions_graph

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/stacktrace"
	"sort"
	"sync"
)

type InstructionGraph struct {
	mutex sync.RWMutex
	nodes map[kurtosis_starlark_framework.InstructionUuid]*InstructionNode
}

func NewInstructionGraph() *InstructionGraph {
	return &InstructionGraph{
		mutex: sync.RWMutex{},
		nodes: map[kurtosis_starlark_framework.InstructionUuid]*InstructionNode{},
	}
}

// AddInstructionToGraph adds an instruction to the instruction graph.
// It makes sure the consistency of the graph is preserved by adding the created node as a child of all of its parents.
// As a consequence, all its parents MUST exist in the graph when this node is added. If one of the parents passed in
// does not exist in the graph, this function will throw
func (graph *InstructionGraph) AddInstructionToGraph(instruction kurtosis_instruction.KurtosisInstruction, parents ...kurtosis_starlark_framework.InstructionUuid) (kurtosis_starlark_framework.InstructionUuid, error) {
	graph.mutex.Lock()
	defer graph.mutex.Unlock()
	instructionUuid := instruction.Uuid()
	if existingInstruction, found := graph.nodes[instructionUuid]; found {
		return "", stacktrace.NewError("Trying to add Kurtosis instruction with UUID '%s' to the graph but an "+
			"instruction with this UUID already exists (new instruction: '%s'; existing instruction: '%s').",
			instructionUuid, instruction.String(), existingInstruction.instruction.String())
	}
	newNode := newInstructionNode(instruction, parents...)
	for _, parentUuid := range parents {
		parent, found := graph.nodes[parentUuid]
		if !found {
			return "", stacktrace.NewError("Parent instruction UUID '%s' not found in graph, this instruction "+
				"cannot be added. All its parents must be present in the graph before it can be added", parentUuid)
		}
		parent.addChildren(instructionUuid)
	}
	graph.nodes[instructionUuid] = newNode
	return instructionUuid, nil
}

// Iter iterates on the graph returning InstructionNodes one by one respecting their dependencies.
// See generatePlan for more details on how the function traverses the graph.
// Note that this function takes a Read Lock on the graph and releases it ONLY when the iteration is completely finished
// No updates to the graph can be made while there's one or more iterator left unfinished
func (graph *InstructionGraph) Iter() (<-chan *InstructionNode, error) {
	graph.mutex.RLock()
	instructions, err := graph.generatePlan()
	if err != nil {
		graph.mutex.Unlock()
		return nil, stacktrace.Propagate(err, "Unable to travers the Kurtosis Plan instructions graph")
	}
	graphIterator := make(chan *InstructionNode)
	go func() {
		defer close(graphIterator)
		graph.mutex.RUnlock()
		for _, instruction := range instructions {
			graphIterator <- instruction
		}
	}()
	return graphIterator, nil
}

func (graph *InstructionGraph) Size() uint32 {
	graph.mutex.RLock()
	defer graph.mutex.RUnlock()
	return uint32(len(graph.nodes))
}

// generatePlan traverses the graph running a topological sort and returns a list of all instructions that respects
// dependencies. If there's a  cycle in the graph, this method throws an error
// Note: This function is deterministic. Calling it multiple times will return the same instruction list.
func (graph *InstructionGraph) generatePlan() ([]*InstructionNode, error) {
	// we start by extracting the head nodes (i.e. nodes with no dependencies) into the nodesToProcessQueue
	// and by indexing each node with the number of parents they have. That's the node weight, in the topological
	// search logic
	var nodesToProcessQueue []*InstructionNode
	nodeWeights := map[kurtosis_starlark_framework.InstructionUuid]int{}
	for _, instructionNode := range graph.nodes {
		nodeWeights[instructionNode.instruction.Uuid()] = len(instructionNode.parents)
		if len(instructionNode.parents) == 0 {
			nodesToProcessQueue = append(nodesToProcessQueue, instructionNode)
		}
	}
	// we sort the head nodes list to keep the topological sort deterministic
	sort.Slice(nodesToProcessQueue, func(i, j int) bool {
		return nodesToProcessQueue[i].instruction.Uuid() < nodesToProcessQueue[j].instruction.Uuid()
	})

	var sortedKurtosisInstructionsList []*InstructionNode
	for len(nodesToProcessQueue) > 0 {
		// pop the first nodes to process out of the queue
		nodeToProcess := nodesToProcessQueue[0]
		nodesToProcessQueue = nodesToProcessQueue[1:]

		// Add it to the sorted list of instruction
		sortedKurtosisInstructionsList = append(sortedKurtosisInstructionsList, nodeToProcess)

		// Now that the node has been processed, update the weight of all its children
		for _, childNodeUuid := range nodeToProcess.children {
			if _, found := nodeWeights[childNodeUuid]; !found {
				return nil, stacktrace.NewError("Unable to find node with UUID '%s' in the node weight index. This is a Kurtosis internal bug", childNodeUuid)
			}
			nodeWeights[childNodeUuid] -= 1
			if nodeWeights[childNodeUuid] == 0 {
				// if the child being updated has no parent anymore, add it to the sorted list of instruction as well
				childNode, found := graph.nodes[childNodeUuid]
				if !found {
					return nil, stacktrace.NewError("Unable to find node with UUID '%s' in the graph nodes map. This is a Kurtosis internal bug", childNodeUuid)
				}
				nodesToProcessQueue = append(nodesToProcessQueue, childNode)
			}
		}
	}

	// Now all nodes weight should be zero, otherwise it means there's at least one cycle in the graph
	for _, nodeWeight := range nodeWeights {
		if nodeWeight > 0 {
			return nil, stacktrace.NewError("The Kurtosis Plan instructions graph has one cycle at least. Kurtosis is not able to process it. This is a Kurtosis internal bug")
		}
	}

	return sortedKurtosisInstructionsList, nil
}
