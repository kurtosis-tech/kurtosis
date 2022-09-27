/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package server

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestOneToOneApiAndPortSpecProtoMapping(t *testing.T) {
	// Ensure all port spec protos are covered
	require.Equal(t, len(kurtosis_core_rpc_api_bindings.Port_Protocol_name), len(apiContainerPortProtoToPortSpecPortProto))
	for enumInt, enumName := range kurtosis_core_rpc_api_bindings.Port_Protocol_name {
		_, found := apiContainerPortProtoToPortSpecPortProto[kurtosis_core_rpc_api_bindings.Port_Protocol(enumInt)]
		require.True(t, found, "No port spec port proto found for API port proto '%v'", enumName)
	}

	// Ensure no duplicates in the kurtosis backend port protos
	require.Equal(t, len(port_spec.PortProtocolValues()), len(apiContainerPortProtoToPortSpecPortProto))
	seenPortSpecProtos := map[port_spec.PortProtocol]kurtosis_core_rpc_api_bindings.Port_Protocol{}
	for apiPortProto, portSpecProto := range apiContainerPortProtoToPortSpecPortProto {
		preexistingApiPortProto, found := seenPortSpecProtos[portSpecProto]
		require.False(
			t,
			found,
			"port spec proto '%v' is already mapped to API port protocol '%v'",
			portSpecProto,
			preexistingApiPortProto.String(),
		)
		seenPortSpecProtos[portSpecProto] = apiPortProto
	}
}
