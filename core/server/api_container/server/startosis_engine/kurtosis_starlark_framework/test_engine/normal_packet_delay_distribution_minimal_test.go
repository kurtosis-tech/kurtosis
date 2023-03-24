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

type normalPacketDelayDistributionMinimalTestCase struct {
	*testing.T
}

func newNormalPacketDelayDistributionMinimalTestCase(t *testing.T) *normalPacketDelayDistributionMinimalTestCase {
	return &normalPacketDelayDistributionMinimalTestCase{
		T: t,
	}
}

func (t *normalPacketDelayDistributionMinimalTestCase) GetId() string {
	return fmt.Sprintf("%s_%s", packet_delay_distribution.NormalPacketDelayDistributionTypeName, "Minimal")
}

func (t *normalPacketDelayDistributionMinimalTestCase) GetTypeConstructor() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return packet_delay_distribution.NewNormalPacketDelayDistributionType()
}

func (t *normalPacketDelayDistributionMinimalTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%d, %s=%d)", packet_delay_distribution.NormalPacketDelayDistributionTypeName, packet_delay_distribution.MeanAttr, 110, packet_delay_distribution.StdDevAttr, 16)
}

func (t *normalPacketDelayDistributionMinimalTestCase) Assert(typeValue builtin_argument.KurtosisValueType) {
	normalPacketDelayDistributionStarlark, ok := typeValue.(*packet_delay_distribution.NormalPacketDelayDistribution)
	require.True(t, ok)
	normalPacketDelayDistribution, err := normalPacketDelayDistributionStarlark.ToKurtosisType()
	require.Nil(t, err)

	expectedNormalPacketDelayDistribution := partition_topology.NewNormalPacketDelayDistribution(110, 16, 0)
	require.Equal(t, expectedNormalPacketDelayDistribution, *normalPacketDelayDistribution)
}
