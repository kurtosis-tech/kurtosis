package grafloki

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/stacktrace"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	grafanaServiceName             = "grafana"
	lokiServiceName                = "loki"
	grafanaDeploymentName          = "grafana-deployment"
	lokiDeploymentName             = "loki-deployment"
	grafanaDatasourceConfigMapName = "grafana-datasources"
	namespace                      = "kurtosis-grafloki"
)

func StartGrafLokiInKubernetes(ctx context.Context) (string, error) {
	kubernetesConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(), nil,
	).ClientConfig()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating Kubernetes configuration")
	}

	clientSet, err := kubernetes.NewForConfig(kubernetesConfig)
	if err != nil {
		return "", stacktrace.Propagate(err, "Expected to be able to create Kubernetes client set using Kubernetes config '%+v', instead a non nil error was returned", kubernetesConfig)
	}

	k8sManager := kubernetes_manager.NewKubernetesManager(clientSet, kubernetesConfig, "")

	if _, err := k8sManager.CreateDeployment(
		ctx,
		namespace,
		lokiDeploymentName,
		map[string]string{},
		map[string]string{},
		[]apiv1.Container{}, // no init containers
		[]apiv1.Container{{
			Name:  "loki",
			Image: grafanaImage,
			Ports: []apiv1.ContainerPort{{ContainerPort: lokiPort}},
		}},
		[]apiv1.Volume{},
		&apiv1.Affinity{}); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating Loki deployment.")
	}

	configMapData := map[string]string{
		"loki-datasource.yaml": fmt.Sprintf(`apiVersion: 1
kind: Datasource
metadata:
  name: Loki
spec:
  name: Loki
  type: loki
  access: proxy
  url: http://%s.%s.svc.cluster.local:%d
  isDefault: true
  jsonData:
    timeInterval: 1s`, lokiServiceName, namespace, lokiPort),
	}
	_, err = k8sManager.CreateConfigMap(ctx,
		namespace,
		grafanaDatasourceConfigMapName,
		map[string]string{},
		map[string]string{},
		configMapData)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating Grafana datasource configmap.")
	}

	_, err = k8sManager.CreateDeployment(
		ctx,
		namespace,
		grafanaDeploymentName,
		map[string]string{},
		map[string]string{},
		[]apiv1.Container{}, // no init containers
		[]apiv1.Container{{
			Name:  "grafana",
			Image: grafanaImage,
			Ports: []apiv1.ContainerPort{{ContainerPort: grafanaPort}},
			VolumeMounts: []apiv1.VolumeMount{{
				Name:      "datasources",
				MountPath: "/etc/grafana/provisioning/datasources",
			}},
		}},
		[]apiv1.Volume{{
			Name: "datasources",
			VolumeSource: apiv1.VolumeSource{
				ConfigMap: &apiv1.ConfigMapVolumeSource{
					LocalObjectReference: apiv1.LocalObjectReference{Name: grafanaDatasourceConfigMapName},
				},
			},
		}},
		&apiv1.Affinity{})
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating Grafana deployment.")
	}

	grafanaPorts := []apiv1.ServicePort{{
		Port:       grafanaPort,
		TargetPort: intstr.FromInt(grafanaPort),
		Protocol:   apiv1.ProtocolTCP,
		NodePort:   30030,
	}}

	lokiPorts := []apiv1.ServicePort{{
		Port:       lokiPort,
		TargetPort: intstr.FromInt(lokiPort),
		Protocol:   apiv1.ProtocolTCP,
		NodePort:   30031,
	}}

	_, err = k8sManager.CreateService(ctx,
		namespace,
		grafanaServiceName,
		map[string]string{},
		nil,
		map[string]string{},
		apiv1.ServiceTypeNodePort,
		grafanaPorts)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating Grafana service")
	}

	_, err = k8sManager.CreateService(ctx,
		namespace,
		lokiServiceName,
		map[string]string{},
		nil,
		map[string]string{},
		apiv1.ServiceTypeNodePort,
		lokiPorts)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating Loki service")
	}

	lokiHost := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", lokiServiceName, namespace, lokiPort)
	return lokiHost, nil
}

func int32Ptr(i int32) *int32 {
	return &i
}
