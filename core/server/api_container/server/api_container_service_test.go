/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package server

import (
	"context"
	"strings"
	"testing"

	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/stretchr/testify/require"
)

func TestOneToOneApiAndPortSpecProtoMapping(t *testing.T) {
	// Ensure all port spec protos are covered
	require.Equal(t, len(kurtosis_core_rpc_api_bindings.Port_TransportProtocol_name), len(apiContainerPortProtoToPortSpecPortProto))
	for enumInt, enumName := range kurtosis_core_rpc_api_bindings.Port_TransportProtocol_name {
		_, found := apiContainerPortProtoToPortSpecPortProto[kurtosis_core_rpc_api_bindings.Port_TransportProtocol(enumInt)]
		require.True(t, found, "No port spec port proto found for API port proto '%v'", enumName)
	}

	// Ensure no duplicates in the kurtosis backend port protos
	require.Equal(t, len(port_spec.TransportProtocolValues()), len(apiContainerPortProtoToPortSpecPortProto))
	seenPortSpecProtos := map[port_spec.TransportProtocol]kurtosis_core_rpc_api_bindings.Port_TransportProtocol{}
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

func TestGetTextRepresentation(t *testing.T) {
	input := `my
line
input
`
	expectedOutput := `my
line
`
	output, err := getTextRepresentation(strings.NewReader(input), 2)
	require.NoError(t, err)
	require.NotNil(t, output)
	require.Equal(t, expectedOutput, *output)
}

func TestCreateSnapshot(t *testing.T) {
	dockerManager, err := docker_manager.CreateDockerManager([]client.Opt{})
	require.NoError(t, err)

	containers, err := dockerManager.GetContainersByLabels(context.Background(), map[string]string{
		docker_label_key.IDDockerLabelKey.GetString(): "test1",
	}, false)
	require.NoError(t, err)
	require.Equal(t, 0, len(containers))
}
