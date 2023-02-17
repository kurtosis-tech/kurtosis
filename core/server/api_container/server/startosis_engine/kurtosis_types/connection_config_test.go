package kurtosis_types

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/partition_topology"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

var (
	noKwargs                  []starlark.Tuple
	noArgs                    = starlark.Tuple{}
	noPacketDelayDistribution = NewUniformPacketDelayDistribution(0)
)

func TestConnectionConfig_StringRepresentation(t *testing.T) {
	connectionConfig := NewConnectionConfig(50.5, nil)
	expectedRepresentation := fmt.Sprintf("%s(%s=50.5)", ConnectionConfigTypeName, packetLossPercentageAttr)
	require.Equal(t, expectedRepresentation, connectionConfig.String())

	connectionConfig = NewConnectionConfig(50.5, noPacketDelayDistribution)
	expectedRepresentation = fmt.Sprintf("%s(%s=50.5, packet_delay_distribution=UniformPacketDelayDistribution(ms=0))", ConnectionConfigTypeName, packetLossPercentageAttr)
	require.Equal(t, expectedRepresentation, connectionConfig.String())

	connectionConfig = NewConnectionConfig(50.5, NewNormalPacketDelayDistribution(100, 10, 10))
	expectedRepresentation = fmt.Sprintf("%s(%s=50.5, packet_delay_distribution=NormalPacketDelayDistribution(mean_ms=100, std_dev_ms=10, correlation=10.0))", ConnectionConfigTypeName, packetLossPercentageAttr)
	require.Equal(t, expectedRepresentation, connectionConfig.String())
}

func TestConnectionConfig_Type(t *testing.T) {
	connectionConfig := NewConnectionConfig(50.5, nil)
	require.Equal(t, ConnectionConfigTypeName, connectionConfig.Type())
}

func TestConnectionConfig_Truth_False(t *testing.T) {
	connectionConfig := NewConnectionConfig(-1, nil)
	require.Equal(t, starlark.Bool(false), connectionConfig.Truth())
}

func TestConnectionConfig_Truth_True(t *testing.T) {
	connectionConfig := NewConnectionConfig(50.5, nil)
	require.Equal(t, starlark.Bool(true), connectionConfig.Truth())
}

func TestConnectionConfig_Attr_Exists(t *testing.T) {
	connectionConfig := NewConnectionConfig(50.5, nil)
	attr, err := connectionConfig.Attr(packetLossPercentageAttr)
	require.Nil(t, err)
	require.Equal(t, starlark.Float(50.5), attr)

	attr, err = connectionConfig.Attr(packetDelayDistributionAttr)
	require.Nil(t, err)
	require.Equal(t, nil, attr)
}

func TestConnectionConfig_Attr_DoesNotExist(t *testing.T) {
	connectionConfig := NewConnectionConfig(50.5, nil)
	attr, err := connectionConfig.Attr("do-not-exist")
	expectedError := fmt.Sprintf("'%s' has no attribute 'do-not-exist'", ConnectionConfigTypeName)
	require.Equal(t, expectedError, err.Error())
	require.Nil(t, attr)
}

func TestConnectionConfig_AttrNames(t *testing.T) {
	connectionConfig := NewConnectionConfig(50.5, nil)
	attrs := connectionConfig.AttrNames()
	expectedAttrs := []string{
		packetLossPercentageAttr,
		packetDelayDistributionAttr,
	}
	require.Equal(t, expectedAttrs, attrs)
}

func TestConnectionConfig_MakeWithArgs(t *testing.T) {
	builtin := &starlark.Builtin{}
	args := starlark.Tuple([]starlark.Value{
		starlark.Float(50),
	})
	connectionConfig, err := MakeConnectionConfig(nil, builtin, args, noKwargs)
	require.Nil(t, err)
	expectedConnectionResult := NewConnectionConfig(50, nil)
	require.Equal(t, expectedConnectionResult, connectionConfig)

	// for args, user would still need to pass 0 as the first argument
	// assumption is that users use named args instead of positional args so should be good
	packetDelay := NewUniformPacketDelayDistribution(100)
	args = []starlark.Value{
		starlark.Float(0),
		packetDelay,
	}
	connectionConfig, err = MakeConnectionConfig(nil, builtin, args, noKwargs)
	require.Nil(t, err)
	expectedConnectionResult = NewConnectionConfig(0, packetDelay)
	require.Equal(t, expectedConnectionResult, connectionConfig)
}

func TestConnectionConfig_MakeWithArgs_FailureBadArgument(t *testing.T) {
	builtin := &starlark.Builtin{}
	args := starlark.Tuple([]starlark.Value{
		starlark.String("hello"),
	})
	connectionConfig, err := MakeConnectionConfig(nil, builtin, args, noKwargs)
	expectedError := fmt.Sprintf(`Cannot construct '%s' from the provided arguments
	Caused by: : for parameter %s: got string, want float`, ConnectionConfigTypeName, packetLossPercentageAttr)
	require.Equal(t, expectedError, err.Error())
	require.Nil(t, connectionConfig)
}

func TestConnectionConfig_MakeWithKwargs_WithPacketLossAndNoDelay(t *testing.T) {
	builtin := &starlark.Builtin{}
	kwargs := []starlark.Tuple{
		starlark.Tuple([]starlark.Value{
			starlark.String(packetLossPercentageAttr),
			starlark.Float(50),
		}),
	}
	connectionConfig, err := MakeConnectionConfig(nil, builtin, noArgs, kwargs)
	require.Nil(t, err)
	expectedConnectionResult := NewConnectionConfig(50, nil)
	require.Equal(t, expectedConnectionResult, connectionConfig)
}

func TestConnectionConfig_MakeWithKwargs_WithNoPacketLossAndPacketDelay(t *testing.T) {
	builtin := &starlark.Builtin{}
	packetDelayDistribution := NewUniformPacketDelayDistribution(100)
	kwargs := []starlark.Tuple{
		[]starlark.Value{
			starlark.String(packetDelayDistributionAttr),
			packetDelayDistribution,
		},
	}
	connectionConfig, err := MakeConnectionConfig(nil, builtin, noArgs, kwargs)
	require.Nil(t, err)
	expectedConnectionResult := NewConnectionConfig(0, packetDelayDistribution)
	require.Equal(t, expectedConnectionResult, connectionConfig)

	packetDelay := NewNormalPacketDelayDistribution(100, 10, 5)
	kwargs = []starlark.Tuple{
		[]starlark.Value{
			starlark.String(packetDelayDistributionAttr),
			packetDelay,
		},
	}
	connectionConfig, err = MakeConnectionConfig(nil, builtin, noArgs, kwargs)
	require.Nil(t, err)
	expectedConnectionResult = NewConnectionConfig(0, packetDelay)
	require.Equal(t, expectedConnectionResult, connectionConfig)
}
func TestConnectionConfig_MakeWithKwargs_WithPacketLossAndDelay(t *testing.T) {
	builtin := &starlark.Builtin{}

	packetDelayDistribution := NewUniformPacketDelayDistribution(100)
	kwargs := []starlark.Tuple{
		[]starlark.Value{
			starlark.String(packetDelayDistributionAttr),
			packetDelayDistribution,
		},
		[]starlark.Value{
			starlark.String(packetLossPercentageAttr),
			starlark.Float(50),
		},
	}

	connectionConfig, err := MakeConnectionConfig(nil, builtin, noArgs, kwargs)
	require.Nil(t, err)
	expectedConnectionResult := NewConnectionConfig(50, packetDelayDistribution)
	require.Equal(t, expectedConnectionResult, connectionConfig)

	packetDelay := NewNormalPacketDelayDistribution(100, 10, 5)
	kwargs = []starlark.Tuple{
		[]starlark.Value{
			starlark.String(packetDelayDistributionAttr),
			packetDelay,
		},
		[]starlark.Value{
			starlark.String(packetLossPercentageAttr),
			starlark.Float(50),
		},
	}

	connectionConfig, err = MakeConnectionConfig(nil, builtin, noArgs, kwargs)
	require.Nil(t, err)
	expectedConnectionResult = NewConnectionConfig(50, packetDelay)
	require.Equal(t, expectedConnectionResult, connectionConfig)
}
func TestConnectionConfig_MakeWithKwargs_WithNoPacketLossAndNoDelay(t *testing.T) {
	builtin := &starlark.Builtin{}
	connectionConfig, err := MakeConnectionConfig(nil, builtin, noArgs, noKwargs)
	require.Nil(t, err)
	expectedConnectionResult := NewConnectionConfig(0, nil)
	require.Equal(t, expectedConnectionResult, connectionConfig)
}

func TestConnectionConfig_MakeWithKwargs_FailureWrongArg(t *testing.T) {
	builtin := &starlark.Builtin{}
	kwargs := []starlark.Tuple{
		starlark.Tuple([]starlark.Value{
			starlark.String(packetLossPercentageAttr),
			starlark.Float(50),
		}),
		starlark.Tuple([]starlark.Value{
			starlark.String("unknown-kwarg"),
			starlark.Float(50),
		}),
	}
	connectionConfig, err := MakeConnectionConfig(nil, builtin, noArgs, kwargs)
	expectedError := fmt.Sprintf(`Cannot construct '%s' from the provided arguments
	Caused by: : unexpected keyword argument "unknown-kwarg"`, ConnectionConfigTypeName)
	require.Equal(t, expectedError, err.Error())
	require.Nil(t, connectionConfig)
}

func TestConnectionConfig_ToKurtosisType(t *testing.T) {
	expectedDefaultUniformDistribution := partition_topology.NewUniformPacketDelayDistribution(0)
	expectedUniformDelayFromPacketDelayDistribution := partition_topology.NewUniformPacketDelayDistribution(500)
	connectionConfigFromPacketDelay := NewConnectionConfig(50, nil)

	expectedKurtosisTypeFromPacketDelay :=
		partition_topology.NewPartitionConnection(
			partition_topology.NewPacketLoss(50),
			expectedDefaultUniformDistribution,
		)

	require.Equal(t, expectedKurtosisTypeFromPacketDelay, *connectionConfigFromPacketDelay.ToKurtosisType())

	connectionConfigFromPacketDelayDistribution :=
		NewConnectionConfig(50, NewUniformPacketDelayDistribution(500))

	expectedKurtosisTypeFromPacketDelayDistribution :=
		partition_topology.NewPartitionConnection(
			partition_topology.NewPacketLoss(50),
			expectedUniformDelayFromPacketDelayDistribution,
		)

	require.Equal(t, expectedKurtosisTypeFromPacketDelayDistribution, *connectionConfigFromPacketDelayDistribution.ToKurtosisType())
}
