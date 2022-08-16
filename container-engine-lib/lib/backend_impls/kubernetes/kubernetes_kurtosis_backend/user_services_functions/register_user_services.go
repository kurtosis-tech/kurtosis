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
	"github.com/kurtosis-tech/container-engine-lib/lib/operation_parallelizer"
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

func RegisterUserServices(
	ctx context.Context,
	enclaveID enclave.EnclaveID,
	serviceIDs map[service.ServiceID]bool,
	cliModeArgs *shared_helpers.CliModeArgs,
	apiContainerModeArgs *shared_helpers.ApiContainerModeArgs,
	engineServerModeArgs *shared_helpers.EngineServerModeArgs,
	kubernetesManager *kubernetes_manager.KubernetesManager) (map[service.ServiceID]*service.ServiceRegistration, map[service.ServiceID]error, error) {
	successfulRegistrations := map[service.ServiceID]*service.ServiceRegistration{}
	failedRegistrations := map[service.ServiceID]error{}

	// If no services were passed in to register, return empty maps
	if len(serviceIDs) == 0 {
		return successfulRegistrations, failedRegistrations, nil
	}

	namespaceName, err := shared_helpers.GetEnclaveNamespaceName(ctx, enclaveID, cliModeArgs, apiContainerModeArgs, engineServerModeArgs, kubernetesManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting namespace name for enclave '%v'", enclaveID)
	}

	objectAttributesProvider := object_attributes_provider.GetKubernetesObjectAttributesProvider()
	enclaveObjAttributesProvider := objectAttributesProvider.ForEnclave(enclaveID)

	registerServiceOperations := map[operation_parallelizer.OperationID]operation_parallelizer.Operation{}
	for serviceID, _ := range serviceIDs {
		registerServiceOperations[operation_parallelizer.OperationID(serviceID)] = createRegisterUserServiceOperation(
			ctx,
			enclaveID,
			serviceID,
			namespaceName,
			enclaveObjAttributesProvider,
			kubernetesManager)
	}

	successfulOperations, failedOperations := operation_parallelizer.RunOperationsInParallel(registerServiceOperations)

	for id, data := range successfulOperations {
		serviceID := service.ServiceID(id)
		serviceRegistration, ok := data.(*service.ServiceRegistration)
		if !ok {
			return nil, nil, stacktrace.NewError(
				"An error occurred downcasting data returned from the register user service operation for service with id: %v." +
					"This is a Kurtosis bug. Make sure the desired type is actually being returned in the corresponding Operation.", serviceID)
		}
		successfulRegistrations[serviceID] = serviceRegistration
	}

	for opID, err := range failedOperations {
		failedRegistrations[service.ServiceID(opID)] = err
	}

	return successfulRegistrations, failedRegistrations, nil
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================
func createRegisterUserServiceOperation(
	ctx context.Context,
	enclaveID enclave.EnclaveID,
	serviceID service.ServiceID,
	namespaceName string,
	enclaveObjAttributesProvider object_attributes_provider.KubernetesEnclaveObjectAttributesProvider,
	kubernetesManager *kubernetes_manager.KubernetesManager) operation_parallelizer.Operation {
	return func() (interface{}, error) {
		serviceGuidStr, err := uuid_generator.GenerateUUIDString()
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred generating a UUID to use for the service GUID")
		}
		serviceGuid := service.ServiceGUID(serviceGuidStr)
		serviceAttributes, err := enclaveObjAttributesProvider.ForUserServiceService(serviceGuid, serviceID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting attributes for the Kubernetes service for user service '%v'", serviceID)
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
		enclaveIdLabelValue, err := kubernetes_label_value.CreateNewKubernetesLabelValue(string(enclaveID))
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating a Kubernetes pod match label value for the enclave ID '%v'", enclaveID)
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
			return nil, stacktrace.Propagate(err, "An error occurred creating Kubernetes service in enclave '%v' with ID '%v'", enclaveID, serviceID)
		}
		shouldDeleteService := true
		defer func() {
			if shouldDeleteService {
				if err := kubernetesManager.RemoveService(ctx, createdService); err != nil {
					logrus.Errorf("Registering service '%v' didn't complete successfully so we tried to remove the Kubernetes service we created but doing so threw an error:\n%v", serviceID, err)
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

		convertedObjects, err := shared_helpers.GetUserServiceObjectsFromKubernetesResources(enclaveID, kubernetesResources)
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

		shouldDeleteService = false
		return objectsAndResources.ServiceRegistration, nil
	}
}