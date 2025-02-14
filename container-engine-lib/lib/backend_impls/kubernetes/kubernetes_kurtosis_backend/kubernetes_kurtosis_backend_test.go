package kubernetes_kurtosis_backend

//func TestCreateLogsCollectorForEnclave(t *testing.T) {
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
//	backend := NewEngineServerKubernetesKurtosisBackend(kubernetesManager)
//
//	logCollector, err := backend.CreateLogsCollectorForEnclave(ctx, "1234", 2020, 2020)
//	require.NoError(t, err)
//	require.NotNil(t, logCollector)
//}
