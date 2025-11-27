package enclave_status_stringifier

import (
	"github.com/dzobbe/PoTE-kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAllEnclaveContainersStatusAreCovered(t *testing.T) {
	for key := range kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus_name {
		_, err := EnclaveContainersStatusStringifier(kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus(key))
		assert.Nil(t, err)
	}
}
