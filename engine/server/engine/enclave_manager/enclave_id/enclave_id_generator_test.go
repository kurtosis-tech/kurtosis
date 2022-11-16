package enclave_id

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetRandomEnclaveIdWithRetriesSuccess(t *testing.T){
	retries := uint16(5)

	noCurrentEnclave := map[enclave.EnclaveID]*enclave.Enclave{}

	randomEnclaveId, err := GetRandomEnclaveIdWithRetries(noCurrentEnclave, retries)
	require.NoError(t, err)
	require.NotEmpty(t, randomEnclaveId)
}
