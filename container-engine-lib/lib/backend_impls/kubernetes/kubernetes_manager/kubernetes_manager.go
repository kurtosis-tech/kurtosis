/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package kubernetes_manager

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_value"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_value"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	apiv1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	applyconfigurationsv1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultPersistentVolumeClaimAccessMode = apiv1.ReadWriteOnce
	binaryMegabytesSuffix         		   = "Mi"
	uintToIntStringConversionBase		   = 10

	waitForPersistentVolumeBoundInitialDelayMilliSeconds = 100
	waitForPersistentVolumeBoundTimeout = 60 * time.Second
	waitForPersistentVolumeBoundRetriesDelayMilliSeconds = 500

	podWaitForAvailabilityTimeout = 15 * time.Minute
	podWaitForAvailabilityTimeBetweenPolls = 500 * time.Millisecond

	// This is a container "reason" (machine-readable string) indicating that the container has some issue with
	// pulling the image (usually, a typo in the image name or the image doesn't exist)
	// Pods in this state don't really recover on their own
	imagePullBackOffContainerReason = "ImagePullBackOff"

	containerStatusLineBulletPoint = " - "

	// Kubernetes unfortunately doesn't have a good way to get the exit code out, so we have to parse it out of a string
	expectedTerminationMessage				= "command terminated with exit code"

	shouldAllocateStdinOnPodExec = false
	shouldAllocatedStdoutOnPodExec = true
	shouldAllocatedStderrOnPodExec = true
	shouldAllocateTtyOnPodExec = false

	successExecCommandExitCode = 0

	// This is the owner string we'll use when updating fields
	fieldManager = "kurtosis"

	shouldFollowContainerLogsWhenPrintingPodInfo = false
	shouldAddTimestampsWhenPrintingPodInfo = true
)

var (
	globalDeletePolicy  = metav1.DeletePropagationForeground
	globalDeleteOptions = metav1.DeleteOptions{
		PropagationPolicy: &globalDeletePolicy,
	}
	globalCreateOptions = metav1.CreateOptions{
		// We need every object to have this field manager so that the Kurtosis objects can all seamlessly modify Kubernetes resources
		FieldManager:    fieldManager,
	}
)

type KubernetesManager struct {
	// The underlying K8s client that will be used to modify the K8s environment
	kubernetesClientSet *kubernetes.Clientset
	// Underlying restClient configuration
	kuberneteRestConfig *rest.Config

}

func NewKubernetesManager(kubernetesClientSet *kubernetes.Clientset, kuberneteRestConfig *rest.Config) *KubernetesManager {
	return &KubernetesManager{
		kubernetesClientSet: kubernetesClientSet,
		kuberneteRestConfig: kuberneteRestConfig,
	}
}

// ---------------------------Services------------------------------------------------------------------------------

// CreateService creates a k8s service in the specified namespace. It connects pods to the service according to the pod labels passed in
func (manager *KubernetesManager) CreateService(ctx context.Context, namespace string, name string, serviceLabels map[string]string, serviceAnnotations map[string]string, matchPodLabels map[string]string, serviceType apiv1.ServiceType, ports []apiv1.ServicePort) (*apiv1.Service, error) {
	servicesClient := manager.kubernetesClientSet.CoreV1().Services(namespace)

	objectMeta := metav1.ObjectMeta{
		Name:        name,
		Labels:      serviceLabels,
		Annotations: serviceAnnotations,
	}

	// Figure out selector api

	// There must be a better way
	serviceSpec := apiv1.ServiceSpec{
		Ports:    ports,
		Selector: matchPodLabels, // these labels are used to match with the Pod
		Type:     serviceType,
	}

	service := &apiv1.Service{
		ObjectMeta: objectMeta,
		Spec:       serviceSpec,
	}

	serviceResult, err := servicesClient.Create(ctx, service, globalCreateOptions)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create service '%s' in namespace '%s'", name, namespace)
	}

	return serviceResult, nil
}

func (manager *KubernetesManager) RemoveService(ctx context.Context, service *apiv1.Service) error {
	namespace := service.Namespace
	serviceName := service.Name
	servicesClient := manager.kubernetesClientSet.CoreV1().Services(namespace)

	if err := servicesClient.Delete(ctx, serviceName, globalDeleteOptions); err != nil {
		return stacktrace.Propagate(err, "Failed to delete service '%s' with delete options '%+v' in namespace '%s'", serviceName, globalDeleteOptions, namespace)
	}

	return nil
}

func (manager *KubernetesManager) UpdateService(
	ctx context.Context,
	namespaceName string,
	serviceName string,
	// We use a configurator, rather than letting the user pass in their own ServiceApplyConfiguration, so that we ensure
	// they use the constructor (and don't do struct instantiation and forget to add the namespace, object name, etc. which
	// would result in removing the object name)
	updateConfigurator func(configuration *applyconfigurationsv1.ServiceApplyConfiguration),
) (*apiv1.Service, error) {
	updatesToApply := applyconfigurationsv1.Service(serviceName, namespaceName)
	updateConfigurator(updatesToApply)

	servicesClient := manager.kubernetesClientSet.CoreV1().Services(namespaceName)

	applyOpts := metav1.ApplyOptions{
		FieldManager: fieldManager,
	}
	result, err := servicesClient.Apply(ctx, updatesToApply, applyOpts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to update service '%v' in namespace '%v'", serviceName, namespaceName)
	}
	return result, nil
}

func (manager *KubernetesManager) GetServicesByLabels(ctx context.Context, namespace string, serviceLabels map[string]string) (*apiv1.ServiceList, error) {
	servicesClient := manager.kubernetesClientSet.CoreV1().Services(namespace)

	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(serviceLabels).String(),
	}

	serviceResult, err := servicesClient.List(ctx, listOptions)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to list services with labels '%+v' in namespace '%s'", serviceLabels, namespace)
	}

	// Only return objects not tombstoned by Kubernetes
	var servicesNotMarkedForDeletionList []apiv1.Service
	for _, service := range serviceResult.Items {
		deletionTimestamp := service.GetObjectMeta().GetDeletionTimestamp()
		if deletionTimestamp == nil {
			servicesNotMarkedForDeletionList = append(servicesNotMarkedForDeletionList, service)
		}
	}
	servicesNotMarkedForDeletionserviceList := apiv1.ServiceList{
		Items:    servicesNotMarkedForDeletionList,
		TypeMeta: serviceResult.TypeMeta,
		ListMeta: serviceResult.ListMeta,
	}
	
	return &servicesNotMarkedForDeletionserviceList, nil
}

// ---------------------------Volumes------------------------------------------------------------------------------
// TODO Delete this after 2022-08-01 if we're still not using PersistentVolumeClaims
/*
func (manager *KubernetesManager) CreatePersistentVolumeClaim(
	ctx context.Context,
	namespace string,
	persistentVolumeClaimName string,
	persistentVolumeClaimLabels map[string]string,
	persistentVolumeClaimAnnotations map[string]string,
	volumeSizeInMegabytes uint,
	storageClassName string,
) (*apiv1.PersistentVolumeClaim, error) {
	volumeClaimsClient := manager.kubernetesClientSet.CoreV1().PersistentVolumeClaims(namespace)

	volumeSizeInMegabytesStr := strconv.FormatUint(uint64(volumeSizeInMegabytes), uintToIntStringConversionBase)
	quantity, err := resource.ParseQuantity(volumeSizeInMegabytesStr + binaryMegabytesSuffix)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to parse volume size in megabytes %d", volumeSizeInMegabytes)
	}

	persistentVolumeClaim := &apiv1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:   persistentVolumeClaimName,
			Labels: persistentVolumeClaimLabels,
			Namespace: namespace,
		},
		Spec: apiv1.PersistentVolumeClaimSpec{
			AccessModes: []apiv1.PersistentVolumeAccessMode{
				defaultPersistentVolumeClaimAccessMode,
			},
			StorageClassName: &storageClassName,
			Resources: apiv1.ResourceRequirements{
				Requests: map[apiv1.ResourceName]resource.Quantity{
					apiv1.ResourceStorage: quantity,
				},
			},
		},
	}

	if _, err := volumeClaimsClient.Create(ctx, persistentVolumeClaim, globalCreateOptions); err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create persistent volume claim with name '%s' in namespace '%s'", persistentVolumeClaimName, namespace)
	}

	// Wait for the PVC to become bound and return that object (which will have VolumeName filled out)
	result, err := manager.waitForPersistentVolumeClaimBinding(ctx, namespace, persistentVolumeClaimName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for persistent volume claim '%v' get bound in namespace '%v'", persistentVolumeClaim.GetName(), persistentVolumeClaim.GetNamespace())
	}

	return result, nil
}

func (manager *KubernetesManager) RemovePersistentVolumeClaim(ctx context.Context, volumeClaim *apiv1.PersistentVolumeClaim) error {
	namespace := volumeClaim.Namespace
	name := volumeClaim.Name
	volumeClaimsClient := manager.kubernetesClientSet.CoreV1().PersistentVolumeClaims(namespace)

	if err := volumeClaimsClient.Delete(ctx, name, globalDeleteOptions); err != nil {
		return stacktrace.Propagate(err, "Failed to delete persistent volume claim with name '%s' with delete options '%+v' in namespace '%s'", name, globalDeleteOptions, namespace)
	}
	return nil
}

func (manager *KubernetesManager) GetPersistentVolumeClaim(ctx context.Context, namespace string, persistentVolumeClaimName string) (*apiv1.PersistentVolumeClaim, error) {
	persistentVolumeClaimsClient := manager.kubernetesClientSet.CoreV1().PersistentVolumeClaims(namespace)

	volumeClaim, err := persistentVolumeClaimsClient.Get(ctx, persistentVolumeClaimName, metav1.GetOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get persistent volume claim with name '%s' in namespace '%s'", persistentVolumeClaimName, namespace)
	}
	
	deletionTimestamp := volumeClaim.GetObjectMeta().GetDeletionTimestamp()
	if deletionTimestamp != nil {
		return nil, stacktrace.Propagate(err, "Persistent volume claim with name '%s' in namespace '%s' has been marked for deletion", persistentVolumeClaimName, namespace)
	}
	return volumeClaim, nil
}

// TODO Make return type an actual list
func (manager *KubernetesManager) GetPersistentVolumeClaimsByLabels(ctx context.Context, namespace string, persistentVolumeClaimLabels map[string]string) (*apiv1.PersistentVolumeClaimList, error) {
	persistentVolumeClaimsClient := manager.kubernetesClientSet.CoreV1().PersistentVolumeClaims(namespace)

	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(persistentVolumeClaimLabels).String(),
	}

	persistentVolumeClaimsResult, err := persistentVolumeClaimsClient.List(ctx, listOptions)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get persistent volume claim with labels '%+v' in namespace '%s'", persistentVolumeClaimLabels, namespace)
	}
 
	// Only return objects not tombstoned by Kubernetes
	var persistentVolumeClaimsNotMarkedForDeletionList []apiv1.PersistentVolumeClaim
	for _, persistentVolumeClaim := range persistentVolumeClaimsResult.Items {
		deletionTimestamp := persistentVolumeClaim.GetObjectMeta().GetDeletionTimestamp()
		if deletionTimestamp == nil {
			persistentVolumeClaimsNotMarkedForDeletionList = append(persistentVolumeClaimsNotMarkedForDeletionList, persistentVolumeClaim)
		}
	}
	persistentVolumeClaimsNotMarkedForDeletionpersistentVolumeClaimList := apiv1.PersistentVolumeClaimList{
		Items:    persistentVolumeClaimsNotMarkedForDeletionList,
		TypeMeta: persistentVolumeClaimsResult.TypeMeta,
		ListMeta: persistentVolumeClaimsResult.ListMeta,
	}
	return &persistentVolumeClaimsNotMarkedForDeletionpersistentVolumeClaimList, nil
}

 */

// ---------------------------namespaces------------------------------------------------------------------------------

func (manager *KubernetesManager) CreateNamespace(ctx context.Context, name string, namespaceLabels map[string]string) (*apiv1.Namespace, error) {
	namespaceClient := manager.kubernetesClientSet.CoreV1().Namespaces()

	namespace := &apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: namespaceLabels,
		},
	}

	namespaceResult, err := namespaceClient.Create(ctx, namespace, globalCreateOptions)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create namespace with name '%s'", name)
	}

	return namespaceResult, nil
}

func (manager *KubernetesManager) RemoveNamespace(ctx context.Context, namespace *apiv1.Namespace) error {
	name := namespace.Name
	namespaceClient := manager.kubernetesClientSet.CoreV1().Namespaces()

	if err := namespaceClient.Delete(ctx, name, globalDeleteOptions); err != nil {
		return stacktrace.Propagate(err, "Failed to delete namespace with name '%s' with delete options '%+v'", name, globalDeleteOptions)
	}

	return nil
}

func (manager *KubernetesManager) GetNamespace(ctx context.Context, name string) (*apiv1.Namespace, error) {
	namespaceClient := manager.kubernetesClientSet.CoreV1().Namespaces()

	namespace, err := namespaceClient.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get namespace with name '%s'", name)
	}
	deletionTimestamp := namespace.GetObjectMeta().GetDeletionTimestamp()
	if deletionTimestamp != nil {
		return nil, stacktrace.Propagate(err, "Namespace with name '%s' has been marked for deletion", namespace)
	}
	return namespace, nil
}

func (manager *KubernetesManager) GetNamespacesByLabels(ctx context.Context, namespaceLabels map[string]string) (*apiv1.NamespaceList, error) {
	namespaceClient := manager.kubernetesClientSet.CoreV1().Namespaces()

	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(namespaceLabels).String(),
	}

	namespaces, err := namespaceClient.List(ctx, listOptions)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to list namespaces with labels '%+v'", namespaceLabels)
	}

	// Only return objects not tombstoned by Kubernetes
	var namespacesNotMarkedForDeletionList []apiv1.Namespace
	for _, namespace := range namespaces.Items {
		deletionTimestamp := namespace.GetObjectMeta().GetDeletionTimestamp()
		if deletionTimestamp == nil {
			namespacesNotMarkedForDeletionList = append(namespacesNotMarkedForDeletionList, namespace)
		}
	}
	namespacesNotMarkedForDeletionnamespaceList := apiv1.NamespaceList{
		Items:    namespacesNotMarkedForDeletionList,
		TypeMeta: namespaces.TypeMeta,
		ListMeta: namespaces.ListMeta,
	}
	return &namespacesNotMarkedForDeletionnamespaceList, nil
}

// ---------------------------service accounts------------------------------------------------------------------------------

func (manager *KubernetesManager) CreateServiceAccount(ctx context.Context, name string, namespace string, labels map[string]string) (*apiv1.ServiceAccount, error) {
	client := manager.kubernetesClientSet.CoreV1().ServiceAccounts(namespace)

	serviceAccount := &apiv1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}

	serviceAccountResult, err := client.Create(ctx, serviceAccount, globalCreateOptions)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create service account with name '%s' in namespace '%v'", name, namespace)
	}
	return serviceAccountResult, nil
}

func (manager *KubernetesManager) GetServiceAccountsByLabels(ctx context.Context, namespace string, serviceAccountsLabels map[string]string) (*apiv1.ServiceAccountList, error) {
	client := manager.kubernetesClientSet.CoreV1().ServiceAccounts(namespace)

	opts := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(serviceAccountsLabels).String(),
	}

	serviceAccounts, err := client.List(ctx, opts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get service accounts with labels '%+v', instead a non-nil error was returned", serviceAccountsLabels)
	}

	// Only return objects not tombstoned by Kubernetes
	var serviceAccountsNotMarkedForDeletionList []apiv1.ServiceAccount
	for _, serviceAccount := range serviceAccounts.Items {
		deletionTimestamp := serviceAccount.GetObjectMeta().GetDeletionTimestamp()
		if deletionTimestamp == nil {
			serviceAccountsNotMarkedForDeletionList = append(serviceAccountsNotMarkedForDeletionList, serviceAccount)
		}
	}
	serviceAccountsNotMarkedForDeletionserviceAccountList := apiv1.ServiceAccountList{
		Items:    serviceAccountsNotMarkedForDeletionList,
		TypeMeta: serviceAccounts.TypeMeta,
		ListMeta: serviceAccounts.ListMeta,
	}
	return &serviceAccountsNotMarkedForDeletionserviceAccountList, nil
}

func (manager *KubernetesManager) RemoveServiceAccount(ctx context.Context, serviceAccount *apiv1.ServiceAccount) error {
	name := serviceAccount.Name
	namespace := serviceAccount.Namespace
	client := manager.kubernetesClientSet.CoreV1().ServiceAccounts(namespace)

	if err := client.Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		return stacktrace.Propagate(err, "Failed to delete service account with name '%s' in namespace '%v'", name, namespace)
	}

	return nil
}

// ---------------------------roles------------------------------------------------------------------------------

func (manager *KubernetesManager) CreateRole(ctx context.Context, name string, namespace string, rules []rbacv1.PolicyRule, labels map[string]string) (*rbacv1.Role, error) {
	client := manager.kubernetesClientSet.RbacV1().Roles(namespace)

	role :=  &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Rules: rules,
	}

	roleResult, err := client.Create(ctx, role, globalCreateOptions)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create role with name '%s' in namespace '%v' and rules '%+v'", name, namespace, rules)
	}

	return roleResult, nil
}

func (manager *KubernetesManager) GetRolesByLabels(ctx context.Context, namespace string, rolesLabels map[string]string) (*rbacv1.RoleList, error) {
	client := manager.kubernetesClientSet.RbacV1().Roles(namespace)

	opts := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(rolesLabels).String(),
	}

	roles, err := client.List(ctx, opts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get roles with labels '%+v', instead a non-nil error was returned", rolesLabels)
	}

	// Only return objects not tombstoned by Kubernetes
	var rolesNotMarkedForDeletionList []rbacv1.Role
	for _, role := range roles.Items {
		deletionTimestamp := role.GetObjectMeta().GetDeletionTimestamp()
		if deletionTimestamp == nil {
			rolesNotMarkedForDeletionList = append(rolesNotMarkedForDeletionList, role)
		}
	}
	rolesNotMarkedForDeletionroleList := rbacv1.RoleList{
		Items:    rolesNotMarkedForDeletionList,
		TypeMeta: roles.TypeMeta,
		ListMeta: roles.ListMeta,
	}
	return &rolesNotMarkedForDeletionroleList, nil
}

func (manager *KubernetesManager) RemoveRole(ctx context.Context, role *rbacv1.Role) error {
	name := role.Name
	namespace := role.Namespace
	client := manager.kubernetesClientSet.RbacV1().Roles(namespace)

	if err := client.Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		return stacktrace.Propagate(err, "Failed to delete role with name '%s' in namespace '%v'", name, namespace)
	}

	return nil
}

// --------------------------- Role Bindings ------------------------------------------------------------------------------

func (manager *KubernetesManager) CreateRoleBindings(ctx context.Context, name string, namespace string, subjects []rbacv1.Subject, roleRef rbacv1.RoleRef, labels map[string]string) (*rbacv1.RoleBinding, error) {
	client := manager.kubernetesClientSet.RbacV1().RoleBindings(namespace)

	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Subjects: subjects,
		RoleRef: roleRef,
	}

	roleBindingResult, err := client.Create(ctx, roleBinding, globalCreateOptions)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create role binding with name '%s', subjects '%+v' and role ref '%v'", name, subjects, roleRef)
	}

	return roleBindingResult, nil
}

func (manager *KubernetesManager) GetRoleBindingsByLabels(ctx context.Context, namespace string, roleBindingsLabels map[string]string) (*rbacv1.RoleBindingList, error) {
	client := manager.kubernetesClientSet.RbacV1().RoleBindings(namespace)

	opts := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(roleBindingsLabels).String(),
	}

	roleBindings, err := client.List(ctx, opts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get role bindings with labels '%+v', instead a non-nil error was returned", roleBindingsLabels)
	}

	// Only return objects not tombstoned by Kubernetes
	var roleBindingsNotMarkedForDeletionList []rbacv1.RoleBinding
	for _, roleBinding := range roleBindings.Items {
		deletionTimestamp := roleBinding.GetObjectMeta().GetDeletionTimestamp()
		if deletionTimestamp == nil {
			roleBindingsNotMarkedForDeletionList = append(roleBindingsNotMarkedForDeletionList, roleBinding)
		}
	}
	roleBindingsNotMarkedForDeletionroleBindingList := rbacv1.RoleBindingList{
		Items:    roleBindingsNotMarkedForDeletionList,
		TypeMeta: roleBindings.TypeMeta,
		ListMeta: roleBindings.ListMeta,
	}
	return &roleBindingsNotMarkedForDeletionroleBindingList, nil
}

func (manager *KubernetesManager) RemoveRoleBindings(ctx context.Context, roleBinding *rbacv1.RoleBinding) error {
	name := roleBinding.Name
	namespace := roleBinding.Namespace
	client := manager.kubernetesClientSet.RbacV1().RoleBindings(namespace)

	if err := client.Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		return stacktrace.Propagate(err, "Failed to delete role bindings with name '%s' in namespace '%v'", name, namespace)
	}

	return nil
}

// ---------------------------cluster roles------------------------------------------------------------------------------

func (manager *KubernetesManager) CreateClusterRoles(ctx context.Context, name string, rules []rbacv1.PolicyRule, labels map[string]string) (*rbacv1.ClusterRole, error) {
	client := manager.kubernetesClientSet.RbacV1().ClusterRoles()

	clusterRole :=  &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Rules: rules,
	}

	clusterRoleResult, err := client.Create(ctx, clusterRole, globalCreateOptions)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create cluster role with name '%s' with rules '%+v'", name, rules)
	}

	return clusterRoleResult, nil
}

func (manager *KubernetesManager) GetClusterRolesByLabels(ctx context.Context, clusterRoleLabels map[string]string) (*rbacv1.ClusterRoleList, error) {
	client := manager.kubernetesClientSet.RbacV1().ClusterRoles()

	opts := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(clusterRoleLabels).String(),
	}

	clusterRoles, err := client.List(ctx, opts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get cluster roles with labels '%+v', instead a non-nil error was returned", clusterRoleLabels)
	}
	
	// Only return objects not tombstoned by Kubernetes
	var clusterRolesNotMarkedForDeletionList []rbacv1.ClusterRole
	for _, clusterRole := range clusterRoles.Items {
		deletionTimestamp := clusterRole.GetObjectMeta().GetDeletionTimestamp()
		if deletionTimestamp == nil {
			clusterRolesNotMarkedForDeletionList = append(clusterRolesNotMarkedForDeletionList, clusterRole)
		}
	}
	clusterRolesNotMarkedForDeletionclusterRoleList := rbacv1.ClusterRoleList{
		Items:    clusterRolesNotMarkedForDeletionList,
		TypeMeta: clusterRoles.TypeMeta,
		ListMeta: clusterRoles.ListMeta,
	}
	return &clusterRolesNotMarkedForDeletionclusterRoleList, nil
}

func (manager *KubernetesManager) RemoveClusterRole(ctx context.Context, clusterRole *rbacv1.ClusterRole) error {
	name := clusterRole.Name
	client := manager.kubernetesClientSet.RbacV1().ClusterRoles()

	if err := client.Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		return stacktrace.Propagate(err, "Failed to delete cluster role with name '%s'", name)
	}

	return nil
}

// --------------------------- Cluster Role Bindings ------------------------------------------------------------------------------

func (manager *KubernetesManager) CreateClusterRoleBindings(ctx context.Context, name string, subjects []rbacv1.Subject, roleRef rbacv1.RoleRef, labels map[string]string) (*rbacv1.ClusterRoleBinding, error) {
	client := manager.kubernetesClientSet.RbacV1().ClusterRoleBindings()

	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Subjects: subjects,
		RoleRef: roleRef,
	}

	clusterRoleBindingResult, err := client.Create(ctx, clusterRoleBinding, globalCreateOptions)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create cluster role binding with name '%s', subjects '%+v' and role ref '%v'", name, subjects, roleRef)
	}

	return clusterRoleBindingResult, nil
}

func (manager *KubernetesManager) GetClusterRoleBindingsByLabels(ctx context.Context, clusterRoleBindingsLabels map[string]string) (*rbacv1.ClusterRoleBindingList, error) {
	client := manager.kubernetesClientSet.RbacV1().ClusterRoleBindings()

	opts := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(clusterRoleBindingsLabels).String(),
	}

	clusterRoleBindings, err := client.List(ctx, opts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get cluster role bindings with labels '%+v', instead a non-nil error was returned", clusterRoleBindingsLabels)
	}

	// Only return objects not tombstoned by Kubernetes
	var clusterRoleBindingsNotMarkedForDeletionList []rbacv1.ClusterRoleBinding
	for _, clusterRoleBindings := range clusterRoleBindings.Items {
		deletionTimestamp := clusterRoleBindings.GetObjectMeta().GetDeletionTimestamp()
		if deletionTimestamp == nil {
			clusterRoleBindingsNotMarkedForDeletionList = append(clusterRoleBindingsNotMarkedForDeletionList, clusterRoleBindings)
		}
	}
	clusterRoleBindingssNotMarkedForDeletionclusterRoleBindingsList := rbacv1.ClusterRoleBindingList{
		Items:    clusterRoleBindingsNotMarkedForDeletionList,
		TypeMeta: clusterRoleBindings.TypeMeta,
		ListMeta: clusterRoleBindings.ListMeta,
	}
	return &clusterRoleBindingssNotMarkedForDeletionclusterRoleBindingsList, nil
}

func (manager *KubernetesManager) RemoveClusterRoleBindings(ctx context.Context, clusterRoleBinding *rbacv1.ClusterRoleBinding) error {
	name := clusterRoleBinding.Name
	client := manager.kubernetesClientSet.RbacV1().ClusterRoleBindings()

	if err := client.Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		return stacktrace.Propagate(err, "Failed to delete cluster role binding with name '%s'", name)
	}

	return nil
}

// ---------------------------pods---------------------------------------------------------------------------------------

func (manager *KubernetesManager) CreatePod(
	ctx context.Context,
	namespaceName string,
	podName string,
	podLabels map[string]string,
	podAnnotations map[string]string,
	initContainers []apiv1.Container,
	podContainers []apiv1.Container,
	podVolumes []apiv1.Volume,
	podServiceAccountName string,
) (*apiv1.Pod, error) {
	podClient := manager.kubernetesClientSet.CoreV1().Pods(namespaceName)

	podMeta := metav1.ObjectMeta{
		Name:        podName,
		Labels:      podLabels,
		Annotations: podAnnotations,
	}
	podSpec := apiv1.PodSpec{
		Volumes:    podVolumes,
		InitContainers: initContainers,
		Containers: podContainers,
		ServiceAccountName: podServiceAccountName,
		// We don't want Kubernetes auto-magically restarting our containers if they fail
		RestartPolicy: apiv1.RestartPolicyNever,
	}

	podToCreate := &apiv1.Pod{
		Spec:       podSpec,
		ObjectMeta: podMeta,
	}

	if podDefinitionBytes, err := json.Marshal(podToCreate); err == nil {
		logrus.Debugf("Going to start pod using the following JSON: %v", string(podDefinitionBytes))
	}

	createdPod, err := podClient.Create(ctx, podToCreate, globalCreateOptions)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to create pod with name '%v' and labels '%+v', instead a non-nil error was returned", podName, podLabels)
	}

	if err := manager.waitForPodAvailability(ctx, namespaceName, podName); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for pod '%v' to become available", podName)
	}

	return createdPod, nil
}

func (manager *KubernetesManager) RemovePod(ctx context.Context, pod *apiv1.Pod) error {
	name := pod.Name
	namespace := pod.Namespace
	client := manager.kubernetesClientSet.CoreV1().Pods(namespace)

	if err := client.Delete(ctx, name, globalDeleteOptions); err != nil {
		return stacktrace.Propagate(err, "Failed to delete pod with name '%s' with delete options '%+v'", name, globalDeleteOptions)
	}

	return nil
}

func (manager *KubernetesManager) GetPod(ctx context.Context, namespace string, name string) (*apiv1.Pod, error) {
	podClient := manager.kubernetesClientSet.CoreV1().Pods(namespace)

	pod, err := podClient.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get pod with name '%s'", name)
	}
	deletionTimestamp := pod.GetObjectMeta().GetDeletionTimestamp()
	if deletionTimestamp != nil {
		return nil, stacktrace.Propagate(err, "Pod with name '%s' in namespace '%s' has been marked for deletion", name, namespace)
	}

	return pod, nil
}

// GetContainerLogs gets the logs for a given container running inside the given pod in the give namespace
// TODO We could upgrade this to get the logs of many containers at once just like kubectl does, see:
//  https://github.com/kubernetes/kubectl/blob/master/pkg/cmd/logs/logs.go#L345
func (manager *KubernetesManager) GetContainerLogs(
	ctx context.Context,
	namespaceName string,
	podName string,
	containerName string,
	shouldFollowLogs bool,
	shouldAddTimestamps bool,
) (
	io.ReadCloser,
	error,
){
	options := &apiv1.PodLogOptions{
		Container: containerName,
		Follow: shouldFollowLogs,
		Timestamps: shouldAddTimestamps,
	}

	getLogsRequest := manager.kubernetesClientSet.CoreV1().Pods(namespaceName).GetLogs(podName, options)
	result, err := getLogsRequest.Stream(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting logs for container '%v' in pod '%v' in namespace '%v'",
			containerName,
			podName,
			namespaceName,
		)
	}
	return result, nil
}

func (manager *KubernetesManager) RunExecCommand(
	namespaceName string,
	podName string,
	containerName string,
	command []string,
	stdOutOutput io.Writer,
	stdErrOutput io.Writer,
) (
	resultExitCode int32,
	resultErr error,
) {
	execOptions := &apiv1.PodExecOptions{
		Container: containerName,
		Command: command,
		Stdin:   shouldAllocateStdinOnPodExec,
		Stdout:  shouldAllocatedStdoutOnPodExec,
		Stderr:  shouldAllocatedStderrOnPodExec,
		TTY:     shouldAllocateTtyOnPodExec,
	}

	//Create a RESTful command request.
	request := manager.kubernetesClientSet.CoreV1().RESTClient().
		Post().
		Namespace(namespaceName).
		Resource("pods").
		Name(podName).
		SubResource("exec").
		VersionedParams(execOptions, scheme.ParameterCodec)
	if request == nil {
		return -1, stacktrace.NewError(
			"Failed to build a working RESTful request for the command '%s'.",
			execOptions.Command,
		)
	}

	exec, err := remotecommand.NewSPDYExecutor(manager.kuberneteRestConfig, http.MethodPost, request.URL())
	if err != nil {
		return -1, stacktrace.Propagate(
			err,
			"Failed to build an executor for the command '%s' with the RESTful endpoint '%s'.",
			execOptions.Command,
			request.URL().String(),
		)
	}

	if err = exec.Stream(remotecommand.StreamOptions{
		Stdout: stdOutOutput,
		Stderr: stdErrOutput,
	}); err != nil {
		// Kubernetes returns the exit code of the command via a string in the error message, so we have to extract it
		statusError := err.Error()
		exitCode, err := getExitCodeFromStatusMessage(statusError)
		if err != nil {
			return exitCode, stacktrace.Propagate(
				err,
				"There was an error trying to parse the message '%s' to an exit code.",
				statusError,
			)
		}

		return exitCode, nil
	}

	return successExecCommandExitCode, nil
}


func (manager *KubernetesManager) GetPodsByLabels(ctx context.Context, namespace string, podLabels map[string]string) (*apiv1.PodList, error) {
	namespacePodClient := manager.kubernetesClientSet.CoreV1().Pods(namespace)

	opts := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(podLabels).String(),
	}

	pods, err := namespacePodClient.List(ctx, opts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get pods with labels '%+v', instead a non-nil error was returned", podLabels)
	}

	// Only return objects not tombstoned by Kubernetes
	var podsNotMarkedForDeletionList []apiv1.Pod
	for _, pod := range pods.Items {
		deletionTimestamp := pod.GetObjectMeta().GetDeletionTimestamp()
		if deletionTimestamp == nil {
			podsNotMarkedForDeletionList = append(podsNotMarkedForDeletionList, pod)
		}
	}
	podsNotMarkedForDeletionPodList := apiv1.PodList{
		Items:    podsNotMarkedForDeletionList,
		TypeMeta: pods.TypeMeta,
		ListMeta: pods.ListMeta,
	}
	return &podsNotMarkedForDeletionPodList, nil
}

func (manager *KubernetesManager) GetPodPortforwardEndpointUrl(namespace string, podName string) *url.URL {
	return manager.kubernetesClientSet.CoreV1().RESTClient().Post().Resource("pods").Namespace(namespace).Name(podName).SubResource("portforward").URL()
}


// TODO Delete this after 2022-08-01 if we're not using Jobs
/*
func (manager *KubernetesManager) CreateJobWithContainerAndVolume(ctx context.Context,
	namespaceName string,
	jobName *kubernetes_object_name.KubernetesObjectName,
	jobLabels map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue,
	jobAnnotations map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue,
	containers []apiv1.Container,
	volumes []apiv1.Volume,
	numRetries int32,
	ttlSecondsAfterFinished uint,
) (*v1.Job, error) {

	jobsClient := manager.kubernetesClientSet.BatchV1().Jobs(namespaceName)
	ttlSecondsAfterFinishedInt32 := int32(ttlSecondsAfterFinished)

	labelStrs := transformTypedLabelsToStrs(jobLabels)
	annotationStrs := transformTypedAnnotationsToStrs(jobAnnotations)

	jobMeta := metav1.ObjectMeta{
		Name:                       jobName.GetString(),
		Labels:                     labelStrs,
		Annotations:                annotationStrs,
	}

	podSpec := apiv1.PodSpec{
		Containers: containers,
		Volumes: volumes,
		// We don't want Kubernetes automagically restarting our containers
		RestartPolicy: apiv1.RestartPolicyNever,
	}

	jobSpec := v1.JobSpec{
		BackoffLimit: &numRetries,
		Template:                apiv1.PodTemplateSpec{
			Spec: podSpec,
		},
		TTLSecondsAfterFinished: &ttlSecondsAfterFinishedInt32,
	}

	jobInput := v1.Job{
		ObjectMeta: jobMeta,
		Spec:       jobSpec,
	}

	logrus.Debugf("Job resource to create: %+v", jobInput)

	job, err := jobsClient.Create(ctx, &jobInput, globalCreateOptions)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Failed to create job '%v' in namespace '%v' with containers '%+v' and volumes '%+v'",
			jobName,
			namespaceName,
			containers,
			volumes,
		)
	}

	return job, nil
}

func (manager *KubernetesManager) DeleteJob(ctx context.Context, namespace string, job *v1.Job) error {
	jobsClient := manager.kubernetesClientSet.BatchV1().Jobs(namespace)
	if jobsClient == nil {
		return stacktrace.NewError("Failed to create a jobs client for namespace '%v'", namespace)
	}
	jobName := job.Name

	if err := jobsClient.Delete(ctx, jobName, globalDeleteOptions); err != nil {
		return stacktrace.Propagate(err, "Failed to delete job '%v' in namespace '%v' with delete options '%+v'", jobName, namespace, globalDeleteOptions)
	}

	return nil
}

func (manager KubernetesManager) GetJobCompletionAndSuccessFlags(ctx context.Context, namespace string, jobName string) (hasCompleted bool, isSuccess bool, resultErr error) {
	job, err := manager.kubernetesClientSet.BatchV1().Jobs(namespace).Get(ctx, jobName, metav1.GetOptions{})
	if err != nil {
		return false, false, stacktrace.Propagate(err, "Failed to get job status for job name '%v' in namespace '%v'", jobName, namespace)
	}

	deletionTimestamp := job.GetObjectMeta().GetDeletionTimestamp()
	if deletionTimestamp != nil {
		return false, false, stacktrace.Propagate(err, "Job with name '%s' in namespace '%s' has been marked for deletion", job.GetName(), namespace)
	}

	// LOGIC FROM https://stackoverflow.com/a/69262406

	// Job hasn't spun up yet
	if job.Status.Active == 0 && job.Status.Succeeded == 0 && job.Status.Failed == 0 {
		return false, false, nil
	}

	// Job is active
	if job.Status.Active > 0 {
		return false, false, nil
	}

	// Job succeeded
	if job.Status.Succeeded > 0 {
		return true, true, nil // Job ran successfully
	}

	return true, false, nil
}
 */

// ====================================================================================================
//                                     Private Helper Methods
// ====================================================================================================
func transformTypedLabelsToStrs(input map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue) map[string]string {
	result := map[string]string{}
	for key, value := range input {
		result[key.GetString()] = value.GetString()
	}
	return result
}

func transformTypedAnnotationsToStrs(input map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue) map[string]string {
	result := map[string]string{}
	for key, value := range input {
		result[key.GetString()] = value.GetString()
	}
	return result
}

// TODO Delete this after 2022-08-01 if we're still not using PersistentVolumeClaims
/*
func (manager *KubernetesManager) waitForPersistentVolumeClaimBinding(
	ctx context.Context,
	namespaceName string,
	persistentVolumeClaimName string,
) (*apiv1.PersistentVolumeClaim, error) {
	deadline := time.Now().Add(waitForPersistentVolumeBoundTimeout)
	time.Sleep(time.Duration(waitForPersistentVolumeBoundInitialDelayMilliSeconds) * time.Millisecond)
	var result *apiv1.PersistentVolumeClaim
	for time.Now().Before(deadline) {
		claim, err := manager.GetPersistentVolumeClaim(ctx, namespaceName, persistentVolumeClaimName)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting persistent volume claim '%v' in namespace '%v", persistentVolumeClaimName, namespaceName)
		}
		result = claim
		claimStatus := claim.Status
		claimPhase := claimStatus.Phase

		switch claimPhase {
		//Success phase, the Persistent Volume got bound
		case apiv1.ClaimBound:
			return result, nil
		//Lost the Persistent Volume phase, unrecoverable state
		case apiv1.ClaimLost:
			return nil, stacktrace.NewError(
				"The persistent volume claim '%v' ended up in unrecoverable state '%v'",
				claim.GetName(),
				claimPhase,
			)
		}

		time.Sleep(time.Duration(waitForPersistentVolumeBoundRetriesDelayMilliSeconds) * time.Millisecond)
	}

	return nil, stacktrace.NewError(
		"Persistent volume claim '%v' in namespace '%v' did not become bound despite waiting for %v with %v " +
			"between polls",
		persistentVolumeClaimName,
		namespaceName,
		waitForPersistentVolumeBoundTimeout,
		waitForPersistentVolumeBoundRetriesDelayMilliSeconds,
	)
}

 */

func (manager *KubernetesManager) waitForPodAvailability(ctx context.Context, namespaceName string, podName string) error {
	// Wait for the pod to start running
	deadline := time.Now().Add(podWaitForAvailabilityTimeout)
	var latestPodStatus *apiv1.PodStatus
	for time.Now().Before(deadline) {
		pod, err := manager.GetPod(ctx, namespaceName, podName)
		if err != nil {
			// We shouldn't get an error on getting the pod, even if it's not ready
			return stacktrace.Propagate(err, "An error occurred getting the just-created pod '%v'", podName)
		}

		latestPodStatus = &pod.Status
		switch latestPodStatus.Phase {
		case apiv1.PodRunning:
			return nil
		case apiv1.PodPending:
			for _, containerStatus := range pod.Status.ContainerStatuses {
				containerName := containerStatus.Name
				maybeContainerWaitingState := containerStatus.State.Waiting
				if maybeContainerWaitingState != nil && maybeContainerWaitingState.Reason == imagePullBackOffContainerReason {
					return stacktrace.NewError(
						"Container '%v' using image '%v' in pod '%v' in namespace '%v' is stuck in state '%v'. This likely means:\n" +
							"1) There's a typo in either the image name or the tag name\n" +
							"2) The image isn't accessible to Kubernetes (e.g. it's a local image, or it's in a private image registry that Kubernetes can't access)",
						containerName,
						containerStatus.Image,
						pod.Name,
						namespaceName,
						imagePullBackOffContainerReason,
					)
				}
			}
		case apiv1.PodFailed:
			podStateStr := manager.getPodInfoBlockStr(ctx, namespaceName, pod)
			return stacktrace.NewError(
				"Pod '%v' failed before availability with the following state:\n%v",
				podName,
				podStateStr,
			)
		case apiv1.PodSucceeded:
			podStateStr := manager.getPodInfoBlockStr(ctx, namespaceName, pod)
			// NOTE: We'll need to change this if we ever expect to run one-off pods
			return stacktrace.NewError(
				"Expected state of pod '%v' to arrive at '%v' but the pod instead landed in '%v' with the following state:\n%v",
				podName,
				apiv1.PodRunning,
				apiv1.PodSucceeded,
				podStateStr,
			)
		}
		time.Sleep(podWaitForAvailabilityTimeBetweenPolls)
	}

	containerStatusStrs := renderContainerStatuses(latestPodStatus.ContainerStatuses, containerStatusLineBulletPoint)
	return stacktrace.NewError(
		"Pod '%v' did not become available after %v; its latest state is '%v' and status message is: %v\n" +
			"The pod's container states are as follows:\n%v",
		podName,
		podWaitForAvailabilityTimeout,
		latestPodStatus.Phase,
		latestPodStatus.Message,
		strings.Join(containerStatusStrs, "\n"),
	)
}

func (manager *KubernetesManager) getPodInfoBlockStr(
	ctx context.Context,
	namespaceName string,
	pod *apiv1.Pod,
) string {
	podName := pod.Name

	// TODO Parallelize to increase perf? But make sure we don't explode memory with huge pod logs
	// We go through all this work so that the user can get detailed information about their pod without needing to dive
	// through the Kubernetes API
	resultStrBuilder := strings.Builder{}
	resultStrBuilder.WriteString(fmt.Sprintf(
		">>>>>>>>>>>>>>>>>>>>>>>>>> Pod %v <<<<<<<<<<<<<<<<<<<<<<<<<<\n",
		podName,
	))
	resultStrBuilder.WriteString("Container Statuses:")
	for _, containerStatusStr := range renderContainerStatuses(pod.Status.ContainerStatuses, containerStatusLineBulletPoint) {
		resultStrBuilder.WriteString(containerStatusStr)
		resultStrBuilder.WriteString("\n")
	}
	resultStrBuilder.WriteString("\n")
	for _, podContainer := range pod.Spec.Containers {
		containerName := podContainer.Name

		resultStrBuilder.WriteString(fmt.Sprintf(
			"-------------------- Container %v Logs --------------------\n",
			podContainer.Name,
		))

		containerLogsStr := manager.getSingleContainerLogs(ctx, namespaceName, podName, containerName)
		resultStrBuilder.WriteString(containerLogsStr)
		resultStrBuilder.WriteString("\n")

		resultStrBuilder.WriteString(fmt.Sprintf(
			"------------------ End Container %v Logs ---------------------\n",
			containerName,
		))
	}
	resultStrBuilder.WriteString(fmt.Sprintf(
		">>>>>>>>>>>>>>>>>>>>>>>> End Pod %v <<<<<<<<<<<<<<<<<<<<<<<<<<<",
		pod.Name,
	))

	return resultStrBuilder.String()
}

func (manager *KubernetesManager) getSingleContainerLogs(ctx context.Context, namespaceName string, podName string, containerName string) string {
	containerLogs, err := manager.GetContainerLogs(ctx, namespaceName, podName, containerName, shouldFollowContainerLogsWhenPrintingPodInfo, shouldAddTimestampsWhenPrintingPodInfo)
	defer containerLogs.Close()
	if err != nil {
		return fmt.Sprintf("Cannot display container logs because an error occurred getting the logs:\n%v", err)
	}

	buffer := &bytes.Buffer{}
	if _, copyErr := io.Copy(buffer, containerLogs); copyErr != nil {
		return fmt.Sprintf("Cannot display container logs because an error occurred saving the logs to memory:\n%v", err)
	}
	return buffer.String()
}

func renderContainerStatuses(containerStatuses []apiv1.ContainerStatus, prefixStr string) []string {
	containerStatusStrs := []string{}
	for _, containerStatus := range containerStatuses {
		containerName := containerStatus.Name
		state := containerStatus.State


		// Okay to do in an if-else because only one will be filled out per Kubernetes docs
		var statusStrForContainer string
		if state.Waiting != nil {
			statusStrForContainer = fmt.Sprintf(
				"WAITING - %v",
				state.Waiting.Message,
			)
		} else if state.Running != nil {
			statusStrForContainer = fmt.Sprintf(
				"RUNNING since %v",
				state.Running.StartedAt,
			)
		} else if state.Terminated != nil {
			terminatedState := state.Terminated
			statusStrForContainer = fmt.Sprintf(
				"TERMINATED with exit code %v - %v",
				terminatedState.ExitCode,
				terminatedState.Message,
			)
		} else {
			statusStrForContainer = fmt.Sprintf(
				"Unrecogznied container state '%+v'; this likely means that Kubernetes " +
					"has added a new container state and Kurtosis needs to be updated to handle it",
				state,
			)
		}

		strForContainer := fmt.Sprintf(
			"%v%v (%v): %v",
			prefixStr,
			containerName,
			containerStatus.Image,
			statusStrForContainer,
		)

		containerStatusStrs = append(
			containerStatusStrs,
			strForContainer,
		)
	}

	return containerStatusStrs
}

// Kubernetes doesn't seem to have a nice API for getting back the exit code of a command (though this seems strange??),
// so we have to parse it out of a status message
func getExitCodeFromStatusMessage(statusMessage string) (int32, error){
	messageSlice := strings.Split(statusMessage, " ")
	if len(messageSlice) != 6 {
		return -1, stacktrace.NewError(
			"Expected the status message to have 6 parts but it has '%v'. This is likely not an exit status message.\n'%s'",
			len(messageSlice),
			statusMessage,
		)
	}

	terminationBaseMessage := strings.Join(messageSlice[0:5], " ")
	if terminationBaseMessage != expectedTerminationMessage {
		return -1, stacktrace.NewError(
			"Received a termination message of '%s' when we expected a message following the pattern of '%s'. This is likely not an exit status message.",
			statusMessage,
			expectedTerminationMessage,
		)
	}

	codeAsString := messageSlice[5]
	codeAsInt64, err := strconv.ParseInt(codeAsString, 0, 32)
	if err != nil {
		return -1, stacktrace.Propagate(err, "Failed to convert '%s' to a base32 int.", codeAsString)
	}
	codeAsInt32 := int32(codeAsInt64)
	return codeAsInt32, nil
}

