package user_services_functions

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_value"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	"net"
)

const (
	// Kubernetes doesn't allow us to create services without ports exposed, but we might not have ports in the following situations:
	//  1) we've registered a service but haven't started a container yet (so ports are yet to come)
	//  2) we've started a container that doesn't listen on any ports
	// In these cases, we use these notional unbound ports
	unboundPortName   = "nonexistent-port"
	unboundPortNumber = 1
)

// Kubernetes doesn't provide public IP or port information; this is instead handled by the Kurtosis gateway that the user uses
// to connect to Kubernetes
var servicePublicIp net.IP = nil
var servicePublicPorts map[string]*port_spec.PortSpec = nil

func RegisterUserService(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceId service.ServiceID,
	cliModeArgs *shared_helpers.CliModeArgs,
	apiContainerModeArgs *shared_helpers.ApiContainerModeArgs,
	engineServerModeArgs *shared_helpers.EngineServerModeArgs,
	kubernetesManager *kubernetes_manager.KubernetesManager) (*service.ServiceRegistration, error) {
	namespaceName, err := shared_helpers.GetEnclaveNamespaceName(ctx, enclaveId, cliModeArgs, apiContainerModeArgs, engineServerModeArgs, kubernetesManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting namespace name for enclave '%v'", enclaveId)
	}

	serviceGuidStr, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred generating a UUID to use for the service GUID")
	}
	serviceGuid := service.ServiceGUID(serviceGuidStr)

	objectAttributesProvider := object_attributes_provider.GetKubernetesObjectAttributesProvider()
	enclaveObjAttributesProvider := objectAttributesProvider.ForEnclave(enclaveId)

	serviceAttributes, err := enclaveObjAttributesProvider.ForUserServiceService(serviceGuid, serviceId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting attributes for the Kubernetes service for user service '%v'", serviceId)
	}

	serviceNameStr := serviceAttributes.GetName().GetString()

	serviceLabelsStrs := shared_helpers.GetStringMapFromLabelMap(serviceAttributes.GetLabels())
	serviceAnnotationsStrs := shared_helpers.GetStringMapFromAnnotationMap(serviceAttributes.GetAnnotations())

	// Set up the labels that the pod will match (i.e. the labels of the pod-to-be)
	// WARNING: We *cannot* use the labels of the Service itself because we're not guaranteed that the labels
	//  between the two will be identical!
	serviceGuidLabelValue, err := kubernetes_label_value.CreateNewKubernetesLabelValue(string(serviceGuid))
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Kubernetes pod match label value for the service GUID '%v'", serviceGuid)
	}
	enclaveIdLabelValue, err := kubernetes_label_value.CreateNewKubernetesLabelValue(string(enclaveId))
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Kubernetes pod match label value for the enclave ID '%v'", enclaveId)
	}
	matchedPodLabels := map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue{
		label_key_consts.AppIDKubernetesLabelKey:     label_value_consts.AppIDKubernetesLabelValue,
		label_key_consts.EnclaveIDKubernetesLabelKey: enclaveIdLabelValue,
		label_key_consts.GUIDKubernetesLabelKey:      serviceGuidLabelValue,
	}
	matchedPodLabelStrs := shared_helpers.GetStringMapFromLabelMap(matchedPodLabels)

	// Kubernetes doesn't allow us to create services without any ports, so we need to set this to a notional value
	// until the user calls StartService
	notionalServicePorts := []apiv1.ServicePort{
		{
			Name: unboundPortName,
			Port: unboundPortNumber,
		},
	}

	createdService, err := kubernetesManager.CreateService(
		ctx,
		namespaceName,
		serviceNameStr,
		serviceLabelsStrs,
		serviceAnnotationsStrs,
		matchedPodLabelStrs,
		apiv1.ServiceTypeClusterIP,
		notionalServicePorts,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating Kubernetes service in enclave '%v' with ID '%v'", enclaveId, serviceId)
	}
	shouldDeleteService := true
	defer func() {
		if shouldDeleteService {
			if err := kubernetesManager.RemoveService(ctx, createdService); err != nil {
				logrus.Errorf("Registering service '%v' didn't complete successfully so we tried to remove the Kubernetes service we created but doing so threw an error:\n%v", serviceId, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to remove service '%v' in namespace '%v' manually!!!", createdService.Name, namespaceName)
			}
		}
	}()

	kubernetesResources := map[service.ServiceGUID]*shared_helpers.UserServiceKubernetesResources{
		serviceGuid: {
			Service: createdService,
			Pod:     nil, // No pod yet
		},
	}

	convertedObjects, err := shared_helpers.GetUserServiceObjectsFromKubernetesResources(enclaveId, kubernetesResources)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting a service registration object from Kubernetes service")
	}
	objectsAndResources, found := convertedObjects[serviceGuid]
	if !found {
		return nil, stacktrace.NewError(
			"Successfully converted the Kubernetes service representing registered service with GUID '%v' to a "+
				"Kurtosis object, but couldn't find that key in the resulting map; this is a bug in Kurtosis",
			serviceGuid,
		)
	}
	serviceRegistration := objectsAndResources.ServiceRegistration

	shouldDeleteService = false
	return serviceRegistration, nil
}