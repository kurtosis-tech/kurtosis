package kubernetes

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager/consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/object_name_constants"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/api_container"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"net"
	"strings"
)

const (
	kurtosisApiContainerContainerName = "kurtosis-core-api"
)

// ====================================================================================================
//                                     API Container CRUD Methods
// ====================================================================================================

func (backend *KubernetesKurtosisBackend) CreateAPIContainer(
	ctx context.Context,
	image string,
	enclaveId enclave.EnclaveID,
	ipAddr net.IP, // TODO REMOVE THIS ONCE WE FIX THE STATIC IP PROBLEM!!
	grpcPortNum uint16,
	grpcProxyPortNum uint16, // TODO remove when we switch fully to enclave data volume
	enclaveDataVolumeDirpath string,
	envVars map[string]string,
) (
	*api_container.APIContainer,
	error,
) {

	//TODO This validation is the same for Docker and for Kubernetes because we are using kurtBackend
	//TODO we could move this to a top layer for validations, perhaps

	// Verify no API container already exists in the enclave
	apiContainersInEnclaveFilters := &api_container.APIContainerFilters{
		EnclaveIDs: map[enclave.EnclaveID]bool{
			enclaveId: true,
		},
	}
	preexistingApiContainersInEnclave, err := backend.GetAPIContainers(ctx, apiContainersInEnclaveFilters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred checking if API containers already exist in enclave '%v'", enclaveId)
	}
	if len(preexistingApiContainersInEnclave) > 0 {
		return nil, stacktrace.NewError("Found existing API container(s) in enclave '%v'; cannot start a new one", enclaveId)
	}

	privateGrpcPortSpec, err := port_spec.NewPortSpec(grpcPortNum, kurtosisServersPortProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the api container's private grpc port spec object using number '%v' and protocol '%v'",
			grpcPortNum,
			kurtosisServersPortProtocol.String(),
		)
	}
	privateGrpcProxyPortSpec, err := port_spec.NewPortSpec(grpcProxyPortNum, kurtosisServersPortProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the api container's private grpc proxy port spec object using number '%v' and protocol '%v'",
			grpcProxyPortNum,
			kurtosisServersPortProtocol.String(),
		)
	}

	enclaveAttributesProvider, err := backend.objAttrsProvider.ForEnclave(enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err,"An error occurred getting the enclave attributes provider using enclave ID '%v'", enclaveId)
	}

	apiContainerAttributesProvider, err := enclaveAttributesProvider.ForApiContainer()
	if err != nil {
		return nil, stacktrace.Propagate(err,"An error occurred getting the api container attributes provider using enclave ID '%v'", enclaveId)
	}

	enclaveNamespace, err := backend.getEnclaveNamespace(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave namespace for enclave with ID '%v'", enclaveId)
	}
	enclaveNamespaceName := enclaveNamespace.GetName()

	//Create api container role based resources and get the api container service account name that was created in the enclave namespace
	apiContainerServiceAccountName, err := backend.createApiContainerRoleBasedResources(ctx, enclaveNamespaceName, apiContainerAttributesProvider)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating api container service account, roles and role bindings for api container for enclave with ID '%v' in namespace '%v'", enclaveId, enclaveNamespaceName)
	}
	shouldRemoveAllApiContainerRoleBasedResources := true
	defer func(){
		if shouldRemoveAllApiContainerRoleBasedResources {
			enclaveIds := map[enclave.EnclaveID]bool {
				enclaveId: true,
			}
			if _, resultErroredApiContainerIds := backend.removeApiContainerRoleBasedResources(ctx, enclaveNamespaceName, enclaveIds); len(resultErroredApiContainerIds) > 0{
				logrus.Errorf("Creating the api container didn't complete successfully, so we tried to delete the api container Kubernetes role based resources that we created for enclave with ID '%v' but an error was thrown:\n%v", enclaveId, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove Kubernetes role based resources for api container in enclave with ID '%v' and namespace '%v'!!!!!!!", enclaveId, enclaveNamespaceName)
			}
		}
	}()

	// Get Pod Attributes
	apiContainerPodAttributes, err := apiContainerAttributesProvider.ForApiContainerPod()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Expected to be able to get attributes for a Kubernetes pod for api container in enclave with id '%v', instead got a non-nil error",
			enclaveId,
		)
	}
	apiContainerPodName := apiContainerPodAttributes.GetName().GetString()
	apiContainerPodLabels := getStringMapFromLabelMap(apiContainerPodAttributes.GetLabels())
	apiContainerPodAnnotations := getStringMapFromAnnotationMap(apiContainerPodAttributes.GetAnnotations())

	enclaveDataPersistentVolumeClaim, err := backend.getEnclaveDataPersistentVolumeClaim(ctx, enclaveNamespaceName, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the enclave data persistent volume claim for enclave '%v' in namespace '%v'", enclaveId, enclaveNamespaceName)
	}

	grpcPortInt32 := int32(grpcPortNum)
	grpcProxyPortInt32 := int32(grpcProxyPortNum)

	containerPorts := []apiv1.ContainerPort{
		{
			Name:          object_name_constants.KurtosisInternalContainerGrpcPortName.GetString(),
			Protocol:      kurtosisInternalContainerGrpcPortProtocol,
			ContainerPort: grpcPortInt32,
		},
		{
			Name:          object_name_constants.KurtosisInternalContainerGrpcProxyPortName.GetString(),
			Protocol:      kurtosisInternalContainerGrpcProxyPortProtocol,
			ContainerPort: grpcProxyPortInt32,
		},
	}

	apiContainerContainers, apiContainerVolumes := getApiContainerContainersAndVolumes(image, containerPorts, envVars, enclaveDataPersistentVolumeClaim, enclaveDataVolumeDirpath)

	// Create pods with api container containers and volumes in Kubernetes
	if _, err = backend.kubernetesManager.CreatePod(ctx, enclaveNamespaceName, apiContainerPodName, apiContainerPodLabels, apiContainerPodAnnotations, apiContainerContainers, apiContainerVolumes, apiContainerServiceAccountName); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the pod with name '%s' in namespace '%s' with image '%s'", apiContainerPodName, enclaveNamespaceName, image)
	}
	var shouldRemovePod = true
	defer func() {
		if shouldRemovePod {
			if err := backend.kubernetesManager.RemovePod(ctx, enclaveNamespaceName, apiContainerPodName); err != nil {
				logrus.Errorf("Creating the api container didn't complete successfully, so we tried to delete Kubernetes pod '%v' that we created but an error was thrown:\n%v", apiContainerPodName, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove Kubernetes pod with name '%v'!!!!!!!", apiContainerPodName)
			}
		}
	}()

	// Get Service Attributes
	apiContainerServiceAttributes, err := apiContainerAttributesProvider.ForApiContainerService(
		kurtosisInternalContainerGrpcPortSpecId,
		privateGrpcPortSpec,
		kurtosisInternalContainerGrpcProxyPortSpecId,
		privateGrpcProxyPortSpec)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting the api container service attributes using private grpc port spec '%+v', and "+
				"private grpc proxy port spec '%+v'",
			privateGrpcPortSpec,
			privateGrpcProxyPortSpec,
		)
	}
	apiContainerServiceName := apiContainerServiceAttributes.GetName().GetString()
	apiContainerServiceLabels := getStringMapFromLabelMap(apiContainerServiceAttributes.GetLabels())
	apiContainerServiceAnnotations := getStringMapFromAnnotationMap(apiContainerServiceAttributes.GetAnnotations())

	// Define service ports. These hook up to ports on the containers running in the api container pod
	// Kubernetes will assign a public port number to them
	servicePorts := []apiv1.ServicePort{
		{
			Name:     object_name_constants.KurtosisInternalContainerGrpcPortName.GetString(),
			Protocol: kurtosisInternalContainerGrpcPortProtocol,
			Port:     grpcPortInt32,
		},
		{
			Name:     object_name_constants.KurtosisInternalContainerGrpcProxyPortName.GetString(),
			Protocol: kurtosisInternalContainerGrpcProxyPortProtocol,
			Port:     grpcProxyPortInt32,
		},
	}

	// Create Service
	service, err := backend.kubernetesManager.CreateService(ctx, enclaveNamespaceName, apiContainerServiceName, apiContainerServiceLabels, apiContainerServiceAnnotations, apiContainerPodLabels, externalServiceType, servicePorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the service with name '%s' in namespace '%s' with ports '%v' and '%v'", apiContainerServiceName, enclaveNamespaceName, grpcPortInt32, grpcProxyPortInt32)
	}
	var shouldRemoveService = true
	defer func() {
		if shouldRemoveService {
			if err := backend.kubernetesManager.RemoveService(ctx, enclaveNamespaceName, apiContainerServiceName); err != nil {
				logrus.Errorf("Creating the api container didn't complete successfully, so we tried to delete Kubernetes service '%v' that we created but an error was thrown:\n%v", apiContainerServiceName, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove Kubernetes service with name '%v'!!!!!!!", apiContainerServiceName)
			}
		}
	}()

	resultApiContainer, err := getApiContainerObjectFromKubernetesService(service)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting api container object from Kubernetes service '%+v'", service)
	}

	shouldRemovePod = false
	shouldRemoveService = false
	shouldRemoveAllApiContainerRoleBasedResources = false
	return resultApiContainer, nil
}

func (backend *KubernetesKurtosisBackend) GetAPIContainers(
	ctx context.Context,
	filters *api_container.APIContainerFilters,
) (
	map[enclave.EnclaveID]*api_container.APIContainer,
	error,
) {
	matchingApiContainersByNamespaceAndServiceName, err := backend.getMatchingApiContainers(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting API containers matching the following filters: %+v", filters)
	}

	matchingApiContainersByEnclaveID := map[enclave.EnclaveID]*api_container.APIContainer{}
	for _, apiContainerServices := range matchingApiContainersByNamespaceAndServiceName {
		for _, apiContainerObj := range apiContainerServices {
			matchingApiContainersByEnclaveID[apiContainerObj.GetEnclaveID()] = apiContainerObj
		}
	}

	return matchingApiContainersByEnclaveID, nil
}

func (backend *KubernetesKurtosisBackend) StopAPIContainers(
	ctx context.Context,
	filters *api_container.APIContainerFilters,
) (
	map[enclave.EnclaveID]bool,
	map[enclave.EnclaveID]error,
	error,
) {
	matchingApiContainersByNamespaceAndServiceName, err := backend.getMatchingApiContainers(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting api containers matching filters '%+v'", filters)
	}

	successfulEnclaveIds :=map[enclave.EnclaveID]bool{}
	erroredEnclaveIds := map[enclave.EnclaveID]error{}

	//First iterate over namespaces
	for enclaveNamespace, apiContainerServices := range matchingApiContainersByNamespaceAndServiceName {
		apiContainerServicesToApiContainerPodsMap := map[string]string{}

		//Then over services
		for apiContainerServiceName, apiContainerObj := range apiContainerServices {
			apiContainerPod, err := backend.getApiContainerPod(ctx, apiContainerObj.GetEnclaveID(), enclaveNamespace)
			if err != nil {
				return nil, nil, stacktrace.Propagate(err, "An error occurred getting the api container pod for enclave with ID '%v' in namespace '%v'", apiContainerObj.GetEnclaveID(), enclaveNamespace)
			}
			if _, found := apiContainerServicesToApiContainerPodsMap[apiContainerServiceName]; found {
				return nil, nil, stacktrace.NewError("Api container service name '%v' already exist in the api container services to api container pod map '%+v'; this should never happen and is a bug", apiContainerServiceName, apiContainerServicesToApiContainerPodsMap)
			}
			apiContainerServicesToApiContainerPodsMap[apiContainerServiceName] = apiContainerPod.GetName()
		}
		successfulServiceNames, erroredServiceNames := backend.removeApiContainerServiceSelectorsAndApiContainerPods(ctx, enclaveNamespace, apiContainerServicesToApiContainerPodsMap)

		removeApiContainerServiceSelectorsAndApiContainerPodsSuccessfulEnclaveIds := map[enclave.EnclaveID]bool{}
		for serviceName := range successfulServiceNames {
			apiContainerObj, found := apiContainerServices[serviceName]
			if !found {
				return nil, nil, stacktrace.NewError("Expected to find service name '%v' in the api container service list '%+v' but it was not found; this is a bug in Kurtosis!", serviceName, apiContainerServices)
			}
			removeApiContainerServiceSelectorsAndApiContainerPodsSuccessfulEnclaveIds[apiContainerObj.GetEnclaveID()] = true
		}

		for serviceName, err := range erroredServiceNames {
			apiContainerObj, found := apiContainerServices[serviceName]
			if !found {
				return nil, nil, stacktrace.NewError("Expected to find service name '%v' in the api container service list '%+v' but it was not found; this is a bug in Kurtosis!", serviceName, apiContainerServices)
			}
			wrappedErr := stacktrace.Propagate(err, "An error occurred removing api container selectors and pods from Kubernetes service for Kurtosis api container in enclave with ID '%v' and Kubernetes service name '%v'", apiContainerObj.GetEnclaveID(), serviceName)
			erroredEnclaveIds[apiContainerObj.GetEnclaveID()] = wrappedErr
		}

		successfulEnclaveIds = removeApiContainerServiceSelectorsAndApiContainerPodsSuccessfulEnclaveIds
	}

	return successfulEnclaveIds, erroredEnclaveIds, nil
}

func (backend *KubernetesKurtosisBackend) DestroyAPIContainers(
	ctx context.Context,
	filters *api_container.APIContainerFilters,
) (
	map[enclave.EnclaveID]bool,
	map[enclave.EnclaveID]error,
	error,
) {

	matchingApiContainersByNamespaceAndServiceName, err := backend.getMatchingApiContainers(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Failed to get api container matching filters '%+v'", filters)
	}

	successfulEnclaveIds :=map[enclave.EnclaveID]bool{}
	erroredEnclaveIds := map[enclave.EnclaveID]error{}

	//First iterate over namespaces
	for enclaveNamespace, apiContainerServices := range matchingApiContainersByNamespaceAndServiceName {
		apiContainerServicesToApiContainerPodsMap := map[string]string{}

		//Then over services
		for apiContainerServiceName, apiContainerObj := range apiContainerServices {
			apiContainerPod, err := backend.getApiContainerPod(ctx, apiContainerObj.GetEnclaveID(), enclaveNamespace)
			if err != nil {
				return nil, nil, stacktrace.Propagate(err, "An error occurred getting the api container pod for enclave with ID '%v' in namespace '%v'", apiContainerObj.GetEnclaveID(), enclaveNamespace)
			}
			if _, found := apiContainerServicesToApiContainerPodsMap[apiContainerServiceName]; found {
				return nil, nil, stacktrace.NewError("Api container service name '%v' already exist in the api container services to api container pod map '%+v'; it should never happens; this is a bug in Kurtosis", apiContainerServiceName, apiContainerServicesToApiContainerPodsMap)
			}
			apiContainerServicesToApiContainerPodsMap[apiContainerServiceName] = apiContainerPod.GetName()
		}

		successfulServiceNames, erroredServiceNames := backend.removeApiContainerServicesAndApiContainerPods(ctx, enclaveNamespace, apiContainerServicesToApiContainerPodsMap)

		removeApiContainerServicesAndApiContainerPodsSuccessfulEnclaveIds := map[enclave.EnclaveID]bool{}
		for serviceName := range successfulServiceNames {
			apiContainerObj, found := apiContainerServices[serviceName]
			if !found {
				return nil, nil, stacktrace.NewError("Expected to find service name '%v' in the api container service list '%+v' but it was not found; this is a bug in Kurtosis!", serviceName, apiContainerServices)
			}
			removeApiContainerServicesAndApiContainerPodsSuccessfulEnclaveIds[apiContainerObj.GetEnclaveID()] = true
		}

		for serviceName, err := range erroredServiceNames {
			apiContainerObj, found := apiContainerServices[serviceName]
			if !found {
				return nil, nil, stacktrace.NewError("Expected to find service name '%v' in the api container service list '%+v' but it was not found; this is a bug in Kurtosis!", serviceName, apiContainerServices)
			}
			wrappedErr := stacktrace.Propagate(err, "An error occurred removing api container service and pod for enclave with ID '%v' and Kubernetes service name '%v'", apiContainerObj.GetEnclaveID(), serviceName)
			erroredEnclaveIds[apiContainerObj.GetEnclaveID()] = wrappedErr
		}
		
		removeRoleBasedResourcesSuccessfulEnclaveIds, removeRoleBasedResourcesErroredEnclaveIds := backend.removeApiContainerRoleBasedResources(ctx, enclaveNamespace, removeApiContainerServicesAndApiContainerPodsSuccessfulEnclaveIds)

		for erroredEnclaveId, removeRoleBasedResourcesErr := range removeRoleBasedResourcesErroredEnclaveIds {
			wrappedErr := stacktrace.Propagate(removeRoleBasedResourcesErr, "An error occurred removing api container role based resources for enclave with ID '%v' ", erroredEnclaveId)
			erroredEnclaveIds[erroredEnclaveId] = wrappedErr
		}
		
		successfulEnclaveIds = removeRoleBasedResourcesSuccessfulEnclaveIds
	}
	
	return successfulEnclaveIds, erroredEnclaveIds, nil
}

// ====================================================================================================
//                                     Private Helper Methods
// ====================================================================================================
// Gets api containers matching the search filters, indexed by their [namespace][service name]
func (backend *KubernetesKurtosisBackend) getMatchingApiContainers(
	ctx context.Context,
	filters *api_container.APIContainerFilters,
) (
	map[string]map[string]*api_container.APIContainer,
	error,
) {
	matchingApiContainers := map[string]map[string]*api_container.APIContainer{}

	apiContainersMatchLabels := getApiContainerMatchLabels()

	enclaveNamespaces, err := backend.getAllEnclaveNamespaces(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting all enclave namespaces")
	}

	for _, enclaveNamespace := range enclaveNamespaces {
		enclaveNamespaceName := enclaveNamespace.GetName()
		enclaveNamespaceLabels := enclaveNamespace.GetLabels()

		enclaveIdStr, found := enclaveNamespaceLabels[label_key_consts.EnclaveIDLabelKey.GetString()]
		if !found {
			return nil, stacktrace.NewError("Expected to find a label with name '%v' in Kubernetes namespace '%v', instead no such label was found", label_key_consts.EnclaveIDLabelKey.GetString(), enclaveNamespaceName)
		}
		enclaveId := enclave.EnclaveID(enclaveIdStr)
		// If the EnclaveIDs filter is specified, drop api containers not matching it
		if filters.EnclaveIDs != nil && len(filters.EnclaveIDs) > 0 {
			if _, found := filters.EnclaveIDs[enclaveId]; !found {
				continue
			}
		}

		serviceList, err := backend.kubernetesManager.GetServicesByLabels(ctx, enclaveNamespaceName, apiContainersMatchLabels)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting api container services using labels: %+v in namespace '%v'", apiContainersMatchLabels, enclaveNamespaceName)
		}

		// Instantiate an empty map for this namespace
		matchingApiContainers[enclaveNamespaceName] = map[string]*api_container.APIContainer{}
		for _, service := range serviceList.Items {
			apiContainerObj, err := getApiContainerObjectFromKubernetesService(&service)
			if err != nil {
				return nil, stacktrace.Propagate(err, "Expected to be able to get a Kurtosis api container object service from Kubernetes service '%v', instead a non-nil error was returned", service.Name)
			}

			// If status filter is specified, drop api containers not matching it
			if filters.Statuses != nil && len(filters.Statuses) > 0 {
				if _, found := filters.Statuses[apiContainerObj.GetStatus()]; !found {
					continue
				}
			}
			matchingApiContainers[enclaveNamespaceName][service.Name] = apiContainerObj
		}
	}

	return matchingApiContainers, nil
}

// TODO parallelize to improve performance
func (backend *KubernetesKurtosisBackend) removeApiContainerServiceSelectorsAndApiContainerPods(
	ctx context.Context,
	enclaveNamespace string,
	serviceNameToPodNameMap map[string]string,
) (
	map[string]bool,
	map[string]error,
) {
	successfulServices := map[string]bool{}
	failedServices := map[string]error{}
	for serviceName, podName := range serviceNameToPodNameMap {
		if err := backend.kubernetesManager.RemoveSelectorsFromService(ctx, enclaveNamespace, serviceName); err != nil {
			wrapperErr := stacktrace.Propagate(err, "An error occurred removing selectors from service '%v' in namespace '%v'", serviceName, enclaveNamespace)
			failedServices[serviceName] = wrapperErr
		} else {
			if err := backend.kubernetesManager.RemovePod(ctx, enclaveNamespace, podName); err != nil {
				wrapperErr := stacktrace.Propagate(err, "An error occurred removing pod '%v' associated with service '%v' in namespace '%v'", podName, serviceName, enclaveNamespace)
				failedServices[serviceName] = wrapperErr
				continue
			}
			successfulServices[serviceName] = true
		}
	}

	return successfulServices, failedServices
}

func (backend *KubernetesKurtosisBackend) removeApiContainerServicesAndApiContainerPods(
	ctx context.Context,
	enclaveNamespace string,
	serviceNameToPodNameMap map[string]string,
) (
	map[string]bool,
	map[string]error,
) {
	successfulServices := map[string]bool{}
	failedServices := map[string]error{}
	for serviceName, podName := range serviceNameToPodNameMap {
		if err := backend.kubernetesManager.RemoveService(ctx, enclaveNamespace, serviceName); err != nil {
			wrapperErr := stacktrace.Propagate(err, "An error occurred removing service '%v' in namespace '%v'", serviceName, enclaveNamespace)
			failedServices[serviceName] = wrapperErr
		} else {
			//First checks if Api Container Pod exist because it could have been destroyed with StopEngines
			pod, err := backend.kubernetesManager.GetPod(ctx, enclaveNamespace, podName)
			if err != nil {
				wrapperErr := stacktrace.Propagate(err, "An error occurred getting pod '%v' in namespace '%v'", podName, enclaveNamespace)
				failedServices[serviceName] = wrapperErr
				continue
			}
			//Remove pod if it exists
			if pod != nil {
				if err := backend.kubernetesManager.RemovePod(ctx, enclaveNamespace, podName); err != nil {
					wrapperErr := stacktrace.Propagate(err, "An error occurred removing pod '%v' associated with service '%v' in namespace '%v'", podName, serviceName, enclaveNamespace)
					failedServices[serviceName] = wrapperErr
					continue
				}
			}
			successfulServices[serviceName] = true
		}
	}

	return successfulServices, failedServices
}

func (backend *KubernetesKurtosisBackend) createApiContainerRoleBasedResources(
	ctx context.Context,
	namespace string,
	apiContainerAttributesProvider object_attributes_provider.KubernetesApiContainerObjectAttributesProvider,
) (
	resultApiContainerServiceAccountName string,
	resultErr error,
) {

	//First create the service account
	serviceAccountAttributes, err := apiContainerAttributesProvider.ForApiContainerServiceAccount()
	if err != nil {
		return "", stacktrace.Propagate(
			err,
			"Expected to be able to get api container attributes for a Kubernetes service account, instead got a non-nil error",
		)
	}

	serviceAccountName := serviceAccountAttributes.GetName().GetString()
	serviceAccountLabels := getStringMapFromLabelMap(serviceAccountAttributes.GetLabels())

	if _, err = backend.kubernetesManager.CreateServiceAccount(ctx, serviceAccountName, namespace, serviceAccountLabels); err != nil {
		return "",  stacktrace.Propagate(err, "An error occurred creating service account '%v' with labels '%+v' in namespace '%v'", serviceAccountName, serviceAccountLabels, namespace)
	}

	//Then the role
	rolesAttributes, err := apiContainerAttributesProvider.ForApiContainerRole()
	if err != nil {
		return "", stacktrace.Propagate(
			err,
			"Expected to be able to get api container attributes for a Kubernetes role, instead got a non-nil error",
		)
	}

	roleName := rolesAttributes.GetName().GetString()
	roleLabels := getStringMapFromLabelMap(rolesAttributes.GetLabels())

	rolePolicyRules := []rbacv1.PolicyRule{
		{
			Verbs: []string{consts.CreateKubernetesVerb, consts.UpdateKubernetesVerb, consts.PatchKubernetesVerb, consts.DeleteKubernetesVerb, consts.GetKubernetesVerb, consts.ListKubernetesVerb, consts.WatchKubernetesVerb},
			APIGroups: []string{rbacv1.APIGroupAll},
			Resources: []string{consts.PodsKubernetesResource, consts.ServicesKubernetesResource, consts.PersistentVolumeClaimsKubernetesResource},
		},
	}

	if _, err = backend.kubernetesManager.CreateRole(ctx, roleName, namespace, rolePolicyRules, roleLabels); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating role '%v' with policy rules '%+v' and labels '%+v' in namespace '%v'", roleName, rolePolicyRules, roleLabels, namespace)
	}

	//And finally, the role binding
	roleBindingsAttributes, err := apiContainerAttributesProvider.ForApiContainerRoleBindings()
	if err != nil {
		return "", stacktrace.Propagate(
			err,
			"Expected to be able to get api container attributes for a Kubernetes role bindings, instead got a non-nil error",
		)
	}

	roleBindingsName := roleBindingsAttributes.GetName().GetString()
	roleBindingsLabels := getStringMapFromLabelMap(roleBindingsAttributes.GetLabels())

	roleBindingsSubjects := []rbacv1.Subject{
		{
			Kind:      rbacv1.ServiceAccountKind,
			Name:      serviceAccountName,
			Namespace: namespace,
		},
	}

	roleBindingsRoleRef := rbacv1.RoleRef{
		APIGroup: consts.RbacAuthorizationApiGroup,
		Kind:     consts.RoleKubernetesResourceType,
		Name:     roleName,
	}

	if _, err := backend.kubernetesManager.CreateRoleBindings(ctx, roleBindingsName, namespace, roleBindingsSubjects, roleBindingsRoleRef, roleBindingsLabels); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating role bindings '%v' with subjects '%+v' and role ref '%+v' in namespace '%v'", roleBindingsName, roleBindingsSubjects, roleBindingsRoleRef, namespace)
	}

	return serviceAccountName, nil
}

// TODO parallelize to improve performance
func (backend *KubernetesKurtosisBackend) removeApiContainerRoleBasedResources(
	ctx context.Context,
	enclaveNamespaceName string,
	enclaveIds map[enclave.EnclaveID]bool,
) (
	resultSuccessfulEnclaveIds map[enclave.EnclaveID]bool,
	resultErroredEnclaveIds map[enclave.EnclaveID]error,
) {

	successfulEnclaveIds := map[enclave.EnclaveID]bool{}
	erroredEnclaveIds := map[enclave.EnclaveID]error{}
	for enclaveId := range enclaveIds {

		//First remove role bindings
		roleBindings, err := backend.getAllApiContainerRoleBindings(ctx, enclaveNamespaceName, enclaveId)
		if err != nil {
			wrapErr := stacktrace.Propagate(err, "An error occurred getting all api container role bindings in enclave with ID '%v' and namespace '%v'", enclaveId, enclaveNamespaceName)
			erroredEnclaveIds[enclaveId] = wrapErr
			continue
		}

		errRoleBindingNames := []string{}
		errRoleBindingErrMsgs := []string{}
		for _, roleBinding := range roleBindings {
			roleBindingName := roleBinding.GetName()
			if err := backend.kubernetesManager.RemoveRoleBindings(ctx, roleBindingName, enclaveNamespaceName); err != nil {
				wrapErr := stacktrace.Propagate(err, "An error occurred removing api container role binding '%v' in namespace '%v'", roleBindingName, enclaveNamespaceName)
				errRoleBindingNames = append(errRoleBindingNames, roleBindingName)
				errRoleBindingErrMsgs = append(errRoleBindingErrMsgs, wrapErr.Error())
			}
		}

		if len(errRoleBindingNames) > 0 {
			errRoleBindingNamesStr := strings.Join(errRoleBindingNames, sentencesSeparator)
			errRoleBindingErrMsgsStr := strings.Join(errRoleBindingErrMsgs, sentencesSeparator)
			wrapErr := stacktrace.NewError("An error occurred removing role bindings '%v' error: %v ", errRoleBindingNames, errRoleBindingErrMsgsStr)
			erroredEnclaveIds[enclaveId] = wrapErr
			logrus.Errorf("Removing the api container role-based resources didn't complete successfully, so we tried to delete Kubernetes role bindings '%v' that we created but an error was thrown:\n%v", errRoleBindingNamesStr, errRoleBindingErrMsgsStr)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually remove Kubernetes role bindings with name '%v'!!!!!!!", errRoleBindingNamesStr)
			continue
		}

		//Then roles
		roles, err := backend.getAllApiContainerRoles(ctx, enclaveNamespaceName, enclaveId)
		if err != nil {
			wrapErr := stacktrace.Propagate(err, "An error occurred getting all api container roles in enclave with ID '%v' and namespace '%v'", enclaveId, enclaveNamespaceName)
			erroredEnclaveIds[enclaveId] = wrapErr
			continue
		}

		errRoleNames := []string{}
		errRoleErrMsgs := []string{}
		for _, role := range roles {
			roleName := role.GetName()
			if err := backend.kubernetesManager.RemoveRole(ctx, roleName, enclaveNamespaceName); err != nil {
				wrapErr := stacktrace.Propagate(err, "An error occurred removing api container role '%v' in namespace '%v'", roleName, enclaveNamespaceName)
				errRoleNames = append(errRoleNames, roleName)
				errRoleErrMsgs = append(errRoleErrMsgs, wrapErr.Error())
			}
		}

		if len(errRoleNames) > 0 {
			errRoleNamesStr := strings.Join(errRoleNames, sentencesSeparator)
			errRoleErrMsgsStr := strings.Join(errRoleErrMsgs, sentencesSeparator)
			wrapErr := stacktrace.NewError("An error occurred removing roles '%v' error: %v ", errRoleNamesStr, errRoleErrMsgsStr)
			erroredEnclaveIds[enclaveId] = wrapErr
			logrus.Errorf("Removing the api container role-based resources didn't complete successfully, so we tried to delete Kubernetes roles '%v' that we created but an error was thrown:\n%v", errRoleNamesStr, errRoleErrMsgsStr)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually remove Kubernetes roles with name '%v'!!!!!!!", errRoleNamesStr)
			continue
		}

		//And finally, the service accounts
		serviceAccounts, err := backend.getAllApiContainerServiceAccounts(ctx, enclaveNamespaceName, enclaveId)
		if err != nil {
			wrapErr := stacktrace.Propagate(err, "An error occurred getting all api container service accounts in enclave with ID '%v' and namespace '%v'", enclaveId, enclaveNamespaceName)
			erroredEnclaveIds[enclaveId] = wrapErr
			continue
		}

		errServiceAccountNames := []string{}
		errServiceAccountErrMsgs := []string{}
		for _, serviceAccount := range serviceAccounts {
			serviceAccountName := serviceAccount.GetName()
			if err := backend.kubernetesManager.RemoveServiceAccount(ctx, serviceAccountName, enclaveNamespaceName); err != nil {
				wrapErr := stacktrace.Propagate(err, "An error occurred removing api container service account '%v' in namespace '%v'", serviceAccountName, enclaveNamespaceName)
				errServiceAccountNames = append(errServiceAccountNames, serviceAccountName)
				errServiceAccountErrMsgs = append(errServiceAccountErrMsgs, wrapErr.Error())
			}
		}

		if len(errServiceAccountNames) > 0 {
			errServiceAccountNamesStr := strings.Join(errServiceAccountNames, sentencesSeparator)
			errServiceAccountErrMsgsStr := strings.Join(errServiceAccountErrMsgs, sentencesSeparator)
			wrapErr := stacktrace.NewError("An error occurred removing service accounts '%v' error: %v ", errServiceAccountNamesStr, errServiceAccountErrMsgsStr)
			erroredEnclaveIds[enclaveId] = wrapErr
			logrus.Errorf("Removing the api container role-based resources didn't complete successfully, so we tried to delete Kubernetes service accounts '%v' that we created but an error was thrown:\n%v", errServiceAccountNamesStr, errServiceAccountErrMsgsStr)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually remove Kubernetes service accounts with name '%v' in namespace '%v'!!!!!!!", errServiceAccountNamesStr, enclaveNamespaceName)
			continue
		}

		successfulEnclaveIds[enclaveId] = true
	}

	return successfulEnclaveIds, erroredEnclaveIds
}

func (backend *KubernetesKurtosisBackend) getAllApiContainerServiceAccounts(ctx context.Context, namespace string, enclaveId enclave.EnclaveID) ([]*apiv1.ServiceAccount, error) {
	enclaveIdStr := string(enclaveId)

	apiContainerMatchLabels := getApiContainerMatchLabels()

	serviceAccounts, err := backend.kubernetesManager.GetServiceAccountsByLabels(ctx, namespace, apiContainerMatchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting api container service accounts using labels '%+v' ", apiContainerMatchLabels)
	}

	filteredServiceAccounts := []*apiv1.ServiceAccount{}

	for _, foundServiceAccount := range serviceAccounts.Items {
		foundServiceAccountLabels := foundServiceAccount.GetLabels()

		foundEnclaveId, found := foundServiceAccountLabels[label_key_consts.EnclaveIDLabelKey.GetString()]
		if !found {
			return nil, stacktrace.NewError("Expected to find enclave ID label key '%v' but none was found", label_key_consts.EnclaveIDLabelKey.GetString())
		}

		if enclaveIdStr == foundEnclaveId {
			filteredServiceAccounts = append(filteredServiceAccounts, &foundServiceAccount)
		}
	}

	return filteredServiceAccounts, nil
}

func (backend *KubernetesKurtosisBackend) getAllApiContainerRoles(ctx context.Context, namespace string, enclaveId enclave.EnclaveID) ([]*rbacv1.Role, error) {
	enclaveIdStr := string(enclaveId)

	apiContainerMatchLabels := getApiContainerMatchLabels()

	roles, err := backend.kubernetesManager.GetRolesByLabels(ctx, namespace, apiContainerMatchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting api container roles using labels '%+v' ", apiContainerMatchLabels)
	}

	filteredRoles := []*rbacv1.Role{}

	for _, foundRole := range roles.Items {
		foundRoleLabels := foundRole.GetLabels()

		foundEnclaveId, found := foundRoleLabels[label_key_consts.EnclaveIDLabelKey.GetString()]
		if !found {
			return nil, stacktrace.NewError("Expected to find enclave ID label key '%v' but none was found", label_key_consts.EnclaveIDLabelKey.GetString())
		}

		if enclaveIdStr == foundEnclaveId {
			filteredRoles = append(filteredRoles, &foundRole)
		}
	}

	return filteredRoles, nil
}

func (backend *KubernetesKurtosisBackend) getAllApiContainerRoleBindings(ctx context.Context, namespace string, enclaveId enclave.EnclaveID) ([]*rbacv1.RoleBinding, error) {
	enclaveIdStr := string(enclaveId)

	apiContainerMatchLabels := getApiContainerMatchLabels()

	roleBindings, err := backend.kubernetesManager.GetRoleBindingsByLabels(ctx, namespace, apiContainerMatchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting api container role bindings using labels '%+v' ", apiContainerMatchLabels)
	}

	filteredRoleBindings := []*rbacv1.RoleBinding{}

	for _, foundRoleBinding := range roleBindings.Items {
		foundRoleBindingLabels := foundRoleBinding.GetLabels()

		foundEnclaveId, found := foundRoleBindingLabels[label_key_consts.EnclaveIDLabelKey.GetString()]
		if !found {
			return nil, stacktrace.NewError("Expected to find enclave ID label key '%v' but none was found", label_key_consts.EnclaveIDLabelKey.GetString())
		}

		if enclaveIdStr == foundEnclaveId {
			filteredRoleBindings = append(filteredRoleBindings, &foundRoleBinding)
		}
	}

	return filteredRoleBindings, nil
}

// The current Kurtosis Kubernetes architecture defines only one pod for Api Container
// This method should be refactored if the architecture changes, and we decide to use replicas for Api Containers
func (backend *KubernetesKurtosisBackend) getApiContainerPod(ctx context.Context, enclaveId enclave.EnclaveID, namespace string) (*apiv1.Pod, error) {
	enclaveIdStr := string(enclaveId)

	apiContainerMatchLabels := getApiContainerMatchLabels()

	pods, err := backend.kubernetesManager.GetPodsByLabels(ctx, namespace, apiContainerMatchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the api container pod in namespace '%v' using labels '%+v' ", namespace, apiContainerMatchLabels)
	}

	filteredPods := []*apiv1.Pod{}

	for _, foundPod := range pods.Items {
		foundPodLabels := foundPod.GetLabels()

		foundEnclaveId, found := foundPodLabels[label_key_consts.EnclaveIDLabelKey.GetString()]
		if !found {
			return nil, stacktrace.NewError("Expected to find enclave ID label key '%v' but none was found", label_key_consts.EnclaveIDLabelKey.GetString())
		}

		if enclaveIdStr == foundEnclaveId {
			filteredPods = append(filteredPods, &foundPod)
		}
	}
	numOfPods := len(filteredPods)
	if numOfPods == 0 {
		return nil, stacktrace.NewError("No pods matching labels '%+v' was found", apiContainerMatchLabels)
	}
	//We are not using replicas for Kurtosis api containers
	if numOfPods > 1 {
		return nil, stacktrace.NewError("Expected to find only one api container pod in enclave with ID '%v', but '%v' was found; this is a bug in Kurtosis", enclaveId, numOfPods)
	}

	resultPod := filteredPods[0]

	return resultPod, nil
}

func getApiContainerObjectFromKubernetesService(service *apiv1.Service) (*api_container.APIContainer, error) {
	enclaveId, found := service.Labels[label_key_consts.EnclaveIDLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected to be able to find label describing the enclave ID on service '%v' with label key '%v', but was unable to", service.Name, label_key_consts.EnclaveIDLabelKey.GetString())
	}

	status := getContainerStatusFromKubernetesService(service)
	var privateIpAddr net.IP
	var privateGrpcPortSpec *port_spec.PortSpec
	var privateGrpcProxyPortSpec *port_spec.PortSpec
	if status == container_status.ContainerStatus_Running {
		privateIpAddr = net.ParseIP(service.Spec.ClusterIP)
		if privateIpAddr == nil {
			return nil, stacktrace.NewError("Expected to be able to get the cluster ip of the api container service, instead parsing the cluster ip of service '%v' returned nil", service.Name)
		}
		var portSpecError error
		privateGrpcPortSpec, privateGrpcProxyPortSpec, portSpecError = getGrpcAndGrpcProxyPortSpecsFromServicePorts(service.Spec.Ports)
		if portSpecError != nil {
			return nil, stacktrace.Propagate(portSpecError, "Expected to be able to determine api container grpc port specs from Kubernetes service ports for api container in enclave with ID '%v', instead a non-nil error was returned", enclaveId)
		}
	}

	// NOTE: We set these to nil because in Kubernetes we have no way of knowing what the public info is!
	var publicIpAddr net.IP = nil
	var publicGrpcPortSpec *port_spec.PortSpec = nil
	var publicGrpcProxyPortSpec *port_spec.PortSpec = nil

	resultApiContainer := api_container.NewAPIContainer(
		enclave.EnclaveID(enclaveId),
		status,
		privateIpAddr,
		privateGrpcPortSpec,
		privateGrpcProxyPortSpec,
		publicIpAddr,
		publicGrpcPortSpec,
		publicGrpcProxyPortSpec,
	)

	return resultApiContainer, nil
}

func getApiContainerContainersAndVolumes(
	containerImageAndTag string,
	containerPorts []apiv1.ContainerPort,
	envVars map[string]string,
	enclaveDataPersistentVolumeClaim *apiv1.PersistentVolumeClaim,
	enclaveDataVolumeDirpath string,
) (
	resultContainers []apiv1.Container,
	resultVolumes []apiv1.Volume,
) {

	var containerEnvVars []apiv1.EnvVar
	for varName, varValue := range envVars {
		envVar := apiv1.EnvVar{
			Name:  varName,
			Value: varValue,
		}
		containerEnvVars = append(containerEnvVars, envVar)
	}
	containers := []apiv1.Container{
		{
			Name:  kurtosisApiContainerContainerName,
			Image: containerImageAndTag,
			Env:   containerEnvVars,
			Ports: containerPorts,
			VolumeMounts: []apiv1.VolumeMount{
				{
					Name:      enclaveDataPersistentVolumeClaim.Spec.VolumeName,
					MountPath: enclaveDataVolumeDirpath,
				},
			},
		},
	}

	volumes := []apiv1.Volume{
		{
			Name: enclaveDataPersistentVolumeClaim.Spec.VolumeName,
			VolumeSource: apiv1.VolumeSource{
				PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
					ClaimName: enclaveDataPersistentVolumeClaim.GetName(),
				},
			},
		},
	}

	return containers, volumes
}

func getContainerStatusFromKubernetesService(service *apiv1.Service) container_status.ContainerStatus {
	// If a Kubernetes Service has selectors, then we assume the container is reachable, and thus not stopped
	serviceSelectors := service.Spec.Selector
	if len(serviceSelectors) == 0 {
		return container_status.ContainerStatus_Stopped
	}
	return container_status.ContainerStatus_Running
}

func getApiContainerMatchLabels() map[string]string {
	engineMatchLabels := map[string]string{
		label_key_consts.AppIDLabelKey.GetString():                label_value_consts.AppIDLabelValue.GetString(),
		label_key_consts.KurtosisResourceTypeLabelKey.GetString(): label_value_consts.APIContainerKurtosisResourceTypeLabelValue.GetString(),
	}
	return engineMatchLabels
}
