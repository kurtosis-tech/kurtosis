package lib

import (
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	kb "github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/metrics_reporting"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/stacktrace"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

// GetLocalDockerKurtosisBackend is the entrypoint method we expect users of container-engine-lib to call
func GetLocalDockerKurtosisBackend() (backend_interface.KurtosisBackend, error) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Docker client connected to the local environment")
	}

	dockerManager := docker_manager.NewDockerManager(dockerClient)

	dockerKurtosisBackend := docker.NewDockerKurtosisBackend(dockerManager)

	wrappedBackend := metrics_reporting.NewMetricsReportingKurtosisBackend(dockerKurtosisBackend)

	return wrappedBackend, nil
}

func GetLocalKubernetesKurtosisBackend(volumeStorageClassName string, volumeSizeInGigabytes int) (backend_interface.KurtosisBackend, error) {
	// TODO Implement GetLocalKubernetesProxyKurtosisBackend?
	kubeConfigFileFilepath := filepath.Join(
		os.Getenv("HOME"), ".kube", "config",
	)
	kubernetesConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigFileFilepath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating kubernetes configuration from flags in file '%v'", kubeConfigFileFilepath)
	}

	wrappedBackend, err:= newWrappedKubernetesKurtosisBackend(kubernetesConfig, volumeStorageClassName, volumeSizeInGigabytes)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating new wrapped Kubernetes Kurtosis Backend using Kubernetes config '%+v', volume storage class name '%v' and size '%v'", kubernetesConfig, volumeStorageClassName, volumeSizeInGigabytes)
	}

	return wrappedBackend, nil
}

func GetInClusterKubernetesKurtosisBackend(volumeStorageClassName string, volumeSizeInGigabytes int) (backend_interface.KurtosisBackend, error) {
	kubernetesConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting in cluster Kubernetes config")
	}

	wrappedBackend, err:= newWrappedKubernetesKurtosisBackend(kubernetesConfig, volumeStorageClassName, volumeSizeInGigabytes)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating new wrapped Kubernetes Kurtosis Backend using Kubernetes config '%+v', volume storage class name '%v' and size '%v'", kubernetesConfig, volumeStorageClassName, volumeSizeInGigabytes)
	}

	return wrappedBackend, nil
}

func newWrappedKubernetesKurtosisBackend(kubernetesConfig *rest.Config, volumeStorageClassName string, volumeSizeInGigabytes int) (*metrics_reporting.MetricsReportingKurtosisBackend, error){
	clientSet, err := kubernetes.NewForConfig(kubernetesConfig)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to create kubernetes client set using Kubernetes config '%+v', instead a non nil error was returned", kubernetesConfig)
	}

	kubernetesManager := kubernetes_manager.NewKubernetesManager(clientSet)

	kurtosisBackend := kb.NewKubernetesKurtosisBackend(kubernetesManager, volumeStorageClassName, volumeSizeInGigabytes)

	wrappedBackend := metrics_reporting.NewMetricsReportingKurtosisBackend(kurtosisBackend)

	return wrappedBackend, nil
}
