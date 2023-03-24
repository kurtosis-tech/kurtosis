package connection_config

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/packet_delay_distribution"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
)

const (
	ConnectionConfigTypeName = "ConnectionConfig"

	PacketLossPercentageAttr    = "packet_loss_percentage"
	PacketDelayDistributionAttr = "packet_delay_distribution"
)

func NewConnectionConfigType() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: ConnectionConfigTypeName,

			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              PacketLossPercentageAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Float],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.FloatInRange(value, PacketLossPercentageAttr, 0, 100)
					},
				},
				{
					Name:              PacketDelayDistributionAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[packet_delay_distribution.PacketDelayDistribution],
					Validator:         nil,
				},
			},
		},

		Instantiate: instantiate,
	}
}

func instantiate(arguments *builtin_argument.ArgumentValuesSet) (builtin_argument.KurtosisValueType, *startosis_errors.InterpretationError) {
	kurtosisValueType, err := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(ConnectionConfigTypeName, arguments)
	if err != nil {
		return nil, err
	}
	return &ConnectionConfig{
		KurtosisValueTypeDefault: kurtosisValueType,
	}, nil
}

type ConnectionConfig struct {
	*kurtosis_type_constructor.KurtosisValueTypeDefault
}

func CreateConnectionConfig(packetLossPercentage starlark.Float) (*ConnectionConfig, *startosis_errors.InterpretationError) {
	args := []starlark.Value{
		packetLossPercentage,
		nil, // no delay distribution as we don't need it
	}
	argumentDefinitions := NewConnectionConfigType().KurtosisBaseBuiltin.Arguments
	argumentValuesSet := builtin_argument.NewArgumentValuesSet(argumentDefinitions, args)
	kurtosisDefaultValue, interpretationErr := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(ConnectionConfigTypeName, argumentValuesSet)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	return &ConnectionConfig{
		KurtosisValueTypeDefault: kurtosisDefaultValue,
	}, nil
}

func (connectionConfig *ConnectionConfig) Copy() (builtin_argument.KurtosisValueType, error) {
	copiedValueType, err := connectionConfig.KurtosisValueTypeDefault.Copy()
	if err != nil {
		return nil, err
	}
	return &ConnectionConfig{
		KurtosisValueTypeDefault: copiedValueType,
	}, nil
}

func (connectionConfig *ConnectionConfig) ToKurtosisType() (*partition_topology.PartitionConnection, *startosis_errors.InterpretationError) {
	packetLossPctStarlark, _, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.Float](
		connectionConfig.KurtosisValueTypeDefault, PacketLossPercentageAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	packetLossPct := float32(packetLossPctStarlark)

	packetDelayDistributionStarlark, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[packet_delay_distribution.PacketDelayDistribution](
		connectionConfig.KurtosisValueTypeDefault, PacketDelayDistributionAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	var packetDelayDistribution partition_topology.PacketDelayDistribution
	if found {
		packetDelayDistributionPtr, interpretationErr := packetDelayDistributionStarlark.ToKurtosisType()
		if interpretationErr != nil {
			return nil, interpretationErr
		}
		packetDelayDistribution = *packetDelayDistributionPtr
	} else {
		packetDelayDistribution = partition_topology.NewUniformPacketDelayDistribution(0)
	}
	partitionConnection := partition_topology.NewPartitionConnection(
		partition_topology.NewPacketLoss(packetLossPct),
		packetDelayDistribution,
	)
	return &partitionConnection, nil
}
