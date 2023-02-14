package enclave_status_stringifier

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAllEnclaveContainersStatusAreCovered(t *testing.T) {
	for key := range kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus_name {
		_, err := EnclaveContainersStatusStringifier(kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus(key))
		assert.Nil(t, err)
	}
}

func TestAllEnclaveAPIContainersStatusStringifier(t *testing.T) {
	for key := range kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus_name {
		_, err := EnclaveAPIContainersStatusStringifier(kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus(key))
		assert.Nil(t, err)
	}
}
