/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package kubernetes_manager

import (
	"context"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"net/url"
	"strconv"
)

const (
	defaultServiceProtocol                 = "TCP"
	defaultPersistentVolumeAccessMode      = apiv1.ReadWriteMany
	defaultPersistentVolumeClaimAccessMode = apiv1.ReadWriteMany
	binaryMegabytesSuffix         		   = "Mi"
	uintToIntStringConversionBase		   = 10
)

var (
	removeObjectDeletePolicy  = metav1.DeletePropagationForeground
	removeObjectDeleteOptions = metav1.DeleteOptions{
		PropagationPolicy: &removeObjectDeletePolicy,
	}
)

type KubernetesManager struct {
	// The logger that all log messages will be written to
	log *logrus.Logger // NOTE: This log should be used for all log statements - the system-wide logger should NOT be used!

	// The underlying K8s client that will be used to modify the K8s environment
	kubernetesClientSet *kubernetes.Clientset
}

/*
NewKubernetesManager
Creates a new K8s manager for manipulating the k8s cluster using the given client.

Args:
	log: The logger that this K8s manager will write all its log messages to.
	kubernetesClientSet: The k8s client that will be used when interacting with the underlying k8s cluster.
*/
func NewKubernetesManager(kubernetesClientSet *kubernetes.Clientset) *KubernetesManager {
	return &KubernetesManager{
		kubernetesClientSet: kubernetesClientSet,
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

	serviceResult, err := servicesClient.Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create service '%s' in namespace '%s'", name, namespace)
	}

	return serviceResult, nil
}

func (manager *KubernetesManager) RemoveService(ctx context.Context, namespace string, name string) error {
	servicesClient := manager.kubernetesClientSet.CoreV1().Services(namespace)

	if err := servicesClient.Delete(ctx, name, removeObjectDeleteOptions); err != nil {
		return stacktrace.Propagate(err, "Failed to delete service '%s' with delete options '%+v' in namespace '%s'", name, removeObjectDeleteOptions, namespace)
	}

	return nil
}

func (manager *KubernetesManager) RemoveSelectorsFromService(ctx context.Context, namespace string, name string) error {
	servicesClient := manager.kubernetesClientSet.CoreV1().Services(namespace)
	service, err := manager.GetServiceByName(ctx, namespace, name)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to find service with name '%v' in namespace '%v'", name, namespace)
	}
	service.Spec.Selector = make(map[string]string)

	updateOpts := metav1.UpdateOptions{}

	if _, err := servicesClient.Update(ctx, service, updateOpts); err != nil {
		return stacktrace.Propagate(err, "Failed to remove selectors from service '%v' in namespace '%v'", name, namespace)
	}

	return nil
}

func (manager *KubernetesManager) GetServiceByName(ctx context.Context, namespace string, name string) (*apiv1.Service, error) {
	servicesClient := manager.kubernetesClientSet.CoreV1().Services(namespace)

	serviceResult, err := servicesClient.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get service '%s' in namespace '%s'", name, namespace)
	}

	return serviceResult, nil
}

func (manager *KubernetesManager) ListServices(ctx context.Context, namespace string) (*apiv1.ServiceList, error) {
	servicesClient := manager.kubernetesClientSet.CoreV1().Services(namespace)

	serviceResult, err := servicesClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to list services in namespace '%s'", namespace)
	}

	return serviceResult, nil
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

	return serviceResult, nil
}

// ---------------------------Volumes------------------------------------------------------------------------------

func (manager *KubernetesManager) CreateStorageClass(ctx context.Context, name string, provisioner string, volumeBindingMode storagev1.VolumeBindingMode) (*storagev1.StorageClass, error) {
	storageClassClient := manager.kubernetesClientSet.StorageV1().StorageClasses()

	//volumeBindingMode := storagev1.VolumeBindingWaitForFirstConsumer
	//provisioner := "kubernetes.io/no-provisioner"

	storageClass := &storagev1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Provisioner:       provisioner,
		VolumeBindingMode: &volumeBindingMode,
	}

	storageClassResult, err := storageClassClient.Create(ctx, storageClass, metav1.CreateOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create storage class with name '%s'", name)
	}

	return storageClassResult, nil
}

func (manager *KubernetesManager) RemoveStorageClass(ctx context.Context, name string) error {
	storageClassClient := manager.kubernetesClientSet.StorageV1().StorageClasses()

	err := storageClassClient.Delete(ctx, name, removeObjectDeleteOptions)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to delete storage class with name '%s' with delete options '%+v'", name, removeObjectDeleteOptions)
	}

	return nil
}

func (manager *KubernetesManager) GetStorageClass(ctx context.Context, name string) (*storagev1.StorageClass, error) {
	storageClassClient := manager.kubernetesClientSet.StorageV1().StorageClasses()

	storageClassResult, err := storageClassClient.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get storage class with name '%s'", name)
	}

	return storageClassResult, nil
}

func (manager *KubernetesManager) CreatePersistentVolumeClaim(ctx context.Context, namespace string, persistentVolumeClaimName string, persistentVolumeClaimLabels map[string]string, volumeSizeInMegabytes uint, storageClassName string) (*apiv1.PersistentVolumeClaim, error) {
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

	persistentVolumeClaimResult, err := volumeClaimsClient.Create(ctx, persistentVolumeClaim, metav1.CreateOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create persistent volume claim with name '%s' in namespace '%s'", persistentVolumeClaimName, namespace)
	}

	return persistentVolumeClaimResult, nil
}

func (manager *KubernetesManager) RemovePersistentVolumeClaim(ctx context.Context, namespace string, persistentVolumeClaimName string) error {
	volumeClaimsClient := manager.kubernetesClientSet.CoreV1().PersistentVolumeClaims(namespace)

	if err := volumeClaimsClient.Delete(ctx, persistentVolumeClaimName, removeObjectDeleteOptions); err != nil {
		return stacktrace.Propagate(err, "Failed to delete persistent volume claim with name '%s' with delete options '%+v' in namespace '%s'", persistentVolumeClaimName, removeObjectDeleteOptions, namespace)
	}

	return nil
}

func (manager *KubernetesManager) GetPersistentVolumeClaim(ctx context.Context, namespace string, persistentVolumeClaimName string) (*apiv1.PersistentVolumeClaim, error) {
	persistentVolumeClaimsClient := manager.kubernetesClientSet.CoreV1().PersistentVolumeClaims(namespace)

	volumeClaim, err := persistentVolumeClaimsClient.Get(ctx, persistentVolumeClaimName, metav1.GetOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get persistent volume claim with name '%s' in namespace '%s'", persistentVolumeClaimName, namespace)
	}

	return volumeClaim, nil
}

func (manager *KubernetesManager) ListPersistentVolumeClaims(ctx context.Context, namespace string) (*apiv1.PersistentVolumeClaimList, error) {
	persistentVolumeClaimsClient := manager.kubernetesClientSet.CoreV1().PersistentVolumeClaims(namespace)

	persistentVolumeClaimsResult, err := persistentVolumeClaimsClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to list persistent volume claims in namespace '%s'", namespace)
	}

	return persistentVolumeClaimsResult, nil
}

func (manager *KubernetesManager) GetPersistentVolumeClaimsByLabels(ctx context.Context, namespace string, persistentVolumeClaimLabels map[string]string) (*apiv1.PersistentVolumeClaimList, error) {
	persistentVolumeClaimsClient := manager.kubernetesClientSet.CoreV1().PersistentVolumeClaims(namespace)

	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(persistentVolumeClaimLabels).String(),
	}

	persistentVolumeClaimsResult, err := persistentVolumeClaimsClient.List(ctx, listOptions)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get persistent volume claim with labels '%+v' in namespace '%s'", persistentVolumeClaimLabels, namespace)
	}

	return persistentVolumeClaimsResult, nil
}

// ---------------------------namespaces------------------------------------------------------------------------------

func (manager *KubernetesManager) CreateNamespace(ctx context.Context, name string, namespaceLabels map[string]string) (*apiv1.Namespace, error) {
	namespaceClient := manager.kubernetesClientSet.CoreV1().Namespaces()

	namespace := &apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: namespaceLabels,
		},
	}

	namespaceResult, err := namespaceClient.Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create namespace with name '%s'", name)
	}

	return namespaceResult, nil
}

func (manager *KubernetesManager) RemoveNamespace(ctx context.Context, name string) error {
	namespaceClient := manager.kubernetesClientSet.CoreV1().Namespaces()

	if err := namespaceClient.Delete(ctx, name, removeObjectDeleteOptions); err != nil {
		return stacktrace.Propagate(err, "Failed to delete namespace with name '%s' with delete options '%+v'", name, removeObjectDeleteOptions)
	}

	return nil
}

func (manager *KubernetesManager) GetNamespace(ctx context.Context, name string) (*apiv1.Namespace, error) {
	namespaceClient := manager.kubernetesClientSet.CoreV1().Namespaces()

	namespace, err := namespaceClient.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get namespace with name '%s'", name)
	}

	return namespace, nil
}

func (manager *KubernetesManager) ListNamespaces(ctx context.Context) (*apiv1.NamespaceList, error) {
	namespaceClient := manager.kubernetesClientSet.CoreV1().Namespaces()

	namespaces, err := namespaceClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to list namespaces")
	}

	return namespaces, nil
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

	return namespaces, nil
}

// ---------------------------DaemonSets------------------------------------------------------------------------------

func (manager *KubernetesManager) CreateDaemonSet(ctx context.Context, namespace string, name string, daemonSetLabels map[string]string) (*appsv1.DaemonSet, error) {
	daemonSetClient := manager.kubernetesClientSet.AppsV1().DaemonSets(namespace)

	daemonSet := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: daemonSetLabels,
		},
		Spec: appsv1.DaemonSetSpec{},
	}

	daemonSetResult, err := daemonSetClient.Create(ctx, daemonSet, metav1.CreateOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create daemonSet with name '%s' in namespace '%s'", name, namespace)
	}

	return daemonSetResult, nil
}

func (manager *KubernetesManager) RemoveDaemonSet(ctx context.Context, name string, namespace string) error {

	daemonSetClient := manager.kubernetesClientSet.AppsV1().DaemonSets(namespace)
	if err := daemonSetClient.Delete(ctx, name, removeObjectDeleteOptions); err != nil {
		return stacktrace.Propagate(err, "Failed to delete daemonSet with name '%s' with delete options '%+v'", name, removeObjectDeleteOptions)
	}

	return nil
}

func (manager *KubernetesManager) GetDaemonSet(ctx context.Context, namespace string, name string) (*appsv1.DaemonSet, error) {
	daemonSetClient := manager.kubernetesClientSet.AppsV1().DaemonSets(namespace)

	daemonSet, err := daemonSetClient.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get daemonSet with name '%s' in namespace '%s'", name, namespace)
	}

	return daemonSet, nil
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

	serviceAccountResult, err := client.Create(ctx, serviceAccount, metav1.CreateOptions{})
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

	pods, err := client.List(ctx, opts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get service accounts with labels '%+v', instead a non-nil error was returned", serviceAccountsLabels)
	}

	return pods, nil
}

func (manager *KubernetesManager) RemoveServiceAccount(ctx context.Context, name string, namespace string) error {
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

	roleResult, err := client.Create(ctx, role, metav1.CreateOptions{})
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

	pods, err := client.List(ctx, opts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get roles with labels '%+v', instead a non-nil error was returned", rolesLabels)
	}

	return pods, nil
}

func (manager *KubernetesManager) RemoveRole(ctx context.Context, name string, namespace string) error {
	client := manager.kubernetesClientSet.RbacV1().Roles(namespace)

	if err := client.Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		return stacktrace.Propagate(err, "Failed to delete role with name '%s' in namespace '%v'", name, namespace)
	}

	return nil
}

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

	roleBindingResult, err := client.Create(ctx, roleBinding, metav1.CreateOptions{})
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

	pods, err := client.List(ctx, opts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get role bindings with labels '%+v', instead a non-nil error was returned", roleBindingsLabels)
	}

	return pods, nil
}

func (manager *KubernetesManager) RemoveRoleBindings(ctx context.Context, name string, namespace string) error {
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

	clusterRoleResult, err := client.Create(ctx, clusterRole, metav1.CreateOptions{})
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

	pods, err := client.List(ctx, opts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get cluster roles with labels '%+v', instead a non-nil error was returned", clusterRoleLabels)
	}

	return pods, nil
}

func (manager *KubernetesManager) RemoveClusterRole(ctx context.Context, name string) error {
	client := manager.kubernetesClientSet.RbacV1().ClusterRoles()

	if err := client.Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		return stacktrace.Propagate(err, "Failed to delete cluster role with name '%s'", name)
	}

	return nil
}

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

	clusterRoleBindingResult, err := client.Create(ctx, clusterRoleBinding, metav1.CreateOptions{})
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

	pods, err := client.List(ctx, opts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get cluster role bindings with labels '%+v', instead a non-nil error was returned", clusterRoleBindingsLabels)
	}

	return pods, nil
}

func (manager *KubernetesManager) RemoveClusterRoleBindings(ctx context.Context, name string) error {
	client := manager.kubernetesClientSet.RbacV1().ClusterRoleBindings()

	if err := client.Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		return stacktrace.Propagate(err, "Failed to delete cluster role binding with name '%s'", name)
	}

	return nil
}

// Private functions
func (manager *KubernetesManager) int32Ptr(i int32) *int32 { return &i }

// Pods
func (manager *KubernetesManager) CreatePod(
	ctx context.Context,
	namespace string,
	name string,
	podLabels map[string]string,
	podAnnotations map[string]string,
	podContainers []apiv1.Container,
	podVolumes []apiv1.Volume,
	podServiceAccountName string,
) (*apiv1.Pod, error) {
	podClient := manager.kubernetesClientSet.CoreV1().Pods(namespace)

	podMeta := metav1.ObjectMeta{
		Name:        name,
		Labels:      podLabels,
		Annotations: podAnnotations,
	}
	podSpec := apiv1.PodSpec{
		Volumes:    podVolumes,
		Containers: podContainers,
		ServiceAccountName: podServiceAccountName,
	}

	podToCreate := &apiv1.Pod{
		Spec:       podSpec,
		ObjectMeta: podMeta,
	}

	createdPod, err := podClient.Create(ctx, podToCreate, metav1.CreateOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to create pod with name '%v' and labels '%+v', instead a non-nil error was returned", name, podLabels)
	}

	return createdPod, nil
}

func (manager *KubernetesManager) RemovePod(ctx context.Context, namespace string, name string) error {
	podClient := manager.kubernetesClientSet.CoreV1().Pods(namespace)

	if err := podClient.Delete(ctx, name, removeObjectDeleteOptions); err != nil {
		return stacktrace.Propagate(err, "Failed to delete pod with name '%s' with delete options '%+v'", name, removeObjectDeleteOptions)
	}

	return nil
}

func (manager *KubernetesManager) GetPod(ctx context.Context, namespace string, name string) (*apiv1.Pod, error) {
	podClient := manager.kubernetesClientSet.CoreV1().Pods(namespace)

	pod, err := podClient.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get pod with name '%s'", name)
	}

	return pod, nil
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

	return pods, nil
}


func (manager *KubernetesManager) GetPodPortforwardEndpointUrl(namespace string, podName string) *url.URL {
	return manager.kubernetesClientSet.RESTClient().Post().Resource("pods").Namespace(namespace).Name(podName).SubResource("portforward").URL()
}
