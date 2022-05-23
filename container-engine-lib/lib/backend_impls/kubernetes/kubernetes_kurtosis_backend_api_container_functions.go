package kubernetes

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager/consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_resource_collectors"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/api_container"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	applyconfigurationsv1 "k8s.io/client-go/applyconfigurations/core/v1"
	"net"
	"time"
)

const (
	kurtosisApiContainerContainerName = "kurtosis-core-api"

	// The name of the environment variable that we'll set using the Kubernetes downward API to tell the API container
	// its own namespace name. This is necessary because the Role that the API container will run as won't have permission
	// to list all namespaces on the cluster.
	ApiContainerOwnNamespaceNameEnvVar = "API_CONTAINER_OWN_NAMESPACE_NAME"

	// The Kubernetes FieldPath string specifying the pod's namespace, which we'll use via the Kubernetes downward API
	// to give the API container an environment variable with its own namespace
	kubernetesResourceOwnNamespaceFieldPath = "metadata.namespace"

	maxWaitForApiContainerContainerAvailabilityRetries         = 30
	timeBetweenWaitForApiContainerContainerAvailabilityRetries = 1 * time.Second

	enclaveDataDirVolumeName = "enclave-data"
)

// Any of these values being nil indicates that the resource doesn't exist
type apiContainerKubernetesResources struct {
	// Will never be nil because an API container is defined by its service
	service *apiv1.Service

	pod *apiv1.Pod

	role *rbacv1.Role

	roleBinding *rbacv1.RoleBinding

	serviceAccount *apiv1.ServiceAccount
}

// ====================================================================================================
//                                     API Container CRUD Methods
// ====================================================================================================

func (backend *KubernetesKurtosisBackend) CreateAPIContainer(
	ctx context.Context,
	image string,
	enclaveId enclave.EnclaveID,
	grpcPortNum uint16,
	grpcProxyPortNum uint16,
	enclaveDataVolumeDirpath string,
	ownIpAddressEnvVar string,
	customEnvVars map[string]string,
) (
	*api_container.APIContainer,
	error,
) {
	grpcPortInt32 := int32(grpcPortNum)
	grpcProxyPortInt32 := int32(grpcProxyPortNum)

	// TODO This validation is the same for Docker and for Kubernetes because we are using kurtBackend
	// TODO we could move this to a top layer for validations, perhaps
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
	privatePortSpecs := map[string]*port_spec.PortSpec{
		kurtosisInternalContainerGrpcPortSpecId: privateGrpcPortSpec,
		kurtosisInternalContainerGrpcProxyPortSpecId: privateGrpcProxyPortSpec,
	}

	enclaveAttributesProvider := backend.objAttrsProvider.ForEnclave(enclaveId)
	apiContainerAttributesProvider:= enclaveAttributesProvider.ForApiContainer()

	enclaveNamespaceName, err := backend.getEnclaveNamespaceName(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting namespace name for enclave with ID '%v'", enclaveId)
	}

	// Get Pod Attributes so that we can select them with the Service
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

	servicePorts, err := getKubernetesServicePortsFromPrivatePortSpecs(privatePortSpecs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating Kubernetes service ports from the API container's private port specs")
	}

	// Create Service BEFORE the pod so that the pod will know its own IP address
	apiContainerService, err := backend.kubernetesManager.CreateService(
		ctx,
		enclaveNamespaceName,
		apiContainerServiceName,
		apiContainerServiceLabels,
		apiContainerServiceAnnotations,
		apiContainerPodLabels,
		apiv1.ServiceTypeClusterIP,
		servicePorts,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the service with name '%s' in namespace '%s' with ports '%v' and '%v'", apiContainerServiceName, enclaveNamespaceName, grpcPortInt32, grpcProxyPortInt32)
	}
	var shouldRemoveService = true
	defer func() {
		if shouldRemoveService {
			if err := backend.kubernetesManager.RemoveService(ctx, apiContainerService); err != nil {
				logrus.Errorf("Creating the api container didn't complete successfully, so we tried to delete Kubernetes service '%v' that we created but an error was thrown:\n%v", apiContainerServiceName, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove Kubernetes service with name '%v'!!!!!!!", apiContainerServiceName)
			}
		}
	}()

	if _, found := customEnvVars[ownIpAddressEnvVar]; found {
		return nil, stacktrace.NewError("Requested own IP environment variable '%v' conflicts with a custom environment variable", ownIpAddressEnvVar)
	}
	envVarsWithOwnIp := map[string]string{
		ownIpAddressEnvVar: apiContainerService.Spec.ClusterIP,
	}
	for key, value := range customEnvVars {
		envVarsWithOwnIp[key] = value
	}

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
	apiContainerServiceAccount, err := backend.kubernetesManager.CreateServiceAccount(ctx, serviceAccountName, enclaveNamespaceName, serviceAccountLabels)
	if err != nil {
		return nil,  stacktrace.Propagate(err, "An error occurred creating service account '%v' with labels '%+v' in namespace '%v'", serviceAccountName, serviceAccountLabels, enclaveNamespaceName)
	}
	apiContainerServiceAccountName := apiContainerServiceAccount.GetName()
	shouldRemoveServiceAccount := true
	defer func() {
		if shouldRemoveServiceAccount {
			if err := backend.kubernetesManager.RemoveServiceAccount(ctx, apiContainerServiceAccount); err != nil {
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
			Verbs: []string{
				consts.CreateKubernetesVerb,
				consts.UpdateKubernetesVerb,
				consts.PatchKubernetesVerb,
				consts.DeleteKubernetesVerb,
				consts.GetKubernetesVerb,
				consts.ListKubernetesVerb,
				consts.WatchKubernetesVerb,
			},
			APIGroups: []string{rbacv1.APIGroupAll},
			Resources: []string{
				consts.PodsKubernetesResource,
				consts.PodExecsKubernetesResource,
				consts.PodLogsKubernetesResource,
				consts.ServicesKubernetesResource,
				consts.PersistentVolumeClaimsKubernetesResource,
				consts.JobsKubernetesResource,
			},
		},
		{
			// Necessary for the API container to get its own namespace
			Verbs: []string{consts.GetKubernetesVerb},
			APIGroups: []string{rbacv1.APIGroupAll},
			Resources: []string{consts.NamespacesKubernetesResource},
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
			if err := backend.kubernetesManager.RemoveRole(ctx, apiContainerRole); err != nil {
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
			if err := backend.kubernetesManager.RemoveRoleBindings(ctx, apiContainerRoleBinding); err != nil {
				logrus.Errorf("Creating the API container didn't complete successfully, so we tried to delete role binding '%v' in namespace '%v' that we created but an error was thrown:\n%v", roleBindingName, enclaveNamespaceName, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove role binding with name '%v'!!!!!!!", roleBindingName)
			}
		}
	}()

	// Create the Pod
	/*
	enclaveDataPersistentVolumeClaim, err := backend.getEnclaveDataPersistentVolumeClaim(ctx, enclaveNamespaceName, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the enclave data persistent volume claim for enclave '%v' in namespace '%v'", enclaveId, enclaveNamespaceName)
	}
	 */

	containerPorts, err := getKubernetesContainerPortsFromPrivatePortSpecs(privatePortSpecs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting container ports from the API container's private port specs")
	}

	// apiContainerContainers, apiContainerVolumes, err := getApiContainerContainersAndVolumes(image, containerPorts, envVarsWithOwnIp, enclaveDataPersistentVolumeClaim, enclaveDataVolumeDirpath)
	apiContainerContainers, apiContainerVolumes, err := getApiContainerContainersAndVolumes(image, containerPorts, envVarsWithOwnIp, enclaveDataVolumeDirpath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting API containers and volumes")
	}

	apiContainerInitContainers := []apiv1.Container{}

	// Create pods with api container containers and volumes in Kubernetes
	apiContainerPod, err := backend.kubernetesManager.CreatePod(
		ctx,
		enclaveNamespaceName,
		apiContainerPodName,
		apiContainerPodLabels,
		apiContainerPodAnnotations,
		apiContainerInitContainers,
		apiContainerContainers,
		apiContainerVolumes,
		apiContainerServiceAccountName,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the pod with name '%s' in namespace '%s' with image '%s'", apiContainerPodName, enclaveNamespaceName, image)
	}
	var shouldRemovePod = true
	defer func() {
		if shouldRemovePod {
			if err := backend.kubernetesManager.RemovePod(ctx, apiContainerPod); err != nil {
				logrus.Errorf("Creating the api container didn't complete successfully, so we tried to delete Kubernetes pod '%v' that we created but an error was thrown:\n%v", apiContainerPodName, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove Kubernetes pod with name '%v'!!!!!!!", apiContainerPodName)
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

	if err := waitForPortAvailabilityUsingNetstat(
		backend.kubernetesManager,
		enclaveNamespaceName,
		apiContainerPodName,
		kurtosisApiContainerContainerName,
		privateGrpcPortSpec,
		maxWaitForApiContainerContainerAvailabilityRetries,
		timeBetweenWaitForApiContainerContainerAvailabilityRetries,
	); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the API container grpc port '%v/%v' to become available", privateGrpcPortSpec.GetProtocol(), privateGrpcPortSpec.GetNumber())
	}

	if err := waitForPortAvailabilityUsingNetstat(
		backend.kubernetesManager,
		enclaveNamespaceName,
		apiContainerPodName,
		kurtosisApiContainerContainerName,
		privateGrpcProxyPortSpec,
		maxWaitForApiContainerContainerAvailabilityRetries,
		timeBetweenWaitForApiContainerContainerAvailabilityRetries,
	); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the API container grpc proxy port '%v/%v' to become available", privateGrpcProxyPortSpec.GetProtocol(), privateGrpcProxyPortSpec.GetNumber())
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
		kubernetesPod := resources.pod
		if kubernetesPod != nil {
			podName := kubernetesPod.GetName()
			namespaceName := kubernetesPod.GetNamespace()
			if err := backend.kubernetesManager.RemovePod(ctx, kubernetesPod); err != nil {
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

		kubernetesService := resources.service
		if kubernetesService != nil {
			serviceName := kubernetesService.GetName()
			namespaceName := kubernetesService.GetNamespace()
			updateConfigurator := func(updatesToApply *applyconfigurationsv1.ServiceApplyConfiguration) {
				specUpdates := applyconfigurationsv1.ServiceSpec().WithSelector(nil)
				updatesToApply.WithSpec(specUpdates)
			}
			if _, err := backend.kubernetesManager.UpdateService(ctx, namespaceName, serviceName, updateConfigurator); err != nil {
				erroredEnclaveIds[enclaveId] = stacktrace.Propagate(
					err,
					"An error occurred removing selectors from service '%v' in namespace '%v' for API container in enclave with ID '%v'",
					kubernetesService.Name,
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

		// Remove Pod
		if resources.pod != nil {
			podName := resources.pod.GetName()
			if err := backend.kubernetesManager.RemovePod(ctx, resources.pod); err != nil {
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
			if err := backend.kubernetesManager.RemoveRoleBindings(ctx, resources.roleBinding); err != nil {
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
			if err := backend.kubernetesManager.RemoveRole(ctx, resources.role); err != nil {
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
			if err := backend.kubernetesManager.RemoveServiceAccount(ctx, resources.serviceAccount); err != nil {
				erroredEnclaveIds[enclaveId] = stacktrace.Propagate(
					err,
					"An error occurred removing service account '%v' for API container in enclave with ID '%v'",
					serviceAccountName,
					enclaveId,
				)
				continue
			}
		}

		// Remove Service
		if resources.service != nil {
			serviceName := resources.service.GetName()
			if err := backend.kubernetesManager.RemoveService(ctx, resources.service); err != nil {
				erroredEnclaveIds[enclaveId] = stacktrace.Propagate(
					err,
					"An error occurred removing service '%v' for API container in enclave with ID '%v'",
					serviceName,
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
		label_key_consts.EnclaveIDKubernetesLabelKey.GetString(),
		enclaveIdsStrSet,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave namespaces matching IDs '%+v'", enclaveIdsStrSet)
	}

	apiContainerMatchLabels := getApiContainerMatchLabels()

	// Per-namespace objects
	result := map[enclave.EnclaveID]*apiContainerKubernetesResources{}
	for enclaveIdStr, namespacesForEnclaveId := range namespaces {
		if len(namespacesForEnclaveId) > 1 {
			return nil, stacktrace.NewError(
				"Expected at most one namespace to match enclave ID '%v', but got '%v'",
				enclaveIdStr,
				len(namespacesForEnclaveId),
			)
		}
		namespaceName := namespacesForEnclaveId[0].GetName()

		// Services (canonical defining resource for an API container)
		// TODO switch to GetSerivcesByLabels since we're already filtering on encalve ID by virtue of passing in namespace
		services, err := kubernetes_resource_collectors.CollectMatchingServices(
			ctx,
			backend.kubernetesManager,
			namespaceName,
			apiContainerMatchLabels,
			label_key_consts.EnclaveIDKubernetesLabelKey.GetString(),
			map[string]bool{
				enclaveIdStr: true,
			},
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting services matching enclave ID '%v' in namespace '%v'", enclaveIdStr, namespaceName)
		}

		servicesForEnclaveId, found := services[enclaveIdStr]
		if !found {
			// No API container services in the enclave means that the enclave doesn't have an API container
			continue
		}
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
		service := servicesForEnclaveId[0]

		// Pods
		pods, err := kubernetes_resource_collectors.CollectMatchingPods(
			ctx,
			backend.kubernetesManager,
			namespaceName,
			apiContainerMatchLabels,
			label_key_consts.EnclaveIDKubernetesLabelKey.GetString(),
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

		//Role Bindings
		roleBindings, err := kubernetes_resource_collectors.CollectMatchingRoleBindings(
			ctx,
			backend.kubernetesManager,
			namespaceName,
			apiContainerMatchLabels,
			label_key_consts.EnclaveIDKubernetesLabelKey.GetString(),
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
			label_key_consts.EnclaveIDKubernetesLabelKey.GetString(),
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
			label_key_consts.EnclaveIDKubernetesLabelKey.GetString(),
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

		enclaveId := enclave.EnclaveID(enclaveIdStr)

		result[enclaveId] = &apiContainerKubernetesResources{
			service:        service,
			pod:            pod,
			role:           role,
			roleBinding:    roleBinding,
			serviceAccount: serviceAccount,
		}
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
		kubernetesService := resourcesForEnclaveId.service
		if kubernetesService == nil {
			return nil, stacktrace.NewError("Expected a Kubernetes service for API container in enclave '%v'", enclaveId)
		}

		status, err := getContainerStatusFromPod(resourcesForEnclaveId.pod)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting container status from Kubernetes pod '%+v'", resourcesForEnclaveId.pod)
		}

		privateIpAddr := net.ParseIP(resourcesForEnclaveId.service.Spec.ClusterIP)
		if privateIpAddr == nil {
			return nil, stacktrace.NewError("Expected to be able to get the cluster ip of the API container service, instead parsing the cluster ip of service '%v' returned nil", resourcesForEnclaveId.service.Name)
		}

		privatePorts, err := getPrivatePortsAndValidatePortExistence(
			kubernetesService,
			map[string]bool{
				kurtosisInternalContainerGrpcPortSpecId: true,
				kurtosisInternalContainerGrpcProxyPortSpecId: true,
			},
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred parsing the API container private port specs and validating gRPC and gRPC proxy port existence")
		}
		privateGrpcPortSpec := privatePorts[kurtosisInternalContainerGrpcPortSpecId]
		privateGrpcProxyPortSpec := privatePorts[kurtosisInternalContainerGrpcProxyPortSpecId]

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
	// enclaveDataPersistentVolumeClaim *apiv1.PersistentVolumeClaim,
	enclaveDataVolumeDirpath string,
) (
	resultContainers []apiv1.Container,
	resultPodVolumes []apiv1.Volume,
	resultErr error,
) {
	if _, found := envVars[ApiContainerOwnNamespaceNameEnvVar]; found {
		return nil, nil, stacktrace.NewError("The environment variable that will contain the API container's own namespace name, '%v', conflicts with an existing environment variable", ApiContainerOwnNamespaceNameEnvVar)
	}

	var containerEnvVars []apiv1.EnvVar
	for varName, varValue := range envVars {
		envVar := apiv1.EnvVar{
			Name:  varName,
			Value: varValue,
		}
		containerEnvVars = append(containerEnvVars, envVar)
	}

	// Using the Kubernetes downward API to tell the API container about its own namespace name
	ownNamespaceEnvVar := apiv1.EnvVar{
		Name: ApiContainerOwnNamespaceNameEnvVar,
		ValueFrom: &apiv1.EnvVarSource{
			FieldRef: &apiv1.ObjectFieldSelector{
				 FieldPath: kubernetesResourceOwnNamespaceFieldPath,
			},
		},
	}
	containerEnvVars = append(containerEnvVars, ownNamespaceEnvVar)

	containers := []apiv1.Container{
		{
			Name:  kurtosisApiContainerContainerName,
			Image: containerImageAndTag,
			Env:   containerEnvVars,
			Ports: containerPorts,
			VolumeMounts: []apiv1.VolumeMount{
				{
					// Name:      enclaveDataPersistentVolumeClaim.Spec.VolumeName,
					Name:      enclaveDataDirVolumeName,
					MountPath: enclaveDataVolumeDirpath,
				},
			},
		},
	}

	volumes := []apiv1.Volume{
		{
			// Name: enclaveDataPersistentVolumeClaim.Spec.VolumeName,
			Name: enclaveDataDirVolumeName,
			VolumeSource: apiv1.VolumeSource{
				EmptyDir: &apiv1.EmptyDirVolumeSource{},
			},
		},
	}

	return containers, volumes, nil
}

func getApiContainerMatchLabels() map[string]string {
	engineMatchLabels := map[string]string{
		label_key_consts.AppIDKubernetesLabelKey.GetString():                label_value_consts.AppIDKubernetesLabelValue.GetString(),
		label_key_consts.KurtosisResourceTypeKubernetesLabelKey.GetString(): label_value_consts.APIContainerKurtosisResourceTypeKubernetesLabelValue.GetString(),
	}
	return engineMatchLabels
}

/*
func (backend *KubernetesKurtosisBackend) getEnclaveDataPersistentVolumeClaim(ctx context.Context, enclaveNamespaceName string, enclaveId enclave.EnclaveID) (*apiv1.PersistentVolumeClaim, error) {
	matchLabels := getEnclaveMatchLabels()
	matchLabels[label_key_consts.KurtosisVolumeTypeKubernetesLabelKey.GetString()] = label_value_consts.EnclaveDataVolumeTypeKubernetesLabelValue.GetString()
	matchLabels[label_key_consts.EnclaveIDKubernetesLabelKey.GetString()] = string(enclaveId)

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


 */