package operation_parallelizer

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

type InvalidOperation struct {
	msg string
}

func NewInvalidOperation(msg string) error {
	return &InvalidOperation{msg}
}

func (err *InvalidOperation) Error() string {
	return fmt.Sprintf("Invalid Job: %s", err.msg)
}

var (
	RandomError = NewInvalidOperation("This error was random.")
	DifferentError = NewInvalidOperation("This error was just different.")

	doSomething Operation = func() error {
		for i := 0; i < 5; i++ {
			// doing something
		}
		return nil
	}
	doSomethingError Operation = func() error{
		// do something
		return RandomError
	}
)


func TestOperationsInParallelReturnsSuccessfulOperations(t *testing.T){
	operations := map[OperationID]Operation{}

	operations["first"] = doSomething
	operations["second"] = doSomething
	operations["third"] = doSomething

	success, failed := RunOperationsInParallel(operations)

	numSucceeded := len(success)
	numFailed := len(failed)

	require.Equal(t, 0, numFailed)
	require.Equal(t, 3, numSucceeded)
}

func TestOperationInParallelReturnsFailedOperations(t *testing.T){
	operations := map[OperationID]Operation{}

	operations["first"] = doSomethingError
	operations["second"] = doSomethingError
	operations["third"] = doSomethingError

	success, failed := RunOperationsInParallel(operations)
	numSucceeded := len(success)
	numFailed := len(failed)


	require.Equal(t, 3, numFailed)
	require.Equal(t, 0, numSucceeded)
	for _, err := range failed {
		require.ErrorIs(t, RandomError, err)
	}
}

func TestOperationInParallelReturnsBothSuccessAndFailedOperations(t *testing.T){
	operations := map[OperationID]Operation{}

	operations["first"] = doSomethingError
	operations["second"] = doSomething
	operations["third"] = doSomething

	success, failed := RunOperationsInParallel(operations)
	numSucceeded := len(success)
	numFailed := len(failed)


	require.Equal(t, 1, numFailed)
	require.Equal(t, 2, numSucceeded)
	for id, err := range failed {
		require.Equal(t, "first", string(id))
		require.ErrorIs(t, RandomError, err)
	}
}

func TestOperationsInParallelUsingSharedVariablesReturnsCorrectResults(t *testing.T){
	operations := map[OperationID]Operation{}

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

	operations["first"] = doSomethingTogether
	operations["second"] = doSomethingTogether
	operations["third"] = doSomethingTogether

	success, _ := RunOperationsInParallel(operations)
	numSucceeded := len(success)

	require.Equal(t, 3, numSucceeded)
	require.Equal(t, 30, p) // p should equal three after all operations increase the counter 10 times
}

func TestOperationsInParallelUsingSharedChannelReturnsCorrectResults(t *testing.T){
	operations := map[OperationID]Operation{}

	operationData := make(chan string, 3)
	var sendDataInChannel Operation = func() error {
		operationData <- "Hello!"
		return nil
	}

	operations["first"] = sendDataInChannel
	operations["second"] = sendDataInChannel
	operations["third"] = sendDataInChannel

	success, _ := RunOperationsInParallel(operations)
	numSucceeded := len(success)

	require.Equal(t, 3, numSucceeded)
	require.Equal(t, 3, len(operationData))
	for _, _ = range success {
		data :=<- operationData
		require.Equal(t, "Hello!", data)
	}
}