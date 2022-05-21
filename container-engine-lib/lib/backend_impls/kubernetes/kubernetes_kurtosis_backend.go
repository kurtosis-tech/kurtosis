package kubernetes

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/annotation_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_key"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_value"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_value"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_port_spec_serializer"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact_expander"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact_expansion_volume"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/networking_sidecar"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	"strings"
	"time"
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
	kurtosisInternalContainerGrpcProxyPortSpecId = "grpc-proxy"

	// Port number string parsing constants
	portNumStrParsingBase = 10
	portNumStrParsingBits = 16

	netstatSuccessExitCode = 0

)

// This maps a Kubernetes pod's phase to a binary "is the pod considered running?" determiner
// Its completeness is enforced via unit test
var isPodRunningDeterminer = map[apiv1.PodPhase]bool{
	apiv1.PodPending: true,
	apiv1.PodRunning: true,
	apiv1.PodSucceeded: false,
	apiv1.PodFailed: false,
	apiv1.PodUnknown: false, //We cannot say that a pod is not running if we don't know the real state
}

// Completeness enforced via unit test
var kurtosisPortProtocolToKubernetesPortProtocolTranslator = map[port_spec.PortProtocol]apiv1.Protocol{
	port_spec.PortProtocol_TCP: apiv1.ProtocolTCP,
	port_spec.PortProtocol_UDP: apiv1.ProtocolUDP,
	port_spec.PortProtocol_SCTP: apiv1.ProtocolSCTP,
}

// TODO Remove this once we split apart the KubernetesKurtosisBackend into multiple backends (which we can only
//  do once the CLI no longer makes any calls directly to the KurtosisBackend, and instead makes all its calls through
//  the API container & engine APIs)
type cliModeArgs struct {
	// No CLI mode args needed for now
}
type apiContainerModeArgs struct {
	ownEnclaveId enclave.EnclaveID

	ownNamespaceName string

	storageClassName string

	// TODO make this more dynamic - maybe guess based on the files artifact size?
	filesArtifactExpansionVolumeSizeInMegabytes uint
}
type engineServerModeArgs struct {
	/*
		StorageClass name to be used for volumes in the cluster
		StorageClasses must be defined by a cluster administrator.
		passes this in when starting Kurtosis with Kubernetes.
	*/
	storageClassName string

	/*
		Enclave availability must be set and defined by a cluster administrator.
		The user passes this in when starting Kurtosis with Kubernetes.
	*/
	enclaveDataVolumeSizeInMegabytes uint
}

type KubernetesKurtosisBackend struct {
	kubernetesManager *kubernetes_manager.KubernetesManager

	objAttrsProvider object_attributes_provider.KubernetesObjectAttributesProvider

	cliModeArgs *cliModeArgs

	engineServerModeArgs *engineServerModeArgs

	// Will only be filled out for the API container
	apiContainerModeArgs *apiContainerModeArgs
}

// Private constructor that the other public constructors will use
func newKubernetesKurtosisBackend(
	kubernetesManager *kubernetes_manager.KubernetesManager,
	cliModeArgs *cliModeArgs,
	engineServerModeArgs *engineServerModeArgs,
	apiContainerModeArgs *apiContainerModeArgs,
) *KubernetesKurtosisBackend {
	objAttrsProvider := object_attributes_provider.GetKubernetesObjectAttributesProvider()
	return &KubernetesKurtosisBackend{
		kubernetesManager:               kubernetesManager,
		objAttrsProvider:                objAttrsProvider,
		cliModeArgs:                     cliModeArgs,
		engineServerModeArgs:            engineServerModeArgs,
		apiContainerModeArgs:            apiContainerModeArgs,
	}
}

func NewAPIContainerKubernetesKurtosisBackend(
	kubernetesManager *kubernetes_manager.KubernetesManager,
	ownEnclaveId enclave.EnclaveID,
	ownNamespaceName string,
	storageClass string,
	filesArtifactExpansionVolumeSizeInMegabytes uint,
) *KubernetesKurtosisBackend {
	modeArgs := &apiContainerModeArgs{
		ownEnclaveId:     ownEnclaveId,
		ownNamespaceName: ownNamespaceName,
		storageClassName: storageClass,
		filesArtifactExpansionVolumeSizeInMegabytes: filesArtifactExpansionVolumeSizeInMegabytes,
	}
	return newKubernetesKurtosisBackend(
		kubernetesManager,
		nil,
		nil,
		modeArgs,
	)
}

func NewEngineServerKubernetesKurtosisBackend(
	kubernetesManager *kubernetes_manager.KubernetesManager,
	storageClass string,
	enclaveDataVolumeSizeInMegabytes uint,
) *KubernetesKurtosisBackend {
	modeArgs := &engineServerModeArgs{
		storageClassName:                 storageClass,
		enclaveDataVolumeSizeInMegabytes: enclaveDataVolumeSizeInMegabytes,
	}
	return newKubernetesKurtosisBackend(
		kubernetesManager,
		nil,
		modeArgs,
		nil,
	)
}

func NewCLIModeKubernetesKurtosisBackend(
	kubernetesManager *kubernetes_manager.KubernetesManager,
) *KubernetesKurtosisBackend {
	modeArgs := &cliModeArgs{}
	return newKubernetesKurtosisBackend(
		kubernetesManager,
		modeArgs,
		nil,
		nil,
	)
}

func NewKubernetesKurtosisBackend(
	kubernetesManager *kubernetes_manager.KubernetesManager,
	// TODO Remove the necessity for these different args by splitting the KubernetesKurtosisBackend into multiple backends per consumer, e.g.
	//  APIContainerKurtosisBackend, CLIKurtosisBackend, EngineKurtosisBackend, etc. This can only happen once the CLI
	//  no longer uses the same functionality as API container, engine, etc. though
	cliModeArgs *cliModeArgs,
	engineServerModeArgs *engineServerModeArgs,
	apiContainerModeargs *apiContainerModeArgs,
) *KubernetesKurtosisBackend {
	objAttrsProvider := object_attributes_provider.GetKubernetesObjectAttributesProvider()
	return &KubernetesKurtosisBackend{
		kubernetesManager:               kubernetesManager,
		objAttrsProvider:                objAttrsProvider,
		cliModeArgs:                     cliModeArgs,
		engineServerModeArgs:            engineServerModeArgs,
		apiContainerModeArgs:            apiContainerModeargs,
	}
}

func (backend *KubernetesKurtosisBackend) PullImage(image string) error {
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

// If no expected-ports list is passed in, no validation is done and all the ports are passed back as-is
func getPrivatePortsAndValidatePortExistence(kubernetesService *apiv1.Service, expectedPortIds map[string]bool) (map[string]*port_spec.PortSpec, error) {
	portSpecsStr, found := kubernetesService.GetAnnotations()[annotation_key_consts.PortSpecsKubernetesAnnotationKey.GetString()]
	if !found {
		return nil, stacktrace.NewError(
			"Couldn't find expected port specs annotation key '%v' on the Kubernetes service",
			annotation_key_consts.PortSpecsKubernetesAnnotationKey.GetString(),
		)
	}
	privatePortSpecs, err := kubernetes_port_spec_serializer.DeserializePortSpecs(portSpecsStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred deserializing private port specs string '%v'", privatePortSpecs)
	}

	if expectedPortIds != nil && len(expectedPortIds) > 0 {
		for portId := range expectedPortIds {
			if _, found := privatePortSpecs[portId]; !found {
				return nil, stacktrace.NewError("Missing private port with ID '%v' in the private ports", portId)
			}
		}
	}
	return privatePortSpecs, nil
}

func getContainerStatusFromPod(pod *apiv1.Pod) (container_status.ContainerStatus, error) {
	// TODO Rename this; this shouldn't be called "ContainerStatus" since there's no longer a 1:1 mapping between container:kurtosis_object
	status := container_status.ContainerStatus_Stopped

	if pod != nil {
		podPhase := pod.Status.Phase
		isPodRunning, found := isPodRunningDeterminer[podPhase]
		if !found {
			// This should never happen because we enforce completeness in a unit test
			return status, stacktrace.NewError("No is-pod-running determination found for pod phase '%v' on pod '%v'; this is a bug in Kurtosis", podPhase, pod.Name)
		}
		if isPodRunning {
			status = container_status.ContainerStatus_Running
		}
	}
	return status, nil
}

func (backend *KubernetesKurtosisBackend) getEnclaveNamespaceName(ctx context.Context, enclaveId enclave.EnclaveID) (string, error) {
	// TODO This is a big janky hack that results from KubernetesKurtosisBackend containing functions for all of API containers, engines, and CLIs
	//  We want to fix this by splitting the KubernetesKurtosisBackend into a bunch of different backends, one per user, but we can only
	//  do this once the CLI no longer uses API container functionality (e.g. GetServices)
	// CLIs and engines can list namespaces so they'll be able to use the regular list-namespaces-and-find-the-one-matching-the-enclave-ID
	// API containers can't list all namespaces due to being namespaced objects themselves (can only view their own namespace, so
	// they can only check if the requested enclave matches the one they have stored
	var namespaceName string
	if backend.cliModeArgs != nil || backend.engineServerModeArgs != nil {
		matchLabels := getEnclaveMatchLabels()
		matchLabels[label_key_consts.EnclaveIDKubernetesLabelKey.GetString()] = string(enclaveId)

		namespaces, err := backend.kubernetesManager.GetNamespacesByLabels(ctx, matchLabels)
		if err != nil {
			return "", stacktrace.Propagate(err, "An error occurred getting the enclave namespace using labels '%+v'", matchLabels)
		}

		numOfNamespaces := len(namespaces.Items)
		if numOfNamespaces == 0 {
			return "", stacktrace.NewError("No namespace matching labels '%+v' was found", matchLabels)
		}
		if numOfNamespaces > 1 {
			return "", stacktrace.NewError("Expected to find only one enclave namespace matching enclave ID '%v', but found '%v'; this is a bug in Kurtosis", enclaveId, numOfNamespaces)
		}

		namespaceName = namespaces.Items[0].Name
	} else if backend.apiContainerModeArgs != nil {
		if enclaveId != backend.apiContainerModeArgs.ownEnclaveId {
			return "", stacktrace.NewError(
				"Received a request to get namespace for enclave '%v', but the Kubernetes Kurtosis backend is running in an API " +
					"container in a different enclave '%v' (so Kubernetes would throw a permission error)",
				enclaveId,
				backend.apiContainerModeArgs.ownEnclaveId,
			)
		}
		namespaceName = backend.apiContainerModeArgs.ownNamespaceName
	} else {
		return "", stacktrace.NewError("Received a request to get an enclave namespace's name, but the Kubernetes Kurtosis backend isn't in any recognized mode; this is a bug in Kurtosis")
	}

	return namespaceName, nil
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

func waitForPortAvailabilityUsingNetstat(
	kubernetesManager *kubernetes_manager.KubernetesManager,
	namespaceName string,
	podName string,
	containerName string,
	portSpec *port_spec.PortSpec,
	maxRetries uint,
	timeBetweenRetries time.Duration,
) error {
	commandStr := fmt.Sprintf(
		"[ -n \"$(netstat -anp %v | grep LISTEN | grep %v)\" ]",
		strings.ToLower(portSpec.GetProtocol().String()),
		portSpec.GetNumber(),
	)
	execCmd := []string{
		"sh",
		"-c",
		commandStr,
	}
	for i := uint(0); i < maxRetries; i++ {
		outputBuffer := &bytes.Buffer{}
		exitCode, err := kubernetesManager.RunExecCommand(namespaceName, podName, containerName, execCmd, outputBuffer)
		if err == nil {
			if exitCode == netstatSuccessExitCode {
				return nil
			}
			logrus.Debugf(
				"Netstat availability-waiting command '%v' returned without a Kubernetes error, but exited with non-%v exit code '%v' and logs:\n%v",
				commandStr,
				netstatSuccessExitCode,
				exitCode,
				outputBuffer.String(),
			)
		} else {
			logrus.Debugf(
				"Netstat availability-waiting command '%v' experienced a Kubernetes error:\n%v",
				commandStr,
				err,
			)
		}

		// Tiny optimization to not sleep if we're not going to run the loop again
		if i < maxRetries {
			time.Sleep(timeBetweenRetries)
		}
	}

	return stacktrace.NewError(
		"The port didn't become available (as measured by the command '%v') even after retrying %v times with %v between retries",
		commandStr,
		maxRetries,
		timeBetweenRetries,
	)
}