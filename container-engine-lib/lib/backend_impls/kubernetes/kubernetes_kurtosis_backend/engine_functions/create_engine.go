package engine_functions

import (
	"context"
	"fmt"
	"time"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	kubernetes_manager_consts "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager/consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_value"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

const (
	kurtosisEngineContainerName = "kurtosis-engine-container"

	maxWaitForEngineContainerAvailabilityRetries         = 30
	timeBetweenWaitForEngineContainerAvailabilityRetries = 1 * time.Second
	httpApplicationProtocol                              = "http"
)

var noWait *port_spec.Wait = nil

// TODO add support for passing toleration to Engine
var noToleration []apiv1.Toleration = nil

func CreateEngine(
	ctx context.Context,
	imageOrgAndRepo string,
	imageVersionTag string,
	grpcPortNum uint16,
	envVars map[string]string,
	kubernetesManager *kubernetes_manager.KubernetesManager,
	objAttrsProvider object_attributes_provider.KubernetesObjectAttributesProvider,
	_ bool, //It's not required to add extra configuration in K8S for enabling the debug server
) (
	*engine.Engine,
	error,
) {
	hasNodes, err := kubernetesManager.HasComputeNodes(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while verifying whether the Kubernetes cluster has any compute nodes")
	}
	if !hasNodes {
		return nil, stacktrace.NewError("Can't start engine on the Kubernetes cluster as it has no compute nodes")
	}

	engineGuidStr, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred generating a UUID string for the engine")
	}
	engineGuid := engine.EngineGUID(engineGuidStr)

	privateGrpcPortSpec, err := port_spec.NewPortSpec(grpcPortNum, consts.KurtosisServersTransportProtocol, httpApplicationProtocol, noWait)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the engine's private grpc port spec object using number '%v' and protocol '%v'",
			grpcPortNum,
			consts.KurtosisServersTransportProtocol.String(),
		)
	}
	privateRESTAPIPortSpec, err := port_spec.NewPortSpec(engine.RESTAPIPortAddr, consts.KurtosisServersTransportProtocol, httpApplicationProtocol, noWait)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the engine's private rest api port spec object using number '%v' and protocol '%v'",
			engine.RESTAPIPortAddr,
			consts.KurtosisServersTransportProtocol.String(),
		)
	}
	privatePortSpecs := map[string]*port_spec.PortSpec{
		consts.KurtosisInternalContainerGrpcPortSpecId: privateGrpcPortSpec,
	}

	engineAttributesProvider := objAttrsProvider.ForEngine(engineGuid)

	namespace, err := createEngineNamespace(ctx, engineAttributesProvider, kubernetesManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the engine namespace")
	}
	shouldRemoveNamespace := true
	defer func() {
		if shouldRemoveNamespace {
			if err := kubernetesManager.RemoveNamespace(ctx, namespace); err != nil {
				logrus.Errorf("Creating the engine didn't complete successfully, so we tried to delete Kubernetes namespace '%v' that we created but an error was thrown:\n%v", namespace.Name, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove Kubernetes namespace with name '%v'!!!!!!!", namespace.Name)
			}
		}
	}()
	namespaceName := namespace.Name

	serviceAccount, err := createEngineServiceAccount(ctx, namespaceName, engineAttributesProvider, kubernetesManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the engine service account")
	}
	shouldRemoveServiceAccount := true
	defer func() {
		if shouldRemoveServiceAccount {
			if err := kubernetesManager.RemoveServiceAccount(ctx, serviceAccount); err != nil {
				logrus.Errorf("Creating the engine didn't complete successfully, so we tried to delete service account '%v' in namespace '%v' that we created but an error was thrown:\n%v", serviceAccount.Name, namespaceName, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove service account with name '%v'!!!!!!!", serviceAccount.Name)
			}
		}
	}()

	clusterRole, err := createEngineClusterRole(ctx, engineAttributesProvider, kubernetesManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the engine cluster role")
	}
	shouldRemoveClusterRole := true
	defer func() {
		if shouldRemoveClusterRole {
			if err := kubernetesManager.RemoveClusterRole(ctx, clusterRole); err != nil {
				logrus.Errorf("Creating the engine didn't complete successfully, so we tried to delete cluster role '%v' that we created but an error was thrown:\n%v", clusterRole.Name, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove cluster role with name '%v'!!!!!!!", clusterRole.Name)
			}
		}
	}()

	clusterRoleBindings, err := createEngineClusterRoleBindings(ctx, engineAttributesProvider, clusterRole.Name, namespaceName, serviceAccount.Name, kubernetesManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the engine cluster role bindings")
	}
	shouldRemoveClusterRoleBinding := true
	defer func() {
		if shouldRemoveClusterRoleBinding {
			if err := kubernetesManager.RemoveClusterRoleBindings(ctx, clusterRoleBindings); err != nil {
				logrus.Errorf("Creating the engine didn't complete successfully, so we tried to delete cluster role bindings '%v' in namespace '%v' that we created but an error was thrown:\n%v", clusterRoleBindings.Name, namespaceName, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove cluster role bindings with name '%v'!!!!!!!", clusterRoleBindings.Name)
			}
		}
	}()

	enginePod, enginePodLabels, err := createEnginePod(ctx, namespaceName, engineAttributesProvider, imageOrgAndRepo, imageVersionTag, envVars, privatePortSpecs, serviceAccount.Name, kubernetesManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the engine pod")
	}
	var shouldRemovePod = true
	defer func() {
		if shouldRemovePod {
			logrus.Debugf("Removing Kurtosis engine Kubernetes pod because something fails during the creation process...")
			if err := kubernetesManager.RemovePod(ctx, enginePod); err != nil {
				logrus.Errorf("Creating the engine didn't complete successfully, so we tried to delete Kubernetes pod '%v' that we created but an error was thrown:\n%v", enginePod.Name, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove Kubernetes pod with name '%v'!!!!!!!", enginePod.Name)
			}
			logrus.Debugf("Removing Kurtosis engine Kubernetes pod succesfully removed")
		}
	}()

	engineService, err := createEngineService(
		ctx,
		namespaceName,
		engineAttributesProvider,
		privateGrpcPortSpec,
		privateRESTAPIPortSpec,
		enginePodLabels,
		kubernetesManager,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the engine service")
	}
	var shouldRemoveService = true
	defer func() {
		if shouldRemoveService {
			if err := kubernetesManager.RemoveService(ctx, engineService); err != nil {
				logrus.Errorf("Creating the engine didn't complete successfully, so we tried to delete Kubernetes service '%v' that we created but an error was thrown:\n%v", engineService.Name, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove Kubernetes service with name '%v'!!!!!!!", engineService.Name)
			}
		}
	}()

	engineIngress, err := createEngineIngress(
		ctx,
		namespaceName,
		engineAttributesProvider,
		privateRESTAPIPortSpec,
		kubernetesManager,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the engine ingress")
	}
	var shouldRemoveIngress = true
	defer func() {
		if shouldRemoveIngress {
			if err := kubernetesManager.RemoveIngress(ctx, engineIngress); err != nil {
				logrus.Errorf("Creating the engine didn't complete successfully, so we tried to delete Kubernetes ingress '%v' that we created but an error was thrown:\n%v", engineIngress.Name, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove Kubernetes ingress with name '%v'!!!!!!!", engineIngress.Name)
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
		ingress:            engineIngress,
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

	if err := shared_helpers.WaitForPortAvailabilityUsingNetstat(
		kubernetesManager,
		namespaceName,
		enginePod.Name,
		kurtosisEngineContainerName,
		privateGrpcPortSpec,
		maxWaitForEngineContainerAvailabilityRetries,
		timeBetweenWaitForEngineContainerAvailabilityRetries,
	); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the engine grpc port '%v/%v' to become available", privateGrpcPortSpec.GetTransportProtocol(), privateGrpcPortSpec.GetNumber())
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
			return nil, stacktrace.Propagate(err, "An error occurred waiting for the engine grpc proxy port '%v/%v' to become available", privateGrpcProxyPortSpec.GetTransportProtocol(), privateGrpcProxyPortSpec.GetNumber())
		}*/

	shouldRemoveNamespace = false
	shouldRemoveServiceAccount = false
	shouldRemoveClusterRole = false
	shouldRemoveClusterRoleBinding = false
	shouldRemovePod = false
	shouldRemoveService = false
	shouldRemoveIngress = false
	return resultEngine, nil
}

func createEngineNamespace(
	ctx context.Context,
	engineAttributesProvider object_attributes_provider.KubernetesEngineObjectAttributesProvider,
	kubernetesManager *kubernetes_manager.KubernetesManager,
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
	engineNamespaceLabels := shared_helpers.GetStringMapFromLabelMap(engineNamespaceAttributes.GetLabels())
	emptyAnnotations := map[string]string{}

	//Create engine's namespace
	engineNamespace, err := kubernetesManager.CreateNamespace(ctx, engineNamespaceName, engineNamespaceLabels, emptyAnnotations)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the namespace '%v' using labels '%+v'", engineNamespace, engineNamespaceLabels)
	}
	return engineNamespace, nil
}

func createEngineServiceAccount(
	ctx context.Context,
	namespace string,
	engineAttributesProvider object_attributes_provider.KubernetesEngineObjectAttributesProvider,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (*apiv1.ServiceAccount, error) {
	serviceAccountAttributes, err := engineAttributesProvider.ForEngineServiceAccount()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Expected to be able to get Kubernetes attributes for engine service account, instead got a non-nil error",
		)
	}
	serviceAccountName := serviceAccountAttributes.GetName().GetString()
	serviceAccountLabels := shared_helpers.GetStringMapFromLabelMap(serviceAccountAttributes.GetLabels())
	serviceAccount, err := kubernetesManager.CreateServiceAccount(ctx, serviceAccountName, namespace, serviceAccountLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating service account '%v' with labels '%+v' in namespace '%v'", serviceAccountName, serviceAccountLabels, namespace)
	}
	return serviceAccount, nil
}

func createEngineClusterRole(
	ctx context.Context,
	engineAttributesProvider object_attributes_provider.KubernetesEngineObjectAttributesProvider,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (*rbacv1.ClusterRole, error) {
	clusterRolesAttributes, err := engineAttributesProvider.ForEngineClusterRole()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Expected to be able to get Kubernetes attributes for a engine cluster role, instead got a non-nil error",
		)
	}
	clusterRoleName := clusterRolesAttributes.GetName().GetString()
	clusterRoleLabels := shared_helpers.GetStringMapFromLabelMap(clusterRolesAttributes.GetLabels())
	clusterRolePolicyRules := []rbacv1.PolicyRule{
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
			APIGroups: []string{
				rbacv1.APIGroupAll,
			},
			Resources: []string{
				kubernetes_manager_consts.NamespacesKubernetesResource,
				kubernetes_manager_consts.ServiceAccountsKubernetesResource,
				kubernetes_manager_consts.ClusterRolesKubernetesResource,
				kubernetes_manager_consts.ClusterRoleBindingsKubernetesResource,
				kubernetes_manager_consts.RolesKubernetesResource,
				kubernetes_manager_consts.RoleBindingsKubernetesResource,
				kubernetes_manager_consts.PodsKubernetesResource,
				kubernetes_manager_consts.PodExecsKubernetesResource,
				kubernetes_manager_consts.PodLogsKubernetesResource,
				kubernetes_manager_consts.ServicesKubernetesResource,
				kubernetes_manager_consts.PersistentVolumesKubernetesResource,
				kubernetes_manager_consts.PersistentVolumeClaimsKubernetesResource,
				kubernetes_manager_consts.IngressesKubernetesResource,
				kubernetes_manager_consts.JobsKubernetesResource, // Necessary so that we can give the API container the permission
			},
		},
		{
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
	clusterRole, err := kubernetesManager.CreateClusterRoles(ctx, clusterRoleName, clusterRolePolicyRules, clusterRoleLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating cluster role '%v' with policy rules '%+v' and labels '%+v'", clusterRoleName, clusterRolePolicyRules, clusterRoleLabels)
	}
	return clusterRole, nil
}

func createEngineClusterRoleBindings(
	ctx context.Context,
	engineAttributesProvider object_attributes_provider.KubernetesEngineObjectAttributesProvider,
	clusterRoleName string,
	namespaceName string,
	serviceAccountName string,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (*rbacv1.ClusterRoleBinding, error) {
	clusterRoleBindingsAttributes, err := engineAttributesProvider.ForEngineClusterRoleBindings()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Expected to be able to get Kubernetes attributes for engine cluster role bindings, but instead got a non-nil error",
		)
	}
	clusterRoleBindingsName := clusterRoleBindingsAttributes.GetName().GetString()
	clusterRoleBindingsLabels := shared_helpers.GetStringMapFromLabelMap(clusterRoleBindingsAttributes.GetLabels())
	clusterRoleBindingsSubjects := []rbacv1.Subject{
		{
			Kind:      rbacv1.ServiceAccountKind,
			Name:      serviceAccountName,
			Namespace: namespaceName,
		},
	}
	clusterRoleBindingsRoleRef := rbacv1.RoleRef{
		APIGroup: kubernetes_manager_consts.RbacAuthorizationApiGroup,
		Kind:     kubernetes_manager_consts.ClusterRoleKubernetesResourceType,
		Name:     clusterRoleName,
	}
	clusterRoleBindings, err := kubernetesManager.CreateClusterRoleBindings(ctx, clusterRoleBindingsName, clusterRoleBindingsSubjects, clusterRoleBindingsRoleRef, clusterRoleBindingsLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating cluster role bindings '%v' with subjects '%+v' and role ref '%+v' in namespace '%v'", clusterRoleBindingsName, clusterRoleBindingsSubjects, clusterRoleBindingsRoleRef, namespaceName)
	}
	return clusterRoleBindings, nil
}

func createEnginePod(
	ctx context.Context,
	namespace string,
	engineAttributesProvider object_attributes_provider.KubernetesEngineObjectAttributesProvider,
	imageOrgAndRepo string,
	imageVersionTag string,
	envVars map[string]string,
	privatePorts map[string]*port_spec.PortSpec,
	serviceAccountName string,
	kubernetesManager *kubernetes_manager.KubernetesManager,
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
	enginePodLabelStrs := shared_helpers.GetStringMapFromLabelMap(enginePodLabels)
	enginePodAnnotationStrs := shared_helpers.GetStringMapFromAnnotationMap(enginePodAttributes.GetAnnotations())

	// Define Containers in our Engine Pod and hook them up to our Engine Volumes
	containerImageAndTag := fmt.Sprintf(
		"%v:%v",
		imageOrgAndRepo,
		imageVersionTag,
	)

	containerPorts, err := shared_helpers.GetKubernetesContainerPortsFromPrivatePortSpecs(privatePorts)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting the engine container ports from the private port specs")
	}

	var engineContainerEnvVars []apiv1.EnvVar
	for varName, varValue := range envVars {
		envVar := apiv1.EnvVar{
			Name:      varName,
			Value:     varValue,
			ValueFrom: nil,
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
	pod, err := kubernetesManager.CreatePod(
		ctx,
		namespace,
		enginePodName,
		enginePodLabelStrs,
		enginePodAnnotationStrs,
		engineInitContainers,
		engineContainers,
		engineVolumes,
		serviceAccountName,
		// Engine doesn't auto restart
		apiv1.RestartPolicyNever,
		noToleration,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred while creating the pod with name '%s' in namespace '%s' with image '%s'", enginePodName, namespace, containerImageAndTag)
	}
	return pod, enginePodLabels, nil
}

func createEngineService(
	ctx context.Context,
	namespace string,
	engineAttributesProvider object_attributes_provider.KubernetesEngineObjectAttributesProvider,
	privateGrpcPortSpec *port_spec.PortSpec,
	privateRESTAPIPortSpec *port_spec.PortSpec,
	podMatchLabels map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (*apiv1.Service, error) {
	engineServiceAttributes, err := engineAttributesProvider.ForEngineService(
		consts.KurtosisInternalContainerGrpcPortSpecId,
		privateGrpcPortSpec,
		consts.KurtosisInternalContainerRESTAPIPortSpecId,
		privateRESTAPIPortSpec)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting the engine service attributes using private grpc port spec '%+v' and private REST API port spec '%+v'",
			privateGrpcPortSpec,
			privateRESTAPIPortSpec,
		)
	}
	engineServiceName := engineServiceAttributes.GetName().GetString()
	engineServiceLabels := shared_helpers.GetStringMapFromLabelMap(engineServiceAttributes.GetLabels())
	engineServiceAnnotations := shared_helpers.GetStringMapFromAnnotationMap(engineServiceAttributes.GetAnnotations())

	// Define service ports. These hook up to ports on the containers running in the engine pod
	servicePorts, err := shared_helpers.GetKubernetesServicePortsFromPrivatePortSpecs(map[string]*port_spec.PortSpec{
		consts.KurtosisInternalContainerGrpcPortSpecId:    privateGrpcPortSpec,
		consts.KurtosisInternalContainerRESTAPIPortSpecId: privateRESTAPIPortSpec,
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the engine service's ports using the engine private port specs")
	}

	podMatchLabelStrs := shared_helpers.GetStringMapFromLabelMap(podMatchLabels)

	// Create Service
	service, err := kubernetesManager.CreateService(
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
			privateRESTAPIPortSpec.GetNumber(),
		)
	}
	return service, nil
}

func createEngineIngress(
	ctx context.Context,
	namespace string,
	engineAttributesProvider object_attributes_provider.KubernetesEngineObjectAttributesProvider,
	privateRESTAPIPortSpec *port_spec.PortSpec,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (*netv1.Ingress, error) {
	engineIngressAttributes, err := engineAttributesProvider.ForEngineIngress()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting the engine ingress attributes",
		)
	}
	engineIngressName := engineIngressAttributes.GetName().GetString()
	engineIngressLabels := shared_helpers.GetStringMapFromLabelMap(engineIngressAttributes.GetLabels())
	engineIngressAnnotations := shared_helpers.GetStringMapFromAnnotationMap(engineIngressAttributes.GetAnnotations())

	engineIngressRules, err := getEngineIngressRules(engineIngressName, privateRESTAPIPortSpec)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the user service ingress rules for ingress service with name '%v'", engineIngressName)
	}

	createdIngress, err := kubernetesManager.CreateIngress(
		ctx,
		namespace,
		engineIngressName,
		engineIngressLabels,
		engineIngressAnnotations,
		engineIngressRules,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the ingress with name '%s' in namespace '%s'", engineIngressName, namespace)
	}

	return createdIngress, nil
}

func getEngineIngressRules(
	engineIngressName string,
	privateRESTAPIPortSpec *port_spec.PortSpec,
) ([]netv1.IngressRule, error) {
	var ingressRules []netv1.IngressRule
	ingressRule := netv1.IngressRule{
		Host: engine.RESTAPIPortHostHeader,
		IngressRuleValue: netv1.IngressRuleValue{
			HTTP: &netv1.HTTPIngressRuleValue{
				Paths: []netv1.HTTPIngressPath{
					{
						Path:     consts.IngressRulePathAllPaths,
						PathType: &consts.IngressRulePathTypePrefix,
						Backend: netv1.IngressBackend{
							Service: &netv1.IngressServiceBackend{
								Name: engineIngressName,
								Port: netv1.ServiceBackendPort{
									Name:   "",
									Number: int32(privateRESTAPIPortSpec.GetNumber()),
								},
							},
							Resource: nil,
						},
					},
				},
			},
		},
	}
	ingressRules = append(ingressRules, ingressRule)

	return ingressRules, nil
}
