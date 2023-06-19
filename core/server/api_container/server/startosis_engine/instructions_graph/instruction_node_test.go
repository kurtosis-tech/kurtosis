package instructions_graph

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInsertSorted_InsertAtTheEnd(t *testing.T) {
	uuid1 := kurtosis_starlark_framework.InstructionUuid("a")
	uuid2 := kurtosis_starlark_framework.InstructionUuid("b")
	uuid3 := kurtosis_starlark_framework.InstructionUuid("c")

	initialSlice := []kurtosis_starlark_framework.InstructionUuid{
		uuid1,
		uuid2,
	}
	result := insertSorted(initialSlice, uuid3)
	expected := []kurtosis_starlark_framework.InstructionUuid{
		uuid1,
		uuid2,
		uuid3,
	}
	require.Equal(t, expected, result)
}

func TestInsertSorted_InsertAtTheBeginning(t *testing.T) {
	uuid1 := kurtosis_starlark_framework.InstructionUuid("a")
	uuid2 := kurtosis_starlark_framework.InstructionUuid("b")
	uuid3 := kurtosis_starlark_framework.InstructionUuid("c")

	initialSlice := []kurtosis_starlark_framework.InstructionUuid{
		uuid2,
		uuid3,
	}
	result := insertSorted(initialSlice, uuid1)
	expected := []kurtosis_starlark_framework.InstructionUuid{
		uuid1,
		uuid2,
		uuid3,
	}
	require.Equal(t, expected, result)
}

func TestInsertSorted_InsertInTheMiddle(t *testing.T) {
	uuid1 := kurtosis_starlark_framework.InstructionUuid("a")
	uuid2 := kurtosis_starlark_framework.InstructionUuid("b")
	uuid3 := kurtosis_starlark_framework.InstructionUuid("c")

	initialSlice := []kurtosis_starlark_framework.InstructionUuid{
		uuid1,
		uuid3,
	}
	result := insertSorted(initialSlice, uuid2)
	expected := []kurtosis_starlark_framework.InstructionUuid{
		uuid1,
		uuid2,
		uuid3,
	}
	require.Equal(t, expected, result)
}
