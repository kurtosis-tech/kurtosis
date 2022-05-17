package kubernetes

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"io"
	apiv1 "k8s.io/api/core/v1"
	"net"
)

/*
      *************************************************************************************************
      * See the documentation on the KurtosisBackend interface for the Kurtosis service state diagram *
      *************************************************************************************************

                                  KUBERNETES SERVICES IMPLEMENTATION

States:
- REGISTERED: a ServiceGUID has been generated, a Kubernetes Service has been created with that ServiceGUID as a label,
	the Kubernetes Service gets it selector set to look for pods with that ServiceGUID (of which there will be none at
	first), and the ClusterIP of the Service is returned to the caller.
- RUNNING: one or more Pods match the ServiceGUID selector of the Service.
- STOPPED = the Service's selectors are set to empty (this is to distinguish between "no Pods exist because they haven't
	been started yet", i.e. REGISTERED, and "no Pods exist because they were stopped", i.e. STOPPED).
- DESTROYED = no Service or Pod for the ServiceGUID exists.

State transitions:
- RegisterService: a Service is created with the ServiceGUID label (for finding later), the Service's selectors are set
	to look for Pods with the ServiceGUID, and the Service's ClusterIP is returned to the caller.
- StartService: a Pod is created with the ServiceGUID, which will hook up the Service to the Pod.
- StopServices: The Pod is destroyed (because Kubernetes doesn't have a way to stop Pods - only remove them) and the Service's
	selectors are set to empty. If we want to keep Kubernetes logs around after a Pod is destroyed, we'll need to implement
	our own logstore.
- DestroyServices: the Service is destroyed, as is any Pod that may have been started.

Implementation notes:
- Kubernetes has no concept of stopping a pod without removing it; if we want Pod logs we'll need to implement our own logstore.
- The IP that the container gets as its "own IP" is technically the IP of the Service. This *should* be fine, but we'll need
	to keep an eye out for it.
*/

// Any of these fields can be nil if they don't exist in Kubernetes, though at least
// one field will be present (else this struct won't exist)
type userServiceKubernetesResources struct {
	service *apiv1.Service

	pod *apiv1.Pod
}

func (backend *KubernetesKurtosisBackend) RegisterUserService(ctx context.Context, enclaveId enclave.EnclaveID, serviceId service.ServiceID, ) (*service.ServiceRegistration, error, ) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) StartUserService(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	guid service.ServiceGUID,
	containerImageName string,
	privatePorts map[string]*port_spec.PortSpec,
	entrypointArgs []string,
	cmdArgs []string,
	envVars map[string]string,
	filesArtifactMountDirpaths map[string]string,
) (newUserService *service.Service, resultErr error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) GetUserServices(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *service.ServiceFilters,
) (successfulUserServices map[service.ServiceGUID]*service.Service, resultError error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) GetUserServiceLogs(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *service.ServiceFilters,
	shouldFollowLogs bool,
) (successfulUserServiceLogs map[service.ServiceGUID]io.ReadCloser, erroredUserServiceGuids map[service.ServiceGUID]error, resultError error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) PauseService(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceId service.ServiceGUID) error {
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) UnpauseService(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceId service.ServiceGUID) error {
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) RunUserServiceExecCommands(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	userServiceCommands map[service.ServiceGUID][]string,
) (
	succesfulUserServiceExecResults map[service.ServiceGUID]*exec_result.ExecResult,
	erroredUserServiceGuids map[service.ServiceGUID]error,
	resultErr error,
) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) GetConnectionWithUserService(ctx context.Context, enclaveId enclave.EnclaveID, serviceGUID service.ServiceGUID) (resultConn net.Conn, resultErr error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) CopyFromUserService(ctx context.Context, enclaveId enclave.EnclaveID, serviceGuid service.ServiceGUID, srcPath string) (resultReadCloser io.ReadCloser, resultErr error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) StopUserServices(ctx context.Context, enclaveId enclave.EnclaveID, filters *service.ServiceFilters) (successfulUserServiceGuids map[service.ServiceGUID]bool, erroredUserServiceGuids map[service.ServiceGUID]error, resultErr error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) DestroyUserServices(ctx context.Context, enclaveId enclave.EnclaveID, filters *service.ServiceFilters) (successfulUserServiceGuids map[service.ServiceGUID]bool, erroredUserServiceGuids map[service.ServiceGUID]error, resultErr error) {
	//TODO implement me
	panic("implement me")
}


// ====================================================================================================
//                                     Private Helper Methods
// ====================================================================================================
func (backend *KubernetesKurtosisBackend) getMatchingUserServiceObjectsAndKubernetesResources(
	ctx context.Context,
	filters *service.ServiceFilters,
) (
	map[service.ServiceGUID]*service.Service,
	map[service.ServiceGUID]*userServiceKubernetesResources,
	error,
) {
	matchingResources, err := backend.getMatchingUserServiceKubernetesResources(ctx, filters.IDs)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting engine Kubernetes resources matching IDs: %+v", filters.IDs)
	}


}

func (backend *KubernetesKurtosisBackend) getMatchingUserServiceKubernetesResources(
	ctx context.Context,
	filters *service.ServiceFilters,
) (
	map[service.ServiceGUID]*userServiceKubernetesResources,
	error,
) {

}

func (backend *KubernetesKurtosisBackend) getUserServiceObjectsFromKubernetesResources(
	map[service.ServiceGUID]*userServiceKubernetesResources,
) (
	map[service.ServiceGUID]*service.Service,
){

}