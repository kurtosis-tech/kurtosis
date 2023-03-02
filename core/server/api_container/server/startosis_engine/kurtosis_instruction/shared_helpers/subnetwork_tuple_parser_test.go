package shared_helpers

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/service_network_types"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

const (
	argName = "subnetwork"
)

func TestParseSubnetworks_ValidArg(t *testing.T) {
	expectedPartition1 := "subnetwork_1"
	expectedPartition2 := "subnetwork_2"
	subnetworks := starlark.Tuple([]starlark.Value{
		starlark.String(expectedPartition1),
		starlark.String(expectedPartition2),
	})
	partition1, partition2, err := ParseSubnetworks(argName, subnetworks)
	require.Nil(t, err)
	require.Equal(t, service_network_types.PartitionID(expectedPartition1), partition1)
	require.Equal(t, service_network_types.PartitionID(expectedPartition2), partition2)
}

func TestParseSubnetworks_TooManySubnetworks(t *testing.T) {
	expectedPartition1 := "subnetwork_1"
	expectedPartition2 := "subnetwork_2"
	expectedPartition3 := "subnetwork_3"
	subnetworks := starlark.Tuple([]starlark.Value{
		starlark.String(expectedPartition1),
		starlark.String(expectedPartition2),
		starlark.String(expectedPartition3),
	})
	partition1, partition2, err := ParseSubnetworks(argName, subnetworks)
	require.Contains(t, err.Error(), "Subnetworks tuple should contain exactly 2 subnetwork names. 3 were provided")
	require.Empty(t, partition1)
	require.Empty(t, partition2)
}

func TestParseSubnetworks_TooFewSubnetworks(t *testing.T) {
	expectedPartition1 := "subnetwork_1"
	subnetworks := starlark.Tuple([]starlark.Value{
		starlark.String(expectedPartition1),
	})
	partition1, partition2, err := ParseSubnetworks(argName, subnetworks)
	require.Contains(t, err.Error(), "Subnetworks tuple should contain exactly 2 subnetwork names. 1 was provided")
	require.Empty(t, partition1)
	require.Empty(t, partition2)
}
