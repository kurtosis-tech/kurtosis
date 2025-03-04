package fluentbit

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"time"
)

const (
	httpProtocolStr = "http"
	emptyUrl        = ""
	retryInterval   = 1 * time.Second
	maxRetries      = 30
)

var noWait *port_spec.Wait = nil

type fluentbitLogsCollector struct{}

func NewFluentbitLogsCollector() *fluentbitLogsCollector {
	return &fluentbitLogsCollector{}
}

func (fluentbit *fluentbitLogsCollector) CreateAndStart(
	ctx context.Context,
	logsAggregatorHost string,
	logsAggregatorPort uint16,
	tcpPortNumber uint16,
	httpPortNumber uint16,
	logsCollectorTcpPortId string,
	logsCollectorHttpPortId string,
	objAttrsProvider object_attributes_provider.KubernetesObjectAttributesProvider,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (
	*appsv1.DaemonSet,
	*apiv1.ConfigMap,
	*apiv1.Namespace,
	*apiv1.ServiceAccount,
	*rbacv1.ClusterRole,
	*rbacv1.ClusterRoleBinding,
	func(),
	error,
) {
	logsCollectorGuidStr, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, stacktrace.Propagate(err, "An error occurred creating uuid for logs collector.")
	}

	logsCollectorGuid := logs_collector.LogsCollectorGuid(logsCollectorGuidStr)
	logsCollectorAttrProvider := objAttrsProvider.ForLogsCollector(logsCollectorGuid)

	namespace, err := createLogsCollectorNamespace(ctx, logsCollectorAttrProvider, kubernetesManager)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, stacktrace.Propagate(err, "An error occurred creating namespace for logs collector.")
	}
	removeNamespaceFunc := func() {
		removeCtx := context.Background()
		if err := kubernetesManager.RemoveNamespace(removeCtx, namespace); err != nil {
			logrus.Errorf(
				"Launching the logs collector daemon set with name '%v' didn't complete successfully so we "+
					"tried to remove the namespace we started, but doing so exited with an error:\n%v",
				namespace.Name,
				err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually remove the logs collector namespace with Kubernetes name '%v'!!!!!!", namespace.Name)
		}
	}
	shouldRemoveLogsCollectorNamespace := true
	defer func() {
		if shouldRemoveLogsCollectorNamespace {
			removeNamespaceFunc()
		}
	}()

	serviceAccount, err := createLogsCollectorServiceAccount(ctx, namespace.Name, logsCollectorAttrProvider, kubernetesManager)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, stacktrace.Propagate(err, "An error occurred while trying to create service account for fluent bit log collector.")
	}
	removeServiceAccountFunc := func() {
		removeCtx := context.Background()
		if err := kubernetesManager.RemoveServiceAccount(removeCtx, serviceAccount); err != nil {
			logrus.Errorf(
				"Launching the logs collector daemon set with name '%v' didn't complete successfully so we "+
					"tried to remove the service account we started, but doing so exited with an error:\n%v",
				serviceAccount.Name,
				err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually remove the logs collector service account with Kubernetes name '%v' in namespace '%v'!!!!!!", serviceAccount.Name, serviceAccount.Namespace)
		}
	}
	shouldRemoveLogsCollectorServiceAccount := true
	defer func() {
		if shouldRemoveLogsCollectorServiceAccount {
			removeServiceAccountFunc()
		}
	}()

	clusterRole, err := createLogsCollectorClusterRole(ctx, logsCollectorAttrProvider, kubernetesManager)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, stacktrace.Propagate(err, "An error occurred while trying to create cluster role for fluent bit log collector.")
	}
	removeClusterRoleFunc := func() {
		removeCtx := context.Background()
		if err := kubernetesManager.RemoveClusterRole(removeCtx, clusterRole); err != nil {
			logrus.Errorf(
				"Launching the logs collector daemon set with name '%v' didn't complete successfully so we "+
					"tried to remove the cluster role we started, but doing so exited with an error:\n%v",
				clusterRole.Name,
				err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually remove the logs collector cluster role with Kubernetes name '%v' in namespace '%v'!!!!!!", clusterRole.Name, clusterRole.Namespace)
		}
	}
	shouldRemoveLogsCollectorClusterRole := true
	defer func() {
		if shouldRemoveLogsCollectorClusterRole {
			removeClusterRoleFunc()
		}
	}()

	clusterRoleBinding, err := createLogsCollectorClusterRoleBinding(ctx, serviceAccount.Name, clusterRole.Name, namespace.Name, logsCollectorAttrProvider, kubernetesManager)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, stacktrace.Propagate(err, "An error occurred while trying to create cluster role binding for fluent bit log collector.")
	}
	removeClusterRoleBindingFunc := func() {
		removeCtx := context.Background()
		if err := kubernetesManager.RemoveClusterRoleBindings(removeCtx, clusterRoleBinding); err != nil {
			logrus.Errorf(
				"Launching the logs collector daemon set with name '%v' didn't complete successfully so we "+
					"tried to remove the cluster role binding we started, but doing so exited with an error:\n%v",
				clusterRoleBinding.Name,
				err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually remove the logs collector cluster role binding with Kubernetes name '%v' in namespace '%v'!!!!!!", clusterRoleBinding.Name, clusterRoleBinding.Namespace)
		}
	}
	shouldRemoveLogsCollectorClusterRoleBinding := true
	defer func() {
		if shouldRemoveLogsCollectorClusterRoleBinding {
			removeClusterRoleBindingFunc()
		}
	}()

	configMap, err := createLogsCollectorConfigMap(ctx, namespace.Name, logsCollectorAttrProvider, kubernetesManager)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, stacktrace.Propagate(err, "An error occurred while trying to create config map for fluent bit log collector.")
	}
	removeConfigMapFunc := func() {
		removeCtx := context.Background()
		if err := kubernetesManager.RemoveConfigMap(removeCtx, namespace.Name, configMap); err != nil {
			logrus.Errorf(
				"Launching the logs collector daemon set with name '%v' didn't complete successfully so we "+
					"tried to remove the config map we started, but doing so exited with an error:\n%v",
				configMap.Name,
				err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually remove the logs collector config map with Kubernetes name '%v' in namespace '%v'!!!!!!", configMap.Name, configMap.Namespace)
		}
	}
	shouldRemoveLogsCollectorConfigMap := true
	defer func() {
		if shouldRemoveLogsCollectorConfigMap {
			removeConfigMapFunc()
		}
	}()

	httpPortSpec, err := port_spec.NewPortSpec(httpPortNumber, port_spec.TransportProtocol_TCP, httpProtocolStr, noWait, emptyUrl)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, stacktrace.Propagate(
			err,
			"An error occurred creating the log collectors HTTP port spec object using number '%v' and protocol '%v'",
			httpPortNumber,
			httpProtocolStr,
		)
	}
	tcpPortSpec, err := port_spec.NewPortSpec(tcpPortNumber, port_spec.TransportProtocol_TCP, httpProtocolStr, noWait, emptyUrl)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, stacktrace.Propagate(
			err,
			"An error occurred creating the log collectors TCP port spec object using number '%v' and protocol '%v'",
			tcpPortNumber,
			port_spec.TransportProtocol_TCP,
		)
	}
	privatePorts := map[string]*port_spec.PortSpec{
		logsCollectorTcpPortId:  tcpPortSpec,
		logsCollectorHttpPortId: httpPortSpec,
	}

	containerPorts, err := shared_helpers.GetKubernetesContainerPortsFromPrivatePortSpecs(privatePorts)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, stacktrace.Propagate(err, "An error occurred getting the logs collector fluent bit container ports from the port specs")
	}

	daemonSet, err := createLogsCollectorDaemonSet(ctx, namespace.Name, configMap.Name, serviceAccount.Name, containerPorts, logsCollectorAttrProvider, kubernetesManager)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, stacktrace.Propagate(err, "An error occurred while trying to create daemon set for fluent bit logs collector.")
	}
	removeDaemonSetFunc := func() {
		removeCtx := context.Background()
		if err := kubernetesManager.RemoveDaemonSet(removeCtx, namespace.Name, daemonSet); err != nil {
			logrus.Errorf(
				"Launching the logs collector daemon with name '%v' didn't complete successfully so we "+
					"tried to remove the daemon set we started, but doing so exited with an error:\n%v",
				daemonSet.Name,
				err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually remove the logs collector daemon set with Kubernetes name '%v' in namespace '%v'!!!!!!", daemonSet.Name, daemonSet.Namespace)
		}
	}
	shouldRemoveLogsCollectorDaemonSet := true
	defer func() {
		if shouldRemoveLogsCollectorDaemonSet {
			removeDaemonSetFunc()
		}
	}()

	// wait until the first pod associated with this daemon set is online before returning
	if err = waitForAtLeastOneActivePodManagedByDaemonSet(ctx, daemonSet, kubernetesManager); err != nil {
		return nil, nil, nil, nil, nil, nil, nil, stacktrace.Propagate(err, "An error occurred waiting for at least one active pod managed by logs collector daemon set '%v'", daemonSet.Name)
	}

	removeLogsCollectorFunc := func() {
		removeDaemonSetFunc()
		removeConfigMapFunc()
		removeClusterRoleBindingFunc()
		removeClusterRoleFunc()
		removeServiceAccountFunc()
		removeNamespaceFunc()
	}

	shouldRemoveLogsCollectorClusterRoleBinding = false
	shouldRemoveLogsCollectorClusterRole = false
	shouldRemoveLogsCollectorServiceAccount = false
	shouldRemoveLogsCollectorNamespace = false
	shouldRemoveLogsCollectorConfigMap = false
	shouldRemoveLogsCollectorDaemonSet = false
	return daemonSet, configMap, namespace, serviceAccount, clusterRole, clusterRoleBinding, removeLogsCollectorFunc, nil
}

func (fluentbit *fluentbitLogsCollector) GetHttpHealthCheckEndpoint() string {
	return healthCheckEndpointPath
}

func createLogsCollectorDaemonSet(
	ctx context.Context,
	namespace string,
	fluentBitCfgConfigMapName string,
	serviceAccountName string,
	ports []apiv1.ContainerPort,
	objAttrProvider object_attributes_provider.KubernetesLogsCollectorObjectAttributesProvider,
	kubernetesManager *kubernetes_manager.KubernetesManager) (*appsv1.DaemonSet, error) {

	daemonSetAttrProvider, err := objAttrProvider.ForLogsCollectorDaemonSet()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting logs collector daemon set attributes provider.")
	}
	name := daemonSetAttrProvider.GetName().GetString()
	labels := shared_helpers.GetStringMapFromLabelMap(daemonSetAttrProvider.GetLabels())
	annotations := shared_helpers.GetStringMapFromAnnotationMap(daemonSetAttrProvider.GetAnnotations())

	containers := []apiv1.Container{
		{
			Name:  fluentBitContainerName,
			Image: fluentBitImage,
			// uses values from official fluent bit helm chart:
			// https://github.com/fluent/helm-charts/blob/f87fec5d36cf871ad5174c021b8d6ceb6a1b2001/charts/fluent-bit/values.yaml#L489
			Command: []string{
				"/fluent-bit/bin/fluent-bit",
			},
			Args: []string{
				"--workdir=/fluent-bit/etc",
				"--config=/fluent-bit/etc/conf/fluent-bit.conf",
			},
			Ports:      ports,
			WorkingDir: "",
			EnvFrom:    nil,
			Env:        nil,
			Resources: apiv1.ResourceRequirements{
				Limits:   nil,
				Requests: nil,
				Claims:   nil,
			},
			ResizePolicy:   nil,
			VolumeDevices:  nil,
			LivenessProbe:  nil,
			ReadinessProbe: nil,
			StartupProbe:   nil,
			// Clean up job to remove the fluent bit checkpoint dbs when they are stopped
			// Note: only runs if container is shut down gracefully, therefore will not remove the checkpoint db if the pod crashes
			// TODO: find a way to remove fluent bit checkpoint db during kurtosis clean -a instead of kurtosis engine stop
			Lifecycle: &apiv1.Lifecycle{
				PostStart: nil,
				PreStop: &apiv1.LifecycleHandler{
					Exec: &apiv1.ExecAction{
						Command: []string{"sh", "-c", fmt.Sprintf("rm -rf %v/*", fluentBitCheckpointDbMountPath)},
					},
					HTTPGet:   nil,
					TCPSocket: nil,
				},
			},
			TerminationMessagePolicy: "",
			ImagePullPolicy:          "",
			SecurityContext:          nil,
			Stdin:                    false,
			StdinOnce:                false,
			TTY:                      false,
			TerminationMessagePath:   "",
			VolumeMounts: []apiv1.VolumeMount{
				// these volumes are where logs from pods on a node in a k8s cluster get stored
				// they get mounted to the fluentbit pod so the fluentbit pod can read them via the `tail` input plugin
				// https://docs.fluentbit.io/manual/installation/kubernetes#details
				//
				{
					Name:             varLogVolumeName,
					ReadOnly:         false,
					MountPath:        varLogMountPath,
					SubPath:          "",
					MountPropagation: nil,
					SubPathExpr:      "",
				},
				{
					Name:             varLibDockerContainersVolumeName,
					ReadOnly:         false,
					MountPath:        varLibDockerContainersMountPath,
					SubPath:          "",
					MountPropagation: nil,
					SubPathExpr:      "",
				},
				{
					Name:             varLogDockerContainersVolumeName,
					ReadOnly:         false,
					MountPath:        varLogDockerContainersMountPath,
					SubPath:          "",
					MountPropagation: nil,
					SubPathExpr:      "",
				},
				{
					Name:             fluentBitConfigVolumeName,
					ReadOnly:         false,
					MountPath:        fluentBitConfigMountPath,
					SubPath:          "",
					MountPropagation: nil,
					SubPathExpr:      "",
				},
				{
					Name:             fluentBitHostLogsVolumeName,
					ReadOnly:         false,
					MountPath:        fluentBitHostLogsMountPath,
					SubPath:          "",
					MountPropagation: nil,
					SubPathExpr:      "",
				},
				{
					Name:             fluentBitCheckpointDbVolumeName,
					ReadOnly:         false,
					MountPath:        fluentBitCheckpointDbMountPath,
					SubPath:          "",
					MountPropagation: nil,
					SubPathExpr:      "",
				},
			},
		},
	}

	volumes := []apiv1.Volume{
		{
			Name:         varLogVolumeName,
			VolumeSource: getVolumeSourceForHostPath(varLogMountPath),
		},
		{
			Name:         varLibDockerContainersVolumeName,
			VolumeSource: getVolumeSourceForHostPath(varLibDockerContainersMountPath),
		},
		{
			Name:         varLogDockerContainersVolumeName,
			VolumeSource: getVolumeSourceForHostPath(varLogDockerContainersMountPath),
		},
		{
			Name:         fluentBitConfigVolumeName,
			VolumeSource: getVolumeSourceForConfigMap(fluentBitCfgConfigMapName),
		},
		{
			Name:         fluentBitHostLogsVolumeName,
			VolumeSource: getVolumeSourceForHostPath(fluentBitHostLogsMountPath),
		},
		{
			Name:         fluentBitCheckpointDbVolumeName,
			VolumeSource: getVolumeSourceForHostPath(fluentBitCheckpointDbMountPath),
		},
	}

	logsCollectorDaemonSet, err := kubernetesManager.CreateDaemonSet(
		ctx,
		namespace,
		name,
		labels,
		annotations,
		serviceAccountName,
		[]apiv1.Container{}, // no need init containers
		containers,
		volumes,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating daemon set for fluent bit logs collector.")
	}

	return logsCollectorDaemonSet, nil
}

func createLogsCollectorConfigMap(
	ctx context.Context,
	namespace string,
	objAttrProvider object_attributes_provider.KubernetesLogsCollectorObjectAttributesProvider,
	kubernetesManager *kubernetes_manager.KubernetesManager) (*apiv1.ConfigMap, error) {
	configMapAttrProvider, err := objAttrProvider.ForLogsCollectorConfigMap()
	if err != nil {
		return nil, err
	}
	name := configMapAttrProvider.GetName().GetString()
	labels := shared_helpers.GetStringMapFromLabelMap(configMapAttrProvider.GetLabels())
	annotations := shared_helpers.GetStringMapFromAnnotationMap(configMapAttrProvider.GetAnnotations())

	configMap, err := kubernetesManager.CreateConfigMap(
		ctx,
		namespace,
		name,
		labels,
		annotations,
		map[string]string{
			fluentBitConfigFileName: fluentBitConfigStr,
		},
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating config map for fluent bit log collector config.")
	}

	return configMap, nil
}

func createLogsCollectorNamespace(
	ctx context.Context,
	objAttrProvider object_attributes_provider.KubernetesLogsCollectorObjectAttributesProvider,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (*apiv1.Namespace, error) {
	namespaceAttrProvider, err := objAttrProvider.ForLogsCollectorNamespace()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting logs collector namespace attributes provider.")
	}
	namespaceName := namespaceAttrProvider.GetName().GetString()
	namespaceLabels := shared_helpers.GetStringMapFromLabelMap(namespaceAttrProvider.GetLabels())
	namespaceAnnotations := shared_helpers.GetStringMapFromAnnotationMap(namespaceAttrProvider.GetAnnotations())

	namespaceObj, err := kubernetesManager.CreateNamespace(ctx, namespaceName, namespaceLabels, namespaceAnnotations)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating namepsace for logs collector with name '%s'", namespaceName)
	}

	return namespaceObj, nil

}

func waitForAtLeastOneActivePodManagedByDaemonSet(ctx context.Context, logsCollectorDaemonSet *appsv1.DaemonSet, kubernetesManager *kubernetes_manager.KubernetesManager) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(maxRetries)*retryInterval)
	defer cancel()

	ticker := time.NewTicker(retryInterval)
	defer ticker.Stop()
	for attempt := 0; attempt < maxRetries; attempt++ {
		select {
		case <-timeoutCtx.Done():
			return stacktrace.NewError(
				"Timeout waiting for a pod managed by logs collector daemon set '%s' to come online",
				logsCollectorDaemonSet.Name,
			)
		case <-ticker.C:
			pods, err := kubernetesManager.GetPodsManagedByDaemonSet(ctx, logsCollectorDaemonSet)
			if err != nil {
				return stacktrace.Propagate(err, "An error occurred getting pods managed by logs collector daemon set '%v'", logsCollectorDaemonSet.Name)
			}
			if len(pods) > 0 && len(pods[0].Status.ContainerStatuses) > 0 && pods[0].Status.ContainerStatuses[0].Ready {
				// found a pod with a running fluent bit container
				return nil
			}
		}
	}
	return stacktrace.NewError(
		"Exceeded max retries (%d) waiting for a pod managed by daemon set '%s' to come online",
		maxRetries, logsCollectorDaemonSet.Name,
	)
}

func getVolumeSourceForHostPath(mountPath string) apiv1.VolumeSource {
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

func getVolumeSourceForConfigMap(configMapName string) apiv1.VolumeSource {
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

func createLogsCollectorServiceAccount(
	ctx context.Context,
	namespace string,
	objAttrProvider object_attributes_provider.KubernetesLogsCollectorObjectAttributesProvider,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (*apiv1.ServiceAccount, error) {
	serviceAccountAttrProvider, err := objAttrProvider.ForLogsCollectorServiceAccount()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting logs collector service account attributes provider.")
	}
	serviceAccountName := serviceAccountAttrProvider.GetName().GetString()
	serviceAccountLabels := shared_helpers.GetStringMapFromLabelMap(serviceAccountAttrProvider.GetLabels())

	serviceAccountObj, err := kubernetesManager.CreateServiceAccount(ctx, serviceAccountName, namespace, serviceAccountLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating service account for logs collector with name '%s'", serviceAccountName)
	}

	return serviceAccountObj, nil
}

func createLogsCollectorClusterRole(
	ctx context.Context,
	objAttrProvider object_attributes_provider.KubernetesLogsCollectorObjectAttributesProvider,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (*rbacv1.ClusterRole, error) {
	clusterRoleAttrProvider, err := objAttrProvider.ForLogsCollectorClusterRole()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting logs collector cluster role attributes provider.")
	}
	clusterRoleName := clusterRoleAttrProvider.GetName().GetString()
	clusterRoleLabels := shared_helpers.GetStringMapFromLabelMap(clusterRoleAttrProvider.GetLabels())

	rules := []rbacv1.PolicyRule{
		{
			Verbs:           []string{"get", "list"},
			APIGroups:       []string{""},
			Resources:       []string{"pods", "pods/logs"},
			ResourceNames:   nil,
			NonResourceURLs: nil,
		},
	}
	clusterRoleObj, err := kubernetesManager.CreateClusterRoles(ctx, clusterRoleName, rules, clusterRoleLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating cluster role for logs collector with name '%s'", clusterRoleName)
	}

	return clusterRoleObj, nil
}

func createLogsCollectorClusterRoleBinding(
	ctx context.Context,
	serviceAccountName string,
	clusterRoleName string,
	namespaceName string,
	objAttrProvider object_attributes_provider.KubernetesLogsCollectorObjectAttributesProvider,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (*rbacv1.ClusterRoleBinding, error) {
	clusterRoleBindingAttrProvider, err := objAttrProvider.ForLogsCollectorClusterRoleBinding()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting logs collector cluster role binding attributes provider.")
	}
	clusterRoleBindingName := clusterRoleBindingAttrProvider.GetName().GetString()
	clusterRoleBindingLabels := shared_helpers.GetStringMapFromLabelMap(clusterRoleBindingAttrProvider.GetLabels())

	subject := []rbacv1.Subject{
		{
			Kind:      "ServiceAccount",
			Name:      serviceAccountName,
			Namespace: namespaceName,
			APIGroup:  "",
		},
	}
	ref := rbacv1.RoleRef{
		Kind:     "ClusterRole",
		Name:     clusterRoleName,
		APIGroup: "rbac.authorization.k8s.io",
	}
	clusterRoleBindingObj, err := kubernetesManager.CreateClusterRoleBindings(ctx, clusterRoleBindingName, subject, ref, clusterRoleBindingLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating cluster role binding for logs collector with name '%s'", clusterRoleBindingName)
	}

	return clusterRoleBindingObj, nil
}
