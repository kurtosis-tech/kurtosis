package kurtosis_types

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
	"strings"
)

const (
	delayAttr                          = "ms"
	UniformPacketDelayDistributionName = "UniformPacketDelayDistribution"
)

type UniformPacketDelayDistribution struct {
	delayMs uint32
}

func NewUniformPacketDelayDistribution(delayMs uint32) *UniformPacketDelayDistribution {
	return &UniformPacketDelayDistribution{
		delayMs: delayMs,
	}
}

func MakeUniformPacketDelayDistribution(_ *starlark.Thread, builtin *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var delayMs uint32
	if err := starlark.UnpackArgs(builtin.Name(), args, kwargs, delayAttr, &delayMs); err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, fmt.Sprintf("Cannot construct a %v from the provided arguments", UniformPacketDelayDistributionName))
	}
	return NewUniformPacketDelayDistribution(delayMs), nil
}

// String the starlark.Value interface
func (ps *UniformPacketDelayDistribution) String() string {
	buffer := new(strings.Builder)
	buffer.WriteString(UniformPacketDelayDistributionName + "(")
	buffer.WriteString(delayAttr + "=")
	buffer.WriteString(fmt.Sprintf("%v)", ps.delayMs))
	return buffer.String()
}

// Type implements the starlark.Value interface
func (ps *UniformPacketDelayDistribution) Type() string {
	return UniformPacketDelayDistributionName
}

// Freeze implements the starlark.Value interface
func (ps *UniformPacketDelayDistribution) Freeze() {
	// this is a no-op its already immutable
}

// Truth implements the starlark.Value interface
func (ps *UniformPacketDelayDistribution) Truth() starlark.Bool {
	return ps.delayMs != 0
}

// Hash implements the starlark.Value interface
// This shouldn't be hashed, users should use a portId instead
func (ps *UniformPacketDelayDistribution) Hash() (uint32, error) {
	return 0, startosis_errors.NewInterpretationError("unhashable type: '%v'", UniformPacketDelayDistributionName)
}

// Attr implements the starlark.HasAttrs interface.
func (ps *UniformPacketDelayDistribution) Attr(name string) (starlark.Value, error) {
	switch name {
	case delayAttr:
		return starlark.MakeInt(int(ps.delayMs)), nil
	default:
		return nil, startosis_errors.NewInterpretationError("'%v' has no attribute '%v;", UniformPacketDelayDistributionName, name)
	}
}

// AttrNames implements the starlark.HasAttrs interface.
func (ps *UniformPacketDelayDistribution) AttrNames() []string {
	return []string{delayAttr}
}

func (ps *UniformPacketDelayDistribution) ToKurtosisType() partition_topology.PacketDelayDistribution {
	return partition_topology.NewUniformPacketDelayDistribution(ps.delayMs)
}
