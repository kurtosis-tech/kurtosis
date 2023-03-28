package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/packet_delay_distribution"
	"github.com/stretchr/testify/require"
	"testing"
)

type uniformPacketDelayDistributionTestCase struct {
	*testing.T
}

func newUniformPacketDelayDistributionTestCase(t *testing.T) *uniformPacketDelayDistributionTestCase {
	return &uniformPacketDelayDistributionTestCase{
		T: t,
	}
}

func (t *uniformPacketDelayDistributionTestCase) GetId() string {
	return packet_delay_distribution.UniformPacketDelayDistributionTypeName
}

func (t *uniformPacketDelayDistributionTestCase) GetTypeConstructor() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return packet_delay_distribution.NewUniformPacketDelayDistributionType()
}

func (t *uniformPacketDelayDistributionTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%d)", packet_delay_distribution.UniformPacketDelayDistributionTypeName, packet_delay_distribution.DelayAttr, 110)
}

func (t *uniformPacketDelayDistributionTestCase) Assert(typeValue builtin_argument.KurtosisValueType) {
	uniformPacketDelayDistributionStarlark, ok := typeValue.(*packet_delay_distribution.UniformPacketDelayDistribution)
	require.True(t, ok)
	uniformPacketDelayDistribution, err := uniformPacketDelayDistributionStarlark.ToKurtosisType()
	require.Nil(t, err)

	expectedUniformPacketDelayDistribution := partition_topology.NewUniformPacketDelayDistribution(110)
	require.Equal(t, expectedUniformPacketDelayDistribution, *uniformPacketDelayDistribution)
}
