package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_key"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_value"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_value"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/object_name_constants"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact_expander"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact_expansion_volume"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/module"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/networking_sidecar"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"io"
	apiv1 "k8s.io/api/core/v1"
	"strconv"
	"strings"
)

const (
	// The Kurtosis servers (Engine and API Container) use gRPC so MUST listen on TCP (no other protocols are supported), which also
	// means that its grpc-proxy must listen on TCP
	kurtosisServersPortProtocol = port_spec.PortProtocol_TCP

	// The ID of the GRPC port for Kurtosis-internal containers (e.g. API container, engine, modules, etc.) which will
	//  be stored in the port spec label
	kurtosisInternalContainerGrpcPortSpecId = "grpc"

	// The ID of the GRPC proxy port for Kurtosis-internal containers. This is necessary because
	// Typescript's grpc-web cannot communicate directly with GRPC ports, so Kurtosis-internal containers
	// need a proxy  that will translate grpc-web requests before they hit the main GRPC server
	kurtosisInternalContainerGrpcProxyPortSpecId = "grpcProxy"

	// Port number string parsing constants
	publicPortNumStrParsingBase = 10
	publicPortNumStrParsingBits = 16
)

// This maps a Kubernetes pod's phase to a binary "is the pod considered running?" determiner
// Its completeness is enforced via unit test
var isPodRunningDeterminer = map[apiv1.PodPhase]bool{
	apiv1.PodPending: true,
	apiv1.PodRunning: true,
	apiv1.PodSucceeded: false,
	apiv1.PodUnknown: true, //We can not say that a pod is not running if we don't know the real state
}

// Completeness enforced via unit test
var kurtosisPortProtocolToKubernetesPortProtocolTranslator = map[port_spec.PortProtocol]apiv1.Protocol{
	port_spec.PortProtocol_TCP: apiv1.ProtocolTCP,
	port_spec.PortProtocol_UDP: apiv1.ProtocolUDP,
	port_spec.PortProtocol_SCTP: apiv1.ProtocolSCTP,
}

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
	volumeSizePerEnclaveInMegabytes uint
}

func (backend *KubernetesKurtosisBackend) PullImage(image string) error {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) CreateModule(ctx context.Context, image string, enclaveId enclave.EnclaveID, id module.ModuleID, guid module.ModuleGUID, grpcPortNum uint16, envVars map[string]string) (newModule *module.Module, resultErr error) {
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

func (backend *KubernetesKurtosisBackend) CreateNetworkingSidecar(ctx context.Context, enclaveId enclave.EnclaveID, serviceGuid service.ServiceGUID) (*networking_sidecar.NetworkingSidecar, error) {
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

func (backend *KubernetesKurtosisBackend) RunFilesArtifactExpander(ctx context.Context, guid files_artifact_expander.FilesArtifactExpanderGUID, enclaveId enclave.EnclaveID, filesArtifactExpansionVolumeName files_artifact_expansion_volume.FilesArtifactExpansionVolumeName, destVolMntDirpathOnExpander string, filesArtifactFilepathRelativeToEnclaveDatadirRoot string) (*files_artifact_expander.FilesArtifactExpander, error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) DestroyFilesArtifactExpanders(ctx context.Context, filters *files_artifact_expander.FilesArtifactExpanderFilters) (successfulFilesArtifactExpanderGuids map[files_artifact_expander.FilesArtifactExpanderGUID]bool, erroredFilesArtifactExpanderGuids map[files_artifact_expander.FilesArtifactExpanderGUID]error, resultError error) {
	//TODO implement me
	panic("implement me")
}

func NewKubernetesKurtosisBackend(kubernetesManager *kubernetes_manager.KubernetesManager, volumeStorageClassName string, volumeSizePerEnclaveInMegabytes uint) *KubernetesKurtosisBackend {
	objAttrsProvider := object_attributes_provider.GetKubernetesObjectAttributesProvider()
	return &KubernetesKurtosisBackend{
		kubernetesManager:               kubernetesManager,
		objAttrsProvider:                objAttrsProvider,
		volumeStorageClassName:          volumeStorageClassName,
		volumeSizePerEnclaveInMegabytes: volumeSizePerEnclaveInMegabytes,
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

// getPublicPortSpecFromServicePort returns a port_spec representing a Kurtosis port spec for a service port in Kubernetes
func getPublicPortSpecFromServicePort(servicePort apiv1.ServicePort, portProtocol port_spec.PortProtocol) (*port_spec.PortSpec, error) {
	publicPortNumStr := strconv.FormatInt(int64(servicePort.Port), publicPortNumStrParsingBase)
	publicPortNumUint64, err := strconv.ParseUint(publicPortNumStr, publicPortNumStrParsingBase, publicPortNumStrParsingBits)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred parsing public port string '%v' using base '%v' and uint bits '%v'",
			publicPortNumStr,
			publicPortNumStrParsingBase,
			publicPortNumStrParsingBits,
		)
	}
	publicPortNum := uint16(publicPortNumUint64) // Safe to do because we pass the requisite number of bits into the parse command
	publicGrpcPort, err := port_spec.NewPortSpec(publicPortNum, portProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to create a port spec describing a public port on a Kubernetes node using number '%v' and protocol '%v', instead a non nil error was returned", publicPortNum, portProtocol)
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
					return nil, nil, stacktrace.Propagate(err, "Expected to be able to create a port spec describing a public grpc port from Kubernetes service port '%v', instead a non nil error was returned", servicePortName)
				}
				publicGrpcPort = publicGrpcPortSpec
			}
		case grpcProxyPortName:
			{
				publicGrpcProxyPortSpec, err := getPublicPortSpecFromServicePort(servicePort, kurtosisServersPortProtocol)
				if err != nil {
					return nil, nil, stacktrace.Propagate(err, "Expected to be able to create a port spec describing a public grpc proxy port from Kubernetes service port '%v', instead a non nil error was returned", servicePortName)
				}
				publicGrpcProxyPort = publicGrpcProxyPortSpec
			}
		}
	}

	if publicGrpcPort == nil || publicGrpcProxyPort == nil {
		return nil, nil, stacktrace.NewError("Expected to get public port specs from Kubernetes service ports, instead got a nil pointer")
	}

	return publicGrpcPort, publicGrpcProxyPort, nil
}

func getContainerStatusFromPod(pod *apiv1.Pod) (container_status.ContainerStatus, error) {
	status := container_status.ContainerStatus_Stopped

	if pod != nil {
		podPhase := pod.Status.Phase
		isPodRunning, found := isPodRunningDeterminer[podPhase]
		if !found {
			// This should never happen because we enforce completeness in a unit test
			return status, stacktrace.NewError("No is-running designation found for pod phase '%v'; this is a bug in Kurtosis!", podPhase)
		}
		if isPodRunning {
			status = container_status.ContainerStatus_Running
		}
	}
	return status, nil
}

func (backend *KubernetesKurtosisBackend) getEnclaveNamespace(ctx context.Context, enclaveId enclave.EnclaveID) (*apiv1.Namespace, error) {

	matchLabels := getEnclaveMatchLabels()
	matchLabels[label_key_consts.EnclaveIDKubernetesLabelKey.GetString()] = string(enclaveId)

	namespaces, err := backend.kubernetesManager.GetNamespacesByLabels(ctx, matchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the enclave namespace using labels '%+v'", matchLabels)
	}

	numOfNamespaces := len(namespaces.Items)
	if numOfNamespaces == 0 {
		return nil, stacktrace.NewError("No namespace matching labels '%+v' was found", matchLabels)
	}
	if numOfNamespaces > 1 {
		return nil, stacktrace.NewError("Expected to find only one API container namespace for API container in enclave ID '%v', but '%v' was found; this is a bug in Kurtosis", enclaveId, numOfNamespaces)
	}

	resultNamespace := &namespaces.Items[0]

	return resultNamespace, nil
}

func getEnclaveMatchLabels() map[string]string {
	matchLabels := map[string]string{
		label_key_consts.AppIDKubernetesLabelKey.GetString():                label_value_consts.AppIDKubernetesLabelValue.GetString(),
		label_key_consts.KurtosisResourceTypeKubernetesLabelKey.GetString(): label_value_consts.EnclaveKurtosisResourceTypeKubernetesLabelValue.GetString(),
	}
	return matchLabels
}

func getKubernetesServicePortsFromPrivatePortSpecs(privatePorts map[string]*port_spec.PortSpec) ([]apiv1.ServicePort, error) {
	result := []apiv1.ServicePort{}
	for portId, portSpec := range privatePorts {
		kurtosisProtocol := portSpec.GetProtocol()
		kubernetesProtocol, found := kurtosisPortProtocolToKubernetesPortProtocolTranslator[kurtosisProtocol]
		if !found {
			// Should never happen because we enforce completeness via unit test
			return nil, stacktrace.NewError("No Kubernetes port protocol was defined for Kurtosis port protocol '%v'; this is a bug in Kurtosis", kurtosisProtocol)
		}

		kubernetesPortObj := apiv1.ServicePort{
			Name:        portId,
			Protocol:    kubernetesProtocol,
			// TODO Specify this!!! Will make for a really nice user interface (e.g. "https")
			AppProtocol: nil,
			// Safe to cast because max uint16 < int32
			Port:        int32(portSpec.GetNumber()),
		}
		result = append(result, kubernetesPortObj)
	}
	return result, nil
}

func getKubernetesContainerPortsFromPrivatePortSpecs(privatePorts map[string]*port_spec.PortSpec) ([]apiv1.ContainerPort, error) {
	result := []apiv1.ContainerPort{}
	for portId, portSpec := range privatePorts {
		kurtosisProtocol := portSpec.GetProtocol()
		kubernetesProtocol, found := kurtosisPortProtocolToKubernetesPortProtocolTranslator[kurtosisProtocol]
		if !found {
			// Should never happen because we enforce completeness via unit test
			return nil, stacktrace.NewError("No Kubernetes port protocol was defined for Kurtosis port protocol '%v'; this is a bug in Kurtosis", kurtosisProtocol)
		}

		kubernetesPortObj := apiv1.ContainerPort{
			Name:          portId,
			// Safe to do because max uint16 < int32
			ContainerPort: int32(portSpec.GetNumber()),
			Protocol:      kubernetesProtocol,
		}
		result = append(result, kubernetesPortObj)
	}
	return result, nil
}

// This is a helper function that will take multiple errors, each identified by an ID, and format them together
// If no errors are returned, this function returns nil
func buildCombinedError(errorsById map[string]error, titleStr string) error {
	allErrorStrs := []string{}
	for errorId, stopErr := range errorsById {
		errorFormatStr := ">>>>>>>>>>>>> %v %v <<<<<<<<<<<<<\n" +
			"%v\n" +
			">>>>>>>>>>>>> END %v %v <<<<<<<<<<<<<"
		errorStr := fmt.Sprintf(
			errorFormatStr,
			strings.ToUpper(titleStr),
			errorId,
			stopErr.Error(),
			strings.ToUpper(titleStr),
			errorId,
		)
		allErrorStrs = append(allErrorStrs, errorStr)
	}

	if len(allErrorStrs) > 0 {
		// NOTE: This is one of the VERY rare cases where we don't want to use stacktrace.Propagate, because
		// attaching stack information for this method (which simply combines errors) just isn't useful. The
		// expected behaviour is that the caller of this function will use stacktrace.Propagate
		return errors.New(strings.Join(
			allErrorStrs,
			"\n\n",
		))
	}

	return nil
}
