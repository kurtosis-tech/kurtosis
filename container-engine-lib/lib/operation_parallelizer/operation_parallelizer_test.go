package operation_parallelizer

import (
	"errors"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

var (
	randomError = errors.New("This error was random.")

	doSomething Operation = func(_ chan OperationData) error {
		for i := 0; i < 5; i++ {
			// doing something
		}
		return nil
	}
	doSomethingError Operation = func(_ chan OperationData) error{
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

	success, failed, _, err := RunOperationsInParallel(operations)
	require.NoError(t, err)

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

	success, failed, _, err := RunOperationsInParallel(operations)
	require.NoError(t, err)

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

	success, failed, _, err:= RunOperationsInParallel(operations)
	require.NoError(t, err)

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
	var doSomethingTogether Operation = func(_ chan OperationData) error {
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

	success, _, _, err := RunOperationsInParallel(operations)
	require.NoError(t, err)

	numSucceeded := len(success)

	require.Equal(t, 3, numSucceeded)
	require.Equal(t, 30, p) // p should equal three after all operations increase the counter 10 times
}

// We still test this case, however most users of RunOperationsInParallel should opt for using the [dataChan] that comes with the module
func TestOperationsInParallelUsingSharedChannelReturnsCorrectResults(t *testing.T){
	operationData := make(chan string, 3)
	var sendDataInChannel Operation = func(_ chan OperationData) error {
		operationData <- "Hello!"
		return nil
	}

	operations := map[OperationID]Operation{
		"first":  sendDataInChannel,
		"second": sendDataInChannel,
		"third":  sendDataInChannel,
	}

	success, _, _, err := RunOperationsInParallel(operations)
	require.NoError(t, err)

	numSucceeded := len(success)

	require.Equal(t, 3, numSucceeded)
	require.Equal(t, 3, len(operationData))
	for _, _ = range success {
		data :=<- operationData
		require.Equal(t, "Hello!", data)
	}
}

func TestOperationsInParallelUsingDeferFunctionsExecuteDeferCorrectly(t *testing.T){
	var operationWithDeferError Operation = func(dataChan chan OperationData) error {
		var p int = 1
		_ = 1 + p // just to make p used

		undo := true
		defer func() {
			if undo {
				p = 3
				dataChan <- OperationData{ID: OperationID("first"), Data: p}
			}
		}()

		return randomError
	}

	var operationWithDeferNoError Operation = func(dataChan chan OperationData) error {
		var p int = 1
		_ = 1 + p // just to make p used

		undo := true
		defer func() {
			if undo {
				p = 2
				dataChan <- OperationData{ID: OperationID("second"), Data: p}
			}
		}()

		undo = false
		return nil
	}

	operations := map[OperationID]Operation{
		"first":  operationWithDeferError,
		"second": operationWithDeferNoError,
	}

	success, failed, data, err := RunOperationsInParallel(operations)
	// should return error because a failed operation returned data, thus successful IDs and data IDs are not 1:1)
	require.ErrorIs(t, err, OperationDataInconsistencyError)

	numSucceeded := len(success)
	numFailed := len(failed)

	require.Equal(t, 1, numSucceeded)
	require.Equal(t, 1, numFailed)
	require.Equal(t, 1, len(data))

	for opID, err := range failed {
		require.Equal(t, "first", string(opID))
		require.ErrorIs(t, err, randomError)

		dataVal, found := data[OperationID("first")]
		require.True(t, found)
		require.Equal(t, "first", string(dataVal.ID))
		require.Equal(t, 3, dataVal.Data.(int))
	}
}

func TestIsDataOneToOneWithSuccessfulOpsReturnsCorrectResult(t *testing.T){
	successfulIDs:= map[OperationID]bool{
		"first":  true,
		"second": true,
	}
	dataIDs := map[OperationID]bool{
		"first":  true,
		"second": true,
	}

	// successfulIDs is not a subset of more data IDs, not 1:1
	moreDataIDs := map[OperationID]bool{
		"first": true,
	}

	require.True(t, isDataOneToOneWithSuccessfulOps(successfulIDs, dataIDs))
	require.False(t, isDataOneToOneWithSuccessfulOps(successfulIDs, moreDataIDs))
}
// add test to test data channel

// test that is returns data correctly
// test that it returns an error when assumptions are broken
