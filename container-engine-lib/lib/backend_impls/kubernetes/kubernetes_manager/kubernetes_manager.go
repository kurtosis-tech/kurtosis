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
	"io"
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
		return nil, stacktrace.Propagate(err, "Namespace with name '%s' has been marked for deletion", namespace)
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

func (manager *KubernetesManager) CreateServiceAccount(ctx context.Context, name string, namespace string, labels map[string]string) (*apiv1.ServiceAccount, error) {
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
		Secrets: nil,
		ImagePullSecrets: []apiv1.LocalObjectReference{
			{
				Name: "kurtosis-image",
			},
		},
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
) (*apiv1.Pod, error) {
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
		HostNetwork:                   false,
		HostPID:                       false,
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

	if err := manager.waitForPodTermination(ctx, namespace, name); err != nil {
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

func (manager *KubernetesManager) waitForPodTermination(ctx context.Context, namespaceName string, podName string) error {
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
