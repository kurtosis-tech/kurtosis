package user_service_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/libs/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/libs/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/libs/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/libs/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"io"
)

// It returns io.ReadCloser which is a tar stream. It's up to the caller to close the reader.
func CopyFilesFromUserService(
	ctx context.Context,
	enclaveId enclave.EnclaveUUID,
	serviceUuid service.ServiceUUID,
	srcPathOnContainer string,
	output io.Writer,
	dockerManager *docker_manager.DockerManager,
) error {
	_, serviceDockerResources, err := shared_helpers.GetSingleUserServiceObjAndResourcesNoMutex(ctx, enclaveId, serviceUuid, dockerManager)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting user service with UUID '%v' in enclave with ID '%v'", serviceUuid, enclaveId)
	}
	container := serviceDockerResources.ServiceContainer

	tarStreamReadCloser, err := dockerManager.CopyFromContainer(ctx, container.GetId(), srcPathOnContainer)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred copying content from sourcepath '%v' in container '%v' for user service '%v' in enclave '%v'",
			srcPathOnContainer,
			container.GetName(),
			serviceUuid,
			enclaveId,
		)
	}
	defer tarStreamReadCloser.Close()

	if _, err := io.Copy(output, tarStreamReadCloser); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred copying the bytes of TAR'd up files at '%v' on service '%v' to the output",
			srcPathOnContainer,
			serviceUuid,
		)
	}

	return nil
}
