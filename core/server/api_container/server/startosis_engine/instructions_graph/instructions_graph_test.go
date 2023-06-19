package instructions_graph

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/mock_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInstructionGraph_SingletonGraph(t *testing.T) {
	graph := NewInstructionGraph()

	instruction1 := generateMockInstruction(t, "instruction1")
	_, err := graph.AddInstructionToGraph(instruction1)
	require.NoError(t, err)

	plan, err := graph.generatePlan()
	require.NoError(t, err)
	require.Len(t, plan, 1)
	require.Equal(t, plan[0].instruction, instruction1)
}

func TestInstructionGraph_EmptyGraph(t *testing.T) {
	graph := NewInstructionGraph()

	plan, err := graph.generatePlan()
	require.NoError(t, err)
	require.Empty(t, plan)
}

func TestInstructionGraph_ThreeNodesChainGraph(t *testing.T) {
	graph := NewInstructionGraph()

	instruction1 := generateMockInstruction(t, "instruction1")
	uuid1, err := graph.AddInstructionToGraph(instruction1)
	require.NoError(t, err)

	instruction2 := generateMockInstruction(t, "instruction2")
	uuid2, err := graph.AddInstructionToGraph(instruction2, uuid1)
	require.NoError(t, err)

	instruction3 := generateMockInstruction(t, "instruction3")
	_, err = graph.AddInstructionToGraph(instruction3, uuid2)
	require.NoError(t, err)

	plan, err := graph.generatePlan()
	require.NoError(t, err)
	require.Len(t, plan, 3)
	require.Equal(t, plan[0].instruction, instruction1)
	require.Equal(t, plan[1].instruction, instruction2)
	require.Equal(t, plan[2].instruction, instruction3)
}

func TestInstructionGraph_ThreeNodesForkGraph(t *testing.T) {
	graph := NewInstructionGraph()

	instruction1 := generateMockInstruction(t, "instruction1")
	uuid1, err := graph.AddInstructionToGraph(instruction1)
	require.NoError(t, err)

	instruction2 := generateMockInstruction(t, "instruction2")
	_, err = graph.AddInstructionToGraph(instruction2, uuid1)
	require.NoError(t, err)

	instruction3 := generateMockInstruction(t, "instruction3")
	_, err = graph.AddInstructionToGraph(instruction3, uuid1)
	require.NoError(t, err)

	plan, err := graph.generatePlan()
	require.NoError(t, err)
	require.Len(t, plan, 3)
	require.Equal(t, plan[0].instruction, instruction1)
	require.Equal(t, plan[1].instruction, instruction2)
	require.Equal(t, plan[2].instruction, instruction3)
}

func TestInstructionGraph_ThreeNodesInvertedForkTwoHeadsGraph(t *testing.T) {
	graph := NewInstructionGraph()

	instruction1 := generateMockInstruction(t, "instruction1")
	uuid1, err := graph.AddInstructionToGraph(instruction1)
	require.NoError(t, err)

	instruction2 := generateMockInstruction(t, "instruction2")
	uuid2, err := graph.AddInstructionToGraph(instruction2)
	require.NoError(t, err)

	instruction3 := generateMockInstruction(t, "instruction3")
	_, err = graph.AddInstructionToGraph(instruction3, uuid1, uuid2)
	require.NoError(t, err)

	plan, err := graph.generatePlan()
	require.NoError(t, err)
	require.Len(t, plan, 3)
	require.Equal(t, plan[0].instruction, instruction1)
	require.Equal(t, plan[1].instruction, instruction2)
	require.Equal(t, plan[2].instruction, instruction3)
}

func TestInstructionGraph_GraphForWhichASimpleBreadthFirstWouldFail(t *testing.T) {
	graph := NewInstructionGraph()

	instruction1 := generateMockInstruction(t, "instruction1")
	uuid1, err := graph.AddInstructionToGraph(instruction1)
	require.NoError(t, err)

	instruction2 := generateMockInstruction(t, "instruction2")
	uuid2, err := graph.AddInstructionToGraph(instruction2, uuid1)
	require.NoError(t, err)

	instruction3 := generateMockInstruction(t, "instruction3")
	uuid3, err := graph.AddInstructionToGraph(instruction3)
	require.NoError(t, err)

	instruction4 := generateMockInstruction(t, "instruction4")
	_, err = graph.AddInstructionToGraph(instruction4, uuid2, uuid3)
	require.NoError(t, err)

	plan, err := graph.generatePlan()
	require.NoError(t, err)
	require.Len(t, plan, 4)
	require.Equal(t, plan[0].instruction, instruction1)
	require.Equal(t, plan[1].instruction, instruction3)
	require.Equal(t, plan[2].instruction, instruction2)
	require.Equal(t, plan[3].instruction, instruction4)
}

func TestInstructionGraph_CyclicGraphWithHeadInCycle(t *testing.T) {
	graph := NewInstructionGraph()

	instruction1 := generateMockInstruction(t, "instruction1")
	uuid1, err := graph.AddInstructionToGraph(instruction1)
	require.NoError(t, err)

	instruction2 := generateMockInstruction(t, "instruction2")
	uuid2, err := graph.AddInstructionToGraph(instruction2, uuid1)
	require.NoError(t, err)

	instruction3 := generateMockInstruction(t, "instruction3")
	uuid3, err := graph.AddInstructionToGraph(instruction3, uuid2)
	require.NoError(t, err)

	// manual update to the graph. By construction, the graph cannot be cyclic. Here we have to hack it to
	// be able to test it, so that even if in the future is becomes possible to have cycles in the graph
	// we will handle it nicely
	graph.nodes[uuid1].parents = append(graph.nodes[uuid1].parents, uuid3)
	graph.nodes[uuid3].children = append(graph.nodes[uuid3].children, uuid1)

	plan, err := graph.generatePlan()
	require.Nil(t, plan)
	require.Contains(t, err.Error(), "The Kurtosis Plan instructions graph has one cycle at least. Kurtosis is not able to process it. This is a Kurtosis internal bug")
}

func TestInstructionGraph_CyclicGraph(t *testing.T) {
	graph := NewInstructionGraph()

	instruction1 := generateMockInstruction(t, "instruction1")
	uuid1, err := graph.AddInstructionToGraph(instruction1)
	require.NoError(t, err)

	instruction2 := generateMockInstruction(t, "instruction2")
	uuid2, err := graph.AddInstructionToGraph(instruction2, uuid1)
	require.NoError(t, err)

	instruction3 := generateMockInstruction(t, "instruction3")
	uuid3, err := graph.AddInstructionToGraph(instruction3, uuid2)
	require.NoError(t, err)

	// manual update to the graph. By construction, the graph cannot be cyclic. Here we have to hack it to
	// be able to test it, so that even if in the future is becomes possible to have cycles in the graph
	// we will handle it nicely
	graph.nodes[uuid2].parents = append(graph.nodes[uuid2].parents, uuid3)
	graph.nodes[uuid3].children = append(graph.nodes[uuid3].children, uuid2)

	plan, err := graph.generatePlan()
	require.Nil(t, plan)
	require.Contains(t, err.Error(), "The Kurtosis Plan instructions graph has one cycle at least. Kurtosis is not able to process it. This is a Kurtosis internal bug")
}

// ///////////////////// TEST HELPERS BELOW ///////////////////////
func generateMockInstruction(t *testing.T, instructionName string) *mock_instruction.MockKurtosisInstruction {
	instruction := mock_instruction.NewMockKurtosisInstruction(t)
	instruction.EXPECT().Uuid().Maybe().Return(kurtosis_starlark_framework.InstructionUuid(instructionName))
	instruction.EXPECT().String().Maybe().Return(instructionName)
	return instruction
}
