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
	"strconv"
	"strings"
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

	// Engine container port number string parsing constants
	publicPortNumStrParsingBase = 10
	publicPortNumStrParsingBits = 16

	sentencesSeparator = ", "
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

	//Create engine's service account, cluster roles and cluster role bindings
	engineServiceAccountName, err := backend.createEngineRoleBasedResources(ctx, engineNamespaceName, engineAttributesProvider)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating engine role based resources for engine '%v' in namespace '%v'", engineIdStr, engineNamespaceName)
	}
	shouldRemoveAllEngineRoleBasedResources := true
	defer func() {
		if shouldRemoveAllEngineRoleBasedResources {
			engineIds := map[string]bool{
				engineIdStr: true,
			}
			if _, resultErroredEngineIds := backend.removeEngineRoleBasedResources(ctx, engineNamespaceName, engineIds); len(resultErroredEngineIds) > 0 {
				logrus.Errorf("Creating the engine didn't complete successfully, so we tried to delete the engine kubernetes role based resources that we created for engine '%v' but an error was thrown:\n%v", engineIdStr, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove kubernetes role based resources for engine '%v' in namespace '%v'!!!!!!!", engineIdStr, engineNamespace)
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
	_, err = backend.kubernetesManager.CreatePod(ctx, engineNamespaceName, enginePodName, enginePodLabels, enginePodAnnotations, engineContainers, engineVolumes, engineServiceAccountName)
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

	// Use cluster IP as public IP
	clusterIp := net.ParseIP(service.Spec.ClusterIP)
	if clusterIp == nil {
		return nil, stacktrace.NewError("Expected to be able to parse cluster IP from the kubernetes spec for service '%v', instead nil was parsed.", service.Name)
	}

	publicGrpcPort, publicGrpcProxyPort, err := getEngineGrpcPortSpecsFromServicePorts(service.Spec.Ports)
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

func (backend *KubernetesKurtosisBackend) GetEngines(ctx context.Context, filters *engine.EngineFilters) (map[string]*engine.Engine, error) {
	matchingEnginesByNamespaceAndServiceName, err := backend.getMatchingEngines(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting engines matching the following filters: %+v", filters)
	}

	matchingEnginesByEngineId := map[string]*engine.Engine{}
	for _, engineServices := range matchingEnginesByNamespaceAndServiceName {
		for _, engineObj := range engineServices {
			matchingEnginesByEngineId[engineObj.GetID()] = engineObj
		}
	}

	return matchingEnginesByEngineId, nil
}

func (backend *KubernetesKurtosisBackend) StopEngines(
	ctx context.Context,
	filters *engine.EngineFilters,
) (
	resultSuccessfulEngineIds map[string]bool,
	resultErroredEngineIds map[string]error,
	resultErr error,
) {
	matchingEnginesByNamespaceAndServiceName, err := backend.getMatchingEngines(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting engines matching filters '%+v'", filters)
	}

	successfulEngineIds := map[string]bool{}
	erroredEngineIds := map[string]error{}
	for engineNamespace, engineServices := range matchingEnginesByNamespaceAndServiceName {
		engineServicesToEnginePodsMap := map[string]string{}
		for engineServiceName, engineObj := range engineServices {
			enginePod, err := backend.getEnginePod(ctx, engineObj.GetID(), engineNamespace)
			if err != nil {
				return nil, nil, stacktrace.Propagate(err, "An error occurred getting the engine pod for engine with ID '%v' in namespace '%v'", engineObj.GetID(), engineNamespace)
			}
			if _, found := engineServicesToEnginePodsMap[engineServiceName]; found {
				return nil, nil, stacktrace.NewError("Engine service name '%v' already exist in the engine services to engines pod map '%+v'; it should never happens; this is a bug in Kurtosis", engineServiceName, engineServicesToEnginePodsMap)
			}
			engineServicesToEnginePodsMap[engineServiceName] = enginePod.GetName()
		}
		successfulServiceNames, erroredServiceNames := backend.removeEngineServiceSelectorsAndEnginePods(ctx, engineNamespace, engineServicesToEnginePodsMap)

		removeEngineServiceSelectorsAndEnginePodsSuccessfulEngineIds := map[string]bool{}
		for serviceName := range successfulServiceNames {
			engineObj, found := engineServices[serviceName]
			if !found {
				return nil, nil, stacktrace.NewError("Expected to find service name '%v' in the engine service list '%+v' but it was not found; this is a bug in Kurtosis!", serviceName, engineServices)
			}
			removeEngineServiceSelectorsAndEnginePodsSuccessfulEngineIds[engineObj.GetID()] = true
		}

		for serviceName, err := range erroredServiceNames {
			engineObj, found := engineServices[serviceName]
			if !found {
				return nil, nil, stacktrace.NewError("Expected to find service name '%v' in the engine service list '%+v' but it was not found; this is a bug in Kurtosis!", serviceName, engineServices)
			}
			wrappedErr := stacktrace.Propagate(err, "An error occurred removing engine selectors and pods from kubernetes service for kurtosis engine with ID '%v' and kubernetes service name '%v'", engineObj.GetID(), serviceName)
			erroredEngineIds[engineObj.GetID()] = wrappedErr
		}

		successfulEngineIds = removeEngineServiceSelectorsAndEnginePodsSuccessfulEngineIds
	}

	return successfulEngineIds, erroredEngineIds, nil
}

func (backend *KubernetesKurtosisBackend) DestroyEngines(
	ctx context.Context,
	filters *engine.EngineFilters,
) (
	successfulEngineIds map[string]bool,
	erroredEngineIds map[string]error,
	resultErr error,
) {
	//TODO implement me
	panic("implement me")

	return nil, nil, nil
}

// ====================================================================================================
//                                     Private Helper Methods
// ====================================================================================================
// Gets engines matching the search filters, indexed by their [namespace][service name]
func (backend *KubernetesKurtosisBackend) getMatchingEngines(ctx context.Context, filters *engine.EngineFilters) (map[string]map[string]*engine.Engine, error) {
	matchingEngines := map[string]map[string]*engine.Engine{}

	engineMatchLabels := getEngineMatchLabels()

	engineNamespaces, err := backend.kubernetesManager.GetNamespacesByLabels(ctx, engineMatchLabels)
	if err != nil {
		return nil, stacktrace.NewError("Expected to be able to get engine namespaces from Kubernetes, instead a non-nil error was returned")
	}
	for _, engineNamespace := range engineNamespaces.Items {
		engineNamespaceName := engineNamespace.GetName()

		engineId, isFound := engineNamespace.Labels[label_key_consts.IDLabelKey.GetString()]
		if !isFound {
			return nil, stacktrace.NewError("Expected to find a label with name '%v' in Kubernetes namespace '%v', instead no such label was found", label_key_consts.IDLabelKey.GetString(), engineNamespaceName)
		}
		// If the ID filter is specified, drop engines not matching it
		if filters.IDs != nil && len(filters.IDs) > 0 {
			if _, found := filters.IDs[engineId]; !found {
				continue
			}
		}

		serviceList, err := backend.kubernetesManager.GetServicesByLabels(ctx, engineNamespaceName, engineMatchLabels)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting engine services using labels: %+v in namespace '%v'", engineMatchLabels, engineNamespaceName)
		}
		// Instantiate an empty map for this namespace
		matchingEngines[engineNamespaceName] = map[string]*engine.Engine{}
		for _, service := range serviceList.Items {
			engineObj, err := getEngineObjectFromKubernetesService(service)
			if err != nil {
				return nil, stacktrace.Propagate(err, "Expected to be able to get a kurtosis engine object service from kubernetes service '%v', instead a non-nil error was returned", service.Name)
			}

			// If status filter is specified, drop engines not matching it
			if filters.Statuses != nil && len(filters.Statuses) > 0 {
				if _, found := filters.Statuses[engineObj.GetStatus()]; !found {
					continue
				}
			}

			matchingEngines[engineNamespaceName][service.Name] = engineObj
		}
	}

	return matchingEngines, nil
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
	for engineId, namespace := range namespaces {
		engineResources, found := result[engineId]
		if !found {
			engineResources = &engineKubernetesResources{}
		}
		engineResources.namespace = namespace
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
	for engineId, clusterRole := range clusterRoles {
		engineResources, found := result[engineId]
		if !found {
			engineResources = &engineKubernetesResources{}
		}
		engineResources.clusterRole = clusterRole
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
	for engineId, clusterRoleBinding := range clusterRoleBindings {
		engineResources, found := result[engineId]
		if !found {
			engineResources = &engineKubernetesResources{}
		}
		engineResources.clusterRoleBinding = clusterRoleBinding
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
		if len(serviceAccounts) > 1 {
			return nil, stacktrace.NewError(
				"Expected at most one engine service account in namespace '%v' for engine with ID '%v' " +
					"but found '%v'",
				namespaceName,
				engineId,
				len(serviceAccounts),
			)
		}
		serviceAccount := serviceAccounts[engineId]

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
		if len(services) > 1 {
			return nil, stacktrace.NewError(
				"Expected at most one engine service in namespace '%v' for engine with ID '%v' " +
					"but found '%v'",
				namespaceName,
				engineId,
				len(services),
			)
		}
		service := services[engineId]

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
		if len(pods) > 1 {
			return nil, stacktrace.NewError(
				"Expected at most one engine pod in namespace '%v' for engine with ID '%v' " +
					"but found '%v'",
				namespaceName,
				engineId,
				len(pods),
			)
		}
		pod := pods[engineId]

		engineResources.service = service
		engineResources.pod = pod
		engineResources.serviceAccount = serviceAccount
	}

	return result, nil
}

// TODO parallelize to improve performance
func (backend *KubernetesKurtosisBackend) removeEngineServiceSelectorsAndEnginePods(ctx context.Context, engineNamespace string, serviceNameToPodNameMap map[string]string) (map[string]bool, map[string]error) {
	successfulServices := map[string]bool{}
	failedServices := map[string]error{}
	for serviceName, podName := range serviceNameToPodNameMap {
		if err := backend.kubernetesManager.RemoveSelectorsFromService(ctx, engineNamespace, serviceName); err != nil {
			failedServices[serviceName] = err
		} else {
			if err := backend.kubernetesManager.RemovePod(ctx, engineNamespace, podName); err != nil {
				failedServices[serviceName] = stacktrace.Propagate(err, "Tried to remove pod '%v' associated with service '%v' in namespace '%v', instead a non-nil err was returned", podName, serviceName, engineNamespace)
			}
			successfulServices[serviceName] = true
		}
	}

	return successfulServices, failedServices
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

// TODO parallelize to improve performance
func (backend *KubernetesKurtosisBackend) removeEngineRoleBasedResources(ctx context.Context, namespace string, engineIds map[string]bool) (resultSuccessfulEngineIds map[string]bool, resultErroredEngineIds map[string]error) {

	successfulEngineIds := map[string]bool{}
	erroredEngineIds := map[string]error{}
	for engineIdStr := range engineIds {

		//First remove cluster role bindings
		clusterRoleBindings, err := backend.getAllEngineClusterRoleBindings(ctx, engineIdStr)
		if err != nil {
			wrapErr := stacktrace.Propagate(err, "An error occurred getting all engine cluster role bindings for engine '%v'", engineIdStr)
			erroredEngineIds[engineIdStr] = wrapErr
			continue
		}

		errClusterRoleBindingNames := []string{}
		errClusterRoleBindingErrMsgs := []string{}
		for _, clusterRoleBinding := range clusterRoleBindings {
			clusterRoleBindingName := clusterRoleBinding.GetName()
			if err := backend.kubernetesManager.RemoveClusterRoleBindings(ctx, clusterRoleBindingName); err != nil {
				wrapErr := stacktrace.Propagate(err, "An error occurred removing engine cluster role binding '%v'", clusterRoleBindingName)
				errClusterRoleBindingNames = append(errClusterRoleBindingNames, clusterRoleBindingName)
				errClusterRoleBindingErrMsgs = append(errClusterRoleBindingErrMsgs, wrapErr.Error())
			}
		}

		if len(errClusterRoleBindingNames) > 0 {
			errClusterRoleBindingNamesStr := strings.Join(errClusterRoleBindingNames, sentencesSeparator)
			errClusterRoleBindingErrMsgsStr := strings.Join(errClusterRoleBindingErrMsgs, sentencesSeparator)
			wrapErr := stacktrace.NewError("An error occurred removing cluster role bindings '%v' error: %v ", errClusterRoleBindingNames, errClusterRoleBindingErrMsgsStr)
			erroredEngineIds[engineIdStr] = wrapErr
			logrus.Errorf("Removing the engine role-based resources didn't complete successfully, so we tried to delete kubernetes cluster role bindings '%v' that we created but an error was thrown:\n%v", errClusterRoleBindingNamesStr, errClusterRoleBindingErrMsgsStr)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually remove kubernetes cluster role bindings with name '%v'!!!!!!!", errClusterRoleBindingNamesStr)
			continue
		}

		//Then cluster roles
		clusterRoles, err := backend.getAllEngineClusterRoles(ctx, engineIdStr)
		if err != nil {
			wrapErr := stacktrace.Propagate(err, "An error occurred getting all engine cluster roles for engine '%v'", engineIdStr)
			erroredEngineIds[engineIdStr] = wrapErr
			continue
		}

		errClusterRoleNames := []string{}
		errClusterRoleErrMsgs := []string{}
		for _, clusterRole := range clusterRoles {
			clusterRoleName := clusterRole.GetName()
			if err := backend.kubernetesManager.RemoveClusterRole(ctx, clusterRoleName); err != nil {
				wrapErr := stacktrace.Propagate(err, "An error occurred removing engine cluster role '%v'", clusterRoleName)
				errClusterRoleNames = append(errClusterRoleNames, clusterRoleName)
				errClusterRoleErrMsgs = append(errClusterRoleErrMsgs, wrapErr.Error())
			}
		}

		if len(errClusterRoleNames) > 0 {
			errClusterRoleNamesStr := strings.Join(errClusterRoleNames, sentencesSeparator)
			errClusterRoleErrMsgsStr := strings.Join(errClusterRoleErrMsgs, sentencesSeparator)
			wrapErr := stacktrace.NewError("An error occurred removing cluster roles '%v' error: %v ", errClusterRoleNamesStr, errClusterRoleErrMsgsStr)
			erroredEngineIds[engineIdStr] = wrapErr
			logrus.Errorf("Removing the engine role-based resources didn't complete successfully, so we tried to delete kubernetes cluster roles '%v' that we created but an error was thrown:\n%v", errClusterRoleNamesStr, errClusterRoleErrMsgsStr)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually remove kubernetes cluster roles with name '%v'!!!!!!!", errClusterRoleNamesStr)
			continue
		}

		//And finally, the service accounts
		serviceAccounts, err := backend.getAllEngineServiceAccounts(ctx, namespace, engineIdStr)
		if err != nil {
			wrapErr := stacktrace.Propagate(err, "An error occurred getting all engine service accounts for engine '%v' in namespace '%v'", engineIdStr, namespace)
			erroredEngineIds[engineIdStr] = wrapErr
			continue
		}

		errServiceAccountNames := []string{}
		errServiceAccountErrMsgs := []string{}
		for _, serviceAccount := range serviceAccounts {
			serviceAccountName := serviceAccount.GetName()
			if err := backend.kubernetesManager.RemoveServiceAccount(ctx, serviceAccountName, namespace); err != nil {
				wrapErr := stacktrace.Propagate(err, "An error occurred removing engine service account '%v' in namespace '%v'", serviceAccountName, namespace)
				errServiceAccountNames = append(errServiceAccountNames, serviceAccountName)
				errServiceAccountErrMsgs = append(errServiceAccountErrMsgs, wrapErr.Error())
			}
		}

		if len(errServiceAccountNames) > 0 {
			errServiceAccountNamesStr := strings.Join(errServiceAccountNames, sentencesSeparator)
			errServiceAccountErrMsgsStr := strings.Join(errServiceAccountErrMsgs, sentencesSeparator)
			wrapErr := stacktrace.NewError("An error occurred removing service accounts '%v' error: %v ", errServiceAccountNamesStr, errServiceAccountErrMsgsStr)
			erroredEngineIds[engineIdStr] = wrapErr
			logrus.Errorf("Removing the engine role-based resources didn't complete successfully, so we tried to delete kubernetes service accounts '%v' that we created but an error was thrown:\n%v", errServiceAccountNamesStr, errServiceAccountErrMsgsStr)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually remove kubernetes service accounts with name '%v' in namespace '%v'!!!!!!!", errServiceAccountNamesStr, namespace)
			continue
		}

		successfulEngineIds[engineIdStr] = true
	}

	return successfulEngineIds, erroredEngineIds
}

func (backend *KubernetesKurtosisBackend) getEngineNamespace(ctx context.Context, engineId string) (*apiv1.Namespace, error) {

	engineMatchLabels := getEngineMatchLabels()

	namespaces, err := backend.kubernetesManager.GetNamespacesByLabels(ctx, engineMatchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the engine namespace using labels '%+v'", engineMatchLabels)
	}

	filteredNamespaces := []*apiv1.Namespace{}

	for _, foundNamespace := range namespaces.Items {
		foundNamespaceLabels := foundNamespace.GetLabels()

		foundEngineId, found := foundNamespaceLabels[label_key_consts.IDLabelKey.GetString()]
		if !found {
			return nil, stacktrace.NewError("Expected to find ID label key '%v' but none was found", label_key_consts.IDLabelKey.GetString())
		}

		if engineId == foundEngineId {
			filteredNamespaces = append(filteredNamespaces, &foundNamespace)
		}
	}
	numOfNamespaces := len(filteredNamespaces)
	if numOfNamespaces == 0 {
		return nil, stacktrace.NewError("No namespace matching labels '%+v' was found", engineMatchLabels)
	}
	if numOfNamespaces > 1 {
		return nil, stacktrace.NewError("Expected to find only one engine namespace for engine ID '%v', but '%v' was found; this is a bug in Kurtosis", engineId, numOfNamespaces)
	}

	resultNamespace := filteredNamespaces[0]

	return resultNamespace, nil
}

// The current Kurtosis Kubernetes architecture defines only one pod for Engine
// This method should be refactored if the architecture changes, and we decide to use replicas for the Engine
func (backend *KubernetesKurtosisBackend) getEnginePod(ctx context.Context, engineId string, namespace string) (*apiv1.Pod, error) {

	engineMatchLabels := getEngineMatchLabels()

	pods, err := backend.kubernetesManager.GetPodsByLabels(ctx, namespace, engineMatchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the engine pod in namespace '%v' using labels '%+v' ", namespace, engineMatchLabels)
	}

	filteredPods := []*apiv1.Pod{}

	for _, foundPod := range pods.Items {
		foundPodLabels := foundPod.GetLabels()

		foundEngineId, found := foundPodLabels[label_key_consts.IDLabelKey.GetString()]
		if !found {
			return nil, stacktrace.NewError("Expected to find ID label key '%v' but none was found", label_key_consts.IDLabelKey.GetString())
		}

		if engineId == foundEngineId {
			filteredPods = append(filteredPods, &foundPod)
		}
	}
	numOfPods := len(filteredPods)
	if numOfPods == 0 {
		return nil, stacktrace.NewError("No pods matching labels '%+v' was found", engineMatchLabels)
	}
	//We are not using replicas for Kurtosis engines
	if numOfPods > 1 {
		return nil, stacktrace.NewError("Expected to find only one engine pod for engine ID '%v', but '%v' was found; this is a bug in Kurtosis", engineId, numOfPods)
	}

	resultPod := filteredPods[0]

	return resultPod, nil
}

func (backend *KubernetesKurtosisBackend) getAllEngineServiceAccounts(ctx context.Context, engineNamespace string, engineId string) ([]*apiv1.ServiceAccount, error) {
	engineMatchLabels := getEngineMatchLabels()

	serviceAccounts, err := backend.kubernetesManager.GetServiceAccountsByLabels(ctx, engineNamespace, engineMatchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting service accounts using labels '%+v' ", engineMatchLabels)
	}

	filteredServiceAccounts := []*apiv1.ServiceAccount{}

	for _, foundServiceAccount := range serviceAccounts.Items {
		foundServiceAccountLabels := foundServiceAccount.GetLabels()

		foundEngineId, found := foundServiceAccountLabels[label_key_consts.IDLabelKey.GetString()]
		if !found {
			return nil, stacktrace.NewError("Expected to find ID label key '%v' but none was found", label_key_consts.IDLabelKey.GetString())
		}

		if engineId == foundEngineId {
			filteredServiceAccounts = append(filteredServiceAccounts, &foundServiceAccount)
		}
	}

	return filteredServiceAccounts, nil
}

func (backend *KubernetesKurtosisBackend) getAllEngineClusterRoles(ctx context.Context, engineId string) ([]*rbacv1.ClusterRole, error) {
	engineMatchLabels := getEngineMatchLabels()

	clusterRoles, err := backend.kubernetesManager.GetClusterRolesByLabels(ctx, engineMatchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting cluster roles using labels '%+v' ", engineMatchLabels)
	}

	filteredClusterRoles := []*rbacv1.ClusterRole{}

	for _, foundClusterRole := range clusterRoles.Items {
		foundClusterRoleLabels := foundClusterRole.GetLabels()

		foundEngineId, found := foundClusterRoleLabels[label_key_consts.IDLabelKey.GetString()]
		if !found {
			return nil, stacktrace.NewError("Expected to find ID label key '%v' but none was found", label_key_consts.IDLabelKey.GetString())
		}

		if engineId == foundEngineId {
			filteredClusterRoles = append(filteredClusterRoles, &foundClusterRole)
		}
	}

	return filteredClusterRoles, nil
}

func (backend *KubernetesKurtosisBackend) getAllEngineClusterRoleBindings(ctx context.Context, engineId string) ([]*rbacv1.ClusterRoleBinding, error) {
	engineMatchLabels := getEngineMatchLabels()

	clusterRoleBindings, err := backend.kubernetesManager.GetClusterRoleBindingsByLabels(ctx, engineMatchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting engine cluster role bindings using labels '%+v' ", engineMatchLabels)
	}

	filteredClusterRoleBindings := []*rbacv1.ClusterRoleBinding{}

	for _, foundClusterRoleBinding := range clusterRoleBindings.Items {
		foundClusterRoleBindingLabels := foundClusterRoleBinding.GetLabels()

		foundEngineId, found := foundClusterRoleBindingLabels[label_key_consts.IDLabelKey.GetString()]
		if !found {
			return nil, stacktrace.NewError("Expected to find ID label key '%v' but none was found", label_key_consts.IDLabelKey.GetString())
		}

		if engineId == foundEngineId {
			filteredClusterRoleBindings = append(filteredClusterRoleBindings, &foundClusterRoleBinding)
		}
	}

	return filteredClusterRoleBindings, nil
}

/*
func (backend *KubernetesKurtosisBackend) destroyEngineResources(ctx context.Context, engineId string) {
	engineObjAttrsProvider, err := backend.objAttrsProvider.ForEngine(engineId)
	engineVolumeAttributes, err := engineObjAttrsProvider.ForEngineVolume()
	enginePodAttributes, err := engineObjAttrsProvider.ForEnginePod()

	// Remove Deployment
	if err := backend.kubernetesManager.RemoveDeployment(ctx, kurtosisEngineNamespace, enginePodAttributes.GetName().GetString()); err != nil {

	}
	// Destroy Service ?

	// Destroy Persistent Volume Claim
	backend.kubernetesManager.RemovePersistentVolumeClaim(ctx, kurtosisEngineNamespace, engineVolumeAttributes.GetName().GetString())

	// Destroy Volume (maybe
}
*/

func getEngineObjectFromKubernetesService(service apiv1.Service) (*engine.Engine, error) {
	engineId, isFound := service.Labels[label_key_consts.IDLabelKey.GetString()]
	if isFound == false {
		return nil, stacktrace.NewError("Expected to be able to find label describing the engine id on service '%v' with label key '%v', but was unable to", service.Name, label_key_consts.IDLabelKey.GetString())
	}
	// the ContainerStatus naming is confusing
	engineStatus := getKurtosisStatusFromKubernetesService(service)
	var publicIpAddr net.IP
	var publicGrpcPortSpec *port_spec.PortSpec
	var publicGrpcProxyPortSpec *port_spec.PortSpec
	if engineStatus == container_status.ContainerStatus_Running {
		publicIpAddr = net.ParseIP(service.Spec.ClusterIP)
		if publicIpAddr == nil {
			return nil, stacktrace.NewError("Expected to be able to get the cluster ip of the engine service, instead parsing the cluster ip of service '%v' returned nil", service.Name)
		}
		var portSpecError error
		publicGrpcPortSpec, publicGrpcProxyPortSpec, portSpecError = getEngineGrpcPortSpecsFromServicePorts(service.Spec.Ports)
		if portSpecError != nil {
			return nil, stacktrace.Propagate(portSpecError, "Expected to be able to determine engine grpc port specs from kubernetes service ports for engine '%v', instead a non-nil error was returned", engineId)
		}
	}

	return engine.NewEngine(engineId, engineStatus, publicIpAddr, publicGrpcPortSpec, publicGrpcProxyPortSpec), nil

}

func getKurtosisStatusFromKubernetesService(service apiv1.Service) container_status.ContainerStatus {
	// If a Kubernetes Service has selectors, then we assume the engine is reachable, and thus not stopped
	// see stopEngineService for how we stop the engine
	// label keys and values used to determine pods this service routes traffic too
	// TODO Better determination of if the engine is reachable? Check that there are two ports with names we expect them to have?
	serviceSelectors := service.Spec.Selector
	if len(serviceSelectors) == 0 {
		return container_status.ContainerStatus_Stopped
	}
	return container_status.ContainerStatus_Running
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

func getEngineGrpcPortSpecsFromServicePorts(servicePorts []apiv1.ServicePort) (resultGrpcPortSpec *port_spec.PortSpec, resultGrpcProxyPortSpec *port_spec.PortSpec, resultErr error) {
	var publicGrpcPort *port_spec.PortSpec
	var publicGrpcProxyPort *port_spec.PortSpec
	grpcPortName := object_name_constants.KurtosisInternalContainerGrpcPortName.GetString()
	grpcProxyPortName := object_name_constants.KurtosisInternalContainerGrpcProxyPortName.GetString()

	for _, servicePort := range servicePorts {
		servicePortName := servicePort.Name
		switch servicePortName {
		case grpcPortName:
			{
				publicGrpcPortSpec, err := getPublicPortSpecFromServicePort(servicePort, enginePortProtocol)
				if err != nil {
					return nil, nil, stacktrace.Propagate(err, "Expected to be able to create a port spec describing an engine's public grpc port from kubernetes service port '%v', instead a non nil error was returned", servicePortName)
				}
				publicGrpcPort = publicGrpcPortSpec
			}
		case grpcProxyPortName:
			{
				publicGrpcProxyPortSpec, err := getPublicPortSpecFromServicePort(servicePort, enginePortProtocol)
				if err != nil {
					return nil, nil, stacktrace.Propagate(err, "Expected to be able to create a port spec describing an engine's public grpc proxy port from kubernetes service port '%v', instead a non nil error was returned", servicePortName)
				}
				publicGrpcProxyPort = publicGrpcProxyPortSpec
			}
		}
	}

	if publicGrpcPort == nil || publicGrpcProxyPort == nil {
		return nil, nil, stacktrace.NewError("Expected to get public port specs from kubernetes service ports, instead got a nil pointer")
	}

	return publicGrpcPort, publicGrpcProxyPort, nil

}

// getPublicPortSpecFromServicePort returns a port_spec representing a kurtosis port spec for a service port in kubernetes
func getPublicPortSpecFromServicePort(servicePort apiv1.ServicePort, portProtocol port_spec.PortProtocol) (*port_spec.PortSpec, error) {
	publicPortNumStr := strconv.FormatInt(int64(servicePort.Port), 10)
	publicPortNumUint64, err := strconv.ParseUint(publicPortNumStr, publicPortNumStrParsingBase, publicPortNumStrParsingBits)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred parsing engine server public port string '%v' using base '%v' and uint bits '%v'",
			publicPortNumStr,
			publicPortNumStrParsingBase,
			publicPortNumStrParsingBits,
		)
	}
	publicPortNum := uint16(publicPortNumUint64) // Safe to do because we pass the requisite number of bits into the parse command
	publicGrpcPort, err := port_spec.NewPortSpec(publicPortNum, portProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to create a port spec describing a public port on a kubernetes node using number '%v' and protocol '%v', instead a non nil error was returned", publicPortNum, portProtocol)
	}

	return publicGrpcPort, nil
}

func getEngineMatchLabels() map[string]string {
	engineMatchLabels := map[string]string{
		label_key_consts.AppIDLabelKey.GetString():        label_value_consts.AppIDLabelValue.GetString(),
		label_key_consts.ResourceTypeLabelKey.GetString(): label_value_consts.EngineResourceTypeLabelValue.GetString(),
	}
	return engineMatchLabels
}
