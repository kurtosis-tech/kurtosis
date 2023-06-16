package graph

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/stacktrace"
)

type InstructionGraph struct {
	nodes map[NodeUuid]*InstructionNode
}

func NewInstructionGraph() *InstructionGraph {
	return &InstructionGraph{
		nodes: map[NodeUuid]*InstructionNode{},
	}
}

func (graph *InstructionGraph) AddNode(instruction kurtosis_instruction.KurtosisInstruction, parents ...NodeUuid) (NodeUuid, error) {
	nodeUuid, err := GenerateNodeUuid()
	if err != nil {
		return "", stacktrace.Propagate(err, "Generate node UUID")
	}
	newNode := newInstructionNode(nodeUuid, instruction, parents...)
	for _, parentUuid := range parents {
		parent, found := graph.nodes[parentUuid]
		if !found {
			return "", stacktrace.NewError("Parent node UUID '%s' not found in graph. Unable to add new node to "+
				"graph. The parent should be added first", parentUuid)
		}
		parent.addChildren(newNode.uuid)
	}
	graph.nodes[newNode.uuid] = newNode
	return newNode.uuid, nil
}

func (graph *InstructionGraph) GetHeadNodes() []*InstructionNode {
	var headNodes []*InstructionNode
	for _, node := range graph.nodes {
		if len(node.parents) == 0 {
			headNodes = append(headNodes, node)
		}
	}
	return headNodes
}

func (graph *InstructionGraph) Traverse() ([]InstructionNode, error) {
	var instructionList []InstructionNode

	var processingNodeQueue []InstructionNode

	// initialize the node queue with all head nodes
	for _, headNode := range graph.GetHeadNodes() {
		processingNodeQueue = append(processingNodeQueue, *headNode)
	}

	// traverse the graph starting for the head nodes.
	// NOTE: for now, since the graph is a chain, there's a single head node, and each node has a single child
	for len(processingNodeQueue) > 0 {
		// add the instruction to the instructionList, and pop the node from the queue
		processedNode := processingNodeQueue[0]
		instructionList = append(instructionList, processedNode)
		processingNodeQueue = processingNodeQueue[1:]

		// add all the node children to the queue for processing
		for _, childUuid := range processedNode.children {
			child, found := graph.nodes[childUuid]
			if !found {
				return nil, stacktrace.NewError("Instruction graph is inconsistent, one of the child (%s) of "+
					"node %s could not be found", childUuid, processedNode.hash)
			}
			processingNodeQueue = append(processingNodeQueue, *child)
		}
	}
	return instructionList, nil
}

func (graph *InstructionGraph) Size() uint32 {
	return uint32(len(graph.nodes))
}
