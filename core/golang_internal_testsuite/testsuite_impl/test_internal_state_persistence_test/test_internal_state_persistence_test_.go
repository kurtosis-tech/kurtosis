/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package test_internal_state_persistence_test

import (
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-testsuite-api-lib/golang/lib/testsuite"
	"github.com/palantir/stacktrace"
)

const (
	newInternalState = "Wow, what a change!"
)

// Tests that the internal state of a test that gets modified in the Setup method gets persisted to the Run method
type TestInternalStatePersistenceTest struct {
	internalState string
}

func (t *TestInternalStatePersistenceTest) Configure(builder *testsuite.TestConfigurationBuilder) {
	builder.WithSetupTimeoutSeconds(10).WithRunTimeoutSeconds(10)
}

func (t *TestInternalStatePersistenceTest) Setup(networkCtx *networks.NetworkContext) (networks.Network, error) {
	t.internalState = newInternalState
	return networkCtx, nil
}

func (t *TestInternalStatePersistenceTest) Run(network networks.Network) error {
	if t.internalState != newInternalState {
		return stacktrace.NewError("Expected test's internal state in the run method to be '%v' but was '%v'", newInternalState, t.internalState)
	}
	return nil
}



