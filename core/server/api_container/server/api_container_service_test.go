/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package server

import (
	"github.com/kurtosis-tech/kurtosis-core/api/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-core/launcher/enclave_container_launcher"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestOneToOneApiAndEnclaveContainerPortProtoMapping(t *testing.T) {
	// Ensure all EnclaveContainerPort protos are covered
	require.Equal(t, len(kurtosis_core_rpc_api_bindings.Port_Protocol_name), len(apiContainerPortProtoToEnclaveContainerPortProto))
	for enumInt, enumName := range kurtosis_core_rpc_api_bindings.Port_Protocol_name {
		_, found := apiContainerPortProtoToEnclaveContainerPortProto[kurtosis_core_rpc_api_bindings.Port_Protocol(enumInt)]
		require.True(t, found, "No enclave conatainer port proto found for API port proto '%v'", enumName)
	}

	// Ensure no duplicates in the object attributes schema port protos
	require.Equal(t, len(enclave_container_launcher.AllEnclaveContainerPortProtocols), len(apiContainerPortProtoToEnclaveContainerPortProto))
	seenEnclaveContainerProtos := map[enclave_container_launcher.EnclaveContainerPortProtocol]kurtosis_core_rpc_api_bindings.Port_Protocol{}
	for apiPortProto, enclaveContainerPortProto := range apiContainerPortProtoToEnclaveContainerPortProto {
		preexistingApiPortProto, found := seenEnclaveContainerProtos[enclaveContainerPortProto]
		require.False(
			t,
			found,
			"Enclave container port proto '%v' is already mapped to API port protocol '%v'",
			enclaveContainerPortProto,
			preexistingApiPortProto.String(),
		)
		seenEnclaveContainerProtos[enclaveContainerPortProto] = apiPortProto
	}
}
