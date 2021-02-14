/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_execution

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStartingState(t *testing.T) {
	machine := newTestExecutionServiceStateMachine()
	assert.Equal(t, machine.stateIdx, 0)
}

func TestAssertOneOfStateSet(t *testing.T) {
	machine := newTestExecutionServiceStateMachine()

	expectedStartState := stateOrder[0]
	assert.Nil(t, machine.assertOneOfSet(map[serviceState]bool{expectedStartState: true}))

	inclusiveStateSet := map[serviceState]bool{
		expectedStartState: true,
		stateOrder[1]: true,
	}
	assert.Nil(t, machine.assertOneOfSet(inclusiveStateSet))

	wrongState := testsuiteExited
	assert.NotNil(t, machine.assertOneOfSet(map[serviceState]bool{wrongState: true}))
}

func TestRegularAssertAndAdvance(t *testing.T) {
	machine := newTestExecutionServiceStateMachine()

	wrongState := testsuiteExited
	assert.NotNil(t, machine.assertAndAdvance(wrongState))
	assert.Equal(t, 0, machine.stateIdx)

	correctState := stateOrder[0]
	assert.Nil(t, machine.assertAndAdvance(correctState))
	assert.Equal(t, 1, machine.stateIdx)
}

func TestAssertAndAdvanceAtEnd(t *testing.T) {
	machine := newTestExecutionServiceStateMachine()

	for i := 0; i < len(stateOrder) - 1; i++ {
		currentState := stateOrder[i]
		assert.Nil(
			t,
			machine.assertAndAdvance(currentState),
			"An error occurred advancing the state machine when it was at state '%v'",
			currentState,
		)
	}

	lastState := stateOrder[len(stateOrder) - 1]
	assert.NotNil(t, machine.assertAndAdvance(lastState))
}
