package grafloki

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"time"
)

const (
	grafanaServiceName             = "kurtosis-grafana-service"
	lokiServiceName                = "kurtosis-loki-service"
	grafanaDeploymentName          = "kurtosis-grafana-deployment"
	lokiDeploymentName             = "kurtosis-loki-deployment"
	grafanaDatasourceConfigMapName = "kurtosis-grafana-datasources"
	graflokiNamespace              = "kurtosis-grafloki"
)

var lokiLabels = map[string]string{
	lokiDeploymentName: "true",
}
var grafanaLabels = map[string]string{
	grafanaDeploymentName: "true",
}

var httpApplicationProtocol = "http"

func StartGrafLokiInKubernetes(ctx context.Context) (string, string, error) {
	k8sManager, err := getKubernetesManager()
	if err != nil {
		return "", "", stacktrace.Propagate(err, "An error occurred getting Kubernetes Manager.")
	}

	var lokiHost string
	doesGrafanaAndLokiExist, lokiHost := existsGrafanaAndLokiDeployments(ctx, k8sManager)
	if err != nil {
		return "", "", stacktrace.Propagate(err, "An error occurred checking if Grafana and Loki exist.")
	}

	if !doesGrafanaAndLokiExist {
		lokiHost, err = createGrafanaAndLokiDeployments(ctx, k8sManager)
		if err != nil {
			return "", "", stacktrace.Propagate(err, "An error occurred creating Grafana and Loki deployments.")
		}
	}

	logrus.Infof("Run `kubectl port-forward -n %v svc/grafana %v:%v` to access Grafana service.", graflokiNamespace, grafanaPort, grafanaPort)
	return lokiHost, getGrafanaUrl(grafanaPort), nil
}

func createGrafanaAndLokiDeployments(ctx context.Context, k8sManager *kubernetes_manager.KubernetesManager) (string, error) {
	_, err := k8sManager.CreateNamespace(ctx, graflokiNamespace, map[string]string{}, map[string]string{})
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating graflokiNamespace '%v'", graflokiNamespace)
	}

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
				ResizePolicy:             nil,
				VolumeMounts:             nil,
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
		[]apiv1.Volume{},
		&apiv1.Affinity{
			NodeAffinity:    nil,
			PodAffinity:     nil,
			PodAntiAffinity: nil,
		})
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating Loki deployment.")
	}
	shouldRemoveLokiDeployment := true
	defer func() {
		if shouldRemoveLokiDeployment {
			if err := k8sManager.RemoveDeployment(ctx, graflokiNamespace, lokiDeployment); err != nil {
				logrus.Warnf("Attempted to remove Loki deployment after an error occurred creating it but an error occurred removing it.")
				logrus.Warnf("Manually remove Loki deployment with name: %v", lokiDeployment.Name)
			}
		}
	}()
	lokiService, err := k8sManager.CreateService(ctx,
		graflokiNamespace,
		lokiServiceName,
		map[string]string{}, // empty labels
		nil,                 // empty annotations
		lokiLabels,          // match loki deployment pod labels
		apiv1.ServiceTypeNodePort,
		[]apiv1.ServicePort{{
			Name:        "logs-listening",
			Port:        lokiPort,
			TargetPort:  intstr.FromInt(lokiPort),
			Protocol:    apiv1.ProtocolTCP,
			NodePort:    lokiPort,
			AppProtocol: &httpApplicationProtocol,
		}})
	shouldRemoveLokiService := true
	defer func() {
		if shouldRemoveLokiService {
			if err := k8sManager.RemoveService(ctx, lokiService); err != nil {
				logrus.Warnf("Attempted to remove Loki service after an error occurred creating it but an error occurred removing it.")
				logrus.Warnf("Manually remove Loki service with name: %v", lokiService.Name)
			}
		}
	}()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating Loki service")
	}
	lokiHost := getLokiHost(lokiServiceName, graflokiNamespace, lokiPort)

	configMapData := map[string]string{
		"loki-datasource.yaml": fmt.Sprintf(`apiVersion: 1
kind: Datasource
metadata:
  name: Loki
spec:
  name: Loki
  type: loki
  access: proxy
  url: %v
  isDefault: true
  jsonData:
    timeInterval: 1s`, lokiHost),
	}
	grafanaConfigMap, err := k8sManager.CreateConfigMap(ctx,
		graflokiNamespace,
		grafanaDatasourceConfigMapName,
		map[string]string{},
		map[string]string{},
		configMapData)
	shouldRemoveGrafanaConfigMap := true
	defer func() {
		if shouldRemoveGrafanaConfigMap {
			if err := k8sManager.RemoveConfigMap(ctx, graflokiNamespace, grafanaConfigMap); err != nil {
				logrus.Warnf("Attempted to remove Grafana datasource config map after an error occurred creating it but an error occurred removing it.")
				logrus.Warnf("Manually remove Grafana datasource config map with name: %v", grafanaConfigMap.Name)
			}
		}
	}()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating Grafana datasource configmap.")
	}

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
						Name:      "GF_AUTH_ANONYMOUS_ENABLED",
						Value:     "true",
						ValueFrom: nil,
					},
					{
						Name:      "GF_AUTH_ANONYMOUS_ORG_ROLE",
						Value:     "Admin",
						ValueFrom: nil,
					},
					{
						Name:      "GF_SECURITY_ALLOW_EMBEDDING",
						Value:     "true",
						ValueFrom: nil,
					},
				},
				VolumeMounts: []apiv1.VolumeMount{
					{
						Name:             "datasources",
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
			Name: "datasources",
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
	shouldRemoveGrafanaDeployment := true
	defer func() {
		if shouldRemoveGrafanaDeployment {
			if err := k8sManager.RemoveDeployment(ctx, graflokiNamespace, grafanaDeployment); err != nil {
				logrus.Warnf("Attempted to remove Loki deployment after an error occurred creating it but an error occurred removing it.")
				logrus.Warnf("Manually remove Loki deployment with name: %v", lokiDeployment.Name)
			}
		}
	}()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating Grafana deployment.")
	}

	grafanaService, err := k8sManager.CreateService(ctx,
		graflokiNamespace,
		grafanaServiceName,
		map[string]string{}, // empty labels
		nil,                 // empty annotations
		grafanaLabels,       // match grafana deployment pod labels
		apiv1.ServiceTypeNodePort,
		[]apiv1.ServicePort{{
			Name:        "grafana-dashboard",
			Port:        grafanaPort,
			TargetPort:  intstr.FromInt(grafanaPort),
			Protocol:    apiv1.ProtocolTCP,
			NodePort:    grafanaPort,
			AppProtocol: &httpApplicationProtocol,
		}})
	shouldRemoveGrafanaService := true
	defer func() {
		if shouldRemoveGrafanaService {
			if err := k8sManager.RemoveService(ctx, grafanaService); err != nil {
				logrus.Warnf("Attempted to remove Grafana service after an error occurred creating it but an error occurred removing it.")
				logrus.Warnf("Manually remove Grafana service with name: %v", grafanaService.Name)
			}
		}
	}()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating Grafana service")
	}

	shouldRemoveLokiDeployment = false
	shouldRemoveGrafanaConfigMap = false
	shouldRemoveGrafanaDeployment = false
	shouldRemoveGrafanaService = false
	return lokiHost, nil
}

func existsGrafanaAndLokiDeployments(ctx context.Context, k8sManager *kubernetes_manager.KubernetesManager) (bool, string) {
	existsLoki := false
	existsGrafana := false
	var lokiHost string
	lokiDeployment, err := k8sManager.GetDeployment(ctx, graflokiNamespace, lokiDeploymentName)
	if err == nil && lokiDeployment != nil {
		existsLoki = true
		lokiHost = getLokiHost(lokiDeploymentName, graflokiNamespace, lokiPort)
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
	ns, err := k8sManager.GetNamespace(ctx, graflokiNamespace)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting graflokiNamespace '%v'.", graflokiNamespace)
	}
	// TODO: add a wait for graflokiNamespace removal
	err = k8sManager.RemoveNamespace(ctx, ns)
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
		clientcmd.NewDefaultClientConfigLoadingRules(), nil,
	).ClientConfig()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating Kubernetes configuration")
	}
	clientSet, err := kubernetes.NewForConfig(kubernetesConfig)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to create Kubernetes client set using Kubernetes config '%+v', instead a non nil error was returned", kubernetesConfig)
	}
	k8sManager := kubernetes_manager.NewKubernetesManager(clientSet, kubernetesConfig, "")
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

	return stacktrace.NewError("Attempted to wait for graflokiNamespace '%v' removal or to be marked for deletion '%v' times but graflokiNamespace was not removed.", namespace, maxTriesToWaitForNamespaceRemoval)
}

func getLokiHost(lokiServiceName, namespace string, lokiPort int) string {
	return fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", lokiServiceName, namespace, lokiPort)
}

func getGrafanaUrl(grafanaPort int) string {
	return fmt.Sprintf("https://localhost:%v", grafanaPort)
}
