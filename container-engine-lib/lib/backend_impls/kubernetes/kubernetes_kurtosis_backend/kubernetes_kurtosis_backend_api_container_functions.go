package kubernetes_kurtosis_backend

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/shared_helpers"
	kubernetes_manager_consts "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager/consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_resource_collectors"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/api_container"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
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

	enclaveDataDirVolumeSize int64 = 1 * 1024 * 1024 * 1024 // 1g minimum size on Kubernetes
)

var noWait *port_spec.Wait = nil

// TODO add support for passing toleration to APIC
var noTolerations []apiv1.Toleration = nil

// TODO: MIGRATE THIS FOLDER TO USE STRUCTURE OF USER_SERVICE_FUNCTIONS MODULE

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
	enclaveId enclave.EnclaveUUID,
	grpcPortNum uint16,
	enclaveDataVolumeDirpath string,
	ownIpAddressEnvVar string,
	customEnvVars map[string]string,
	shouldStartInDebugMode bool,
) (
	*api_container.APIContainer,
	error,
) {
	grpcPortInt32 := int32(grpcPortNum)

	// TODO This validation is the same for Docker and for Kubernetes because we are using kurtBackend
	// TODO we could move this to a top layer for validations, perhaps
	// Verify no API container already exists in the enclave
	apiContainersInEnclaveFilters := &api_container.APIContainerFilters{
		EnclaveIDs: map[enclave.EnclaveUUID]bool{
			enclaveId: true,
		},
		Statuses: nil,
	}
	preexistingApiContainersInEnclave, err := backend.GetAPIContainers(ctx, apiContainersInEnclaveFilters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred checking if API containers already exist in enclave '%v'", enclaveId)
	}
	if len(preexistingApiContainersInEnclave) > 0 {
		return nil, stacktrace.NewError("Found existing API container(s) in enclave '%v'; cannot start a new one", enclaveId)
	}

	privateGrpcPortSpec, err := port_spec.NewPortSpec(grpcPortNum, consts.KurtosisServersTransportProtocol, consts.HttpApplicationProtocol, noWait)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the API container's private grpc port spec object using number '%v' and protocol '%v'",
			grpcPortNum,
			consts.KurtosisServersTransportProtocol.String(),
		)
	}
	privatePortSpecs := map[string]*port_spec.PortSpec{
		consts.KurtosisInternalContainerGrpcPortSpecId: privateGrpcPortSpec,
	}

	enclaveAttributesProvider := backend.objAttrsProvider.ForEnclave(enclaveId)
	apiContainerAttributesProvider := enclaveAttributesProvider.ForApiContainer()

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
	apiContainerPodLabels := shared_helpers.GetStringMapFromLabelMap(apiContainerPodAttributes.GetLabels())
	apiContainerPodAnnotations := shared_helpers.GetStringMapFromAnnotationMap(apiContainerPodAttributes.GetAnnotations())

	// Get Service Attributes
	apiContainerServiceAttributes, err := apiContainerAttributesProvider.ForApiContainerService(
		consts.KurtosisInternalContainerGrpcPortSpecId,
		privateGrpcPortSpec)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting the API container service attributes using private grpc port spec '%+v'a",
			privateGrpcPortSpec,
		)
	}
	apiContainerServiceName := apiContainerServiceAttributes.GetName().GetString()
	apiContainerServiceLabels := shared_helpers.GetStringMapFromLabelMap(apiContainerServiceAttributes.GetLabels())
	apiContainerServiceAnnotations := shared_helpers.GetStringMapFromAnnotationMap(apiContainerServiceAttributes.GetAnnotations())

	// Define service ports. These hook up to ports on the containers running in the API container pod
	// Kubernetes will assign a public port number to them

	servicePorts, err := shared_helpers.GetKubernetesServicePortsFromPrivatePortSpecs(privatePortSpecs)
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
		errMsg := fmt.Sprintf("An error occurred while creating the service with name '%s' in namespace '%s' with ports '%v'", apiContainerServiceName, enclaveNamespaceName, grpcPortInt32)
		logrus.Errorf("%s. Error was:\n%s", errMsg, err)
		return nil, stacktrace.Propagate(err, errMsg)
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
			"Expected to be able to get API container attributes for a Kubernetes service account, "+
				"instead got a non-nil error",
		)
	}

	serviceAccountName := serviceAccountAttributes.GetName().GetString()
	serviceAccountLabels := shared_helpers.GetStringMapFromLabelMap(serviceAccountAttributes.GetLabels())
	apiContainerServiceAccount, err := backend.kubernetesManager.CreateServiceAccount(ctx, serviceAccountName, enclaveNamespaceName, serviceAccountLabels)
	if err != nil {
		errMsg := fmt.Sprintf("An error occurred creating service account '%v' with labels '%+v' in namespace '%v'", serviceAccountName, serviceAccountLabels, enclaveNamespaceName)
		logrus.Errorf("%s. Error was:\n%s", errMsg, err)
		return nil, stacktrace.Propagate(err, errMsg)
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

	//Create the cluster role
	clusterRolesAttributes, err := apiContainerAttributesProvider.ForApiContainerClusterRole()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Expected to be able to get API container attributes for a Kubernetes cluster role, "+
				"instead got a non-nil error",
		)
	}

	clusterRoleName := clusterRolesAttributes.GetName().GetString()
	clusterRoleLabels := shared_helpers.GetStringMapFromLabelMap(clusterRolesAttributes.GetLabels())
	clusterRolePolicyRules := []rbacv1.PolicyRule{
		{
			Verbs: []string{
				kubernetes_manager_consts.CreateKubernetesVerb,
				kubernetes_manager_consts.DeleteKubernetesVerb,
				kubernetes_manager_consts.GetKubernetesVerb,
			},
			APIGroups: []string{
				rbacv1.APIGroupAll,
			},
			Resources: []string{
				kubernetes_manager_consts.PersistentVolumesKubernetesResource,
			},
		},
		{
			// Necessary for the API container to list all nodes
			// Temporarily adding perms to list nodes here to check for mono-node deployment for persistent volumes.
			// TODO: remove once we support multi nodes deployments with persistent volumes
			Verbs: []string{
				kubernetes_manager_consts.ListKubernetesVerb,
			},
			APIGroups: []string{
				rbacv1.APIGroupAll,
			},
			Resources: []string{
				kubernetes_manager_consts.NodesKubernetesResource,
			},
		},
	}

	apiContainerClusterRole, err := backend.kubernetesManager.CreateClusterRoles(ctx, clusterRoleName, clusterRolePolicyRules, clusterRoleLabels)
	if err != nil {
		errMsg := fmt.Sprintf("An error occurred creating cluster role '%v' with policy rules '%+v' "+
			"and labels '%+v' in namespace '%v'", clusterRoleName, clusterRolePolicyRules, clusterRoleLabels, enclaveNamespaceName)
		logrus.Errorf("%s. Error was:\n%s", errMsg, err)
		return nil, stacktrace.Propagate(err, errMsg)
	}
	shouldRemoveClusterRole := true
	defer func() {
		if shouldRemoveClusterRole {
			if err := backend.kubernetesManager.RemoveClusterRole(ctx, apiContainerClusterRole); err != nil {
				logrus.Errorf("Creating the API container didn't complete successfully, so we tried to delete cluster role '%v' that we created but an error was thrown:\n%v", clusterRoleName, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove cluster role with name '%v'!!!!!!!", clusterRoleName)
			}
		}
	}()

	//Create the cluster role binding to join the service account with the cluster role
	clusterRoleBindingsAttributes, err := apiContainerAttributesProvider.ForApiContainerClusterRoleBindings()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Expected to be able to get API container attributes for a Kubernetes cluster role bindings, "+
				"instead got a non-nil error",
		)
	}

	clusterRoleBindingName := clusterRoleBindingsAttributes.GetName().GetString()
	clusterRoleBindingsLabels := shared_helpers.GetStringMapFromLabelMap(clusterRoleBindingsAttributes.GetLabels())
	clusterRoleBindingsSubjects := []rbacv1.Subject{
		{
			Kind:      rbacv1.ServiceAccountKind,
			Name:      serviceAccountName,
			Namespace: enclaveNamespaceName,
		},
	}

	clusterRoleBindingsRoleRef := rbacv1.RoleRef{
		APIGroup: kubernetes_manager_consts.RbacAuthorizationApiGroup,
		Kind:     kubernetes_manager_consts.ClusterRoleKubernetesResourceType,
		Name:     clusterRoleName,
	}

	apiContainerClusterRoleBinding, err := backend.kubernetesManager.CreateClusterRoleBindings(ctx, clusterRoleBindingName, clusterRoleBindingsSubjects, clusterRoleBindingsRoleRef, clusterRoleBindingsLabels)
	if err != nil {
		errMsg := fmt.Sprintf("An error occurred creating cluster role bindings '%v' with subjects "+
			"'%+v' and role ref '%+v' in namespace '%v'", clusterRoleBindingName, clusterRoleBindingsSubjects, clusterRoleBindingsRoleRef, enclaveNamespaceName)
		logrus.Errorf("%s. Error was:\n%s", errMsg, err)
		return nil, stacktrace.Propagate(err, errMsg)
	}
	shouldRemoveClusterRoleBinding := true
	defer func() {
		if shouldRemoveClusterRoleBinding {
			if err := backend.kubernetesManager.RemoveClusterRoleBindings(ctx, apiContainerClusterRoleBinding); err != nil {
				logrus.Errorf("Creating the API container didn't complete successfully, so we tried to delete cluster role binding '%v' that we created but an error was thrown:\n%v", clusterRoleBindingName, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove cluster role binding with name '%v'!!!!!!!", clusterRoleBindingName)
			}
		}
	}()

	//Create the role
	rolesAttributes, err := apiContainerAttributesProvider.ForApiContainerRole()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Expected to be able to get API container attributes for a Kubernetes role, "+
				"instead got a non-nil error",
		)
	}

	roleName := rolesAttributes.GetName().GetString()
	roleLabels := shared_helpers.GetStringMapFromLabelMap(rolesAttributes.GetLabels())
	rolePolicyRules := []rbacv1.PolicyRule{
		{
			Verbs: []string{
				kubernetes_manager_consts.CreateKubernetesVerb,
				kubernetes_manager_consts.UpdateKubernetesVerb,
				kubernetes_manager_consts.PatchKubernetesVerb,
				kubernetes_manager_consts.DeleteKubernetesVerb,
				kubernetes_manager_consts.GetKubernetesVerb,
				kubernetes_manager_consts.ListKubernetesVerb,
				kubernetes_manager_consts.WatchKubernetesVerb,
			},
			APIGroups: []string{rbacv1.APIGroupAll},
			Resources: []string{
				kubernetes_manager_consts.PodsKubernetesResource,
				kubernetes_manager_consts.PodExecsKubernetesResource,
				kubernetes_manager_consts.PodLogsKubernetesResource,
				kubernetes_manager_consts.ServicesKubernetesResource,
				kubernetes_manager_consts.JobsKubernetesResource,
				kubernetes_manager_consts.PersistentVolumeClaimsKubernetesResource,
				kubernetes_manager_consts.IngressesKubernetesResource,
			},
		},
		{
			// Necessary for the API container to get its own namespace
			Verbs:     []string{kubernetes_manager_consts.GetKubernetesVerb},
			APIGroups: []string{rbacv1.APIGroupAll},
			Resources: []string{kubernetes_manager_consts.NamespacesKubernetesResource},
		},
	}

	apiContainerRole, err := backend.kubernetesManager.CreateRole(ctx, roleName, enclaveNamespaceName, rolePolicyRules, roleLabels)
	if err != nil {
		errMsg := fmt.Sprintf("An error occurred creating role '%v' with policy rules '%+v' "+
			"and labels '%+v' in namespace '%v'", roleName, rolePolicyRules, roleLabels, enclaveNamespaceName)
		logrus.Errorf("%s. Error was:\n%s", errMsg, err)
		return nil, stacktrace.Propagate(err, errMsg)
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
			"Expected to be able to get API container attributes for a Kubernetes role bindings, "+
				"instead got a non-nil error",
		)
	}

	roleBindingName := roleBindingsAttributes.GetName().GetString()
	roleBindingsLabels := shared_helpers.GetStringMapFromLabelMap(roleBindingsAttributes.GetLabels())
	roleBindingsSubjects := []rbacv1.Subject{
		{
			Kind:      rbacv1.ServiceAccountKind,
			Name:      serviceAccountName,
			Namespace: enclaveNamespaceName,
		},
	}

	roleBindingsRoleRef := rbacv1.RoleRef{
		APIGroup: kubernetes_manager_consts.RbacAuthorizationApiGroup,
		Kind:     kubernetes_manager_consts.RoleKubernetesResourceType,
		Name:     roleName,
	}

	apiContainerRoleBinding, err := backend.kubernetesManager.CreateRoleBindings(ctx, roleBindingName, enclaveNamespaceName, roleBindingsSubjects, roleBindingsRoleRef, roleBindingsLabels)
	if err != nil {
		errMsg := fmt.Sprintf("An error occurred creating role bindings '%v' with subjects "+
			"'%+v' and role ref '%+v' in namespace '%v'", roleBindingName, roleBindingsSubjects, roleBindingsRoleRef, enclaveNamespaceName)
		logrus.Errorf("%s. Error was:\n%s", errMsg, err)
		return nil, stacktrace.Propagate(err, errMsg)
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

	containerPorts, err := shared_helpers.GetKubernetesContainerPortsFromPrivatePortSpecs(privatePortSpecs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting container ports from the API container's private port specs")
	}

	volumeAttrs, err := enclaveAttributesProvider.ForEnclaveDataDirVolume()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the labels for enclave data dir volume")
	}

	volumeLabelsStrs := map[string]string{}
	for key, value := range volumeAttrs.GetLabels() {
		volumeLabelsStrs[key.GetString()] = value.GetString()
	}
	if _, err = backend.kubernetesManager.CreatePersistentVolumeClaim(ctx, enclaveNamespaceName, enclaveDataDirVolumeName, volumeLabelsStrs, enclaveDataDirVolumeSize); err != nil {
		errMsg := fmt.Sprintf("An error occurred creating the persistent volume claim for enclave data dir volume for enclave '%s'", enclaveDataDirVolumeName)
		logrus.Errorf("%s. Error was:\n%s", errMsg, err)
		return nil, stacktrace.Propagate(err, errMsg)
	}
	shouldDeleteVolumeClaim := true

	defer func() {
		if !shouldDeleteVolumeClaim {
			return
		}
		if err := backend.kubernetesManager.RemovePersistentVolumeClaim(context.Background(), enclaveNamespaceName, enclaveDataDirVolumeName); err != nil {
			logrus.Warnf(
				"Creating pod didn't finish successfully - we tried removing the PVC %v but failed with error %v",
				enclaveDataDirVolumeName,
				err,
			)
			logrus.Warnf("You'll need to clean up volume claim '%v' manually!", enclaveDataDirVolumeName)
		}
	}()

	apiContainerContainers, apiContainerVolumes, err := getApiContainerContainersAndVolumes(image, containerPorts, envVarsWithOwnIp, enclaveDataVolumeDirpath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting API containers and volumes")
	}

	apiContainerInitContainers := []apiv1.Container{}

	// Data is always persistent we can always restart like Docker
	apiContainerRestartPolicy := apiv1.RestartPolicyOnFailure

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
		apiContainerRestartPolicy,
		noTolerations,
	)
	if err != nil {
		errMsg := fmt.Sprintf("An error occurred while creating the pod with name '%s' in namespace '%s' with image '%s'", apiContainerPodName, enclaveNamespaceName, image)
		logrus.Errorf("%s. Error was:\n%s", errMsg, err)
		return nil, stacktrace.Propagate(err, errMsg)
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
		role:           apiContainerRole,
		roleBinding:    apiContainerRoleBinding,
		serviceAccount: apiContainerServiceAccount,
		service:        apiContainerService,
		pod:            apiContainerPod,
	}
	apiContainerObjsById, err := getApiContainerObjectsFromKubernetesResources(map[enclave.EnclaveUUID]*apiContainerKubernetesResources{
		enclaveId: apiContainerResources,
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred converting the new API container's Kubernetes resources to API container objects")
	}
	resultApiContainer, found := apiContainerObjsById[enclaveId]
	if !found {
		return nil, stacktrace.NewError("Successfully converted the new API container's Kubernetes resources to an API container object, but the resulting map didn't have an entry for enclave ID '%v'", enclaveId)
	}

	if err := shared_helpers.WaitForPortAvailabilityUsingNetstat(
		backend.kubernetesManager,
		enclaveNamespaceName,
		apiContainerPodName,
		kurtosisApiContainerContainerName,
		privateGrpcPortSpec,
		maxWaitForApiContainerContainerAvailabilityRetries,
		timeBetweenWaitForApiContainerContainerAvailabilityRetries,
	); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the API container grpc port '%v/%v' to become available", privateGrpcPortSpec.GetTransportProtocol(), privateGrpcPortSpec.GetNumber())
	}

	shouldRemoveClusterRole = false
	shouldRemoveClusterRoleBinding = false
	shouldRemoveRoleBinding = false
	shouldRemoveRole = false
	shouldRemoveServiceAccount = false
	shouldRemovePod = false
	shouldRemoveService = false
	shouldDeleteVolumeClaim = false
	return resultApiContainer, nil
}

func (backend *KubernetesKurtosisBackend) GetAPIContainers(
	ctx context.Context,
	filters *api_container.APIContainerFilters,
) (
	map[enclave.EnclaveUUID]*api_container.APIContainer,
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
	map[enclave.EnclaveUUID]bool,
	map[enclave.EnclaveUUID]error,
	error,
) {
	_, matchingKubernetesResources, err := backend.getMatchingApiContainerObjectsAndKubernetesResources(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting API containers and Kubernetes resources matching filters '%+v'", filters)
	}

	successfulEnclaveIds := map[enclave.EnclaveUUID]bool{}
	erroredEnclaveIds := map[enclave.EnclaveUUID]error{}
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
	map[enclave.EnclaveUUID]bool,
	map[enclave.EnclaveUUID]error,
	error,
) {

	_, matchingResources, err := backend.getMatchingApiContainerObjectsAndKubernetesResources(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting API container Kubernetes resources matching filters: %+v", filters)
	}

	successfulEnclaveIds := map[enclave.EnclaveUUID]bool{}
	erroredEnclaveIds := map[enclave.EnclaveUUID]error{}
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
//
//	Private Helper Methods
//
// ====================================================================================================
func (backend *KubernetesKurtosisBackend) getMatchingApiContainerObjectsAndKubernetesResources(
	ctx context.Context,
	filters *api_container.APIContainerFilters,
) (
	map[enclave.EnclaveUUID]*api_container.APIContainer,
	map[enclave.EnclaveUUID]*apiContainerKubernetesResources,
	error,
) {
	matchingResources, err := backend.getMatchingApiContainerKubernetesResources(ctx, filters.EnclaveIDs)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting API container Kubernetes resources matching enclave UUIDs: %+v", filters.EnclaveIDs)
	}

	apiContainerObjects, err := getApiContainerObjectsFromKubernetesResources(matchingResources)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting API container objects from Kubernetes resources")
	}

	// Finally, apply the filters
	resultApiContainerObjs := map[enclave.EnclaveUUID]*api_container.APIContainer{}
	resultKubernetesResources := map[enclave.EnclaveUUID]*apiContainerKubernetesResources{}
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
func (backend *KubernetesKurtosisBackend) getMatchingApiContainerKubernetesResources(ctx context.Context, enclaveUuids map[enclave.EnclaveUUID]bool) (
	map[enclave.EnclaveUUID]*apiContainerKubernetesResources,
	error,
) {
	enclaveMatchLabels := getEnclaveMatchLabels()

	enclaveIdsStrSet := map[string]bool{}
	for enclaveId, booleanValue := range enclaveUuids {
		enclaveIdStr := string(enclaveId)
		enclaveIdsStrSet[enclaveIdStr] = booleanValue
	}

	// Namespaces
	namespaces, err := kubernetes_resource_collectors.CollectMatchingNamespaces(
		ctx,
		backend.kubernetesManager,
		enclaveMatchLabels,
		kubernetes_label_key.EnclaveUUIDKubernetesLabelKey.GetString(),
		enclaveIdsStrSet,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave namespaces matching UUIDs '%+v'", enclaveIdsStrSet)
	}

	apiContainerMatchLabels := getApiContainerMatchLabels()

	// Per-namespace objects
	result := map[enclave.EnclaveUUID]*apiContainerKubernetesResources{}
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
		// TODO switch to GetSerivcesByLabels since we're already filtering on enclave ID by virtue of passing in namespace
		services, err := kubernetes_resource_collectors.CollectMatchingServices(
			ctx,
			backend.kubernetesManager,
			namespaceName,
			apiContainerMatchLabels,
			kubernetes_label_key.EnclaveUUIDKubernetesLabelKey.GetString(),
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
				"Expected to find one API container service in namespace '%v' for enclave with ID '%v' "+
					"but none was found",
				namespaceName,
				enclaveIdStr,
			)
		}
		if len(servicesForEnclaveId) > 1 {
			return nil, stacktrace.NewError(
				"Expected at most one API container service in namespace '%v' for enclave with ID '%v' "+
					"but found '%v'",
				namespaceName,
				enclaveIdStr,
				len(servicesForEnclaveId),
			)
		}
		service := servicesForEnclaveId[0]

		// Pods
		pods, err := kubernetes_resource_collectors.CollectMatchingPods(
			ctx,
			backend.kubernetesManager,
			namespaceName,
			apiContainerMatchLabels,
			kubernetes_label_key.EnclaveUUIDKubernetesLabelKey.GetString(),
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
					"Expected to find one API container pod in namespace '%v' for enclave with ID '%v' "+
						"but none was found",
					namespaceName,
					enclaveIdStr,
				)
			}
			if len(podsForEnclaveId) > 1 {
				return nil, stacktrace.NewError(
					"Expected at most one API container pod in namespace '%v' for enclave with ID '%v' "+
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
			kubernetes_label_key.EnclaveUUIDKubernetesLabelKey.GetString(),
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
					"Expected to find one API container role binding in namespace '%v' for enclave with ID '%v' "+
						"but none was found",
					namespaceName,
					enclaveIdStr,
				)
			}
			if len(roleBindingsForEnclaveId) > 1 {
				return nil, stacktrace.NewError(
					"Expected at most one API container role binding in namespace '%v' for enclave with ID '%v' "+
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
			kubernetes_label_key.EnclaveUUIDKubernetesLabelKey.GetString(),
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
					"Expected to find one API container role in namespace '%v' for enclave with ID '%v' "+
						"but none was found",
					namespaceName,
					enclaveIdStr,
				)
			}
			if len(rolesForEnclaveId) > 1 {
				return nil, stacktrace.NewError(
					"Expected at most one API container role in namespace '%v' for enclave with ID '%v' "+
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
			kubernetes_label_key.EnclaveUUIDKubernetesLabelKey.GetString(),
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
					"Expected to find one API container service account in namespace '%v' for enclave with ID '%v' "+
						"but none was found",
					namespaceName,
					enclaveIdStr,
				)
			}
			if len(serviceAccountsForEnclaveId) > 1 {
				return nil, stacktrace.NewError(
					"Expected at most one API container service account in namespace '%v' for enclave with ID '%v' "+
						"but found '%v'",
					namespaceName,
					enclaveIdStr,
					len(serviceAccounts),
				)
			}
			serviceAccount = serviceAccountsForEnclaveId[0]
		}

		enclaveId := enclave.EnclaveUUID(enclaveIdStr)

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
	allResources map[enclave.EnclaveUUID]*apiContainerKubernetesResources,
) (
	map[enclave.EnclaveUUID]*api_container.APIContainer,
	error,
) {
	result := map[enclave.EnclaveUUID]*api_container.APIContainer{}

	for enclaveId, resourcesForEnclaveId := range allResources {
		kubernetesService := resourcesForEnclaveId.service
		if kubernetesService == nil {
			return nil, stacktrace.NewError("Expected a Kubernetes service for API container in enclave '%v'", enclaveId)
		}

		status, err := shared_helpers.GetContainerStatusFromPod(resourcesForEnclaveId.pod)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting container status from Kubernetes pod '%+v'", resourcesForEnclaveId.pod)
		}

		privateIpAddr := net.ParseIP(resourcesForEnclaveId.service.Spec.ClusterIP)
		if privateIpAddr == nil {
			return nil, stacktrace.NewError("Expected to be able to get the cluster ip of the API container service, instead parsing the cluster ip of service '%v' returned nil", resourcesForEnclaveId.service.Name)
		}

		privatePorts, err := shared_helpers.GetPrivatePortsAndValidatePortExistence(
			kubernetesService,
			map[string]bool{
				consts.KurtosisInternalContainerGrpcPortSpecId: true,
			},
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred parsing the API container private port specs and validating gRPC and gRPC proxy port existence")
		}
		privateGrpcPortSpec := privatePorts[consts.KurtosisInternalContainerGrpcPortSpecId]

		// NOTE: We set these to nil because in Kubernetes we have no way of knowing what the public info is!
		var publicIpAddr net.IP = nil
		var publicGrpcPortSpec *port_spec.PortSpec = nil

		apiContainerObj := api_container.NewAPIContainer(
			enclaveId,
			status,
			privateIpAddr,
			privateGrpcPortSpec,
			publicIpAddr,
			publicGrpcPortSpec,
			nil,
		)

		result[enclaveId] = apiContainerObj
	}
	return result, nil
}

func getApiContainerContainersAndVolumes(
	containerImageAndTag string,
	containerPorts []apiv1.ContainerPort,
	envVars map[string]string,
	enclaveDataVolumeDirpath string,
) (
	resultContainers []apiv1.Container,
	resultVolumes []apiv1.Volume,
	resultErr error,
) {
	if _, found := envVars[ApiContainerOwnNamespaceNameEnvVar]; found {
		return nil, nil, stacktrace.NewError("The environment variable that will contain the API container's own namespace name, '%v', conflicts with an existing environment variable", ApiContainerOwnNamespaceNameEnvVar)
	}

	var containerEnvVars []apiv1.EnvVar
	for varName, varValue := range envVars {
		envVar := apiv1.EnvVar{
			Name:      varName,
			Value:     varValue,
			ValueFrom: nil,
		}
		containerEnvVars = append(containerEnvVars, envVar)
	}

	// Using the Kubernetes downward API to tell the API container about its own namespace name
	ownNamespaceEnvVar := apiv1.EnvVar{
		Name:  ApiContainerOwnNamespaceNameEnvVar,
		Value: "",
		ValueFrom: &apiv1.EnvVarSource{
			FieldRef: &apiv1.ObjectFieldSelector{
				APIVersion: "",
				FieldPath:  kubernetesResourceOwnNamespaceFieldPath,
			},
			ResourceFieldRef: nil,
			ConfigMapKeyRef:  nil,
			SecretKeyRef:     nil,
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
					Name:      enclaveDataDirVolumeName,
					MountPath: enclaveDataVolumeDirpath,
				},
			},
		},
	}

	volumes := []apiv1.Volume{
		{
			Name: enclaveDataDirVolumeName,
			VolumeSource: apiv1.VolumeSource{
				HostPath:             nil,
				EmptyDir:             nil,
				GCEPersistentDisk:    nil,
				AWSElasticBlockStore: nil,
				GitRepo:              nil,
				Secret:               nil,
				NFS:                  nil,
				ISCSI:                nil,
				Glusterfs:            nil,
				PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
					ClaimName: enclaveDataDirVolumeName,
					ReadOnly:  false,
				},
				RBD:                  nil,
				FlexVolume:           nil,
				Cinder:               nil,
				CephFS:               nil,
				Flocker:              nil,
				DownwardAPI:          nil,
				FC:                   nil,
				AzureFile:            nil,
				ConfigMap:            nil,
				VsphereVolume:        nil,
				Quobyte:              nil,
				AzureDisk:            nil,
				PhotonPersistentDisk: nil,
				Projected:            nil,
				PortworxVolume:       nil,
				ScaleIO:              nil,
				StorageOS:            nil,
				CSI:                  nil,
				Ephemeral:            nil,
			},
		},
	}

	return containers, volumes, nil
}

func getApiContainerMatchLabels() map[string]string {
	engineMatchLabels := map[string]string{
		kubernetes_label_key.AppIDKubernetesLabelKey.GetString():                label_value_consts.AppIDKubernetesLabelValue.GetString(),
		kubernetes_label_key.KurtosisResourceTypeKubernetesLabelKey.GetString(): label_value_consts.APIContainerKurtosisResourceTypeKubernetesLabelValue.GetString(),
	}
	return engineMatchLabels
}
