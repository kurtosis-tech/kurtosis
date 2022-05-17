package args

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis-engine-server/launcher/args/kurtosis_backend_config"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	kubernetesArgsJson   = "{\"grpcListenPortNum\":9710,\"grpcProxyListenPortNum\":9711,\"logLevelStr\":\"debug\",\"imageVersionTag\":\"X.X.X\",\"metricsUserId\":\"5e9d668ad9b004ba16def3ee14c271f5134e1df57a4d4996924e6544e6b0e9be\",\"didUserAcceptSendingMetrics\":true,\"kurtosisBackendType\":\"kubernetes\",\"kurtosisBackendConfig\":{\"storageClass\":\"standard\",\"enclaveSizeInMegabytes\":10}}"
	dockerArgsJson   = "{\"grpcListenPortNum\":9710,\"grpcProxyListenPortNum\":9711,\"logLevelStr\":\"debug\",\"imageVersionTag\":\"X.X.X\",\"metricsUserId\":\"5e9d668ad9b004ba16def3ee14c271f5134e1df57a4d4996924e6544e6b0e9be\",\"didUserAcceptSendingMetrics\":true,\"kurtosisBackendType\":\"docker\",\"kurtosisBackendConfig\":{}}"
	expectedStorageClass = "standard"
	expectedEnclaveSize = uint(10)
)

func TestArgsUnmarshalKubernetes(t *testing.T) {
	paramsJsonBytes := []byte(kubernetesArgsJson)
	var args EngineServerArgs
	err := json.Unmarshal(paramsJsonBytes, &args)
	require.NoError(t, err)
	backendConfig := args.KurtosisBackendConfig
	kubernetesBackendConfig := backendConfig.(kurtosis_backend_config.KubernetesBackendConfig)
	require.Equal(t, expectedStorageClass, kubernetesBackendConfig.StorageClass)
	require.Equal(t, expectedEnclaveSize, kubernetesBackendConfig.EnclaveSizeInMegabytes)
}

func TestArgsUnmarshalDocker(t *testing.T) {
	paramsJsonBytes := []byte(dockerArgsJson)
	var args EngineServerArgs
	err := json.Unmarshal(paramsJsonBytes, &args)
	require.NoError(t, err)
}