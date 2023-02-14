package operation_parallelizer

import (
	"errors"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

var (
	randomError = errors.New("This error was random.")

	doSomething Operation = func() (interface{}, error) {
		for i := 0; i < 5; i++ {
			// doing something
		}
		return nil, nil
	}
	doSomethingError Operation = func() (interface{}, error){
		// do something
		return nil, randomError
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
	var doSomethingTogether Operation = func() (interface{}, error) {
		for i := 0; i < 10; i++ {
			incLock.Lock()
			p++
			incLock.Unlock()
		}
		return nil, nil
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

func TestOperationsInParallelReturnsDataCorrectly(t *testing.T) {
	type CustomType string

	var doSomethingWithData Operation = func() (interface{}, error){
		return CustomType("Hello!"), nil
	}

	operations := map[OperationID]Operation{
		"first": doSomethingWithData,
	}

	success, _ := RunOperationsInParallel(operations)

	numSucceeded := len(success)

	require.Equal(t, 1, numSucceeded)
	dataVal, found := success[OperationID("first")]
	require.True(t, found)
	require.NotNil(t, dataVal)

	val, ok := dataVal.(CustomType)
	require.True(t, ok)
	require.Equal(t, "Hello!", string(val))
}

func TestOperationsInParallelUsingDeferFunctionsExecuteDeferCorrectly(t *testing.T){
	operationData := make(chan string, 1)
	var operationWithDeferError Operation = func() (interface{}, error) {
		undo := true
		defer func() {
			if undo {
				operationData <- "Hello!"
			}
		}()

		return nil, randomError
	}

	var operationWithDeferNoError Operation = func() (interface{}, error) {
		var p *int = new(int)
		*p = 1

		undo := true
		defer func() {
			if undo {
				*p = 2
			}
		}()

		undo = false
		return p, nil
	}

	operations := map[OperationID]Operation{
		"first":  operationWithDeferError,
		"second": operationWithDeferNoError,
	}

	success, failed := RunOperationsInParallel(operations)

	numSucceeded := len(success)
	numFailed := len(failed)

	require.Equal(t, 1, numSucceeded)
	require.Equal(t, 1, numFailed)

	dataVal, found := success["second"]
	require.True(t, found)
	require.NotNil(t, dataVal)
	val, ok := dataVal.(*int)
	require.True(t, ok)
	require.Equal(t, 1, *val)

	_, found = failed["first"]
	require.True(t, found)
	require.Equal(t, 1, len(operationData))
	chanData :=<- operationData
	require.Equal(t, "Hello!", chanData)
}

// Most users of RunOperationsInParallel should opt for using the return interface{} to return results, but we still test this case
func TestOperationsInParallelUsingSharedChannelReturnsCorrectResults(t *testing.T){
	operationData := make(chan string, 3)
	var sendDataInChannel Operation = func() (interface{}, error) {
		operationData <- "Hello!"
		return nil, nil
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
	for range success {
		data := <-operationData
		require.Equal(t, "Hello!", data)
	}
}
