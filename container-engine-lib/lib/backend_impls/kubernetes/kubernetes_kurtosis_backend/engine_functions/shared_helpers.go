package engine_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_resource_collectors"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	apiv1 "k8s.io/api/core/v1"
	"net"
)

func getEngineObjectsFromKubernetesResources(allResources map[engine.EngineGUID]*engineKubernetesResources) (map[engine.EngineGUID]*engine.Engine, error) {
	result := map[engine.EngineGUID]*engine.Engine{}

	for engineGuid, resourcesForId := range allResources {
		engineStatus, err := shared_helpers.GetContainerStatusFromPod(resourcesForId.pod)
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

func getMatchingEngineObjectsAndKubernetesResources(
	ctx context.Context,
	filters *engine.EngineFilters,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (
	map[engine.EngineGUID]*engine.Engine,
	map[engine.EngineGUID]*engineKubernetesResources,
	error,
) {
	matchingResources, err := getMatchingEngineKubernetesResources(ctx, filters.GUIDs, kubernetesManager)
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
func getMatchingEngineKubernetesResources(
	ctx context.Context,
	engineGuids map[engine.EngineGUID]bool,
	kubernetesManager *kubernetes_manager.KubernetesManager,
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
		kubernetesManager,
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
		kubernetesManager,
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
		kubernetesManager,
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
			kubernetesManager,
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
			kubernetesManager,
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
			kubernetesManager,
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

func getEngineMatchLabels() map[string]string {
	engineMatchLabels := map[string]string{
		label_key_consts.AppIDKubernetesLabelKey.GetString():                label_value_consts.AppIDKubernetesLabelValue.GetString(),
		label_key_consts.KurtosisResourceTypeKubernetesLabelKey.GetString(): label_value_consts.EngineKurtosisResourceTypeKubernetesLabelValue.GetString(),
	}
	return engineMatchLabels
}
