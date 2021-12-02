/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package user_service_launcher

import (
	"github.com/kurtosis-tech/kurtosis-core/launcher/enclave_container_launcher"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestOneToOneEnclaveContainerAndObjAttrsSchemaPortProtoMapping(t *testing.T) {
	// Ensure all EnclaveContainerPort protos are covered
	require.Equal(t, len(enclave_container_launcher.AllEnclaveContainerPortProtocols), len(enclaveContainerPortProtosToObjAttrsPortProtos))
	for proto := range enclave_container_launcher.AllEnclaveContainerPortProtocols {
		_, found := enclaveContainerPortProtosToObjAttrsPortProtos[proto]
		require.True(t, found, "No object attributes schema port proto found for enclave container port proto '%v'", proto)
	}

	// Ensure no duplicates in the object attributes schema port protos
	require.Equal(t, len(schema.AllowedProtocols), len(enclaveContainerPortProtosToObjAttrsPortProtos))
	seenObjAttrsProtos := map[schema.PortProtocol]enclave_container_launcher.EnclaveContainerPortProtocol{}
	for enclaveContainerPortProto, objAttrsPortProto := range enclaveContainerPortProtosToObjAttrsPortProtos {
		preexistingEnclaveContainerPortProto, found := seenObjAttrsProtos[objAttrsPortProto]
		require.False(
			t,
			found,
			"Object attributes schema proto '%v' is already mapped to enclave container port protocol '%v'",
			objAttrsPortProto,
			preexistingEnclaveContainerPortProto,
		)
		seenObjAttrsProtos[objAttrsPortProto] = enclaveContainerPortProto
	}
}
