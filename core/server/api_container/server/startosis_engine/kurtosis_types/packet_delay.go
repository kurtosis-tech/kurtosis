package kurtosis_types

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
	"strings"
)

const (
	delayInMsAttr   = "delay_ms"
	PacketDelayName = "PacketDelay"
)

// PacketDelay TODO: added ability to update jitter and maybe correlation, depending upon the product!
type PacketDelay struct {
	avgDelayMs uint32
}

var (
	NoPacketDelay = NewPacketDelay(0)
)

func NewPacketDelay(delayInMs uint32) *PacketDelay {
	return &PacketDelay{
		avgDelayMs: delayInMs,
	}
}

//MakePacketDelay The reason I have delayInMsAttr optional because it could be possible to only set std dev and not avg delay
func MakePacketDelay(_ *starlark.Thread, builtin *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var delayInMs uint32
	if err := starlark.UnpackArgs(builtin.Name(), args, kwargs, delayInMsAttr, &delayInMs); err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Cannot construct a PacketDelay from the provided arguments")
	}
	return NewPacketDelay(delayInMs), nil
}

// String the starlark.Value interface
func (ps *PacketDelay) String() string {
	buffer := new(strings.Builder)
	buffer.WriteString(PacketDelayName + "(")
	buffer.WriteString(delayInMsAttr + "=")
	buffer.WriteString(fmt.Sprintf("%v)", ps.avgDelayMs))
	return buffer.String()
}

// Type implements the starlark.Value interface
func (ps *PacketDelay) Type() string {
	return PacketDelayName
}

// Freeze implements the starlark.Value interface
func (ps *PacketDelay) Freeze() {
	// this is a no-op its already immutable
}

// Truth implements the starlark.Value interface
func (ps *PacketDelay) Truth() starlark.Bool {
	return ps.avgDelayMs != 0
}

// Hash implements the starlark.Value interface
// This shouldn't be hashed, users should use a portId instead
func (ps *PacketDelay) Hash() (uint32, error) {
	return 0, startosis_errors.NewInterpretationError("unhashable type: '%v'", PacketDelayName)
}

// Attr implements the starlark.HasAttrs interface.
func (ps *PacketDelay) Attr(name string) (starlark.Value, error) {
	switch name {
	case delayInMsAttr:
		return starlark.MakeInt(int(ps.avgDelayMs)), nil
	default:
		return nil, startosis_errors.NewInterpretationError("'%v' has no attribute '%v;", PacketDelayName, name)
	}
}

// AttrNames implements the starlark.HasAttrs interface.
func (ps *PacketDelay) AttrNames() []string {
	return []string{delayInMsAttr}
}

func (ps *PacketDelay) ToKurtosisType() partition_topology.PacketDelay {
	return partition_topology.NewPacketDelay(ps.avgDelayMs)
}
