package operation_parallelizer

import (
	"errors"
	"github.com/gammazero/workerpool"
)

const (
	maxNumConcurrentRequests = 25
)

type OperationID string

type operationResult struct {
	id OperationID

	// Nil error indicates the operation executed successfully
	resultErr error
}

// Users can attach any data to this struct and downcast to their desired type when consuming this data.
type OperationData struct {
	ID OperationID

	Data interface{}
}

var (
	OperationDataInconsistencyError = errors.New(
		"set of OperationIDs sent over the data channel is not 1:1 with set of OperationIDs of successful operations,"+
		"this could mean two things: 1. an error occurred in operation logic. or 2. one of the assumptions was broken")
)

type Operation func(dataChan chan OperationData) error

// In order to allow users to safely send and retrieve data in parallel operations, we gurantee that all data returned come from
// successful operations. To do that, we make the following assumptions:
// 1. If data is sent through [dataChan] in an operation, the operation will send data in all cases of that operations logic.
//	(meaning its not the case that data is not sent in some cases and not others)
// 2. If any operation in [operations] sends data, then all operations send data. (meaning its not the case that one operation does send data and another does not)
func RunOperationsInParallel(operations map[OperationID]Operation) (map[OperationID]bool, map[OperationID]error, chan OperationData, error) {
	workerPool := workerpool.New(maxNumConcurrentRequests)
	resultsChan := make(chan operationResult, len(operations))
	dataChan := make(chan OperationData, len(operations))

	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	// It's VERY important that we call a function to generate the lambda, rather than inlining a lambda,
	// because if we don't then 'id' will be the same for all tasks (and it will be the
	// value of the last iteration of the loop)
	// https://medium.com/swlh/use-pointer-of-for-range-loop-variable-in-go-3d3481f7ffc9
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	for id, op := range operations {
		workerPool.Submit(getWorkerTask(id, op, resultsChan, dataChan))
	}

	workerPool.StopWait()
	close(resultsChan)
	close(dataChan)

	successfulOperationIDs := map[OperationID]bool{}
	failedOperationIDs := map[OperationID]error{}
	for taskResult := range resultsChan {
		id := taskResult.id
		err := taskResult.resultErr
		if err == nil {
			successfulOperationIDs[id] = true
		} else {
			failedOperationIDs[id] = err
		}
	}

	// If data has been sent, make sure that data comes from successful operations
	if len(dataChan) > 0 {
		dataIDs := map[OperationID]bool{}
		for data := range dataChan {
			dataIDs[data.ID] = true
		}
		if !isDataOneToOneWithSuccessfulOps(dataIDs, successfulOperationIDs) {
			return nil, nil, nil, OperationDataInconsistencyError
		}
	}

	return successfulOperationIDs, failedOperationIDs, dataChan, nil
}

func getWorkerTask(id OperationID, operation Operation, resultsChan chan operationResult, dataChan chan OperationData) func(){
	return func() {
		operationResultErr := operation(dataChan)
		resultsChan <- operationResult{
			id: id,
			resultErr: operationResultErr,
		}
	}
}

// This is to ensure two things:
//	1. If data is returned to the [dataChan], it comes from a successful operation, so the user does not consume data from failed operations.
//  2. If an operation is successful and should have returned data, it does.
func isDataOneToOneWithSuccessfulOps(successfulIDs map[OperationID]bool, dataIDs map[OperationID]bool) bool {
	// Check if [dataIDs] is a subset of [successfulIDs]
	for dataID := range dataIDs {
		if _, found := successfulIDs[dataID]; !found {
			return false
		}
	}
	// Check if len of sets are equal (ensuring 1:1)
	if len(successfulIDs) != len(dataIDs) {
		return false
	}
	return true
}

