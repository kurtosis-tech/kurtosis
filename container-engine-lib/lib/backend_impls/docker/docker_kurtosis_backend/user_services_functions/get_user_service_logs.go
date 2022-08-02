package user_service_functions

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_log_streaming_readcloser"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"io"
)

func GetUserServiceLogs(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *service.ServiceFilters,
	shouldFollowLogs bool,
	dockerManager *docker_manager.DockerManager,
) (
	map[service.ServiceGUID]io.ReadCloser,
	map[service.ServiceGUID]error,
	error,
) {
	_, allDockerResources, err := shared_helpers.GetMatchingUserServiceObjsAndDockerResourcesNoMutex(ctx, enclaveId, filters, dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting user services matching filters '%+v'", filters)
	}

	//TODO use concurrency to improve perf
	successfulUserServicesLogs := map[service.ServiceGUID]io.ReadCloser{}
	erroredUserServices := map[service.ServiceGUID]error{}
	shouldCloseLogStreams := true
	for guid, resourcesForService := range allDockerResources {
		container := resourcesForService.ServiceContainer
		if container == nil {
			erroredUserServices[guid] = stacktrace.NewError("Cannot get logs for service '%v' as it has no container", guid)
			continue
		}

		rawDockerLogStream, err := dockerManager.GetContainerLogs(ctx, container.GetId(), shouldFollowLogs)
		if err != nil {
			serviceError := stacktrace.Propagate(err, "An error occurred getting logs for container '%v' for user service with GUID '%v'", container.GetName(), guid)
			erroredUserServices[guid] = serviceError
			continue
		}
		defer func() {
			if shouldCloseLogStreams {
				rawDockerLogStream.Close()
			}
		}()

		demultiplexedLogStream := docker_log_streaming_readcloser.NewDockerLogStreamingReadCloser(rawDockerLogStream)
		defer func() {
			if shouldCloseLogStreams {
				demultiplexedLogStream.Close()
			}
		}()

		successfulUserServicesLogs[guid] = demultiplexedLogStream
	}

	shouldCloseLogStreams = false
	return successfulUserServicesLogs, erroredUserServices, nil
}