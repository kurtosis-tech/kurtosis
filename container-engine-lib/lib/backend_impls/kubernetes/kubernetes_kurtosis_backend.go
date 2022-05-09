package kubernetes

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_key"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_value"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_value"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/object_name_constants"
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
	"github.com/kurtosis-tech/stacktrace"
	"io"
	apiv1 "k8s.io/api/core/v1"
	"net"
	"strconv"
)

const (
	// The Kurtosis servers (Engine and API Container) uses gRPC so MUST listen on TCP (no other protocols are supported), which also
	// means that its grpc-proxy must listen on TCP
	kurtosisServersPortProtocol = port_spec.PortProtocol_TCP

	// Port number string parsing constants
	publicPortNumStrParsingBase = 10
	publicPortNumStrParsingBits = 16

	externalServiceType = "ClusterIP"

	sentencesSeparator = ", "
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

func (backend *KubernetesKurtosisBackend) StopAPIContainers(ctx context.Context, filters *api_container.APIContainerFilters) (successfulApiContainerIds map[enclave.EnclaveID]bool, erroredApiContainerIds map[enclave.EnclaveID]error, resultErr error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) DestroyAPIContainers(ctx context.Context, filters *api_container.APIContainerFilters) (successfulApiContainerIds map[enclave.EnclaveID]bool, erroredApiContainerIds map[enclave.EnclaveID]error, resultErr error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) CreateModule(ctx context.Context, image string, enclaveId enclave.EnclaveID, id module.ModuleID, guid module.ModuleGUID, ipAddr net.IP, grpcPortNum uint16, envVars map[string]string) (newModule *module.Module, resultErr error) {
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

func (backend *KubernetesKurtosisBackend) CreateUserService(ctx context.Context, id service.ServiceID, guid service.ServiceGUID, containerImageName string, enclaveId enclave.EnclaveID, ipAddr net.IP, privatePorts map[string]*port_spec.PortSpec, entrypointArgs []string, cmdArgs []string, envVars map[string]string, filesArtifactMountDirpaths map[string]string) (newUserService *service.Service, resultErr error) {
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

func (backend *KubernetesKurtosisBackend) RunFilesArtifactExpander(ctx context.Context, guid files_artifact_expander.FilesArtifactExpanderGUID, enclaveId enclave.EnclaveID, filesArtifactExpansionVolumeName files_artifact_expansion_volume.FilesArtifactExpansionVolumeName, destVolMntDirpathOnExpander string, filesArtifactFilepathRelativeToEnclaveDatadirRoot string, ipAddr net.IP) (*files_artifact_expander.FilesArtifactExpander, error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) DestroyFilesArtifactExpanders(ctx context.Context, filters *files_artifact_expander.FilesArtifactExpanderFilters) (successfulFilesArtifactExpanderGuids map[files_artifact_expander.FilesArtifactExpanderGUID]bool, erroredFilesArtifactExpanderGuids map[files_artifact_expander.FilesArtifactExpanderGUID]error, resultError error) {
	//TODO implement me
	panic("implement me")
}

func NewKubernetesKurtosisBackend(kubernetesManager *kubernetes_manager.KubernetesManager, volumeStorageClassName string, volumeSizePerEnclaveInGigabytes int) *KubernetesKurtosisBackend {
	objAttrsProvider := object_attributes_provider.GetKubernetesObjectAttributesProvider()
	return &KubernetesKurtosisBackend{
		kubernetesManager: kubernetesManager,
		objAttrsProvider:  objAttrsProvider,
		volumeStorageClassName: volumeStorageClassName,
		volumeSizePerEnclaveInGigabytes: volumeSizePerEnclaveInGigabytes,
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

// getPublicPortSpecFromServicePort returns a port_spec representing a kurtosis port spec for a service port in kubernetes
func getPublicPortSpecFromServicePort(servicePort apiv1.ServicePort, portProtocol port_spec.PortProtocol) (*port_spec.PortSpec, error) {
	publicPortNumStr := strconv.FormatInt(int64(servicePort.Port), 10)
	publicPortNumUint64, err := strconv.ParseUint(publicPortNumStr, publicPortNumStrParsingBase, publicPortNumStrParsingBits)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred parsing engine server public port string '%v' using base '%v' and uint bits '%v'",
			publicPortNumStr,
			publicPortNumStrParsingBase,
			publicPortNumStrParsingBits,
		)
	}
	publicPortNum := uint16(publicPortNumUint64) // Safe to do because we pass the requisite number of bits into the parse command
	publicGrpcPort, err := port_spec.NewPortSpec(publicPortNum, portProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to create a port spec describing a public port on a kubernetes node using number '%v' and protocol '%v', instead a non nil error was returned", publicPortNum, portProtocol)
	}

	return publicGrpcPort, nil
}

func getGrpcAndGrpcProxyPortSpecsFromServicePorts(servicePorts []apiv1.ServicePort) (resultGrpcPortSpec *port_spec.PortSpec, resultGrpcProxyPortSpec *port_spec.PortSpec, resultErr error) {
	var publicGrpcPort *port_spec.PortSpec
	var publicGrpcProxyPort *port_spec.PortSpec
	grpcPortName := object_name_constants.KurtosisInternalContainerGrpcPortName.GetString()
	grpcProxyPortName := object_name_constants.KurtosisInternalContainerGrpcProxyPortName.GetString()

	for _, servicePort := range servicePorts {
		servicePortName := servicePort.Name
		switch servicePortName {
		case grpcPortName:
			{
				publicGrpcPortSpec, err := getPublicPortSpecFromServicePort(servicePort, kurtosisServersPortProtocol)
				if err != nil {
					return nil, nil, stacktrace.Propagate(err, "Expected to be able to create a port spec describing a public grpc port from kubernetes service port '%v', instead a non nil error was returned", servicePortName)
				}
				publicGrpcPort = publicGrpcPortSpec
			}
		case grpcProxyPortName:
			{
				publicGrpcProxyPortSpec, err := getPublicPortSpecFromServicePort(servicePort, kurtosisServersPortProtocol)
				if err != nil {
					return nil, nil, stacktrace.Propagate(err, "Expected to be able to create a port spec describing a public grpc proxy port from kubernetes service port '%v', instead a non nil error was returned", servicePortName)
				}
				publicGrpcProxyPort = publicGrpcProxyPortSpec
			}
		}
	}

	if publicGrpcPort == nil || publicGrpcProxyPort == nil {
		return nil, nil, stacktrace.NewError("Expected to get public port specs from kubernetes service ports, instead got a nil pointer")
	}

	return publicGrpcPort, publicGrpcProxyPort, nil
}

func (backend *KubernetesKurtosisBackend) getEnclaveNamespace(ctx context.Context, enclaveId enclave.EnclaveID) (*apiv1.Namespace, error) {

	matchLabels := getEnclaveMatchLabels(enclaveId)

	namespaces, err := backend.kubernetesManager.GetNamespacesByLabels(ctx, matchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the enclave namespace using labels '%+v'", matchLabels)
	}

	numOfNamespaces := len(namespaces.Items)
	if numOfNamespaces == 0 {
		return nil, stacktrace.NewError("No namespace matching labels '%+v' was found", matchLabels)
	}
	if numOfNamespaces > 1 {
		return nil, stacktrace.NewError("Expected to find only one api container namespace for api container in enclave ID '%v', but '%v' was found; this is a bug in Kurtosis", enclaveId, numOfNamespaces)
	}

	resultNamespace := &namespaces.Items[0]

	return resultNamespace, nil
}

func (backend *KubernetesKurtosisBackend) getEnclaveDataPersistentVolumeClaim(ctx context.Context, enclaveNamespaceName string, enclaveId enclave.EnclaveID) (*apiv1.PersistentVolumeClaim, error) {
	matchLabels :=  map[string]string{
		label_key_consts.AppIDLabelKey.GetString():     label_value_consts.AppIDLabelValue.GetString(),
		label_key_consts.VolumeTypeLabelKey.GetString(): label_value_consts.EnclaveDataVolumeTypeLabelValue.GetString(),
		label_key_consts.EnclaveIDLabelKey.GetString(): string(enclaveId),
	}

	persistentVolumeClaims, err := backend.kubernetesManager.GetPersistentVolumeClaimsByLabels(ctx, enclaveNamespaceName, matchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the enclave persistent volume claim using labels '%+v'", matchLabels)
	}

	numOfPersistentVolumeClaims := len(persistentVolumeClaims.Items)
	if numOfPersistentVolumeClaims == 0 {
		return nil, stacktrace.NewError("No persistent volume claim matching labels '%+v' was found", matchLabels)
	}
	if numOfPersistentVolumeClaims > 1 {
		return nil, stacktrace.NewError("Expected to find only one enclave data persistent volume claim for enclave ID '%v', but '%v' was found; this is a bug in Kurtosis", enclaveId, numOfPersistentVolumeClaims)
	}

	resultPersistentVolumeClaim := &persistentVolumeClaims.Items[0]

	return resultPersistentVolumeClaim, nil
}

func getEnclaveMatchLabels(enclaveId enclave.EnclaveID) map[string]string {
	enclaveIdStr := string(enclaveId)

	matchLabels := map[string]string{
		label_key_consts.AppIDLabelKey.GetString():     label_value_consts.AppIDLabelValue.GetString(),
		label_key_consts.EnclaveIDLabelKey.GetString(): enclaveIdStr,
	}

	return matchLabels
}

