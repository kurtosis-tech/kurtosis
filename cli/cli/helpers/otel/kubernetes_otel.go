package otel

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// kubernetes_otel.go is the Kubernetes analog of docker_otel.go: it deploys the
// same OTel collector + ClickHouse stack (reusing the embedded init.sql and
// collector-bootstrap-config.yaml) as in-cluster Deployments/Services so that
// enclave services and the engine's logs aggregator can ship to it over cluster
// DNS instead of host ports. Wired into StartOtel/StopOtel's Kubernetes case.

const (
	otelNamespace = "kurtosis-otel"

	clickHouseDeploymentName = "kurtosis-otel-clickhouse"
	clickHouseServiceName    = "kurtosis-otel-clickhouse"
	collectorDeploymentName  = "kurtosis-otel-collector"
	collectorServiceName     = "kurtosis-otel-collector"

	clickHouseInitConfigMapName = "kurtosis-otel-clickhouse-init"
	collectorConfigMapName      = "kurtosis-otel-collector-config"
	clickHouseInitDirMountPath  = "/docker-entrypoint-initdb.d"
	collectorConfigDirMountPath = "/etc/otelcol"
	collectorConfigFileName     = "config.yaml"

	otelDeploymentMaxRetries     = 90
	otelDeploymentRetryInterval  = 1 * time.Second
	emptyOtelStorageClassName    = ""
)

var otelNamespaceLabels = map[string]string{"kurtosistech.com/app-id": "kurtosis", "kurtosistech.com/resource-type": "otel"}
var clickHouseLabels = map[string]string{"kurtosistech.com/app-id": "kurtosis", "kurtosistech.com/resource-type": "otel-clickhouse"}
var collectorLabels = map[string]string{"kurtosistech.com/app-id": "kurtosis", "kurtosistech.com/resource-type": "otel-collector"}

// StartOtelInK8s deploys the collector + ClickHouse into the cluster and returns
// their in-cluster endpoints. Idempotent: reuses existing deployments by label.
func StartOtelInK8s(ctx context.Context) (*Endpoints, error) {
	k8sManager, err := getKubernetesManagerForOtel()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting a Kubernetes manager for otel.")
	}

	if _, err := k8sManager.CreateNamespace(ctx, otelNamespace, otelNamespaceLabels, map[string]string{}); err != nil {
		// CreateNamespace errors if it already exists; tolerate that for idempotency.
		logrus.Debugf("Namespace '%v' may already exist: %v", otelNamespace, err)
	}

	clickHouseNativeAddr := fmt.Sprintf("%v.%v.svc.cluster.local:%d", clickHouseServiceName, otelNamespace, clickHouseNativePort)
	if err := ensureClickHouse(ctx, k8sManager); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred ensuring the ClickHouse deployment.")
	}
	if err := ensureCollector(ctx, k8sManager, clickHouseNativeAddr); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred ensuring the OTel collector deployment.")
	}

	collectorHost := fmt.Sprintf("%v.%v.svc.cluster.local", collectorServiceName, otelNamespace)
	endpoints := &Endpoints{
		ClickHouseHTTPURL:       fmt.Sprintf("http://%v.%v.svc.cluster.local:%d", clickHouseServiceName, otelNamespace, clickHouseHTTPPort),
		ClickHouseNativeAddress: clickHouseNativeAddr,
		CollectorOTLPGRPCURL:    fmt.Sprintf("%v:%d", collectorHost, collectorOTLPGRPCPort),
		CollectorOTLPHTTPURL:    fmt.Sprintf("http://%v:%d", collectorHost, collectorOTLPHTTPPort),
		CollectorLokiURL:        fmt.Sprintf("http://%v:%d", collectorHost, collectorLokiPort),
	}
	logrus.Infof("otel stack ready in-cluster: ClickHouse %v, collector OTLP %v / Loki %v", endpoints.ClickHouseHTTPURL, endpoints.CollectorOTLPGRPCURL, endpoints.CollectorLokiURL)
	return endpoints, nil
}

// StopOtelInK8s removes the otel namespace (and everything in it).
func StopOtelInK8s(ctx context.Context) error {
	k8sManager, err := getKubernetesManagerForOtel()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting a Kubernetes manager for otel.")
	}
	if err := k8sManager.RemoveNamespace(ctx, &apiv1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: otelNamespace}}); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing the otel namespace '%v'.", otelNamespace)
	}
	return nil
}

func ensureClickHouse(ctx context.Context, k8sManager *kubernetes_manager.KubernetesManager) error {
	if existing, err := k8sManager.GetDeploymentsByLabels(ctx, otelNamespace, clickHouseLabels); err == nil && existing != nil && len(existing.Items) > 0 {
		logrus.Debugf("ClickHouse deployment already exists; reusing it.")
		return nil
	}
	if _, err := k8sManager.CreateConfigMap(ctx, otelNamespace, clickHouseInitConfigMapName, clickHouseLabels, map[string]string{}, map[string]string{"init.sql": clickHouseInitSQL}); err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the ClickHouse init.sql ConfigMap.")
	}
	deployment, err := k8sManager.CreateDeployment(ctx, otelNamespace, clickHouseDeploymentName, clickHouseLabels, map[string]string{},
		nil,
		[]apiv1.Container{{
			Name:  "clickhouse",
			Image: defaultClickHouseImage,
			// Mirrors docker_otel.go: CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT lets the
			// collector connect as the default user from another pod (without it,
			// remote exports fail with code 516 "Authentication failed").
			Env: []apiv1.EnvVar{
				{Name: "CLICKHOUSE_DB", Value: "otel"},
				{Name: "CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT", Value: "1"},
			},
			Ports: []apiv1.ContainerPort{
				{Name: "http", ContainerPort: int32(clickHouseHTTPPort), Protocol: apiv1.ProtocolTCP},
				{Name: "native", ContainerPort: int32(clickHouseNativePort), Protocol: apiv1.ProtocolTCP},
			},
			VolumeMounts: []apiv1.VolumeMount{{Name: "init", MountPath: clickHouseInitDirMountPath}},
			ReadinessProbe: &apiv1.Probe{
				ProbeHandler:        apiv1.ProbeHandler{HTTPGet: &apiv1.HTTPGetAction{Path: "/ping", Port: intstr.FromInt(int(clickHouseHTTPPort))}},
				InitialDelaySeconds: 3, PeriodSeconds: 2, TimeoutSeconds: 3,
			},
		}},
		[]apiv1.Volume{{
			Name:         "init",
			VolumeSource: apiv1.VolumeSource{ConfigMap: &apiv1.ConfigMapVolumeSource{LocalObjectReference: apiv1.LocalObjectReference{Name: clickHouseInitConfigMapName}}},
		}},
		nil, nil, nil)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the ClickHouse deployment.")
	}
	logrus.Infof("Waiting for ClickHouse to come online...")
	if err := k8sManager.WaitForPodManagedByDeployment(ctx, deployment, otelDeploymentMaxRetries, otelDeploymentRetryInterval); err != nil {
		return stacktrace.Propagate(err, "An error occurred waiting for the ClickHouse deployment to come online.")
	}
	if _, err := k8sManager.CreateService(ctx, otelNamespace, clickHouseServiceName, clickHouseLabels, map[string]string{}, clickHouseLabels, apiv1.ServiceTypeClusterIP,
		[]apiv1.ServicePort{
			{Name: "http", Port: int32(clickHouseHTTPPort), TargetPort: intstr.FromInt(int(clickHouseHTTPPort)), Protocol: apiv1.ProtocolTCP},
			{Name: "native", Port: int32(clickHouseNativePort), TargetPort: intstr.FromInt(int(clickHouseNativePort)), Protocol: apiv1.ProtocolTCP},
		}); err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the ClickHouse service.")
	}
	return nil
}

func ensureCollector(ctx context.Context, k8sManager *kubernetes_manager.KubernetesManager, clickHouseNativeAddr string) error {
	if existing, err := k8sManager.GetDeploymentsByLabels(ctx, otelNamespace, collectorLabels); err == nil && existing != nil && len(existing.Items) > 0 {
		logrus.Debugf("Collector deployment already exists; reusing it.")
		return nil
	}
	collectorConfig := strings.ReplaceAll(collectorBootstrapConfig, "{{CLICKHOUSE_NATIVE_ADDRESS}}", clickHouseNativeAddr)
	if _, err := k8sManager.CreateConfigMap(ctx, otelNamespace, collectorConfigMapName, collectorLabels, map[string]string{}, map[string]string{collectorConfigFileName: collectorConfig}); err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the collector config ConfigMap.")
	}
	deployment, err := k8sManager.CreateDeployment(ctx, otelNamespace, collectorDeploymentName, collectorLabels, map[string]string{},
		nil,
		[]apiv1.Container{{
			Name:  "collector",
			Image: defaultCollectorImage,
			Args:  []string{fmt.Sprintf("--config=%v/%v", collectorConfigDirMountPath, collectorConfigFileName)},
			Ports: []apiv1.ContainerPort{
				{Name: "otlp-grpc", ContainerPort: int32(collectorOTLPGRPCPort), Protocol: apiv1.ProtocolTCP},
				{Name: "otlp-http", ContainerPort: int32(collectorOTLPHTTPPort), Protocol: apiv1.ProtocolTCP},
				{Name: "loki", ContainerPort: int32(collectorLokiPort), Protocol: apiv1.ProtocolTCP},
				{Name: "health", ContainerPort: int32(collectorHealthPort), Protocol: apiv1.ProtocolTCP},
			},
			VolumeMounts: []apiv1.VolumeMount{{Name: "config", MountPath: collectorConfigDirMountPath}},
			ReadinessProbe: &apiv1.Probe{
				ProbeHandler:        apiv1.ProbeHandler{HTTPGet: &apiv1.HTTPGetAction{Path: "/", Port: intstr.FromInt(int(collectorHealthPort))}},
				InitialDelaySeconds: 3, PeriodSeconds: 2, TimeoutSeconds: 3,
			},
		}},
		[]apiv1.Volume{{
			Name:         "config",
			VolumeSource: apiv1.VolumeSource{ConfigMap: &apiv1.ConfigMapVolumeSource{LocalObjectReference: apiv1.LocalObjectReference{Name: collectorConfigMapName}}},
		}},
		nil, nil, nil)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the collector deployment.")
	}
	logrus.Infof("Waiting for OTel collector to come online...")
	if err := k8sManager.WaitForPodManagedByDeployment(ctx, deployment, otelDeploymentMaxRetries, otelDeploymentRetryInterval); err != nil {
		return stacktrace.Propagate(err, "An error occurred waiting for the collector deployment to come online.")
	}
	if _, err := k8sManager.CreateService(ctx, otelNamespace, collectorServiceName, collectorLabels, map[string]string{}, collectorLabels, apiv1.ServiceTypeClusterIP,
		[]apiv1.ServicePort{
			{Name: "otlp-grpc", Port: int32(collectorOTLPGRPCPort), TargetPort: intstr.FromInt(int(collectorOTLPGRPCPort)), Protocol: apiv1.ProtocolTCP},
			{Name: "otlp-http", Port: int32(collectorOTLPHTTPPort), TargetPort: intstr.FromInt(int(collectorOTLPHTTPPort)), Protocol: apiv1.ProtocolTCP},
			{Name: "loki", Port: int32(collectorLokiPort), TargetPort: intstr.FromInt(int(collectorLokiPort)), Protocol: apiv1.ProtocolTCP},
		}); err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the collector service.")
	}
	return nil
}

func getKubernetesManagerForOtel() (*kubernetes_manager.KubernetesManager, error) {
	kubernetesConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		nil,
	).ClientConfig()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating Kubernetes configuration.")
	}
	clientSet, err := kubernetes.NewForConfig(kubernetesConfig)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the Kubernetes client set.")
	}
	return kubernetes_manager.NewKubernetesManager(clientSet, kubernetesConfig, emptyOtelStorageClassName), nil
}
