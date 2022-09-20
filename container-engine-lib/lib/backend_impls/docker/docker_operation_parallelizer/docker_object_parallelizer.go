package docker_operation_parallelizer

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/operation_parallelizer"
	"github.com/sirupsen/logrus"
)

// DockerOperation represents an operation done on a Docker object (identified by Docker object ID)
type DockerOperation func(ctx context.Context, dockerManager *docker_manager.DockerManager, dockerObjectId string) error

// RunDockerOperationInParallel will run a Docker operation on each of the object IDs, in parallel
// NOTE: Each call to this will get its own threadpool, so it's possible overwhelm Docker with many calls to this;
// we can fix this if it becomes problematic
func RunDockerOperationInParallel(
	ctx context.Context,
	dockerObjectIdSet map[string]bool, // The IDs of the Docker objects to operate on
	dockerManager *docker_manager.DockerManager,
	operationToApplyToAllDockerObjects DockerOperation,
) (
	map[string]bool,
	map[string]error,
) {
	logrus.Debugf("Called RunDockerOperationInParallel on the following Docker object IDs: %+v", dockerObjectIdSet)
	dockerOperations := map[operation_parallelizer.OperationID]operation_parallelizer.Operation{}
	for dockerObjectID, _ := range dockerObjectIdSet {
		opID := operation_parallelizer.OperationID(dockerObjectID)

		// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
		// It's VERY important that we call a function to generate the lambda, rather than inlining a lambda,
		// because if we don't then 'dockerObjectID' will be the same for all tasks (and it will be the
		// value of the last iteration of the loop)
		// https://medium.com/swlh/use-pointer-of-for-range-loop-variable-in-go-3d3481f7ffc9
		// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
		dockerOperations[opID] = createDockerOperation(ctx, dockerObjectID, dockerManager, operationToApplyToAllDockerObjects)
	}

	successfulOperations, failedOperations := operation_parallelizer.RunOperationsInParallel(dockerOperations)

	successfulOperationIDStrs := map[string]bool{}
	failedOperationIDStrs := map[string]error{}
	for opID, _ := range successfulOperations {
		successfulOperationIDStrs[string(opID)] = true
	}
	for opID, _ := range failedOperations {
		failedOperationIDStrs[string(opID)] = failedOperations[opID]
	}

	return successfulOperationIDStrs, failedOperationIDStrs
}

func createDockerOperation(
	ctx context.Context,
	dockerObjectID string,
	dockerManager *docker_manager.DockerManager,
	operationToApplyToAllDockerObjects DockerOperation) operation_parallelizer.Operation {
	return func() (interface{}, error) {
		return nil, operationToApplyToAllDockerObjects(ctx, dockerManager, dockerObjectID)
	}
}
