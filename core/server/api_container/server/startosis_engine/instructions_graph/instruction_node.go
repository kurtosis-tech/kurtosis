package instructions_graph

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
)

// InstructionNode is a thin wrapper around a KurtosisInstruction to record its parents and children in the instruction
// graph
type InstructionNode struct {
	instruction kurtosis_instruction.KurtosisInstruction

	parents []kurtosis_starlark_framework.InstructionUuid

	children []kurtosis_starlark_framework.InstructionUuid
}

// newInstructionNode builds a new InstructionNode.
//
// Only the parents of the instructions are passed in here because at construction time, usually only the parents
// of a node are known. Children for this instruction can be added later with the addChildren function
func newInstructionNode(instruction kurtosis_instruction.KurtosisInstruction, parents ...kurtosis_starlark_framework.InstructionUuid) *InstructionNode {
	var sortedParents []kurtosis_starlark_framework.InstructionUuid
	for _, parentUuid := range parents {
		sortedParents = insertSorted(sortedParents, parentUuid)
	}

	return &InstructionNode{
		instruction: instruction,
		parents:     sortedParents,
		children:    []kurtosis_starlark_framework.InstructionUuid{},
	}
}

func (node *InstructionNode) GetInstruction() kurtosis_instruction.KurtosisInstruction {
	return node.instruction
}

func (node *InstructionNode) addChildren(children ...kurtosis_starlark_framework.InstructionUuid) {
	for _, child := range children {
		node.children = insertSorted(node.children, child)
	}
}

func insertSorted(sortedSlice []kurtosis_starlark_framework.InstructionUuid, uuidToInsert kurtosis_starlark_framework.InstructionUuid) []kurtosis_starlark_framework.InstructionUuid {
	newSortedSlice := make([]kurtosis_starlark_framework.InstructionUuid, len(sortedSlice)+1)
	newSortedSliceIdx := 0
	for sortedSliceIdx := 0; sortedSliceIdx < len(sortedSlice); sortedSliceIdx += 1 {
		newElementHasBeenInserted := newSortedSliceIdx > sortedSliceIdx
		if newElementHasBeenInserted || sortedSlice[sortedSliceIdx] <= uuidToInsert {
			newSortedSlice[newSortedSliceIdx] = sortedSlice[sortedSliceIdx]
			newSortedSliceIdx += 1
		} else {
			newSortedSlice[newSortedSliceIdx] = uuidToInsert
			newSortedSlice[newSortedSliceIdx+1] = sortedSlice[sortedSliceIdx]
			newSortedSliceIdx += 2
		}
	}
	if newSortedSliceIdx < len(newSortedSlice) {
		// if the new element hasn't been inserted after going through the entire sortedSlice, it must be inserted
		// at the end
		newSortedSlice[newSortedSliceIdx] = uuidToInsert
	}
	return newSortedSlice
}
