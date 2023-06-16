package graph

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
)

type InstructionNode struct {
	uuid NodeUuid

	hash        InstructionHash
	instruction kurtosis_instruction.KurtosisInstruction

	parents []NodeUuid

	children []NodeUuid
}

func newInstructionNode(uuid NodeUuid, instruction kurtosis_instruction.KurtosisInstruction, parents ...NodeUuid) *InstructionNode {
	return &InstructionNode{
		uuid:        uuid,
		hash:        newInstructionHash(instruction),
		instruction: instruction,
		parents:     parents,
		children:    []NodeUuid{},
	}
}

func (node *InstructionNode) addChildren(children ...NodeUuid) {
	node.children = append(node.children, children...)
}

func (node *InstructionNode) GetUuid() NodeUuid {
	return node.uuid
}

func (node *InstructionNode) GetHash() InstructionHash {
	return node.hash
}

func (node *InstructionNode) GetInstruction() kurtosis_instruction.KurtosisInstruction {
	return node.instruction
}
