/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package k8s_manager

import (
	"context"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

type K8sManager struct {
	// The logger that all log messages will be written to
	log *logrus.Logger // NOTE: This log should be used for all log statements - the system-wide logger should NOT be used!

	// The underlying K8s client that will be used to /modify the K8s environment
	k8sClientSet *kubernetes.Clientset
}

/*
NewK8sManager
Creates a new K8s manager for manipulating the k8s cluster using the given client.

Args:
	log: The logger that this K8s manager will write all its log messages to.
	k8sClientSet: The k8s client that will be used when interacting with the underlying k8s cluster.
*/
func NewK8sManager(log *logrus.Logger, k8sClientSet *kubernetes.Clientset) *K8sManager {
	return &K8sManager{
		log:          log,
		k8sClientSet: k8sClientSet,
	}
}

// ---------------------------Deployments------------------------------------------------------------------------------

/*
CreateDeployment
Creates a new k8s deployment with the given parameters

Args:


Returns:
	id: The deployment ID
*/
func (manager K8sManager) CreateDeployment(context context.Context, deploymentName string, namespace string, labels map[string]string, containerImage string, replicas int32, volumes []apiv1.Volume, volumeMounts []apiv1.VolumeMount, envVars map[string]string, containerName string) (*appsv1.Deployment, error) {
	deploymentsClient := manager.k8sClientSet.AppsV1().Deployments(namespace)

	var podEnvVars []apiv1.EnvVar

	for k, v := range envVars {
		envVar := apiv1.EnvVar{
			Name:  k,
			Value: v,
		}
		podEnvVars = append(podEnvVars, envVar)
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:   deploymentName,
			Labels: labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: manager.int32Ptr(replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: apiv1.PodSpec{
					Volumes: volumes,
					Containers: []apiv1.Container{
						{
							Name:         containerName,
							Image:        containerImage,
							VolumeMounts: volumeMounts,
							Env:          podEnvVars,
						},
					},
				},
			},
		},
	}

	deploymentResult, err := deploymentsClient.Create(context, deployment, metav1.CreateOptions{})

	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create deployment %s", deploymentName)
	}

	return deploymentResult, nil
}

func (manager K8sManager) ListDeployments(ctx context.Context) (*appsv1.DeploymentList, error) {
	deploymentsClient := manager.k8sClientSet.AppsV1().Deployments(apiv1.NamespaceDefault)

	list, err := deploymentsClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to list deployments")
	}

	return list, nil
}

func (manager K8sManager) RemoveDeployment(context context.Context, namespace string, name string) error {
	deploymentsClient := manager.k8sClientSet.AppsV1().Deployments(namespace)

	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	if err := deploymentsClient.Delete(context, name, deleteOptions); err != nil {
		return stacktrace.Propagate(err, "Failed to delete deployment %s", name)
	}
	return nil
}

func (manager K8sManager) GetDeploymentByName(context context.Context, namespace string, name string) (*appsv1.Deployment, error) {
	deploymentsClient := manager.k8sClientSet.AppsV1().Deployments(namespace)

	deployment, err := deploymentsClient.Get(context, name, metav1.GetOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get deployment %s", name)
	}
	return deployment, nil
}

func (manager K8sManager) GetDeploymentsByLabels(context context.Context, namespace string, deploymentLabels map[string]string) (*appsv1.DeploymentList, error) {
	deploymentsClient := manager.k8sClientSet.AppsV1().Deployments(namespace)

	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(deploymentLabels).String(),
	}

	deployments, err := deploymentsClient.List(context, listOptions)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get deployment by labels %v", deploymentLabels)
	}
	return deployments, nil
}

func (manager K8sManager) UpdateDeploymentReplicas(context context.Context, namespace string, deploymentLabels map[string]string, replicas int32) error {
	deploymentsClient := manager.k8sClientSet.AppsV1().Deployments(apiv1.NamespaceDefault)

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Deployment before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		result, err := manager.GetDeploymentsByLabels(context, namespace, deploymentLabels)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to get deployments by labels %v", deploymentLabels)
		}

		for _, deployment := range result.Items {
			deployment.Spec.Replicas = manager.int32Ptr(replicas)
			_, err = deploymentsClient.Update(context, &deployment, metav1.UpdateOptions{})
			if err != nil{
				return err
			}
		}
		return err
	})
	if retryErr != nil {
		stacktrace.Propagate(retryErr, "Failed to update deployment replicas by labels %v", deploymentLabels)
	}

	return nil
}

// ---------------------------Services------------------------------------------------------------------------------

func (manager K8sManager) CreateService(context context.Context, name string, namespace string, serviceLabels map[string]string, serviceType apiv1.ServiceType, port int32, targetPort int32) (*apiv1.Service, error) {

	servicesClient := manager.k8sClientSet.CoreV1().Services(namespace)

	service := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: serviceLabels,
		},
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort{
				{
					Protocol: "TCP",
					Port:     port,
					TargetPort: intstr.IntOrString{
						IntVal: targetPort, // internal container port
					},
				},
			},
			Selector: serviceLabels, // these labels are used to match with the Pod
			Type:     serviceType,
		},
	}

	serviceResult, err := servicesClient.Create(context, service, metav1.CreateOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create service %s", name)
	}

	return serviceResult, nil
}

func (manager K8sManager) RemoveService(context context.Context, namespace string, name string) error {
	servicesClient := manager.k8sClientSet.CoreV1().Services(namespace)

	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	if err := servicesClient.Delete(context, name, deleteOptions); err != nil {
		return stacktrace.Propagate(err, "Failed to delete service %s", name)
	}

	return nil
}

func (manager K8sManager) GetServiceByName(context context.Context, namespace string, name string) (*apiv1.Service, error) {
	servicesClient := manager.k8sClientSet.CoreV1().Services(namespace)

	serviceResult, err := servicesClient.Get(context, name, metav1.GetOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get service %s", name)
	}

	return serviceResult, nil
}

func (manager K8sManager) ListServices(context context.Context) (*apiv1.ServiceList, error) {
	servicesClient := manager.k8sClientSet.CoreV1().Services(apiv1.NamespaceDefault)

	serviceResult, err := servicesClient.List(context, metav1.ListOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to list services")
	}

	return serviceResult, nil
}

func (manager K8sManager) GetServicesByLabels(context context.Context, serviceLabels map[string]string) (*apiv1.ServiceList, error) {
	servicesClient := manager.k8sClientSet.CoreV1().Services(apiv1.NamespaceDefault)

	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(serviceLabels).String(),
	}

	serviceResult, err := servicesClient.List(context, listOptions)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to list services")
	}

	return serviceResult, nil
}

// ---------------------------Volumes------------------------------------------------------------------------------

func (manager K8sManager) CreateStorageClass(context context.Context, name string, provisioner string, volumeBindingMode storagev1.VolumeBindingMode) (*storagev1.StorageClass, error) {
	storageClassClient := manager.k8sClientSet.StorageV1().StorageClasses()

	//volumeBindingMode := storagev1.VolumeBindingWaitForFirstConsumer
	//provisioner := "kubernetes.io/no-provisioner"

	storageClass := &storagev1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Provisioner:       provisioner,
		VolumeBindingMode: &volumeBindingMode,
	}

	storageClassResult, err := storageClassClient.Create(context, storageClass, metav1.CreateOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create storage class with name %s", name)
	}

	return storageClassResult, nil
}

func (manager K8sManager) RemoveStorageClass(context context.Context, name string) error {
	storageClassClient := manager.k8sClientSet.StorageV1().StorageClasses()

	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	err := storageClassClient.Delete(context, name, deleteOptions)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to delete storage class with name %s", name)
	}

	return nil
}

func (manager K8sManager) GetStorageClass(context context.Context, name string) (*storagev1.StorageClass, error) {
	storageClassClient := manager.k8sClientSet.StorageV1().StorageClasses()

	storageClassResult, err := storageClassClient.Get(context, name, metav1.GetOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get storage class with name %s", name)
	}

	return storageClassResult, nil
}

func (manager K8sManager) CreatePersistentVolume(context context.Context, volumeName string, volumeLabels map[string]string, quantity string, path string, storageClassName string) (*apiv1.PersistentVolume, error) {
	volumesClient := manager.k8sClientSet.CoreV1().PersistentVolumes()

	//quantity := "100Gi"
	//storageClassName := "my-local-storage"
	//path := "/Users/mariofernandez/Library/Application Support/kurtosis/engine-data"

	persistentVolume := &apiv1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:   volumeName,
			Labels: volumeLabels,
		},
		Spec: apiv1.PersistentVolumeSpec{
			Capacity: map[apiv1.ResourceName]resource.Quantity{
				apiv1.ResourceStorage: resource.MustParse(quantity),
			},
			PersistentVolumeSource: apiv1.PersistentVolumeSource{
				HostPath: &apiv1.HostPathVolumeSource{
					Path: path,
				},
			},
			AccessModes: []apiv1.PersistentVolumeAccessMode{
				apiv1.ReadWriteOnce,
			},
			StorageClassName: storageClassName,
		},
	}

	persistentVolumeResult, err := volumesClient.Create(context, persistentVolume, metav1.CreateOptions{})

	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create persistent volume with name %s", volumeName)
	}

	return persistentVolumeResult, nil
}

func (manager K8sManager) RemovePersistentVolume(context context.Context, volumeName string) error {
	volumesClient := manager.k8sClientSet.CoreV1().PersistentVolumes()

	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	if err := volumesClient.Delete(context, volumeName, deleteOptions); err != nil {
		return stacktrace.Propagate(err, "Failed to create persistent volume with name %s", volumeName)
	}

	return nil
}

func (manager K8sManager) GetPersistentVolume(context context.Context, volumeName string) (*apiv1.PersistentVolume, error) {
	volumesClient := manager.k8sClientSet.CoreV1().PersistentVolumes()

	persistentVolumeResult, err := volumesClient.Get(context, volumeName, metav1.GetOptions{})

	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create persistent volume with name %s", volumeName)
	}

	return persistentVolumeResult, nil
}

func (manager K8sManager) ListPersistentVolumes(context context.Context, volumeName string) (*apiv1.PersistentVolumeList, error) {
	volumesClient := manager.k8sClientSet.CoreV1().PersistentVolumes()

	persistentVolumesResult, err := volumesClient.List(context, metav1.ListOptions{})

	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create persistent volume with name %s", volumeName)
	}

	return persistentVolumesResult, nil
}

func (manager K8sManager) GetPersistentVolumesByLabels(context context.Context, persistentVolumeLabels map[string]string) (*apiv1.PersistentVolumeList, error) {
	volumesClient := manager.k8sClientSet.CoreV1().PersistentVolumes()

	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(persistentVolumeLabels).String(),
	}

	persistentVolumesResult, err := volumesClient.List(context, listOptions)

	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create persistent volume with name %v", persistentVolumeLabels)
	}

	return persistentVolumesResult, nil
}

func (manager K8sManager) CreatePersistentVolumeClaim(context context.Context, namespace string, persistentVolumeClaimName string, persistentVolumeClaimLabels map[string]string, quantity string, storageClassName string) (*apiv1.PersistentVolumeClaim, error) {
	volumeClaimsClient := manager.k8sClientSet.CoreV1().PersistentVolumeClaims(namespace)

	//storageClassName := "my-local-storage"
	//quantity := "10Gi"

	persistentVolumeClaim := &apiv1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:   persistentVolumeClaimName,
			Labels: persistentVolumeClaimLabels,
		},
		Spec: apiv1.PersistentVolumeClaimSpec{
			AccessModes: []apiv1.PersistentVolumeAccessMode{
				apiv1.ReadWriteOnce,
			},
			StorageClassName: &storageClassName,
			Resources: apiv1.ResourceRequirements{
				Requests: map[apiv1.ResourceName]resource.Quantity{
					apiv1.ResourceStorage: resource.MustParse(quantity),
				},
			},
		},
	}

	persistentVolumeClaimResult, err := volumeClaimsClient.Create(context, persistentVolumeClaim, metav1.CreateOptions{})

	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create persistent volume claim with name %s", persistentVolumeClaimName)
	}

	return persistentVolumeClaimResult, nil
}

func (manager K8sManager) RemovePersistentVolumeClaim(context context.Context, namespace string, persistentVolumeClaimName string) error {
	volumeClaimsClient := manager.k8sClientSet.CoreV1().PersistentVolumeClaims(namespace)

	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	if err := volumeClaimsClient.Delete(context, persistentVolumeClaimName, deleteOptions); err != nil {
		return stacktrace.Propagate(err, "Failed to delete persistent volume claim with name %s", persistentVolumeClaimName)
	}

	return nil
}

func (manager K8sManager) GetPersistentVolumeClaim(context context.Context, namespace string, persistentVolumeClaimName string) (*apiv1.PersistentVolumeClaim, error) {
	persistentVolumeClaimsClient := manager.k8sClientSet.CoreV1().PersistentVolumeClaims(namespace)

	volumeClaim, err := persistentVolumeClaimsClient.Get(context, persistentVolumeClaimName, metav1.GetOptions{})

	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get persistent volume claim with name %s", persistentVolumeClaimName)
	}

	return volumeClaim, nil
}

func (manager K8sManager) ListPersistentVolumeClaims(context context.Context, namespace string) (*apiv1.PersistentVolumeClaimList, error) {
	persistentVolumeClaimsClient := manager.k8sClientSet.CoreV1().PersistentVolumeClaims(namespace)

	persistentVolumeClaimsResult, err := persistentVolumeClaimsClient.List(context, metav1.ListOptions{})

	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to list persistent volume claims")
	}

	return persistentVolumeClaimsResult, nil
}

func (manager K8sManager) GetPersistentVolumeClaimsByLabels(context context.Context, namespace string, persistentVolumeClaimLabels map[string]string) (*apiv1.PersistentVolumeClaimList, error) {
	persistentVolumeClaimsClient := manager.k8sClientSet.CoreV1().PersistentVolumeClaims(namespace)

	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(persistentVolumeClaimLabels).String(),
	}

	persistentVolumeClaimsResult, err := persistentVolumeClaimsClient.List(context, listOptions)

	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to Get persistent volume claim with labels %v", persistentVolumeClaimLabels)
	}

	return persistentVolumeClaimsResult, nil
}

// ---------------------------namespaces------------------------------------------------------------------------------

func (manager K8sManager) CreateNamespace(context context.Context, name string, namespaceLabels map[string]string) (*apiv1.Namespace, error) {
	namespaceClient := manager.k8sClientSet.CoreV1().Namespaces()

	namespace := &apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: namespaceLabels,
		},
	}

	namespaceResult, err := namespaceClient.Create(context, namespace, metav1.CreateOptions{})

	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create namespace with name %s", name)
	}

	return namespaceResult, nil
}

func (manager K8sManager) RemoveNamespace(context context.Context, name string) error {
	namespaceClient := manager.k8sClientSet.CoreV1().Namespaces()

	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	if err := namespaceClient.Delete(context, name, deleteOptions); err != nil {
		return stacktrace.Propagate(err, "Failed to delete namespace with name %s", name)
	}

	return nil
}

func (manager K8sManager) GetNamespace(context context.Context, name string) (*apiv1.Namespace, error) {
	namespaceClient := manager.k8sClientSet.CoreV1().Namespaces()

	namespace, err := namespaceClient.Get(context, name, metav1.GetOptions{})

	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get namespace with name %s", name)
	}

	return namespace, nil
}

func (manager K8sManager) ListNamespaces(context context.Context) (*apiv1.NamespaceList, error) {
	namespaceClient := manager.k8sClientSet.CoreV1().Namespaces()

	namespaces, err := namespaceClient.List(context, metav1.ListOptions{})

	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to list namespaces")
	}

	return namespaces, nil
}

func (manager K8sManager) GetNamespacesByLabels(context context.Context, namespaceLabels map[string]string) (*apiv1.NamespaceList, error) {
	namespaceClient := manager.k8sClientSet.CoreV1().Namespaces()

	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(namespaceLabels).String(),
	}

	namespaces, err := namespaceClient.List(context, listOptions)

	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to list namespaces with labels %v", namespaceLabels)
	}

	return namespaces, nil
}

// ---------------------------DaemonSets------------------------------------------------------------------------------

func (manager K8sManager) CreateDaemonSet(context context.Context, namespace string, name string, daemonSetLabels map[string]string) (*appsv1.DaemonSet, error) {
	daemonSetClient := manager.k8sClientSet.AppsV1().DaemonSets(namespace)

	daemonSet := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: daemonSetLabels,
		},
		Spec: appsv1.DaemonSetSpec{},
	}

	daemonSetResult, err := daemonSetClient.Create(context, daemonSet, metav1.CreateOptions{})

	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create daemonSet with name %s", name)
	}

	return daemonSetResult, nil
}

func (manager K8sManager) RemoveDaemonSet(context context.Context, name string) error {

	daemonSetClient := manager.k8sClientSet.AppsV1().DaemonSets(apiv1.NamespaceDefault)

	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	if err := daemonSetClient.Delete(context, name, deleteOptions); err != nil {
		return stacktrace.Propagate(err, "Failed to delete daemonSet with name %s", name)
	}

	return nil
}

func (manager K8sManager) GetDaemonSet(context context.Context, name string) (*appsv1.DaemonSet, error) {

	daemonSetClient := manager.k8sClientSet.AppsV1().DaemonSets(apiv1.NamespaceDefault)

	daemonSet, err := daemonSetClient.Get(context, name, metav1.GetOptions{})

	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get daemonSet with name %s", name)
	}

	return daemonSet, nil
}

// Private functions
func (manager K8sManager) int32Ptr(i int32) *int32 { return &i }
