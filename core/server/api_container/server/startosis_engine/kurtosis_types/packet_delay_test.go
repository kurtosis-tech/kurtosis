package kurtosis_types

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/partition_topology"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

func TestMakePacketDelay(t *testing.T) {
	input := starlark.Tuple([]starlark.Value{
		starlark.String(delayInMsAttr), starlark.MakeInt(100),
	})

	kwargs := []starlark.Tuple{
		input,
	}

	builtin := &starlark.Builtin{}

	packetDelay, err := MakePacketDelay(nil, builtin, noArgs, kwargs)
	expectedPacketDelay := NewPacketDelay(100)

	require.Nil(t, err)
	require.NotNil(t, packetDelay)
	require.Equal(t, expectedPacketDelay, packetDelay)
}

func TestMakePacketDelayWithNoNamedArgs(t *testing.T) {
	builtin := &starlark.Builtin{}

	_, err := MakePacketDelay(nil, builtin, noArgs, noKwargs)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Cannot construct a PacketDelay from the provided arguments")
}

func TestMakePacketDelayWithNotRecognizedAttr(t *testing.T) {
	input := starlark.Tuple([]starlark.Value{
		starlark.String(packetDelayAttr), starlark.MakeInt(100),
	})

	kwargs := []starlark.Tuple{
		input,
	}

	builtin := &starlark.Builtin{}

	_, err := MakePacketDelay(nil, builtin, noArgs, kwargs)
	require.NotNil(t, err)
	require.ErrorContains(t, err, `unexpected keyword argument "packet_delay"`)
}

func TestPacketDelay_Attr(t *testing.T) {
	packetDelay := NewPacketDelay(100)
	delayInMs, err := packetDelay.Attr(delayInMsAttr)
	require.Nil(t, err)
	require.Equal(t, starlark.MakeInt(100), delayInMs)
}

func TestPacketDelay_AttrNames(t *testing.T) {
	packetDelay := NewPacketDelay(100)
	actual := packetDelay.AttrNames()
	expected := []string{delayInMsAttr}
	require.Equal(t, expected, actual)
}

func TestPacketDelay_String(t *testing.T) {
	packetDelay := NewPacketDelay(100)
	packetDelayStr := packetDelay.String()
	require.Equal(t, "PacketDelay(delay_ms=100)", packetDelayStr)

	packetDelay = NewPacketDelay(0)
	packetDelayStr = packetDelay.String()
	require.Equal(t, "PacketDelay(delay_ms=0)", packetDelayStr)
}

func TestPacketDelay_ToKurtosisType(t *testing.T) {
	packetDelay := NewPacketDelay(100)
	actual := packetDelay.ToKurtosisType()
	expected := partition_topology.NewUniformPacketDelayDistribution(100)
	require.Equal(t, expected, actual)
}
