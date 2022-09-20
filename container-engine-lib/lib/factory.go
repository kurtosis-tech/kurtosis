package lib

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/metrics_reporting"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/stacktrace"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

func GetCLIKubernetesKurtosisBackend(ctx context.Context) (backend_interface.KurtosisBackend, error) {
	kubeConfigFileFilepath := filepath.Join(
		os.Getenv("HOME"), ".kube", "config",
	)
	kubernetesConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigFileFilepath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating kubernetes configuration from flags in file '%v'", kubeConfigFileFilepath)
	}

	backendSupplier := func(_ context.Context, kubernetesManager *kubernetes_manager.KubernetesManager) (*kubernetes_kurtosis_backend.KubernetesKurtosisBackend, error) {
		return kubernetes_kurtosis_backend.NewCLIModeKubernetesKurtosisBackend(kubernetesManager), nil
	}

	wrappedBackend, err := getWrappedKubernetesKurtosisBackend(
		ctx,
		kubernetesConfig,
		backendSupplier,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred wrapping the CLI Kubernetes backend")
	}

	return wrappedBackend, nil
}

func GetEngineServerKubernetesKurtosisBackend(
	ctx context.Context,
) (backend_interface.KurtosisBackend, error) {
	kubernetesConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting in cluster Kubernetes config")
	}

	backendSupplier := func(_ context.Context, kubernetesManager *kubernetes_manager.KubernetesManager) (*kubernetes_kurtosis_backend.KubernetesKurtosisBackend, error) {
		return kubernetes_kurtosis_backend.NewEngineServerKubernetesKurtosisBackend(
			kubernetesManager,
		), nil
	}

	wrappedBackend, err := getWrappedKubernetesKurtosisBackend(
		ctx,
		kubernetesConfig,
		backendSupplier,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred wrapping the CLI Kubernetes backend")
	}

	return wrappedBackend, nil
}

func GetApiContainerKubernetesKurtosisBackend(
	ctx context.Context,
) (backend_interface.KurtosisBackend, error) {
	kubernetesConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting in cluster Kubernetes config")
	}

	namespaceName := os.Getenv(kubernetes_kurtosis_backend.ApiContainerOwnNamespaceNameEnvVar)
	if namespaceName == "" {
		return nil, stacktrace.NewError("Expected to find environment variable '%v' containing own namespace information when instantiating an API container Kurtosis backend, but none was found", kubernetes_kurtosis_backend.ApiContainerOwnNamespaceNameEnvVar)
	}

	backendSupplier := func(ctx context.Context, kubernetesManager *kubernetes_manager.KubernetesManager) (*kubernetes_kurtosis_backend.KubernetesKurtosisBackend, error) {
		namespace, err := kubernetesManager.GetNamespace(ctx, namespaceName)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting the API container's own namespace '%v'", namespaceName)
		}

		namespaceLabels := namespace.GetLabels()
		enclaveIdStr, found := namespaceLabels[label_key_consts.IDKubernetesLabelKey.GetString()]
		if !found {
			return nil, stacktrace.NewError(
				"Expected to find enclave ID label '%v' on namespace '%v' but none was found",
				label_key_consts.IDKubernetesLabelKey.GetString(),
				namespaceName,
			)
		}
		enclaveId := enclave.EnclaveID(enclaveIdStr)

		return kubernetes_kurtosis_backend.NewAPIContainerKubernetesKurtosisBackend(
			kubernetesManager,
			enclaveId,
			namespaceName,
		), nil
	}

	wrappedBackend, err := getWrappedKubernetesKurtosisBackend(
		ctx,
		kubernetesConfig,
		backendSupplier,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred wrapping the CLI Kubernetes backend")
	}

	return wrappedBackend, nil
}

// ====================================================================================================
//                                      Private Helper Functions
// ====================================================================================================
func getWrappedKubernetesKurtosisBackend(
	ctx context.Context,
	kubernetesConfig *rest.Config,
	kurtosisBackendSupplier func(context.Context, *kubernetes_manager.KubernetesManager) (*kubernetes_kurtosis_backend.KubernetesKurtosisBackend, error),
) (*metrics_reporting.MetricsReportingKurtosisBackend, error) {
	clientSet, err := kubernetes.NewForConfig(kubernetesConfig)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to create kubernetes client set using Kubernetes config '%+v', instead a non nil error was returned", kubernetesConfig)
	}

	kubernetesManager := kubernetes_manager.NewKubernetesManager(clientSet, kubernetesConfig)

	kubernetesBackend, err := kurtosisBackendSupplier(ctx, kubernetesManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the Kurtosis backend")
	}

	wrappedBackend := metrics_reporting.NewMetricsReportingKurtosisBackend(kubernetesBackend)
	return wrappedBackend, nil
}
