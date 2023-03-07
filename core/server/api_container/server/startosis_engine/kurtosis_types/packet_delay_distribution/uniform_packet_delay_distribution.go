package packet_delay_distribution

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
	"math"
)

const (
	UniformPacketDelayDistributionTypeName = "UniformPacketDelayDistribution"

	DelayAttr = "ms"
)

func NewUniformPacketDelayDistributionType() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: UniformPacketDelayDistributionTypeName,

			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              DelayAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Int],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.Uint64InRange(value, DelayAttr, 0, math.MaxUint32)
					},
				},
			},
		},

		Instantiate: instantiateUniformPacketDelayDistribution,
	}
}

func instantiateUniformPacketDelayDistribution(arguments *builtin_argument.ArgumentValuesSet) (kurtosis_type_constructor.KurtosisValueType, *startosis_errors.InterpretationError) {
	kurtosisValueType, err := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(UniformPacketDelayDistributionTypeName, arguments)
	if err != nil {
		return nil, err
	}
	return &UniformPacketDelayDistribution{
		KurtosisValueTypeDefault: kurtosisValueType,
	}, nil
}

type UniformPacketDelayDistribution struct {
	*kurtosis_type_constructor.KurtosisValueTypeDefault
}

func (packetDelayDistribution *UniformPacketDelayDistribution) ToKurtosisType() (*partition_topology.PacketDelayDistribution, *startosis_errors.InterpretationError) {
	delayMsStarlark, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.Int](
		packetDelayDistribution.KurtosisValueTypeDefault, DelayAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if !found {
		return nil, startosis_errors.NewInterpretationError("Required argument '%s' could not be found on '%s'", DelayAttr, UniformPacketDelayDistributionTypeName)
	}
	delayMs, ok := delayMsStarlark.Uint64()
	if !ok {
		return nil, startosis_errors.NewInterpretationError("Argument '%s' on '%s' was out of bounds", DelayAttr, UniformPacketDelayDistributionTypeName)
	}

	normalPacketDelayDistribution := partition_topology.NewUniformPacketDelayDistribution(uint32(delayMs))
	return &normalPacketDelayDistribution, nil
}
