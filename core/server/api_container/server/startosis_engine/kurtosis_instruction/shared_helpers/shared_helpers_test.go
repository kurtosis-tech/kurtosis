package shared_helpers

import (
	"context"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/stretchr/testify/assert"
	"go.starlark.net/starlark"
	"testing"
	"time"
)

const emptyServiceName = ""

func TestExecuteServiceAssertionWithRecipeWithTicker_ExecuteOnceAndTimeoutReached(t *testing.T) {
	executionTickChan := make(chan time.Time, 2)
	timeoutChan := make(chan time.Time, 2)
	execRunCount := 0
	assertRunCount := 0

	execFunc := func() (map[string]starlark.Comparable, error) {
		execRunCount += 1
		return nil, stacktrace.NewError("Exec Error")
	}
	assertFunc := func(map[string]starlark.Comparable) error {
		assertRunCount += 1
		return nil
	}
	executionTickChan <- time.Now()
	go func() {
		time.Sleep(10 * time.Millisecond)
		timeoutChan <- time.Now()
	}()
	_, tries, err := executeServiceAssertionWithRecipeWithTicker(emptyServiceName, execFunc, assertFunc, executionTickChan, timeoutChan)
	assert.NotNil(t, err)
	assert.Equal(t, tries, 1)
	assert.Equal(t, execRunCount, 1)
	assert.Equal(t, assertRunCount, 0)
}

func TestExecuteServiceAssertionWithRecipeWithTicker_ExecuteTriceAndSucceeds(t *testing.T) {
	executionTickChan := make(chan time.Time, 100)
	timeoutChan := make(chan time.Time)

	execFirstRun := true
	execFunc := func() (map[string]starlark.Comparable, error) {
		if execFirstRun {
			execFirstRun = false
			return nil, stacktrace.NewError("Exec Error")
		}
		return map[string]starlark.Comparable{}, nil
	}
	assertFirstRun := true
	assertFunc := func(map[string]starlark.Comparable) error {
		if assertFirstRun {
			assertFirstRun = false
			return stacktrace.NewError("Assert Error")
		}
		return nil
	}
	go func() {
		for i := 0; i < 10; i++ {
			time.Sleep(10 * time.Millisecond)
			executionTickChan <- time.Now()
		}
	}()
	_, tries, err := executeServiceAssertionWithRecipeWithTicker(emptyServiceName, execFunc, assertFunc, executionTickChan, timeoutChan)
	assert.Nil(t, err)
	assert.Equal(t, tries, 3)
}

func TestExecuteServiceAssertionWithRecipeWithTicker_ExecuteTimeoutAndCancelExec(t *testing.T) {
	executionTickChan := make(chan time.Time, 2)
	timeoutChan := make(chan time.Time, 2)
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	execFunc := func() (map[string]starlark.Comparable, error) {
		<-ctx.Done()
		return nil, stacktrace.NewError("Exec Timeout")
	}
	assertFunc := func(map[string]starlark.Comparable) error {
		return nil
	}
	executionTickChan <- time.Now()
	go func() {
		time.Sleep(10 * time.Millisecond)
		timeoutChan <- time.Now()
		cancelFunc()
	}()
	_, tries, err := executeServiceAssertionWithRecipeWithTicker(emptyServiceName, execFunc, assertFunc, executionTickChan, timeoutChan)
	assert.NotNil(t, err)
	assert.Equal(t, tries, 1)
}
