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
//	//logsCollector, err := backend.CreateLogsCollectorForEnclave(ctx, "1234", 2020, 2020)
//	//require.NoError(t, err)
//	//require.NotNil(t, logsCollector)
//	//logsCollector, err := backend.GetLogsCollectorForEnclave(ctx, "")
//	//require.NoError(t, err)
//	//require.NotNil(t, logsCollector)
//	//require.Equal(t, logsCollector.GetStatus(), container.ContainerStatus_Running)
//
//	// test the retrieval and test the destroying
//	err = backend.DestroyLogsCollectorForEnclave(ctx, "")
//	require.NoError(t, err)
//
//	logsCollectorObj, err := backend.CreateLogsCollectorForEnclave(ctx, "", 0, 0)
//	require.NoError(t, err)
//	require.Equal(t, logsCollectorObj.GetStatus(), container.ContainerStatus_Running)
//
//	//logsCollectorObj, err = backend.GetLogsCollectorForEnclave(ctx, "")
//	//require.NoError(t, err)
//	//require.NotNil(t, logsCollectorObj)
//}
