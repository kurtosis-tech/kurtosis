package operation_parallelizer

import (
	"github.com/gammazero/workerpool"
)

type OperationID string

type OperationResult struct {
	id OperationID
	resultErr error
}

type Operation struct {
	Task func() error
}

const (
	maxNumConcurrentRequests = 25
)

func RunOperationsInParallel(operations map[OperationID]Operation) (map[OperationID]bool, map[OperationID]error) {
	workerPool := workerpool.New(maxNumConcurrentRequests)
	resultsChan := make(chan OperationResult, len(operations))

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

func getWorkerTask(id OperationID, operation Operation, resultsChan chan OperationResult) func(){
	return func() {
		OperationResultErr := operation.Task()
		resultsChan <- OperationResult{
			id: id,
			resultErr: OperationResultErr,
		}
	}
}