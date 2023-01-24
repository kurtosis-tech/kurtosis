package kurtosis_types

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"strings"
)

const (
	ConnectionConfigTypeName = "ConnectionConfig"
	packetLossPercentageAttr = "packet_loss_percentage"
	packetDelayAttr          = "packet_delay"
)

var (
	PreBuiltConnectionConfigs = &starlarkstruct.Module{
		Name: "connection",
		Members: starlark.StringDict{
			"BLOCKED": NewConnectionConfig(100, NoPacketDelay),
			"ALLOWED": NewConnectionConfig(0, NoPacketDelay),
		},
	}
)

// ConnectionConfig A starlark.Value that represents a connection config between 2 subnetworks
type ConnectionConfig struct {
	packetLossPercentage starlark.Float
	packetDelay          *PacketDelay
}

func NewConnectionConfig(packetLossPercentage starlark.Float, packetDelay *PacketDelay) *ConnectionConfig {
	return &ConnectionConfig{
		packetLossPercentage: packetLossPercentage,
		packetDelay:          packetDelay,
	}
}

func MakeConnectionConfig(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var packetLossPercentage starlark.Float
	var packetDelay *PacketDelay

	if err := starlark.UnpackArgs(b.Name(), args, kwargs, MakeOptional(packetLossPercentageAttr), &packetLossPercentage, MakeOptional(packetDelayAttr), &packetDelay); err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Cannot construct '%s' from the provided arguments. Expecting a single argument '%s'", ConnectionConfigTypeName, packetLossPercentageAttr)
	}
	if packetLossPercentage < 0 || packetLossPercentage > 100 {
		return nil, startosis_errors.NewInterpretationError("Invalid attribute. '%s' in '%s' should be greater than 0 and lower than 100. Got '%v'", packetLossPercentageAttr, ConnectionConfigTypeName, packetLossPercentage)
	}

	// if packetLossPercentage is not set, intrepreter will automatically default it to 0
	// which matches to no packet loss
	if packetDelay != nil {
		return NewConnectionConfig(packetLossPercentage, packetDelay), nil
	}
	return NewConnectionConfig(packetLossPercentage, NoPacketDelay), nil
}

// String the starlark.Value interface
func (connectionConfig *ConnectionConfig) String() string {
	buffer := new(strings.Builder)
	buffer.WriteString(ConnectionConfigTypeName + "(")
	buffer.WriteString(packetLossPercentageAttr + "=")
	buffer.WriteString(fmt.Sprintf("%v, ", connectionConfig.packetLossPercentage))
	buffer.WriteString(packetDelayAttr + "=")
	buffer.WriteString(fmt.Sprintf("%v)", connectionConfig.packetDelay))
	return buffer.String()
}

// Type implements the starlark.Value interface
func (connectionConfig *ConnectionConfig) Type() string {
	return ConnectionConfigTypeName
}

// Freeze implements the starlark.Value interface
func (connectionConfig *ConnectionConfig) Freeze() {
	// this is a no-op its already immutable
}

// Truth implements the starlark.Value interface
func (connectionConfig *ConnectionConfig) Truth() starlark.Bool {
	if connectionConfig.packetLossPercentage < starlark.Float(0) {
		return starlark.False
	}
	if connectionConfig.packetLossPercentage > starlark.Float(100) {
		return starlark.False
	}
	if connectionConfig.packetDelay == nil {
		return starlark.False
	}
	return starlark.True
}

// Hash implements the starlark.Value interface
// This shouldn't be hashed
func (connectionConfig *ConnectionConfig) Hash() (uint32, error) {
	return 0, startosis_errors.NewInterpretationError("unhashable type: '%s'", ConnectionConfigTypeName)
}

// Attr implements the starlark.HasAttrs interface.
func (connectionConfig *ConnectionConfig) Attr(name string) (starlark.Value, error) {
	switch name {
	case packetLossPercentageAttr:
		return connectionConfig.packetLossPercentage, nil
	case packetDelayAttr:
		return connectionConfig.packetDelay, nil
	default:
		return nil, startosis_errors.NewInterpretationError("'%s' has no attribute '%s'", ConnectionConfigTypeName, name)
	}
}

// AttrNames implements the starlark.HasAttrs interface.
func (connectionConfig *ConnectionConfig) AttrNames() []string {
	return []string{packetLossPercentageAttr, packetDelayAttr}
}

func (connectionConfig *ConnectionConfig) ToKurtosisType() partition_topology.PartitionConnection {
	return partition_topology.NewPartitionConnection(
		partition_topology.NewPacketLoss(float32(connectionConfig.packetLossPercentage)),
		connectionConfig.packetDelay.ToKurtosisType(),
	)
}
