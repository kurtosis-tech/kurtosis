package enclave_manager

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetRandomEnclaveIdWithRetriesSuccess(t *testing.T) {
	retries := uint16(5)

	noCurrentEnclave := map[enclave.EnclaveUUID]*enclave.Enclave{}

	enclaveIdGeneratorObj := GetEnclaveNameGenerator()

	randomEnclaveId, err := enclaveIdGeneratorObj.GetRandomEnclaveNameWithRetries(noCurrentEnclave, retries)
	require.NoError(t, err)
	require.NotEmpty(t, randomEnclaveId)
}
