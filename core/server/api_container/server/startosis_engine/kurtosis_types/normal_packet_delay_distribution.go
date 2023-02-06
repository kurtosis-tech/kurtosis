package kurtosis_types

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
	"strings"
)

const (
	meanAttr                          = "mean_ms"
	stdDevAttr                        = "std_dev_ms"
	correlationAttr                   = "correlation"
	NormalPacketDelayDistributionName = "NormalPacketDelayDistribution"
)

type NormalPacketDelayDistribution struct {
	meanMs      uint32
	stdDevMs    uint32
	correlation starlark.Float
}

func NewNormalPacketDelayDistribution(meanMs uint32, stdDevMs uint32, correlation starlark.Float) *NormalPacketDelayDistribution {
	return &NormalPacketDelayDistribution{
		meanMs:      meanMs,
		stdDevMs:    stdDevMs,
		correlation: correlation,
	}
}

func MakeNormalPacketDelayDistribution(_ *starlark.Thread, builtin *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var meanMs uint32
	var stdDevMs uint32
	var correlation starlark.Float

	if err := starlark.UnpackArgs(
		builtin.Name(),
		args,
		kwargs,
		meanAttr, &meanMs,
		stdDevAttr, &stdDevMs,
		MakeOptional(correlationAttr), &correlation); err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, fmt.Sprintf("Cannot construct a %v from the provided arguments", NormalPacketDelayDistributionName))
	}

	if correlation < 0 || correlation > 100 {
		return nil, startosis_errors.NewInterpretationError("Invalid attribute. '%s' in '%s' should be greater than 0 and lower than 100. Got '%v'", correlationAttr, NormalPacketDelayDistributionName, correlation)
	}

	return NewNormalPacketDelayDistribution(meanMs, stdDevMs, correlation), nil
}

// String the starlark.Value interface
func (ps *NormalPacketDelayDistribution) String() string {
	buffer := new(strings.Builder)
	buffer.WriteString(NormalPacketDelayDistributionName + "(")
	buffer.WriteString(meanAttr + "=")
	buffer.WriteString(fmt.Sprintf("%v, ", ps.meanMs))
	buffer.WriteString(stdDevAttr + "=")
	buffer.WriteString(fmt.Sprintf("%v, ", ps.stdDevMs))
	buffer.WriteString(correlationAttr + "=")
	buffer.WriteString(fmt.Sprintf("%v)", ps.correlation))
	return buffer.String()
}

// Type implements the starlark.Value interface
func (ps *NormalPacketDelayDistribution) Type() string {
	return NormalPacketDelayDistributionName
}

// Freeze implements the starlark.Value interface
func (ps *NormalPacketDelayDistribution) Freeze() {
	// this is a no-op its already immutable
}

// Truth implements the starlark.Value interface
func (ps *NormalPacketDelayDistribution) Truth() starlark.Bool {
	return ps.meanMs != 0 && ps.stdDevMs != 0 && ps.correlation >= starlark.Float(0) && ps.correlation <= starlark.Float(100)
}

// Hash implements the starlark.Value interface
// This shouldn't be hashed, users should use a portId instead
func (ps *NormalPacketDelayDistribution) Hash() (uint32, error) {
	return 0, startosis_errors.NewInterpretationError("unhashable type: '%v'", NormalPacketDelayDistributionName)
}

// Attr implements the starlark.HasAttrs interface.
func (ps *NormalPacketDelayDistribution) Attr(name string) (starlark.Value, error) {
	switch name {
	case meanAttr:
		return starlark.MakeInt(int(ps.meanMs)), nil
	case stdDevAttr:
		return starlark.MakeInt(int(ps.stdDevMs)), nil
	case correlationAttr:
		return ps.correlation, nil
	default:
		return nil, startosis_errors.NewInterpretationError("'%v' has no attribute '%v;", NormalPacketDelayDistributionName, name)
	}
}

// AttrNames implements the starlark.HasAttrs interface.
func (ps *NormalPacketDelayDistribution) AttrNames() []string {
	return []string{meanAttr, stdDevAttr, correlationAttr}
}

func (ps *NormalPacketDelayDistribution) ToKurtosisType() partition_topology.PacketDelayDistribution {
	return partition_topology.NewNormalPacketDelayDistribution(ps.meanMs, ps.stdDevMs, float32(ps.correlation))
}
