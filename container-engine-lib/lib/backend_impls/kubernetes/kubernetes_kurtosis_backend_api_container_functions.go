package kubernetes

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager/consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/object_name_constants"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/api_container"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"net"
)

const (
	kurtosisApiContainerContainerName = "kurtosis-core_api"
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
	enclaveDataDirpathOnHostMachine string, // The dirpath on the API container where the enclave data volume should be mounted
	enclaveDataVolumeDirpath string,
	envVars map[string]string,
) (*api_container.APIContainer, error) {

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

	apiContainerAttributesProvider, err := backend.objAttrsProvider.ForApiContainer(enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err,"An error occurred getting the api container attributes provider using enclave ID '%v'", enclaveId)
	}

	enclaveNamespace, err := backend.getEnclaveNamespace(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave namespace for enclave with ID '%v'", enclaveId)
	}
	enclaveNamespaceName := enclaveNamespace.GetName()

	enclavePersistentVolumeClaim, err := backend.getEnclavePersistentVolumeClaim(ctx, enclaveNamespaceName, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the enclave persistent volume claim for enclave '%v'", enclaveId)
	}

	//Create api container's service account, roles and role bindings
	apiContainerServiceAccountName, removeAllApiContainerRoleBasedResourcesFunc, err := backend.createApiContainerRoleBasedResources(ctx, enclaveNamespaceName, apiContainerAttributesProvider)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating api container service account, roles and role bindings for api container for enclave with ID '%v' in namespace '%v'", enclaveId, enclaveNamespaceName)
	}
	shouldRemoveAllApiContainerRoleBasedResources := true
	defer func(){
		if shouldRemoveAllApiContainerRoleBasedResources {
			removeAllApiContainerRoleBasedResourcesFunc()
		}
	}()

	// Get Pod Attributes
	apiContainerPodAttributes, err := apiContainerAttributesProvider.ForApiContainerPod()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Expected to be able to get attributes for a kubernetes pod for api container in enclave with id '%v', instead got a non-nil error",
			enclaveId,
		)
	}
	apiContainerPodName := apiContainerPodAttributes.GetName().GetString()
	apiContainerPodLabels := getStringMapFromLabelMap(apiContainerPodAttributes.GetLabels())
	apiContainerPodAnnotations := getStringMapFromAnnotationMap(apiContainerPodAttributes.GetAnnotations())

	apiContainerContainers, apiContainerVolumes := getApiContainerContainersAndVolumes(image, envVars, enclavePersistentVolumeClaim, enclaveDataVolumeDirpath)
	// Create pods with api container containers and volumes in kubernetes
	_, err = backend.kubernetesManager.CreatePod(ctx, enclaveNamespaceName, apiContainerPodName, apiContainerPodLabels, apiContainerPodAnnotations, apiContainerContainers, apiContainerVolumes, apiContainerServiceAccountName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the pod with name '%s' in namespace '%s' with image '%s'", apiContainerPodName, enclaveNamespaceName, image)
	}
	var shouldRemovePod = true
	defer func() {
		if shouldRemovePod {
			if err := backend.kubernetesManager.RemovePod(ctx, enclaveNamespaceName, apiContainerPodName); err != nil {
				logrus.Errorf("Creating the api container didn't complete successfully, so we tried to delete kubernetes pod '%v' that we created but an error was thrown:\n%v", apiContainerPodName, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove kubernetes pod with name '%v'!!!!!!!", apiContainerPodName)
			}
		}
	}()

	// Get Service Attributes
	apiContainerServiceAttributes, err := apiContainerAttributesProvider.ForApiContainerService(object_name_constants.KurtosisInternalContainerGrpcPortName.GetString(), privateGrpcPortSpec, object_name_constants.KurtosisInternalContainerGrpcProxyPortName.GetString(), privateGrpcProxyPortSpec)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting the api container service attributes using private grpc port spec '%+v', and "+
				"private grpc proxy port spec '%v'",
			privateGrpcPortSpec,
			privateGrpcProxyPortSpec,
		)
	}
	apiContainerServiceName := apiContainerServiceAttributes.GetName().GetString()
	apiContainerServiceLabels := getStringMapFromLabelMap(apiContainerServiceAttributes.GetLabels())
	apiContainerServiceAnnotations := getStringMapFromAnnotationMap(apiContainerServiceAttributes.GetAnnotations())
	grpcPortInt32 := int32(grpcPortNum)
	grpcProxyPortInt32 := int32(grpcProxyPortNum)
	// Define service ports. These hook up to ports on the containers running in the api container pod
	// Kubernetes will assign a public port number to them
	servicePorts := []apiv1.ServicePort{
		{
			Name:     object_name_constants.KurtosisInternalContainerGrpcPortName.GetString(),
			Protocol: apiv1.ProtocolTCP,
			Port:     grpcPortInt32,
		},
		{
			Name:     object_name_constants.KurtosisInternalContainerGrpcProxyPortName.GetString(),
			Protocol: apiv1.ProtocolTCP,
			Port:     grpcProxyPortInt32,
		},
	}

	// Create Service
	service, err := backend.kubernetesManager.CreateService(ctx, engineNamespaceName, apiContainerServiceName, apiContainerServiceLabels, apiContainerServiceAnnotations, enginePodLabels, externalServiceType, servicePorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the service with name '%s' in namespace '%s' with ports '%v' and '%v'", apiContainerServiceName, engineNamespaceName, grpcPortInt32, grpcProxyPortInt32)
	}
	var shouldRemoveService = true
	defer func() {
		if shouldRemoveService {
			if err := backend.kubernetesManager.RemoveService(ctx, engineNamespaceName, apiContainerServiceName); err != nil {
				logrus.Errorf("Creating the engine didn't complete successfully, so we tried to delete kubernetes service '%v' that we created but an error was thrown:\n%v", apiContainerServiceName, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove kubernetes service with name '%v'!!!!!!!", apiContainerServiceName)
			}
		}
	}()

	service, err = backend.kubernetesManager.GetServiceByName(ctx, engineNamespaceName, service.Name)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the service with name '%v' in namespace '%v'", service.Name, engineNamespaceName)
	}

	// Use cluster IP as public IP
	clusterIp := net.ParseIP(service.Spec.ClusterIP)
	if clusterIp == nil {
		return nil, stacktrace.NewError("Expected to be able to parse cluster IP from the kubernetes spec for service '%v', instead nil was parsed.", service.Name)
	}

	publicGrpcPort, publicGrpcProxyPort, err := getGrpcAndGrpcProxyPortSpecsFromServicePorts(service.Spec.Ports)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to determine kurtosis port specs from kubernetes service '%v', instead a non-nil err was returned", service.Name)
	}

	resultEngine := engine.NewEngine(
		engineIdStr,
		container_status.ContainerStatus_Running,
		clusterIp, publicGrpcPort, publicGrpcProxyPort)

	shouldRemoveNamespace = false
	shouldRemovePod = false
	shouldRemoveService = false
	shouldRemoveAllEngineRoleBasedResources = false
	return resultEngine, nil

}

func (backend *KubernetesKurtosisBackend) GetAPIContainers(ctx context.Context, filters *api_container.APIContainerFilters) (map[enclave.EnclaveID]*api_container.APIContainer, error) {
	matchingApiContainers, err := backend.getMatchingApiContainers(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting API containers matching the following filters: %+v", filters)
	}

	matchingApiContainersByEnclaveID := map[enclave.EnclaveID]*api_container.APIContainer{}
	for _, apicObjs := range matchingApiContainers {
		for _, apicObj := range apicObjs {
			matchingApiContainersByEnclaveID[apicObj.GetEnclaveID()] = apicObj
		}
	}

	return matchingApiContainersByEnclaveID, nil
}

// ====================================================================================================
//                                     Private Helper Methods
// ====================================================================================================
// Gets api containers matching the search filters, indexed by their [namespace][service name]
func (backend *KubernetesKurtosisBackend) getMatchingApiContainers(ctx context.Context, filters *api_container.APIContainerFilters) (map[string]map[string]*api_container.APIContainer, error) {
	matchingApiContainers := map[string]map[string]*api_container.APIContainer{}
	apiContainersMatchLabels := map[string]string{
		label_key_consts.AppIDLabelKey.GetString():                label_value_consts.AppIDLabelValue.GetString(),
		label_key_consts.KurtosisResourceTypeLabelKey.GetString(): label_value_consts.APIContainerContainerTypeLabelValue.GetString(),
	}

	for enclaveId := range filters.EnclaveIDs {
		enclaveNamespace, err := backend.getEnclaveNamespace(ctx, enclaveId)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting the namespace for api container in enclave with ID '%v'", enclaveId)
		}
		enclaveNamespaceName := enclaveNamespace.GetName()

		serviceList, err := backend.kubernetesManager.GetServicesByLabels(ctx, enclaveNamespaceName, apiContainersMatchLabels)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting api container services using labels: %+v in namespace '%v'", apiContainersMatchLabels, enclaveNamespaceName)
		}

		for _, service := range serviceList.Items {
			apiContainerObj, err := getApiContainerObjectFromKubernetesService(service)
			if err != nil {
				return nil, stacktrace.Propagate(err, "Expected to be able to get a kurtosis api container object service from kubernetes service '%v', instead a non-nil error was returned", service.Name)
			}
			// If the EnclaveIDs filter is specified, drop api containers not matching it
			if filters.EnclaveIDs != nil && len(filters.EnclaveIDs) > 0 {
				if _, found := filters.EnclaveIDs[apiContainerObj.GetEnclaveID()]; !found {
					continue
				}
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

func getApiContainerObjectFromKubernetesService(service apiv1.Service) (*api_container.APIContainer, error) {
	enclaveId, found := service.Labels[label_key_consts.EnclaveIDLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected to be able to find label describing the enclave ID on service '%v' with label key '%v', but was unable to", service.Name, label_key_consts.EnclaveIDLabelKey.GetString())
	}

	status := getKurtosisStatusFromKubernetesService(service)
	var publicIpAddr net.IP
	var publicGrpcPortSpec *port_spec.PortSpec
	var publicGrpcProxyPortSpec *port_spec.PortSpec
	var privateIpAddr net.IP
	var privateGrpcPortSpec *port_spec.PortSpec
	var privateGrpcProxyPortSpec *port_spec.PortSpec
	if status == container_status.ContainerStatus_Running {
		publicIpAddr = net.ParseIP(service.Spec.ClusterIP)
		if publicIpAddr == nil {
			return nil, stacktrace.NewError("Expected to be able to get the cluster ip of the api container service, instead parsing the cluster ip of service '%v' returned nil", service.Name)
		}
		privateIpAddr = publicIpAddr //TODO I'm not really sure of it, I've to check it
		var portSpecError error
		publicGrpcPortSpec, publicGrpcProxyPortSpec, portSpecError = getGrpcAndGrpcProxyPortSpecsFromServicePorts(service.Spec.Ports)
		if portSpecError != nil {
			return nil, stacktrace.Propagate(portSpecError, "Expected to be able to determine api container grpc port specs from kubernetes service ports for api container in enclave with ID '%v', instead a non-nil error was returned", enclaveId)
		}
		privateGrpcPortSpec, privateGrpcProxyPortSpec = publicGrpcPortSpec, publicGrpcProxyPortSpec //TODO I'm not really sure of it, I've to check it
	}

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

func (backend *KubernetesKurtosisBackend) createApiContainerRoleBasedResources(
	ctx context.Context,
	namespace string,
	apiContainerAttributesProvider object_attributes_provider.KubernetesApiContainerObjectAttributesProvider,
) (
resultApiContainerServiceAccountName string,
resultRemoveAllApiContainerRoleBasedResourcesFunc func(),
resultErr error,
) {

	serviceAccountName, serviceAccountLabels, err := getApiContainerServiceAccountNameAndLabels(apiContainerAttributesProvider)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred getting service account name and labels in namespace '%v'", namespace)
	}

	if _, err = backend.kubernetesManager.CreateServiceAccount(ctx, serviceAccountName, namespace, serviceAccountLabels); err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred creating service account '%v' with labels '%+v' in namespace '%v'", serviceAccountName, serviceAccountLabels, namespace)
	}
	shouldRemoveServiceAccount := true
	removeServiceAccountFunc := func() {
		if err := backend.kubernetesManager.RemoveServiceAccount(ctx, serviceAccountName, namespace); err != nil {
			logrus.Errorf("Creating the api container didn't complete successfully, so we tried to delete kubernetes service account '%v' that we created but an error was thrown:\n%v", serviceAccountName, err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually remove kubernetes service account with name '%v'!!!!!!!", serviceAccountName)
		}
	}
	defer func () {
		if shouldRemoveServiceAccount {
			removeServiceAccountFunc()
		}
	}()

	roleName, roleLabels, err := getApiContainerRoleNameAndLabels(apiContainerAttributesProvider)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred getting api container role name and labels in namespace '%v'", namespace)
	}

	rolePolicyRules := []rbacv1.PolicyRule{
		{
			Verbs: []string{consts.CreateKubernetesVerb, consts.UpdateKubernetesVerb, consts.PatchKubernetesVerb, consts.DeleteKubernetesVerb, consts.GetKubernetesVerb, consts.ListKubernetesVerb, consts.WatchKubernetesVerb},
			APIGroups: []string{rbacv1.APIGroupAll},
			Resources: []string{consts.PodsKubernetesResource, consts.ServicesKubernetesResource, consts.PersistentVolumeClaimsKubernetesResource},
		},
	}

	if _, err = backend.kubernetesManager.CreateRole(ctx, roleName, namespace, rolePolicyRules, roleLabels); err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred creating role '%v' with policy rules '%+v' and labels '%+v' in namespace '%v'", roleName, rolePolicyRules, roleLabels, namespace)
	}
	shouldRemoveRole := true
	removeRoleFunc := func() {
		if err := backend.kubernetesManager.RemoveRole(ctx, namespace, roleName); err != nil {
			logrus.Errorf("Creating the api container didn't complete successfully, so we tried to delete kubernetes role '%v' that we created but an error was thrown:\n%v", roleName, err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually remove kubernetes role with name '%v'!!!!!!!", roleName)
		}
	}
	defer func () {
		if shouldRemoveRole {
			removeRoleFunc()
		}
	}()

	roleBindingsName, roleBindingsLabels, err := getApiContainerRoleBindingsNameAndLabels(apiContainerAttributesProvider, serviceAccountName, roleName)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred getting api container role bindings name and labels")
	}

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
		return "", nil, stacktrace.Propagate(err, "An error occurred creating role bindings '%v' with subjects '%+v' and role ref '%+v' in namespace '%v'", roleBindingsName, roleBindingsSubjects, roleBindingsRoleRef, namespace)
	}
	shouldRemoveRoleBindings := true
	removeRoleBindingsFunc := func() {
		if err := backend.kubernetesManager.RemoveRoleBindings(ctx, roleBindingsName, namespace); err != nil {
			logrus.Errorf("Creating the api container didn't complete successfully, so we tried to delete kubernetes role bindings '%v' that we created but an error was thrown:\n%v", roleBindingsName, err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually remove kubernetes role bindings with name '%v'!!!!!!!", roleBindingsName)
		}
	}
	defer func () {
		if shouldRemoveRoleBindings {
			removeRoleBindingsFunc()
		}
	}()

	shouldRemoveServiceAccount = false
	shouldRemoveRole = false
	shouldRemoveRoleBindings = false

	removeAllRoleBasedResourcesFunc := func() {
		removeServiceAccountFunc()
		removeRoleFunc()
		removeRoleBindingsFunc()
	}
	return serviceAccountName, removeAllRoleBasedResourcesFunc, nil
}

func getApiContainerServiceAccountNameAndLabels(apiContainerAttributesProvider object_attributes_provider.KubernetesApiContainerObjectAttributesProvider) (resultApiContainerServiceAccountName string, resultApiContainerServiceAccountLabels map[string]string, resultErr error) {
	serviceAccountAttributes, err := apiContainerAttributesProvider.ForApiContainerServiceAccount()
	if err != nil {
		return "", nil, stacktrace.Propagate(
			err,
			"Expected to be able to get api container attributes for a kubernetes service account, instead got a non-nil error",
		)
	}

	serviceAccountName := serviceAccountAttributes.GetName().GetString()
	serviceAccountLabels := getStringMapFromLabelMap(serviceAccountAttributes.GetLabels())

	return serviceAccountName, serviceAccountLabels, nil
}

func getApiContainerRoleNameAndLabels(apiContainerAttributesProvider object_attributes_provider.KubernetesApiContainerObjectAttributesProvider) (resultApiContainerRoleName string, resultApiContainerRoleLabels map[string]string, resultErr error) {
	rolesAttributes, err := apiContainerAttributesProvider.ForApiContainerRole()
	if err != nil {
		return "", nil, stacktrace.Propagate(
			err,
			"Expected to be able to get api container attributes for a kubernetes role, instead got a non-nil error",
		)
	}

	roleName := rolesAttributes.GetName().GetString()
	roleLabels := getStringMapFromLabelMap(rolesAttributes.GetLabels())

	return roleName, roleLabels, nil
}

func getApiContainerRoleBindingsNameAndLabels(apiContainerAttributesProvider object_attributes_provider.KubernetesApiContainerObjectAttributesProvider, serviceAccountName string, roleName string) (resultApiContainerRoleBindingsName string, resultApiContainerRoleBindingsLabels map[string]string, resultErr error) {
	roleBindingsAttributes, err := apiContainerAttributesProvider.ForApiContainerRoleBindings(serviceAccountName, roleName)
	if err != nil {
		return "", nil, stacktrace.Propagate(
			err,
			"Expected to be able to get api container attributes for a kubernetes role bindings, instead got a non-nil error",
		)
	}

	roleBindingsName := roleBindingsAttributes.GetName().GetString()
	roleBindingsLabels := getStringMapFromLabelMap(roleBindingsAttributes.GetLabels())

	return roleBindingsName, roleBindingsLabels, nil
}

func getApiContainerContainersAndVolumes(
	containerImageAndTag string,
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