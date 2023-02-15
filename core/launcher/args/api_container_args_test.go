/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package args

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	kubernetesArgsJson = `{"version": "X.X.X", "grpcListenPortNum":9710,"grpcProxyListenPortNum":9711,"logLevelStr":"debug", "enclaveUuid": "enclave-id", "isPartitioningEnabled": false, "metricsUserId":"5e9d668ad9b004ba16def3ee14c271f5134e1df57a4d4996924e6544e6b0e9be", "enclaveDataVolumeDirpath": "/path/", "didUserAcceptSendingMetrics":true,"kurtosisBackendType":"kubernetes","kurtosisBackendConfig":{}}`
	dockerArgsJson     = `{"version": "X.X.X", "grpcListenPortNum":9710,"grpcProxyListenPortNum":9711,"logLevelStr":"debug", "enclaveUuid": "enclave-id", "isPartitioningEnabled": false, "metricsUserId":"5e9d668ad9b004ba16def3ee14c271f5134e1df57a4d4996924e6544e6b0e9be", "enclaveDataVolumeDirpath": "/path/", "didUserAcceptSendingMetrics":true,"kurtosisBackendType":"docker","kurtosisBackendConfig":{}}`
)

func TestArgsUnmarshalKubernetes(t *testing.T) {
	paramsJsonBytes := []byte(kubernetesArgsJson)
	var args APIContainerArgs
	err := json.Unmarshal(paramsJsonBytes, &args)
	require.NoError(t, err)
}

func TestArgsUnmarshalDocker(t *testing.T) {
	paramsJsonBytes := []byte(dockerArgsJson)
	var args APIContainerArgs
	err := json.Unmarshal(paramsJsonBytes, &args)
	require.NoError(t, err)
}
