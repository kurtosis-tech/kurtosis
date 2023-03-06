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
	NormalPacketDelayDistributionTypeName = "NormalPacketDelayDistribution"

	MeanAttr        = "mean_ms"
	StdDevAttr      = "std_dev_ms"
	CorrelationAttr = "correlation"
)

func NewNormalPacketDelayDistributionType() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: NormalPacketDelayDistributionTypeName,

			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              MeanAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Int],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.Uint64InRange(value, MeanAttr, 0, math.MaxUint32)
					},
				},
				{
					Name:              StdDevAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Int],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.Uint64InRange(value, StdDevAttr, 0, math.MaxUint32)
					},
				},
				{
					Name:              CorrelationAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Float],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {

						return builtin_argument.FloatInRange(value, CorrelationAttr, 0, 100)
					},
				},
			},
		},

		Instantiate: instantiateNormalPacketDelayDistribution,
	}
}

func instantiateNormalPacketDelayDistribution(arguments *builtin_argument.ArgumentValuesSet) (kurtosis_type_constructor.KurtosisValueType, *startosis_errors.InterpretationError) {
	kurtosisValueType, err := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(NormalPacketDelayDistributionTypeName, arguments)
	if err != nil {
		return nil, err
	}
	return &NormalPacketDelayDistribution{
		KurtosisValueTypeDefault: kurtosisValueType,
	}, nil
}

type NormalPacketDelayDistribution struct {
	*kurtosis_type_constructor.KurtosisValueTypeDefault
}

func (packetDelayDistribution *NormalPacketDelayDistribution) ToKurtosisType() (*partition_topology.PacketDelayDistribution, *startosis_errors.InterpretationError) {
	meanStarlark, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.Int](
		packetDelayDistribution.KurtosisValueTypeDefault, MeanAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if !found {
		return nil, startosis_errors.NewInterpretationError("Required argument '%s' could not be found on '%s'", MeanAttr, NormalPacketDelayDistributionTypeName)
	}
	mean, ok := meanStarlark.Uint64()
	if !ok {
		return nil, startosis_errors.NewInterpretationError("Argument '%s' on '%s' was out of bounds", MeanAttr, NormalPacketDelayDistributionTypeName)
	}

	stdDevStarlark, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.Int](
		packetDelayDistribution.KurtosisValueTypeDefault, StdDevAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if !found {
		return nil, startosis_errors.NewInterpretationError("Required argument '%s' could not be found on '%s'", StdDevAttr, NormalPacketDelayDistributionTypeName)
	}
	stdDev, ok := stdDevStarlark.Uint64()
	if !ok {
		return nil, startosis_errors.NewInterpretationError("Argument '%s' on '%s' was out of bounds", StdDevAttr, NormalPacketDelayDistributionTypeName)
	}

	correlationStarlark, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.Float](
		packetDelayDistribution.KurtosisValueTypeDefault, CorrelationAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	var correlation float32
	if !found {
		correlation = 0
	} else {
		correlation = float32(correlationStarlark)
	}

	normalPacketDelayDistribution := partition_topology.NewNormalPacketDelayDistribution(uint32(mean), uint32(stdDev), correlation)
	return &normalPacketDelayDistribution, nil
}
