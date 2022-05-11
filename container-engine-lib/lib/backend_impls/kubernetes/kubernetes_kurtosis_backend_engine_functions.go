package kubernetes

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager/consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_resource_collectors"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/object_name_constants"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"net"
	"time"
)

const (
	// The ID of the GRPC port for Kurtosis-internal containers (e.g. API container, engine, modules, etc.) which will
	//  be stored in the port spec label
	kurtosisInternalContainerGrpcPortSpecId = "grpc"

	// The ID of the GRPC proxy port for Kurtosis-internal containers. This is necessary because
	// Typescript's grpc-web cannot communicate directly with GRPC ports, so Kurtosis-internal containers
	// need a proxy  that will translate grpc-web requests before they hit the main GRPC server
	kurtosisInternalContainerGrpcProxyPortSpecId = "grpcProxy"

	// The engine server uses gRPC so MUST listen on TCP (no other protocols are supported), which also
	// means that its grpc-proxy must listen on TCP
	enginePortProtocol = port_spec.PortProtocol_TCP

	externalServiceType = "ClusterIP"
)

// Any of these values being nil indicates that the resource doesn't exist
type engineKubernetesResources struct {
	clusterRole *rbacv1.ClusterRole

	clusterRoleBinding *rbacv1.ClusterRoleBinding

	namespace *apiv1.Namespace

	// Should always be nil if namespace is nil
	serviceAccount *apiv1.ServiceAccount

	// Should always be nil if namespace is nil
	service *apiv1.Service

	// Should always be nil if namespace is nil
	pod *apiv1.Pod
}

// ====================================================================================================
//                                     Engine CRUD Methods
// ====================================================================================================

func (backend *KubernetesKurtosisBackend) CreateEngine(
	ctx context.Context,
	imageOrgAndRepo string,
	imageVersionTag string,
	grpcPortNum uint16,
	grpcProxyPortNum uint16,
	envVars map[string]string,
) (
	*engine.Engine,
	error,
) {

	containerStartTimeUnixSecs := time.Now().Unix()
	engineIdStr := fmt.Sprintf("%v", containerStartTimeUnixSecs)

	privateGrpcPortSpec, err := port_spec.NewPortSpec(grpcPortNum, enginePortProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the engine's private grpc port spec object using number '%v' and protocol '%v'",
			grpcPortNum,
			enginePortProtocol.String(),
		)
	}
	privateGrpcProxyPortSpec, err := port_spec.NewPortSpec(grpcProxyPortNum, enginePortProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the engine's private grpc proxy port spec object using number '%v' and protocol '%v'",
			grpcProxyPortNum,
			enginePortProtocol.String(),
		)
	}
	engineAttributesProvider, err := backend.objAttrsProvider.ForEngine(engineIdStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the engine attributes provider using id '%v'", engineIdStr)
	}

	// Get Namespace Attributes
	engineNamespaceAttributes, err := engineAttributesProvider.ForEngineNamespace()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Expected to be able to get attributes for a kubernetes namespace for engine with  id '%v', instead got a non-nil error",
			engineIdStr,
		)
	}
	engineNamespaceName := engineNamespaceAttributes.GetName().GetString()
	engineNamespaceLabels := getStringMapFromLabelMap(engineNamespaceAttributes.GetLabels())

	//Create engine's namespace
	engineNamespace, err := backend.kubernetesManager.CreateNamespace(ctx, engineNamespaceName, engineNamespaceLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the namespace '%v' using labels '%+v'", engineNamespace, engineNamespaceLabels)
	}
	shouldRemoveNamespace := true
	defer func() {
		if shouldRemoveNamespace {
			if err := backend.kubernetesManager.RemoveNamespace(ctx, engineNamespaceName); err != nil {
				logrus.Errorf("Creating the engine didn't complete successfully, so we tried to delete kubernetes namespace '%v' that we created but an error was thrown:\n%v", engineNamespaceName, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove kubernetes namespace with name '%v'!!!!!!!", engineNamespaceName)
			}
		}
	}()

	serviceAccountAttributes, err := engineAttributesProvider.ForEngineServiceAccount()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Expected to be able to get Kubernetes attributes for engine service account, instead got a non-nil error",
		)
	}
	serviceAccountName := serviceAccountAttributes.GetName().GetString()
	serviceAccountLabels := getStringMapFromLabelMap(serviceAccountAttributes.GetLabels())
	if _, err = backend.kubernetesManager.CreateServiceAccount(ctx, serviceAccountName, engineNamespaceName, serviceAccountLabels); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating service account '%v' with labels '%+v' in namespace '%v'", serviceAccountName, serviceAccountLabels, engineNamespaceName)
	}
	shouldRemoveServiceAccount := true
	defer func() {
		if shouldRemoveServiceAccount {
			if err := backend.kubernetesManager.RemoveServiceAccount(ctx, serviceAccountName, engineNamespaceName); err != nil {
				logrus.Errorf("Creating the engine didn't complete successfully, so we tried to delete service account '%v' in namespace '%v' that we created but an error was thrown:\n%v", serviceAccountName, engineNamespaceName, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove service account with name '%v'!!!!!!!", serviceAccountName)
			}
		}
	}()

	clusterRolesAttributes, err := engineAttributesProvider.ForEngineClusterRole()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Expected to be able to get Kubernetes attributes for a engine cluster role, instead got a non-nil error",
		)
	}
	clusterRoleName := clusterRolesAttributes.GetName().GetString()
	clusterRoleLabels := getStringMapFromLabelMap(clusterRolesAttributes.GetLabels())
	clusterRolePolicyRules := []rbacv1.PolicyRule{
		{
			Verbs:     []string{consts.CreateKubernetesVerb, consts.UpdateKubernetesVerb, consts.PatchKubernetesVerb, consts.DeleteKubernetesVerb, consts.GetKubernetesVerb, consts.ListKubernetesVerb, consts.WatchKubernetesVerb},
			APIGroups: []string{rbacv1.APIGroupAll},
			Resources: []string{consts.NamespacesKubernetesResource, consts.ServiceAccountsKubernetesResource, consts.RolesKubernetesResource, consts.RoleBindingsKubernetesResource, consts.PodsKubernetesResource, consts.ServicesKubernetesResource, consts.PersistentVolumeClaimsKubernetesResource},
		},
	}
	if _, err = backend.kubernetesManager.CreateClusterRoles(ctx, clusterRoleName, clusterRolePolicyRules, clusterRoleLabels); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating cluster role '%v' with policy rules '%+v' and labels '%+v'", clusterRoleName, clusterRolePolicyRules, clusterRoleLabels)
	}
	shouldRemoveClusterRole := true
	defer func() {
		if shouldRemoveClusterRole {
			if err := backend.kubernetesManager.RemoveClusterRole(ctx, clusterRoleName); err != nil {
				logrus.Errorf("Creating the engine didn't complete successfully, so we tried to delete cluster role '%v' that we created but an error was thrown:\n%v", clusterRoleName, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove cluster role with name '%v'!!!!!!!", clusterRoleName)
			}
		}
	}()

	clusterRoleBindingsAttributes, err := engineAttributesProvider.ForEngineClusterRoleBindings()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Expected to be able to get engine attributes for a kubernetes cluster role bindings, instead got a non-nil error",
		)
	}
	clusterRoleBindingsName := clusterRoleBindingsAttributes.GetName().GetString()
	clusterRoleBindingsLabels := getStringMapFromLabelMap(clusterRoleBindingsAttributes.GetLabels())
	clusterRoleBindingsSubjects := []rbacv1.Subject{
		{
			Kind:      rbacv1.ServiceAccountKind,
			Name:      serviceAccountName,
			Namespace: engineNamespaceName,
		},
	}
	clusterRoleBindingsRoleRef := rbacv1.RoleRef{
		APIGroup: consts.RbacAuthorizationApiGroup,
		Kind:     consts.ClusterRoleKubernetesResourceType,
		Name:     clusterRoleName,
	}
	if _, err := backend.kubernetesManager.CreateClusterRoleBindings(ctx, clusterRoleBindingsName, clusterRoleBindingsSubjects, clusterRoleBindingsRoleRef, clusterRoleBindingsLabels); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating cluster role bindings '%v' with subjects '%+v' and role ref '%+v' in namespace '%v'", clusterRoleBindingsName, clusterRoleBindingsSubjects, clusterRoleBindingsRoleRef, engineNamespaceName)
	}
	shouldRemoveClusterRoleBinding := true
	defer func() {
		if shouldRemoveClusterRoleBinding {
			if err := backend.kubernetesManager.RemoveClusterRoleBindings(ctx, clusterRoleBindingsName); err != nil {
				logrus.Errorf("Creating the engine didn't complete successfully, so we tried to delete cluster role bindings '%v' in namespace '%v' that we created but an error was thrown:\n%v", clusterRoleBindingsName, engineNamespaceName, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove cluster role bindings with name '%v'!!!!!!!", clusterRoleBindingsName)
			}
		}
	}()

	// Get Pod Attributes
	enginePodAttributes, err := engineAttributesProvider.ForEnginePod()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Expected to be able to get attributes for a kubernetes pod for engine with id '%v', instead got a non-nil error",
			engineIdStr,
		)
	}
	enginePodName := enginePodAttributes.GetName().GetString()
	enginePodLabels := getStringMapFromLabelMap(enginePodAttributes.GetLabels())
	enginePodAnnotations := getStringMapFromAnnotationMap(enginePodAttributes.GetAnnotations())

	// Define Containers in our Engine Pod and hook them up to our Engine Volumes
	containerImageAndTag := fmt.Sprintf(
		"%v:%v",
		imageOrgAndRepo,
		imageVersionTag,
	)
	engineContainers, engineVolumes := getEngineContainers(containerImageAndTag, envVars)
	// Create pods with engine containers and volumes in kubernetes
	_, err = backend.kubernetesManager.CreatePod(ctx, engineNamespaceName, enginePodName, enginePodLabels, enginePodAnnotations, engineContainers, engineVolumes, serviceAccountName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the pod with name '%s' in namespace '%s' with image '%s'", enginePodName, engineNamespaceName, containerImageAndTag)
	}
	var shouldRemovePod = true
	defer func() {
		if shouldRemovePod {
			if err := backend.kubernetesManager.RemovePod(ctx, engineNamespaceName, enginePodName); err != nil {
				logrus.Errorf("Creating the engine didn't complete successfully, so we tried to delete kubernetes pod '%v' that we created but an error was thrown:\n%v", enginePodName, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove kubernetes pod with name '%v'!!!!!!!", enginePodName)
			}
		}
	}()

	// Get Service Attributes
	engineServiceAttributes, err := engineAttributesProvider.ForEngineService(kurtosisInternalContainerGrpcPortSpecId, privateGrpcPortSpec, kurtosisInternalContainerGrpcProxyPortSpecId, privateGrpcProxyPortSpec)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting the engine service attributes using private grpc port spec '%+v', and "+
				"private grpc proxy port spec '%v'",
			privateGrpcPortSpec,
			privateGrpcProxyPortSpec,
		)
	}
	engineServiceName := engineServiceAttributes.GetName().GetString()
	engineServiceLabels := getStringMapFromLabelMap(engineServiceAttributes.GetLabels())
	engineServiceAnnotations := getStringMapFromAnnotationMap(engineServiceAttributes.GetAnnotations())
	grpcPortInt32 := int32(grpcPortNum)
	grpcProxyPortInt32 := int32(grpcProxyPortNum)
	// Define service ports. These hook up to ports on the containers running in the engine pod
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
	service, err := backend.kubernetesManager.CreateService(ctx, engineNamespaceName, engineServiceName, engineServiceLabels, engineServiceAnnotations, enginePodLabels, externalServiceType, servicePorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the service with name '%s' in namespace '%s' with ports '%v' and '%v'", engineServiceName, engineNamespaceName, grpcPortInt32, grpcProxyPortInt32)
	}
	var shouldRemoveService = true
	defer func() {
		if shouldRemoveService {
			if err := backend.kubernetesManager.RemoveService(ctx, engineNamespaceName, engineServiceName); err != nil {
				logrus.Errorf("Creating the engine didn't complete successfully, so we tried to delete kubernetes service '%v' that we created but an error was thrown:\n%v", engineServiceName, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove kubernetes service with name '%v'!!!!!!!", engineServiceName)
			}
		}
	}()

	service, err = backend.kubernetesManager.GetServiceByName(ctx, engineNamespaceName, service.Name)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the service with name '%v' in namespace '%v'", service.Name, engineNamespaceName)
	}

	// Left a nil because the KurtosisBackend has no way of knowing what the public port spec will be
	var publicIpAddr net.IP
	var publicGrpcPort *port_spec.PortSpec
	var publicGrpcProxyPort *port_spec.PortSpec

	resultEngine := engine.NewEngine(
		engineIdStr,
		container_status.ContainerStatus_Running,
		publicIpAddr,
		publicGrpcPort,
		publicGrpcProxyPort,
	)

	shouldRemoveNamespace = false
	shouldRemoveServiceAccount = false
	shouldRemoveClusterRole = false
	shouldRemoveClusterRoleBinding = false
	shouldRemovePod = false
	shouldRemoveService = false
	return resultEngine, nil
}

func (backend *KubernetesKurtosisBackend) GetEngines(ctx context.Context, filters *engine.EngineFilters) (map[string]*engine.Engine, error) {
	matchingEngines, _, err := backend.getMatchingEnginesAndKubernetesResources(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting engines matching the following filters: %+v", filters)
	}
	return matchingEngines, nil
}

func (backend *KubernetesKurtosisBackend) StopEngines(
	ctx context.Context,
	filters *engine.EngineFilters,
) (
	resultSuccessfulEngineIds map[string]bool,
	resultErroredEngineIds map[string]error,
	resultErr error,
) {
	_, matchingKubernetesResources, err := backend.getMatchingEnginesAndKubernetesResources(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting engines and Kubernetes resources matching filters '%+v'", filters)
	}

	successfulEngineIds := map[string]bool{}
	erroredEngineIds := map[string]error{}
	for engineId, resources := range matchingKubernetesResources {
		if resources.namespace == nil {
			// No namespace means nothing needs stopping
			successfulEngineIds[engineId] = true
			continue
		}
		namespaceName := resources.namespace.Name

		if resources.service != nil {
			serviceName := resources.service.Name
			if err := backend.kubernetesManager.RemoveSelectorsFromService(ctx, namespaceName, serviceName); err != nil {
				erroredEngineIds[engineId] = stacktrace.Propagate(
					err,
					"An error occurred removing selectors from service '%v' in namespace '%v' for engine '%v'",
					serviceName,
					namespaceName,
					engineId,
				)
				continue
			}
		}

		if resources.pod != nil {
			podName := resources.pod.Name
			if err := backend.kubernetesManager.RemovePod(ctx, namespaceName, podName); err != nil {
				erroredEngineIds[engineId] = stacktrace.Propagate(
					err,
					"An error occurred removing pod '%v' in namespace '%v' for engine '%v'",
					podName,
					namespaceName,
					engineId,
				)
				continue
			}
		}

		successfulEngineIds[engineId] = true
	}

	return successfulEngineIds, erroredEngineIds, nil
}

func (backend *KubernetesKurtosisBackend) DestroyEngines(
	ctx context.Context,
	filters *engine.EngineFilters,
) (
	resultSuccessfulEngineIds map[string]bool,
	resultErroredEngineIds map[string]error,
	resultErr error,
) {
	_, matchingResources, err := backend.getMatchingEnginesAndKubernetesResources(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting engine Kubernetes resources matching filters: %+v", filters)
	}

	successfulEngineIds := map[string]bool{}
	erroredEngineIds := map[string]error{}
	for engineId, resources := range matchingResources {
		// Remove ClusterRoleBinding
		if resources.clusterRoleBinding != nil {
			roleBindingName := resources.clusterRoleBinding.Name
			if err := backend.kubernetesManager.RemoveClusterRoleBindings(ctx, roleBindingName); err != nil {
				erroredEngineIds[engineId] = stacktrace.Propagate(
					err,
					"An error occurred removing cluster role binding '%v' for engine '%v'",
					roleBindingName,
					engineId,
				)
				continue
			}
		}

		// Remove ClusterRole
		if resources.clusterRole != nil {
			roleName := resources.clusterRole.Name
			if err := backend.kubernetesManager.RemoveClusterRole(ctx, roleName); err != nil {
				erroredEngineIds[engineId] = stacktrace.Propagate(
					err,
					"An error occurred removing cluster role '%v' for engine '%v'",
					roleName,
					engineId,
				)
				continue
			}
		}

		// Remove the namespace
		if resources.namespace != nil {
			namespaceName := resources.namespace.Name
			if err := backend.kubernetesManager.RemoveNamespace(ctx, namespaceName); err != nil {
				erroredEngineIds[engineId] = stacktrace.Propagate(
					err,
					"An error occurred removing namespace '%v' for engine '%v'",
					namespaceName,
					engineId,
				)
				continue
			}
		}

		successfulEngineIds[engineId] = true
	}
	return successfulEngineIds, erroredEngineIds, nil
}

// ====================================================================================================
//                                     Private Helper Methods
// ====================================================================================================
func (backend *KubernetesKurtosisBackend) getMatchingEnginesAndKubernetesResources(
	ctx context.Context,
	filters *engine.EngineFilters,
) (
	map[string]*engine.Engine,
	map[string]*engineKubernetesResources,
	error,
) {
	matchingResources, err := backend.getMatchingKubernetesResources(ctx, filters.IDs)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting engine Kubernetes resources matching IDs: %+v", filters.IDs)
	}

	engineObjects, err := getEngineObjectsFromKubernetesResources(matchingResources)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting engine objects from Kubernetes resources")
	}

	// Finally, apply the filters
	resultEngineObjs := map[string]*engine.Engine{}
	resultKubernetesResources := map[string]*engineKubernetesResources{}
	for engineId, engineObj := range engineObjects {
		if filters.IDs != nil && len(filters.IDs) > 0 {
			if _, found := filters.IDs[engineObj.GetID()]; !found {
				continue
			}
		}

		if filters.Statuses != nil && len(filters.IDs) > 0 {
			if _, found := filters.Statuses[engineObj.GetStatus()]; !found {
				continue
			}
		}

		resultEngineObjs[engineId] = engineObj
		// Okay to do because we're guaranteed a 1:1 mapping between engine_obj:engine_resources
		resultKubernetesResources[engineId] = matchingResources[engineId]
	}

	return resultEngineObjs, resultKubernetesResources, nil
}

// Get back any and all Kubernetes resources matching the given IDs, where a nil or empty map == "match all IDs"
func (backend *KubernetesKurtosisBackend) getMatchingKubernetesResources(ctx context.Context, engineIds map[string]bool) (
	map[string]*engineKubernetesResources,
	error,
) {
	engineMatchLabels := getEngineMatchLabels()

	result := map[string]*engineKubernetesResources{}

	// Namespaces
	namespaces, err := kubernetes_resource_collectors.CollectMatchingNamespaces(
		ctx,
		backend.kubernetesManager,
		engineMatchLabels,
		label_key_consts.IDLabelKey.GetString(),
		engineIds,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting engine namespaces matching IDs '%+v'", engineIds)
	}
	for engineId, namespacesForId := range namespaces {
		if len(namespacesForId) > 1 {
			return nil, stacktrace.NewError(
				"Expected at most one namespace to match engine ID '%v', but got '%v'",
				len(namespacesForId),
				engineId,
			)
		}
		engineResources, found := result[engineId]
		if !found {
			engineResources = &engineKubernetesResources{}
		}
		engineResources.namespace = namespacesForId[0]
		result[engineId] = engineResources
	}

	// Cluster roles
	clusterRoles, err := kubernetes_resource_collectors.CollectMatchingClusterRoles(
		ctx,
		backend.kubernetesManager,
		engineMatchLabels,
		label_key_consts.IDLabelKey.GetString(),
		engineIds,
	)
	for engineId, clusterRolesForId := range clusterRoles {
		if len(clusterRolesForId) > 1 {
			return nil, stacktrace.NewError(
				"Expected at most one cluster role to match engine ID '%v', but got '%v'",
				len(clusterRolesForId),
				engineId,
			)
		}
		engineResources, found := result[engineId]
		if !found {
			engineResources = &engineKubernetesResources{}
		}
		engineResources.clusterRole = clusterRolesForId[0]
		result[engineId] = engineResources
	}

	// Cluster role bindings
	clusterRoleBindings, err := kubernetes_resource_collectors.CollectMatchingClusterRoleBindings(
		ctx,
		backend.kubernetesManager,
		engineMatchLabels,
		label_key_consts.IDLabelKey.GetString(),
		engineIds,
	)
	for engineId, clusterRoleBindingsForId := range clusterRoleBindings {
		if len(clusterRoleBindingsForId) > 1 {
			return nil, stacktrace.NewError(
				"Expected at most one cluster role binding to match engine ID '%v', but got '%v'",
				len(clusterRoleBindingsForId),
				engineId,
			)
		}
		engineResources, found := result[engineId]
		if !found {
			engineResources = &engineKubernetesResources{}
		}
		engineResources.clusterRoleBinding = clusterRoleBindingsForId[0]
		result[engineId] = engineResources
	}

	// Per-namespace objects
	for engineId, engineResources := range result {
		if engineResources.namespace == nil {
			continue
		}
		namespaceName := engineResources.namespace.Name

		// Service accounts
		serviceAccounts, err := kubernetes_resource_collectors.CollectMatchingServiceAccounts(
			ctx,
			backend.kubernetesManager,
			namespaceName,
			engineMatchLabels,
			label_key_consts.IDLabelKey.GetString(),
			map[string]bool{
				engineId: true,
			},
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting service accounts matching engine ID '%v' in namespace '%v'", engineId, namespaceName)
		}
		var serviceAccount *apiv1.ServiceAccount
		if serviceAccountsForId, found := serviceAccounts[engineId]; found {
			if len(serviceAccountsForId) > 1 {
				return nil, stacktrace.NewError(
					"Expected at most one engine service account in namespace '%v' for engine with ID '%v' " +
						"but found '%v'",
					namespaceName,
					engineId,
					len(serviceAccounts),
				)
			}
			serviceAccount = serviceAccountsForId[0]
		}

		// Services
		services, err := kubernetes_resource_collectors.CollectMatchingServices(
			ctx,
			backend.kubernetesManager,
			namespaceName,
			engineMatchLabels,
			label_key_consts.IDLabelKey.GetString(),
			map[string]bool{
				engineId: true,
			},
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting services matching engine ID '%v' in namespace '%v'", engineId, namespaceName)
		}
		var service *apiv1.Service
		if servicesForId, found := services[engineId]; found {
			if len(servicesForId) > 1 {
				return nil, stacktrace.NewError(
					"Expected at most one engine service in namespace '%v' for engine with ID '%v' " +
						"but found '%v'",
					namespaceName,
					engineId,
					len(services),
				)
			}
			service = servicesForId[0]
		}

		// Pods
		pods, err := kubernetes_resource_collectors.CollectMatchingPods(
			ctx,
			backend.kubernetesManager,
			namespaceName,
			engineMatchLabels,
			label_key_consts.IDLabelKey.GetString(),
			map[string]bool{
				engineId: true,
			},
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting pods matching engine ID '%v' in namespace '%v'", engineId, namespaceName)
		}
		var pod *apiv1.Pod
		if podsForId, found := pods[engineId]; found {
			if len(podsForId) > 1 {
				return nil, stacktrace.NewError(
					"Expected at most one engine pod in namespace '%v' for engine with ID '%v' " +
						"but found '%v'",
					namespaceName,
					engineId,
					len(pods),
				)
			}
			pod = podsForId[0]
		}

		engineResources.service = service
		engineResources.pod = pod
		engineResources.serviceAccount = serviceAccount
	}

	return result, nil
}

func getEngineObjectsFromKubernetesResources(allResources map[string]*engineKubernetesResources) (map[string]*engine.Engine, error) {
	result := map[string]*engine.Engine{}

	for engineId, resourcesForId := range allResources {

		engineStatus := container_status.ContainerStatus_Stopped
		if resourcesForId.pod != nil {
			engineStatus = container_status.ContainerStatus_Running
		}
		if resourcesForId.service != nil && len(resourcesForId.service.Spec.Selector) > 0 {
			engineStatus = container_status.ContainerStatus_Running
		}

		// NOTE: We set these to nil because in Kubernetes we have no way of knowing what the public info is!
		var publicIpAddr net.IP = nil
		var publicGrpcPortSpec *port_spec.PortSpec = nil
		var publicGrpcProxyPortSpec *port_spec.PortSpec = nil

		engineObj := engine.NewEngine(
			engineId,
			engineStatus,
			publicIpAddr,
			publicGrpcPortSpec,
			publicGrpcProxyPortSpec,
		)
		result[engineId] = engineObj
	}
	return result, nil
}

func (backend *KubernetesKurtosisBackend) createEngineRoleBasedResources(ctx context.Context, namespace string, engineAttributesProvider object_attributes_provider.KubernetesEngineObjectAttributesProvider) (resultEngineServiceAccountName string, resultErr error) {

	serviceAccountAttributes, err := engineAttributesProvider.ForEngineServiceAccount()
	if err != nil {
		return "", stacktrace.Propagate(
			err,
			"Expected to be able to get engine attributes for a kubernetes service account, instead got a non-nil error",
		)
	}

	serviceAccountName := serviceAccountAttributes.GetName().GetString()
	serviceAccountLabels := getStringMapFromLabelMap(serviceAccountAttributes.GetLabels())

	if _, err = backend.kubernetesManager.CreateServiceAccount(ctx, serviceAccountName, namespace, serviceAccountLabels); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating service account '%v' with labels '%+v' in namespace '%v'", serviceAccountName, serviceAccountLabels, namespace)
	}

	clusterRolesAttributes, err := engineAttributesProvider.ForEngineClusterRole()
	if err != nil {
		return "", stacktrace.Propagate(
			err,
			"Expected to be able to get engine attributes for a kubernetes cluster role, instead got a non-nil error",
		)
	}

	clusterRoleName := clusterRolesAttributes.GetName().GetString()
	clusterRoleLabels := getStringMapFromLabelMap(clusterRolesAttributes.GetLabels())

	clusterRolePolicyRules := []rbacv1.PolicyRule{
		{
			Verbs:     []string{consts.CreateKubernetesVerb, consts.UpdateKubernetesVerb, consts.PatchKubernetesVerb, consts.DeleteKubernetesVerb, consts.GetKubernetesVerb, consts.ListKubernetesVerb, consts.WatchKubernetesVerb},
			APIGroups: []string{rbacv1.APIGroupAll},
			Resources: []string{consts.NamespacesKubernetesResource, consts.ServiceAccountsKubernetesResource, consts.RolesKubernetesResource, consts.RoleBindingsKubernetesResource, consts.PodsKubernetesResource, consts.ServicesKubernetesResource, consts.PersistentVolumeClaimsKubernetesResource},
		},
	}

	if _, err = backend.kubernetesManager.CreateClusterRoles(ctx, clusterRoleName, clusterRolePolicyRules, clusterRoleLabels); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating cluster role '%v' with policy rules '%+v' and labels '%+v' in namespace '%v'", clusterRoleName, clusterRolePolicyRules, clusterRoleLabels, namespace)
	}

	clusterRoleBindingsAttributes, err := engineAttributesProvider.ForEngineClusterRoleBindings()
	if err != nil {
		return "", stacktrace.Propagate(
			err,
			"Expected to be able to get engine attributes for a kubernetes cluster role bindings, instead got a non-nil error",
		)
	}

	clusterRoleBindingsName := clusterRoleBindingsAttributes.GetName().GetString()
	clusterRoleBindingsLabels := getStringMapFromLabelMap(clusterRoleBindingsAttributes.GetLabels())

	clusterRoleBindingsSubjects := []rbacv1.Subject{
		{
			Kind:      rbacv1.ServiceAccountKind,
			Name:      serviceAccountName,
			Namespace: namespace,
		},
	}

	clusterRoleBindingsRoleRef := rbacv1.RoleRef{
		APIGroup: consts.RbacAuthorizationApiGroup,
		Kind:     consts.ClusterRoleKubernetesResourceType,
		Name:     clusterRoleName,
	}

	if _, err := backend.kubernetesManager.CreateClusterRoleBindings(ctx, clusterRoleBindingsName, clusterRoleBindingsSubjects, clusterRoleBindingsRoleRef, clusterRoleBindingsLabels); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating cluster role bindings '%v' with subjects '%+v' and role ref '%+v' in namespace '%v'", clusterRoleBindingsName, clusterRoleBindingsSubjects, clusterRoleBindingsRoleRef, namespace)
	}

	return serviceAccountName, nil
}

func getEngineContainers(containerImageAndTag string, engineEnvVars map[string]string) (resultContainers []apiv1.Container, resultVolumes []apiv1.Volume) {
	containerName := "kurtosis-engine-container"

	var engineContainerEnvVars []apiv1.EnvVar
	for varName, varValue := range engineEnvVars {
		envVar := apiv1.EnvVar{
			Name:  varName,
			Value: varValue,
		}
		engineContainerEnvVars = append(engineContainerEnvVars, envVar)
	}
	containers := []apiv1.Container{
		{
			Name:  containerName,
			Image: containerImageAndTag,
			Env:   engineContainerEnvVars,
		},
	}

	return containers, nil
}

func getEngineMatchLabels() map[string]string {
	engineMatchLabels := map[string]string{
		label_key_consts.AppIDLabelKey.GetString():        label_value_consts.AppIDLabelValue.GetString(),
		label_key_consts.ResourceTypeLabelKey.GetString(): label_value_consts.EngineResourceTypeLabelValue.GetString(),
	}
	return engineMatchLabels
}
