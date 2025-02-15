package kubernetes_kurtosis_backend

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"testing"
)

func TestCreateLogsCollectorForEnclave(t *testing.T) {
	kubernetesConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigFileFilepath)
	require.NoError(t, err)

	clientSet, err := kubernetes.NewForConfig(kubernetesConfig)
	require.NoError(t, err)

	kubernetesManager := kubernetes_manager.NewKubernetesManager(clientSet, kubernetesConfig, "")

	ctx := context.Background()

	backend := NewEngineServerKubernetesKurtosisBackend(kubernetesManager)

	logsCollector, err := backend.CreateLogsCollectorForEnclave(ctx, "1234", 2020, 2020)
	require.NoError(t, err)
	require.NotNil(t, logsCollector)
}
