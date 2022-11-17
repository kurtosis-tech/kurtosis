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
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net"
	"time"
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
			engineResources = &engineKubernetesResources{
				clusterRole: &rbacv1.ClusterRole{
					TypeMeta: metav1.TypeMeta{
						Kind:       "",
						APIVersion: "",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:            "",
						GenerateName:    "",
						Namespace:       "",
						SelfLink:        "",
						UID:             "",
						ResourceVersion: "",
						Generation:      0,
						CreationTimestamp: metav1.Time{
							Time: time.Time{},
						},
						DeletionTimestamp: &metav1.Time{
							Time: time.Time{},
						},
						DeletionGracePeriodSeconds: nil,
						Labels:                     nil,
						Annotations:                nil,
						OwnerReferences:            nil,
						Finalizers:                 nil,
						ZZZ_DeprecatedClusterName:  "",
						ManagedFields:              nil,
					},
					Rules: nil,
					AggregationRule: &rbacv1.AggregationRule{
						ClusterRoleSelectors: nil,
					},
				},
				clusterRoleBinding: &rbacv1.ClusterRoleBinding{
					TypeMeta: metav1.TypeMeta{
						Kind:       "",
						APIVersion: "",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:            "",
						GenerateName:    "",
						Namespace:       "",
						SelfLink:        "",
						UID:             "",
						ResourceVersion: "",
						Generation:      0,
						CreationTimestamp: metav1.Time{
							Time: time.Time{},
						},
						DeletionTimestamp: &metav1.Time{
							Time: time.Time{},
						},
						DeletionGracePeriodSeconds: nil,
						Labels:                     nil,
						Annotations:                nil,
						OwnerReferences:            nil,
						Finalizers:                 nil,
						ZZZ_DeprecatedClusterName:  "",
						ManagedFields:              nil,
					},
					Subjects: nil,
					RoleRef: rbacv1.RoleRef{
						APIGroup: "",
						Kind:     "",
						Name:     "",
					},
				},
				namespace: &apiv1.Namespace{
					TypeMeta: metav1.TypeMeta{
						Kind:       "",
						APIVersion: "",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:            "",
						GenerateName:    "",
						Namespace:       "",
						SelfLink:        "",
						UID:             "",
						ResourceVersion: "",
						Generation:      0,
						CreationTimestamp: metav1.Time{
							Time: time.Time{},
						},
						DeletionTimestamp: &metav1.Time{
							Time: time.Time{},
						},
						DeletionGracePeriodSeconds: nil,
						Labels:                     nil,
						Annotations:                nil,
						OwnerReferences:            nil,
						Finalizers:                 nil,
						ZZZ_DeprecatedClusterName:  "",
						ManagedFields:              nil,
					},
					Spec: apiv1.NamespaceSpec{
						Finalizers: nil,
					},
					Status: apiv1.NamespaceStatus{
						Phase:      "",
						Conditions: nil,
					},
				},
				serviceAccount: &apiv1.ServiceAccount{
					TypeMeta: metav1.TypeMeta{
						Kind:       "",
						APIVersion: "",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:            "",
						GenerateName:    "",
						Namespace:       "",
						SelfLink:        "",
						UID:             "",
						ResourceVersion: "",
						Generation:      0,
						CreationTimestamp: metav1.Time{
							Time: time.Time{},
						},
						DeletionTimestamp: &metav1.Time{
							Time: time.Time{},
						},
						DeletionGracePeriodSeconds: nil,
						Labels:                     nil,
						Annotations:                nil,
						OwnerReferences:            nil,
						Finalizers:                 nil,
						ZZZ_DeprecatedClusterName:  "",
						ManagedFields:              nil,
					},
					Secrets:                      nil,
					ImagePullSecrets:             nil,
					AutomountServiceAccountToken: nil,
				},
				service: &apiv1.Service{
					TypeMeta: metav1.TypeMeta{
						Kind:       "",
						APIVersion: "",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:            "",
						GenerateName:    "",
						Namespace:       "",
						SelfLink:        "",
						UID:             "",
						ResourceVersion: "",
						Generation:      0,
						CreationTimestamp: metav1.Time{
							Time: time.Time{},
						},
						DeletionTimestamp: &metav1.Time{
							Time: time.Time{},
						},
						DeletionGracePeriodSeconds: nil,
						Labels:                     nil,
						Annotations:                nil,
						OwnerReferences:            nil,
						Finalizers:                 nil,
						ZZZ_DeprecatedClusterName:  "",
						ManagedFields:              nil,
					},
					Spec: apiv1.ServiceSpec{
						Ports:                    nil,
						Selector:                 nil,
						ClusterIP:                "",
						ClusterIPs:               nil,
						Type:                     "",
						ExternalIPs:              nil,
						SessionAffinity:          "",
						LoadBalancerIP:           "",
						LoadBalancerSourceRanges: nil,
						ExternalName:             "",
						ExternalTrafficPolicy:    "",
						HealthCheckNodePort:      0,
						PublishNotReadyAddresses: false,
						SessionAffinityConfig: &apiv1.SessionAffinityConfig{
							ClientIP: &apiv1.ClientIPConfig{
								TimeoutSeconds: nil,
							},
						},
						IPFamilies:                    nil,
						IPFamilyPolicy:                nil,
						AllocateLoadBalancerNodePorts: nil,
						LoadBalancerClass:             nil,
						InternalTrafficPolicy:         nil,
					},
					Status: apiv1.ServiceStatus{
						LoadBalancer: apiv1.LoadBalancerStatus{
							Ingress: nil,
						},
						Conditions: nil,
					},
				},
				pod: &apiv1.Pod{
					TypeMeta: metav1.TypeMeta{
						Kind:       "",
						APIVersion: "",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:            "",
						GenerateName:    "",
						Namespace:       "",
						SelfLink:        "",
						UID:             "",
						ResourceVersion: "",
						Generation:      0,
						CreationTimestamp: metav1.Time{
							Time: time.Time{},
						},
						DeletionTimestamp: &metav1.Time{
							Time: time.Time{},
						},
						DeletionGracePeriodSeconds: nil,
						Labels:                     nil,
						Annotations:                nil,
						OwnerReferences:            nil,
						Finalizers:                 nil,
						ZZZ_DeprecatedClusterName:  "",
						ManagedFields:              nil,
					},
					Spec: apiv1.PodSpec{
						Volumes:                       nil,
						InitContainers:                nil,
						Containers:                    nil,
						EphemeralContainers:           nil,
						RestartPolicy:                 "",
						TerminationGracePeriodSeconds: nil,
						ActiveDeadlineSeconds:         nil,
						DNSPolicy:                     "",
						NodeSelector:                  nil,
						ServiceAccountName:            "",
						DeprecatedServiceAccount:      "",
						AutomountServiceAccountToken:  nil,
						NodeName:                      "",
						HostNetwork:                   false,
						HostPID:                       false,
						HostIPC:                       false,
						ShareProcessNamespace:         nil,
						SecurityContext: &apiv1.PodSecurityContext{
							SELinuxOptions: &apiv1.SELinuxOptions{
								User:  "",
								Role:  "",
								Type:  "",
								Level: "",
							},
							WindowsOptions: &apiv1.WindowsSecurityContextOptions{
								GMSACredentialSpecName: nil,
								GMSACredentialSpec:     nil,
								RunAsUserName:          nil,
								HostProcess:            nil,
							},
							RunAsUser:           nil,
							RunAsGroup:          nil,
							RunAsNonRoot:        nil,
							SupplementalGroups:  nil,
							FSGroup:             nil,
							Sysctls:             nil,
							FSGroupChangePolicy: nil,
							SeccompProfile: &apiv1.SeccompProfile{
								Type:             "",
								LocalhostProfile: nil,
							},
						},
						ImagePullSecrets: nil,
						Hostname:         "",
						Subdomain:        "",
						Affinity: &apiv1.Affinity{
							NodeAffinity: &apiv1.NodeAffinity{
								RequiredDuringSchedulingIgnoredDuringExecution: &apiv1.NodeSelector{
									NodeSelectorTerms: nil,
								},
								PreferredDuringSchedulingIgnoredDuringExecution: nil,
							},
							PodAffinity: &apiv1.PodAffinity{
								RequiredDuringSchedulingIgnoredDuringExecution:  nil,
								PreferredDuringSchedulingIgnoredDuringExecution: nil,
							},
							PodAntiAffinity: &apiv1.PodAntiAffinity{
								RequiredDuringSchedulingIgnoredDuringExecution:  nil,
								PreferredDuringSchedulingIgnoredDuringExecution: nil,
							},
						},
						SchedulerName:     "",
						Tolerations:       nil,
						HostAliases:       nil,
						PriorityClassName: "",
						Priority:          nil,
						DNSConfig: &apiv1.PodDNSConfig{
							Nameservers: nil,
							Searches:    nil,
							Options:     nil,
						},
						ReadinessGates:            nil,
						RuntimeClassName:          nil,
						EnableServiceLinks:        nil,
						PreemptionPolicy:          nil,
						Overhead:                  nil,
						TopologySpreadConstraints: nil,
						SetHostnameAsFQDN:         nil,
						OS: &apiv1.PodOS{
							Name: "",
						},
					},
					Status: apiv1.PodStatus{
						Phase:             "",
						Conditions:        nil,
						Message:           "",
						Reason:            "",
						NominatedNodeName: "",
						HostIP:            "",
						PodIP:             "",
						PodIPs:            nil,
						StartTime: &metav1.Time{
							Time: time.Time{},
						},
						InitContainerStatuses:      nil,
						ContainerStatuses:          nil,
						QOSClass:                   "",
						EphemeralContainerStatuses: nil,
					},
				},
			}
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
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while collecting matching cluster roles")
	}
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
			engineResources = &engineKubernetesResources{
				clusterRole:        nil,
				clusterRoleBinding: nil,
				namespace:          nil,
				serviceAccount:     nil,
				service:            nil,
				pod:                nil,
			}
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
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while collecting matching cluster role bindings")
	}
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
			engineResources = &engineKubernetesResources{
				clusterRole:        nil,
				clusterRoleBinding: nil,
				namespace:          nil,
				serviceAccount:     nil,
				service:            nil,
				pod:                nil,
			}
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
					"Expected at most one engine service account in namespace '%v' for engine with GUID '%v' "+
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
					"Expected at most one engine service in namespace '%v' for engine with GUID '%v' "+
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
					"Expected at most one engine pod in namespace '%v' for engine with GUID '%v' "+
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
