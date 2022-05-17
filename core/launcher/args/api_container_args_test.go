/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package args

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis-core/launcher/args/kurtosis_backend_config"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	kubernetesArgsJson   = `{"version": "X.X.X", "grpcListenPortNum":9710,"grpcProxyListenPortNum":9711,"logLevelStr":"debug", "enclaveId": "enclave-id", "isPartitioningEnabled": false, "metricsUserId":"5e9d668ad9b004ba16def3ee14c271f5134e1df57a4d4996924e6544e6b0e9be", "enclaveDataVolumeDirpath": "/path/", "didUserAcceptSendingMetrics":true,"kurtosisBackendType":"kubernetes","kurtosisBackendConfig":{"storageClass":"standard","enclaveSizeInMegabytes":10}}`
	dockerArgsJson   = `{"version": "X.X.X", "grpcListenPortNum":9710,"grpcProxyListenPortNum":9711,"logLevelStr":"debug", "enclaveId": "enclave-id", "isPartitioningEnabled": false, "metricsUserId":"5e9d668ad9b004ba16def3ee14c271f5134e1df57a4d4996924e6544e6b0e9be", "enclaveDataVolumeDirpath": "/path/", "didUserAcceptSendingMetrics":true,"kurtosisBackendType":"docker","kurtosisBackendConfig":{}}`
	expectedStorageClass = "standard"
	expectedEnclaveSize = uint(10)
)

func TestArgsUnmarshalKubernetes(t *testing.T) {
	paramsJsonBytes := []byte(kubernetesArgsJson)
	var args APIContainerArgs
	err := json.Unmarshal(paramsJsonBytes, &args)
	require.NoError(t, err)
	backendConfig := args.KurtosisBackendConfig
	kubernetesBackendConfig := backendConfig.(kurtosis_backend_config.KubernetesBackendConfig)
	require.Equal(t, expectedStorageClass, kubernetesBackendConfig.StorageClass)
	require.Equal(t, expectedEnclaveSize, kubernetesBackendConfig.EnclaveSizeInMegabytes)
}

func TestArgsUnmarshalDocker(t *testing.T) {
	paramsJsonBytes := []byte(dockerArgsJson)
	var args APIContainerArgs
	err := json.Unmarshal(paramsJsonBytes, &args)
	require.NoError(t, err)
}
