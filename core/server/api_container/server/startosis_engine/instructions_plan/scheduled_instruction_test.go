package instructions_plan

import (
	"testing"

	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/mock_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/types"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
)

func TestExecutedDefaultAndUpdate(t *testing.T) {
	instruction1Uuid := types.ScheduledInstructionUuid("instruction1")
	instruction1 := mock_instruction.NewMockKurtosisInstruction(t)
	instruction1ReturnedValue := starlark.MakeInt(1)
	scheduleInstruction := NewScheduledInstruction(instruction1Uuid, instruction1, instruction1ReturnedValue)

	require.False(t, scheduleInstruction.executed)
	scheduleInstruction.Executed(true)
	require.True(t, scheduleInstruction.executed)
}
