package docker_task_parallelizer

import (
	"context"
	"github.com/gammazero/workerpool"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
)

type dockerTaskResult struct {
	containerId string
	resultErr   error // Nil if no issue
}

type ContainerIdConsumingTask func(ctx context.Context, dockerManager *docker_manager.DockerManager, containerId string) error

// RunDockerTaskInParallel executes the given function for all the container IDs using the requested parallelism level,
// and stores the results.
func RunDockerTaskInParallel(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
	parallelism int,
	funcToApply ContainerIdConsumingTask,
	containerIdsToOperateOn map[string]bool,
) (
	resultSuccessfulContainerIds map[string]bool,
	resultErroredContainerIds map[string]error,
) {
	workerPool := workerpool.New(parallelism)

	resultsChan := make(chan dockerTaskResult, len(containerIdsToOperateOn))
	for containerId := range containerIdsToOperateOn {
		workerPool.Submit(func(){
			taskExecutionResult := funcToApply(ctx, dockerManager, containerId)
			resultsChan <- dockerTaskResult{
				containerId: containerId,
				resultErr:   taskExecutionResult,
			}
		})
	}
	workerPool.StopWait()
	close(resultsChan)

	successfulContainerIds := map[string]bool{}
	erroredContainerIds := map[string]error{}
	for result := range resultsChan {
		containerId := result.containerId
		resultErr := result.resultErr
		if resultErr == nil {
			successfulContainerIds[containerId] = true
		} else {
			erroredContainerIds[containerId] = resultErr
		}
	}
	return successfulContainerIds, erroredContainerIds
}
