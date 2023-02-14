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

	data interface{}

	// Nil error indicates the operation executed successfully
	resultErr error
}

// Users can return any data to this through interface{} and downcast to their desired type when consuming the data in [successfulOps] or leave nil
type Operation func() (interface{}, error)

func RunOperationsInParallel(operations map[OperationID]Operation) (map[OperationID]interface{}, map[OperationID]error) {
	workerPool := workerpool.New(maxNumConcurrentRequests)
	resultsChan := make(chan operationResult, len(operations))

	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	// It's VERY important that we call a function to generate the lambda, rather than inlining a lambda,
	// because if we don't then 'id' will be the same for all tasks (and it will be the
	// value of the last iteration of the loop)
	// https://medium.com/swlh/use-pointer-of-for-range-loop-variable-in-go-3d3481f7ffc9
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	for id, op := range operations {
		workerPool.Submit(getWorkerTask(id, op, resultsChan))
	}

	workerPool.StopWait()
	close(resultsChan)

	successfulOperations := map[OperationID]interface{}{}
	failedOperations := map[OperationID]error{}
	for taskResult := range resultsChan {
		id := taskResult.id
		data := taskResult.data
		err := taskResult.resultErr
		if err == nil {
			successfulOperations[id] = data
		} else {
			failedOperations[id] = err
		}
	}

	return successfulOperations, failedOperations
}

func getWorkerTask(id OperationID, operation Operation, resultsChan chan operationResult) func(){
	return func() {
		data, err := operation()
		resultsChan <- operationResult{
			id: id,
			data: data,
			resultErr: err,
		}
	}
}