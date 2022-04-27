package connection

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/lib/kurtosis_context"
	"github.com/kurtosis-tech/stacktrace"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	"net/url"
)

const (
	grpcPortIdStr = "grpc"
)

type GatewayConnectionProvider struct {
	config            *restclient.Config
	kubernetesManager *kubernetes_manager.KubernetesManager
	providerContext   context.Context
}

func NewGatewayConnectionProvider(ctx context.Context, kubernetesConfig *restclient.Config) (*GatewayConnectionProvider, error) {
	// Necessary to set these fields for kubernetes portforwarder
	kubernetesConfig.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	kubernetesConfig.GroupVersion = &schema.GroupVersion{Group: "", Version: "v1"}

	clientSet, err := kubernetes.NewForConfig(kubernetesConfig)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get config for Kubernetes client set, instead a non nil error was returned")
	}
	kubernetesManager := kubernetes_manager.NewKubernetesManager(clientSet)

	return &GatewayConnectionProvider{
		config:            kubernetesConfig,
		kubernetesManager: kubernetesManager,
		providerContext:   ctx,
	}, nil
}

func (provider *GatewayConnectionProvider) ForEngine(engine *engine.Engine) (GatewayConnectionToKurtosis, error) {
	// Forward public GRPC ports of engine
	enginePublicGrpcPortSpec, err := port_spec.NewPortSpec(kurtosis_context.DefaultKurtosisEngineServerGrpcPortNum, port_spec.PortProtocol_TCP)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get a port-spec describing the public GRPC port of a Kurotsis engine, instead a non-nil error was returned")
	}
	enginePorts := map[string]*port_spec.PortSpec{
		grpcPortIdStr: enginePublicGrpcPortSpec,
	}
	podPortforwardEndpoint, err := provider.getEnginePodPortforwardEndpoint(engine.GetID())
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to find an api endpoint for Kubernetes portforward to engine '%v', instead a non-nil error was returned", engine.GetID())
	}
	engineConnection, err := newLocalPortToPodPortConnection(provider.config, podPortforwardEndpoint, enginePorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get a connection to engine '%v', instead a non-nil error was returned", engine.GetID())
	}
	return engineConnection, nil
}

func (provider *GatewayConnectionProvider) ForEnclaveApiContainer(enclaveInfo *kurtosis_engine_rpc_api_bindings.EnclaveInfo) (GatewayConnectionToKurtosis, error) {
	apiContainerInfo := enclaveInfo.ApiContainerInfo
	// We want the port on the kubernetes pod that tbe api container is listening on
	grpcPortUint16 := uint16(apiContainerInfo.GetGrpcPortInsideEnclave())
	apiContainerGrpcPortSpec, err := port_spec.NewPortSpec(grpcPortUint16, port_spec.PortProtocol_TCP)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get a port spec describing api container GRPC port on port number'%v', instead a non-nil error was returned", grpcPortUint16)
	}
	apiContainerPorts := map[string]*port_spec.PortSpec{
		grpcPortIdStr: apiContainerGrpcPortSpec,
	}
	enclaveId := enclaveInfo.GetEnclaveId()
	podPortforwardEndpoint, err := provider.getApiContainerPodPortforwardEndpoint(enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get an endpoint for portforwarding to the API Container in enclave '%v', instead a non-nil error was returned", enclaveId)
	}
	apiContainerConnection, err := newLocalPortToPodPortConnection(provider.config, podPortforwardEndpoint, apiContainerPorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to connect to api container in enclave '%v', instead a non-nil error was returned", enclaveId)
	}

	return apiContainerConnection, nil
}

func (provider *GatewayConnectionProvider) ForUserService(userService *service.Service) (GatewayConnectionToKurtosis, error) {
	portSpecsToForward := userService.GetMaybePublicPorts()
	serviceId := string(userService.GetRegistration().GetID())
	podPortforwardEndpoint := provider.getUserServicePodPortforwardEndpoint(serviceId)
	userServiceConnection, err := newLocalPortToPodPortConnection(provider.config, podPortforwardEndpoint, portSpecsToForward)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to connect to user service with id '%v', instead a non-nil error was returned", serviceId)
	}

	return userServiceConnection, nil
}

func (provider *GatewayConnectionProvider) getEnginePodPortforwardEndpoint(engineId string) (*url.URL, error) {
	engineLabels := map[string]string{
		label_key_consts.IDLabelKey.GetString():                   engineId,
		label_key_consts.KurtosisResourceTypeLabelKey.GetString(): label_value_consts.EngineKurtosisResourceTypeLabelValue.GetString(),
		label_key_consts.AppIDLabelKey.GetString():                label_value_consts.AppIDLabelValue.GetString(),
	}
	// Call k8s to find our engine namespace
	engineNamespaceList, err := provider.kubernetesManager.GetNamespacesByLabels(provider.providerContext, engineLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get engine namespaces with labels '%+v`, instead a non-nil error was returned", engineLabels)
	}
	if len(engineNamespaceList.Items) != 1 {
		return nil, stacktrace.NewError("Expected to find exactly 1 engine namespace with enclaveId '%v', instead found '%v'", engineId, len(engineNamespaceList.Items))
	}
	engineNamespaceName := engineNamespaceList.Items[0].Name

	// Get running Engine pods from Kubernetes
	runningEnginePodNames, err := provider.getRunningPodNamesByLabels(engineNamespaceName, engineLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get the names of running engine pods with labels '%+v', instead a non-nil error was returned", engineLabels)
	}
	if len(runningEnginePodNames) != 1 {
		return nil, stacktrace.NewError("Expected to find exactly 1 running Kurtosis Engine pod, instead found '%v'", len(runningEnginePodNames))
	}
	enginePodName := runningEnginePodNames[0]

	return provider.kubernetesManager.GetPodPortforwardEndpointUrl(engineNamespaceName, enginePodName), nil
}

func (provider *GatewayConnectionProvider) getApiContainerPodPortforwardEndpoint(enclaveId string) (*url.URL, error) {
	enclaveLabels := map[string]string{
		label_key_consts.EnclaveIDLabelKey.GetString():            enclaveId,
		label_key_consts.KurtosisResourceTypeLabelKey.GetString(): label_value_consts.EnclaveKurtosisResourceTypeLabelValue.GetString(),
		label_key_consts.AppIDLabelKey.GetString():                label_value_consts.AppIDLabelValue.GetString(),
	}
	enclaveNamespaceList, err := provider.kubernetesManager.GetNamespacesByLabels(provider.providerContext, enclaveLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get enclaves namespaces with labels '%+v`, instead a non-nil error was returned", enclaveLabels)
	}
	if len(enclaveNamespaceList.Items) != 1 {
		return nil, stacktrace.NewError("Expected to find exactly 1 enclave namespace with enclaveId '%v', instead found '%v'", enclaveId, len(enclaveNamespaceList.Items))
	}
	enclaveNamespaceName := enclaveNamespaceList.Items[0].Name

	// Get running API Container pods from Kubernetes
	apiContainerPodLabels := map[string]string{
		label_key_consts.KurtosisResourceTypeLabelKey.GetString(): label_value_consts.APIContainerKurtosisResourceTypeLabelValue.GetString(),
		label_key_consts.AppIDLabelKey.GetString():                label_key_consts.AppIDLabelKey.GetString(),
	}
	runningApiContainerPodNames, err := provider.getRunningPodNamesByLabels(enclaveNamespaceName, apiContainerPodLabels)
	if len(runningApiContainerPodNames) != 1 {
		return nil, stacktrace.NewError("Expected to find exactly 1 running API container pod in enclave '%v', instead found '%v'", enclaveId, len(runningApiContainerPodNames))
	}
	apiContainerPodName := runningApiContainerPodNames[0]

	return provider.kubernetesManager.GetPodPortforwardEndpointUrl(enclaveNamespaceName, apiContainerPodName), nil
}

// TODO Update to call Kubernetes api
func (provider *GatewayConnectionProvider) getUserServicePodPortforwardEndpoint(serviceId string) *url.URL {
	// TODO naming conventions for pods with user services and enclave namespaces
	userServiceNamespace := "enclave-namespace"
	userServicePodName := fmt.Sprintf("kurtosis-user-service--%v--pod", serviceId)
	return provider.kubernetesManager.GetPodPortforwardEndpointUrl(userServiceNamespace, userServicePodName)
}

func (provider *GatewayConnectionProvider) getRunningPodNamesByLabels(namespace string, podLabels map[string]string) ([]string, error) {
	podList, err := provider.kubernetesManager.GetPodsByLabels(provider.providerContext, namespace, podLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get pods with labels '%+v', instead a non-nil error was returned", podLabels)
	}

	runningPodNames := []string{}
	for _, pod := range podList.Items {
		podPhase := pod.Status.Phase

		if podPhase == v1.PodRunning {
			runningPodNames = append(runningPodNames, pod.Name)
		}
	}
	return runningPodNames, nil
}
