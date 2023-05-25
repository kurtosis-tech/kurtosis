package kubernetes_kurtosis_backend

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/networking_sidecar"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
)

// TODO: MIGRATE THIS FOLDER TO USE STRUCTURE OF USER_SERVICE_FUNCTIONS MODULE

func (backend *KubernetesKurtosisBackend) CreateNetworkingSidecar(ctx context.Context, enclaveUuid enclave.EnclaveUUID, serviceUuid service.ServiceUUID) (*networking_sidecar.NetworkingSidecar, error) {
	return nil, stacktrace.NewError("Networking sidecars aren't implemented for Kubernetes yet")
}

func (backend *KubernetesKurtosisBackend) GetNetworkingSidecars(ctx context.Context, filters *networking_sidecar.NetworkingSidecarFilters) (map[service.ServiceUUID]*networking_sidecar.NetworkingSidecar, error) {
	return nil, stacktrace.NewError("Networking sidecars aren't implemented for Kubernetes yet")
}

func (backend *KubernetesKurtosisBackend) RunNetworkingSidecarExecCommands(ctx context.Context, enclaveUuid enclave.EnclaveUUID, networkingSidecarsCommands map[service.ServiceUUID][]string) (successfulNetworkingSidecarExecResults map[service.ServiceUUID]*exec_result.ExecResult, erroredUserServiceUuids map[service.ServiceUUID]error, resultErr error) {
	return nil, nil, stacktrace.NewError("Networking sidecars aren't implemented for Kubernetes yet")
}

func (backend *KubernetesKurtosisBackend) StopNetworkingSidecars(ctx context.Context, filters *networking_sidecar.NetworkingSidecarFilters) (successfulUserServiceUuids map[service.ServiceUUID]bool, erroredUserServiceUuids map[service.ServiceUUID]error, resultErr error) {
	return nil, nil, stacktrace.NewError("Networking sidecars aren't implemented for Kubernetes yet")
}

func (backend *KubernetesKurtosisBackend) DestroyNetworkingSidecars(ctx context.Context, filters *networking_sidecar.NetworkingSidecarFilters) (successfulUserServiceUuids map[service.ServiceUUID]bool, erroredUserServiceUuids map[service.ServiceUUID]error, resultErr error) {
	return nil, nil, stacktrace.NewError("Networking sidecars aren't implemented for Kubernetes yet")
}
