package kurtosis_types

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/port_spec"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

func TestMakeUniformPacketDelayDistribution(t *testing.T) {
	input := starlark.Tuple([]starlark.Value{
		starlark.String(delayAttr), starlark.MakeInt(100),
	})

	kwargs := []starlark.Tuple{
		input,
	}

	builtin := &starlark.Builtin{}

	packetUniformDelayDistribution, err := MakeUniformPacketDelayDistribution(nil, builtin, noArgs, kwargs)
	expectedUniformDelayDistribution := NewUniformPacketDelayDistribution(100)

	require.Nil(t, err)
	require.NotNil(t, packetUniformDelayDistribution)
	require.Equal(t, expectedUniformDelayDistribution, packetUniformDelayDistribution)
}

func TestMakeUniformPacketDelayDistribution_NoNamedArgs(t *testing.T) {
	builtin := &starlark.Builtin{}

	_, err := MakeUniformPacketDelayDistribution(nil, builtin, noArgs, noKwargs)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Cannot construct a UniformPacketDelayDistribution from the provided arguments")
}

func TestMakeUniformPacketDelayDistribution_WithNotRecognizedAttr(t *testing.T) {
	input := starlark.Tuple([]starlark.Value{
		starlark.String(port_spec.PortApplicationProtocolAttr), starlark.MakeInt(100),
	})

	kwargs := []starlark.Tuple{
		input,
	}

	builtin := &starlark.Builtin{}

	_, err := MakeUniformPacketDelayDistribution(nil, builtin, noArgs, kwargs)
	require.NotNil(t, err)
	require.ErrorContains(t, err, `unexpected keyword argument "application_protocol"`)
}

func TestUniformPacketDelayDistribution_Attr(t *testing.T) {
	packetDelay := NewUniformPacketDelayDistribution(100)
	delayInMs, err := packetDelay.Attr(delayAttr)
	require.Nil(t, err)
	require.Equal(t, starlark.MakeInt(100), delayInMs)
}

func TestUniformPacketDelayDistribution_AttrNames(t *testing.T) {
	packetDelay := NewUniformPacketDelayDistribution(100)
	actual := packetDelay.AttrNames()
	expected := []string{delayAttr}
	require.Equal(t, expected, actual)
}

func TestUniformPacketDelayDistribution_String(t *testing.T) {
	packetDelay := NewUniformPacketDelayDistribution(100)
	packetDelayStr := packetDelay.String()
	require.Equal(t, "UniformPacketDelayDistribution(ms=100)", packetDelayStr)

	packetDelay = NewUniformPacketDelayDistribution(0)
	packetDelayStr = packetDelay.String()
	require.Equal(t, "UniformPacketDelayDistribution(ms=0)", packetDelayStr)
}

func TestUniformPacketDelayDistribution_Type(t *testing.T) {
	packetDelay := NewUniformPacketDelayDistribution(100)
	actual := packetDelay.ToKurtosisType()
	expected := partition_topology.NewUniformPacketDelayDistribution(100)
	require.Equal(t, expected, actual)
}
