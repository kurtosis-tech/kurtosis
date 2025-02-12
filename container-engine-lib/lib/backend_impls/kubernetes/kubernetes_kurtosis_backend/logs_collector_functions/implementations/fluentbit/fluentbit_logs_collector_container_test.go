package fluentbit

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"testing"
)

var kubeConfigFileFilepath = filepath.Join(
	os.Getenv("HOME"), ".kube", "config",
)

func TestCreatingFluentbitLogCollector(t *testing.T) {
	kubernetesConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigFileFilepath)
	require.NoError(t, err)

	clientSet, err := kubernetes.NewForConfig(kubernetesConfig)
	require.NoError(t, err)

	kubernetesManager := kubernetes_manager.NewKubernetesManager(clientSet, kubernetesConfig, "")

	ctx := context.Background()

	err = CreateLogsCollectorConfigMap(ctx, kubernetesManager)
	require.NoError(t, err)

	_, err = CreateLogsCollectorDaemonSet(ctx, kubernetesManager)
	require.NoError(t, err)
}
