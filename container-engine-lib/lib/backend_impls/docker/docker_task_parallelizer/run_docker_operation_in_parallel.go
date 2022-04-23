package docker_task_parallelizer

import (
	"context"
	"github.com/gammazero/workerpool"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
)

// RunDockerOperationInParallel will run a Docker operation on each of the object IDs, in parallel
func RunDockerOperationInParallel(
	ctx context.Context,
	// The IDs of the Docker objects to operate on
	dockerObjectIdSet map[string]bool,
	dockerManager *docker_manager.DockerManager,
	operationToApplyToAllDockerObjects DockerOperation,
) (
	map[string]bool,
	map[string]error,
){
	workerPool := workerpool.New(maxNumConcurrentRequestsToDocker)

	resultsChan := make(chan dockerOperationResult, len(dockerObjectIdSet))
	for dockerObjectId := range dockerObjectIdSet {
		workerPool.Submit(func(){
			operationResultErr := operationToApplyToAllDockerObjects(ctx, dockerManager, dockerObjectId)
			resultsChan <- dockerOperationResult{
				dockerObjectId: dockerObjectId,
				resultErr:      operationResultErr,
			}
		})
	}
	workerPool.StopWait()
	close(resultsChan)

	successfulDockerObjectIds := map[string]bool{}
	erroredDockerObjectIds := map[string]error{}
	for taskResult := range resultsChan {
		dockerObjectId := taskResult.dockerObjectId
		taskResultErr := taskResult.resultErr
		if taskResultErr == nil {
			successfulDockerObjectIds[dockerObjectId] = true
		} else {
			erroredDockerObjectIds[dockerObjectId] = taskResultErr
		}
	}
	return successfulDockerObjectIds, erroredDockerObjectIds
}
