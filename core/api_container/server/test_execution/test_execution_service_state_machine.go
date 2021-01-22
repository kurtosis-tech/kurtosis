/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_execution

import (
	"github.com/palantir/stacktrace"
	"sync"
)

const (
	// These form a linear state machine, where progress is one-way
	waitingForSuiteRegistration         serviceState = "WAITING_FOR_SUITE_REGISTRATION"
	waitingForTestExecutionRegistration serviceState = "WAITING_FOR_TEST_EXECUTION_REGISTRATION"
	waitingForExecutionCompletion       serviceState = "WAITING_FOR_TEST_EXECUTION_COMPLETION"
	testsuiteExited                     serviceState = "TESTSUITE_EXITED"

)

type serviceState string
var stateOrder = []serviceState{
	waitingForSuiteRegistration,
	waitingForTestExecutionRegistration,
	waitingForExecutionCompletion,
	testsuiteExited,
}

type stateAssertion func(expected serviceState, actual serviceState) bool
var inStateAssertion = func(a serviceState, b serviceState) bool {
	return a == b
}
var notInStateAssertion = func(a serviceState, b serviceState) bool {
	return a != b
}

type testExecutionServiceStateMachine struct {
	mutex *sync.Mutex
	stateIdx int
}

// TODO Write tests for me!!!
func newTestExecutionServiceStateMachine() *testExecutionServiceStateMachine {
	if len(stateOrder) == 0 {
		panic("Cannot construct state machine with no states!")
	}
	return &testExecutionServiceStateMachine{
		mutex: &sync.Mutex{},
		stateIdx: 0,
	}
}

func (machine *testExecutionServiceStateMachine) assert(expectedState serviceState) error {
	machine.mutex.Lock()
	defer machine.mutex.Unlock()

	if err := machine.throwErrorIfStatesDontMatch(expectedState); err != nil {
		return stacktrace.Propagate(err, "Actual state doesn't match expected state")
	}

	return nil
}

func (machine *testExecutionServiceStateMachine) assertAndAdvance(expectedState serviceState) error {
	machine.mutex.Lock()
	defer machine.mutex.Unlock()

	if err := machine.throwErrorIfStatesDontMatch(expectedState); err != nil {
		return stacktrace.Propagate(err, "Couldn't advance state machine; actual state doesn't match expected state")
	}

	if machine.stateIdx == len(stateOrder) - 1 {
		return stacktrace.NewError("Cannot advance test execution state machine; already in final state")
	}
	machine.stateIdx++

	return nil
}

func (machine testExecutionServiceStateMachine) throwErrorIfStatesDontMatch(expectedState serviceState) error {
	actualState := stateOrder[machine.stateIdx]
	if actualState != expectedState {
		return stacktrace.NewError(
			"Actual state '%v' doesn't match expected state '%v'",
			actualState,
			expectedState)
	}
	return nil
}

