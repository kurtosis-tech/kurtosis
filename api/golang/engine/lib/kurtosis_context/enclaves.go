package kurtosis_context

import "github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"

// Enclaves A collection of enclaves by uuid, name and shortened uuid
type Enclaves struct {
	enclavesByUuid          map[string]*kurtosis_engine_rpc_api_bindings.EnclaveInfo
	enclavesByName          map[string][]*kurtosis_engine_rpc_api_bindings.EnclaveInfo
	enclavesByShortenedUuid map[string][]*kurtosis_engine_rpc_api_bindings.EnclaveInfo
}

func (enclaves *Enclaves) GetEnclavesByUuid() map[string]*kurtosis_engine_rpc_api_bindings.EnclaveInfo {
	return enclaves.enclavesByUuid
}

func (enclaves *Enclaves) GetEnclavesByName() map[string][]*kurtosis_engine_rpc_api_bindings.EnclaveInfo {
	return enclaves.enclavesByName
}
func (enclaves *Enclaves) GetEnclavesByShortenedUuid() map[string][]*kurtosis_engine_rpc_api_bindings.EnclaveInfo {
	return enclaves.enclavesByShortenedUuid
}
