package fluentbit

import (
	"os"
	"path/filepath"
)

var kubeConfigFileFilepath = filepath.Join(
	os.Getenv("HOME"), ".kube", "config",
)

//func TestCreatingFluentbitLogCollector(t *testing.T) {
//	kubernetesConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigFileFilepath)
//	require.NoError(t, err)
//
//	clientSet, err := kubernetes.NewForConfig(kubernetesConfig)
//	require.NoError(t, err)
//
//	kubernetesManager := kubernetes_manager.NewKubernetesManager(clientSet, kubernetesConfig, "")
//
//	ctx := context.Background()
//
//	err = CreateLogsCollectorConfigMap(ctx, kubernetesManager)
//	require.NoError(t, err)
//
//	_, err = CreateLogsCollectorDaemonSet(ctx, kubernetesManager)
//	require.NoError(t, err)
//}
