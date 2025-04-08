package grafloki

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"time"
)

const (
	grafanaServiceName                   = "kurtosis-grafana-service"
	lokiServiceName                      = "kurtosis-loki-service"
	grafanaDeploymentName                = "kurtosis-grafana-deployment"
	lokiDeploymentName                   = "kurtosis-loki-deployment"
	grafanaDatasourceConfigMapName       = "kurtosis-grafana-datasources"
	graflokiNamespace                    = "kurtosis-grafloki"
	grafanaNodePort                int32 = 30030
	lokiNodePort                   int32 = 30031
	lokiProbeInitialDelaySeconds         = 5
	lokiProbePeriodSeconds               = 10
	lokiProbeTimeoutSeconds              = 10

	// takes around 30 seconds for loki pod to become ready
	lokiDeploymentMaxRetries    = 40
	lokiDeploymentRetryInterval = 1 * time.Second
	defaultStorageClass         = ""
)

var lokiLabels = map[string]string{
	kubernetes_label_key.KurtosisResourceTypeKubernetesLabelKey.GetString(): lokiDeploymentName,
}
var grafanaLabels = map[string]string{
	kubernetes_label_key.KurtosisResourceTypeKubernetesLabelKey.GetString(): grafanaDeploymentName,
}

var httpApplicationProtocol = "http"

func StartGrafLokiInKubernetes(ctx context.Context) (string, string, error) {
	k8sManager, err := getKubernetesManager()
	if err != nil {
		return "", "", stacktrace.Propagate(err, "An error occurred getting Kubernetes Manager.")
	}

	var lokiHost string
	var removeGrafanaAndLokiFunc func()
	shouldRemoveGrafanaAndLoki := false
	doesGrafanaAndLokiExist, lokiHost := checkGrafanaAndLokiDeploymentExistence(ctx, k8sManager)
	if !doesGrafanaAndLokiExist {
		lokiHost, removeGrafanaAndLokiFunc, err = createGrafanaAndLokiDeployments(ctx, k8sManager)
		if err != nil {
			return "", "", stacktrace.Propagate(err, "An error occurred creating Grafana and Loki deployments.")
		}
		shouldRemoveGrafanaAndLoki = true
		defer func() {
			if shouldRemoveGrafanaAndLoki {
				removeGrafanaAndLokiFunc()
			}
		}()
	}

	logrus.Infof("Run `kubectl port-forward -n %v svc/%v %v:%v` to access Grafana service.", graflokiNamespace, grafanaServiceName, grafanaPort, grafanaNodePort)
	shouldRemoveGrafanaAndLoki = false
	return lokiHost, getGrafanaUrl(grafanaPort), nil
}

func createGrafanaAndLokiDeployments(ctx context.Context, k8sManager *kubernetes_manager.KubernetesManager) (string, func(), error) {
	graflokilNamespaceObj, err := k8sManager.CreateNamespace(ctx, graflokiNamespace, map[string]string{}, map[string]string{})
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred creating namespace '%v'", graflokiNamespace)
	}
	shouldRemoveNamespace := false
	removeGraflokiNamespaceFunc := func() {
		if err := k8sManager.RemoveNamespace(ctx, graflokilNamespaceObj); err != nil {
			logrus.Warnf("Attempted to remove namespace '%v' after an error occurred but an error occurred removing it.", graflokiNamespace)
			logrus.Warnf("!! ACTION REQUIRED !! Manually remove namespace with name: %v", graflokilNamespaceObj.Name)
		}
	}
	defer func() {
		if shouldRemoveNamespace {
			removeGraflokiNamespaceFunc()
		}
	}()

	lokiDeployment, err := k8sManager.CreateDeployment(
		ctx,
		graflokiNamespace,
		lokiDeploymentName,
		lokiLabels,
		map[string]string{},
		[]apiv1.Container{}, // no init containers
		[]apiv1.Container{
			{
				Name:  "loki",
				Image: lokiImage,
				Ports: []apiv1.ContainerPort{
					{
						Name:          "",
						HostPort:      0,
						ContainerPort: lokiPort,
						Protocol:      "",
						HostIP:        "",
					},
				},
				Command:    nil,
				Args:       nil,
				WorkingDir: "",
				EnvFrom:    nil,
				Env:        nil,
				Resources: apiv1.ResourceRequirements{
					Limits:   nil,
					Requests: nil,
					Claims:   nil,
				},
				ResizePolicy:  nil,
				VolumeMounts:  nil,
				VolumeDevices: nil,
				LivenessProbe: nil,
				ReadinessProbe: &apiv1.Probe{
					ProbeHandler: apiv1.ProbeHandler{
						Exec: nil,
						HTTPGet: &apiv1.HTTPGetAction{
							Path:        "/ready",
							Port:        intstr.FromInt(lokiPort),
							Host:        "",
							Scheme:      "",
							HTTPHeaders: nil,
						},
						TCPSocket: nil,
						GRPC:      nil,
					},
					InitialDelaySeconds:           lokiProbeInitialDelaySeconds,
					TimeoutSeconds:                lokiProbeTimeoutSeconds,
					PeriodSeconds:                 lokiProbePeriodSeconds,
					SuccessThreshold:              0,
					FailureThreshold:              0,
					TerminationGracePeriodSeconds: nil,
				},
				StartupProbe:             nil,
				Lifecycle:                nil,
				TerminationMessagePath:   "",
				TerminationMessagePolicy: "",
				ImagePullPolicy:          "",
				SecurityContext:          nil,
				Stdin:                    false,
				StdinOnce:                false,
				TTY:                      false,
			},
		},
		[]apiv1.Volume{},
		&apiv1.Affinity{
			NodeAffinity:    nil,
			PodAffinity:     nil,
			PodAntiAffinity: nil,
		})
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred creating Loki deployment.")
	}
	shouldRemoveLokiDeployment := false
	removeLokiDeploymentFunc := func() {
		if err := k8sManager.RemoveDeployment(ctx, graflokiNamespace, lokiDeployment); err != nil {
			logrus.Warnf("Attempted to remove Loki deployment after an error occurred but an error occurred removing it.")
			logrus.Warnf("!! ACTION REQUIRED !! Manually remove Loki deployment with name: %v", lokiDeployment.Name)
		}
	}
	defer func() {
		if shouldRemoveLokiDeployment {
			removeLokiDeploymentFunc()
		}
	}()
	logrus.Infof("Waiting for Loki deployment to come online (can take around 30s)... ")
	if err := k8sManager.WaitForPodManagedByDeployment(ctx, lokiDeployment, lokiDeploymentMaxRetries, lokiDeploymentRetryInterval); err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred while waiting for pod managed by Loki deployment '%v' to come online.", lokiDeploymentName)
	}

	lokiService, err := k8sManager.CreateService(ctx,
		graflokiNamespace,
		lokiServiceName,
		map[string]string{}, // empty labels
		map[string]string{}, // empty annotations
		lokiLabels,          // match loki deployment pod labels
		apiv1.ServiceTypeNodePort,
		[]apiv1.ServicePort{{
			Name:        "logs-listening",
			Port:        lokiNodePort,
			TargetPort:  intstr.FromInt(lokiPort),
			Protocol:    apiv1.ProtocolTCP,
			NodePort:    lokiNodePort,
			AppProtocol: &httpApplicationProtocol,
		}})
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred creating Loki service")
	}
	shouldRemoveLokiService := false
	removeLokiServiceFunc := func() {
		if err := k8sManager.RemoveService(ctx, lokiService); err != nil {
			logrus.Warnf("Attempted to remove Loki service after an error occurred but an error occurred removing it.")
			logrus.Warnf("!! ACTION REQUIRED !! Manually remove Loki service with name: %v", lokiService.Name)
		}
	}
	defer func() {
		if shouldRemoveLokiService {
			removeLokiServiceFunc()
		}
	}()
	lokiHost := getLokiHost(lokiServiceName, graflokiNamespace, lokiNodePort)

	grafanaDatasource := GrafanaDatasources{
		apiVersion: "1",
		datasources: []GrafanaDatasource{
			{
				name:      lokiServiceName,
				type_:     "loki",
				access:    "proxy",
				url:       lokiHost,
				isDefault: true,
				editable:  true,
			},
		}}
	grafanaDatasourceYaml, err := yaml.Marshal(grafanaDatasource)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred serializing Grafana datasource to yaml: %v", grafanaDatasourceYaml)
	}

	configMapData := map[string]string{
		"loki-datasource.yaml": string(grafanaDatasourceYaml),
	}
	grafanaConfigMap, err := k8sManager.CreateConfigMap(ctx,
		graflokiNamespace,
		grafanaDatasourceConfigMapName,
		map[string]string{}, // empty labels
		map[string]string{}, // empty annotations
		configMapData)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred creating Grafana datasource configmap.")
	}
	shouldRemoveGrafanaConfigMap := true
	removeGrafanaConfigMapFunc := func() {
		if err := k8sManager.RemoveConfigMap(ctx, graflokiNamespace, grafanaConfigMap); err != nil {
			logrus.Warnf("Attempted to remove Grafana datasource config map after an error occurred but an error occurred removing it.")
			logrus.Warnf("!! ACTION REQUIRED !! Manually remove Grafana datasource config map with name: %v", grafanaConfigMap.Name)
		}
	}
	defer func() {
		if shouldRemoveGrafanaConfigMap {
			removeGrafanaConfigMapFunc()
		}
	}()

	grafanaDeployment, err := k8sManager.CreateDeployment(
		ctx,
		graflokiNamespace,
		grafanaDeploymentName,
		grafanaLabels,
		map[string]string{}, // empty annotations
		[]apiv1.Container{}, // no init containers
		[]apiv1.Container{
			{
				Name:  "grafana",
				Image: grafanaImage,
				Ports: []apiv1.ContainerPort{
					{
						Name:          "",
						ContainerPort: grafanaPort,
						HostPort:      0,
						Protocol:      "",
						HostIP:        "",
					},
				},
				Env: []apiv1.EnvVar{
					{
						Name:      grafanaAuthAnonymousEnabledEnvVarKey,
						Value:     grafanaAuthAnonymousEnabledEnvVarVal,
						ValueFrom: nil,
					},
					{
						Name:      grafanaAuthAnonymousOrgRoleEnvVarKey,
						Value:     grafanaAuthAnonymousOrgRoleEnvVarVal,
						ValueFrom: nil,
					},
					{
						Name:      grafanaSecurityAllowEmbeddingEnvVarKey,
						Value:     grafanaSecurityAllowEmbeddingEnvVarVal,
						ValueFrom: nil,
					},
				},
				VolumeMounts: []apiv1.VolumeMount{
					{
						Name:             grafanaDatasourcesKey,
						MountPath:        "/etc/grafana/provisioning/datasources",
						ReadOnly:         false,
						SubPath:          "",
						MountPropagation: nil,
						SubPathExpr:      "",
					},
				},
				Command:    nil,
				Args:       nil,
				WorkingDir: "",
				EnvFrom:    nil,
				Resources: apiv1.ResourceRequirements{
					Limits:   nil,
					Requests: nil,
					Claims:   nil,
				},
				ResizePolicy:             nil,
				VolumeDevices:            nil,
				LivenessProbe:            nil,
				ReadinessProbe:           nil,
				StartupProbe:             nil,
				Lifecycle:                nil,
				TerminationMessagePath:   "",
				TerminationMessagePolicy: "",
				ImagePullPolicy:          "",
				SecurityContext:          nil,
				Stdin:                    false,
				StdinOnce:                false,
				TTY:                      false,
			},
		},
		[]apiv1.Volume{{
			Name: grafanaDatasourcesKey,
			VolumeSource: apiv1.VolumeSource{
				ConfigMap: &apiv1.ConfigMapVolumeSource{
					LocalObjectReference: apiv1.LocalObjectReference{
						Name: grafanaDatasourceConfigMapName,
					},
					Items:       nil,
					DefaultMode: nil,
					Optional:    nil,
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
			},
		}},
		&apiv1.Affinity{
			NodeAffinity:    nil,
			PodAffinity:     nil,
			PodAntiAffinity: nil,
		})
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred creating Grafana deployment.")
	}
	shouldRemoveGrafanaDeployment := true
	removeGrafanaDeploymentFunc := func() {
		if err := k8sManager.RemoveDeployment(ctx, graflokiNamespace, grafanaDeployment); err != nil {
			logrus.Warnf("Attempted to remove Loki deployment after an error occurred but an error occurred removing it.")
			logrus.Warnf("!! ACTION REQUIRED !! Manually remove Loki deployment with name: %v", lokiDeployment.Name)
		}
	}
	defer func() {
		if shouldRemoveGrafanaDeployment {
			removeGrafanaDeploymentFunc()
		}
	}()

	grafanaService, err := k8sManager.CreateService(ctx,
		graflokiNamespace,
		grafanaServiceName,
		map[string]string{}, // empty labels
		nil,                 // empty annotations
		grafanaLabels,       // match grafana deployment pod labels
		apiv1.ServiceTypeNodePort,
		[]apiv1.ServicePort{{
			Name:        "grafana-dashboard",
			Port:        grafanaNodePort,
			TargetPort:  intstr.FromInt(grafanaPort),
			Protocol:    apiv1.ProtocolTCP,
			NodePort:    grafanaNodePort,
			AppProtocol: &httpApplicationProtocol,
		}})
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred creating Grafana service")
	}
	shouldRemoveGrafanaService := true
	removeGrafanaServiceFunc := func() {
		if err := k8sManager.RemoveService(ctx, grafanaService); err != nil {
			logrus.Warnf("Attempted to remove Grafana service after an error occurred but an error occurred removing it.")
			logrus.Warnf("!! ACTION REQUIRED !! Manually remove Grafana service with name: %v", grafanaService.Name)
		}
	}
	defer func() {
		if shouldRemoveGrafanaService {
			removeGrafanaServiceFunc()
		}
	}()

	removeGrafanaAndLokiDeploymentsFunc := func() {
		removeGraflokiNamespaceFunc()
		removeLokiDeploymentFunc()
		removeLokiServiceFunc()
		removeGrafanaConfigMapFunc()
		removeGrafanaDeploymentFunc()
		removeGrafanaServiceFunc()
	}

	shouldRemoveLokiDeployment = false
	shouldRemoveGrafanaConfigMap = false
	shouldRemoveGrafanaDeployment = false
	shouldRemoveGrafanaService = false
	shouldRemoveNamespace = false
	shouldRemoveLokiService = false
	return lokiHost, removeGrafanaAndLokiDeploymentsFunc, nil
}

func checkGrafanaAndLokiDeploymentExistence(ctx context.Context, k8sManager *kubernetes_manager.KubernetesManager) (bool, string) {
	existsLoki := false
	existsGrafana := false
	var lokiHost string
	lokiDeployment, err := k8sManager.GetDeployment(ctx, graflokiNamespace, lokiDeploymentName)
	if err == nil && lokiDeployment != nil {
		existsLoki = true
		lokiHost = getLokiHost(lokiServiceName, graflokiNamespace, lokiNodePort)
	}

	grafanaDeployment, err := k8sManager.GetDeployment(ctx, graflokiNamespace, grafanaDeploymentName)
	if err == nil && grafanaDeployment != nil {
		existsGrafana = true
	}

	return existsLoki && existsGrafana, lokiHost
}

func StopGrafLokiInKubernetes(ctx context.Context) error {
	k8sManager, err := getKubernetesManager()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting Kubernetes Manager.")
	}
	graflokiNamespaceObj, err := k8sManager.GetNamespace(ctx, graflokiNamespace)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting graflokiNamespace '%v'.", graflokiNamespace)
	}
	err = k8sManager.RemoveNamespace(ctx, graflokiNamespaceObj)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred removing graflokiNamespace '%v'.", graflokiNamespace)
	}
	err = waitForNamespaceRemoval(ctx, graflokiNamespace, k8sManager)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while waiting for graflokiNamespace '%v' removal.", graflokiNamespace)
	}
	return nil
}

func getKubernetesManager() (*kubernetes_manager.KubernetesManager, error) {
	kubernetesConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		nil, // empty overrides
	).ClientConfig()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating Kubernetes configuration")
	}
	clientSet, err := kubernetes.NewForConfig(kubernetesConfig)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to create Kubernetes client set using Kubernetes config '%+v', instead a non nil error was returned", kubernetesConfig)
	}
	k8sManager := kubernetes_manager.NewKubernetesManager(clientSet, kubernetesConfig, defaultStorageClass)
	return k8sManager, nil
}

func waitForNamespaceRemoval(
	ctx context.Context,
	namespace string,
	kubernetesManager *kubernetes_manager.KubernetesManager) error {
	var (
		maxTriesToWaitForNamespaceRemoval       uint = 30
		timeToWaitBetweenNamespaceRemovalChecks      = 1 * time.Second
	)

	for i := uint(0); i < maxTriesToWaitForNamespaceRemoval; i++ {
		if _, err := kubernetesManager.GetNamespace(ctx, namespace); err != nil {
			// if err was returned, graflokiNamespace doesn't exist, or it's been marked for deleted
			return nil
		}

		// Tiny optimization to not sleep if we're not going to run the loop again
		if i < maxTriesToWaitForNamespaceRemoval-1 {
			time.Sleep(timeToWaitBetweenNamespaceRemovalChecks)
		}
	}

	return stacktrace.NewError("Attempted to wait for namespace '%v' removal or to be marked for deletion '%v' times but '%v' was not removed.", namespace, maxTriesToWaitForNamespaceRemoval, namespace)
}

func getLokiHost(lokiServiceName, namespace string, lokiPort int32) string {
	return fmt.Sprintf("http://%v.%v.svc.cluster.local:%v", lokiServiceName, namespace, lokiPort)
}

func getGrafanaUrl(grafanaPort int) string {
	return fmt.Sprintf("https://127.0.0.1:%v", grafanaPort)
}
