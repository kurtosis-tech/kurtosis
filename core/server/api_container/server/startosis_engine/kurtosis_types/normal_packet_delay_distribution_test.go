package kurtosis_types

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/port_spec"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

func TestMakeNormalPacketDelayDistribution(t *testing.T) {
	kwargs := []starlark.Tuple{
		starlark.Tuple([]starlark.Value{
			starlark.String(meanAttr),
			starlark.MakeInt(100),
		}),
		starlark.Tuple([]starlark.Value{
			starlark.String(stdDevAttr),
			starlark.MakeInt(10),
		}),
		starlark.Tuple([]starlark.Value{
			starlark.String(correlationAttr),
			starlark.Float(10),
		}),
	}

	builtin := &starlark.Builtin{}

	packetUniformDelayDistribution, err := MakeNormalPacketDelayDistribution(nil, builtin, noArgs, kwargs)
	expectedUniformDelayDistribution := NewNormalPacketDelayDistribution(100, 10, 10)

	require.Nil(t, err)
	require.NotNil(t, packetUniformDelayDistribution)
	require.Equal(t, expectedUniformDelayDistribution, packetUniformDelayDistribution)
}

func TestMakeNormalPacketDelayDistribution_WithNoCorrelation(t *testing.T) {
	kwargs := []starlark.Tuple{
		starlark.Tuple([]starlark.Value{
			starlark.String(meanAttr),
			starlark.MakeInt(100),
		}),
		starlark.Tuple([]starlark.Value{
			starlark.String(stdDevAttr),
			starlark.MakeInt(10),
		}),
	}

	builtin := &starlark.Builtin{}
	packetUniformDelayDistribution, err := MakeNormalPacketDelayDistribution(nil, builtin, noArgs, kwargs)
	expectedUniformDelayDistribution := NewNormalPacketDelayDistribution(100, 10, 0)

	require.Nil(t, err)
	require.NotNil(t, packetUniformDelayDistribution)
	require.Equal(t, expectedUniformDelayDistribution, packetUniformDelayDistribution)
}

func TestMakeNormalPacketDelayDistribution_WithInvalidCorrelation(t *testing.T) {
	kwargs := []starlark.Tuple{
		starlark.Tuple([]starlark.Value{
			starlark.String(meanAttr),
			starlark.MakeInt(100),
		}),
		starlark.Tuple([]starlark.Value{
			starlark.String(stdDevAttr),
			starlark.MakeInt(10),
		}),
		starlark.Tuple([]starlark.Value{
			starlark.String(correlationAttr),
			starlark.Float(110.0),
		}),
	}

	builtin := &starlark.Builtin{}

	_, err := MakeNormalPacketDelayDistribution(nil, builtin, noArgs, kwargs)
	require.NotNil(t, err)
	require.EqualError(
		t,
		err,
		fmt.Sprintf("Invalid attribute. '%s' in '%s' should be greater than 0 and lower than 100. Got '%v'",
			correlationAttr, NormalPacketDelayDistributionName, starlark.Float(110),
		))
}

func TestMakeNormalPacketDelayDistribution_NoNamedArgs(t *testing.T) {
	builtin := &starlark.Builtin{}

	_, err := MakeNormalPacketDelayDistribution(nil, builtin, noArgs, noKwargs)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Cannot construct a NormalPacketDelayDistribution from the provided arguments")
}

func TestMakeNormalPacketDelayDistribution_WithNotRecognizedAttr(t *testing.T) {
	input := starlark.Tuple([]starlark.Value{
		starlark.String(port_spec.PortApplicationProtocolAttr), starlark.MakeInt(100),
	})

	kwargs := []starlark.Tuple{
		input,
	}

	builtin := &starlark.Builtin{}

	_, err := MakeNormalPacketDelayDistribution(nil, builtin, noArgs, kwargs)
	require.NotNil(t, err)
	require.ErrorContains(t, err, `unexpected keyword argument "application_protocol"`)
}

func TestNormalPacketDelayDistribution_Attr(t *testing.T) {
	packetDelay := NewNormalPacketDelayDistribution(100, 10, 0)
	delayMs, err := packetDelay.Attr(meanAttr)
	require.Nil(t, err)
	require.Equal(t, starlark.MakeInt(100), delayMs)

	stdDevMs, err := packetDelay.Attr(stdDevAttr)
	require.Nil(t, err)
	require.Equal(t, starlark.MakeInt(10), stdDevMs)

	correlation, err := packetDelay.Attr(correlationAttr)
	require.Nil(t, err)
	require.Equal(t, starlark.Float(0), correlation)
}

func TestNormalPacketDelayDistribution_AttrNames(t *testing.T) {
	packetDelay := NewNormalPacketDelayDistribution(100, 10, 0)
	actual := packetDelay.AttrNames()
	expected := []string{meanAttr, stdDevAttr, correlationAttr}
	require.Equal(t, expected, actual)
}

func TestNormalPacketDelayDistribution_String(t *testing.T) {
	packetDelay := NewNormalPacketDelayDistribution(100, 10, 0)
	packetDelayStr := packetDelay.String()
	require.Equal(t, "NormalPacketDelayDistribution(mean_ms=100, std_dev_ms=10, correlation=0.0)", packetDelayStr)

	packetDelay = NewNormalPacketDelayDistribution(100, 10, 25.5)
	packetDelayStr = packetDelay.String()
	require.Equal(t, "NormalPacketDelayDistribution(mean_ms=100, std_dev_ms=10, correlation=25.5)", packetDelayStr)
}

func TestNormalPacketDelayDistribution_Type(t *testing.T) {
	packetDelay := NewNormalPacketDelayDistribution(100, 10, 10.5)
	actual := packetDelay.ToKurtosisType()
	expected := partition_topology.NewNormalPacketDelayDistribution(100, 10, 10.5)
	require.Equal(t, expected, actual)
}
