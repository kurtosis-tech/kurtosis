package docker_task_parallelizer

import (
	"context"
	"github.com/gammazero/workerpool"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/sirupsen/logrus"
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
		// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
		// It's VERY important that we call a function to generate the lambda, rather than inlining a lambda,
		// because if we don't then 'dockerObjectId' will be the same for all tasks (and it will be the
		// value of the last iteration of the loop)
		// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
		workerPool.Submit(getWorkerTask(
			ctx,
			dockerManager,
			dockerObjectId,
			operationToApplyToAllDockerObjects,
			resultsChan,
		))
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

// NOTE: It's very important we do this, rather than inlining the lambda; see the place where this is called
// for more information
func getWorkerTask(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
	dockerObjectId string,
	operationToApplyToAllDockerObjects DockerOperation,
	resultsChan chan dockerOperationResult,
) func() {
	logrus.Debugf("Creating a Docker concurrent operation task to operate on Docker object with ID '%v'", dockerObjectId)
	return func(){
		operationResultErr := operationToApplyToAllDockerObjects(ctx, dockerManager, dockerObjectId)
		resultsChan <- dockerOperationResult{
			dockerObjectId: dockerObjectId,
			resultErr:      operationResultErr,
		}
	}
}
