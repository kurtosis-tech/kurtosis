package docker_task_parallelizer

import (
	"context"
	"github.com/gammazero/workerpool"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	// This should probably (?) be fine
	maxNumConcurrentRequestsToDocker = 25
)

type dockerOperationResult struct {
	dockerObjectId string
	resultErr   error // Nil if no issue
}

// An operation that consumes the Docker object ID, does something, and returns an error (or not)
type DockerOperation func(ctx context.Context, dockerManager *docker_manager.DockerManager, dockerObjectId string) error

// RunDockerOperationInParallelForKurtosisObject abstracts away a very common pattern that we have in DockerKurtosisBackend:
//  1) take a list of Kurtosis objects, keyed by its Docker ID
//  2) extract the Docker ID only
//  3) call an arbitrary Docker function using the ID
//  4) collect the results
//  5) key the results by the Kurtosis ID
func RunDockerOperationInParallelForKurtosisObject(
	ctx context.Context,
	// The objects that will be operated upon, keyed by their Docker ID
	dockerKeyedKurtosisObjects map[string]interface{},
	dockerManager *docker_manager.DockerManager,
	// Function that will be applied to each Kurtosis object for extracting its key
	// when categorizing the final results
	kurtosisKeyExtractor func(kurtosisObj interface{}) (string, error),
	operationToApplyToAllDockerObjects DockerOperation,
) (
	// Results of the Docker operation, keyed by Kurtosis object IDs (needs to be converted to the
	// proper type). Nil error == no error occurred
	resultSuccessfulKurtosisObjectIds map[string]bool,
	resultErroredKurtosisObjectIds map[string]error,
	resultErr error,
) {
	workerPool := workerpool.New(maxNumConcurrentRequestsToDocker)

	resultsChan := make(chan dockerOperationResult, len(dockerKeyedKurtosisObjects))
	for dockerObjectId := range dockerKeyedKurtosisObjects {
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

	successfulKurtosisObjIds := map[string]bool{}
	erroredKurtosisObjIds := map[string]error{}
	for taskResult := range resultsChan {
		dockerObjectId := taskResult.dockerObjectId
		kurtosisObj, found := dockerKeyedKurtosisObjects[dockerObjectId]
		if !found {
			return nil, nil, stacktrace.NewError("Unrequested Docker object with ID '%v was operated on!", dockerObjectId)
		}
		kurtosisObjectId, err := kurtosisKeyExtractor(kurtosisObj)
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "Couldn't extract Kurtosis object key for object with Docker object ID '%v'", dockerObjectId)
		}
		taskResultErr := taskResult.resultErr
		if taskResultErr == nil {
			successfulKurtosisObjIds[kurtosisObjectId] = true
		} else {
			erroredKurtosisObjIds[kurtosisObjectId] = taskResultErr
		}
	}
	return successfulKurtosisObjIds, erroredKurtosisObjIds, nil
}
