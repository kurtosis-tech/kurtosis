package kubernetes

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager/consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_resource_collectors"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_value"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	applyconfigurationsv1 "k8s.io/client-go/applyconfigurations/core/v1"
	"net"
	"time"
)

const (
	kurtosisEngineContainerName = "kurtosis-engine-container"

	maxWaitForEngineContainerAvailabilityRetries         = 30
	timeBetweenWaitForEngineContainerAvailabilityRetries = 1 * time.Second
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
	engineGuidStr, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred generating a UUID string for the engine")
	}
	engineGuid := engine.EngineGUID(engineGuidStr)

	privateGrpcPortSpec, err := port_spec.NewPortSpec(grpcPortNum, kurtosisServersPortProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the engine's private grpc port spec object using number '%v' and protocol '%v'",
			grpcPortNum,
			kurtosisServersPortProtocol.String(),
		)
	}
	privateGrpcProxyPortSpec, err := port_spec.NewPortSpec(grpcProxyPortNum, kurtosisServersPortProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the engine's private grpc proxy port spec object using number '%v' and protocol '%v'",
			grpcProxyPortNum,
			kurtosisServersPortProtocol.String(),
		)
	}
	privatePortSpecs := map[string]*port_spec.PortSpec{
		kurtosisInternalContainerGrpcPortSpecId: privateGrpcPortSpec,
		kurtosisInternalContainerGrpcProxyPortSpecId: privateGrpcProxyPortSpec,
	}

	engineAttributesProvider := backend.objAttrsProvider.ForEngine(engineGuid)

	namespace, err := backend.createEngineNamespace(ctx, engineAttributesProvider)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the engine namespace")
	}
	shouldRemoveNamespace := true
	defer func() {
		if shouldRemoveNamespace {
			if err := backend.kubernetesManager.RemoveNamespace(ctx, namespace); err != nil {
				logrus.Errorf("Creating the engine didn't complete successfully, so we tried to delete Kubernetes namespace '%v' that we created but an error was thrown:\n%v", namespace.Name, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove Kubernetes namespace with name '%v'!!!!!!!", namespace.Name)
			}
		}
	}()
	namespaceName := namespace.Name

	serviceAccount, err := backend.createEngineServiceAccount(ctx, namespaceName, engineAttributesProvider)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the engine service account")
	}
	shouldRemoveServiceAccount := true
	defer func() {
		if shouldRemoveServiceAccount {
			if err := backend.kubernetesManager.RemoveServiceAccount(ctx, serviceAccount); err != nil {
				logrus.Errorf("Creating the engine didn't complete successfully, so we tried to delete service account '%v' in namespace '%v' that we created but an error was thrown:\n%v", serviceAccount.Name, namespaceName, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove service account with name '%v'!!!!!!!", serviceAccount.Name)
			}
		}
	}()

	clusterRole, err := backend.createEngineClusterRole(ctx, engineAttributesProvider)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the engine cluster role")
	}
	shouldRemoveClusterRole := true
	defer func() {
		if shouldRemoveClusterRole {
			if err := backend.kubernetesManager.RemoveClusterRole(ctx, clusterRole); err != nil {
				logrus.Errorf("Creating the engine didn't complete successfully, so we tried to delete cluster role '%v' that we created but an error was thrown:\n%v", clusterRole.Name, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove cluster role with name '%v'!!!!!!!", clusterRole.Name)
			}
		}
	}()

	clusterRoleBindings, err := backend.createEngineClusterRoleBindings(ctx, engineAttributesProvider, clusterRole.Name, namespaceName, serviceAccount.Name)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the engine cluster role bindings")
	}
	shouldRemoveClusterRoleBinding := true
	defer func() {
		if shouldRemoveClusterRoleBinding {
			if err := backend.kubernetesManager.RemoveClusterRoleBindings(ctx, clusterRoleBindings); err != nil {
				logrus.Errorf("Creating the engine didn't complete successfully, so we tried to delete cluster role bindings '%v' in namespace '%v' that we created but an error was thrown:\n%v", clusterRoleBindings.Name, namespaceName, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove cluster role bindings with name '%v'!!!!!!!", clusterRoleBindings.Name)
			}
		}
	}()

	enginePod, enginePodLabels, err := backend.createEnginePod(ctx, namespaceName, engineAttributesProvider, imageOrgAndRepo, imageVersionTag, envVars, privatePortSpecs, serviceAccount.Name)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the engine pod")
	}
	var shouldRemovePod = true
	defer func() {
		if shouldRemovePod {
			if err := backend.kubernetesManager.RemovePod(ctx, enginePod); err != nil {
				logrus.Errorf("Creating the engine didn't complete successfully, so we tried to delete Kubernetes pod '%v' that we created but an error was thrown:\n%v", enginePod.Name, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove Kubernetes pod with name '%v'!!!!!!!", enginePod.Name)
			}
		}
	}()

	engineService, err := backend.createEngineService(
		ctx,
		namespaceName,
		engineAttributesProvider,
		privateGrpcPortSpec,
		privateGrpcProxyPortSpec,
		enginePodLabels,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the engine service")
	}
	var shouldRemoveService = true
	defer func() {
		if shouldRemoveService {
			if err := backend.kubernetesManager.RemoveService(ctx, engineService); err != nil {
				logrus.Errorf("Creating the engine didn't complete successfully, so we tried to delete Kubernetes service '%v' that we created but an error was thrown:\n%v", engineService.Name, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove Kubernetes service with name '%v'!!!!!!!", engineService.Name)
			}
		}
	}()

	engineResources := &engineKubernetesResources{
		clusterRole:        clusterRole,
		clusterRoleBinding: clusterRoleBindings,
		namespace:          namespace,
		serviceAccount:     serviceAccount,
		service:            engineService,
		pod:                enginePod,
	}
	engineObjsById, err := getEngineObjectsFromKubernetesResources(map[engine.EngineGUID]*engineKubernetesResources{
		engineGuid: engineResources,
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred converting the new engine's Kubernetes resources to engine objects")
	}
	resultEngine, found := engineObjsById[engineGuid]
	if !found {
		return nil, stacktrace.NewError("Successfully converted the new engine's Kubernetes resources to an engine object, but the resulting map didn't have an entry for engine GUID '%v'", engineGuid)
	}

	if err := waitForPortAvailabilityUsingNetstat(
		backend.kubernetesManager,
		namespaceName,
		enginePod.Name,
		kurtosisEngineContainerName,
		privateGrpcPortSpec,
		maxWaitForEngineContainerAvailabilityRetries,
		timeBetweenWaitForEngineContainerAvailabilityRetries,
	); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the engine grpc port '%v/%v' to become available", privateGrpcPortSpec.GetProtocol(), privateGrpcPortSpec.GetNumber())
	}

	// TODO UNCOMMENT THIS ONCE WE HAVE GRPC-PROXY WIRED UP!!
	/*
	if err := waitForPortAvailabilityUsingNetstat(
		backend.kubernetesManager,
		namespaceName,
		enginePod.Name,
		kurtosisEngineContainerName,
		privateGrpcProxyPortSpec,
		maxWaitForEngineContainerAvailabilityRetries,
		timeBetweenWaitForEngineContainerAvailabilityRetries,
	); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the engine grpc proxy port '%v/%v' to become available", privateGrpcProxyPortSpec.GetProtocol(), privateGrpcProxyPortSpec.GetNumber())
	}*/

	shouldRemoveNamespace = false
	shouldRemoveServiceAccount = false
	shouldRemoveClusterRole = false
	shouldRemoveClusterRoleBinding = false
	shouldRemovePod = false
	shouldRemoveService = false
	return resultEngine, nil
}

func (backend *KubernetesKurtosisBackend) GetEngines(ctx context.Context, filters *engine.EngineFilters) (map[engine.EngineGUID]*engine.Engine, error) {
	matchingEngines, _, err := backend.getMatchingEngineObjectsAndKubernetesResources(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting engines matching the following filters: %+v", filters)
	}
	return matchingEngines, nil
}

func (backend *KubernetesKurtosisBackend) StopEngines(
	ctx context.Context,
	filters *engine.EngineFilters,
) (
	resultSuccessfulEngineGuids map[engine.EngineGUID]bool,
	resultErroredEngineGuids map[engine.EngineGUID]error,
	resultErr error,
) {
	_, matchingKubernetesResources, err := backend.getMatchingEngineObjectsAndKubernetesResources(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting engines and Kubernetes resources matching filters '%+v'", filters)
	}

	successfulEngineGuids := map[engine.EngineGUID]bool{}
	erroredEngineGuids := map[engine.EngineGUID]error{}
	for engineGuid, resources := range matchingKubernetesResources {
		if resources.namespace == nil {
			// No namespace means nothing needs stopping
			successfulEngineGuids[engineGuid] = true
			continue
		}
		namespaceName := resources.namespace.Name

		if resources.pod != nil {
			podName := resources.pod.Name
			if err := backend.kubernetesManager.RemovePod(ctx, resources.pod); err != nil {
				erroredEngineGuids[engineGuid] = stacktrace.Propagate(
					err,
					"An error occurred removing pod '%v' in namespace '%v' for engine '%v'",
					podName,
					namespaceName,
					engineGuid,
				)
				continue
			}
		}

		kubernetesService := resources.service
		if kubernetesService != nil {
			serviceName := kubernetesService.Name
			updateConfigurator := func(updatesToApply *applyconfigurationsv1.ServiceApplyConfiguration) {
				specUpdates := applyconfigurationsv1.ServiceSpec().WithSelector(nil)
				updatesToApply.WithSpec(specUpdates)
			}
			if _, err := backend.kubernetesManager.UpdateService(ctx, namespaceName, serviceName, updateConfigurator); err != nil {
				erroredEngineGuids[engineGuid] = stacktrace.Propagate(
					err,
					"An error occurred removing selectors from service '%v' in namespace '%v' for engine '%v'",
					kubernetesService.Name,
					namespaceName,
					engineGuid,
				)
				continue
			}
		}

		successfulEngineGuids[engineGuid] = true
	}

	return successfulEngineGuids, erroredEngineGuids, nil
}

func (backend *KubernetesKurtosisBackend) DestroyEngines(
	ctx context.Context,
	filters *engine.EngineFilters,
) (
	resultSuccessfulEngineGuids map[engine.EngineGUID]bool,
	resultErroredEngineGuids map[engine.EngineGUID]error,
	resultErr error,
) {
	_, matchingResources, err := backend.getMatchingEngineObjectsAndKubernetesResources(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting engine Kubernetes resources matching filters: %+v", filters)
	}

	successfulEngineGuids := map[engine.EngineGUID]bool{}
	erroredEngineGuids := map[engine.EngineGUID]error{}
	for engineGuid, resources := range matchingResources {
		// Remove ClusterRoleBinding
		if resources.clusterRoleBinding != nil {
			roleBindingName := resources.clusterRoleBinding.Name
			if err := backend.kubernetesManager.RemoveClusterRoleBindings(ctx, resources.clusterRoleBinding); err != nil {
				erroredEngineGuids[engineGuid] = stacktrace.Propagate(
					err,
					"An error occurred removing cluster role binding '%v' for engine '%v'",
					roleBindingName,
					engineGuid,
				)
				continue
			}
		}

		// Remove ClusterRole
		if resources.clusterRole != nil {
			roleName := resources.clusterRole.Name
			if err := backend.kubernetesManager.RemoveClusterRole(ctx, resources.clusterRole); err != nil {
				erroredEngineGuids[engineGuid] = stacktrace.Propagate(
					err,
					"An error occurred removing cluster role '%v' for engine '%v'",
					roleName,
					engineGuid,
				)
				continue
			}
		}

		// Remove the namespace
		if resources.namespace != nil {
			namespaceName := resources.namespace.Name
			if err := backend.kubernetesManager.RemoveNamespace(ctx, resources.namespace); err != nil {
				erroredEngineGuids[engineGuid] = stacktrace.Propagate(
					err,
					"An error occurred removing namespace '%v' for engine '%v'",
					namespaceName,
					engineGuid,
				)
				continue
			}
		}

		successfulEngineGuids[engineGuid] = true
	}
	return successfulEngineGuids, erroredEngineGuids, nil
}

// ====================================================================================================
//                                     Private Helper Methods
// ====================================================================================================
func (backend *KubernetesKurtosisBackend) getMatchingEngineObjectsAndKubernetesResources(
	ctx context.Context,
	filters *engine.EngineFilters,
) (
	map[engine.EngineGUID]*engine.Engine,
	map[engine.EngineGUID]*engineKubernetesResources,
	error,
) {
	matchingResources, err := backend.getMatchingEngineKubernetesResources(ctx, filters.GUIDs)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting engine Kubernetes resources matching GUIDs: %+v", filters.GUIDs)
	}

	engineObjects, err := getEngineObjectsFromKubernetesResources(matchingResources)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting engine objects from Kubernetes resources")
	}

	// Finally, apply the filters
	resultEngineObjs := map[engine.EngineGUID]*engine.Engine{}
	resultKubernetesResources := map[engine.EngineGUID]*engineKubernetesResources{}
	for engineGuid, engineObj := range engineObjects {
		if filters.GUIDs != nil && len(filters.GUIDs) > 0 {
			if _, found := filters.GUIDs[engineObj.GetGUID()]; !found {
				continue
			}
		}

		if filters.Statuses != nil && len(filters.Statuses) > 0 {
			if _, found := filters.Statuses[engineObj.GetStatus()]; !found {
				continue
			}
		}

		resultEngineObjs[engineGuid] = engineObj
		// Okay to do because we're guaranteed a 1:1 mapping between engine_obj:engine_resources
		resultKubernetesResources[engineGuid] = matchingResources[engineGuid]
	}

	return resultEngineObjs, resultKubernetesResources, nil
}

// Get back any and all engine's Kubernetes resources matching the given GUIDs, where a nil or empty map == "match all GUIDs"
func (backend *KubernetesKurtosisBackend) getMatchingEngineKubernetesResources(
	ctx context.Context,
	engineGuids map[engine.EngineGUID]bool,
) (
	map[engine.EngineGUID]*engineKubernetesResources,
	error,
) {
	engineMatchLabels := getEngineMatchLabels()

	result := map[engine.EngineGUID]*engineKubernetesResources{}

	engineGuidStrs := map[string]bool{}
	for engineGuid := range engineGuids {
		engineGuidStrs[string(engineGuid)] = true
	}

	// Namespaces
	namespaces, err := kubernetes_resource_collectors.CollectMatchingNamespaces(
		ctx,
		backend.kubernetesManager,
		engineMatchLabels,
		label_key_consts.IDKubernetesLabelKey.GetString(),
		engineGuidStrs,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting engine namespaces matching GUIDs '%+v'", engineGuids)
	}
	for engineGuidStr, namespacesForId := range namespaces {
		engineGuid := engine.EngineGUID(engineGuidStr)
		if len(namespacesForId) > 1 {
			return nil, stacktrace.NewError(
				"Expected at most one namespace to match engine GUID '%v', but got '%v'",
				len(namespacesForId),
				engineGuidStr,
			)
		}
		engineResources, found := result[engineGuid]
		if !found {
			engineResources = &engineKubernetesResources{}
		}
		engineResources.namespace = namespacesForId[0]
		result[engineGuid] = engineResources
	}

	// Cluster roles
	clusterRoles, err := kubernetes_resource_collectors.CollectMatchingClusterRoles(
		ctx,
		backend.kubernetesManager,
		engineMatchLabels,
		label_key_consts.IDKubernetesLabelKey.GetString(),
		engineGuidStrs,
	)
	for engineGuidStr, clusterRolesForId := range clusterRoles {
		engineGuid := engine.EngineGUID(engineGuidStr)
		if len(clusterRolesForId) > 1 {
			return nil, stacktrace.NewError(
				"Expected at most one cluster role to match engine GUID '%v', but got '%v'",
				len(clusterRolesForId),
				engineGuidStr,
			)
		}
		engineResources, found := result[engineGuid]
		if !found {
			engineResources = &engineKubernetesResources{}
		}
		engineResources.clusterRole = clusterRolesForId[0]
		result[engineGuid] = engineResources
	}

	// Cluster role bindings
	clusterRoleBindings, err := kubernetes_resource_collectors.CollectMatchingClusterRoleBindings(
		ctx,
		backend.kubernetesManager,
		engineMatchLabels,
		label_key_consts.IDKubernetesLabelKey.GetString(),
		engineGuidStrs,
	)
	for engineGuidStr, clusterRoleBindingsForId := range clusterRoleBindings {
		engineGuid := engine.EngineGUID(engineGuidStr)
		if len(clusterRoleBindingsForId) > 1 {
			return nil, stacktrace.NewError(
				"Expected at most one cluster role binding to match engine GUID '%v', but got '%v'",
				len(clusterRoleBindingsForId),
				engineGuidStr,
			)
		}
		engineResources, found := result[engineGuid]
		if !found {
			engineResources = &engineKubernetesResources{}
		}
		engineResources.clusterRoleBinding = clusterRoleBindingsForId[0]
		result[engineGuid] = engineResources
	}

	// Per-namespace objects
	for engineGuid, engineResources := range result {
		if engineResources.namespace == nil {
			continue
		}
		namespaceName := engineResources.namespace.Name

		engineGuidStr := string(engineGuid)

		// Service accounts
		serviceAccounts, err := kubernetes_resource_collectors.CollectMatchingServiceAccounts(
			ctx,
			backend.kubernetesManager,
			namespaceName,
			engineMatchLabels,
			label_key_consts.IDKubernetesLabelKey.GetString(),
			map[string]bool{
				engineGuidStr: true,
			},
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting service accounts matching engine GUID '%v' in namespace '%v'", engineGuid, namespaceName)
		}
		var serviceAccount *apiv1.ServiceAccount
		if serviceAccountsForId, found := serviceAccounts[engineGuidStr]; found {
			if len(serviceAccountsForId) > 1 {
				return nil, stacktrace.NewError(
					"Expected at most one engine service account in namespace '%v' for engine with GUID '%v' " +
						"but found '%v'",
					namespaceName,
					engineGuid,
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
			label_key_consts.IDKubernetesLabelKey.GetString(),
			map[string]bool{
				engineGuidStr: true,
			},
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting services matching engine GUID '%v' in namespace '%v'", engineGuid, namespaceName)
		}
		var service *apiv1.Service
		if servicesForId, found := services[engineGuidStr]; found {
			if len(servicesForId) > 1 {
				return nil, stacktrace.NewError(
					"Expected at most one engine service in namespace '%v' for engine with GUID '%v' " +
						"but found '%v'",
					namespaceName,
					engineGuid,
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
			label_key_consts.IDKubernetesLabelKey.GetString(),
			map[string]bool{
				engineGuidStr: true,
			},
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting pods matching engine GUID '%v' in namespace '%v'", engineGuid, namespaceName)
		}
		var pod *apiv1.Pod
		if podsForId, found := pods[engineGuidStr]; found {
			if len(podsForId) > 1 {
				return nil, stacktrace.NewError(
					"Expected at most one engine pod in namespace '%v' for engine with GUID '%v' " +
						"but found '%v'",
					namespaceName,
					engineGuid,
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

func getEngineObjectsFromKubernetesResources(allResources map[engine.EngineGUID]*engineKubernetesResources) (map[engine.EngineGUID]*engine.Engine, error) {
	result := map[engine.EngineGUID]*engine.Engine{}

	for engineGuid, resourcesForId := range allResources {
		engineStatus, err := getContainerStatusFromPod(resourcesForId.pod)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting engine status from engine pod")
		}

		// NOTE: We set these to nil because in Kubernetes we have no way of knowing what the public info is!
		var publicIpAddr net.IP = nil
		var publicGrpcPortSpec *port_spec.PortSpec = nil
		var publicGrpcProxyPortSpec *port_spec.PortSpec = nil

		engineObj := engine.NewEngine(
			engineGuid,
			engineStatus,
			publicIpAddr,
			publicGrpcPortSpec,
			publicGrpcProxyPortSpec,
		)
		result[engineGuid] = engineObj
	}
	return result, nil
}

func getEngineMatchLabels() map[string]string {
	engineMatchLabels := map[string]string{
		label_key_consts.AppIDKubernetesLabelKey.GetString():                label_value_consts.AppIDKubernetesLabelValue.GetString(),
		label_key_consts.KurtosisResourceTypeKubernetesLabelKey.GetString(): label_value_consts.EngineKurtosisResourceTypeKubernetesLabelValue.GetString(),
	}
	return engineMatchLabels
}

func (backend *KubernetesKurtosisBackend) createEngineNamespace(
	ctx context.Context,
	engineAttributesProvider object_attributes_provider.KubernetesEngineObjectAttributesProvider,
) (*apiv1.Namespace, error) {
	// Get Namespace Attributes
	engineNamespaceAttributes, err := engineAttributesProvider.ForEngineNamespace()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting the Kubernetes attributes for the namespace",
		)
	}
	engineNamespaceName := engineNamespaceAttributes.GetName().GetString()
	engineNamespaceLabels := getStringMapFromLabelMap(engineNamespaceAttributes.GetLabels())

	//Create engine's namespace
	engineNamespace, err := backend.kubernetesManager.CreateNamespace(ctx, engineNamespaceName, engineNamespaceLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the namespace '%v' using labels '%+v'", engineNamespace, engineNamespaceLabels)
	}
	return engineNamespace, nil
}

func (backend *KubernetesKurtosisBackend) createEngineServiceAccount(
	ctx context.Context,
	namespace string,
	engineAttributesProvider object_attributes_provider.KubernetesEngineObjectAttributesProvider,
) (*apiv1.ServiceAccount, error) {
	serviceAccountAttributes, err := engineAttributesProvider.ForEngineServiceAccount()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Expected to be able to get Kubernetes attributes for engine service account, instead got a non-nil error",
		)
	}
	serviceAccountName := serviceAccountAttributes.GetName().GetString()
	serviceAccountLabels := getStringMapFromLabelMap(serviceAccountAttributes.GetLabels())
	serviceAccount, err := backend.kubernetesManager.CreateServiceAccount(ctx, serviceAccountName, namespace, serviceAccountLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating service account '%v' with labels '%+v' in namespace '%v'", serviceAccountName, serviceAccountLabels, namespace)
	}
	return serviceAccount, nil
}

func (backend *KubernetesKurtosisBackend) createEngineClusterRole(
	ctx context.Context,
	engineAttributesProvider object_attributes_provider.KubernetesEngineObjectAttributesProvider,
) (*rbacv1.ClusterRole, error) {
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
			Verbs:     []string{
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
				consts.NamespacesKubernetesResource,
				consts.ServiceAccountsKubernetesResource,
				consts.RolesKubernetesResource,
				consts.RoleBindingsKubernetesResource,
				consts.PodsKubernetesResource,
				consts.PodExecsKubernetesResource,
				consts.PodLogsKubernetesResource,
				consts.ServicesKubernetesResource,
				consts.PersistentVolumeClaimsKubernetesResource,
				consts.JobsKubernetesResource,  // Necessary so that we can give the API container the permission
			},
		},
	}
	clusterRole, err := backend.kubernetesManager.CreateClusterRoles(ctx, clusterRoleName, clusterRolePolicyRules, clusterRoleLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating cluster role '%v' with policy rules '%+v' and labels '%+v'", clusterRoleName, clusterRolePolicyRules, clusterRoleLabels)
	}
	return clusterRole, nil
}

func (backend *KubernetesKurtosisBackend) createEngineClusterRoleBindings(
	ctx context.Context,
	engineAttributesProvider object_attributes_provider.KubernetesEngineObjectAttributesProvider,
	clusterRoleName string,
	namespaceName string,
	serviceAccountName string,
) (*rbacv1.ClusterRoleBinding, error) {
	clusterRoleBindingsAttributes, err := engineAttributesProvider.ForEngineClusterRoleBindings()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Expected to be able to get Kubernetes attributes for engine cluster role bindings, but instead got a non-nil error",
		)
	}
	clusterRoleBindingsName := clusterRoleBindingsAttributes.GetName().GetString()
	clusterRoleBindingsLabels := getStringMapFromLabelMap(clusterRoleBindingsAttributes.GetLabels())
	clusterRoleBindingsSubjects := []rbacv1.Subject{
		{
			Kind:      rbacv1.ServiceAccountKind,
			Name:      serviceAccountName,
			Namespace: namespaceName,
		},
	}
	clusterRoleBindingsRoleRef := rbacv1.RoleRef{
		APIGroup: consts.RbacAuthorizationApiGroup,
		Kind:     consts.ClusterRoleKubernetesResourceType,
		Name:     clusterRoleName,
	}
	clusterRoleBindings, err := backend.kubernetesManager.CreateClusterRoleBindings(ctx, clusterRoleBindingsName, clusterRoleBindingsSubjects, clusterRoleBindingsRoleRef, clusterRoleBindingsLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating cluster role bindings '%v' with subjects '%+v' and role ref '%+v' in namespace '%v'", clusterRoleBindingsName, clusterRoleBindingsSubjects, clusterRoleBindingsRoleRef, namespaceName)
	}
	return clusterRoleBindings, nil
}

func (backend *KubernetesKurtosisBackend) createEnginePod(
	ctx context.Context,
	namespace string,
	engineAttributesProvider object_attributes_provider.KubernetesEngineObjectAttributesProvider,
	imageOrgAndRepo string,
	imageVersionTag string,
	envVars map[string]string,
	privatePorts map[string]*port_spec.PortSpec,
	serviceAccountName string,
) (*apiv1.Pod, map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue, error) {
	// Get Pod Attributes
	enginePodAttributes, err := engineAttributesProvider.ForEnginePod()
	if err != nil {
		return nil, nil, stacktrace.Propagate(
			err,
			"Expected to be able to get Kubernetes attributes for engine pod, instead got a non-nil error",
		)
	}
	enginePodName := enginePodAttributes.GetName().GetString()
	enginePodLabels := enginePodAttributes.GetLabels()
	enginePodLabelStrs := getStringMapFromLabelMap(enginePodLabels)
	enginePodAnnotationStrs := getStringMapFromAnnotationMap(enginePodAttributes.GetAnnotations())

	// Define Containers in our Engine Pod and hook them up to our Engine Volumes
	containerImageAndTag := fmt.Sprintf(
		"%v:%v",
		imageOrgAndRepo,
		imageVersionTag,
	)

	containerPorts, err := getKubernetesContainerPortsFromPrivatePortSpecs(privatePorts)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting the engine container ports from the private port specs")
	}

	var engineContainerEnvVars []apiv1.EnvVar
	for varName, varValue := range envVars {
		envVar := apiv1.EnvVar{
			Name:  varName,
			Value: varValue,
		}
		engineContainerEnvVars = append(engineContainerEnvVars, envVar)
	}
	engineContainers := []apiv1.Container{
		{
			Name:  kurtosisEngineContainerName,
			Image: containerImageAndTag,
			Env:   engineContainerEnvVars,
			Ports: containerPorts,
		},
	}

	engineVolumes := []apiv1.Volume{}

	engineInitContainers := []apiv1.Container{}

	// Create pods with engine containers and volumes in kubernetes
	pod, err := backend.kubernetesManager.CreatePod(
		ctx,
		namespace,
		enginePodName,
		enginePodLabelStrs,
		enginePodAnnotationStrs,
		engineInitContainers,
		engineContainers,
		engineVolumes,
		serviceAccountName,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred while creating the pod with name '%s' in namespace '%s' with image '%s'", enginePodName, namespace, containerImageAndTag)
	}
	return pod, enginePodLabels, nil
}

func (backend *KubernetesKurtosisBackend) createEngineService(
	ctx context.Context,
	namespace string,
	engineAttributesProvider object_attributes_provider.KubernetesEngineObjectAttributesProvider,
	privateGrpcPortSpec *port_spec.PortSpec,
	privateGrpcProxyPortSpec *port_spec.PortSpec,
	podMatchLabels map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue,
) (*apiv1.Service, error) {
	engineServiceAttributes, err := engineAttributesProvider.ForEngineService(
		kurtosisInternalContainerGrpcPortSpecId,
		privateGrpcPortSpec,
		kurtosisInternalContainerGrpcProxyPortSpecId,
		privateGrpcProxyPortSpec)
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

	// Define service ports. These hook up to ports on the containers running in the engine pod
	servicePorts, err := getKubernetesServicePortsFromPrivatePortSpecs(map[string]*port_spec.PortSpec{
		kurtosisInternalContainerGrpcPortSpecId: privateGrpcPortSpec,
		kurtosisInternalContainerGrpcProxyPortSpecId: privateGrpcProxyPortSpec,
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the engine service's ports using the engine private port specs")
	}

	podMatchLabelStrs := getStringMapFromLabelMap(podMatchLabels)

	// Create Service
	service, err := backend.kubernetesManager.CreateService(
		ctx,
		namespace,
		engineServiceName,
		engineServiceLabels,
		engineServiceAnnotations,
		podMatchLabelStrs,
		apiv1.ServiceTypeClusterIP,
		servicePorts,
	)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred while creating the service with name '%s' in namespace '%s' with ports '%v' and '%v'",
			engineServiceName,
			namespace,
			privateGrpcPortSpec.GetNumber(),
			privateGrpcProxyPortSpec.GetNumber(),
		)
	}
	return service, nil
}
