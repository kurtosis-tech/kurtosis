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
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/concurrent_writer"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"io"
	v1 "k8s.io/api/apps/v1"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/channel_writer"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	terminal "golang.org/x/term"
	apiv1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	applyconfigurationsv1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

const (
	podWaitForAvailabilityTimeout          = 15 * time.Minute
	podWaitForAvailabilityTimeBetweenPolls = 500 * time.Millisecond
	podWaitForDeletionTimeout              = 5 * time.Minute
	podWaitForDeletionTimeBetweenPolls     = 500 * time.Millisecond
	podWaitForTerminationTimeout           = 5 * time.Minute
	podWaitForTerminationTimeBetweenPolls  = 500 * time.Millisecond

	// This is a container "reason" (machine-readable string) indicating that the container has some issue with
	// pulling the image (usually, a typo in the image name or the image doesn't exist)
	// Pods in this state don't really recover on their own
	imagePullBackOffContainerReason = "ImagePullBackOff"

	containerStatusLineBulletPoint = " - "

	// Kubernetes unfortunately doesn't have a good way to get the exit code out, so we have to parse it out of a string
	expectedTerminationMessage = "command terminated with exit code"

	shouldAllocateStdinOnPodExec   = false
	shouldAllocatedStdoutOnPodExec = true
	shouldAllocatedStderrOnPodExec = true
	shouldAllocateTtyOnPodExec     = false

	successExecCommandExitCode = 0

	// This is the owner string we'll use when updating fields
	fieldManager = "kurtosis"

	shouldFollowContainerLogsWhenPrintingPodInfo = false
	shouldAddTimestampsWhenPrintingPodInfo       = true

	listOptionsTimeoutSeconds      int64 = 10
	contextDeadlineExceeded              = "context deadline exceeded"
	expectedStatusMessageSliceSize       = 6

	shouldUseHostPidsNamespace     = true
	shouldUseHostNetworksNamespace = true
)

// We'll try to use the nicer-to-use shells first before we drop down to the lower shells
var commandToRunWhenCreatingUserServiceShell = []string{
	"sh",
	"-c",
	`if command -v 'bash' > /dev/null; then
		echo "Found bash on container; creating bash shell..."; bash; 
       else 
		echo "No bash found on container; dropping down to sh shell..."; sh; 
	fi`,
}

var (
	globalDeletePolicy  = metav1.DeletePropagationForeground
	globalDeleteOptions = metav1.DeleteOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		GracePeriodSeconds: nil,
		Preconditions:      nil,
		OrphanDependents:   nil,
		PropagationPolicy:  &globalDeletePolicy,
		DryRun:             nil,
	}
	globalCreateOptions = metav1.CreateOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		DryRun: nil,
		// We need every object to have this field manager so that the Kurtosis objects can all seamlessly modify Kubernetes resources
		FieldManager:    fieldManager,
		FieldValidation: "",
	}
	globalGetOptions = metav1.GetOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		ResourceVersion: "",
	}
	globalListOptions = metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		LabelSelector:        "",
		FieldSelector:        "",
		Watch:                false,
		AllowWatchBookmarks:  false,
		ResourceVersion:      "",
		ResourceVersionMatch: "",
		TimeoutSeconds:       nil,
		Limit:                0,
		Continue:             "",
		SendInitialEvents:    nil,
	}
)

type KubernetesManager struct {
	// The underlying K8s client that will be used to modify the K8s environment
	kubernetesClientSet *kubernetes.Clientset

	// Underlying restClient configuration
	kuberneteRestConfig *rest.Config
	// The storage class name as specified in the `kurtosis-config.yaml`
	storageClass string
}

func int64Ptr(i int64) *int64 { return &i }

func NewKubernetesManager(kubernetesClientSet *kubernetes.Clientset, kuberneteRestConfig *rest.Config, storageClass string) *KubernetesManager {
	return &KubernetesManager{
		kubernetesClientSet: kubernetesClientSet,
		kuberneteRestConfig: kuberneteRestConfig,
		storageClass:        storageClass,
	}
}

// ---------------------------Services------------------------------------------------------------------------------

// CreateService creates a k8s service in the specified namespace. It connects pods to the service according to the pod labels passed in
func (manager *KubernetesManager) CreateService(ctx context.Context, namespace string, name string, serviceLabels map[string]string, serviceAnnotations map[string]string, matchPodLabels map[string]string, serviceType apiv1.ServiceType, ports []apiv1.ServicePort) (*apiv1.Service, error) {
	servicesClient := manager.kubernetesClientSet.CoreV1().Services(namespace)

	objectMeta := metav1.ObjectMeta{
		Name:            name,
		GenerateName:    "",
		Namespace:       "",
		SelfLink:        "",
		UID:             "",
		ResourceVersion: "",
		Generation:      0,
		CreationTimestamp: metav1.Time{
			Time: time.Time{},
		},
		DeletionTimestamp:          nil,
		DeletionGracePeriodSeconds: nil,
		Labels:                     serviceLabels,
		Annotations:                serviceAnnotations,
		OwnerReferences:            nil,
		Finalizers:                 nil,
		ManagedFields:              nil,
	}

	// Figure out selector api

	// There must be a better way
	serviceSpec := apiv1.ServiceSpec{
		Ports:                         ports,
		Selector:                      matchPodLabels, // these labels are used to match with the Pod
		ClusterIP:                     "",
		ClusterIPs:                    nil,
		Type:                          serviceType,
		ExternalIPs:                   nil,
		SessionAffinity:               "",
		LoadBalancerIP:                "",
		LoadBalancerSourceRanges:      nil,
		ExternalName:                  "",
		ExternalTrafficPolicy:         "",
		HealthCheckNodePort:           0,
		PublishNotReadyAddresses:      false,
		SessionAffinityConfig:         nil,
		IPFamilies:                    nil,
		IPFamilyPolicy:                nil,
		AllocateLoadBalancerNodePorts: nil,
		LoadBalancerClass:             nil,
		InternalTrafficPolicy:         nil,
	}

	service := &apiv1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		ObjectMeta: objectMeta,
		Spec:       serviceSpec,
		Status: apiv1.ServiceStatus{
			LoadBalancer: apiv1.LoadBalancerStatus{
				Ingress: nil,
			},
			Conditions: nil,
		},
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
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		DryRun:       nil,
		Force:        false,
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

	opts := buildListOptionsFromLabels(serviceLabels)
	serviceResult, err := servicesClient.List(ctx, opts)
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

func (manager *KubernetesManager) GetIngressesByLabels(ctx context.Context, namespace string, ingressLabels map[string]string) (*netv1.IngressList, error) {
	ingressesClient := manager.kubernetesClientSet.NetworkingV1().Ingresses(namespace)

	opts := buildListOptionsFromLabels(ingressLabels)
	ingressResult, err := ingressesClient.List(ctx, opts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to list ingresses with labels '%+v' in namespace '%s'", ingressLabels, namespace)
	}

	// Only return objects not tombstoned by Kubernetes
	var ingressesNotMarkedForDeletionList []netv1.Ingress
	for _, service := range ingressResult.Items {
		deletionTimestamp := service.GetObjectMeta().GetDeletionTimestamp()
		if deletionTimestamp == nil {
			ingressesNotMarkedForDeletionList = append(ingressesNotMarkedForDeletionList, service)
		}
	}
	ingressesNotMarkedForDeletionserviceList := netv1.IngressList{
		Items:    ingressesNotMarkedForDeletionList,
		TypeMeta: ingressResult.TypeMeta,
		ListMeta: ingressResult.ListMeta,
	}

	return &ingressesNotMarkedForDeletionserviceList, nil
}

// ---------------------------Volumes------------------------------------------------------------------------------

func (manager *KubernetesManager) GetPersistentVolumesByLabels(ctx context.Context, persistentVolumeLabels map[string]string) (*apiv1.PersistentVolumeList, error) {
	persistentVolumesClient := manager.kubernetesClientSet.CoreV1().PersistentVolumes()

	listOptions := buildListOptionsFromLabels(persistentVolumeLabels)
	persistentVolumesResult, err := persistentVolumesClient.List(ctx, listOptions)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to list persistent volumes with labels '%+v'", persistentVolumeLabels)
	}

	// Only return objects not tombstoned by Kubernetes
	var persistentVolumesNotMarkedForDeletionList []apiv1.PersistentVolume
	for _, persistentVolume := range persistentVolumesResult.Items {
		deletionTimestamp := persistentVolume.GetObjectMeta().GetDeletionTimestamp()
		if deletionTimestamp == nil {
			persistentVolumesNotMarkedForDeletionList = append(persistentVolumesNotMarkedForDeletionList, persistentVolume)
		}
	}
	persistentVolumesNotMarkedForDeletionserviceList := apiv1.PersistentVolumeList{
		Items:    persistentVolumesNotMarkedForDeletionList,
		TypeMeta: persistentVolumesResult.TypeMeta,
		ListMeta: persistentVolumesResult.ListMeta,
	}

	return &persistentVolumesNotMarkedForDeletionserviceList, nil
}

func (manager *KubernetesManager) CreatePersistentVolumeClaim(
	ctx context.Context,
	namespace string,
	volumeClaimName string,
	labels map[string]string,
	requiredSize int64,
) (*apiv1.PersistentVolumeClaim, error) {
	if requiredSize == 0 {
		return nil, stacktrace.NewError("Cannot create volume '%v' of 0 size; need a value greater than 0", volumeClaimName)
	}

	volumeClaimsClient := manager.kubernetesClientSet.CoreV1().PersistentVolumeClaims(namespace)

	volumeClaimsDefinition := apiv1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            volumeClaimName,
			GenerateName:    "",
			Namespace:       namespace,
			SelfLink:        "",
			UID:             "",
			ResourceVersion: "",
			Generation:      0,
			CreationTimestamp: metav1.Time{
				Time: time.Time{},
			},
			DeletionTimestamp:          nil,
			DeletionGracePeriodSeconds: nil,
			Labels:                     labels,
			Annotations:                nil,
			OwnerReferences:            nil,
			Finalizers:                 nil,
			ManagedFields:              nil,
		},
		Spec: apiv1.PersistentVolumeClaimSpec{
			AccessModes: []apiv1.PersistentVolumeAccessMode{
				apiv1.ReadWriteOnce, // ReadWriteOncePod would be better, but it's a fairly recent feature
			},
			Selector: nil,
			Resources: apiv1.ResourceRequirements{
				Limits: nil,
				Requests: apiv1.ResourceList{
					// we give each claim 100% of the corresponding volume. Since we have a 1:1 mapping between volumes
					// and claims right now, it's the best we can do
					apiv1.ResourceStorage: *resource.NewQuantity(requiredSize, resource.BinarySI),
				},
				Claims: nil,
			},
			VolumeName:       "", // we use dynamic provisioning this should happen automagically
			StorageClassName: &manager.storageClass,
			VolumeMode:       nil,
			DataSource:       nil,
			DataSourceRef:    nil,
		},
		Status: apiv1.PersistentVolumeClaimStatus{
			Phase:              "",
			AccessModes:        nil,
			Capacity:           nil,
			Conditions:         nil,
			AllocatedResources: nil,
			ResizeStatus:       nil,
		},
	}

	volumeClaim, err := volumeClaimsClient.Create(ctx, &volumeClaimsDefinition, globalCreateOptions)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create volume claim '%s'", volumeClaimName)
	}

	return volumeClaim, err
}

func (manager *KubernetesManager) RemovePersistentVolumeClaim(
	ctx context.Context,
	namespace string,
	volumeClaimName string,
) error {
	volumesClient := manager.kubernetesClientSet.CoreV1().PersistentVolumeClaims(namespace)
	if err := volumesClient.Delete(ctx, volumeClaimName, globalDeleteOptions); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing the persistent volume claim '%s' in namespace '%s'",
			volumeClaimName, namespace)
	}
	return nil
}

func (manager *KubernetesManager) GetPersistentVolumeClaim(
	ctx context.Context,
	namespace string,
	volumeClaimName string,
) (*apiv1.PersistentVolumeClaim, error) {
	volumesClient := manager.kubernetesClientSet.CoreV1().PersistentVolumeClaims(namespace)
	volumeClaim, err := volumesClient.Get(ctx, volumeClaimName, globalGetOptions)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the persistent volume claim '%s' in namespace '%s'", volumeClaimName, namespace)
	}
	deletionTimestamp := volumeClaim.GetObjectMeta().GetDeletionTimestamp()
	if deletionTimestamp != nil {
		return nil, stacktrace.Propagate(err, "Persistent volume claim with name '%s' in namespace '%s' has been marked for deletion",
			volumeClaim.Name, namespace)
	}
	return volumeClaim, nil
}

// ---------------------------namespaces------------------------------------------------------------------------------

func (manager *KubernetesManager) CreateNamespace(
	ctx context.Context,
	name string,
	namespaceLabels map[string]string,
	namespaceAnnotations map[string]string,
) (*apiv1.Namespace, error) {
	namespaceClient := manager.kubernetesClientSet.CoreV1().Namespaces()

	namespace := &apiv1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			GenerateName:    "",
			Namespace:       "",
			SelfLink:        "",
			UID:             "",
			ResourceVersion: "",
			Generation:      0,
			CreationTimestamp: metav1.Time{
				Time: time.Time{},
			},
			DeletionTimestamp:          nil,
			DeletionGracePeriodSeconds: nil,
			Labels:                     namespaceLabels,
			Annotations:                namespaceAnnotations,
			OwnerReferences:            nil,
			Finalizers:                 nil,
			ManagedFields:              nil,
		},
		Spec: apiv1.NamespaceSpec{
			Finalizers: nil,
		},
		Status: apiv1.NamespaceStatus{
			Phase:      "",
			Conditions: nil,
		},
	}

	namespaceResult, err := namespaceClient.Create(ctx, namespace, globalCreateOptions)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create namespace with name '%s'", name)
	}

	return namespaceResult, nil
}

func (manager *KubernetesManager) UpdateNamespace(
	ctx context.Context,
	namespaceName string,
	// We use a configurator, rather than letting the user pass in their own NamespaceApplyConfiguration, so that we ensure
	// they use the constructor (and don't do struct instantiation and forget to add the object name, etc. which
	// would result in removing the object name)
	updateConfigurator func(configuration *applyconfigurationsv1.NamespaceApplyConfiguration),
) (*apiv1.Namespace, error) {
	updatesToApply := applyconfigurationsv1.Namespace(namespaceName)
	updateConfigurator(updatesToApply)

	namespaceClient := manager.kubernetesClientSet.CoreV1().Namespaces()

	applyOpts := metav1.ApplyOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		DryRun:       nil,
		Force:        true, //We need to use force to avoid conflict errors
		FieldManager: fieldManager,
	}
	result, err := namespaceClient.Apply(ctx, updatesToApply, applyOpts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to update namespace '%v' ", namespaceName)
	}
	return result, nil
}

func (manager *KubernetesManager) RemoveNamespace(ctx context.Context, namespace *apiv1.Namespace) error {
	name := namespace.Name
	namespaceClient := manager.kubernetesClientSet.CoreV1().Namespaces()

	if err := namespaceClient.Delete(ctx, name, globalDeleteOptions); err != nil {
		return stacktrace.Propagate(err, "Failed to delete namespace with name '%s' with delete options '%+v'", name, globalDeleteOptions)
	}

	return nil
}

// GetNamespace returns the namespace object associated with [name] or returns err
// - if err occurred getting namespace,
// - the namespace doesn't exist
// - the namespace has been marked for deletions
func (manager *KubernetesManager) GetNamespace(ctx context.Context, name string) (*apiv1.Namespace, error) {
	namespaceClient := manager.kubernetesClientSet.CoreV1().Namespaces()

	namespace, err := namespaceClient.Get(ctx, name, metav1.GetOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		ResourceVersion: "",
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get namespace with name '%s'", name)
	}
	deletionTimestamp := namespace.GetObjectMeta().GetDeletionTimestamp()
	if deletionTimestamp != nil {
		return nil, stacktrace.NewError("Namespace with name '%s' has been marked for deletion", namespace)
	}
	return namespace, nil
}

func (manager *KubernetesManager) GetNamespacesByLabels(ctx context.Context, namespaceLabels map[string]string) (*apiv1.NamespaceList, error) {
	namespaceClient := manager.kubernetesClientSet.CoreV1().Namespaces()

	listOptions := buildListOptionsFromLabels(namespaceLabels)
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

func (manager *KubernetesManager) CreateServiceAccount(ctx context.Context, name string, namespace string, labels map[string]string, imagePullSecrets []apiv1.LocalObjectReference) (*apiv1.ServiceAccount, error) {
	client := manager.kubernetesClientSet.CoreV1().ServiceAccounts(namespace)

	serviceAccount := &apiv1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			GenerateName:    "",
			Namespace:       "",
			SelfLink:        "",
			UID:             "",
			ResourceVersion: "",
			Generation:      0,
			CreationTimestamp: metav1.Time{
				Time: time.Time{},
			},
			DeletionTimestamp:          nil,
			DeletionGracePeriodSeconds: nil,
			Labels:                     labels,
			Annotations:                nil,
			OwnerReferences:            nil,
			Finalizers:                 nil,
			ManagedFields:              nil,
		},
		Secrets:                      nil,
		ImagePullSecrets:             imagePullSecrets,
		AutomountServiceAccountToken: nil,
	}

	serviceAccountResult, err := client.Create(ctx, serviceAccount, globalCreateOptions)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create service account with name '%s' in namespace '%v'", name, namespace)
	}
	return serviceAccountResult, nil
}

func (manager *KubernetesManager) GetServiceAccountsByLabels(ctx context.Context, namespace string, serviceAccountsLabels map[string]string) (*apiv1.ServiceAccountList, error) {
	client := manager.kubernetesClientSet.CoreV1().ServiceAccounts(namespace)

	opts := buildListOptionsFromLabels(serviceAccountsLabels)
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

	deleteOptions := metav1.DeleteOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		GracePeriodSeconds: nil,
		Preconditions:      nil,
		OrphanDependents:   nil,
		PropagationPolicy:  nil,
		DryRun:             nil,
	}
	if err := client.Delete(ctx, name, deleteOptions); err != nil {
		return stacktrace.Propagate(err, "Failed to delete service account with name '%s' in namespace '%v'", name, namespace)
	}

	return nil
}

// ---------------------------roles------------------------------------------------------------------------------

func (manager *KubernetesManager) CreateRole(ctx context.Context, name string, namespace string, rules []rbacv1.PolicyRule, labels map[string]string) (*rbacv1.Role, error) {
	client := manager.kubernetesClientSet.RbacV1().Roles(namespace)

	role := &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			GenerateName:    "",
			Namespace:       "",
			SelfLink:        "",
			UID:             "",
			ResourceVersion: "",
			Generation:      0,
			CreationTimestamp: metav1.Time{
				Time: time.Time{},
			},
			DeletionTimestamp:          nil,
			DeletionGracePeriodSeconds: nil,
			Labels:                     labels,
			Annotations:                nil,
			OwnerReferences:            nil,
			Finalizers:                 nil,
			ManagedFields:              nil,
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

	opts := buildListOptionsFromLabels(rolesLabels)
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

	deleteOptions := metav1.DeleteOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		GracePeriodSeconds: nil,
		Preconditions:      nil,
		OrphanDependents:   nil,
		PropagationPolicy:  nil,
		DryRun:             nil,
	}
	if err := client.Delete(ctx, name, deleteOptions); err != nil {
		return stacktrace.Propagate(err, "Failed to delete role with name '%s' in namespace '%v'", name, namespace)
	}

	return nil
}

// --------------------------- Role Bindings ------------------------------------------------------------------------------

func (manager *KubernetesManager) CreateRoleBindings(ctx context.Context, name string, namespace string, subjects []rbacv1.Subject, roleRef rbacv1.RoleRef, labels map[string]string) (*rbacv1.RoleBinding, error) {
	client := manager.kubernetesClientSet.RbacV1().RoleBindings(namespace)

	roleBinding := &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			GenerateName:    "",
			Namespace:       "",
			SelfLink:        "",
			UID:             "",
			ResourceVersion: "",
			Generation:      0,
			CreationTimestamp: metav1.Time{
				Time: time.Time{},
			},
			DeletionTimestamp:          nil,
			DeletionGracePeriodSeconds: nil,
			Labels:                     labels,
			Annotations:                nil,
			OwnerReferences:            nil,
			Finalizers:                 nil,
			ManagedFields:              nil,
		},
		Subjects: subjects,
		RoleRef:  roleRef,
	}

	roleBindingResult, err := client.Create(ctx, roleBinding, globalCreateOptions)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create role binding with name '%s', subjects '%+v' and role ref '%v'", name, subjects, roleRef)
	}

	return roleBindingResult, nil
}

func (manager *KubernetesManager) GetRoleBindingsByLabels(ctx context.Context, namespace string, roleBindingsLabels map[string]string) (*rbacv1.RoleBindingList, error) {
	client := manager.kubernetesClientSet.RbacV1().RoleBindings(namespace)

	opts := buildListOptionsFromLabels(roleBindingsLabels)
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

	deleteOptions := metav1.DeleteOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		GracePeriodSeconds: nil,
		Preconditions:      nil,
		OrphanDependents:   nil,
		PropagationPolicy:  nil,
		DryRun:             nil,
	}
	if err := client.Delete(ctx, name, deleteOptions); err != nil {
		return stacktrace.Propagate(err, "Failed to delete role bindings with name '%s' in namespace '%v'", name, namespace)
	}

	return nil
}

// ---------------------------cluster roles------------------------------------------------------------------------------

func (manager *KubernetesManager) CreateClusterRoles(ctx context.Context, name string, rules []rbacv1.PolicyRule, labels map[string]string) (*rbacv1.ClusterRole, error) {
	client := manager.kubernetesClientSet.RbacV1().ClusterRoles()

	clusterRole := &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			GenerateName:    "",
			Namespace:       "",
			SelfLink:        "",
			UID:             "",
			ResourceVersion: "",
			Generation:      0,
			CreationTimestamp: metav1.Time{
				Time: time.Time{},
			},
			DeletionTimestamp:          nil,
			DeletionGracePeriodSeconds: nil,
			Labels:                     labels,
			Annotations:                nil,
			OwnerReferences:            nil,
			Finalizers:                 nil,
			ManagedFields:              nil,
		},
		Rules:           rules,
		AggregationRule: nil,
	}

	clusterRoleResult, err := client.Create(ctx, clusterRole, globalCreateOptions)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create cluster role with name '%s' with rules '%+v'", name, rules)
	}

	return clusterRoleResult, nil
}

func (manager *KubernetesManager) GetClusterRolesByLabels(ctx context.Context, clusterRoleLabels map[string]string) (*rbacv1.ClusterRoleList, error) {
	client := manager.kubernetesClientSet.RbacV1().ClusterRoles()

	opts := buildListOptionsFromLabels(clusterRoleLabels)
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

	deleteOptions := metav1.DeleteOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		GracePeriodSeconds: nil,
		Preconditions:      nil,
		OrphanDependents:   nil,
		PropagationPolicy:  nil,
		DryRun:             nil,
	}
	if err := client.Delete(ctx, name, deleteOptions); err != nil {
		return stacktrace.Propagate(err, "Failed to delete cluster role with name '%s'", name)
	}

	return nil
}

// --------------------------- Cluster Role Bindings ------------------------------------------------------------------------------

func (manager *KubernetesManager) CreateClusterRoleBindings(ctx context.Context, name string, subjects []rbacv1.Subject, roleRef rbacv1.RoleRef, labels map[string]string) (*rbacv1.ClusterRoleBinding, error) {
	client := manager.kubernetesClientSet.RbacV1().ClusterRoleBindings()

	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			GenerateName:    "",
			Namespace:       "",
			SelfLink:        "",
			UID:             "",
			ResourceVersion: "",
			Generation:      0,
			CreationTimestamp: metav1.Time{
				Time: time.Time{},
			},
			DeletionTimestamp:          nil,
			DeletionGracePeriodSeconds: nil,
			Labels:                     labels,
			Annotations:                nil,
			OwnerReferences:            nil,
			Finalizers:                 nil,
			ManagedFields:              nil,
		},
		Subjects: subjects,
		RoleRef:  roleRef,
	}

	clusterRoleBindingResult, err := client.Create(ctx, clusterRoleBinding, globalCreateOptions)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create cluster role binding with name '%s', subjects '%+v' and role ref '%v'", name, subjects, roleRef)
	}

	return clusterRoleBindingResult, nil
}

func (manager *KubernetesManager) GetClusterRoleBindingsByLabels(ctx context.Context, clusterRoleBindingsLabels map[string]string) (*rbacv1.ClusterRoleBindingList, error) {
	client := manager.kubernetesClientSet.RbacV1().ClusterRoleBindings()

	opts := buildListOptionsFromLabels(clusterRoleBindingsLabels)
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

	deleteOptions := metav1.DeleteOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		GracePeriodSeconds: nil,
		Preconditions:      nil,
		OrphanDependents:   nil,
		PropagationPolicy:  nil,
		DryRun:             nil,
	}
	if err := client.Delete(ctx, name, deleteOptions); err != nil {
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
	restartPolicy apiv1.RestartPolicy,
	tolerations []apiv1.Toleration,
	nodeSelectors map[string]string,
	shouldUseHostPidsNamespace bool,
	shouldUseHostNetworksNamespace bool,
) (
	*apiv1.Pod,
	error,
) {
	podClient := manager.kubernetesClientSet.CoreV1().Pods(namespaceName)

	podMeta := metav1.ObjectMeta{
		Name:            podName,
		GenerateName:    "",
		Namespace:       "",
		SelfLink:        "",
		UID:             "",
		ResourceVersion: "",
		Generation:      0,
		CreationTimestamp: metav1.Time{
			Time: time.Time{},
		},
		DeletionTimestamp:          nil,
		DeletionGracePeriodSeconds: nil,
		Labels:                     podLabels,
		Annotations:                podAnnotations,
		OwnerReferences:            nil,
		Finalizers:                 nil,
		ManagedFields:              nil,
	}
	podSpec := apiv1.PodSpec{
		Volumes:                       podVolumes,
		InitContainers:                initContainers,
		Containers:                    podContainers,
		EphemeralContainers:           nil,
		RestartPolicy:                 restartPolicy,
		TerminationGracePeriodSeconds: nil,
		ActiveDeadlineSeconds:         nil,
		DNSPolicy:                     "",
		NodeSelector:                  nodeSelectors,
		ServiceAccountName:            podServiceAccountName,
		DeprecatedServiceAccount:      "",
		AutomountServiceAccountToken:  nil,
		NodeName:                      "",
		HostNetwork:                   shouldUseHostNetworksNamespace,
		HostPID:                       shouldUseHostPidsNamespace,
		HostIPC:                       false,
		ShareProcessNamespace:         nil,
		SecurityContext:               nil,
		// TODO add support for ImageRegistrySpec to Kubernetes by adding the right secret here
		// You will have to first publish the secret using the Kubernetes API
		ImagePullSecrets:          nil,
		Hostname:                  "",
		Subdomain:                 "",
		Affinity:                  nil,
		SchedulerName:             "",
		Tolerations:               tolerations,
		HostAliases:               nil,
		PriorityClassName:         "",
		Priority:                  nil,
		DNSConfig:                 nil,
		ReadinessGates:            nil,
		RuntimeClassName:          nil,
		EnableServiceLinks:        nil,
		PreemptionPolicy:          nil,
		Overhead:                  nil,
		TopologySpreadConstraints: nil,
		SetHostnameAsFQDN:         nil,
		OS:                        nil,
		HostUsers:                 nil,
		SchedulingGates:           nil,
		ResourceClaims:            nil,
	}

	podToCreate := &apiv1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		ObjectMeta: podMeta,
		Spec:       podSpec,
		Status: apiv1.PodStatus{
			Phase:                      "",
			Conditions:                 nil,
			Message:                    "",
			Reason:                     "",
			NominatedNodeName:          "",
			HostIP:                     "",
			PodIP:                      "",
			PodIPs:                     nil,
			StartTime:                  nil,
			InitContainerStatuses:      nil,
			ContainerStatuses:          nil,
			QOSClass:                   "",
			EphemeralContainerStatuses: nil,
			Resize:                     "",
		},
	}

	if podDefinitionBytes, err := json.Marshal(podToCreate); err == nil {
		logrus.Debugf("Going to start pod using the following JSON: %v", string(podDefinitionBytes))
	}

	// In case of a service update, it's possible the pod has not been fully deleted yet. It's still "scheduled for
	// deletion", and so we wait for it to be deleted before creating it again. There's probably a way to optimize this
	// a bit more using native k8s pod update operation
	if err := manager.waitForPodDeletion(ctx, namespaceName, podName); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for pod '%v' to be completely removed", podName)
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

	if err := manager.WaitForPodTermination(ctx, namespace, name); err != nil {
		return stacktrace.Propagate(err, "An error occurred waiting for pod '%v' to terminate", name)
	}

	return nil
}

func (manager *KubernetesManager) GetPod(ctx context.Context, namespace string, name string) (*apiv1.Pod, error) {
	podClient := manager.kubernetesClientSet.CoreV1().Pods(namespace)

	pod, err := podClient.Get(ctx, name, metav1.GetOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		ResourceVersion: "",
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get pod with name '%s'", name)
	}

	return pod, nil
}

// ---------------------------daemon sets---------------------------------------------------------------------------------------
func (manager *KubernetesManager) RemoveDaemonSet(ctx context.Context, namespace string, daemonSet *v1.DaemonSet) error {
	client := manager.kubernetesClientSet.AppsV1().DaemonSets(namespace)

	if err := client.Delete(ctx, daemonSet.Name, globalDeleteOptions); err != nil {
		return stacktrace.Propagate(err, "Failed to delete daemon set with name '%s' with delete options '%+v'", daemonSet.Name, globalDeleteOptions)
	}

	// TODO: maybe add a termination wait here?
	return nil
}

func (manager *KubernetesManager) GetDaemonSet(ctx context.Context, namespace string, name string) (*v1.DaemonSet, error) {
	daemonSetClient := manager.kubernetesClientSet.AppsV1().DaemonSets(namespace)

	daemonSet, err := daemonSetClient.Get(ctx, name, metav1.GetOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		ResourceVersion: "",
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get daemon set with name '%s'", name)
	}

	return daemonSet, nil
}

func (manager *KubernetesManager) CreateDaemonSet(
	ctx context.Context,
	namespaceName string,
	daemonSetName string,
	daemonSetLabels map[string]string,
	daemonSetAnnotations map[string]string,
	daemonSetServiceAccountName string,
	initContainers []apiv1.Container,
	containers []apiv1.Container,
	volumes []apiv1.Volume,
) (*v1.DaemonSet, error) {
	daemonSetClient := manager.kubernetesClientSet.AppsV1().DaemonSets(namespaceName)

	daemonSetMeta := metav1.ObjectMeta{
		Name:            daemonSetName,
		GenerateName:    "",
		Namespace:       namespaceName,
		SelfLink:        "",
		UID:             "",
		ResourceVersion: "",
		Generation:      0,
		CreationTimestamp: metav1.Time{
			Time: time.Time{},
		},
		DeletionTimestamp:          nil,
		DeletionGracePeriodSeconds: nil,
		Labels:                     daemonSetLabels,
		Annotations:                daemonSetAnnotations,
		OwnerReferences:            nil,
		Finalizers:                 nil,
		ManagedFields:              nil,
	}

	daemonSetSpec := v1.DaemonSetSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels:      daemonSetLabels,
			MatchExpressions: nil,
		},
		Template: apiv1.PodTemplateSpec{
			ObjectMeta: daemonSetMeta,
			Spec: apiv1.PodSpec{
				Volumes:                       volumes,
				InitContainers:                initContainers,
				Containers:                    containers,
				EphemeralContainers:           nil,
				RestartPolicy:                 "",
				TerminationGracePeriodSeconds: nil,
				ActiveDeadlineSeconds:         nil,
				DNSPolicy:                     "",
				NodeSelector:                  nil,
				ServiceAccountName:            daemonSetServiceAccountName,
				DeprecatedServiceAccount:      "",
				AutomountServiceAccountToken:  nil,
				NodeName:                      "",
				HostNetwork:                   false,
				HostPID:                       false,
				HostIPC:                       false,
				ShareProcessNamespace:         nil,
				SecurityContext:               nil,
				ImagePullSecrets:              nil,
				Hostname:                      "",
				Subdomain:                     "",
				Affinity:                      nil,
				SchedulerName:                 "",
				Tolerations:                   nil,
				HostAliases:                   nil,
				PriorityClassName:             "",
				Priority:                      nil,
				DNSConfig:                     nil,
				ReadinessGates:                nil,
				RuntimeClassName:              nil,
				EnableServiceLinks:            nil,
				PreemptionPolicy:              nil,
				Overhead:                      nil,
				TopologySpreadConstraints:     nil,
				SetHostnameAsFQDN:             nil,
				OS:                            nil,
				HostUsers:                     nil,
				SchedulingGates:               nil,
				ResourceClaims:                nil,
			},
		},
		UpdateStrategy: v1.DaemonSetUpdateStrategy{
			Type:          "",
			RollingUpdate: nil,
		},
		MinReadySeconds:      0,
		RevisionHistoryLimit: nil,
	}

	daemonSetToCreate := &v1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		ObjectMeta: daemonSetMeta,
		Spec:       daemonSetSpec,
		Status: v1.DaemonSetStatus{
			CurrentNumberScheduled: 0,
			NumberMisscheduled:     0,
			DesiredNumberScheduled: 0,
			NumberReady:            0,
			ObservedGeneration:     0,
			UpdatedNumberScheduled: 0,
			NumberAvailable:        0,
			NumberUnavailable:      0,
			CollisionCount:         nil,
			Conditions:             nil,
		},
	}

	if daemonSetDefinitionBytes, err := json.Marshal(daemonSetToCreate); err == nil {
		logrus.Debugf("Going to start daemon set using the following JSON: %v", string(daemonSetDefinitionBytes))
	}

	createdDaemonSet, err := daemonSetClient.Create(ctx, daemonSetToCreate, globalCreateOptions)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating daemon set.")
	}

	return createdDaemonSet, nil
}

func (manager *KubernetesManager) GetPodsManagedByDaemonSet(ctx context.Context, daemonSet *v1.DaemonSet) ([]*apiv1.Pod, error) {
	podsClient := manager.kubernetesClientSet.CoreV1().Pods(daemonSet.Namespace)

	selector := metav1.FormatLabelSelector(daemonSet.Spec.Selector)

	pods, err := podsClient.List(ctx, metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		LabelSelector:        selector,
		FieldSelector:        "",
		Watch:                false,
		AllowWatchBookmarks:  false,
		ResourceVersion:      "",
		ResourceVersionMatch: "",
		TimeoutSeconds:       nil,
		Limit:                0,
		Continue:             "",
		SendInitialEvents:    nil,
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred retrieving list of pods in namespace '%v' with label selectors: %v.", daemonSet.Namespace, selector)
	}

	var podsManagedByDaemonSet []*apiv1.Pod
	for _, pod := range pods.Items {
		podToAdd := pod
		podsManagedByDaemonSet = append(podsManagedByDaemonSet, &podToAdd)
	}

	return podsManagedByDaemonSet, nil
}

func (manager *KubernetesManager) UpdateDaemonSetWithNodeSelectors(ctx context.Context, daemonSet *v1.DaemonSet, nodeSelector map[string]string) (*v1.DaemonSet, error) {
	daemonSet.Spec.Template.Spec.NodeSelector = nodeSelector

	daemonSetName := daemonSet.Name
	daemonSetNamespace := daemonSet.Namespace

	daemonSet, err := manager.kubernetesClientSet.AppsV1().DaemonSets(daemonSet.Namespace).Update(
		ctx,
		daemonSet,
		metav1.UpdateOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "",
				APIVersion: "",
			},
			DryRun:          nil,
			FieldManager:    "",
			FieldValidation: "",
		},
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred updating daemon set '%v' in namespace '%v'.", daemonSetName, daemonSetNamespace)
	}

	logrus.Debugf("Successfully updated daemon set with node selector %v", nodeSelector)

	updatedDaemonSet, err := manager.kubernetesClientSet.AppsV1().DaemonSets(daemonSet.Namespace).Get(
		ctx,
		daemonSet.Name,
		globalGetOptions,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the updated daemon set '%v' in namespace '%v'.", daemonSet.Name, daemonSet.Namespace)
	}

	return updatedDaemonSet, nil
}

//
//func (manager *KubernetesManager) UpdateDaemonSetWithNodeSelectors(ctx context.Context, daemonSet *v1.DaemonSet, nodeSelector map[string]string) error {
//	patchData, err := json.Marshal([]map[string]interface{}{
//		{"op": "replace", "path": "/spec/template/spec/nodeSelector", "value": nodeSelector},
//	})
//	if err != nil {
//		return stacktrace.Propagate(err, "An error occurred marshaling data to patch daemon set '%v' with node selectors '%v':\n%v.", daemonSet.Name, nodeSelector, patchData)
//	}
//
//	_, err = manager.kubernetesClientSet.AppsV1().DaemonSets(daemonSet.Namespace).Patch(
//		ctx,
//		daemonSet.Name,
//		types.JSONPatchType,
//		patchData,
//		metav1.PatchOptions{
//			TypeMeta: metav1.TypeMeta{
//				Kind:       "",
//				APIVersion: "",
//			},
//			DryRun:          nil,
//			Force:           nil,
//			FieldManager:    "",
//			FieldValidation: "",
//		},
//	)
//	if err != nil {
//		return stacktrace.Propagate(err, "An error occurred patching daemon set '%v' in namespace '%v' with patch data '%v'.", daemonSet.Name, daemonSet.Namespace, patchData)
//	}
//	logrus.Debugf("Successfully patched daemon set with node selector %v and patch data '%v'", nodeSelector, patchData)
//
//	return nil
//}

// ---------------------------deployments---------------------------------------------------------------------------------------
func (manager *KubernetesManager) RemoveDeployment(ctx context.Context, namespace string, deployment *v1.Deployment) error {
	client := manager.kubernetesClientSet.AppsV1().Deployments(namespace)

	if err := client.Delete(ctx, deployment.Name, globalDeleteOptions); err != nil {
		return stacktrace.Propagate(err, "Failed to delete deployment with name '%s' with delete options '%+v'", deployment.Name, globalDeleteOptions)
	}

	// TODO: maybe add a termination wait here?
	return nil
}

func (manager *KubernetesManager) GetDeployment(ctx context.Context, namespace string, name string) (*v1.Deployment, error) {
	daemonSetClient := manager.kubernetesClientSet.AppsV1().Deployments(namespace)

	daemonSet, err := daemonSetClient.Get(ctx, name, metav1.GetOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		ResourceVersion: "",
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get daemon set with name '%s'", name)
	}

	return daemonSet, nil
}

func (manager *KubernetesManager) CreateDeployment(
	ctx context.Context,
	namespaceName string,
	deploymentName string,
	deploymentLabels map[string]string,
	deploymentAnnotations map[string]string,
	initContainers []apiv1.Container,
	containers []apiv1.Container,
	volumes []apiv1.Volume,
	affinity *apiv1.Affinity,
) (*v1.Deployment, error) {
	deploymentClient := manager.kubernetesClientSet.AppsV1().Deployments(namespaceName)

	deploymentMeta := metav1.ObjectMeta{
		Name:            deploymentName,
		GenerateName:    "",
		Namespace:       namespaceName,
		SelfLink:        "",
		UID:             "",
		ResourceVersion: "",
		Generation:      0,
		CreationTimestamp: metav1.Time{
			Time: time.Time{},
		},
		DeletionTimestamp:          nil,
		DeletionGracePeriodSeconds: nil,
		Labels:                     deploymentLabels,
		Annotations:                deploymentAnnotations,
		OwnerReferences:            nil,
		Finalizers:                 nil,
		ManagedFields:              nil,
	}

	numReplicas := int32(1)
	deploymentSpec := v1.DeploymentSpec{
		Replicas: &numReplicas,
		Strategy: v1.DeploymentStrategy{
			Type:          "",
			RollingUpdate: nil,
		},
		Paused:                  false,
		ProgressDeadlineSeconds: nil,
		Selector: &metav1.LabelSelector{
			MatchLabels:      deploymentLabels,
			MatchExpressions: nil,
		},
		Template: apiv1.PodTemplateSpec{
			ObjectMeta: deploymentMeta,
			Spec: apiv1.PodSpec{
				ShareProcessNamespace:         nil,
				Volumes:                       volumes,
				InitContainers:                initContainers,
				Containers:                    containers,
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
				SecurityContext:               nil,
				ImagePullSecrets:              nil,
				Hostname:                      "",
				Subdomain:                     "",
				Affinity:                      affinity,
				SchedulerName:                 "",
				Tolerations:                   nil,
				HostAliases:                   nil,
				PriorityClassName:             "",
				Priority:                      nil,
				DNSConfig:                     nil,
				ReadinessGates:                nil,
				RuntimeClassName:              nil,
				EnableServiceLinks:            nil,
				PreemptionPolicy:              nil,
				Overhead:                      nil,
				TopologySpreadConstraints:     nil,
				SetHostnameAsFQDN:             nil,
				OS:                            nil,
				HostUsers:                     nil,
				SchedulingGates:               nil,
				ResourceClaims:                nil,
			},
		},
		MinReadySeconds:      0,
		RevisionHistoryLimit: nil,
	}

	deploymentToCreate := &v1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		ObjectMeta: deploymentMeta,
		Spec:       deploymentSpec,
		Status: v1.DeploymentStatus{
			ObservedGeneration:  0,
			CollisionCount:      nil,
			Conditions:          nil,
			ReadyReplicas:       0,
			Replicas:            0,
			UpdatedReplicas:     0,
			AvailableReplicas:   0,
			UnavailableReplicas: 0,
		},
	}

	if deploymentDefinitionBytes, err := json.Marshal(deploymentToCreate); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred marshaling deployment object for '%v' into json.", deploymentName)
	} else {
		logrus.Debugf("Going to start deployment using the following JSON: %v", string(deploymentDefinitionBytes))
	}

	createdDeployment, err := deploymentClient.Create(ctx, deploymentToCreate, globalCreateOptions)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating deployment.")
	}

	return createdDeployment, nil
}

func (manager *KubernetesManager) ScaleDeployment(ctx context.Context, namespace, name string, replicas int32) error {
	deploymentClient := manager.kubernetesClientSet.AppsV1().Deployments(namespace)

	scale, err := deploymentClient.GetScale(ctx, name, globalGetOptions)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting scale of deployment '%v' in namespace '%v'.", name, namespace)
	}

	oldReplicas := scale.Spec.Replicas
	scale.Spec.Replicas = replicas
	_, err = deploymentClient.UpdateScale(ctx, name, scale, metav1.UpdateOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		DryRun:          nil,
		FieldManager:    "",
		FieldValidation: "",
	})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred updating scale of deployment '%v' in namespace '%v' from '%v' to '%v'.", name, namespace, oldReplicas, replicas)
	}

	logrus.Debugf("Deployment '%s' scaled to %d\n", name, replicas)

	return nil
}

func (manager *KubernetesManager) GetPodsManagedByDeployment(ctx context.Context, deployment *v1.Deployment) ([]*apiv1.Pod, error) {
	podsClient := manager.kubernetesClientSet.CoreV1().Pods(deployment.Namespace)

	selector := metav1.FormatLabelSelector(deployment.Spec.Selector)

	pods, err := podsClient.List(ctx, metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		LabelSelector:        selector,
		FieldSelector:        "",
		Watch:                false,
		AllowWatchBookmarks:  false,
		ResourceVersion:      "",
		ResourceVersionMatch: "",
		TimeoutSeconds:       nil,
		Limit:                0,
		Continue:             "",
		SendInitialEvents:    nil,
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred retrieving list of pods in namespace '%v' with label selectors: %v.", deployment.Namespace, selector)
	}

	var podsManagedByDeployment []*apiv1.Pod
	for _, pod := range pods.Items {
		podToAdd := pod
		podsManagedByDeployment = append(podsManagedByDeployment, &podToAdd)
	}

	return podsManagedByDeployment, nil
}

// ---------------------------config map---------------------------------------------------------------------------------------
func (manager *KubernetesManager) RemoveConfigMap(ctx context.Context, namespace string, configMap *apiv1.ConfigMap) error {
	client := manager.kubernetesClientSet.CoreV1().ConfigMaps(namespace)

	if err := client.Delete(ctx, configMap.Name, globalDeleteOptions); err != nil {
		return stacktrace.Propagate(err, "Failed to delete config map with name '%s' with delete options '%+v'", configMap.Name, globalDeleteOptions)
	}

	return nil
}

func (manager *KubernetesManager) GetConfigMap(ctx context.Context, namespace string, name string) (*apiv1.ConfigMap, error) {
	client := manager.kubernetesClientSet.CoreV1().ConfigMaps(namespace)

	configMap, err := client.Get(ctx, name, metav1.GetOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		ResourceVersion: "",
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get config map with name '%s'", name)
	}

	return configMap, nil
}

func (manager *KubernetesManager) CreateConfigMap(
	ctx context.Context,
	namespaceName string,
	configMapName string,
	labels map[string]string,
	annotations map[string]string,
	data map[string]string,
) (*apiv1.ConfigMap, error) {
	client := manager.kubernetesClientSet.CoreV1().ConfigMaps(namespaceName)

	configMapToCreate := &apiv1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            configMapName,
			GenerateName:    "",
			Namespace:       namespaceName,
			SelfLink:        "",
			UID:             "",
			ResourceVersion: "",
			Generation:      0,
			CreationTimestamp: metav1.Time{
				Time: time.Time{},
			},
			DeletionTimestamp:          nil,
			DeletionGracePeriodSeconds: nil,
			Labels:                     labels,
			Annotations:                annotations,
			OwnerReferences:            nil,
			Finalizers:                 nil,
			ManagedFields:              nil,
		},
		Immutable:  nil,
		Data:       data,
		BinaryData: nil,
	}

	createdConfigMap, err := client.Create(ctx, configMapToCreate, globalCreateOptions)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating config map.")
	}

	return createdConfigMap, nil
}

func (kubernetesManager *KubernetesManager) GetVolumeSourceForHostPath(mountPath string) apiv1.VolumeSource {
	return apiv1.VolumeSource{
		HostPath: &apiv1.HostPathVolumeSource{
			Path: mountPath,
			Type: nil,
		},
		EmptyDir:              nil,
		GCEPersistentDisk:     nil,
		AWSElasticBlockStore:  nil,
		GitRepo:               nil,
		Secret:                nil,
		NFS:                   nil,
		ISCSI:                 nil,
		Glusterfs:             nil,
		PersistentVolumeClaim: nil,
		RBD:                   nil,
		FlexVolume:            nil,
		Cinder:                nil,
		CephFS:                nil,
		Flocker:               nil,
		DownwardAPI:           nil,
		FC:                    nil,
		AzureFile:             nil,
		ConfigMap:             nil,
		VsphereVolume:         nil,
		Quobyte:               nil,
		AzureDisk:             nil,
		PhotonPersistentDisk:  nil,
		Projected:             nil,
		PortworxVolume:        nil,
		ScaleIO:               nil,
		StorageOS:             nil,
		CSI:                   nil,
		Ephemeral:             nil,
	}
}

func (kubernetesManager *KubernetesManager) GetVolumeSourceForConfigMap(configMapName string) apiv1.VolumeSource {
	return apiv1.VolumeSource{
		ConfigMap: &apiv1.ConfigMapVolumeSource{
			LocalObjectReference: apiv1.LocalObjectReference{Name: configMapName},
			Items:                nil,
			DefaultMode:          nil,
			Optional:             nil,
		},
		HostPath:              nil,
		EmptyDir:              nil,
		GCEPersistentDisk:     nil,
		AWSElasticBlockStore:  nil,
		GitRepo:               nil,
		Secret:                nil,
		NFS:                   nil,
		ISCSI:                 nil,
		Glusterfs:             nil,
		PersistentVolumeClaim: nil,
		RBD:                   nil,
		FlexVolume:            nil,
		Cinder:                nil,
		CephFS:                nil,
		Flocker:               nil,
		DownwardAPI:           nil,
		FC:                    nil,
		AzureFile:             nil,
		VsphereVolume:         nil,
		Quobyte:               nil,
		AzureDisk:             nil,
		PhotonPersistentDisk:  nil,
		Projected:             nil,
		PortworxVolume:        nil,
		ScaleIO:               nil,
		StorageOS:             nil,
		CSI:                   nil,
		Ephemeral:             nil,
	}
}

// GetContainerLogs gets the logs for a given container running inside the given pod in the give namespace
// TODO We could upgrade this to get the logs of many containers at once just like kubectl does, see:
//
//	https://github.com/kubernetes/kubectl/blob/master/pkg/cmd/logs/logs.go#L345
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
) {
	options := &apiv1.PodLogOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		Container:                    containerName,
		Follow:                       shouldFollowLogs,
		Previous:                     false,
		SinceSeconds:                 nil,
		SinceTime:                    nil,
		Timestamps:                   shouldAddTimestamps,
		TailLines:                    nil,
		LimitBytes:                   nil,
		InsecureSkipTLSVerifyBackend: false,
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

// RunExecCommandWithContext This runs the exec to kubernetes with context, therefore
// when context timeouts it stops the process.
// TODO: merge RunExecCommand and this to one method
//
//	Doing this for now to unblock myself for wait worflows for k8s
//	In next PR, will include add context to WaitForPortAvailabilityUsingNetstat and
//	CopyFilesFromUserService method. I am doing this to reduce the blast radius.
func (manager *KubernetesManager) RunExecCommandWithContext(
	ctx context.Context,
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
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		Stdin:     shouldAllocateStdinOnPodExec,
		Stdout:    shouldAllocatedStdoutOnPodExec,
		Stderr:    shouldAllocatedStderrOnPodExec,
		TTY:       shouldAllocateTtyOnPodExec,
		Container: containerName,
		Command:   command,
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

	if err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:             nil,
		Stdout:            stdOutOutput,
		Stderr:            stdErrOutput,
		Tty:               false,
		TerminalSizeQueue: nil,
	}); err != nil {
		// Kubernetes returns the exit code of the command via a string in the error message, so we have to extract it
		statusError := err.Error()

		// this means that context deadline has exceeded
		if strings.Contains(statusError, contextDeadlineExceeded) {
			return 1, stacktrace.Propagate(err, "There was an error occurred while executing commands on the container")
		}

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
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		Stdin:     shouldAllocateStdinOnPodExec,
		Stdout:    shouldAllocatedStdoutOnPodExec,
		Stderr:    shouldAllocatedStderrOnPodExec,
		TTY:       shouldAllocateTtyOnPodExec,
		Container: containerName,
		Command:   command,
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

	if err = exec.StreamWithContext(context.Background(), remotecommand.StreamOptions{
		Stdin:             nil,
		Stdout:            stdOutOutput,
		Stderr:            stdErrOutput,
		Tty:               false,
		TerminalSizeQueue: nil,
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

func (manager *KubernetesManager) RunExecCommandWithStreamedOutput(
	ctx context.Context,
	namespaceName string,
	podName string,
	containerName string,
	command []string,
) (chan string, chan *exec_result.ExecResult, error) {
	execOptions := &apiv1.PodExecOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
		Container: containerName,
		Command:   command,
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
		return nil, nil, stacktrace.NewError("Failed to build a working RESTful request for the command '%s'.", execOptions.Command)
	}

	exec, err := remotecommand.NewSPDYExecutor(manager.kuberneteRestConfig, http.MethodPost, request.URL())
	if err != nil {
		return nil, nil, stacktrace.Propagate(err,
			"Failed to build an executor for the command '%s' with the RESTful endpoint '%s'.",
			execOptions.Command,
			request.URL().String())
	}

	execOutputChan := make(chan string)
	finalExecResultChan := make(chan *exec_result.ExecResult)
	go func() {
		defer func() {
			close(execOutputChan)
			close(finalExecResultChan)
		}()

		// Stream output from k8s to output channel
		channelWriter := channel_writer.NewChannelWriter(execOutputChan)
		if err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
			Stdin:             nil,
			Stdout:            channelWriter,
			Stderr:            channelWriter,
			Tty:               true,
			TerminalSizeQueue: nil,
		}); err != nil {
			// Kubernetes returns the exit code of the command via a string in the error message, so we have to extract it
			statusError := err.Error()

			// this means that context deadline has exceeded
			if strings.Contains(statusError, contextDeadlineExceeded) {
				execOutputChan <- stacktrace.Propagate(err, "There was an error occurred while executing commands on the container").Error()
				return
			}

			exitCode, err := getExitCodeFromStatusMessage(statusError)
			if err != nil {
				execOutputChan <- stacktrace.Propagate(err, "There was an error trying to parse the message '%s' to an exit code.", statusError).Error()
				return
			}

			// Don't send output in final result because it was already streamed
			finalExecResultChan <- exec_result.NewExecResult(exitCode, "")
		}
	}()
	return execOutputChan, finalExecResultChan, nil
}

// RemoveDirPathFromNode removes the contents and path to [dirPathToRemove] by creating a pod in [namespace] with privileged access to the [nodeName]'s filesystem
// The host filesystem is mounted onto the pod as a volume and then a rm -rf is run at the location on the pod where [dirPathToRemove] is mounted
func (manager *KubernetesManager) RemoveDirPathFromNode(ctx context.Context, namespace string, nodeName string, dirPathToRemove string) error {
	// rm the directory from the node using a privileged pod
	removeContainerName := "remove-dir-container"
	// pod needs to be privileged to access host filesystem
	isPrivileged := true
	removeDataDirPodUUID, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred generating uuid for remove data dir pod.")
	}
	removeDataDirPodName := fmt.Sprintf("remove-dir-pod-%v", removeDataDirPodUUID)
	hostVolumeName := "remove-dir-vol"
	mountPath := "/dir-to-remove"
	nodeSelectorsToSchedulePodOnNode := map[string]string{
		apiv1.LabelHostname: nodeName,
	}
	removeDataDirPod, err := manager.CreatePod(ctx,
		namespace,
		removeDataDirPodName,
		nil,
		nil,
		nil,
		[]apiv1.Container{
			{
				Name:  removeContainerName,
				Image: "busybox",
				Command: []string{
					"sh",
					"-c",
					"sleep 10000000s",
				},
				Args:       nil,
				WorkingDir: "",
				Ports:      nil,
				EnvFrom:    nil,
				Env:        nil,
				Resources: apiv1.ResourceRequirements{
					Limits:   nil,
					Requests: nil,
					Claims:   nil,
				},
				ResizePolicy: nil,
				VolumeMounts: []apiv1.VolumeMount{
					{
						Name:             hostVolumeName,
						ReadOnly:         false,
						MountPath:        mountPath,
						SubPath:          "",
						MountPropagation: nil,
						SubPathExpr:      "",
					},
				},
				VolumeDevices:            nil,
				LivenessProbe:            nil,
				ReadinessProbe:           nil,
				StartupProbe:             nil,
				Lifecycle:                nil,
				TerminationMessagePath:   "",
				TerminationMessagePolicy: "",
				ImagePullPolicy:          "",
				SecurityContext: &apiv1.SecurityContext{
					Privileged:               &isPrivileged,
					Capabilities:             nil,
					SeccompProfile:           nil,
					ProcMount:                nil,
					ReadOnlyRootFilesystem:   nil,
					AllowPrivilegeEscalation: nil,
					RunAsNonRoot:             nil,
					RunAsGroup:               nil,
					RunAsUser:                nil,
					SELinuxOptions:           nil,
					WindowsOptions:           nil,
				},
				Stdin:     false,
				StdinOnce: false,
				TTY:       false,
			},
		},
		[]apiv1.Volume{
			{
				Name:         hostVolumeName,
				VolumeSource: manager.GetVolumeSourceForHostPath(dirPathToRemove), // mount the entire host filesystem in this volume
			},
		},
		"",
		"",
		nil,
		nodeSelectorsToSchedulePodOnNode,
		shouldUseHostPidsNamespace,
		shouldUseHostNetworksNamespace,
	)
	defer func() {
		// Don't block on removing this remove directory pod because this can take a while sometimes in k8s
		go func() {
			removeCtx := context.Background()
			if removeDataDirPod != nil {
				err := manager.RemovePod(removeCtx, removeDataDirPod)
				if err != nil {
					logrus.Warnf("Attempted to remove pod '%v' in namespace '%v' but an error occurred:\n%v", removeDataDirPod.Name, namespace, err.Error())
					logrus.Warn("You may have to remove this pod manually.")
				}
			}
		}()
	}()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating pod '%v' in namespace '%v'.", removeDataDirPodName, namespace)
	}

	removeDirSuccessExitCode := int32(0)
	dirPathToRemoveAndEmpty := fmt.Sprintf("%v/*", mountPath)
	removeDirCmd := []string{"rm", "-rf", dirPathToRemoveAndEmpty}
	output := &bytes.Buffer{}
	concurrentWriter := concurrent_writer.NewConcurrentWriter(output)
	resultExitCode, err := manager.RunExecCommand(
		removeDataDirPod.Namespace,
		removeDataDirPod.Name,
		removeContainerName,
		removeDirCmd,
		concurrentWriter,
		concurrentWriter,
	)
	logrus.Debugf("Output of remove directory '%v': %v, exit code: %v", removeDirCmd, output.String(), resultExitCode)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred running exec command '%v' on pod '%v' in namespace '%v' with output '%v'.", removeDirCmd, removeDataDirPod.Name, removeDataDirPod.Namespace, output.String())
	}
	if resultExitCode != removeDirSuccessExitCode {
		return stacktrace.NewError("Running exec command '%v' on pod '%v' in namespace '%v' returned a non-%v exit code: '%v' and output '%v'.", removeDirCmd, removeDataDirPod.Name, removeDataDirPod.Namespace, removeDirSuccessExitCode, resultExitCode, output.String())
	}
	if output.String() != "" {
		return stacktrace.NewError("Expected empty output from running exec command '%v' but instead retrieved output string '%v'", removeDirCmd, output.String())
	}

	logrus.Debugf("Successfully removed contents of dir path '%v' on node '%v'.", dirPathToRemove, nodeName)
	return nil
}

func (manager *KubernetesManager) GetAllEnclaveResourcesByLabels(ctx context.Context, namespace string, labels map[string]string) (*apiv1.PodList, *apiv1.ServiceList, *rbacv1.ClusterRoleList, *rbacv1.ClusterRoleBindingList, error) {

	var (
		wg                      = sync.WaitGroup{}
		errChan                 = make(chan error)
		allCallsDoneChan        = make(chan bool)
		podsList                *apiv1.PodList
		servicesList            *apiv1.ServiceList
		clusterRolesList        *rbacv1.ClusterRoleList
		clusterRoleBindingsList *rbacv1.ClusterRoleBindingList
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		podsList, err = manager.GetPodsByLabels(ctx, namespace, labels)
		if err != nil {
			errChan <- stacktrace.Propagate(err, "Expected to be able to get pods with labels '%+v', instead a non-nil error was returned", labels)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		servicesList, err = manager.GetServicesByLabels(ctx, namespace, labels)
		if err != nil {
			errChan <- stacktrace.Propagate(err, "Expected to be able to get services with labels '%+v', instead a non-nil error was returned", labels)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		clusterRolesList, err = manager.GetClusterRolesByLabels(ctx, labels)
		if err != nil {
			errChan <- stacktrace.Propagate(err, "Expected to be able to get services with labels '%+v', instead a non-nil error was returned", labels)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		clusterRoleBindingsList, err = manager.GetClusterRoleBindingsByLabels(ctx, labels)
		if err != nil {
			errChan <- stacktrace.Propagate(err, "Expected to be able to get services with labels '%+v', instead a non-nil error was returned", labels)
		}
	}()

	go func() {
		wg.Wait()
		close(allCallsDoneChan)
	}()

	select {
	case <-allCallsDoneChan:
		break
	case err, isChanOpen := <-errChan:
		if isChanOpen {
			return nil, nil, nil, nil, stacktrace.NewError("The error chan has been closed; this is a bug in Kurtosis")
		}
		if err != nil {
			return nil, nil, nil, nil, stacktrace.Propagate(err, "An error occurred getting pods and services for labels '%+v' in namespace '%s'", labels, namespace)
		}
	}

	return podsList, servicesList, clusterRolesList, clusterRoleBindingsList, nil
}

func (manager *KubernetesManager) GetPodsByLabels(ctx context.Context, namespace string, podLabels map[string]string) (*apiv1.PodList, error) {
	namespacePodClient := manager.kubernetesClientSet.CoreV1().Pods(namespace)

	opts := buildListOptionsFromLabels(podLabels)
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

func (manager *KubernetesManager) GetDaemonSetsByLabels(ctx context.Context, namespace string, daemonSetLabels map[string]string) (*v1.DaemonSetList, error) {
	namespaceDaemonSetClient := manager.kubernetesClientSet.AppsV1().DaemonSets(namespace)

	opts := buildListOptionsFromLabels(daemonSetLabels)
	daemonSets, err := namespaceDaemonSetClient.List(ctx, opts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get daemon sets with labels '%+v', instead a non-nil error was returned", daemonSetLabels)
	}

	return daemonSets, nil
}

func (manager *KubernetesManager) GetDeploymentsByLabels(ctx context.Context, namespace string, deploymentLabels map[string]string) (*v1.DeploymentList, error) {
	deploymentsClient := manager.kubernetesClientSet.AppsV1().Deployments(namespace)

	opts := buildListOptionsFromLabels(deploymentLabels)
	deployment, err := deploymentsClient.List(ctx, opts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get deployment with labels '%+v', instead a non-nil error was returned", deploymentLabels)
	}

	return deployment, nil
}

func (manager *KubernetesManager) GetConfigMapByLabels(ctx context.Context, namespace string, configMapLabels map[string]string) (*apiv1.ConfigMapList, error) {
	configMapClient := manager.kubernetesClientSet.CoreV1().ConfigMaps(namespace)

	opts := buildListOptionsFromLabels(configMapLabels)
	configMaps, err := configMapClient.List(ctx, opts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get daemonsets with labels '%+v', instead a non-nil error was returned", configMapLabels)
	}

	return configMaps, nil
}

func (manager *KubernetesManager) GetPodPortforwardEndpointUrl(namespace string, podName string) *url.URL {
	return manager.kubernetesClientSet.CoreV1().RESTClient().Post().Resource("pods").Namespace(namespace).Name(podName).SubResource("portforward").URL()
}

func (manager *KubernetesManager) GetExecStream(ctx context.Context, pod *apiv1.Pod) error {
	containerName := pod.Spec.Containers[0].Name
	request := manager.kubernetesClientSet.CoreV1().RESTClient().Post().Resource("pods").Name(pod.Name).Namespace(pod.Namespace).SubResource("exec")
	// lifted from https://github.com/kubernetes/client-go/issues/912 - the terminal magic is still magical
	request.VersionedParams(&apiv1.PodExecOptions{
		Container: containerName,
		Command:   commandToRunWhenCreatingUserServiceShell,
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
	}, scheme.ParameterCodec)
	exec, err := remotecommand.NewSPDYExecutor(manager.kuberneteRestConfig, "POST", request.URL())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while creating a new SPDY executor")
	}
	stdinFd := int(os.Stdin.Fd())
	var oldState *terminal.State
	if terminal.IsTerminal(stdinFd) {
		oldState, err = terminal.MakeRaw(stdinFd)
		if err != nil {
			// print error
			return stacktrace.Propagate(err, "An error occurred making STDIN stream raw")
		}
		defer func() {
			if err = terminal.Restore(stdinFd, oldState); err != nil {
				logrus.Warn("An error occurred while restoring the terminal to its normal state. Your terminal might look funny; we recommend closing and starting a new terminal.")
			}
		}()
	}
	return exec.StreamWithContext(
		ctx,
		remotecommand.StreamOptions{
			TerminalSizeQueue: nil,
			Stdin:             os.Stdin,
			Stdout:            os.Stdout,
			Stderr:            os.Stderr,
			Tty:               true,
		})
}

func (manager *KubernetesManager) HasComputeNodes(ctx context.Context) (bool, error) {
	nodes, err := manager.kubernetesClientSet.CoreV1().Nodes().List(ctx, globalListOptions)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred while checking if the Kubernetes cluster has any nodes")
	}
	return len(nodes.Items) != 0, nil
}

// AddLabelsToNode will add kurtosis related [labels] from [nodeName] - non Kurtosis labels will not be allowed for addition
func (manager *KubernetesManager) AddLabelsToNode(ctx context.Context, nodeName string, labels map[string]string) error {
	for k := range labels {
		if !strings.HasPrefix(k, kubernetes_label_key.KurtosisDomainLabelKeyPrefix.GetString()) {
			return stacktrace.NewError("Found label '%v' not prefixed with Kurtosis app id label '%v'. Adding non-Kurtosis label is disallowed.", k, kubernetes_label_key.AppIDKubernetesLabelKey.GetString())
		}
	}
	nodeClient := manager.kubernetesClientSet.CoreV1().Nodes()

	node, err := nodeClient.Get(ctx, nodeName, globalGetOptions)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while trying to get node '%v'. Ensure node with name '%v' exists in cluster.", nodeName, nodeName)
	}

	// add to existing labels
	for k, v := range labels {
		node.Labels[k] = v
	}

	_, err = nodeClient.Update(ctx, node, metav1.UpdateOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		DryRun:          nil,
		FieldManager:    "",
		FieldValidation: "",
	})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while trying to update node '%v' with labels '%v'", nodeName, labels)
	}

	return nil
}

// RemoveLabelsFromNode will remove kurtosis related [labels] from [nodeName] - non Kurtosis labels will not be allowed for removal
func (manager *KubernetesManager) RemoveLabelsFromNode(ctx context.Context, nodeName string, labels map[string]bool) error {
	for k := range labels {
		if !strings.HasPrefix(k, kubernetes_label_key.KurtosisDomainLabelKeyPrefix.GetString()) {
			return stacktrace.NewError("Found label '%v' not prefixed with Kurtosis domain prefix '%v'. Removing non-Kurtosis label is disallowed.", k, kubernetes_label_key.AppIDKubernetesLabelKey.GetString())
		}
	}
	nodeClient := manager.kubernetesClientSet.CoreV1().Nodes()

	// TODO: add check here
	node, err := nodeClient.Get(ctx, nodeName, globalGetOptions)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while trying to get node '%v'. Ensure node with name '%v' exists in cluster.", nodeName, nodeName)
	}

	// add node selectors to existing labels
	for k := range labels {
		delete(node.Labels, k)
	}

	_, err = nodeClient.Update(ctx, node, metav1.UpdateOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		DryRun:          nil,
		FieldManager:    "",
		FieldValidation: "",
	})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while trying to remove labels '%v' from node '%v'", labels, nodeName)
	}

	logrus.Debugf("Successfullly removed label '%v' from node '%v'.", labels, nodeName)

	return nil
}

func (manager *KubernetesManager) GetLabelsOnNode(ctx context.Context, nodeName string) (map[string]string, error) {
	nodeClient := manager.kubernetesClientSet.CoreV1().Nodes()

	node, err := nodeClient.Get(ctx, nodeName, globalGetOptions)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while trying to get node '%v'. Ensure node with name '%v' exists in cluster.", nodeName, nodeName)
	}

	return node.Labels, nil
}

// ---------------------------Ingresses------------------------------------------------------------------------------

func (manager *KubernetesManager) CreateIngress(ctx context.Context, namespace string, name string, labels map[string]string, annotations map[string]string, rules []netv1.IngressRule) (*netv1.Ingress, error) {
	client := manager.kubernetesClientSet.NetworkingV1().Ingresses(namespace)

	ingress := &netv1.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			GenerateName:    "",
			Namespace:       "",
			SelfLink:        "",
			UID:             "",
			ResourceVersion: "",
			Generation:      0,
			CreationTimestamp: metav1.Time{
				Time: time.Time{},
			},
			DeletionTimestamp:          nil,
			DeletionGracePeriodSeconds: nil,
			Labels:                     labels,
			Annotations:                annotations,
			OwnerReferences:            nil,
			Finalizers:                 nil,
			ManagedFields:              nil,
		},
		Spec: netv1.IngressSpec{
			IngressClassName: nil,
			DefaultBackend:   nil,
			TLS:              nil,
			Rules:            rules,
		},
		Status: netv1.IngressStatus{
			LoadBalancer: netv1.IngressLoadBalancerStatus{
				Ingress: nil,
			},
		},
	}

	ingressResult, err := client.Create(ctx, ingress, globalCreateOptions)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create the ingress with name '%s' in namespace '%v'", name, namespace)
	}
	return ingressResult, nil
}

func (manager *KubernetesManager) RemoveIngress(ctx context.Context, ingress *netv1.Ingress) error {
	namespace := ingress.Namespace
	ingressName := ingress.Name
	ingressesClient := manager.kubernetesClientSet.NetworkingV1().Ingresses(namespace)

	if err := ingressesClient.Delete(ctx, ingressName, globalDeleteOptions); err != nil {
		return stacktrace.Propagate(err, "Failed to delete ingress '%s' with delete options '%+v' in namespace '%s'", ingressName, globalDeleteOptions, namespace)
	}

	return nil
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
// TODO delete the following assuming we don't use it.
/*
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
		case apiv1.PodUnknown:
			// not impl - skipping
		case apiv1.PodRunning:
			return nil
		case apiv1.PodPending:
			for _, containerStatus := range pod.Status.ContainerStatuses {
				containerName := containerStatus.Name
				maybeContainerWaitingState := containerStatus.State.Waiting
				if maybeContainerWaitingState != nil && maybeContainerWaitingState.Reason == imagePullBackOffContainerReason {
					return stacktrace.NewError(
						"Container '%v' using image '%v' in pod '%v' in namespace '%v' is stuck in state '%v'. This likely means:\n"+
							"1) There's a typo in either the image name or the tag name\n"+
							"2) The image isn't accessible to Kubernetes (e.g. it's a local image, or it's in a private image registry that Kubernetes can't access)\n"+
							"3) The image's platform/architecture might not match",
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
			//NOTE: We'll need to change this if we ever expect to run one-off pods
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
		"Pod '%v' did not become available after %v; its latest state is '%v' and status message is: %v\n"+
			"The pod's container states are as follows:\n%v",
		podName,
		podWaitForAvailabilityTimeout,
		latestPodStatus.Phase,
		latestPodStatus.Message,
		strings.Join(containerStatusStrs, "\n"),
	)
}

// waitForPodDeletion waits for the pod to be fully deleted if it has been marked for deletion
func (manager *KubernetesManager) waitForPodDeletion(ctx context.Context, namespaceName string, podName string) error {
	// Wait for the pod to start running
	deadline := time.Now().Add(podWaitForDeletionTimeout)
	var latestPodStatus *apiv1.PodStatus
	for time.Now().Before(deadline) {
		pod, err := manager.GetPod(ctx, namespaceName, podName)
		if err != nil {
			// If an error has been returned, it's likely the pod does not exist. Continue
			return nil
		}
		if pod.DeletionTimestamp == nil {
			return stacktrace.NewError("The pod '%s' currently exists in namespace '%s' and is not scheduled for deletion",
				podName, namespaceName)
		}
		time.Sleep(podWaitForDeletionTimeBetweenPolls)
	}

	containerStatusStrs := renderContainerStatuses(latestPodStatus.ContainerStatuses, containerStatusLineBulletPoint)
	return stacktrace.NewError(
		"Pod '%v' wasn't deleted after %v; its latest state is '%v' and status message is: %v\n"+
			"The pod's container states are as follows:\n%v",
		podName,
		podWaitForAvailabilityTimeout,
		latestPodStatus.Phase,
		latestPodStatus.Message,
		strings.Join(containerStatusStrs, "\n"),
	)
}

func (manager *KubernetesManager) WaitForPodTermination(ctx context.Context, namespaceName string, podName string) error {
	deadline := time.Now().Add(podWaitForTerminationTimeout)
	var latestPodStatus *apiv1.PodStatus
	for time.Now().Before(deadline) {
		pod, err := manager.GetPod(ctx, namespaceName, podName)
		if err != nil {
			// The pod info is not always available after deletion, so we handle that gracefully
			logrus.Debugf("An error occurred trying to retrieve the just-deleted pod '%v': %v, for checking if it was successfully terminated; but we can ignore this error because the pod info is not always available after deletion", podName, err)
			return nil
		}

		latestPodStatus = &pod.Status
		switch latestPodStatus.Phase {
		case apiv1.PodPending:
		case apiv1.PodRunning:
		case apiv1.PodUnknown:
		case apiv1.PodSucceeded:
		case apiv1.PodFailed:
			return nil
		}
		time.Sleep(podWaitForTerminationTimeBetweenPolls)
	}

	containerStatusStrs := renderContainerStatuses(latestPodStatus.ContainerStatuses, containerStatusLineBulletPoint)
	return stacktrace.NewError(
		"Pod '%v' did not terminate after %v; its latest state is '%v' and status message is: %v\n"+
			"The pod's container states are as follows:\n%v",
		podName,
		podWaitForTerminationTimeout,
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
	if err != nil {
		return fmt.Sprintf("Cannot display container logs because an error occurred getting the logs:\n%v", err)
	}
	defer containerLogs.Close()

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
				"Unrecogznied container state '%+v'; this likely means that Kubernetes "+
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
func getExitCodeFromStatusMessage(statusMessage string) (int32, error) {
	messageSlice := strings.Split(statusMessage, " ")
	if len(messageSlice) != expectedStatusMessageSliceSize {
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

func buildListOptionsFromLabels(labelsMap map[string]string) metav1.ListOptions {
	return metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		LabelSelector:        labels.SelectorFromSet(labelsMap).String(),
		FieldSelector:        "",
		Watch:                false,
		AllowWatchBookmarks:  false,
		ResourceVersion:      "",
		ResourceVersionMatch: "",
		TimeoutSeconds:       int64Ptr(listOptionsTimeoutSeconds),
		Limit:                0,
		Continue:             "",
		SendInitialEvents:    nil,
	}
}
