package kubernetes

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_key"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_value"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_value"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/api_container"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact_expander"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact_expansion_volume"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/module"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/networking_sidecar"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/wait_for_availability_http_methods"
	"io"
	"net"
)

type KubernetesKurtosisBackend struct {
	kubernetesManager *kubernetes_manager.KubernetesManager

	objAttrsProvider object_attributes_provider.KubernetesObjectAttributesProvider

	/*
		StorageClass name to be used for volumes in the cluster
		StorageClasses must be defined by a cluster administrator.
		passes this in when starting Kurtosis with Kubernetes.
	*/
	volumeStorageClassName string
	/*
		Enclave availability must be set and defined by a cluster administrator.
		The user passes this in when starting Kurtosis with Kubernetes.
	 */
	volumeSizePerEnclaveInGigabytes int
}

func (backend *KubernetesKurtosisBackend) PullImage(image string) error {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) CreateAPIContainer(
	ctx context.Context,
	image string,
	enclaveId enclave.EnclaveID,
	ipAddr net.IP,
	grpcPortNum uint16,
	grpcProxyPortNum uint16,
	enclaveDataDirpathOnHostMachine string,
	enclaveDataVolumeDirpath string,
	envVars map[string]string,
) (*api_container.APIContainer, error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) GetAPIContainers(ctx context.Context, filters *api_container.APIContainerFilters) (map[enclave.EnclaveID]*api_container.APIContainer, error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) StopAPIContainers(ctx context.Context, filters *api_container.APIContainerFilters) (successfulApiContainerIds map[enclave.EnclaveID]bool, erroredApiContainerIds map[enclave.EnclaveID]error, resultErr error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) DestroyAPIContainers(ctx context.Context, filters *api_container.APIContainerFilters) (successfulApiContainerIds map[enclave.EnclaveID]bool, erroredApiContainerIds map[enclave.EnclaveID]error, resultErr error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) CreateModule(ctx context.Context, image string, enclaveId enclave.EnclaveID, id module.ModuleID, guid module.ModuleGUID, ipAddr net.IP, grpcPortNum uint16, enclaveDataDirpathOnHostMachine string, envVars map[string]string) (newModule *module.Module, resultErr error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) GetModules(ctx context.Context, filters *module.ModuleFilters) (map[module.ModuleGUID]*module.Module, error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) GetModuleLogs(ctx context.Context, filters *module.ModuleFilters, shouldFollowLogs bool) (successfulModuleLogs map[module.ModuleGUID]io.ReadCloser, erroredModuleGuids map[module.ModuleGUID]error, resultError error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) StopModules(ctx context.Context, filters *module.ModuleFilters) (successfulModuleIds map[module.ModuleGUID]bool, erroredModuleIds map[module.ModuleGUID]error, resultErr error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) DestroyModules(ctx context.Context, filters *module.ModuleFilters) (successfulModuleIds map[module.ModuleGUID]bool, erroredModuleIds map[module.ModuleGUID]error, resultErr error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) CreateUserService(ctx context.Context, id service.ServiceID, guid service.ServiceGUID, containerImageName string, enclaveId enclave.EnclaveID, ipAddr net.IP, privatePorts map[string]*port_spec.PortSpec, entrypointArgs []string, cmdArgs []string, envVars map[string]string, enclaveDataDirpathOnHostMachine string, enclaveDataDirpathOnServiceContainer string, filesArtifactMountDirpaths map[string]string) (newUserService *service.Service, resultErr error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) GetUserServices(ctx context.Context, filters *service.ServiceFilters) (successfulUserServices map[service.ServiceGUID]*service.Service, resultError error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) GetUserServiceLogs(ctx context.Context, filters *service.ServiceFilters, shouldFollowLogs bool) (successfulUserServiceLogs map[service.ServiceGUID]io.ReadCloser, erroredUserServiceGuids map[service.ServiceGUID]error, resultError error) {
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

func (backend *KubernetesKurtosisBackend) RunUserServiceExecCommands(ctx context.Context, enclaveId enclave.EnclaveID, userServiceCommands map[service.ServiceGUID][]string) (succesfulUserServiceExecResults map[service.ServiceGUID]*exec_result.ExecResult, erroredUserServiceGuids map[service.ServiceGUID]error, resultErr error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) WaitForUserServiceHttpEndpointAvailability(ctx context.Context, enclaveId enclave.EnclaveID, serviceGUID service.ServiceGUID, httpMethod wait_for_availability_http_methods.WaitForAvailabilityHttpMethod, port uint32, path string, requestBody string, expectedResponseBody string, initialDelayMilliseconds uint32, retries uint32, retriesDelayMilliseconds uint32) (resultErr error) {
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

func (backend *KubernetesKurtosisBackend) StopUserServices(ctx context.Context, filters *service.ServiceFilters) (successfulUserServiceGuids map[service.ServiceGUID]bool, erroredUserServiceGuids map[service.ServiceGUID]error, resultErr error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) DestroyUserServices(ctx context.Context, filters *service.ServiceFilters) (successfulUserServiceGuids map[service.ServiceGUID]bool, erroredUserServiceGuids map[service.ServiceGUID]error, resultErr error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) CreateNetworkingSidecar(ctx context.Context, enclaveId enclave.EnclaveID, serviceGuid service.ServiceGUID, ipAddr net.IP) (*networking_sidecar.NetworkingSidecar, error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) GetNetworkingSidecars(ctx context.Context, filters *networking_sidecar.NetworkingSidecarFilters) (map[service.ServiceGUID]*networking_sidecar.NetworkingSidecar, error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) RunNetworkingSidecarExecCommands(ctx context.Context, enclaveId enclave.EnclaveID, networkingSidecarsCommands map[service.ServiceGUID][]string) (successfulNetworkingSidecarExecResults map[service.ServiceGUID]*exec_result.ExecResult, erroredUserServiceGuids map[service.ServiceGUID]error, resultErr error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) StopNetworkingSidecars(ctx context.Context, filters *networking_sidecar.NetworkingSidecarFilters) (successfulUserServiceGuids map[service.ServiceGUID]bool, erroredUserServiceGuids map[service.ServiceGUID]error, resultErr error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) DestroyNetworkingSidecars(ctx context.Context, filters *networking_sidecar.NetworkingSidecarFilters) (successfulUserServiceGuids map[service.ServiceGUID]bool, erroredUserServiceGuids map[service.ServiceGUID]error, resultErr error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) CreateFilesArtifactExpansionVolume(ctx context.Context, enclaveId enclave.EnclaveID, serviceGuid service.ServiceGUID, filesArtifactId service.FilesArtifactID) (*files_artifact_expansion_volume.FilesArtifactExpansionVolume, error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) DestroyFilesArtifactExpansionVolumes(ctx context.Context, filters *files_artifact_expansion_volume.FilesArtifactExpansionVolumeFilters) (successfulFileArtifactExpansionVolumeNames map[files_artifact_expansion_volume.FilesArtifactExpansionVolumeName]bool, erroredFileArtifactExpansionVolumeNames map[files_artifact_expansion_volume.FilesArtifactExpansionVolumeName]error, resultErr error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) RunFilesArtifactExpander(ctx context.Context, guid files_artifact_expander.FilesArtifactExpanderGUID, enclaveId enclave.EnclaveID, filesArtifactExpansionVolumeName files_artifact_expansion_volume.FilesArtifactExpansionVolumeName, enclaveDataDirpathOnHostMachine string, destVolMntDirpathOnExpander string, filesArtifactFilepathRelativeToEnclaveDatadirRoot string, ipAddr net.IP) (*files_artifact_expander.FilesArtifactExpander, error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) DestroyFilesArtifactExpanders(ctx context.Context, filters *files_artifact_expander.FilesArtifactExpanderFilters) (successfulFilesArtifactExpanderGuids map[files_artifact_expander.FilesArtifactExpanderGUID]bool, erroredFilesArtifactExpanderGuids map[files_artifact_expander.FilesArtifactExpanderGUID]error, resultError error) {
	//TODO implement me
	panic("implement me")
}

func NewKubernetesKurtosisBackend(kubernetesManager *kubernetes_manager.KubernetesManager, volumeStorageClassName string) *KubernetesKurtosisBackend {
	objAttrsProvider := object_attributes_provider.GetKubernetesObjectAttributesProvider()
	return &KubernetesKurtosisBackend{
		kubernetesManager: kubernetesManager,
		objAttrsProvider:  objAttrsProvider,
		volumeStorageClassName: volumeStorageClassName,
	}
}

func getStringMapFromLabelMap(labelMap map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue) map[string]string {
	strMap := map[string]string{}
	for labelKey, labelValue := range labelMap {
		strMap[labelKey.GetString()] = labelValue.GetString()
	}
	return strMap
}

func getStringMapFromAnnotationMap(labelMap map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue) map[string]string {
	strMap := map[string]string{}
	for labelKey, labelValue := range labelMap {
		strMap[labelKey.GetString()] = labelValue.GetString()
	}
	return strMap
}