package instructions_plan

import (
	"testing"

	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/mock_instruction"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"

	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/types"
)

func TestAddInstruction(t *testing.T) {
	plan := NewInstructionsPlan()

	instruction1 := mock_instruction.NewMockKurtosisInstruction(t)
	instruction1ReturnedValue := starlark.None
	require.NoError(t, plan.AddInstruction(instruction1, instruction1ReturnedValue))

	require.Len(t, plan.instructionsSequence, 1)
	scheduledInstructionUuid := plan.instructionsSequence[0]
	require.Contains(t, plan.scheduledInstructionsIndex, scheduledInstructionUuid)
	scheduledInstruction, found := plan.scheduledInstructionsIndex[scheduledInstructionUuid]
	require.True(t, found)
	require.Equal(t, scheduledInstruction.GetInstruction(), instruction1)
	require.False(t, scheduledInstruction.IsExecuted())
	require.Equal(t, scheduledInstruction.GetReturnedValue(), instruction1ReturnedValue)
}

func TestAddScheduledInstruction(t *testing.T) {
	plan := NewInstructionsPlan()

	instruction1Uuid := types.ScheduledInstructionUuid("instruction1")
	instruction1 := mock_instruction.NewMockKurtosisInstruction(t)
	instruction1ReturnedValue := starlark.MakeInt(1)
	scheduleInstruction := NewScheduledInstruction(instruction1Uuid, instruction1, instruction1ReturnedValue)
	scheduleInstruction.executed = true

	plan.AddScheduledInstruction(scheduleInstruction)

	require.Len(t, plan.instructionsSequence, 1)
	scheduledInstructionUuid := plan.instructionsSequence[0]
	require.Equal(t, instruction1Uuid, scheduledInstructionUuid)
	require.Contains(t, plan.scheduledInstructionsIndex, scheduledInstructionUuid)
	addedScheduledInstruction, found := plan.scheduledInstructionsIndex[scheduledInstructionUuid]
	require.True(t, found)
	require.NotSame(t, scheduleInstruction, addedScheduledInstruction) // validate the instruction was cloned
	require.Equal(t, addedScheduledInstruction.GetInstruction(), instruction1)
	require.True(t, addedScheduledInstruction.IsExecuted())
	require.Equal(t, addedScheduledInstruction.GetReturnedValue(), instruction1ReturnedValue)
}

func TestGeneratePlan(t *testing.T) {
	plan := NewInstructionsPlan()

	// add instruction1 which is marked as executed
	instruction1Uuid := types.ScheduledInstructionUuid("instruction1")
	instruction1 := mock_instruction.NewMockKurtosisInstruction(t)
	instruction1ReturnedValue := starlark.None
	scheduleInstruction1 := NewScheduledInstruction(instruction1Uuid, instruction1, instruction1ReturnedValue)
	scheduleInstruction1.Executed(true)

	plan.scheduledInstructionsIndex[instruction1Uuid] = scheduleInstruction1
	plan.instructionsSequence = append(plan.instructionsSequence, instruction1Uuid)

	// add instruction2 which by default is not executed
	instruction2Uuid := types.ScheduledInstructionUuid("instruction2")
	instruction2 := mock_instruction.NewMockKurtosisInstruction(t)
	instruction2ReturnedValue := starlark.MakeInt(1)
	scheduleInstruction2 := NewScheduledInstruction(instruction2Uuid, instruction2, instruction2ReturnedValue)

	plan.scheduledInstructionsIndex[instruction2Uuid] = scheduleInstruction2
	plan.instructionsSequence = append(plan.instructionsSequence, instruction2Uuid)

	// generate plan and validate it
	instructionsSequence, err := plan.GeneratePlan()
	require.Nil(t, err)

	require.Len(t, instructionsSequence, 2)
	scheduledInstruction1 := instructionsSequence[0]
	require.Equal(t, scheduledInstruction1.GetInstruction(), instruction1)
	require.True(t, scheduledInstruction1.IsExecuted())
	require.Equal(t, scheduledInstruction1.GetReturnedValue(), instruction1ReturnedValue)

	scheduledInstruction2 := instructionsSequence[1]
	require.Equal(t, scheduledInstruction2.GetInstruction(), instruction2)
	require.False(t, scheduledInstruction2.IsExecuted())
	require.Equal(t, scheduledInstruction2.GetReturnedValue(), instruction2ReturnedValue)
}
