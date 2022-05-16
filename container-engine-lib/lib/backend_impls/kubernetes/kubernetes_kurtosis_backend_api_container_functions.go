package kubernetes

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager/consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_resource_collectors"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/object_name_constants"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/api_container"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"net"
)

const (
	kurtosisApiContainerContainerName = "kurtosis-core-api"
)

// Any of these values being nil indicates that the resource doesn't exist
type apiContainerKubernetesResources struct {
	role *rbacv1.Role

	roleBinding *rbacv1.RoleBinding

	serviceAccount *apiv1.ServiceAccount

	service *apiv1.Service

	pod *apiv1.Pod
}

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
			"An error occurred creating the API container's private grpc port spec object using number '%v' and protocol '%v'",
			grpcPortNum,
			kurtosisServersPortProtocol.String(),
		)
	}
	privateGrpcProxyPortSpec, err := port_spec.NewPortSpec(grpcProxyPortNum, kurtosisServersPortProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the API container's private grpc proxy port spec object using number '%v' and protocol '%v'",
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
		return nil, stacktrace.Propagate(err,"An error occurred getting the API container attributes provider using enclave ID '%v'", enclaveId)
	}

	enclaveNamespace, err := backend.getEnclaveNamespace(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave namespace for enclave with ID '%v'", enclaveId)
	}
	enclaveNamespaceName := enclaveNamespace.GetName()

	//Create the service account
	serviceAccountAttributes, err := apiContainerAttributesProvider.ForApiContainerServiceAccount()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Expected to be able to get API container attributes for a Kubernetes service account, " +
				"instead got a non-nil error",
		)
	}

	serviceAccountName := serviceAccountAttributes.GetName().GetString()
	serviceAccountLabels := getStringMapFromLabelMap(serviceAccountAttributes.GetLabels())
	apiContainerServiceAccount, err := backend.kubernetesManager.CreateServiceAccount(ctx, serviceAccountName, enclaveNamespaceName, serviceAccountLabels);
	if err != nil {
		return nil,  stacktrace.Propagate(err, "An error occurred creating service account '%v' with labels '%+v' in namespace '%v'", serviceAccountName, serviceAccountLabels, enclaveNamespaceName)
	}
	apiContainerServiceAccountName := apiContainerServiceAccount.GetName()
	shouldRemoveServiceAccount := true
	defer func() {
		if shouldRemoveServiceAccount {
			if err := backend.kubernetesManager.RemoveServiceAccount(ctx, apiContainerServiceAccountName, enclaveNamespaceName); err != nil {
				logrus.Errorf("Creating the API container didn't complete successfully, so we tried to delete service account '%v' in namespace '%v' that we created but an error was thrown:\n%v", apiContainerServiceAccountName, enclaveNamespaceName, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove service account with name '%v'!!!!!!!", apiContainerServiceAccountName)
			}
		}
	}()

	//Create the role
	rolesAttributes, err := apiContainerAttributesProvider.ForApiContainerRole()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Expected to be able to get API container attributes for a Kubernetes role, " +
				"instead got a non-nil error",
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

	apiContainerRole, err := backend.kubernetesManager.CreateRole(ctx, roleName, enclaveNamespaceName, rolePolicyRules, roleLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating role '%v' with policy rules '%+v' " +
			"and labels '%+v' in namespace '%v'", roleName, rolePolicyRules, roleLabels, enclaveNamespaceName)
	}
	shouldRemoveRole := true
	defer func() {
		if shouldRemoveRole {
			if err := backend.kubernetesManager.RemoveRole(ctx, roleName, enclaveNamespaceName); err != nil {
				logrus.Errorf("Creating the API container didn't complete successfully, so we tried to delete role '%v' in namespace '%v' that we created but an error was thrown:\n%v", roleName, enclaveNamespaceName, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove role with name '%v'!!!!!!!", roleName)
			}
		}
	}()

	//Create the role binding to join the service account with the role
	roleBindingsAttributes, err := apiContainerAttributesProvider.ForApiContainerRoleBindings()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Expected to be able to get API container attributes for a Kubernetes role bindings, " +
				"instead got a non-nil error",
		)
	}

	roleBindingName := roleBindingsAttributes.GetName().GetString()
	roleBindingsLabels := getStringMapFromLabelMap(roleBindingsAttributes.GetLabels())
	roleBindingsSubjects := []rbacv1.Subject{
		{
			Kind:      rbacv1.ServiceAccountKind,
			Name:      serviceAccountName,
			Namespace: enclaveNamespaceName,
		},
	}

	roleBindingsRoleRef := rbacv1.RoleRef{
		APIGroup: consts.RbacAuthorizationApiGroup,
		Kind:     consts.RoleKubernetesResourceType,
		Name:     roleName,
	}

	 apiContainerRoleBinding, err := backend.kubernetesManager.CreateRoleBindings(ctx, roleBindingName, enclaveNamespaceName, roleBindingsSubjects, roleBindingsRoleRef, roleBindingsLabels)
	 if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating role bindings '%v' with subjects " +
			"'%+v' and role ref '%+v' in namespace '%v'", roleBindingName, roleBindingsSubjects, roleBindingsRoleRef, enclaveNamespaceName)
	}
	shouldRemoveRoleBinding := true
	defer func() {
		if shouldRemoveRoleBinding {
			if err := backend.kubernetesManager.RemoveRoleBindings(ctx, roleBindingName, enclaveNamespaceName); err != nil {
				logrus.Errorf("Creating the API container didn't complete successfully, so we tried to delete role binding '%v' in namespace '%v' that we created but an error was thrown:\n%v", roleBindingName, enclaveNamespaceName, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove role binding with name '%v'!!!!!!!", roleBindingName)
			}
		}
	}()

	// Get Pod Attributes
	apiContainerPodAttributes, err := apiContainerAttributesProvider.ForApiContainerPod()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Expected to be able to get attributes for a Kubernetes pod for API container in enclave with id '%v', instead got a non-nil error",
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

	// Create pods with API container containers and volumes in Kubernetes
	apiContainerPod, err := backend.kubernetesManager.CreatePod(ctx, enclaveNamespaceName, apiContainerPodName, apiContainerPodLabels, apiContainerPodAnnotations, apiContainerContainers, apiContainerVolumes, apiContainerServiceAccountName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the pod with name '%s' in namespace '%s' with image '%s'", apiContainerPodName, enclaveNamespaceName, image)
	}
	var shouldRemovePod = true
	defer func() {
		if shouldRemovePod {
			if err := backend.kubernetesManager.RemovePod(ctx, enclaveNamespaceName, apiContainerPodName); err != nil {
				logrus.Errorf("Creating the API container didn't complete successfully, so we tried to delete Kubernetes pod '%v' that we created but an error was thrown:\n%v", apiContainerPodName, err)
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
			"An error occurred getting the API container service attributes using private grpc port spec '%+v', and "+
				"private grpc proxy port spec '%+v'",
			privateGrpcPortSpec,
			privateGrpcProxyPortSpec,
		)
	}
	apiContainerServiceName := apiContainerServiceAttributes.GetName().GetString()
	apiContainerServiceLabels := getStringMapFromLabelMap(apiContainerServiceAttributes.GetLabels())
	apiContainerServiceAnnotations := getStringMapFromAnnotationMap(apiContainerServiceAttributes.GetAnnotations())

	// Define service ports. These hook up to ports on the containers running in the API container pod
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
	apiContainerService, err := backend.kubernetesManager.CreateService(ctx, enclaveNamespaceName, apiContainerServiceName, apiContainerServiceLabels, apiContainerServiceAnnotations, apiContainerPodLabels, externalServiceType, servicePorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the service with name '%s' in namespace '%s' with ports '%v' and '%v'", apiContainerServiceName, enclaveNamespaceName, grpcPortInt32, grpcProxyPortInt32)
	}
	var shouldRemoveService = true
	defer func() {
		if shouldRemoveService {
			if err := backend.kubernetesManager.RemoveService(ctx, enclaveNamespaceName, apiContainerServiceName); err != nil {
				logrus.Errorf("Creating the API container didn't complete successfully, so we tried to delete Kubernetes service '%v' that we created but an error was thrown:\n%v", apiContainerServiceName, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove Kubernetes service with name '%v'!!!!!!!", apiContainerServiceName)
			}
		}
	}()

	apiContainerResources := &apiContainerKubernetesResources{
		role:             apiContainerRole,
		roleBinding:      apiContainerRoleBinding,
		serviceAccount:   apiContainerServiceAccount,
		service:          apiContainerService,
		pod:              apiContainerPod,
	}
	apiContainerObjsById, err := getApiContainerObjectsFromKubernetesResources(map[enclave.EnclaveID]*apiContainerKubernetesResources{
		enclaveId: apiContainerResources,
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred converting the new API container's Kubernetes resources to API container objects")
	}
	resultApiContainer, found := apiContainerObjsById[enclaveId]
	if !found {
		return nil, stacktrace.NewError("Successfully converted the new API container's Kubernetes resources to an API container object, but the resulting map didn't have an entry for enclave ID '%v'", enclaveId)
	}

	shouldRemoveRoleBinding = false
	shouldRemoveRole = false
	shouldRemoveServiceAccount = false
	shouldRemovePod = false
	shouldRemoveService = false
	return resultApiContainer, nil
}

func (backend *KubernetesKurtosisBackend) GetAPIContainers(
	ctx context.Context,
	filters *api_container.APIContainerFilters,
) (
	map[enclave.EnclaveID]*api_container.APIContainer,
	error,
) {
	matchingApiContainers, _, err := backend.getMatchingApiContainerObjectsAndKubernetesResources(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting API containers matching the following filters: %+v", filters)
	}
	return matchingApiContainers, nil
}

func (backend *KubernetesKurtosisBackend) StopAPIContainers(
	ctx context.Context,
	filters *api_container.APIContainerFilters,
) (
	map[enclave.EnclaveID]bool,
	map[enclave.EnclaveID]error,
	error,
) {
	_, matchingKubernetesResources, err := backend.getMatchingApiContainerObjectsAndKubernetesResources(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting API containers and Kubernetes resources matching filters '%+v'", filters)
	}

	successfulEnclaveIds := map[enclave.EnclaveID]bool{}
	erroredEnclaveIds := map[enclave.EnclaveID]error{}
	for enclaveId, resources := range matchingKubernetesResources {

		if resources.service != nil {
			serviceName := resources.service.GetName()
			namespaceName := resources.service.GetNamespace()
			if err := backend.kubernetesManager.RemoveSelectorsFromService(ctx, namespaceName, serviceName); err != nil {
				erroredEnclaveIds[enclaveId] = stacktrace.Propagate(
					err,
					"An error occurred removing selectors from service '%v' in namespace '%v' for API container in enclave with ID '%v'",
					serviceName,
					namespaceName,
					enclaveId,
				)
				continue
			}
		}

		if resources.pod != nil {
			podName := resources.pod.GetName()
			namespaceName := resources.pod.GetNamespace()
			if err := backend.kubernetesManager.RemovePod(ctx, namespaceName, podName); err != nil {
				erroredEnclaveIds[enclaveId] = stacktrace.Propagate(
					err,
					"An error occurred removing pod '%v' in namespace '%v' for API container in enclave with ID '%v'",
					podName,
					namespaceName,
					enclaveId,
				)
				continue
			}
		}

		successfulEnclaveIds[enclaveId] = true
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

	_, matchingResources, err := backend.getMatchingApiContainerObjectsAndKubernetesResources(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting API container Kubernetes resources matching filters: %+v", filters)
	}

	successfulEnclaveIds := map[enclave.EnclaveID]bool{}
	erroredEnclaveIds := map[enclave.EnclaveID]error{}
	for enclaveId, resources := range matchingResources {

		// Remove Service
		if resources.service != nil {
			serviceName := resources.service.GetName()
			namespaceName := resources.service.GetNamespace()
			if err := backend.kubernetesManager.RemoveService(ctx, serviceName, namespaceName); err != nil {
				erroredEnclaveIds[enclaveId] = stacktrace.Propagate(
					err,
					"An error occurred removing service '%v' for API container in enclave with ID '%v'",
					serviceName,
					enclaveId,
				)
				continue
			}
		}

		// Remove Pod
		if resources.pod != nil {
			podName := resources.pod.GetName()
			namespaceName := resources.pod.GetNamespace()
			if err := backend.kubernetesManager.RemovePod(ctx, podName, namespaceName); err != nil {
				erroredEnclaveIds[enclaveId] = stacktrace.Propagate(
					err,
					"An error occurred removing pod '%v' for API container in enclave with ID '%v'",
					podName,
					enclaveId,
				)
				continue
			}
		}

		// Remove RoleBinding
		if resources.roleBinding != nil {
			roleBindingName := resources.roleBinding.GetName()
			namespaceName := resources.roleBinding.GetNamespace()
			if err := backend.kubernetesManager.RemoveRoleBindings(ctx, roleBindingName, namespaceName); err != nil {
				erroredEnclaveIds[enclaveId] = stacktrace.Propagate(
					err,
					"An error occurred removing role binding '%v' for API container in enclave with ID '%v'",
					roleBindingName,
					enclaveId,
				)
				continue
			}
		}

		// Remove Role
		if resources.role != nil {
			roleName := resources.role.GetName()
			namespaceName := resources.role.GetNamespace()
			if err := backend.kubernetesManager.RemoveRole(ctx, roleName, namespaceName); err != nil {
				erroredEnclaveIds[enclaveId] = stacktrace.Propagate(
					err,
					"An error occurred removing role '%v' for API container in enclave with ID '%v'",
					roleName,
					enclaveId,
				)
				continue
			}
		}

		// Remove Service Account
		if resources.serviceAccount != nil {
			serviceAccountName := resources.serviceAccount.GetName()
			namespaceName := resources.serviceAccount.GetNamespace()
			if err := backend.kubernetesManager.RemoveServiceAccount(ctx, serviceAccountName, namespaceName); err != nil {
				erroredEnclaveIds[enclaveId] = stacktrace.Propagate(
					err,
					"An error occurred removing service account '%v' for API container in enclave with ID '%v'",
					serviceAccountName,
					enclaveId,
				)
				continue
			}
		}

		successfulEnclaveIds[enclaveId] = true
	}
	return successfulEnclaveIds, erroredEnclaveIds, nil

}

// ====================================================================================================
//                                     Private Helper Methods
// ====================================================================================================
func (backend *KubernetesKurtosisBackend) getMatchingApiContainerObjectsAndKubernetesResources(
	ctx context.Context,
	filters *api_container.APIContainerFilters,
) (
	map[enclave.EnclaveID]*api_container.APIContainer,
	map[enclave.EnclaveID]*apiContainerKubernetesResources,
	error,
) {
	matchingResources, err := backend.getMatchingApiContainerKubernetesResources(ctx, filters.EnclaveIDs)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting API container Kubernetes resources matching enclave IDs: %+v", filters.EnclaveIDs)
	}

	apiContainerObjects, err := getApiContainerObjectsFromKubernetesResources(matchingResources)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting API container objects from Kubernetes resources")
	}

	// Finally, apply the filters
	resultApiContainerObjs := map[enclave.EnclaveID]*api_container.APIContainer{}
	resultKubernetesResources := map[enclave.EnclaveID]*apiContainerKubernetesResources{}
	for enclaveId, apiContainerObj := range apiContainerObjects {
		if filters.EnclaveIDs != nil && len(filters.EnclaveIDs) > 0 {
			if _, found := filters.EnclaveIDs[apiContainerObj.GetEnclaveID()]; !found {
				continue
			}
		}

		if filters.Statuses != nil && len(filters.Statuses) > 0 {
			if _, found := filters.Statuses[apiContainerObj.GetStatus()]; !found {
				continue
			}
		}

		resultApiContainerObjs[enclaveId] = apiContainerObj
		// Okay to do because we're guaranteed a 1:1 mapping between api_container_obj:api_container_resources
		resultKubernetesResources[enclaveId] = matchingResources[enclaveId]
	}

	return resultApiContainerObjs, resultKubernetesResources, nil
}

// Get back any and all API container's Kubernetes resources matching the given enclave IDs, where a nil or empty map == "match all enclave IDs"
func (backend *KubernetesKurtosisBackend) getMatchingApiContainerKubernetesResources(ctx context.Context, enclaveIds map[enclave.EnclaveID]bool) (
	map[enclave.EnclaveID]*apiContainerKubernetesResources,
	error,
) {

	enclaveMatchLabels := getEnclaveMatchLabels()

	result := map[enclave.EnclaveID]*apiContainerKubernetesResources{}

	enclaveIdsStrSet := map[string]bool{}
	for enclaveId, booleanValue := range enclaveIds {
		enclaveIdStr := string(enclaveId)
		enclaveIdsStrSet[enclaveIdStr] = booleanValue
	}

	// Namespaces
	namespaces, err := kubernetes_resource_collectors.CollectMatchingNamespaces(
		ctx,
		backend.kubernetesManager,
		enclaveMatchLabels,
		label_key_consts.EnclaveIDLabelKey.GetString(),
		enclaveIdsStrSet,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave namespaces matching IDs '%+v'", enclaveIdsStrSet)
	}

	apiContainerMatchLabels := getApiContainerMatchLabels()

	// Per-namespace objects
	for enclaveIdStr, namespacesForEnclaveId := range namespaces {
		if len(namespacesForEnclaveId) > 1 {
			return nil, stacktrace.NewError(
				"Expected at most one namespace to match enclave ID '%v', but got '%v'",
				enclaveIdStr,
				len(namespacesForEnclaveId),
			)
		}
		namespaceName := namespacesForEnclaveId[0].GetName()

		//Role Bindings
		roleBindings, err := kubernetes_resource_collectors.CollectMatchingRoleBindings(
			ctx,
			backend.kubernetesManager,
			namespaceName,
			apiContainerMatchLabels,
			label_key_consts.EnclaveIDLabelKey.GetString(),
			map[string]bool{
				enclaveIdStr: true,
			},
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting role bindings matching enclave ID '%v' in namespace '%v'", enclaveIdStr, namespaceName)
		}
		var roleBinding *rbacv1.RoleBinding
		if roleBindingsForEnclaveId, found := roleBindings[enclaveIdStr]; found {
			if len(roleBindingsForEnclaveId) == 0 {
				return nil, stacktrace.NewError(
						"Expected to find one API container role binding in namespace '%v' for enclave with ID '%v' " +
						"but none was found",
					namespaceName,
					enclaveIdStr,
				)
			}
			if len(roleBindingsForEnclaveId) > 1 {
				return nil, stacktrace.NewError(
					"Expected at most one API container role binding in namespace '%v' for enclave with ID '%v' " +
						"but found '%v'",
					namespaceName,
					enclaveIdStr,
					len(roleBindings),
				)
			}
			roleBinding = roleBindingsForEnclaveId[0]
		}

		//Roles
		roles, err := kubernetes_resource_collectors.CollectMatchingRoles(
			ctx,
			backend.kubernetesManager,
			namespaceName,
			apiContainerMatchLabels,
			label_key_consts.EnclaveIDLabelKey.GetString(),
			map[string]bool{
				enclaveIdStr: true,
			},
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting roles matching enclave ID '%v' in namespace '%v'", enclaveIdStr, namespaceName)
		}
		var role *rbacv1.Role
		if rolesForEnclaveId, found := roles[enclaveIdStr]; found {
			if len(rolesForEnclaveId) == 0 {
				return nil, stacktrace.NewError(
					"Expected to find one API container role in namespace '%v' for enclave with ID '%v' " +
						"but none was found",
					namespaceName,
					enclaveIdStr,
				)
			}
			if len(rolesForEnclaveId) > 1 {
				return nil, stacktrace.NewError(
					"Expected at most one API container role in namespace '%v' for enclave with ID '%v' " +
						"but found '%v'",
					namespaceName,
					enclaveIdStr,
					len(roles),
				)
			}
			role = rolesForEnclaveId[0]
		}

		// Service accounts
		serviceAccounts, err := kubernetes_resource_collectors.CollectMatchingServiceAccounts(
			ctx,
			backend.kubernetesManager,
			namespaceName,
			apiContainerMatchLabels,
			label_key_consts.EnclaveIDLabelKey.GetString(),
			map[string]bool{
				enclaveIdStr: true,
			},
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting service accounts matching enclave ID '%v' in namespace '%v'", enclaveIdStr, namespaceName)
		}
		var serviceAccount *apiv1.ServiceAccount
		if serviceAccountsForEnclaveId, found := serviceAccounts[enclaveIdStr]; found {
			if len(serviceAccountsForEnclaveId) == 0 {
				return nil, stacktrace.NewError(
					"Expected to find one API container service account in namespace '%v' for enclave with ID '%v' " +
						"but none was found",
					namespaceName,
					enclaveIdStr,
				)
			}
			if len(serviceAccountsForEnclaveId) > 1 {
				return nil, stacktrace.NewError(
					"Expected at most one API container service account in namespace '%v' for enclave with ID '%v' " +
						"but found '%v'",
					namespaceName,
					enclaveIdStr,
					len(serviceAccounts),
				)
			}
			serviceAccount = serviceAccountsForEnclaveId[0]
		}

		// Services
		services, err := kubernetes_resource_collectors.CollectMatchingServices(
			ctx,
			backend.kubernetesManager,
			namespaceName,
			apiContainerMatchLabels,
			label_key_consts.EnclaveIDLabelKey.GetString(),
			map[string]bool{
				enclaveIdStr: true,
			},
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting services matching enclave ID '%v' in namespace '%v'", enclaveIdStr, namespaceName)
		}
		var service *apiv1.Service
		if servicesForEnclaveId, found := services[enclaveIdStr]; found {
			if len(servicesForEnclaveId) == 0 {
				return nil, stacktrace.NewError(
					"Expected to find one API container service in namespace '%v' for enclave with ID '%v' " +
						"but none was found",
					namespaceName,
					enclaveIdStr,
				)
			}
			if len(servicesForEnclaveId) > 1 {
				return nil, stacktrace.NewError(
					"Expected at most one API container service in namespace '%v' for enclave with ID '%v' " +
						"but found '%v'",
					namespaceName,
					enclaveIdStr,
					len(services),
				)
			}
			service = servicesForEnclaveId[0]
		}

		// Pods
		pods, err := kubernetes_resource_collectors.CollectMatchingPods(
			ctx,
			backend.kubernetesManager,
			namespaceName,
			apiContainerMatchLabels,
			label_key_consts.EnclaveIDLabelKey.GetString(),
			map[string]bool{
				enclaveIdStr: true,
			},
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting pods matching enclave ID '%v' in namespace '%v'", enclaveIdStr, namespaceName)
		}
		var pod *apiv1.Pod
		if podsForEnclaveId, found := pods[enclaveIdStr]; found {
			if len(podsForEnclaveId) == 0 {
				return nil, stacktrace.NewError(
					"Expected to find one API container pod in namespace '%v' for enclave with ID '%v' " +
						"but none was found",
					namespaceName,
					enclaveIdStr,
				)
			}
			if len(podsForEnclaveId) > 1 {
				return nil, stacktrace.NewError(
					"Expected at most one API container pod in namespace '%v' for enclave with ID '%v' " +
						"but found '%v'",
					namespaceName,
					enclaveIdStr,
					len(pods),
				)
			}
			pod = podsForEnclaveId[0]
		}

		enclaveId := enclave.EnclaveID(enclaveIdStr)

		apiContainerResources, found := result[enclaveId]
		if !found {
			apiContainerResources = &apiContainerKubernetesResources{}
		}
		apiContainerResources.service = service
		apiContainerResources.pod = pod
		apiContainerResources.serviceAccount = serviceAccount
		apiContainerResources.role = role
		apiContainerResources.roleBinding = roleBinding

		result[enclaveId] = apiContainerResources
	}

	return result, nil
}

func getApiContainerObjectsFromKubernetesResources(
	allResources map[enclave.EnclaveID]*apiContainerKubernetesResources,
) (
	map[enclave.EnclaveID]*api_container.APIContainer,
	error,
) {
	result := map[enclave.EnclaveID]*api_container.APIContainer{}

	for enclaveId, resourcesForEnclaveId := range allResources {

		if resourcesForEnclaveId.service == nil {
			return nil, stacktrace.NewError("Can not create an API container object if there is not an API container's Kubernetes service in enclave '%v'", enclaveId)
		}

		status, err := getContainerStatusFromPod(resourcesForEnclaveId.pod)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting container status from Kubernetes pod '%+v'", resourcesForEnclaveId.pod)
		}

		privateIpAddr := net.ParseIP(resourcesForEnclaveId.service.Spec.ClusterIP)
		if privateIpAddr == nil {
			return nil, stacktrace.NewError("Expected to be able to get the cluster ip of the API container service, instead parsing the cluster ip of service '%v' returned nil", resourcesForEnclaveId.service.Name)
		}
		var portSpecError error
		privateGrpcPortSpec, privateGrpcProxyPortSpec, portSpecError := getGrpcAndGrpcProxyPortSpecsFromServicePorts(resourcesForEnclaveId.service.Spec.Ports)
		if portSpecError != nil {
			return nil, stacktrace.Propagate(portSpecError, "Expected to be able to determine API container grpc port specs from Kubernetes service ports for API container in enclave with ID '%v', instead a non-nil error was returned", enclaveId)
		}

		// NOTE: We set these to nil because in Kubernetes we have no way of knowing what the public info is!
		var publicIpAddr net.IP = nil
		var publicGrpcPortSpec *port_spec.PortSpec = nil
		var publicGrpcProxyPortSpec *port_spec.PortSpec = nil

		apiContainerObj := api_container.NewAPIContainer(
			enclaveId,
			status,
			privateIpAddr,
			privateGrpcPortSpec,
			privateGrpcProxyPortSpec,
			publicIpAddr,
			publicGrpcPortSpec,
			publicGrpcProxyPortSpec,
		)

		result[enclaveId] = apiContainerObj
	}
	return result, nil
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

func getApiContainerMatchLabels() map[string]string {
	engineMatchLabels := map[string]string{
		label_key_consts.AppIDLabelKey.GetString():                label_value_consts.AppIDLabelValue.GetString(),
		label_key_consts.KurtosisResourceTypeLabelKey.GetString(): label_value_consts.APIContainerKurtosisResourceTypeLabelValue.GetString(),
	}
	return engineMatchLabels
}
