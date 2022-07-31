package operation_parallelizer

import (
	"errors"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

var (
	randomError = errors.New("This error was random.")

	doSomething Operation = func() error {
		for i := 0; i < 5; i++ {
			// doing something
		}
		return nil
	}
	doSomethingError Operation = func() error{
		// do something
		return randomError
	}
)


func TestOperationsInParallelReturnsSuccessfulOperations(t *testing.T){
	operations := map[OperationID]Operation{
		"first": doSomething,
		"second": doSomething,
		"third": doSomething,
	}

	success, failed := RunOperationsInParallel(operations)

	numSucceeded := len(success)
	numFailed := len(failed)

	require.Equal(t, 0, numFailed)
	require.Equal(t, 3, numSucceeded)
}

func TestOperationInParallelReturnsFailedOperations(t *testing.T){
	operations := map[OperationID]Operation{
		"first": doSomethingError,
		"second": doSomethingError,
		"third": doSomethingError,
	}

	success, failed := RunOperationsInParallel(operations)
	numSucceeded := len(success)
	numFailed := len(failed)


	require.Equal(t, 3, numFailed)
	require.Equal(t, 0, numSucceeded)
	for _, err := range failed {
		require.ErrorIs(t, randomError, err)
	}
}

func TestOperationInParallelReturnsBothSuccessAndFailedOperations(t *testing.T){
	operations := map[OperationID]Operation{
		"first":  doSomethingError,
		"second": doSomething,
		"third":  doSomething,
	}

	success, failed := RunOperationsInParallel(operations)
	numSucceeded := len(success)
	numFailed := len(failed)


	require.Equal(t, 1, numFailed)
	require.Equal(t, 2, numSucceeded)
	for id, err := range failed {
		require.Equal(t, "first", string(id))
		require.ErrorIs(t, randomError, err)
	}
}

func TestOperationsInParallelUsingSharedVariablesReturnsCorrectResults(t *testing.T){
	p := 0
	incLock := sync.Mutex{}
	var doSomethingTogether Operation = func() error {
		for i := 0; i < 10; i++ {
			incLock.Lock()
			p++
			incLock.Unlock()
		}
		return nil
	}

	operations := map[OperationID]Operation{
		"first":  doSomethingTogether,
		"second": doSomethingTogether,
		"third":  doSomethingTogether,
	}

	success, _ := RunOperationsInParallel(operations)
	numSucceeded := len(success)

	require.Equal(t, 3, numSucceeded)
	require.Equal(t, 30, p) // p should equal three after all operations increase the counter 10 times
}

func TestOperationsInParallelUsingSharedChannelReturnsCorrectResults(t *testing.T){
	operationData := make(chan string, 3)
	var sendDataInChannel Operation = func() error {
		operationData <- "Hello!"
		return nil
	}

	operations := map[OperationID]Operation{
		"first":  sendDataInChannel,
		"second": sendDataInChannel,
		"third":  sendDataInChannel,
	}

	success, _ := RunOperationsInParallel(operations)
	numSucceeded := len(success)

	require.Equal(t, 3, numSucceeded)
	require.Equal(t, 3, len(operationData))
	for _, _ = range success {
		data :=<- operationData
		require.Equal(t, "Hello!", data)
	}
}