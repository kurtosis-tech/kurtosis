package operation_parallelizer

import (
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

type Operation func() error

func RunOperationsInParallel(operations map[OperationID]Operation) (map[OperationID]bool, map[OperationID]error) {
	workerPool := workerpool.New(maxNumConcurrentRequests)
	resultsChan := make(chan operationResult, len(operations))

	for id, op := range operations {
		workerPool.Submit(getWorkerTask(id, op, resultsChan))
	}

	workerPool.StopWait()
	close(resultsChan)

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

	return successfulOperationIDs, failedOperationIDs
}

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
// It's VERY important that we call a function to generate the lambda, rather than inlining a lambda,
// because if we don't then 'dockerObjectId' will be the same for all tasks (and it will be the
// value of the last iteration of the loop)
// https://medium.com/swlh/use-pointer-of-for-range-loop-variable-in-go-3d3481f7ffc9
// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
func getWorkerTask(id OperationID, operation Operation, resultsChan chan operationResult) func(){
	return func() {
		operationResultErr := operation()
		resultsChan <- operationResult{
			id: id,
			resultErr: operationResultErr,
		}
	}
}