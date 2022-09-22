package docker_operation_parallelizer

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/stacktrace"
)

// RunDockerOperationInParallelForKurtosisObjects sits on top of RunDockerOperationInParallel, abstracting away a very
// common pattern that we have in DockerKurtosisBackend:
//  1) take a list of Kurtosis objects, keyed by its Docker ID
//  2) extract the Docker ID only
//  3) call an arbitrary Docker function using the ID
//  4) collect the results
//  5) key the results by the Kurtosis ID
func RunDockerOperationInParallelForKurtosisObjects(
	ctx context.Context,
	// The objects that will be operated upon, keyed by their Docker ID
	// TODO Replace this stupid interface{} thing when we get generics!!
	dockerKeyedKurtosisObjects map[string]interface{},
	dockerManager *docker_manager.DockerManager,
	// Function that will be applied to each Kurtosis object for extracting its key
	// when categorizing the final results
	// TODO Replace this stupid interface{} thing when we get generics!!
	kurtosisKeyExtractor func(kurtosisObj interface{}) (string, error),
	operationToApplyToAllDockerObjects DockerOperation,
) (
	// Results of the Docker operation, keyed by Kurtosis object IDs (needs to be converted to the
	// proper type). Nil error == no error occurred
	resultSuccessfulKurtosisObjectIds map[string]bool,
	resultErroredKurtosisObjectIds map[string]error,
	resultErr error,
) {
	dockerObjectIdSet := map[string]bool{}
	for dockerObjectId := range dockerKeyedKurtosisObjects {
		dockerObjectIdSet[dockerObjectId] = true
	}

	successfulDockerObjectIds, erroredDockerObjectIds := RunDockerOperationInParallel(
		ctx,
		dockerObjectIdSet,
		dockerManager,
		operationToApplyToAllDockerObjects,
	)

	successfulKurtosisObjIds := map[string]bool{}
	for dockerObjectId := range successfulDockerObjectIds {
		kurtosisObj, found := dockerKeyedKurtosisObjects[dockerObjectId]
		if !found {
			return nil, nil, stacktrace.NewError("Successfully ran Docker operation on Docker object with ID '%v', but that object wasn't requested to be operated on", dockerObjectId)
		}
		kurtosisObjectId, err := kurtosisKeyExtractor(kurtosisObj)
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "Couldn't extract Kurtosis object key for Docker object with ID '%v' that was successfully operated on", dockerObjectId)
		}
		successfulKurtosisObjIds[kurtosisObjectId] = true
	}

	erroredKurtosisObjIds := map[string]error{}
	for dockerObjectId, dockerOperationErr := range erroredDockerObjectIds {
		kurtosisObj, found := dockerKeyedKurtosisObjects[dockerObjectId]
		if !found {
			return nil, nil, stacktrace.NewError("An error occurred running Docker operation on Docker object with ID '%v', but that object wasn't requested to be operated on", dockerObjectId)
		}
		kurtosisObjectId, err := kurtosisKeyExtractor(kurtosisObj)
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "Couldn't extract Kurtosis object key for Docker object with ID '%v' that threw the following error when being operated on:\n%v", dockerObjectId, dockerOperationErr)
		}
		erroredKurtosisObjIds[kurtosisObjectId] = dockerOperationErr
	}

	return successfulKurtosisObjIds, erroredKurtosisObjIds, nil
}
